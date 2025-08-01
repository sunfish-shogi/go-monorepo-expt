package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"maps"
	"os"
	"path/filepath"

	detector "github.com/sunfish-shogi/go-change-detector"
	"github.com/sunfish-shogi/go-monorepo-expt/monorepo"
	"github.com/sunfish-shogi/go-monorepo-expt/monorepo/internal/golang"
)

func main() {
	gitRootPath := flag.String("git-root", ".", "Path to the git root directory")
	goWorkPath := flag.String("go-work", "go.work", "Path to the go.work file")
	baseCommit := flag.String("base-commit", "HEAD~", "Base commit to compare against")
	flag.Parse()

	if gitDir, err := os.Lstat(filepath.Join(*gitRootPath, ".git")); err != nil || !gitDir.IsDir() {
		panic("The specified git root path does not contain a .git directory")
	}

	goModulePaths, err := golang.ReadWorkspace(*goWorkPath)
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

	config, err := monorepo.ReadMonorepoConfig(filepath.Join(*gitRootPath, "go-monorepo.yaml"))
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
				maps.Copy(props, target.Props)
				props["name"] = target.Name
				props["path"] = target.Path
				targets = append(targets, props)
				break
			}
		}
	}
	output, err := json.Marshal(targets)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(output))
}
