package collect

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/bluekeyes/go-gitdiff/gitdiff"
)

// ParseDiffFromFile reads a unified diff from a file path.
func ParseDiffFromFile(path string) ([]*gitdiff.File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open diff file: %w", err)
	}
	defer f.Close()
	return parseDiff(f)
}

// ParseDiffFromReader reads a unified diff from an io.Reader.
func ParseDiffFromReader(r io.Reader) ([]*gitdiff.File, error) {
	return parseDiff(r)
}

// ParseDiffFromGit runs git diff between base and head refs and parses the result.
func ParseDiffFromGit(base, head string) ([]*gitdiff.File, error) {
	if err := validateRef(base); err != nil {
		return nil, fmt.Errorf("invalid base ref: %w", err)
	}
	if err := validateRef(head); err != nil {
		return nil, fmt.Errorf("invalid head ref: %w", err)
	}

	if base == "" {
		mb, err := mergeBase(head)
		if err != nil {
			return nil, fmt.Errorf("determine merge-base: %w", err)
		}
		base = mb
	}

	cmd := exec.Command("git", "diff", base+".."+head)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff %s..%s: %w", base, head, err)
	}

	return parseDiff(strings.NewReader(string(out)))
}

func validateRef(ref string) error {
	if ref != "" && strings.HasPrefix(ref, "-") {
		return fmt.Errorf("ref %q must not start with '-'", ref)
	}
	return nil
}

func mergeBase(head string) (string, error) {
	cmd := exec.Command("git", "merge-base", "--", "main", head)
	out, err := cmd.Output()
	if err != nil {
		cmd2 := exec.Command("git", "merge-base", "--", "master", head)
		out, err = cmd2.Output()
		if err != nil {
			return "", fmt.Errorf("could not find merge-base (tried main and master). Use --base to specify the base ref explicitly: %w", err)
		}
	}
	return strings.TrimSpace(string(out)), nil
}

func parseDiff(r io.Reader) ([]*gitdiff.File, error) {
	files, _, err := gitdiff.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("parse diff: %w", err)
	}
	return files, nil
}

// FilesToChangedFiles converts parsed gitdiff files into our ChangedFile type.
func FilesToChangedFiles(files []*gitdiff.File) []ChangedFile {
	result := make([]ChangedFile, 0, len(files))
	for _, f := range files {
		path := filePath(f)
		added, deleted := countLines(f)
		cf := ChangedFile{
			Path:     path,
			Added:    added,
			Deleted:  deleted,
			IsBinary: f.IsBinary,
			HunkText: renderHunks(f),
		}
		result = append(result, cf)
	}
	return result
}

func filePath(f *gitdiff.File) string {
	if f.NewName != "" {
		return f.NewName
	}
	if f.OldName != "" {
		return f.OldName
	}
	return "(unknown)"
}

func countLines(f *gitdiff.File) (added, deleted int) {
	for _, frag := range f.TextFragments {
		for _, line := range frag.Lines {
			switch line.Op {
			case gitdiff.OpAdd:
				added++
			case gitdiff.OpDelete:
				deleted++
			}
		}
	}
	return
}

func renderHunks(f *gitdiff.File) string {
	if f.IsBinary {
		return "(binary file)"
	}
	var sb strings.Builder
	for _, frag := range f.TextFragments {
		fmt.Fprintf(&sb, "@@ -%d,%d +%d,%d @@ %s\n",
			frag.OldPosition, frag.OldLines,
			frag.NewPosition, frag.NewLines,
			frag.Comment)
		for _, line := range frag.Lines {
			switch line.Op {
			case gitdiff.OpAdd:
				sb.WriteString("+" + line.Line)
			case gitdiff.OpDelete:
				sb.WriteString("-" + line.Line)
			default:
				sb.WriteString(" " + line.Line)
			}
			if !strings.HasSuffix(line.Line, "\n") {
				sb.WriteString("\n")
			}
		}
	}
	return sb.String()
}
