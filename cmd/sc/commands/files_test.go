package commands

import (
	"testing"
)

func Test_ShouldProcessSync(t *testing.T) {

	client, err := createClient(config)

	if err != nil {
		t.Fatalf("failed to create client: %s", err)
	}

	results, err := runFileProcess(client, "testdata/eicar.txt", 1, false, "m1=v1", false)
	if err != nil {
		t.Fatalf("failed to process file: %s", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	// check result has findings:
	if len(results[0].findings) == 0 {
		t.Fatalf("expected findings, got none")
	}

	if results[0].findings[0] != "content.malicious.eicar-test-signature" {
		t.Fatalf("expected finding content.malicious.eicar-test-signature, got %s", results[0].findings[0])
	}

	if results[0].checksum != "cf8bd9dfddff007f75adf4c2be48005cea317c62" {
		t.Fatalf("expected checksum cf8bd9dfddff007f75adf4c2be48005cea317c62, got %s", results[0].checksum)
	}

	// check result has metadata:
	if results[0].metadata["m1"] != "v1" {
		t.Fatalf("expected metadata m1=v1, got %s", results[0].metadata["m1"])
	}

}

func Test_ShouldProcessAsync(t *testing.T) {

	client, err := createClient(config)

	if err != nil {
		t.Fatalf("failed to create client: %s", err)
	}

	results, err := runFileProcess(client, "testdata/eicar.txt", 1, false, "m1=v1", true)
	if err != nil {
		t.Fatalf("failed to process file: %s", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	// ensure pending result has an id and a location;
	if results[0].id == "" {
		t.Fatalf("expected result to have an id")
	}
	if results[0].location == "" {
		t.Fatalf("expected result to have a location")
	}

}

func Test_ShouldProcessFetch(t *testing.T) {

	client, err := createClient(config)

	if err != nil {
		t.Fatalf("failed to create client: %s", err)
	}

	result, err := runFetch(client, "https://www.google.com", "", "m1=v1")
	if err != nil {
		t.Fatalf("failed to process file: %s", err)
	}

	// ensure pending result has an id and a location;
	if result.id == "" {
		t.Fatalf("expected result to have an id")
	}
	if result.location == "" {
		t.Fatalf("expected result to have a location")
	}

}
