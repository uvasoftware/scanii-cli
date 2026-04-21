package ping

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/uvasoftware/scanii-cli/internal/commands/profile"
	"github.com/uvasoftware/scanii-cli/internal/terminal"
)

// Command returns the ping cobra command.
func Command(ctx context.Context, profileName *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ping",
		Short: "API operations for the ping resource",
		Long:  `Ping API operation. Detailed API documentation can be found here: https://uvasoftware.github.io/openapi/v22/#/General/ping`,
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := profile.Load(*profileName)
			if err != nil {
				return err
			}
			terminal.Info(fmt.Sprintf("Using endpoint: %s and API key: %s", config.Endpoint, config.ApiKey()))

			client, err := config.Client()
			if err != nil {
				return err
			}
			startTime := time.Now()
			_, err = client.Ping(ctx)
			if err != nil {
				return err
			}
			terminal.Success(fmt.Sprintf("Ping successful in %s using endpoint %s", terminal.FormatDuration(time.Since(startTime)), config.Endpoint))
			return nil
		},
	}
	return cmd
}
