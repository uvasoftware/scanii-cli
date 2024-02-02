package engine

import (
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/gabriel-vasile/mimetype"
	"io"
	"log/slog"
	"strings"
	"time"
)

//go:embed default.json
var defaultConfig string

type Engine struct {
	config *Config
}

type Rule struct {
	Format  string `json:"format"`
	Content string `json:"content"`
	Result  string `json:"result"`
}
type Config struct {
	Rules []Rule `json:"rules"`
}

func New() (*Engine, error) {
	rules := make([]Rule, 0)

	// default rule
	rules = append(rules, Rule{
		Format:  "sha256",
		Content: "",
	})

	engine := &Engine{
		config: &Config{
			Rules: make([]Rule, 0),
		},
	}
	slog.Debug("loading default config")
	err := engine.LoadConfig(strings.NewReader(defaultConfig))
	if err != nil {
		return nil, err
	}

	return engine, nil
}

func (e *Engine) LoadConfig(reader io.Reader) error {
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(e.config)
	if err != nil {
		return err
	}
	return nil
}

func (e *Engine) RuleCount() int {
	return len(e.config.Rules)

}

type Result struct {
	Id            string
	Sha1          string
	Sha256        string
	ContentLength uint64
	Findings      []string
	ContentType   string
	CreationDate  string
	Metadata      map[string]string
	Error         string
}

func (e *Engine) Process(contents io.Reader) (Result, error) {
	result := Result{
		CreationDate: time.Now().UTC().Format(time.RFC3339Nano),
	}

	result.Findings = []string{}
	s1 := sha1.New()
	s2 := sha256.New()
	dest := io.MultiWriter(s1, s2)

	// detecting mime type
	mime, recycledInput, err := recycleReader(contents)
	if err != nil {
		return result, err
	}

	i, err := io.Copy(dest, recycledInput)
	if err != nil {
		return result, err
	}

	result.Sha1 = fmt.Sprintf("%x", s1.Sum(nil))
	result.Sha256 = fmt.Sprintf("%x", s2.Sum(nil))
	result.ContentLength = uint64(i)

	if err != nil {
		slog.Error("error detecting mime type", "error", err.Error())
	}
	result.ContentType = mime

	appendIfMissing := func(slice []string, s string) []string {
		for _, ele := range slice {
			if ele == s {
				return slice
			}
		}
		return append(slice, s)
	}

	// looking for matches in the rules:
	for _, rule := range e.config.Rules {
		switch rule.Format {
		case "sha1":
			if result.Sha1 == rule.Content {
				result.Findings = appendIfMissing(result.Findings, rule.Result)
			}
		case "sha256":
			if result.Sha256 == rule.Content {
				result.Findings = appendIfMissing(result.Findings, rule.Result)
			}
		}
	}

	return result, nil

}

// recycleReader returns the MIME type of input and a new reader
// containing the whole data from input.
func recycleReader(input io.Reader) (mimeType string, recycled io.Reader, err error) {
	// header will store the bytes mimetype uses for detection.
	header := bytes.NewBuffer(nil)

	// After DetectReader, the data read from input is copied into header.
	mtype, err := mimetype.DetectReader(io.TeeReader(input, header))
	if err != nil {
		return
	}

	// Concatenate back the header to the rest of the file.
	// recycled now contains the complete, original data.
	recycled = io.MultiReader(header, input)

	return mtype.String(), recycled, err
}
