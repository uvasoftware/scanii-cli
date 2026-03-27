package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSaveAndLoadConfig(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	c := &configuration{
		Credentials: "testkey:testsecret",
		Endpoint:    "api-us1.scanii.com",
	}

	err := saveConfig("testprofile", c)
	if err != nil {
		t.Fatalf("failed to save config: %s", err)
	}

	// verify file exists at expected path
	expectedPath := filepath.Join(tmpHome, ".config", "scanii-cli", "testprofile.json")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Fatalf("expected config file at %s", expectedPath)
	}

	// load it back
	loaded, err := loadConfig("testprofile")
	if err != nil {
		t.Fatalf("failed to load config: %s", err)
	}

	if loaded.Credentials != "testkey:testsecret" {
		t.Fatalf("expected Credentials 'testkey:testsecret', got %q", loaded.Credentials)
	}
	if loaded.apiKey() != "testkey" {
		t.Fatalf("expected apiKey 'testkey', got %q", loaded.apiKey())
	}
	if loaded.apiSecret() != "testsecret" {
		t.Fatalf("expected apiSecret 'testsecret', got %q", loaded.apiSecret())
	}
	if loaded.Endpoint != "api-us1.scanii.com" {
		t.Fatalf("expected Endpoint 'api-us1.scanii.com', got %q", loaded.Endpoint)
	}
	if loaded.CreatedAt.IsZero() {
		t.Fatalf("expected CreatedAt to be set")
	}
}

func TestLoadConfigNonExistent(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	_, err := loadConfig("doesnotexist")
	if err == nil {
		t.Fatalf("expected error loading non-existent profile")
	}
}

func TestSaveConfigOverwrite(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	c1 := &configuration{
		Credentials: "key1:secret1",
		Endpoint:    "api-us1.scanii.com",
	}
	err := saveConfig("myprofile", c1)
	if err != nil {
		t.Fatalf("failed to save config: %s", err)
	}

	c2 := &configuration{
		Credentials: "key2:secret2",
		Endpoint:    "api-eu1.scanii.com",
	}
	err = saveConfig("myprofile", c2)
	if err != nil {
		t.Fatalf("failed to overwrite config: %s", err)
	}

	loaded, err := loadConfig("myprofile")
	if err != nil {
		t.Fatalf("failed to load config: %s", err)
	}
	if loaded.Credentials != "key2:secret2" {
		t.Fatalf("expected Credentials 'key2:secret2', got %q", loaded.Credentials)
	}
	if loaded.Endpoint != "api-eu1.scanii.com" {
		t.Fatalf("expected Endpoint 'api-eu1.scanii.com', got %q", loaded.Endpoint)
	}
}

func TestMultipleProfiles(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	profiles := map[string]*configuration{
		"dev": {
			Credentials: "devkey:devsecret",
			Endpoint:    "localhost:4000",
		},
		"staging": {
			Credentials: "stagingkey:stagingsecret",
			Endpoint:    "api-eu1.scanii.com",
		},
		"prod": {
			Credentials: "prodkey:prodsecret",
			Endpoint:    "api-us1.scanii.com",
		},
	}

	for name, cfg := range profiles {
		if err := saveConfig(name, cfg); err != nil {
			t.Fatalf("failed to save profile %q: %s", name, err)
		}
	}

	for name, expected := range profiles {
		loaded, err := loadConfig(name)
		if err != nil {
			t.Fatalf("failed to load profile %q: %s", name, err)
		}
		if loaded.Credentials != expected.Credentials {
			t.Fatalf("profile %q: expected Credentials %q, got %q", name, expected.Credentials, loaded.Credentials)
		}
		if loaded.Endpoint != expected.Endpoint {
			t.Fatalf("profile %q: expected Endpoint %q, got %q", name, expected.Endpoint, loaded.Endpoint)
		}
	}
}

func TestConfigDir(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	dir, err := configDir()
	if err != nil {
		t.Fatalf("failed to get config dir: %s", err)
	}

	expected := filepath.Join(tmpHome, ".config", "scanii-cli")
	if dir != expected {
		t.Fatalf("expected config dir %q, got %q", expected, dir)
	}
}

func TestConfigPath(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	p, err := configPath("myprofile")
	if err != nil {
		t.Fatalf("failed to get config path: %s", err)
	}

	expected := filepath.Join(tmpHome, ".config", "scanii-cli", "myprofile.json")
	if p != expected {
		t.Fatalf("expected config path %q, got %q", expected, p)
	}
}

func TestSaveConfigSetsCreatedAtTime(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	before := time.Now().Add(-time.Second)
	c := &configuration{
		Credentials: "k:s",
		Endpoint:    "api-us1.scanii.com",
	}
	err := saveConfig("timetest", c)
	if err != nil {
		t.Fatalf("failed to save: %s", err)
	}
	after := time.Now().Add(time.Second)

	loaded, err := loadConfig("timetest")
	if err != nil {
		t.Fatalf("failed to load: %s", err)
	}

	if loaded.CreatedAt.Before(before) || loaded.CreatedAt.After(after) {
		t.Fatalf("CreatedAt time %v not in expected range [%v, %v]", loaded.CreatedAt, before, after)
	}
}

func TestConfigFilePermissions(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	c := &configuration{
		Credentials: "k:s",
		Endpoint:    "api-us1.scanii.com",
	}
	err := saveConfig("permtest", c)
	if err != nil {
		t.Fatalf("failed to save: %s", err)
	}

	p, _ := configPath("permtest")
	info, err := os.Stat(p)
	if err != nil {
		t.Fatalf("failed to stat: %s", err)
	}

	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Fatalf("expected file permissions 0600, got %o", perm)
	}
}

func TestProfileDeleteCommand(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	c := &configuration{
		Credentials: "k:s",
		Endpoint:    "api-us1.scanii.com",
	}
	err := saveConfig("todelete", c)
	if err != nil {
		t.Fatalf("failed to save: %s", err)
	}

	// verify it exists
	_, err = loadConfig("todelete")
	if err != nil {
		t.Fatalf("profile should exist: %s", err)
	}

	// delete via file removal (same as delete command logic)
	p, _ := configPath("todelete")
	err = os.Remove(p)
	if err != nil {
		t.Fatalf("failed to delete: %s", err)
	}

	// verify it's gone
	_, err = loadConfig("todelete")
	if err == nil {
		t.Fatalf("expected error after deletion")
	}
}

func TestDefaultProfileName(t *testing.T) {
	if defaultProfileName != "default" {
		t.Fatalf("expected default profile name to be 'default', got %q", defaultProfileName)
	}
}

func TestProfileCommandStructure(t *testing.T) {
	cmd := ProfileCommand()

	if cmd.Use != "profile" {
		t.Fatalf("expected Use 'profile', got %q", cmd.Use)
	}

	subcommands := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		subcommands[sub.Use] = true
	}

	expected := []string{"create [name]", "list [name]", "delete <name>"}
	for _, e := range expected {
		if !subcommands[e] {
			t.Fatalf("expected subcommand %q not found", e)
		}
	}
}

func TestJsonFormat(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	c := &configuration{
		Credentials: "sck_abc123:scks_secret456",
		Endpoint:    "api-us1.staging.scanii.com",
	}

	err := saveConfig("jsontest", c)
	if err != nil {
		t.Fatalf("failed to save: %s", err)
	}

	// read raw JSON and verify structure
	p, _ := configPath("jsontest")
	raw, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("failed to read file: %s", err)
	}

	var parsed map[string]interface{}
	err = json.Unmarshal(raw, &parsed)
	if err != nil {
		t.Fatalf("failed to parse JSON: %s", err)
	}

	// verify expected keys exist
	expectedKeys := []string{"endpoint", "createdAt", "version", "credentials"}
	for _, key := range expectedKeys {
		if _, ok := parsed[key]; !ok {
			t.Fatalf("expected key %q in JSON output, got keys: %v", key, parsed)
		}
	}

	// verify no old keys exist
	oldKeys := []string{"apiKey", "apiSecret", "updated"}
	for _, key := range oldKeys {
		if _, ok := parsed[key]; ok {
			t.Fatalf("unexpected old key %q found in JSON output", key)
		}
	}

	// verify values
	if parsed["endpoint"] != "api-us1.staging.scanii.com" {
		t.Fatalf("expected endpoint 'api-us1.staging.scanii.com', got %v", parsed["endpoint"])
	}
	if parsed["credentials"] != "sck_abc123:scks_secret456" {
		t.Fatalf("expected credentials 'sck_abc123:scks_secret456', got %v", parsed["credentials"])
	}
	if parsed["version"] != nil {
		t.Fatalf("expected version to be null, got %v", parsed["version"])
	}
}

func TestApiKeyAndSecretParsing(t *testing.T) {
	c := &configuration{
		Credentials: "mykey:mysecret",
	}
	if c.apiKey() != "mykey" {
		t.Fatalf("expected apiKey 'mykey', got %q", c.apiKey())
	}
	if c.apiSecret() != "mysecret" {
		t.Fatalf("expected apiSecret 'mysecret', got %q", c.apiSecret())
	}

	// credentials with no colon
	c2 := &configuration{
		Credentials: "onlykey",
	}
	if c2.apiKey() != "onlykey" {
		t.Fatalf("expected apiKey 'onlykey', got %q", c2.apiKey())
	}
	if c2.apiSecret() != "" {
		t.Fatalf("expected apiSecret '', got %q", c2.apiSecret())
	}

	// credentials with multiple colons (secret contains colon)
	c3 := &configuration{
		Credentials: "key:secret:with:colons",
	}
	if c3.apiKey() != "key" {
		t.Fatalf("expected apiKey 'key', got %q", c3.apiKey())
	}
	if c3.apiSecret() != "secret:with:colons" {
		t.Fatalf("expected apiSecret 'secret:with:colons', got %q", c3.apiSecret())
	}
}
