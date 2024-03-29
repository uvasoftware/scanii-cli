package commands

import (
	"context"
	"testing"
	"time"
)

func TestAuthTokenLifeCycle(t *testing.T) {
	client, err := createClient(config)
	if err != nil {
		t.Fatalf("failed to create client: %s", err)
	}
	var tokenID string

	// create:
	targetTimeout := 10_000
	token, err := callCreateAuthToken(context.Background(), client, targetTimeout)
	if err != nil {
		t.Fatalf("failed to create token: %s", err)
	}
	expDate, err := time.Parse(time.RFC3339, *token.ExpirationDate)
	if err != nil {
		t.Fatalf("failed to parse expiration date: %s", err)
	}

	createDate, err := time.Parse(time.RFC3339, *token.CreationDate)
	if err != nil {
		t.Fatalf("failed to parse expiration date: %s", err)
	}

	if expDate.Sub(createDate).Seconds() != float64(targetTimeout) {
		t.Fatalf("expected expiration date to be %d seconds after creation date", targetTimeout)
	}

	tokenID = *token.Id

	t.Run("use", func(t *testing.T) {
		// create a config copy
		configCopy := *config
		configCopy.APIKey = tokenID
		configCopy.APISecret = ""

		client2, err := createClient(&configCopy)
		if err != nil {
			t.Fatalf("failed to create client2: %s", err)
		}
		// now try to use it
		result, err := callFileProcess(context.Background(), client2, "testdata/eicar.txt", 1, true, "", false)
		if err != nil {
			t.Fatalf("failed to process file: %s", err)
		}

		if result[0].id == "" {
			t.Fatalf("expected result to have an id")
		}

	})

	t.Run("delete", func(t *testing.T) {
		found, err := callDeleteAuthToken(context.Background(), client, tokenID)
		if err != nil {
			t.Fatalf("failed to delete token: %s", err)
		}
		if found != true {
			t.Fatalf("expected found to be true")
		}

		// now we try to use it again and it should fail
		configCopy := *config
		configCopy.APIKey = tokenID
		configCopy.APISecret = ""

		client2, err := createClient(&configCopy)
		if err != nil {
			t.Fatalf("failed to create client2: %s", err)
		}

		result, err := callFileProcess(context.Background(), client2, "testdata/eicar.txt", 1, true, "", false)
		if err != nil {
			t.Fatalf("failed to process file: %s", err)
		}

		if result[0].id != "" {
			t.Fatalf("we expected an invalid token error")
		}

	})

}
