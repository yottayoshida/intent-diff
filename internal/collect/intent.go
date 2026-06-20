package collect

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// ReadIntentFromFile reads PR description text from a markdown file.
func ReadIntentFromFile(path string) (string, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", fmt.Errorf("read intent file: %w", err)
	}
	return strings.TrimSpace(string(data)), "file:" + path, nil
}

// ReadIntentFromPRJSON reads title and body from gh pr view --json output.
func ReadIntentFromPRJSON(path string) (string, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", fmt.Errorf("read pr-json file: %w", err)
	}

	var pr struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}
	if err := json.Unmarshal(data, &pr); err != nil {
		return "", "", fmt.Errorf("parse pr-json: %w", err)
	}

	intent := pr.Title
	if pr.Body != "" {
		intent += "\n\n" + pr.Body
	}

	return strings.TrimSpace(intent), "pr-json:" + path, nil
}

// ReadIntentFromStdin reads PR description from stdin.
func ReadIntentFromStdin(r io.Reader) (string, string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return "", "", fmt.Errorf("read stdin: %w", err)
	}
	return strings.TrimSpace(string(data)), "stdin", nil
}

const minIntentLength = 10

// IsEmptyIntent returns true if the intent text is too short to be useful.
func IsEmptyIntent(intent string) bool {
	return len(strings.TrimSpace(intent)) < minIntentLength
}

// IsCopyPasteIntent returns true if the intent is suspiciously similar to the diff.
func IsCopyPasteIntent(intent, diffText string) bool {
	if len(intent) == 0 || len(diffText) == 0 {
		return false
	}
	intentNorm := normalize(intent)
	diffNorm := normalize(diffText)
	if len(intentNorm) == 0 || len(diffNorm) == 0 {
		return false
	}

	shorter, longer := intentNorm, diffNorm
	if len(shorter) > len(longer) {
		shorter, longer = longer, shorter
	}
	if len(shorter) == 0 {
		return false
	}

	overlap := longestCommonSubstringLen(shorter, longer)
	ratio := float64(overlap) / float64(len(shorter))
	return ratio > 0.8
}

func normalize(s string) string {
	s = strings.ToLower(s)
	s = strings.Join(strings.Fields(s), " ")
	return s
}

func longestCommonSubstringLen(a, b string) int {
	ra := []rune(a)
	rb := []rune(b)

	if len(ra) > 500 {
		ra = ra[:500]
	}
	if len(rb) > 500 {
		rb = rb[:500]
	}

	m, n := len(ra), len(rb)
	prev := make([]int, n+1)
	curr := make([]int, n+1)
	maxLen := 0

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if ra[i-1] == rb[j-1] {
				curr[j] = prev[j-1] + 1
				if curr[j] > maxLen {
					maxLen = curr[j]
				}
			} else {
				curr[j] = 0
			}
		}
		prev, curr = curr, prev
		for j := range curr {
			curr[j] = 0
		}
	}
	return maxLen
}
