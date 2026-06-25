package render

// RenderMetadata carries collect-stage facts to the render stage.
type RenderMetadata struct {
	Truncated     bool
	TruncatedFiles []string
	ExcludedFiles []string
	FilesAnalyzed int
	FilesTotal    int
	BudgetChars   int
}
