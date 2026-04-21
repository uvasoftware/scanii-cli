package ping

import (
	"context"
	"net/http"
	"testing"

	"github.com/uvasoftware/scanii-cli/internal/testutil"
)

var ts *testutil.Server

func init() {
	ts = testutil.StartServer()
}

func TestPing(t *testing.T) {
	c, err := ts.Profile.Client()
	if err != nil {
		t.Fatalf("failed to create client: %s", err)
	}
	result, err := c.Ping(context.Background())
	if err != nil {
		t.Fatalf("failed to run ping: %s", err)
	}

	if result.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", result.StatusCode)
	}

	if result.Message != "pong" {
		t.Fatalf("expected pong, got %s", result.Message)
	}
}
