package sleekid

import (
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rs/xid"
	"gotest.tools/assert"
)

func BenchmarkNew(b *testing.B) {
	gen := NewGenerator(GeneratorInit{ChecksumToken: 30, RandomDigitsLength: 15})
	prefix := "usr"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gen.New(prefix)
	}
}

func BenchmarkPrefix(b *testing.B) {
	gen := NewGenerator(GeneratorInit{ChecksumToken: 30, RandomDigitsLength: 15, ChecksumLength: 4, Delimiter: '_'})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.Prefix([]byte("usr_ajijoi"))
	}
}

func BenchmarkTimestamp(b *testing.B) {
	gen := NewGenerator(GeneratorInit{ChecksumToken: 30, RandomDigitsLength: 2, TimestampLength: 5})
	id, _ := gen.New("usr")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.Timestamp(id)
	}
}

func BenchmarkValidate(b *testing.B) {
	gen := NewGenerator(GeneratorInit{ChecksumToken: 30, RandomDigitsLength: 15})
	prefix := "usr"
	id, _ := gen.New(prefix)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.Validate(id)
	}
}

func BenchmarkValidateWithPrefix(b *testing.B) {
	gen := NewGenerator(GeneratorInit{ChecksumToken: 30, RandomDigitsLength: 15})
	prefix := "usr"
	id, _ := gen.New(prefix)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.ValidateWithPrefix(prefix, id)
	}
}

func BenchmarkNewUUID(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = uuid.NewRandom()
	}
}

func BenchmarkNewXid(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = xid.New()
	}
}

func TestNew(t *testing.T) {
	gen := NewGenerator(GeneratorInit{ChecksumToken: 30, RandomDigitsLength: 10})
	id, _ := gen.New("usr")

	assert.Equal(t, 3+1+5+10+2, len(id))
	assert.Assert(t, regexp.MustCompile(`^[a-z]+_[A-Za-z0-9]+$`).MatchString(string(id)))
}

func TestPrefix(t *testing.T) {
	gen := NewGenerator(GeneratorInit{ChecksumToken: 30, RandomDigitsLength: 10})
	id, _ := gen.New("usr")
	assert.Equal(t, "usr", gen.Prefix(id))
}

func TestPrefix_shouldReturnEmpty(t *testing.T) {
	gen := NewGenerator(GeneratorInit{ChecksumToken: 30, RandomDigitsLength: 10})
	assert.Equal(t, "", gen.Prefix([]byte("ajijoi")))
}

func TestTimestamp(t *testing.T) {
	gen := NewGenerator(GeneratorInit{ChecksumToken: 30, RandomDigitsLength: 2, TimestampLength: 5})
	id, _ := gen.New("usr")
	now := time.Unix(time.Now().Unix(), 0)
	timestmap := gen.Timestamp(id)
	assert.Equal(t, now, timestmap)
}

func TestTimestampWithCustomTimestampLength(t *testing.T) {
	gen := NewGenerator(GeneratorInit{ChecksumToken: 30, RandomDigitsLength: 2, TimestampLength: 6})
	id, _ := gen.New("usr")
	now := time.Unix(time.Now().Unix(), 0)
	timestmap := gen.Timestamp(id)
	assert.Equal(t, now, timestmap)
}

func TestValidate(t *testing.T) {
	gen := NewGenerator(GeneratorInit{ChecksumToken: 30, RandomDigitsLength: 15})
	id, _ := gen.New("usr")
	assert.Equal(t, true, gen.Validate(id))
	assert.Equal(t, false, gen.Validate([]byte("aoj_aoisjdfoi")))
	assert.Equal(t, false, gen.Validate([]byte("asf")))
}

func TestValidateWithPrefix_shouldFailWhenPrefixIsDifferent(t *testing.T) {
	gen := NewGenerator(GeneratorInit{ChecksumToken: 30, RandomDigitsLength: 15})
	id, _ := gen.New("usr")
	assert.Equal(t, false, gen.ValidateWithPrefix("usr2", id))
}

func TestValidateWithPrefix_shouldFailWhenIdIsTooShort(t *testing.T) {
	gen := NewGenerator(GeneratorInit{ChecksumToken: 30, RandomDigitsLength: 15})
	assert.Equal(t, false, gen.ValidateWithPrefix("usr", []byte("a")))
	assert.Equal(t, false, gen.ValidateWithPrefix("usr", []byte("usr_")))
	assert.Equal(t, false, gen.ValidateWithPrefix("usr", []byte("usr_aos")))
}

func TestValidateWithPrefix_shouldFailWhenTokenIsDifferent(t *testing.T) {
	gen := NewGenerator(GeneratorInit{ChecksumToken: 30, RandomDigitsLength: 15})
	id, _ := gen.New("usr")
	gen2 := NewGenerator(GeneratorInit{ChecksumToken: 100, RandomDigitsLength: 15})
	assert.Equal(t, false, gen2.ValidateWithPrefix("usr", id))
}

func TestValidateWithPrefix_shouldWorkWhenAllIsGood(t *testing.T) {
	gen := NewGenerator(GeneratorInit{ChecksumToken: 30, RandomDigitsLength: 15})
	id, _ := gen.New("usr")
	gen2 := NewGenerator(GeneratorInit{ChecksumToken: 30, RandomDigitsLength: 15})
	assert.Equal(t, true, gen2.ValidateWithPrefix("usr", id))
}

func TestNewConcurrentGeneration(t *testing.T) {
	const numGoroutines = 150
	const numIDs = 2000

	gen := NewGenerator(GeneratorInit{ChecksumToken: 30, RandomDigitsLength: 15})

	seen := sync.Map{}
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numIDs; j++ {
				id, _ := gen.New("test")
				_, loaded := seen.LoadOrStore(string(id), true)
				assert.Equal(t, false, loaded, "Duplicate ID generated")
			}
		}()
	}
	wg.Wait()
}

func TestCustomChecksumLength(t *testing.T) {
	gen := NewGenerator(GeneratorInit{
		ChecksumToken:      100,
		RandomDigitsLength: 15,
		ChecksumLength:     4,
		TimestampLength:    5,
		Delimiter:          '_',
	})
	id, _ := gen.New("usr")
	assert.Equal(t, true, gen.Validate(id))
	assert.Equal(t, 3+1+5+15+4, len(id))
	assert.Assert(t, regexp.MustCompile(`^[a-z]+_[A-Za-z0-9]+$`).MatchString(string(id)))
}

func TestCustomTimestampLength(t *testing.T) {
	gen := NewGenerator(GeneratorInit{
		ChecksumToken:      100,
		RandomDigitsLength: 15,
		ChecksumLength:     2,
		TimestampLength:    6,
		Delimiter:          '_',
	})
	id, _ := gen.New("usr")
	assert.Equal(t, true, gen.Validate(id))
	assert.Equal(t, 3+1+6+15+2, len(id))
	assert.Assert(t, regexp.MustCompile(`^[a-z]+_[A-Za-z0-9]+$`).MatchString(string(id)))
}

func TestTimestampPadding(t *testing.T) {
	alphabet := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	str := timestampToSortableString(time.Unix(baseUnixEpoch+1, 0), 6, alphabet)
	assert.Equal(t, "000001", string(str))
	str = timestampToSortableString(time.Unix(baseUnixEpoch+60, 0), 6, alphabet)
	assert.Equal(t, "00000y", string(str))
	str = timestampToSortableString(time.Unix(baseUnixEpoch+61, 0), 6, alphabet)
	assert.Equal(t, "00000z", string(str))
	str = timestampToSortableString(time.Unix(baseUnixEpoch+62, 0), 6, alphabet)
	assert.Equal(t, "000010", string(str))
}
