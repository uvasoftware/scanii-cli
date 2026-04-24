package engine

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

type callbackItem struct {
	result      *Result
	destination string
}

type callback struct {
	ID            string            `json:"id"`
	Checksum      string            `json:"checksum,omitempty"`
	ContentLength uint64            `json:"content_length,omitempty"`
	ContentType   string            `json:"content_type,omitempty"`
	Findings      []string          `json:"findings"`
	CreationDate  string            `json:"creation_date,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	Error         string            `json:"error,omitempty"`
}

// newRunner creates a new runner returning a channel to submit callbacks to
func (e *Engine) newRunner() (chan callbackItem, error) {
	const UA = "scanii/jackfruit (see https://www.scanii.com)"

	queue := make(chan callbackItem, 100)
	wait := time.Duration(0)
	if e.config.CallbackWait != nil {
		wait = *e.config.CallbackWait
	}
	slog.Debug("callback wait", "wait", wait)
	client := &http.Client{Timeout: 30 * time.Second}

	go func() {
		for msg := range queue {
			slog.Debug("sending callback", "destination", msg.destination)
			time.Sleep(wait)
			slog.Debug("post wait", "destination", msg.destination)

			body, err := json.Marshal(&callback{
				ID:            msg.result.ID,
				ContentLength: msg.result.ContentLength,
				ContentType:   msg.result.ContentType,
				Checksum:      msg.result.Sha1,
				Findings:      msg.result.Findings,
				CreationDate:  msg.result.CreationDate,
				Metadata:      msg.result.Metadata,
				Error:         msg.result.Error,
			})
			if err != nil {
				slog.Error("failed to marshal callback", "error", err)
				return
			}
			req, err := http.NewRequest("POST", msg.destination, bytes.NewReader(body))
			if err != nil {
				slog.Error("failed to create request", "error", err)
				return
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("User-Agent", UA)

			resp, err := client.Do(req)
			if err != nil {
				return
			}
			resp.Body.Close()
			slog.Debug("callback delivered", "destination", msg.destination, "status", resp.StatusCode)
		}
	}()
	return queue, nil
}
