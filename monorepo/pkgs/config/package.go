package config

import (
	"os"

	"github.com/goccy/go-yaml"
)

const PackageConfigFileName = "go-mono-pkg.yaml"

type PackageConfig struct {
	ExtraDependencies []string `yaml:"extra_dependencies,omitempty"`
}

func ReadPackageConfig(path string) (*PackageConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config PackageConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
