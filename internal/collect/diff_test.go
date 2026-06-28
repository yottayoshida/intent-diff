package collect

import (
	"strings"
	"testing"
)

const sampleDiff = `diff --git a/main.go b/main.go
index 1234567..abcdefg 100644
--- a/main.go
+++ b/main.go
@@ -1,3 +1,4 @@
 package main

+import "fmt"
 func main() {
`

func TestParseDiffFromReader(t *testing.T) {
	files, err := ParseDiffFromReader(strings.NewReader(sampleDiff))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	if files[0].NewName != "main.go" {
		t.Errorf("expected main.go, got %s", files[0].NewName)
	}
}

func TestFilesToChangedFiles(t *testing.T) {
	files, err := ParseDiffFromReader(strings.NewReader(sampleDiff))
	if err != nil {
		t.Fatal(err)
	}

	changed := FilesToChangedFiles(files)
	if len(changed) != 1 {
		t.Fatalf("expected 1 changed file, got %d", len(changed))
	}
	if changed[0].Path != "main.go" {
		t.Errorf("expected path main.go, got %s", changed[0].Path)
	}
	if changed[0].Added != 1 {
		t.Errorf("expected 1 added line, got %d", changed[0].Added)
	}
	if changed[0].HunkText == "" {
		t.Error("expected non-empty hunk text")
	}
}

func TestParseDiffFromReader_Empty(t *testing.T) {
	files, err := ParseDiffFromReader(strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 0 {
		t.Errorf("expected 0 files for empty diff, got %d", len(files))
	}
}

func TestValidateRef(t *testing.T) {
	tests := []struct {
		ref     string
		wantErr bool
	}{
		{"", false},
		{"HEAD", false},
		{"main", false},
		{"abc123def", false},
		{"refs/heads/feature/foo", false},
		{"v0.1.0", false},
		{"HEAD~3", false},
		{"HEAD^2", false},
		{"origin/main", false},
		{"-flag", true},
		{"ref with space", true},
		{"ref;inject", true},
		{"ref|pipe", true},
		{"ref$(cmd)", true},
		{"ref`cmd`", true},
		{"ref&bg", true},
		{"ref\nnewline", true},
	}
	for _, tt := range tests {
		t.Run(tt.ref, func(t *testing.T) {
			err := validateRef(tt.ref)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRef(%q) error = %v, wantErr %v", tt.ref, err, tt.wantErr)
			}
		})
	}
}
