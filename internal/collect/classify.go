package collect

import (
	"path/filepath"
	"strings"
)

// ClassifyFile determines the FileCategory for a given file path.
func ClassifyFile(path string) FileCategory {
	base := filepath.Base(path)
	dir := filepath.Dir(path)
	ext := filepath.Ext(path)

	if isLockfile(base) {
		return CategoryLockfile
	}
	if isGenerated(path, base) {
		return CategoryGenerated
	}
	if isVendor(path) {
		return CategoryVendor
	}
	if isTest(path, base) {
		return CategoryTest
	}
	if isDocs(path, base, ext) {
		return CategoryDocs
	}
	if isConfig(path, base, dir, ext) {
		return CategoryConfig
	}
	return CategorySource
}

func isLockfile(base string) bool {
	lockfiles := []string{
		"package-lock.json", "yarn.lock", "pnpm-lock.yaml",
		"Cargo.lock", "Gemfile.lock", "poetry.lock",
		"go.sum", "composer.lock", "Pipfile.lock",
		"bun.lockb", "uv.lock",
	}
	for _, lf := range lockfiles {
		if base == lf {
			return true
		}
	}
	return false
}

func isGenerated(path, base string) bool {
	if strings.HasSuffix(base, ".gen.go") ||
		strings.HasSuffix(base, ".gen.ts") ||
		strings.HasSuffix(base, ".pb.go") ||
		strings.HasSuffix(base, ".generated.ts") {
		return true
	}
	generatedDirs := []string{"dist/", "build/", "__generated__/", ".next/"}
	for _, d := range generatedDirs {
		if strings.HasPrefix(path, d) {
			return true
		}
	}
	return false
}

func isVendor(path string) bool {
	vendorPrefixes := []string{"vendor/", "node_modules/", "third_party/"}
	for _, v := range vendorPrefixes {
		if strings.HasPrefix(path, v) {
			return true
		}
	}
	return false
}

func isTest(path, base string) bool {
	testPatterns := []string{"_test.go", ".test.ts", ".test.js", ".spec.ts", ".spec.js", ".test.tsx", ".spec.tsx"}
	for _, p := range testPatterns {
		if strings.HasSuffix(base, p) {
			return true
		}
	}
	testDirs := []string{"test/", "tests/", "testdata/", "__tests__/", "spec/"}
	for _, d := range testDirs {
		if strings.HasPrefix(path, d) || strings.Contains(path, "/"+d) {
			return true
		}
	}
	if strings.HasPrefix(base, "test_") && strings.HasSuffix(base, ".py") {
		return true
	}
	return false
}

func isDocs(path, base, ext string) bool {
	docExts := []string{".md", ".rst", ".txt", ".adoc"}
	for _, de := range docExts {
		if ext == de {
			return true
		}
	}
	docDirs := []string{"docs/", "doc/", "documentation/"}
	for _, d := range docDirs {
		if strings.HasPrefix(path, d) || strings.Contains(path, "/"+d) {
			return true
		}
	}
	docFiles := []string{"LICENSE", "NOTICE", "AUTHORS", "CONTRIBUTORS"}
	for _, df := range docFiles {
		if base == df {
			return true
		}
	}
	return false
}

func isConfig(path, base, dir, ext string) bool {
	configExts := []string{".yml", ".yaml", ".toml", ".ini", ".cfg", ".conf"}
	for _, ce := range configExts {
		if ext == ce {
			return true
		}
	}
	configFiles := []string{
		".gitignore", ".gitattributes", ".editorconfig",
		"Makefile", "Dockerfile", "docker-compose.yml",
		".eslintrc.js", ".prettierrc", "tsconfig.json",
		"pyproject.toml", "setup.cfg", "setup.py",
		"Cargo.toml", "go.mod",
		".github", ".gitlab-ci.yml",
	}
	for _, cf := range configFiles {
		if base == cf || strings.HasPrefix(path, cf) {
			return true
		}
	}
	if strings.HasPrefix(dir, ".github") || strings.HasPrefix(path, ".github/") {
		return true
	}
	if base == "package.json" || base == "manifest.json" {
		return true
	}
	return false
}
