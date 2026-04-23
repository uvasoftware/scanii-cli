package engine

import (
	"io"
	"os"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	engine, err := New()
	if err != nil {
		t.Fatal(err.Error())
	}
	if len(engine.config.Rules) == 0 {
		t.Fatalf("ruleset was not loaded")
	}
}

func TestRuleCount(t *testing.T) {
	engine, err := New()
	if err != nil {
		t.Fatal(err.Error())
	}
	if engine.RuleCount() == 0 {
		t.Fatalf("expected non-zero rule count")
	}
	if engine.RuleCount() != len(engine.config.Rules) {
		t.Fatalf("RuleCount() mismatch: got %d, want %d", engine.RuleCount(), len(engine.config.Rules))
	}
}

func TestLoadConfigCustomRules(t *testing.T) {
	engine := &Engine{config: &Config{Rules: make([]Rule, 0)}}
	config := `{"rules": [{"format": "sha1", "content": "abc123", "result": "test.finding"}]}`
	err := engine.LoadConfig(strings.NewReader(config))
	if err != nil {
		t.Fatal(err.Error())
	}
	if engine.RuleCount() != 1 {
		t.Fatalf("expected 1 rule, got %d", engine.RuleCount())
	}
	if engine.config.Rules[0].Result != "test.finding" {
		t.Fatalf("expected rule result test.finding, got %s", engine.config.Rules[0].Result)
	}
}

func TestLoadConfigInvalidJSON(t *testing.T) {
	engine := &Engine{config: &Config{Rules: make([]Rule, 0)}}
	err := engine.LoadConfig(strings.NewReader("not json"))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestProcessEmptyInput(t *testing.T) {
	engine, err := New()
	if err != nil {
		t.Fatal(err.Error())
	}
	result, err := engine.Process(strings.NewReader(""))
	if err != nil {
		t.Fatal(err.Error())
	}
	if result.ContentLength != 0 {
		t.Fatalf("expected content length 0, got %d", result.ContentLength)
	}
	if len(result.Findings) != 0 {
		t.Fatalf("expected no findings, got %v", result.Findings)
	}
}

func TestIdentifyEicar(t *testing.T) {
	engine, err := New()
	if err != nil {
		t.Fatal(err.Error())
	}

	result, err := engine.Process(strings.NewReader("X5O!P%@AP[4\\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*"))
	if err != nil {
		return
	}

	if result.ContentLength == 0 {
		t.Fatalf("content length was not calculated")
	}

	if result.Findings[0] != "content.malicious.eicar-test-signature" {
		t.Fatalf("eicar file was not identified")
	}

	if result.ContentType != "text/plain; charset=utf-8" {
		t.Fatalf("content type was not identifiedd, got %s", result.ContentType)
	}

}

type junkReader struct {
	length    uint64
	readIndex uint64
}

func (r *junkReader) Read(p []byte) (n int, err error) {
	if r.readIndex >= r.length {
		return 0, io.EOF
	}
	for i := range p {
		p[i] = 7
	}

	r.readIndex += uint64(len(p))
	return len(p), nil
}

func TestEngine_Process(t *testing.T) {
	r := &junkReader{
		length: 11 * 1024 * 1024,
	}

	engine, err := New()
	if err != nil {
		t.Fatal(err.Error())
	}

	_, err = engine.Process(r)
	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestSamples(t *testing.T) {
	tests := []struct {
		file            string
		expectedFinding string
	}{
		{"testdata/image.jpg", "content.image.nsfw.nudity"},
		{"testdata/language.txt", "content.en.language.nsfw.0"},
	}
	for _, test := range tests {
		t.Run(test.file, func(t *testing.T) {
			if open, err := os.Open(test.file); err != nil {
				t.Fatalf("failed to open file: %s", err)
			} else {
				defer open.Close()
				engine, err := New()
				if err != nil {
					t.Fatal(err.Error())
				}
				result, err := engine.Process(open)
				if err != nil {
					t.Fatal(err.Error())
				}
				if len(result.Findings) == 0 {
					t.Fatalf("expected findings, got none")
				}
				if result.Findings[0] != test.expectedFinding {
					t.Fatalf("expected finding %s, got %s", test.expectedFinding, result.Findings[0])
				}
			}

		})
	}
}

func BenchmarkEngine(b *testing.B) {
	// 11 mb of junk
	r := &junkReader{
		length: 10 * 1024 * 1024,
	}
	engine, err := New()
	if err != nil {
		b.Fatal(err.Error())
	}

	for i := 0; i < b.N; i++ {
		_, err = engine.Process(r)
		if err != nil {
			b.Fatal(err.Error())
		}
	}
}
