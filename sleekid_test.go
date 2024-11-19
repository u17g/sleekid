package sleekid

import (
	"regexp"
	"sync"
	"testing"

	"github.com/google/uuid"
	"gotest.tools/assert"
)

func BenchmarkNew(b *testing.B) {
	gen := NewGenerator(GeneratorInit{Token: 30, RandomDigitsLength: 15})
	prefix := "usr"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.New(prefix)
	}
}

func BenchmarkValidate(b *testing.B) {
	gen := NewGenerator(GeneratorInit{Token: 30, RandomDigitsLength: 15})
	prefix := "usr"
	id, _ := gen.New(prefix)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.Validate(prefix, id)
	}
}

func BenchmarkNewUUID(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		uuid.NewRandom()
	}
}

func TestNew(t *testing.T) {
	gen := NewGenerator(GeneratorInit{Token: 30, RandomDigitsLength: 10})
	id, _ := gen.New("usr")

	assert.Equal(t, 3+1+5+10+2, len(id))
	assert.Assert(t, regexp.MustCompile(`^[a-z]+_[A-Za-z0-9]+$`).MatchString(string(id)))
}

func TestValidate_shouldFailWhenPrefixIsDifferent(t *testing.T) {
	gen := NewGenerator(GeneratorInit{Token: 30, RandomDigitsLength: 15})
	id, _ := gen.New("usr")
	assert.Equal(t, false, gen.Validate("usr2", id))
}

func TestValidate_shouldFailWhenIdIsTooShort(t *testing.T) {
	gen := NewGenerator(GeneratorInit{Token: 30, RandomDigitsLength: 15})
	assert.Equal(t, false, gen.Validate("usr", []byte("a")))
	assert.Equal(t, false, gen.Validate("usr", []byte("usr_")))
	assert.Equal(t, false, gen.Validate("usr", []byte("usr_aos")))
}

func TestValidate_shouldFailWhenTokenIsDifferent(t *testing.T) {
	gen := NewGenerator(GeneratorInit{Token: 30, RandomDigitsLength: 15})
	id, _ := gen.New("usr")
	gen2 := NewGenerator(GeneratorInit{Token: 100, RandomDigitsLength: 15})
	assert.Equal(t, false, gen2.Validate("usr", id))
}

func TestValidate_shouldWorkWhenAllIsGood(t *testing.T) {
	gen := NewGenerator(GeneratorInit{Token: 30, RandomDigitsLength: 15})
	id, _ := gen.New("usr")
	gen2 := NewGenerator(GeneratorInit{Token: 30, RandomDigitsLength: 15})
	assert.Equal(t, true, gen2.Validate("usr", id))
}

func TestNewConcurrentGeneration(t *testing.T) {
	const numGoroutines = 150
	const numIDs = 2000

	gen := NewGenerator(GeneratorInit{Token: 30, RandomDigitsLength: 15})

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
