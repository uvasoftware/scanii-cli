package v22

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
)

type store struct {
	path string
}

func (s store) load(key string, v any) error {
	// same rules as json.unmarshal
	if v == nil {
		return fmt.Errorf("v cannot be nil")
	}
	if reflect.ValueOf(v).Kind() != reflect.Ptr {
		return fmt.Errorf("v must be a pointer")
	}

	dest := filepath.Join(s.path, fmt.Sprintf("%s.json", key))
	slog.Debug("loading result", "dest", dest)
	js, err := os.ReadFile(dest)
	if err != nil {
		return err
	}

	err = json.Unmarshal(js, &v)
	if err != nil {
		return err
	}
	return nil
}

func (s store) save(key string, v any) error {
	js, err := json.Marshal(v)
	if err != nil {
		return err
	}

	dest := filepath.Join(s.path, fmt.Sprintf("%s.json", key))
	err = os.WriteFile(dest, js, 0644)
	if err != nil {
		return err
	}

	slog.Info("saved value", "dest", dest)
	return nil
}

func (s store) remove(key string) (bool, error) {
	dest := filepath.Join(s.path, fmt.Sprintf("%s.json", key))
	err := os.Remove(dest)
	if err != nil {
		switch err {
		case os.ErrNotExist:
			return false, nil
		}
		return false, err
	}
	return true, nil
}
