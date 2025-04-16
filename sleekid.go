package sleekid

import (
	"bytes"
	"fmt"
	"time"

	"crypto/rand"
	"crypto/subtle"
)

var generator Generator = NewGenerator(GeneratorInit{})

// Setup initializes the generator.
//
//	sleekid.Setup(sleekid.GeneratorInit{
//		ChecksumToken: 12345,
//	})
func Setup(init GeneratorInit) {
	generator = NewGenerator(init)
}

// New generates a new id with the given prefix.
//
//	id, err := sleekid.New("usr")
//	id, err := sleekid.New("usr", sleekid.WithRandomBytes(16))
func New(prefix string, options ...*GenerateOption) (SleekId, error) {
	return generator.New(prefix, options...)
}

// Prefix returns the prefix of the given id.
//
//	sleekid.Prefix(id)
func Prefix(id SleekId) string {
	return generator.Prefix(id)
}

// Timestamp returns the unix time of the timestamp part.
//
//	sleekid.Timestamp(id)
func Timestamp(id SleekId) time.Time {
	return generator.Timestamp(id)
}

// Validate checks if the given the timestamp and random part is valid.
//
//	sleekid.Validate(id)
func Validate(id SleekId) bool {
	return generator.Validate(id)
}

// ValidateWithPrefix checks if the given id is valid with the specific prefix..
//
//	sleekid.ValidateWithPrefix("usr", id)
func ValidateWithPrefix(prefix string, id SleekId) bool {
	return generator.ValidateWithPrefix(prefix, id)
}

type SleekId []byte

func (id SleekId) String() string {
	// return *(*string)(unsafe.Pointer(&id))
	return string(id)
}

type GenerateOption struct {
	// RandomDigitsLength is the length of the random part of the id.
	RandomDigitsLength int
}

type TimestampOrder int

const (
	TimestampOrderAlphabetical TimestampOrder = 0
	TimestampOrderASCII        TimestampOrder = 1
)

type GeneratorInit struct {
	// delimiter is the character used to separate the prefix from the rest of the id.
	//
	// Default is "_".
	Delimiter rune

	// checksumToken is used to verify the id. Don't expose it to the public.
	//
	// Default is 4567890. Change it on your production environment.
	ChecksumToken uint64

	// ChecksumLength is the length of the checksum part of the id.
	// This will increase the precisition of the false detection rate.
	//
	// the probability of false detection is 1 - 1/62^ChecksumLength.
	// 2 is enough for most cases. It's 99.97%.
	//
	// Default is 2.
	ChecksumLength int

	// RandomDigitsLength is the length of the random part of the id.
	// Also you can customize this length when you call New() with WithRandomBytes().
	//
	// Default is 12.
	RandomDigitsLength int

	// TimestampLength is the length of the timestamp part of the id.
	//
	// Default is 5.
	// must be 4 <= TimestampLength <= 6.
	TimestampLength int

	// TimestampOrder is the order of the timestamp part of the id that sleekid generates.
	//
	// Default is Alphabetical order.
	TimestampOrder TimestampOrder
}

// WithRandomBytes is a helper function to set the RandomDigitsLength option.
func WithRandomBytes(length int) *GenerateOption {
	return &GenerateOption{RandomDigitsLength: length}
}

type Generator interface {
	// New generates a new id with the given prefix.
	//
	//	gen := NewGenerator(GeneratorInit{...})
	//	id, err := gen.New("usr")
	//	id, err := gen.New("usr", WithRandomBytes(16))
	New(prefix string, options ...*GenerateOption) (SleekId, error)

	// Prefix returns the prefix of the given id.
	//
	//	gen := NewGenerator(GeneratorInit{...})
	//	prefix := gen.Prefix(id)
	Prefix(id SleekId) string

	// Timestamp returns the unix time of the timestamp part.
	//
	//	gen := NewGenerator(GeneratorInit{...})
	//	timestamp := gen.Timestamp(id)
	Timestamp(id SleekId) time.Time

	// Validate checks if the given the timestamp and random part is valid.
	//
	//	gen := NewGenerator(GeneratorInit{...})
	//	valid := gen.Validate(id)
	Validate(id SleekId) bool

	// ValidateWithPrefix checks if the given id is valid with the specific prefix.
	//
	//	gen := NewGenerator(GeneratorInit{...})
	//	valid := gen.ValidateWithPrefix("usr", id)
	ValidateWithPrefix(prefix string, id SleekId) bool
}

type sleekIdGen struct {
	delimiter          byte
	checksumToken      uint64
	checksumLength     int
	randomDigitsLength int
	timestampLength    int
	alphabet           string
	alphabetBytes      []byte
}

const baseUnixEpoch = 1704067200 // 2024-01-01 00:00:00 UTC

func NewGenerator(init GeneratorInit) Generator {
	delimiter := byte('_')
	if init.Delimiter != 0 {
		delimiter = byte(init.Delimiter)
	}
	checksumLength := 2
	if init.ChecksumLength != 0 {
		checksumLength = init.ChecksumLength
	}
	randomDigitsLength := 12
	if init.RandomDigitsLength != 0 {
		randomDigitsLength = init.RandomDigitsLength
	}
	checksumToken := uint64(4567890)
	if init.ChecksumToken != 0 {
		checksumToken = init.ChecksumToken
	}
	timestampLength := 5
	if 4 <= init.TimestampLength && init.TimestampLength <= 6 {
		timestampLength = init.TimestampLength
	} else if init.TimestampLength != 0 {
		panic("TimestampLength must be 4 <= TimestampLength <= 6")
	}
	// Alphabetical order: 0-9 < a-z < A-Z
	alphabet := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	if init.TimestampOrder == TimestampOrderASCII {
		// ASCII order: 0-9 < A-Z < a-z
		alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	}
	return &sleekIdGen{
		delimiter:          delimiter,
		checksumToken:      checksumToken,
		checksumLength:     checksumLength,
		randomDigitsLength: randomDigitsLength,
		timestampLength:    timestampLength,
		alphabet:           alphabet,
		alphabetBytes:      []byte(alphabet),
	}
}

func (o *sleekIdGen) New(prefix string, options ...*GenerateOption) (SleekId, error) {
	randomDigitsLength := o.randomDigitsLength
	if len(options) > 0 {
		randomDigitsLength = options[0].RandomDigitsLength
	}
	timestamp := timestampToSortableString(time.Now(), o.timestampLength, o.alphabet)
	randomBytes := make([]byte, randomDigitsLength)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	for i, b := range randomBytes {
		randomBytes[i] = o.alphabet[b%62]
	}

	// prefix + "_" + random bits + checksum bits
	id := make([]byte, 0, len(prefix)+1+len(timestamp)+randomDigitsLength+o.checksumLength)
	id = append(id, prefix...)
	id = append(id, o.delimiter)
	id = append(id, timestamp...)
	id = append(id, randomBytes...)
	checksum := generateChecksum(id, o.checksumLength, o.checksumToken, o.alphabet)
	id = append(id, checksum...)
	return SleekId(id), nil
}

func (o *sleekIdGen) Prefix(id SleekId) string {
	delimiterPos := -1
	for i, b := range id {
		if b == o.delimiter {
			delimiterPos = i
			break
		}
	}
	if delimiterPos == -1 {
		return ""
	}
	return string(id[:delimiterPos])
}

func (o *sleekIdGen) Timestamp(id SleekId) time.Time {
	prefix := o.Prefix(id)
	timestamp := id[len(prefix)+1 : len(prefix)+1+o.timestampLength]
	return timestampToUnixTime(timestamp, o.alphabetBytes)
}

func (o *sleekIdGen) Validate(id SleekId) bool {
	if len(id) < o.checksumLength+o.timestampLength {
		return false
	}
	idPart, checksum := id[:len(id)-o.checksumLength], id[len(id)-o.checksumLength:]
	return subtle.ConstantTimeCompare(checksum, generateChecksum(idPart, o.checksumLength, o.checksumToken, o.alphabet)) == 1
}

func (o *sleekIdGen) ValidateWithPrefix(prefix string, id SleekId) bool {
	if !bytes.HasPrefix(id, append([]byte(prefix), o.delimiter)) {
		return false
	}
	return o.Validate(id)
}

func timestampToSortableString(t time.Time, length int, alphabet string) []byte {
	timeValue := t.Unix() - baseUnixEpoch

	result := make([]byte, 0, length)

	for timeValue > 0 {
		result = append(result, alphabet[int(timeValue%62)])
		timeValue = timeValue / 62
	}

	for len(result) < length {
		result = append(result, alphabet[0])
	}

	if len(result) > length {
		result = result[:length]
	}

	// make sortable
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result
}

func timestampToUnixTime(s []byte, alphabetBytes []byte) time.Time {
	// convert 62 base to 10 base
	var timeValue int64 = 0
	multiplier := int64(1)

	for i := len(s) - 1; i >= 0; i-- {
		c := s[i]
		// find the position of the alphabet
		pos := -1
		for i, char := range alphabetBytes {
			if c == char {
				pos = i
				break
			}
		}
		if pos == -1 {
			return time.Time{}
		}

		timeValue += int64(pos) * multiplier
		multiplier *= 62
	}

	// add baseUnixEpoch to get the original unix timestamp
	timeValue += baseUnixEpoch

	// convert unix timestamp to time.Time
	return time.Unix(timeValue, 0)
}

// generateChecksum creates a 2-character checksum for the given string
func generateChecksum(str []byte, length int, token uint64, alphabet string) []byte {
	var hash1 uint32 = uint32(token >> 32) // top 32 bits
	var hash2 uint32 = uint32(token)

	for i := 0; i < len(str); i++ {
		hash1 = ((hash1 << 5) - hash1) + uint32(str[i])
		hash1 ^= (hash1 >> 16)

		hash2 = ((hash2 << 5) - hash2) + uint32(str[i])
		hash2 ^= (hash2 >> 16)
	}

	combined := hash1 ^ hash2

	// Convert to length-character base62
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = alphabet[combined%62]
		combined /= 62
	}
	return result
}
