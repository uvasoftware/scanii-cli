package account

import (
	"context"
	"testing"

	"github.com/uvasoftware/scanii-cli/internal/testutil"
)

var ts *testutil.Server

func init() {
	ts = testutil.StartServer()
}

func TestShouldCallAccountEndpoint(t *testing.T) {
	client, err := ts.Profile.Client()
	if err != nil {
		t.Fatalf("failed to create client: %s", err)
	}

	pa, err := callAccountEndpoint(context.Background(), client)
	if err != nil {
		t.Fatalf("failed to call account endpoint: %s", err)
	}

	if pa.Name == nil {
		t.Fatalf("expected result to have a name")
	}
	if pa.Balance == nil {
		t.Fatalf("expected result to have a balance")
	}
	if pa.Users == nil {
		t.Fatalf("expected result to have users")
	}
	if pa.Keys == nil {
		t.Fatalf("expected result to have keys")
	}
}
