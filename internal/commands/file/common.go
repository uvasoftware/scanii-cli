package file

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/uvasoftware/scanii-cli/internal/terminal"
)

type resultRecord struct {
	path          string
	err           error
	contentType   string
	findings      []string
	checksum      string
	id            string
	location      string
	contentLength uint64
	creationDate  string
	metadata      map[string]string
}

// extractMetadata parses the metadata string and returns a map of key/value pairs.
func extractMetadata(metadata string) map[string]string {
	result := make(map[string]string)
	if metadata == "" {
		return result
	}

	parts := strings.Split(metadata, ",")
	for _, p := range parts {
		kv := strings.Split(p, "=")
		if len(kv) != 2 {
			slog.Warn("invalid metadata entry", "entry", p)
			continue
		}
		result[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
	}
	return result
}

func printFileResult(result *resultRecord) {
	if result.path != "" {
		terminal.Title(fmt.Sprintf("%s:", result.path))
	}

	if result.err != nil {
		terminal.KeyValue("error:", result.err.Error())
		return
	}

	terminal.KeyValue("id:", result.id)

	if result.checksum != "" {
		terminal.KeyValue("checksum/sha1:", result.checksum)
	}

	if result.location != "" {
		terminal.KeyValue("location:", result.location)
	}

	if result.contentType != "" {
		terminal.KeyValue("content type:", result.contentType)
	}

	if result.contentLength != 0 {
		terminal.KeyValueW("content length:", terminal.FormatBytes(result.contentLength), 16)
	}

	if result.creationDate != "" {
		terminal.KeyValueW("creation date:", terminal.FormatTime(result.creationDate), 16)
	}

	if len(result.findings) > 0 {
		terminal.KeyValue("findings:", strings.Join(result.findings, ","))
	} else {
		terminal.KeyValue("findings:", "none")
	}

	if len(result.metadata) > 0 {
		fmt.Println("  metadata:")
		for k, v := range result.metadata {
			fmt.Printf("    %s → %s\n", k, v)
		}
	} else {
		terminal.KeyValue("metadata:", "none")
	}
}
