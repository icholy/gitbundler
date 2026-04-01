package main

import (
	"fmt"
	"os"
	"regexp"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DataDir       string       `yaml:"data_dir"`
	Addr          string       `yaml:"addr"`
	MaxConcurrent int          `yaml:"max_concurrent"`
	Repos         []RepoConfig `yaml:"repos"`
}

type RepoConfig struct {
	Name        string            `yaml:"name"`
	URL         string            `yaml:"url"`
	Interval    time.Duration     `yaml:"interval"`
	Env         map[string]string `yaml:"env"`
	Repack      bool              `yaml:"repack"`
	CloneFlags  []string          `yaml:"clone_flags"`
	FetchFlags  []string          `yaml:"fetch_flags"`
	BundleFlags []string          `yaml:"bundle_flags"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}
	expanded, err := Expand(string(data), func(namespace, value string) (string, error) {
		if namespace == "env" {
			return os.Getenv(value), nil
		}
		return "", fmt.Errorf("unknown variable namespace: %s", namespace)
	})
	if err != nil {
		return nil, fmt.Errorf("expanding variables: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	if cfg.Addr == "" {
		cfg.Addr = ":8080"
	}
	if cfg.DataDir == "" {
		cfg.DataDir = "data"
	}
	for i, repo := range cfg.Repos {
		if repo.Name == "" {
			return nil, fmt.Errorf("repo at index %d has no name", i)
		}
		if repo.URL == "" {
			return nil, fmt.Errorf("repo at index %d has no url", i)
		}
		if repo.Interval == 0 {
			cfg.Repos[i].Interval = 5 * time.Minute
		}
	}
	return &cfg, nil
}

var expandPattern = regexp.MustCompile(`\$\{([^:}]+):([^}]+)\}`)

// Expand replaces ${namespace:value} patterns in the input string
// using the provided replace function.
func Expand(input string, replace func(namespace, value string) (string, error)) (string, error) {
	var lastErr error
	result := expandPattern.ReplaceAllStringFunc(input, func(match string) string {
		parts := expandPattern.FindStringSubmatch(match)
		if len(parts) != 3 {
			return match
		}
		expanded, err := replace(parts[1], parts[2])
		if err != nil {
			lastErr = err
			return match
		}
		return expanded
	})
	return result, lastErr
}
