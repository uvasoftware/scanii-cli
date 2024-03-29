package commands

import (
	"context"
	"fmt"
	v22 "github.com/uvasoftware/scanii-cli/internal/v22"
	"github.com/uvasoftware/scanii-cli/internal/vcs"
	"net/http"
	"strings"
	"time"
)

// createClient creates a new Scanii client
func createClient(config *configuration) (*v22.Client, error) {

	dest := fmt.Sprintf("https://%s/v2.2", config.Endpoint)
	if strings.HasPrefix(config.Endpoint, "localhost") {
		dest = fmt.Sprintf("http://%s/v2.2", config.Endpoint)
	}

	customizer := v22.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		req.SetBasicAuth(config.APIKey, config.APISecret)
		req.Header.Add("User-Agent", fmt.Sprintf("github.com/uvasoftware/scanii-cli/v%s", vcs.Version()))
		return nil
	})

	client, err := v22.NewClient(dest, customizer, v22.WithHTTPClient(&http.Client{
		Timeout: 30 * time.Second,
	}))
	if err != nil {
		return nil, err
	}

	return client, nil

}
