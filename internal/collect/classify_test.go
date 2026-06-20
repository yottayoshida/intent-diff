package collect

import "testing"

func TestClassifyFile(t *testing.T) {
	tests := []struct {
		path     string
		expected FileCategory
	}{
		// Source
		{"src/main.go", CategorySource},
		{"internal/handler.go", CategorySource},
		{"app.py", CategorySource},
		{"lib/utils.ts", CategorySource},

		// Test
		{"main_test.go", CategoryTest},
		{"src/handler_test.go", CategoryTest},
		{"app.test.ts", CategoryTest},
		{"component.spec.tsx", CategoryTest},
		{"tests/unit/test_auth.py", CategoryTest},
		{"test/helpers.go", CategoryTest},
		{"__tests__/app.test.js", CategoryTest},
		{"test_auth.py", CategoryTest},

		// Config
		{".github/workflows/ci.yml", CategoryConfig},
		{"Dockerfile", CategoryConfig},
		{"pyproject.toml", CategoryConfig},
		{"Cargo.toml", CategoryConfig},
		{"go.mod", CategoryConfig},
		{"package.json", CategoryConfig},
		{"tsconfig.json", CategoryConfig},
		{".gitignore", CategoryConfig},

		// Docs
		{"README.md", CategoryDocs},
		{"CHANGELOG.md", CategoryDocs},
		{"docs/guide.md", CategoryDocs},
		{"LICENSE", CategoryDocs},

		// Lockfile
		{"package-lock.json", CategoryLockfile},
		{"yarn.lock", CategoryLockfile},
		{"Cargo.lock", CategoryLockfile},
		{"go.sum", CategoryLockfile},
		{"poetry.lock", CategoryLockfile},

		// Generated
		{"dist/bundle.js", CategoryGenerated},
		{"api.pb.go", CategoryGenerated},
		{"types.gen.ts", CategoryGenerated},

		// Vendor
		{"vendor/github.com/pkg/errors/errors.go", CategoryVendor},
		{"node_modules/lodash/index.js", CategoryVendor},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := ClassifyFile(tt.path)
			if got != tt.expected {
				t.Errorf("ClassifyFile(%q) = %q, want %q", tt.path, got, tt.expected)
			}
		})
	}
}
