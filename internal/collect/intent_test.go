package collect

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadIntentFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pr.md")
	if err := os.WriteFile(path, []byte("# Fix auth bug\n\nThis PR fixes the session timeout."), 0o644); err != nil {
		t.Fatal(err)
	}

	intent, source, err := ReadIntentFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(intent, "Fix auth bug") {
		t.Errorf("expected intent to contain 'Fix auth bug', got %q", intent)
	}
	if !strings.HasPrefix(source, "file:") {
		t.Errorf("expected source to start with 'file:', got %q", source)
	}
}

func TestReadIntentFromPRJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pr.json")
	if err := os.WriteFile(path, []byte(`{"title":"Fix auth","body":"Fixes session handling."}`), 0o644); err != nil {
		t.Fatal(err)
	}

	intent, source, err := ReadIntentFromPRJSON(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(intent, "Fix auth") {
		t.Errorf("expected intent to contain 'Fix auth', got %q", intent)
	}
	if !strings.Contains(intent, "Fixes session handling") {
		t.Errorf("expected intent to contain body, got %q", intent)
	}
	if !strings.HasPrefix(source, "pr-json:") {
		t.Errorf("expected source to start with 'pr-json:', got %q", source)
	}
}

func TestReadIntentFromStdin(t *testing.T) {
	r := strings.NewReader("Refactor middleware")
	intent, source, err := ReadIntentFromStdin(r)
	if err != nil {
		t.Fatal(err)
	}
	if intent != "Refactor middleware" {
		t.Errorf("got %q", intent)
	}
	if source != "stdin" {
		t.Errorf("got source %q", source)
	}
}

func TestIsEmptyIntent(t *testing.T) {
	tests := []struct {
		intent string
		empty  bool
	}{
		{"", true},
		{"   ", true},
		{"hi", true},
		{"short txt", true},
		{"This is a proper PR description", false},
	}
	for _, tt := range tests {
		if got := IsEmptyIntent(tt.intent); got != tt.empty {
			t.Errorf("IsEmptyIntent(%q) = %v, want %v", tt.intent, got, tt.empty)
		}
	}
}

func TestIsCopyPasteIntent(t *testing.T) {
	diff := "func main() {\n\tfmt.Println(\"hello\")\n}"
	if IsCopyPasteIntent(diff, diff) != true {
		t.Error("identical text should be detected as copy-paste")
	}
	if IsCopyPasteIntent("Refactor auth module", diff) != false {
		t.Error("different text should not be copy-paste")
	}
	if IsCopyPasteIntent("", diff) != false {
		t.Error("empty intent should not be copy-paste")
	}
}
