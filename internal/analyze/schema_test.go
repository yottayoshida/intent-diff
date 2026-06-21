package analyze

import (
	"encoding/json"
	"testing"
)

func TestJSONSchema_ValidJSON(t *testing.T) {
	schema := JSONSchema()
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(schema), &parsed); err != nil {
		t.Fatalf("JSONSchema() is not valid JSON: %v", err)
	}
}

func TestJSONSchema_VersionConstrainedToSingleValue(t *testing.T) {
	schema := JSONSchema()
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(schema), &parsed); err != nil {
		t.Fatalf("JSONSchema() is not valid JSON: %v", err)
	}

	props := parsed["properties"].(map[string]interface{})
	version := props["version"].(map[string]interface{})

	// The version field must be constrained to exactly "0.1".
	// Implementation may use "enum" or "const" — we test the guarantee, not the keyword.
	var allowed []string
	if enumVal, ok := version["enum"]; ok {
		arr, ok := enumVal.([]interface{})
		if !ok {
			t.Fatalf("enum should be an array, got %T", enumVal)
		}
		for _, v := range arr {
			s, ok := v.(string)
			if !ok {
				t.Fatalf("enum element should be string, got %T", v)
			}
			allowed = append(allowed, s)
		}
	} else if constVal, ok := version["const"]; ok {
		s, ok := constVal.(string)
		if !ok {
			t.Fatalf("const should be string, got %T", constVal)
		}
		allowed = []string{s}
	} else {
		t.Fatal("version field must be constrained via 'enum' or 'const'")
	}

	if len(allowed) != 1 || allowed[0] != "0.1" {
		t.Errorf("version must be constrained to exactly [\"0.1\"], got %v", allowed)
	}
}

func TestJSONSchema_RequiredFields(t *testing.T) {
	schema := JSONSchema()
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(schema), &parsed); err != nil {
		t.Fatalf("JSONSchema() is not valid JSON: %v", err)
	}

	required := parsed["required"].([]interface{})
	expected := []string{
		"version", "alignment", "claimed_intent",
		"implementation_evidence", "behavior_impact_hypotheses",
		"mismatches", "attention_map", "suggested_pr_description",
	}

	reqSet := make(map[string]bool)
	for _, r := range required {
		reqSet[r.(string)] = true
	}

	for _, field := range expected {
		if !reqSet[field] {
			t.Errorf("missing required field: %s", field)
		}
	}
}

func TestJSONSchema_GradeEnum(t *testing.T) {
	schema := JSONSchema()
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(schema), &parsed); err != nil {
		t.Fatalf("JSONSchema() is not valid JSON: %v", err)
	}

	props := parsed["properties"].(map[string]interface{})
	alignment := props["alignment"].(map[string]interface{})
	alignProps := alignment["properties"].(map[string]interface{})
	grade := alignProps["grade"].(map[string]interface{})

	enumVal := grade["enum"].([]interface{})
	expectedGrades := []string{"A", "B", "C", "D", "E"}

	if len(enumVal) != len(expectedGrades) {
		t.Fatalf("expected %d grades, got %d", len(expectedGrades), len(enumVal))
	}

	for i, g := range expectedGrades {
		if enumVal[i] != g {
			t.Errorf("grade[%d] = %v, want %s", i, enumVal[i], g)
		}
	}
}
