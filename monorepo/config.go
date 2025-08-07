package monorepo

import (
	"os"

	"github.com/goccy/go-yaml"
)

type GoMonorepoConfig struct {
	GHA `yaml:"gha"`
}

type GHA struct {
	Targets []BuildTarget `yaml:"targets"`
}

type BuildTarget struct {
	ID    string         `yaml:"id"`
	Path  string         `yaml:"path"`
	Jobs  []Job          `yaml:"jobs,omitempty"`
	Props map[string]any `yaml:"props,omitempty"`
}

type Job struct {
	Name string `yaml:"name"`
}

func ReadMonorepoConfig(path string) (*GoMonorepoConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config GoMonorepoConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
