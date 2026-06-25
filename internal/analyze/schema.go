package analyze

// AnalysisResult is the top-level structured output from the LLM.
type AnalysisResult struct {
	Version                  string          `json:"version"`
	Alignment                Alignment       `json:"alignment"`
	ClaimedIntent            string          `json:"claimed_intent"`
	ImplementationEvidence   []Evidence      `json:"implementation_evidence"`
	BehaviorImpactHypotheses []Hypothesis    `json:"behavior_impact_hypotheses"`
	Mismatches               []Mismatch      `json:"mismatches"`
	AttentionMap             []AttentionItem `json:"attention_map"`
	SuggestedPRDescription   string          `json:"suggested_pr_description"`
}

// Alignment represents the overall alignment assessment.
type Alignment struct {
	Grade               string  `json:"grade"`
	Score               float64 `json:"score"`
	Confidence          string  `json:"confidence"`
	HighestRiskCategory string  `json:"highest_risk_category"`
}

// Evidence represents a piece of implementation evidence from the diff.
type Evidence struct {
	File        string `json:"file"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

// Hypothesis represents a behavior-impact hypothesis that needs verification.
type Hypothesis struct {
	Description      string   `json:"description"`
	AffectedFiles    []string `json:"affected_files"`
	VerificationHint string   `json:"verification_hint"`
}

// MismatchCategory enumerates the 9-category taxonomy.
type MismatchCategory string

const (
	MismatchScope               MismatchCategory = "scope"
	MismatchContract            MismatchCategory = "contract"
	MismatchRisk                MismatchCategory = "risk"
	MismatchTest                MismatchCategory = "test"
	MismatchIntentUnderSpec     MismatchCategory = "intent_under_specification"
	MismatchNonCodeImpact       MismatchCategory = "non_code_impact"
	MismatchBehavioralAmbiguity MismatchCategory = "behavioral_ambiguity"
	MismatchDocumentation       MismatchCategory = "documentation"
	MismatchDependencyRisk      MismatchCategory = "dependency_risk"
)

// Mismatch represents a single detected mismatch.
type Mismatch struct {
	Category          MismatchCategory `json:"category"`
	Severity          string           `json:"severity"`
	Confidence        string           `json:"confidence"`
	Claim             string           `json:"claim"`
	Observation       string           `json:"observation"`
	Evidence          []string         `json:"evidence"`
	RecommendedAction string           `json:"recommended_action"`
}

// AttentionItem represents a file or area the reviewer should focus on.
type AttentionItem struct {
	File     string `json:"file"`
	Reason   string `json:"reason"`
	Priority string `json:"priority"`
}

// JSONSchema returns the JSON schema string for structured output.
func JSONSchema() string {
	return `{
  "type": "object",
  "required": ["version", "alignment", "claimed_intent", "implementation_evidence", "behavior_impact_hypotheses", "mismatches", "attention_map", "suggested_pr_description"],
  "properties": {
    "version": {
      "type": "string",
      "enum": ["0.1"],
      "description": "Schema version. Uses enum (not const) for broader structured-output tool compatibility."
    },
    "alignment": {
      "type": "object",
      "required": ["grade", "score", "confidence", "highest_risk_category"],
      "properties": {
        "grade": {
          "type": "string",
          "enum": ["A", "B", "C", "D", "E"],
          "description": "A=Well-aligned: no material mismatches. B=Minor omissions: small gaps but no risk-bearing changes undocumented. C=Material omissions: risk-bearing changes not mentioned or scope significantly exceeds claims. D=Significant mismatches: explicit claims contradicted by diff evidence. E=Critical mismatches: safety/auth/data changes undocumented or explicit non-goals violated."
        },
        "score": {
          "type": "number",
          "minimum": 0,
          "maximum": 1,
          "description": "Alignment score from 0 (completely misaligned) to 1 (perfectly aligned)"
        },
        "confidence": {
          "type": "string",
          "enum": ["high", "medium", "low"],
          "description": "Confidence in the assessment. low if intent is empty/vague or diff was truncated."
        },
        "highest_risk_category": {
          "type": "string",
          "description": "The mismatch category with the highest severity, or 'none' if no mismatches."
        }
      },
      "additionalProperties": false
    },
    "claimed_intent": {
      "type": "string",
      "description": "Summary of what the PR description claims this change does."
    },
    "implementation_evidence": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["file", "description", "category"],
        "properties": {
          "file": { "type": "string" },
          "description": { "type": "string" },
          "category": { "type": "string" }
        },
        "additionalProperties": false
      }
    },
    "behavior_impact_hypotheses": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["description", "affected_files", "verification_hint"],
        "properties": {
          "description": { "type": "string" },
          "affected_files": { "type": "array", "items": { "type": "string" } },
          "verification_hint": { "type": "string" }
        },
        "additionalProperties": false
      }
    },
    "mismatches": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["category", "severity", "confidence", "claim", "observation", "evidence", "recommended_action"],
        "properties": {
          "category": {
            "type": "string",
            "enum": ["scope", "contract", "risk", "test", "intent_under_specification", "non_code_impact", "behavioral_ambiguity", "documentation", "dependency_risk"]
          },
          "severity": {
            "type": "string",
            "enum": ["critical", "high", "medium", "low", "info"]
          },
          "confidence": {
            "type": "string",
            "enum": ["high", "medium", "low"]
          },
          "claim": { "type": "string" },
          "observation": { "type": "string" },
          "evidence": { "type": "array", "items": { "type": "string" } },
          "recommended_action": { "type": "string" }
        },
        "additionalProperties": false
      }
    },
    "attention_map": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["file", "reason", "priority"],
        "properties": {
          "file": { "type": "string" },
          "reason": { "type": "string" },
          "priority": {
            "type": "string",
            "enum": ["critical", "high", "medium", "low"]
          }
        },
        "additionalProperties": false
      }
    },
    "suggested_pr_description": {
      "type": "string",
      "description": "An improved PR description that accurately reflects the implementation."
    }
  },
  "additionalProperties": false
}`
}
