package main

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"

	detector "github.com/sunfish-shogi/go-change-detector"
	"golang.org/x/mod/modfile"
)

func main() {
	gitRootPath := flag.String("git-root", ".", "Path to the git root directory")
	goWorkPath := flag.String("go-work", "go.work", "Path to the go.work file")
	baseCommit := flag.String("base-commit", "HEAD~", "Base commit to compare against")
	flag.Parse()

	if gitDir, err := os.Lstat(filepath.Join(*gitRootPath, ".git")); err != nil || !gitDir.IsDir() {
		panic("The specified git root path does not contain a .git directory")
	}

	goModulePaths, err := readWorkspace(*goWorkPath)
	if err != nil {
		panic(err)
	}

	changedPackages, err := detector.DetectChangedPackages(&detector.Config{
		GitRootPath:   *gitRootPath,
		BaseCommit:    *baseCommit,
		GoModulePaths: goModulePaths,
	})
	if err != nil {
		panic(err)
	}

	config, err := readMonorepoConfig(filepath.Join(*gitRootPath, "go-monorepo.yaml"))
	if err != nil {
		panic(err)
	}

	targets := make([]map[string]string, 0, len(config.GHA.Targets))
	for _, target := range config.GHA.Targets {
		cleanedPath := filepath.Clean(target.Path)
		for _, pkgPath := range changedPackages {
			cleanedPkgPath := filepath.Clean(pkgPath.Dir)
			if cleanedPkgPath == cleanedPath {
				props := make(map[string]string, len(target.Props)+2)
				props["name"] = target.Name
				props["path"] = target.Path
				for k, v := range target.Props {
					props[k] = v
				}
				targets = append(targets, props)
				break
			}
		}
	}
	output, err := json.Marshal(targets)
	if err != nil {
		panic(err)
	}
	println(string(output))
}

func readWorkspace(goWorkPath string) ([]string, error) {
	dir := filepath.Dir(goWorkPath)
	data, err := os.ReadFile(goWorkPath)
	if err != nil {
		return nil, err
	}
	workFile, err := modfile.ParseWork(filepath.Base(goWorkPath), data, nil)
	if err != nil {
		return nil, err
	}
	var goModulePaths []string
	for _, use := range workFile.Use {
		path := filepath.Join(dir, use.Path)
		goModulePaths = append(goModulePaths, path)
	}
	// FIXME: support replace-directives
	return goModulePaths, nil
}

type GoMonorepoConfig struct {
	GHA `yaml:"gha"`
}

type GHA struct {
	Targets []BuildTarget `yaml:"targets"`
}

type BuildTarget struct {
	Name  string            `yaml:"name"`
	Path  string            `yaml:"path"`
	Jobs  []Job             `yaml:"jobs,omitempty"`
	Props map[string]string `yaml:"props,omitempty"`
}

type Job struct {
	Name string `yaml:"name"`
}

func readMonorepoConfig(path string) (*GoMonorepoConfig, error) {
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
