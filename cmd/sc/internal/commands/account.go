package commands

import (
	"context"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
	"net/http"
	v22 "scanii-cli/internal/v22"
	"strings"
)

func AccountCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account",
		Short: "API operations for the account resource",
		Long:  `Account API operations. Detailed API documentation can be found here: https://uvasoftware.github.io/openapi/v22/#/General/account`,
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := loadConfig()
			if err != nil {
				return err
			}
			fmt.Printf("Using endpoint: %s and key: %s\n", config.Endpoint, config.APIKey)

			client, err := createClient(config)
			if err != nil {
				return err
			}

			pa, err := callAccountEndpoint(client)
			if err != nil {
				return err
			}

			if pa.Name != nil {
				fmt.Printf("Account: %s\n", *pa.Name)
			}
			if pa.Balance != nil {
				fmt.Printf("Balance: %s\n", humanize.Comma(int64(*pa.Balance)))
			}
			if pa.StartingBalance != nil {
				fmt.Printf("Starting Balance: %s\n", humanize.Comma(int64(*pa.StartingBalance)))
			}
			if pa.CreationDate != nil {
				fmt.Printf("Created: %s\n", *pa.CreationDate)
			}
			if pa.ModificationDate != nil {
				fmt.Printf("Modification Date: %s\n", *pa.CreationDate)
			}

			if pa.Users != nil && len(*pa.Users) > 0 {
				fmt.Printf("------\n")
				fmt.Printf("Users:\n")
				for k, v := range *pa.Users {
					fmt.Printf("  Email: %s\n", k)
					if v.CreationDate != nil {
						fmt.Printf("  Creation Date: %s\n", *v.CreationDate)
					}
					if v.LastLogin != nil {
						fmt.Printf("  Last Log in: %s\n", *v.LastLogin)
					}
				}
			}

			if pa.Keys != nil && len(*pa.Keys) > 0 {
				fmt.Printf("------\n")
				fmt.Printf("Keys:\n")
				for k, v := range *pa.Keys {
					fmt.Printf("  Key: %s\n", k)
					if v.CreationDate != nil {
						fmt.Printf("  Creation Date: %s\n", *v.CreationDate)
					}
					if v.DetectionCategoriesEnabled != nil {
						fmt.Printf("  Engines Enabled: %s\n", strings.Join(*v.DetectionCategoriesEnabled, ","))
					}
					if v.Tags != nil {
						fmt.Printf("  Tags: %s\n", strings.Join(*v.Tags, ","))
					}
					if v.LastSeenDate != nil {
						fmt.Printf("  Last Seen: %s\n", *v.LastSeenDate)
					}
					if v.Active != nil {
						fmt.Printf("  Active: %t\n", *v.Active)
					}

				}
			}

			return nil

		},
	}

	return cmd
}

func callAccountEndpoint(client *v22.Client) (*v22.AccountInfo, error) {
	r, err := client.Account(context.Background())
	if err != nil {
		return nil, err
	}

	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected http response status: %d", r.StatusCode)
	}

	defer r.Body.Close()
	parsedResponse, err := v22.ParseAccountResponse(r)
	if err != nil {
		return nil, err
	}

	return parsedResponse.JSON200, nil
}
