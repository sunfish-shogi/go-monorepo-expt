package detector

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sunfish-shogi/go-monorepo-expt/monorepo/internal/git"
	"github.com/sunfish-shogi/go-monorepo-expt/monorepo/internal/golang"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
	"golang.org/x/tools/go/packages"
)

type Config struct {
	GitRootPath   string   // path to the root directory of the git repository
	BaseCommit    string   // base commit revision to compare against, e.g., "HEAD~" for the previous commit
	GoModulePaths []string // paths to go modules
}

type Package struct {
	Dir        string // directory of the package
	ImportPath string // import path of the package
}

func DetectChangedPackages(ctx context.Context, config *Config) ([]Package, error) {
	detector, err := newChangeDetector(ctx, config)
	if err != nil {
		return nil, err
	}
	return detector.detectChangedPackages()
}

type changeDetector struct {
	context             context.Context
	config              *Config
	gitRootFullPath     string                         // full path to the git root directory
	changedModulesCache map[string]map[string]struct{} // cache for changed modules
	changedFiles        map[string]struct{}            // full paths of changed files
}

func newChangeDetector(ctx context.Context, config *Config) (*changeDetector, error) {
	gitRootPath := config.GitRootPath
	if gitRootPath == "" {
		gitRootPath = "."
	}
	var gitRootFullPath string
	if config != nil && config.GitRootPath != "" {
		path, err := filepath.Abs(config.GitRootPath)
		if err != nil {
			return nil, err
		}
		gitRootFullPath = path
	}
	changedFiles, err := git.ChangedFilesFrom(ctx, gitRootFullPath, config.BaseCommit)
	if err != nil {
		return nil, err
	}
	changedFilesMap := make(map[string]struct{}, len(changedFiles))
	for _, file := range changedFiles {
		changedFilesMap[filepath.Join(gitRootFullPath, file)] = struct{}{}
	}
	return &changeDetector{
		context:             ctx,
		config:              config,
		gitRootFullPath:     gitRootFullPath,
		changedModulesCache: make(map[string]map[string]struct{}),
		changedFiles:        changedFilesMap,
	}, nil
}

func (cd *changeDetector) detectChangedPackages() ([]Package, error) {
	// List all Go packages in the specified modules
	goPackages := make([]*packages.Package, 0, 64)
	for _, modulePath := range cd.config.GoModulePaths {
		pkgs, err := golang.ListPackages(cd.context, modulePath)
		if err != nil {
			return nil, err
		}
		goPackages = append(goPackages, pkgs...)
	}

	// Filter packages that have changed
	var changedPackages = make(map[string]*packages.Package)
	for _, goPackage := range goPackages {
		if cd.isPackageChanged(goPackage) {
			changedPackages[goPackage.PkgPath] = goPackage
		}
	}

	// List packages that have changed by dependencies
	for _, goPackage := range goPackages {
		if _, exists := changedPackages[goPackage.PkgPath]; exists {
			continue // Skip packages that are already marked as changed
		} else if updated, err := cd.isModuleUpdated(goPackage, changedPackages); err != nil {
			return nil, err
		} else if updated {
			changedPackages[goPackage.PkgPath] = goPackage
		}
	}

	var results []Package
	for pkg := range changedPackages {
		relativeDir, err := filepath.Rel(cd.gitRootFullPath, changedPackages[pkg].Dir)
		if err != nil {
			return nil, err
		}
		results = append(results, Package{
			Dir:        relativeDir,
			ImportPath: changedPackages[pkg].PkgPath,
		})
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].ImportPath < results[j].ImportPath
	})
	return results, nil
}

func (cd *changeDetector) isPackageChanged(goPackage *packages.Package) bool {
	files := goPackage.GoFiles
	files = append(files, goPackage.OtherFiles...)
	files = append(files, goPackage.EmbedFiles...)
	for _, filePath := range files {
		if _, exists := cd.changedFiles[filePath]; exists {
			return true
		}
	}
	return false
}

func (cd *changeDetector) isModuleUpdated(goPackage *packages.Package, changedPackages map[string]*packages.Package) (bool, error) {
	changedModules, err := cd.listChangedModules(goPackage)
	if err != nil {
		return false, err
	}
	for _, importPackage := range goPackage.Imports {
		path := importPackage.PkgPath
		// 1 Check other packages in the same module
		if _, exists := changedPackages[path]; exists {
			return true, nil
		}
		// 2 Check if the third-party module has changed
		for path != "" {
			if _, exists := changedModules[path]; exists {
				return true, nil
			}

			// Move to the parent module
			lastSlashIndex := strings.LastIndex(path, "/")
			if lastSlashIndex == -1 {
				break // No more parent module
			} else {
				path = path[:lastSlashIndex]
			}
		}
	}
	return false, nil
}

func (cd *changeDetector) listChangedModules(goPackage *packages.Package) (map[string]struct{}, error) {
	goModFullPath := goPackage.Module.GoMod
	if cache, ok := cd.changedModulesCache[goModFullPath]; ok {
		return cache, nil
	}
	currentGoMod, err := cd.readGoModFile(goModFullPath)
	if err != nil {
		return nil, err
	}
	previousGoMod, err := cd.readGoModFileFromGit(goModFullPath, cd.config.BaseCommit)
	if err != nil {
		return nil, err
	}
	currentVersions := getModuleVersions(currentGoMod)
	previousVersions := getModuleVersions(previousGoMod)
	changedModules := make(map[string]struct{})
	for module, currentVersion := range currentVersions {
		previousVersion, exists := previousVersions[module]
		if !exists || currentVersion != previousVersion {
			changedModules[module] = struct{}{}
		}
	}
	cd.changedModulesCache[goModFullPath] = changedModules
	return changedModules, nil
}

func getModuleVersions(goMod *modfile.File) map[string]string {
	if goMod == nil {
		return nil
	}
	replaceMap := make(map[module.Version]module.Version)
	for _, replace := range goMod.Replace {
		replaceMap[replace.Old] = replace.New
	}
	moduleVersions := make(map[string]string)
	for _, req := range goMod.Require {
		replace, exists := replaceMap[req.Mod]
		if !exists && req.Mod.Version != "" {
			replace, exists = replaceMap[module.Version{Path: req.Mod.Path, Version: ""}]
		}
		if exists {
			if replace.Version != "" {
				moduleVersions[req.Mod.Path] = fmt.Sprintf("%s@%s", replace.Path, replace.Version)
			} else {
				moduleVersions[req.Mod.Path] = fmt.Sprintf("%s@%s", replace.Path, req.Mod.Version)
			}
		} else {
			moduleVersions[req.Mod.Path] = fmt.Sprintf("%s@%s", req.Mod.Path, req.Mod.Version)
		}
	}
	return moduleVersions
}

func (cd *changeDetector) readGoModFile(fullPath string) (*modfile.File, error) {
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}
	return parseGoModFile(fullPath, data)
}

func (cd *changeDetector) readGoModFileFromGit(fullPath string, gitRevision string) (*modfile.File, error) {
	if !strings.HasPrefix(fullPath, cd.gitRootFullPath) {
		return nil, errors.New("go.mod file is not in the git project")
	}
	gitPath := strings.TrimPrefix(fullPath, cd.gitRootFullPath+"/")
	data, exists, err := git.ReadFile(cd.context, cd.gitRootFullPath, gitRevision, gitPath)
	if err != nil || !exists {
		return nil, err
	}
	return parseGoModFile(fullPath, data)
}

func parseGoModFile(fullPath string, data []byte) (*modfile.File, error) {
	mod, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to parse go.mod file %s: %w", fullPath, err)
	}
	return mod, nil
}
