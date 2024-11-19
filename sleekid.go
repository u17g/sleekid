package sleekid

import (
	"bytes"
	"errors"
	"fmt"
	"time"

	"crypto/rand"
)

var generator Generator

// Setup initializes the generator.
//
// It must be called before using New or Validate.
//
//	sleekid.Setup(sleekid.GeneratorInit{
//		Token: 12345,
//		RandomDigitsLength: 12,
//	})
func Setup(init GeneratorInit) {
	generator = NewGenerator(init)
}

// New generates a new id with the given prefix.
//
// Before using New, Setup must be called.
//
//	sleekid.Setup(sleekid.GeneratorInit{
//		Token: 12345,
//		RandomDigitsLength: 12,
//	})
//	id, err := sleekid.New("usr")
//	id, err := sleekid.New("usr", sleekid.WithRandomBytes(16))
func New(prefix string, options ...*GenerateOption) (SleekId, error) {
	if generator == nil {
		return nil, errors.New("must be initialized by sleekid.Setup")
	}
	return generator.New(prefix, options...)
}

// Validate checks if the given id is valid.
//
// Before using Validate, Setup must be called.
//
//	sleekid.Setup(sleekid.GeneratorInit{
//		Token: 12345,
//		RandomDigitsLength: 12,
//	})
//	sleekid.Validate("usr", id)
func Validate(prefix string, id SleekId) bool {
	if generator == nil {
		return false
	}
	return generator.Validate(prefix, id)
}

type SleekId []byte

func (id SleekId) String() string {
	return string(id)
}

type GenerateOption struct {
	RandomDigitsLength int
}
type GeneratorInit struct {
	// token is used to verify the id.
	Token uint64
	// RandomDigitsLength is the length of the random part of the id.
	RandomDigitsLength int
}

func WithRandomBytes(length int) *GenerateOption {
	return &GenerateOption{RandomDigitsLength: length}
}

type Generator interface {
	New(prefix string, options ...*GenerateOption) (SleekId, error)
	Validate(prefix string, id SleekId) bool
}

type sleekIdGen struct {
	Token              uint64
	RandomDigitsLength int
}

// 0-9 < a-z < A-Z
const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const alphabetLength = len(alphabet)
const alphabetLength64 = int64(alphabetLength)
const baseUnixEpoch = 1704067200 // 2024-01-01 00:00:00 UTC

func NewGenerator(init GeneratorInit) Generator {
	return &sleekIdGen{Token: init.Token, RandomDigitsLength: init.RandomDigitsLength}
}

// Returns id with len(prefix) + 4~6 + RandomDigitsLength + 2
func (o *sleekIdGen) New(prefix string, options ...*GenerateOption) (SleekId, error) {
	randomDigitsLength := o.RandomDigitsLength
	if len(options) > 0 {
		randomDigitsLength = options[0].RandomDigitsLength
	}
	timestamp := timestampToSortableString(time.Now())
	randomBytes := make([]byte, randomDigitsLength)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	for i, b := range randomBytes {
		randomBytes[i] = alphabet[b%byte(alphabetLength)]
	}
	idPart := append(timestamp, randomBytes...)
	checksum := generateChecksum(idPart, o.Token)

	// prefix + "_" + random bits + checksum bits
	id := make([]byte, 0, len(prefix)+1+len(idPart)+len(checksum))
	id = append(id, prefix...)
	id = append(id, '_')
	id = append(id, idPart...)
	id = append(id, checksum...)
	return SleekId(id), nil
}

func (o *sleekIdGen) Validate(prefix string, id SleekId) bool {
	if !bytes.HasPrefix(id, append([]byte(prefix), '_')) {
		return false
	}
	id = id[len(prefix)+1:]
	if len(id) < 2 {
		return false
	}
	idPart, checksum := id[:len(id)-2], id[len(id)-2:]
	return bytes.Equal(checksum, generateChecksum(idPart, o.Token))
}

func timestampToSortableString(t time.Time) []byte {
	timeValue := t.Unix() - baseUnixEpoch

	result := make([]byte, 0, 4)

	for timeValue > 0 {
		result = append(result, alphabet[int(timeValue%alphabetLength64)])
		timeValue = timeValue / alphabetLength64
	}

	for len(result) < 4 {
		result = append(result, alphabet[0])
	}

	// make sortable
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result
}

// generateChecksum creates a 2-character checksum for the given string
func generateChecksum(str []byte, token uint64) []byte {
	var hash1 uint32 = uint32(token >> 32) // top 32 bits
	var hash2 uint32 = uint32(token)

	for i := 0; i < len(str); i++ {
		hash1 = ((hash1 << 5) - hash1) + uint32(str[i])
		hash1 ^= (hash1 >> 16)

		hash2 = ((hash2 << 5) - hash2) + uint32(str[i])
		hash2 ^= (hash2 >> 16)
	}

	combined := hash1 ^ hash2

	// Convert to 2-character base62
	return []byte{
		alphabet[combined%62],
		alphabet[(combined/62)%62],
	}
}
