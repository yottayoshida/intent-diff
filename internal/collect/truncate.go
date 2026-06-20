package collect

import (
	"sort"
)

const DefaultBudgetChars = 100_000

// TruncateFiles applies the diff truncation strategy:
// 1. Exclude generated/vendor/binary/lockfile
// 2. Sort remaining by risk (high risk first)
// 3. Truncate low-risk hunks to fit within char budget
func TruncateFiles(files []ChangedFile, budget int) (included []ChangedFile, excluded []string, truncated bool) {
	if budget <= 0 {
		budget = DefaultBudgetChars
	}

	var incl []ChangedFile
	for _, f := range files {
		switch f.Category {
		case CategoryGenerated, CategoryVendor, CategoryBinary, CategoryLockfile:
			excluded = append(excluded, f.Path)
		default:
			if f.IsBinary {
				excluded = append(excluded, f.Path)
			} else {
				incl = append(incl, f)
			}
		}
	}

	sort.Slice(incl, func(i, j int) bool {
		ri, rj := RiskOrder(incl[i].Risk), RiskOrder(incl[j].Risk)
		if ri != rj {
			return ri < rj
		}
		return (incl[i].Added + incl[i].Deleted) > (incl[j].Added + incl[j].Deleted)
	})

	totalChars := 0
	for i := range incl {
		totalChars += len(incl[i].HunkText)
	}

	if totalChars <= budget {
		return incl, excluded, false
	}

	usedChars := 0
	for i := range incl {
		hunkLen := len(incl[i].HunkText)
		if usedChars+hunkLen <= budget {
			usedChars += hunkLen
			continue
		}

		remaining := budget - usedChars
		if remaining > 0 {
			incl[i].HunkText = incl[i].HunkText[:remaining] + "\n... (truncated)"
			incl[i].Truncated = true
			usedChars = budget
		} else {
			incl[i].HunkText = "(truncated — over budget)"
			incl[i].Truncated = true
		}
		truncated = true
	}

	return incl, excluded, truncated
}
