package account

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
	"github.com/uvasoftware/scanii-cli/internal/client"
	"github.com/uvasoftware/scanii-cli/internal/commands/profile"
	"github.com/uvasoftware/scanii-cli/internal/terminal"
)

// Command returns the account cobra command.
func Command(ctx context.Context, profileName *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account",
		Short: "API operations for the account resource",
		Long:  `Account API operations. Detailed API documentation can be found here: https://uvasoftware.github.io/openapi/v22/#/General/account`,
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := profile.Load(*profileName)
			if err != nil {
				return err
			}
			terminal.Info(fmt.Sprintf("Using endpoint: %s and API key: %s", config.Endpoint, config.APIKey()))

			c, err := config.Client()
			if err != nil {
				return err
			}

			pa, err := callAccountEndpoint(ctx, c)
			if err != nil {
				return err
			}

			terminal.Section("Account information")
			if pa.Name != nil {
				terminal.KeyValue("Name:", *pa.Name)
			}
			if pa.Balance != nil {
				if pa.StartingBalance != nil {
					terminal.KeyValue("Balance:", fmt.Sprintf("%s/%s", terminal.FormatNumber(int64(*pa.Balance)), terminal.FormatNumber(int64(*pa.StartingBalance))))
				} else {
					terminal.KeyValue("Balance:", terminal.FormatNumber(int64(*pa.Balance)))
				}
			}
			if pa.CreationDate != nil {
				terminal.KeyValue("Created:", terminal.FormatTime(*pa.CreationDate))
			}
			if pa.ModificationDate != nil {
				terminal.KeyValue("Modified:", terminal.FormatTime(*pa.ModificationDate))
			}

			if pa.Keys != nil && len(*pa.Keys) > 0 {
				terminal.Section("API keys")
				for k, v := range *pa.Keys {
					terminal.KeyValue("Key:", k)
					if v.Active != nil {
						terminal.KeyValue("Active:", fmt.Sprintf("%t", *v.Active))
					}
					if v.Tags != nil {
						terminal.KeyValue("Tags:", strings.Join(*v.Tags, ", "))
					}
					if v.DetectionCategoriesEnabled != nil {
						terminal.KeyValueW("Engines enabled:", strings.Join(*v.DetectionCategoriesEnabled, ", "), 18)
					}
					if v.CreationDate != nil {
						terminal.KeyValue("Created:", terminal.FormatTime(*v.CreationDate))
					}
					if v.LastSeenDate != nil {
						terminal.KeyValue("Last seen:", terminal.FormatTime(*v.LastSeenDate))
					}
				}
			}

			if pa.Users != nil && len(*pa.Users) > 0 {
				terminal.Section("Users")
				for k, v := range *pa.Users {
					terminal.KeyValue("Email:", k)
					if v.CreationDate != nil {
						terminal.KeyValue("Created:", terminal.FormatTime(*v.CreationDate))
					}
					if v.LastLogin != nil {
						terminal.KeyValue("Last login:", terminal.FormatTime(*v.LastLogin))
					}
				}
			}

			return nil
		},
	}

	return cmd
}

func callAccountEndpoint(ctx context.Context, c *client.Client) (*client.AccountInfo, error) {
	result, err := c.Account(ctx)
	if err != nil {
		return nil, err
	}

	if result.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected http response status: %d", result.StatusCode)
	}

	return result.Account, nil
}
