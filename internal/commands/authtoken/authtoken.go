package authtoken

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/uvasoftware/scanii-cli/internal/client"
	"github.com/uvasoftware/scanii-cli/internal/commands/profile"
	"github.com/uvasoftware/scanii-cli/internal/terminal"
)

// Command returns the auth-token cobra command.
func Command(ctx context.Context, profileName *string) *cobra.Command {
	timeout := 300

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new authentication token",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := profile.Load(*profileName)
			if err != nil {
				return err
			}
			c, err := config.Client()
			if err != nil {
				return err
			}
			token, err := callCreateAuthToken(ctx, c, timeout)
			if err != nil {
				return err
			}

			printToken(token)
			return nil
		},
	}

	createCmd.PersistentFlags().IntVarP(&timeout, "timeout", "t", 300, "Timeout for created token in seconds")

	retrieveCmd := &cobra.Command{
		Use:        "retrieve [id]",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"id"},
		Short:      "Retrieve an existing authentication token",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := profile.Load(*profileName)
			if err != nil {
				return err
			}
			c, err := config.Client()
			if err != nil {
				return err
			}
			token, err := callRetrieveAuthToken(ctx, c, args[0])
			if err != nil {
				return err
			}

			printToken(token)
			return nil
		},
	}

	deleteCmd := &cobra.Command{
		Use:        "delete [id]",
		Short:      "Delete an existing authentication token",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"id"},
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := profile.Load(*profileName)
			if err != nil {
				return err
			}
			c, err := config.Client()
			if err != nil {
				return err
			}
			_, err = callDeleteAuthToken(ctx, c, args[0])
			if err != nil {
				return err
			}

			terminal.Success(fmt.Sprintf("Token %s deleted", args[0]))
			return nil
		},
	}

	cmd := &cobra.Command{
		Use:   "auth-token",
		Short: "API operations for the authentication token resource",
		Long:  `Auth Token API operations. Detailed API documentation can be found here: https://uvasoftware.github.io/openapi/v22/#/Authentication%20Token`,
	}

	cmd.AddCommand(createCmd)
	cmd.AddCommand(retrieveCmd)
	cmd.AddCommand(deleteCmd)
	return cmd
}

func callDeleteAuthToken(ctx context.Context, c *client.Client, s string) (bool, error) {
	result, err := c.DeleteToken(ctx, s)
	if err != nil {
		return false, err
	}

	if result.StatusCode != 204 {
		slog.Error("failed to delete token", "status", result.StatusCode)
		return false, nil
	}
	return true, nil
}

func callRetrieveAuthToken(ctx context.Context, c *client.Client, s string) (*client.AuthToken, error) {
	result, err := c.RetrieveToken(ctx, s)
	if err != nil {
		return nil, err
	}

	if result.StatusCode != 200 {
		return nil, fmt.Errorf("failed to retrieve token: %d", result.StatusCode)
	}

	return result.Token, nil
}

func callCreateAuthToken(ctx context.Context, c *client.Client, timeoutInSeconds int) (*client.AuthToken, error) {
	form := url.Values{}
	form.Add("timeout", strconv.Itoa(timeoutInSeconds))
	result, err := c.CreateToken(ctx, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}

	if result.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to create token: %d", result.StatusCode)
	}

	return result.Token, nil
}

func printToken(token *client.AuthToken) {
	terminal.KeyValue("id:", *token.Id)
	terminal.KeyValue("created:", terminal.FormatTime(*token.CreationDate))
	terminal.KeyValue("expires:", terminal.FormatTime(*token.ExpirationDate))
}
