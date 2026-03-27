package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

const defaultProfileName = "default"

func ProfileCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage CLI profiles",
		Long:  "Manage multiple CLI profiles with different API credentials and endpoints",
	}

	cmd.AddCommand(profileCreateCommand())
	cmd.AddCommand(profileListCommand())
	cmd.AddCommand(profileDeleteCommand())

	return cmd
}

func profileCreateCommand() *cobra.Command {
	var endpoint, credentials string

	cmd := &cobra.Command{
		Use:   "create [name]",
		Short: "Create or update a profile",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := ""
			if len(args) > 0 {
				name = args[0]
			}

			config := &configuration{}

			if endpoint == "" && credentials == "" {
				// interactive mode: prompt for name if not provided
				if name == "" {
					namePrompt := promptui.Prompt{
						Label:   "Profile name",
						Default: defaultProfileName,
					}
					result, err := namePrompt.Run()
					if err != nil {
						return err
					}
					name = result
				}

				prompt := promptui.Prompt{
					Label: "API Credentials in the format key:secret",
					Validate: func(s string) error {
						if strings.Trim(s, " ") == "" {
							return fmt.Errorf("credentials cannot be empty")
						}
						if !strings.Contains(s, ":") {
							return fmt.Errorf("credentials must be in the format key:secret")
						}
						return nil
					},
				}

				result, _ := prompt.Run()
				config.Credentials = result

				endpointSelect := promptui.Select{
					Label: "Select endpoint to use",
					Items: []string{"api-us1.scanii.com", "api-eu1.scanii.com", "api-eu2.scanii.com", "api-ap1.scanii.com", "api-ap2.scanii.com", "api-ca1.scanii.com", "localhost:4000"},
				}
				_, selectedEndpoint, err := endpointSelect.Run()
				if err != nil {
					return err
				}
				config.Endpoint = selectedEndpoint
			} else {
				// non-interactive mode: default to "default" if no name provided
				if name == "" {
					name = defaultProfileName
				}
				if endpoint != "" {
					config.Endpoint = endpoint
				}
				if credentials != "" {
					config.Credentials = credentials
				}
			}

			err := saveConfig(name, config)
			if err != nil {
				return err
			}
			fmt.Printf("Profile %q saved\n", name)
			return nil
		},
	}

	cmd.Flags().StringVarP(&endpoint, "endpoint", "e", "", "Endpoint to use, see https://docs.scanii.com/article/161-endpoints-and-regions")
	cmd.Flags().StringVarP(&credentials, "credentials", "c", "", "API credentials to use in the format key:secret")

	return cmd
}

func profileListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list [name]",
		Short: "List all profiles or show details for a specific profile",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				name := args[0]
				config, err := loadConfig(name)
				if err != nil {
					return fmt.Errorf("profile %q not found: %w", name, err)
				}

				fmt.Printf("Profile: %s\n", name)
				fmt.Printf("  Endpoint:     %s\n", config.Endpoint)
				fmt.Printf("  Credentials:  %s\n", config.Credentials)
				fmt.Printf("  Created At:   %s\n", config.CreatedAt.Format(time.RFC3339))
				return nil
			}

			dir, err := configDir()
			if err != nil {
				return err
			}

			entries, err := os.ReadDir(dir)
			if err != nil {
				if os.IsNotExist(err) {
					fmt.Println("No profiles found")
					return nil
				}
				return err
			}

			found := false
			for _, entry := range entries {
				if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
					continue
				}
				name := strings.TrimSuffix(entry.Name(), ".json")
				config, err := loadConfig(name)
				if err != nil {
					fmt.Printf("  %s (error: %s)\n", name, err)
					continue
				}
				fmt.Printf("  %-20s endpoint=%s key=%s\n", name, config.Endpoint, config.apiKey())
				found = true
			}

			if !found {
				fmt.Println("No profiles found")
			}
			return nil
		},
	}
}

func profileDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			p, err := configPath(name)
			if err != nil {
				return err
			}

			if _, err := os.Stat(p); os.IsNotExist(err) {
				return fmt.Errorf("profile %q not found", name)
			}

			err = os.Remove(p)
			if err != nil {
				return err
			}

			fmt.Printf("Profile %q deleted\n", name)
			return nil
		},
	}
}

type configuration struct {
	Endpoint    string    `json:"endpoint"`
	CreatedAt   time.Time `json:"createdAt"`
	Version     *string   `json:"version"`
	Credentials string    `json:"credentials"`
}

// apiKey returns the key portion of the credentials (before the colon).
func (c *configuration) apiKey() string {
	parts := strings.SplitN(c.Credentials, ":", 2)
	return parts[0]
}

// apiSecret returns the secret portion of the credentials (after the colon).
func (c *configuration) apiSecret() string {
	parts := strings.SplitN(c.Credentials, ":", 2)
	if len(parts) < 2 {
		return ""
	}
	return parts[1]
}

// configDir returns the base directory for storing profiles.
// Uses $HOME/.config/scanii-cli on all platforms.
func configDir() (string, error) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dirname, ".config", "scanii-cli"), nil
}

func configPath(profileName string) (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, profileName+".json"), nil
}

func ensureConfigDir() error {
	dir, err := configDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(dir, 0755)
}

func loadConfig(profileName string) (*configuration, error) {
	config := &configuration{}

	cp, err := configPath(profileName)
	if err != nil {
		return nil, err
	}

	slog.Debug("loading config", "path", cp, "profile", profileName)

	fd, err := os.Open(cp)
	if err != nil {
		if os.IsNotExist(err) {
			if profileName == defaultProfileName {
				return nil, fmt.Errorf("no default profile configured, create one with: sc profile create")
			}
			return nil, fmt.Errorf("profile %q not found, create one with: sc profile create %s", profileName, profileName)
		}
		return nil, err
	}
	defer fd.Close()

	contents, err := io.ReadAll(fd)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(contents, &config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func saveConfig(profileName string, config *configuration) error {
	config.CreatedAt = time.Now()

	if err := ensureConfigDir(); err != nil {
		return err
	}

	cp, err := configPath(profileName)
	if err != nil {
		return err
	}

	js, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(cp, js, 0600)
	if err != nil {
		return err
	}

	slog.Debug("config saved", "path", cp, "profile", profileName)
	return nil
}
