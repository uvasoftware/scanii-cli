package engine

import (
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
		t.Fatalf("content lenght was not calculated")
	}

	if result.Findings[0] != "content.malicious.eicar-test-signature" {
		t.Fatalf("eicar file was not identified")
	}

}
