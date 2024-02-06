package commands

import (
	"encoding/json"
	"fmt"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"io"
	"log/slog"
	"os"
	"path"
	"strings"
	"time"
)

func ConfigureCommand() *cobra.Command {

	var (
		endpoint, credentials string
	)

	cmd := &cobra.Command{
		Use:   "configure",
		Short: "Configure the CLI",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := loadConfig()
			if err != nil {
				println("could not load config file")
				return err
			}

			if endpoint == "" && credentials == "" {

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
				config.APIKey = strings.Split(result, ":")[0]
				config.APISecret = strings.Split(result, ":")[1]

				endpointSelect := promptui.Select{
					Label: "Select endpoint to use",
					Items: []string{"api-us1.scanii.com", "api-eu1.scanii.com", "api-eu2.scanii.com", "api-ap1.scanii.com", "api-ap2.scanii.com"},
				}
				_, selectedEndpoint, err2 := endpointSelect.Run()
				if err2 != nil {
					return err2
				}
				config.Endpoint = selectedEndpoint

			} else {
				if endpoint != "" {
					config.Endpoint = endpoint
				}
				if credentials != "" {
					config.APIKey = strings.Split(credentials, ":")[0]
					config.APISecret = strings.Split(credentials, ":")[1]
				}
			}

			err = saveConfig(config)
			if err != nil {
				return err
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&endpoint, "endpoint", "e", "", "Endpoint to use, see https://docs.scanii.com/article/161-endpoints-and-regions")
	cmd.Flags().StringVarP(&credentials, "credentials", "c", "", "API credentials to use in the format key:secret")

	return cmd
}

type configuration struct {
	APIKey    string    `json:"apiKey"`
	APISecret string    `json:"apiSecret"`
	Endpoint  string    `json:"endpoint"`
	Updated   time.Time `json:"updated"`
}

func configPath() (string, error) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	cp := path.Join(dirname, ".scanii-cli", "config.json")
	_, err = os.Stat(cp)

	if err != nil {
		if os.IsNotExist(err) {
			// create the directory
			err = os.MkdirAll(path.Dir(cp), 0755)
			if err != nil {
				return "", err
			}

		} else {
			return "", err
		}
	}

	return cp, nil

}

func loadConfig() (*configuration, error) {
	config := &configuration{}

	cp, err := configPath()
	if err != nil {
		return nil, err
	}

	slog.Debug("saving/loading config from ", "path", cp)

	fd, err := os.Open(cp)
	if err != nil {
		return nil, err
	}

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

func saveConfig(config *configuration) error {
	config.Updated = time.Now()
	// saving config
	cp, err := configPath()
	if err != nil {
		return err
	}

	js, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(cp, js, 0644)
	if err != nil {
		return err

	}

	slog.Debug("config saved", "path", cp)

	return nil

}
