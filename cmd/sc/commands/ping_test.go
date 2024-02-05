package commands

import (
	"testing"
)

func Test_runPing(t *testing.T) {

	client, err := createClient(config)

	if err != nil {
		t.Fatalf("failed to create client: %s", err)
	}
	ping, err := callPingEndpoint(client)
	if err != nil {
		t.Fatalf("failed to run ping %s", err)
	}
	if !ping {
		t.Fatalf("ping failed")
	}
}
