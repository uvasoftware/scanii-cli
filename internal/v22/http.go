package v22

import (
	"encoding/json"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, data any, headers http.Header) error {
	// this ident matches current production settings
	js, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	if headers != nil {
		for key, value := range headers {
			w.Header()[key] = value
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(js)
	if err != nil {
		return err
	}

	return nil
}
