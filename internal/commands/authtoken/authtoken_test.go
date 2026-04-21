package authtoken

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/uvasoftware/scanii-cli/internal/testutil"
)

var ts *testutil.Server

func init() {
	ts = testutil.StartServer()
}

func TestAuthTokenLifeCycle(t *testing.T) {
	c, err := ts.Profile.Client()
	if err != nil {
		t.Fatalf("failed to create client: %s", err)
	}
	var tokenID string

	// create:
	targetTimeout := 10_000
	token, err := callCreateAuthToken(context.Background(), c, targetTimeout)
	if err != nil {
		t.Fatalf("failed to create token: %s", err)
	}
	expDate, err := time.Parse(time.RFC3339, *token.ExpirationDate)
	if err != nil {
		t.Fatalf("failed to parse expiration date: %s", err)
	}

	createDate, err := time.Parse(time.RFC3339, *token.CreationDate)
	if err != nil {
		t.Fatalf("failed to parse creation date: %s", err)
	}

	if expDate.Sub(createDate).Seconds() != float64(targetTimeout) {
		t.Fatalf("expected expiration date to be %d seconds after creation date", targetTimeout)
	}

	tokenID = *token.Id

	t.Run("retrieve", func(t *testing.T) {
		retrieved, err := callRetrieveAuthToken(context.Background(), c, tokenID)
		if err != nil {
			t.Fatalf("failed to retrieve token: %s", err)
		}
		if *retrieved.Id != tokenID {
			t.Fatalf("expected token id %s, got %s", tokenID, *retrieved.Id)
		}
	})

	t.Run("use", func(t *testing.T) {
		// verify the token works by pinging
		configCopy := *ts.Profile
		configCopy.Credentials = tokenID + ":"

		client2, err := configCopy.Client()
		if err != nil {
			t.Fatalf("failed to create client2: %s", err)
		}

		result, err := client2.Ping(context.Background())
		if err != nil {
			t.Fatalf("failed to ping with token: %s", err)
		}

		if result.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", result.StatusCode)
		}

		if result.Message != "pong" {
			t.Fatalf("expected pong, got %s", result.Message)
		}
	})

	t.Run("delete", func(t *testing.T) {
		found, err := callDeleteAuthToken(context.Background(), c, tokenID)
		if err != nil {
			t.Fatalf("failed to delete token: %s", err)
		}
		if found != true {
			t.Fatalf("expected found to be true")
		}

		// now we try to use it again and it should fail
		configCopy := *ts.Profile
		configCopy.Credentials = tokenID + ":"

		client2, err := configCopy.Client()
		if err != nil {
			t.Fatalf("failed to create client2: %s", err)
		}

		result, err := client2.Ping(context.Background())
		if err != nil {
			t.Fatalf("failed to ping: %s", err)
		}

		if result.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected status 401 with deleted token, got %d", result.StatusCode)
		}
	})
}
