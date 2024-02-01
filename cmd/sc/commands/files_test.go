package commands

import (
	"fmt"
	"github.com/lmittmann/tint"
	"log/slog"
	"os"
	"testing"
)

func init() {
	handler := tint.NewHandler(os.Stdout, &tint.Options{
		AddSource:  true,
		Level:      slog.LevelDebug,
		TimeFormat: "2006/01/02 15:04",
		NoColor:    false,
	})
	slog.SetDefault(slog.New(handler))
}

const eicarSample = "testdata/eicar.txt"

func Test_ShouldProcessSync(t *testing.T) {

	client, err := createClient(config)

	if err != nil {
		t.Fatalf("failed to create client: %s", err)
	}

	results, err := runFileProcess(client, eicarSample, 1, false, "m1=v1", false)
	if err != nil {
		t.Fatalf("failed to process file: %s", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	assertEicar(t, &results[0])

}

func Test_ShouldProcessLocationSync(t *testing.T) {

	client, err := createClient(config)

	if err != nil {
		t.Fatalf("failed to create client: %s", err)
	}

	result, err := runLocationProcess(client, fmt.Sprintf("http://%s/static/eicar.txt", endpoint), "", "m1=v1")
	if err != nil {
		t.Fatalf("failed to process file: %s", err)
	}

	assertEicar(t, result)

}

// assertEicar checks that the result is the expected eicar test file
// https://en.wikipedia.org/wiki/EICAR_test_file
func assertEicar(t *testing.T, result *resultRecord) {
	// check result has findings:
	if len(result.findings) == 0 {
		t.Fatalf("expected findings, got none")
	}

	if result.findings[0] != "content.malicious.eicar-test-signature" {
		t.Fatalf("expected finding content.malicious.eicar-test-signature, got %s", result.findings[0])
	}

	if result.checksum != "3395856ce81f2b7382dee72602f798b642f14140" {
		t.Fatalf("expected checksum cf8bd9dfddff007f75adf4c2be48005cea317c62, got %s", result.checksum)
	}

	if result.contentLength != 68 {
		t.Fatalf("expected content length 68, got %d", result.contentLength)
	}

	// check result has metadata:
	if result.metadata["m1"] != "v1" {
		t.Fatalf("expected metadata m1=v1, got %s", result.metadata["m1"])
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

func Test_shouldRetrievePreviousFiles(t *testing.T) {
	client, err := createClient(config)

	if err != nil {
		t.Fatalf("failed to create client: %s", err)
	}

	results, err := runFileProcess(client, "testdata/eicar.txt", 1, false, "m1=v1", true)
	if err != nil {
		t.Fatalf("failed to process file: %s", err)
	}

	t.Run("valid_id", func(t *testing.T) {
		retrieve, err := runFileRetrieve(client, results[0].id)
		if err != nil {
			t.Fatalf("failed to retrieve file: %s", err)
		}
		assertEicar(t, retrieve)

	})

}
