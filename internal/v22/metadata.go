package v22

import (
	"regexp"
	"strings"
)

func extractMetadataKey(content string) string {
	regex := regexp.MustCompile(`^metadata\[(.*)]$`)
	parts := regex.FindAllStringSubmatch(content, -1)
	if len(parts) == 1 {
		v := strings.Trim(parts[0][1], " ")
		if v != "" {
			return v
		}
	}
	return ""
}
