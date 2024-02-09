package commands

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	v22 "github.com/uvasoftware/scanii-cli/internal/v22"
	"log/slog"
	"net/url"
	"strconv"
	"strings"
)

func AuthTokenCommand(ctx context.Context) *cobra.Command {
	timeout := 300

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new authentication token",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := loadConfig()
			if err != nil {
				return err
			}
			client, err := createClient(config)
			if err != nil {
				return err
			}
			token, err := callCreateAuthToken(ctx, client, timeout)
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
		Short:      "Retrieves an existing authentication token",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := loadConfig()
			if err != nil {
				return err
			}
			client, err := createClient(config)
			if err != nil {
				return err
			}
			token, err := callRetrieveAuthToken(ctx, client, args[0])
			if err != nil {
				return err
			}

			printToken(token)
			return nil
		},
	}

	deleteCmd := &cobra.Command{
		Use:        "delete [id]",
		Short:      "Deletes an existing authentication token",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"id"},
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := loadConfig()
			if err != nil {
				return err
			}
			client, err := createClient(config)
			if err != nil {
				return err
			}
			_, err = callDeleteAuthToken(ctx, client, args[0])
			if err != nil {
				return err
			}

			fmt.Println("Token deleted")
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

func callDeleteAuthToken(ctx context.Context, client *v22.Client, s string) (bool, error) {
	httpResp, err := client.DeleteToken(ctx, s)
	if err != nil {
		return false, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != 204 {
		slog.Error("failed to delete token", "status", httpResp.Status)
		return false, nil
	}
	return true, nil

}

func callRetrieveAuthToken(ctx context.Context, client *v22.Client, s string) (*v22.AuthToken, error) {
	httpResp, err := client.RetrieveToken(ctx, s)

	if err != nil {
		return nil, err
	}

	defer httpResp.Body.Close()

	if httpResp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to create token: %s", httpResp.Status)
	}

	token, err := v22.ParseRetrieveTokenResponse(httpResp)
	if err != nil {
		return nil, err
	}
	return token.JSON200, nil
}

func callCreateAuthToken(ctx context.Context, client *v22.Client, timeoutInSeconds int) (*v22.AuthToken, error) {
	form := url.Values{}
	form.Add("timeout", strconv.Itoa(timeoutInSeconds))
	httpResp, err := client.CreateTokenWithBody(ctx, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to create token: %s", httpResp.Status)
	}

	token, err := v22.ParseCreateTokenResponse(httpResp)
	if err != nil {
		return nil, err
	}
	return token.JSON200, nil

}
func printToken(token *v22.AuthToken) {
	fmt.Println("Token ID:", *token.Id)
	fmt.Println("Expiration Date:", *token.ExpirationDate)
	fmt.Println("Creation Date:", *token.CreationDate)
}
