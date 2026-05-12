package server

import (
	"net/http"
	"testing"
)

// TestCORS_PreflightReturnsAllowHeaders verifies that an OPTIONS preflight
// against an authenticated route succeeds without credentials and returns
// the Access-Control headers a browser needs to permit the follow-up call.
func TestCORS_PreflightReturnsAllowHeaders(t *testing.T) {
	ts := startServer(t)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodOptions, ts.URL+"/v2.2/files", http.NoBody)
	if err != nil {
		t.Fatalf("new request: %s", err)
	}
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "authorization")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("preflight: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("preflight status: want 200, got %d", resp.StatusCode)
	}

	checks := map[string]string{
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "GET, POST, HEAD, OPTIONS, DELETE",
		"Access-Control-Allow-Headers": "Authorization, User-Agent",
		"Access-Control-Max-Age":       "300",
	}
	for header, want := range checks {
		if got := resp.Header.Get(header); got != want {
			t.Errorf("%s: want %q, got %q", header, want, got)
		}
	}
}

// TestCORS_AllowOriginOnAuthenticatedResponse verifies that the Allow-Origin
// header is also present on the actual cross-origin request (not just the
// preflight) so the browser will let the calling script read the response.
func TestCORS_AllowOriginOnAuthenticatedResponse(t *testing.T) {
	ts := startServer(t)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, ts.URL+"/v2.2/ping", http.NoBody)
	if err != nil {
		t.Fatalf("new request: %s", err)
	}
	req.SetBasicAuth("key", "secret")
	req.Header.Set("Origin", "http://example.com")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("get: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("ping status: want 200, got %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("Access-Control-Allow-Origin: want %q, got %q", "*", got)
	}
}
