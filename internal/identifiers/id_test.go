//go:build !integration

package identifiers

import (
	"testing"
)

func TestGenerate(t *testing.T) {
	// attempting concurrent generation
	const iterations = 100
	results := make(chan string, iterations)

	for i := 0; i < iterations; i++ {
		go func() {
			results <- Generate()
		}()
	}

	for i := 0; i < iterations; i++ {
		select {
		case r := <-results:
			if len(r) == 0 {
				t.Fatalf("empty result")
			}
		}

	}
}
func TestGenerateShort(t *testing.T) {
	// attempting concurrent generation
	const iterations = 100
	results := make(chan string, iterations)

	for i := 0; i < iterations; i++ {
		go func() {
			results <- GenerateShort()
		}()
	}

	for i := 0; i < iterations; i++ {
		select {
		case r := <-results:
			if len(r) == 0 {
				t.Fatalf("empty result")
			}
		}

	}
}

func BenchmarkGenerate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Generate()
	}
}

func BenchmarkGenerateSecure(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateSecure()
	}
}
func BenchmarkGenerateShort(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateShort()
	}
}
