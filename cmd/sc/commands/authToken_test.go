package commands

import (
	"testing"
	"time"
)

func TestAuthTokenLifeCycle(t *testing.T) {
	client, err := createClient(config)
	if err != nil {
		t.Fatalf("failed to create client: %s", err)
	}
	var tokenId string

	// create:
	targetTimeout := 10_000
	token, err := callCreateAuthToken(client, targetTimeout)
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

	tokenId = *token.Id

	t.Run("use", func(t *testing.T) {
		// create a config copy
		configCopy := *config
		configCopy.ApiKey = tokenId
		configCopy.ApiSecret = ""

		client2, err := createClient(&configCopy)
		if err != nil {
			t.Fatalf("failed to create client2: %s", err)
		}
		// now try to use it
		result, err := callFileProcess(client2, "testdata/eicar.txt", 1, true, "", false)
		if err != nil {
			t.Fatalf("failed to process file: %s", err)
		}

		if result[0].id == "" {
			t.Fatalf("expected result to have an id")
		}

	})

	t.Run("delete", func(t *testing.T) {
		found, err := callDeleteAuthToken(client, tokenId)
		if err != nil {
			t.Fatalf("failed to delete token: %s", err)
		}
		if found != true {
			t.Fatalf("expected found to be true")
		}

		// now we try to use it again and it should fail
		configCopy := *config
		configCopy.ApiKey = tokenId
		configCopy.ApiSecret = ""

		client2, err := createClient(&configCopy)
		result, err := callFileProcess(client2, "testdata/eicar.txt", 1, true, "", false)
		if err != nil {
			t.Fatalf("failed to process file: %s", err)
		}

		if result[0].id != "" {
			t.Fatalf("we expected an invalid token error")
		}

	})

}
