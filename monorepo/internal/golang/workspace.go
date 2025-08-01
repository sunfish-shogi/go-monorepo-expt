package golang

import (
	"os"
	"path/filepath"

	"golang.org/x/mod/modfile"
)

func ReadWorkspace(goWorkPath string) ([]string, error) {
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
