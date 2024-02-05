package commands

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"log/slog"
	"net/http"
	v22 "scanii-cli/internal/v22"
	"time"
)

func PingCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ping",
		Short: "Ping the Scanii API",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := loadConfig()
			if err != nil {
				return err
			}
			fmt.Printf("Using endpoint: %s and key: %s\n", config.Endpoint, config.ApiKey)

			client, err := createClient(config)
			if err != nil {
				return err
			}
			_, err = callPingEndpoint(client)
			if err != nil {
				return err
			}
			return nil
		},
	}
	return cmd
}

func callPingEndpoint(client *v22.Client) (bool, error) {
	startTime := time.Now()

	r, err := client.Ping(context.Background())
	if err != nil {
		return false, err
	}

	if r.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected parsedResponse response status: %d", r.StatusCode)
	}

	parsedResponse, err := v22.ParsePingResponse(r)
	if err != nil {
		return false, err
	}
	slog.Debug("parsedResponse response", "status", parsedResponse.Status(), "body", string(parsedResponse.Body))

	if *parsedResponse.JSON200.Message != "pong" {
		return false, fmt.Errorf("unexpected parsedResponse response: %s", *parsedResponse.JSON200.Message)
	}
	fmt.Printf("success in %s\n", time.Since(startTime).String())

	return true, nil
}
