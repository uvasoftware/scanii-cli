package v22

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"scanii-cli/internal/engine"
)

type store struct {
	path string
}

func (s store) load(key string) (*engine.Result, error) {
	dest := filepath.Join(s.path, fmt.Sprintf("%s.json", key))
	slog.Debug("loading result", "dest", dest)
	js, err := os.ReadFile(dest)
	if err != nil {
		return nil, err
	}

	result := engine.Result{}
	err = json.Unmarshal(js, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (s store) save(key string, result *engine.Result) error {
	js, err := json.Marshal(result)
	if err != nil {
		return err
	}

	dest := filepath.Join(s.path, fmt.Sprintf("%s.json", key))
	err = os.WriteFile(dest, js, 0644)
	if err != nil {
		return err
	}

	slog.Debug("saved result", "dest", dest)
	return nil
}
