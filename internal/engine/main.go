package engine

import (
	"crypto/sha1"
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"
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

type Result struct {
	Sha1          string
	Sha256        string
	ContentLength uint64
	Findings      []string
}

func (e *Engine) Process(contents io.Reader) (Result, error) {
	result := Result{}
	result.Findings = []string{}

	s1 := sha1.New()
	s2 := sha256.New()
	dest := io.MultiWriter(s1, s2)

	i, err := io.Copy(dest, contents)
	if err != nil {
		return result, err
	}

	result.Sha1 = fmt.Sprintf("%x", s1.Sum(nil))
	result.Sha256 = fmt.Sprintf("%x", s2.Sum(nil))
	result.ContentLength = uint64(i)

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
