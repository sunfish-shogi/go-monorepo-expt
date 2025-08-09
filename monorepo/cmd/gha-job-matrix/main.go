package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"

	"github.com/sunfish-shogi/go-monorepo-expt/monorepo"
	"github.com/sunfish-shogi/go-monorepo-expt/monorepo/internal/detect"
	"github.com/sunfish-shogi/go-monorepo-expt/monorepo/internal/golang"
)

func main() {
	gitRootPath := flag.String("git-root", ".", "Path to the git root directory")
	goWorkPath := flag.String("go-work", "go.work", "Path to the go.work file")
	baseCommit := flag.String("base-commit", "HEAD~", "Base commit to compare against")
	appendGitHubOutput := flag.Bool("append-github-output", false, "Append GitHub output")
	flag.Parse()

	if gitDir, err := os.Lstat(filepath.Join(*gitRootPath, ".git")); err != nil || !gitDir.IsDir() {
		panic("The specified git root path does not contain a .git directory")
	}

	goModulePaths, err := golang.ReadWorkspace(*goWorkPath)
	if err != nil {
		panic(err)
	}

	changedPackages, err := detect.DetectChangedPackages(context.Background(), &detect.Config{
		GitRootPath:   *gitRootPath,
		BaseCommit:    *baseCommit,
		GoModulePaths: goModulePaths,
	})
	if err != nil {
		panic(err)
	}

	config, err := monorepo.ReadRepoConfig(filepath.Join(*gitRootPath, monorepo.RepoConfigFileName))
	if err != nil {
		panic(err)
	}

	jobNames := make(map[string]struct{})
	changedTargets := make([]monorepo.BuildTarget, 0, len(config.GHA.Targets))
	for _, target := range config.GHA.Targets {
		for _, job := range target.Jobs {
			jobNames[job.Name] = struct{}{}
		}
		cleanedPath := filepath.Clean(target.Path)
		for _, pkgPath := range changedPackages {
			cleanedPkgPath := filepath.Clean(pkgPath.Dir)
			if cleanedPkgPath == cleanedPath {
				changedTargets = append(changedTargets, target)
				break
			}
		}
	}

	out := io.Writer(os.Stdout)

	if *appendGitHubOutput {
		gitHubOutputPath := os.Getenv("GITHUB_OUTPUT")
		file, err := os.OpenFile(gitHubOutputPath, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			panic(fmt.Errorf("failed to open GITHUB_OUTPUT file: %w", err))
		}
		defer file.Close()
		out = io.MultiWriter(os.Stdout, file)
	}

	for jobName := range jobNames {
		matrixItems := make([]MatrixItem, 0, len(changedTargets))
		for _, target := range changedTargets {
			if slices.ContainsFunc(target.Jobs, func(job monorepo.Job) bool {
				return job.Name == jobName
			}) {
				matrixItems = append(matrixItems, newMatrixItem(target))
			}
		}
		matrix := Matrix{Target: matrixItems}
		jsonData, err := json.Marshal(matrix)
		if err != nil {
			panic(err)
		}
		fmt.Fprintf(out, "%s_matrix=%s\n", jobName, jsonData)
		fmt.Fprintf(out, "needs_%s=%t\n", jobName, len(matrixItems) > 0)
	}
}

type MatrixItem map[string]any

type Matrix struct {
	Target []MatrixItem `json:"target"`
}

func newMatrixItem(target monorepo.BuildTarget) MatrixItem {
	item := make(MatrixItem, len(target.Props)+2)
	maps.Copy(item, target.Props)
	item["id"] = target.ID
	item["path"] = target.Path
	return item
}
