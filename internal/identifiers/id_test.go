//go:build !integration

package identifiers

import (
	"strings"
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
		r := <-results
		if r == "" {
			t.Fatalf("empty result")
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
		r := <-results
		if r == "" {
			t.Fatalf("empty result")
		}
	}
}

func TestGenerateLength(t *testing.T) {
	id := Generate()
	if len(id) != 32 {
		t.Fatalf("expected length 32, got %d", len(id))
	}
}

func TestGenerateShortLength(t *testing.T) {
	id := GenerateShort()
	if len(id) != 16 {
		t.Fatalf("expected length 16, got %d", len(id))
	}
}

func TestGenerateSecure(t *testing.T) {
	const iterations = 100
	results := make(chan string, iterations)

	for i := 0; i < iterations; i++ {
		go func() {
			results <- GenerateSecure()
		}()
	}

	seen := make(map[string]bool)
	for i := 0; i < iterations; i++ {
		r := <-results
		if r == "" {
			t.Fatalf("empty result")
		}
		if len(r) != 32 {
			t.Fatalf("expected length 32, got %d", len(r))
		}
		if seen[r] {
			t.Fatalf("duplicate result: %s", r)
		}
		seen[r] = true
	}
}

func TestGenerateShortSecure(t *testing.T) {
	const iterations = 100
	results := make(chan string, iterations)

	for i := 0; i < iterations; i++ {
		go func() {
			results <- GenerateShortSecure()
		}()
	}

	for i := 0; i < iterations; i++ {
		r := <-results
		if r == "" {
			t.Fatalf("empty result")
		}
		if len(r) != 16 {
			t.Fatalf("expected length 16, got %d", len(r))
		}
	}
}

func TestGenerateCharset(t *testing.T) {
	id := Generate()
	for _, c := range id {
		if !strings.ContainsRune(runes, c) {
			t.Fatalf("unexpected character %c in generated id", c)
		}
	}
}

func TestGenerateSecureCharset(t *testing.T) {
	id := GenerateSecure()
	for _, c := range id {
		if !strings.ContainsRune(runes, c) {
			t.Fatalf("unexpected character %c in generated id", c)
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
