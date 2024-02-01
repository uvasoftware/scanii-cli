package commands

import (
	"context"
	"fmt"
	"net/http"
	v22 "scanii-cli/internal/v22"
	"scanii-cli/internal/vcs"
	"strings"
)

// createClient creates a new Scanii client
func createClient(config *configuration) (*v22.Client, error) {

	dest := fmt.Sprintf("https://%s/v2.2", config.Endpoint)
	if strings.HasPrefix(config.Endpoint, "localhost") {
		dest = fmt.Sprintf("http://%s/v2.2", config.Endpoint)
	}

	customizer := v22.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		req.SetBasicAuth(config.ApiKey, config.ApiSecret)
		req.Header.Add("User-Agent", fmt.Sprintf("scanii-cli/v%s", vcs.Version()))
		return nil
	})

	client, err := v22.NewClient(dest, customizer, v22.WithHTTPClient(&http.Client{
		//Timeout: 30 * time.Second,
		//Transport: &loggingTransport{},
	}))
	if err != nil {
		return nil, err
	}

	return client, nil

}

//type loggingTransport struct{}
//
//func (s *loggingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
//	bytes, _ := httputil.DumpRequestOut(r, true)
//
//	resp, err := http.DefaultTransport.RoundTrip(r)
//	// err is returned after dumping the response
//
//	respBytes, _ := httputil.DumpResponse(resp, true)
//	bytes = append(bytes, respBytes...)
//
//	fmt.Printf("%s\n", bytes)
//
//	return resp, err
//}
