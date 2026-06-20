package collect

// FileCategory classifies a changed file by its role in the project.
type FileCategory string

const (
	CategorySource    FileCategory = "source"
	CategoryTest      FileCategory = "test"
	CategoryConfig    FileCategory = "config"
	CategoryDocs      FileCategory = "docs"
	CategoryGenerated FileCategory = "generated"
	CategoryVendor    FileCategory = "vendor"
	CategoryBinary    FileCategory = "binary"
	CategoryLockfile  FileCategory = "lockfile"
)

// RiskTag indicates the risk level of a changed file based on its path.
type RiskTag string

const (
	RiskAuth  RiskTag = "auth"
	RiskAPI   RiskTag = "api"
	RiskData  RiskTag = "data"
	RiskInfra RiskTag = "infra"
	RiskOther RiskTag = "other"
)

// RiskOrder returns a numeric priority for sorting (lower = higher risk).
func RiskOrder(r RiskTag) int {
	switch r {
	case RiskAuth:
		return 0
	case RiskAPI:
		return 1
	case RiskData:
		return 2
	case RiskInfra:
		return 3
	default:
		return 4
	}
}

// ChangedFile represents a single file from the diff with classification metadata.
type ChangedFile struct {
	Path      string       `json:"path"`
	Category  FileCategory `json:"category"`
	Risk      RiskTag      `json:"risk"`
	Added     int          `json:"added"`
	Deleted   int          `json:"deleted"`
	IsBinary  bool         `json:"is_binary"`
	HunkText  string       `json:"hunk_text,omitempty"`
	Truncated bool         `json:"truncated,omitempty"`
}

// CollectResult is the output of the Collect stage, ready for the Analyze stage.
type CollectResult struct {
	Intent        string        `json:"intent"`
	IntentSource  string        `json:"intent_source"`
	Files         []ChangedFile `json:"files"`
	TotalAdded    int           `json:"total_added"`
	TotalDeleted  int           `json:"total_deleted"`
	ExcludedFiles []string      `json:"excluded_files,omitempty"`
	Truncated     bool          `json:"truncated"`
	DiffChars     int           `json:"diff_chars"`
	BudgetChars   int           `json:"budget_chars"`
}
