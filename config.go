package main

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DataDir string       `yaml:"data_dir"`
	Addr    string       `yaml:"addr"`
	Repos   []RepoConfig `yaml:"repos"`
}

type RepoConfig struct {
	Name     string        `yaml:"name"`
	URL      string        `yaml:"url"`
	Interval time.Duration `yaml:"interval"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
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
