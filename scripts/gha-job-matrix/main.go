package main

import (
	"flag"
	"os"
	"path/filepath"

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
	for _, pkg := range changedPackages {
		println(pkg.Dir)
	}
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
