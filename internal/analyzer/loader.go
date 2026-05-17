package analyzer

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

type LoadRequest struct {
	RootPath     string
	IncludeTests bool
}

type PackageInfo struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	PkgPath string   `json:"pkg_path"`
	Files   []string `json:"files"`
}

type AnalyzeWarning struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	PackageID string `json:"package_id,omitempty"`
}

type LoadResult struct {
	Packages []PackageInfo    `json:"packages"`
	Warnings []AnalyzeWarning `json:"warnings"`
}

func LoadPackages(req LoadRequest) (LoadResult, error) {
	if strings.TrimSpace(req.RootPath) == "" {
		return LoadResult{}, errors.New("root path is required")
	}

	absRoot, err := filepath.Abs(req.RootPath)
	if err != nil {
		return LoadResult{}, fmt.Errorf("resolve absolute path: %w", err)
	}

	info, err := os.Stat(absRoot)
	if err != nil {
		return LoadResult{}, fmt.Errorf("stat root path %q: %w", req.RootPath, err)
	}
	if !info.IsDir() {
		return LoadResult{}, fmt.Errorf("root path %q is not a directory", req.RootPath)
	}

	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedSyntax |
			packages.NeedTypes |
			packages.NeedTypesInfo |
			packages.NeedModule,
		Dir:   absRoot,
		Tests: req.IncludeTests,
	}

	pkgs, loadErr := packages.Load(cfg, "./...")
	result := LoadResult{
		Packages: make([]PackageInfo, 0, len(pkgs)),
		Warnings: make([]AnalyzeWarning, 0),
	}

	if loadErr != nil {
		result.Warnings = append(result.Warnings, AnalyzeWarning{
			Code:    "packages-load-error",
			Message: loadErr.Error(),
		})
	}

	for _, pkg := range pkgs {
		for _, pkgErr := range pkg.Errors {
			result.Warnings = append(result.Warnings, AnalyzeWarning{
				Code:      "package-error",
				Message:   pkgErr.Error(),
				PackageID: pkg.ID,
			})
		}

		files := make([]string, 0, len(pkg.GoFiles))
		for _, file := range pkg.GoFiles {
			if !req.IncludeTests && strings.HasSuffix(file, "_test.go") {
				continue
			}
			rel, relErr := filepath.Rel(absRoot, file)
			if relErr != nil {
				files = append(files, filepath.ToSlash(file))
				continue
			}
			files = append(files, filepath.ToSlash(rel))
		}
		sort.Strings(files)

		result.Packages = append(result.Packages, PackageInfo{
			ID:      pkg.ID,
			Name:    pkg.Name,
			PkgPath: pkg.PkgPath,
			Files:   files,
		})
	}

	return result, nil
}
