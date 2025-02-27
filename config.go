package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/creasty/defaults"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Repos   map[string][]string `yaml:"repositories"`
	Labels  []string            `yaml:"labels" default:"[\"good first issue\"]"`
	PerPage int                 `yaml:"per_page" default:"100"`
}

func loadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	if err := defaults.Set(&config); err != nil {
		return nil, err
	}
	return &config, nil
}

func parseRepoURL(url string) (owner, repo string, err error) {
	parts := strings.Split(strings.TrimPrefix(url, "https://github.com/"), "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("Invalid repository URL: %s", url)
	}
	return parts[0], parts[1], nil
}
