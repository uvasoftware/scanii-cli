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
		Short: "Account related operations",
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

			httpResponse, err := client.Account(context.Background())
			if err != nil {
				return err
			}

			if httpResponse.StatusCode != http.StatusOK {
				return fmt.Errorf("unexpected parsedResponse response status: %d", httpResponse.StatusCode)
			}

			parsedAccount, err := v22.ParseAccountResponse(httpResponse)
			if err != nil {
				return err
			}

			pa := parsedAccount.JSON200

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
