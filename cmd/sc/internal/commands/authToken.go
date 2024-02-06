package commands

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"log/slog"
	"net/url"
	v22 "scanii-cli/internal/v22"
	"strconv"
	"strings"
)

func AuthTokenCommand() *cobra.Command {
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
			_, err = callCreateAuthToken(client, 3600)
			if err != nil {
				return err
			}

			return nil
		},
	}

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
			_, err = callRetrieveAuthToken(client, args[0])
			if err != nil {
				return err
			}

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
			_, err = callDeleteAuthToken(client, args[0])
			if err != nil {
				return err
			}

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

func callDeleteAuthToken(client *v22.Client, s string) (bool, error) {
	httpResp, err := client.DeleteToken(context.Background(), s)
	if err != nil {
		return false, err
	}
	if httpResp.StatusCode != 204 {
		slog.Error("failed to delete token", "status", httpResp.Status)
		return false, nil
	}
	return true, nil

}

func callRetrieveAuthToken(client *v22.Client, s string) (*v22.AuthToken, error) {
	httpResp, err := client.RetrieveToken(context.Background(), s)
	if err != nil {
		return nil, err
	}
	if httpResp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to create token: %s", httpResp.Status)
	}

	token, err := v22.ParseRetrieveTokenResponse(httpResp)
	if err != nil {
		return nil, err
	}
	return token.JSON200, nil
}

func callCreateAuthToken(client *v22.Client, timeoutInSeconds int) (*v22.AuthToken, error) {
	form := url.Values{}
	form.Add("timeout", strconv.Itoa(timeoutInSeconds))
	httpResp, err := client.CreateTokenWithBody(context.Background(), "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	if httpResp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to create token: %s", httpResp.Status)
	}

	token, err := v22.ParseCreateTokenResponse(httpResp)
	if err != nil {
		return nil, err
	}
	return token.JSON200, nil

}
