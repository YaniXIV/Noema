package evaluate

import (
	"strings"
	"testing"
)

func TestBuildSystemPrompt_IncludesStrictJSONAndReasoning(t *testing.T) {
	prompt := buildSystemPrompt()
	if !strings.Contains(prompt, "Return ONLY valid JSON") {
		t.Fatalf("expected system prompt to require JSON-only output")
	}
	if !strings.Contains(prompt, "reason about the dataset") {
		t.Fatalf("expected system prompt to require dataset reasoning")
	}
}

func TestBuildUserPrompt_IncludesStrictSchemaAndDataset(t *testing.T) {
	cfg := PolicyConfig{
		PolicyVersion: "noema_policy_v1",
		Constraints: []PolicyConstraint{
			{ID: "pii_exposure_risk", Enabled: true, MaxAllowed: 1},
		},
	}
	prompt := buildUserPrompt(cfg, []byte(`{"items":[{"id":"1","text":"hello"}]}`), nil)
	if !strings.Contains(prompt, "\"eval_version\":\"noema_eval_v1\"") {
		t.Fatalf("expected user prompt to include eval_version schema")
	}
	if !strings.Contains(prompt, "Include one result per constraint id provided") {
		t.Fatalf("expected user prompt to enforce per-id results")
	}
	if !strings.Contains(prompt, "Dataset JSON") {
		t.Fatalf("expected user prompt to include dataset content")
	}
}

func TestEvalResponseSchema_MatchesContract(t *testing.T) {
	schema := evalResponseSchema()
	required, ok := schema["required"].([]any)
	if !ok {
		t.Fatalf("expected required field list")
	}
	foundEval := false
	foundResults := false
	for _, v := range required {
		if v == "eval_version" {
			foundEval = true
		}
		if v == "results" {
			foundResults = true
		}
	}
	if !foundEval || !foundResults {
		t.Fatalf("expected schema to require eval_version and results")
	}
	props := schema["properties"].(map[string]any)
	results := props["results"].(map[string]any)
	items := results["items"].(map[string]any)
	req := items["required"].([]any)
	hasID := false
	hasSeverity := false
	for _, v := range req {
		if v == "id" {
			hasID = true
		}
		if v == "severity" {
			hasSeverity = true
		}
	}
	if !hasID || !hasSeverity {
		t.Fatalf("expected results items to require id and severity")
	}
}

func TestValidateEvaluationResult_RejectsMissingIDs(t *testing.T) {
	cfg := PolicyConfig{
		PolicyVersion: "noema_policy_v1",
		Constraints: []PolicyConstraint{
			{ID: "a", Enabled: true, MaxAllowed: 1},
			{ID: "b", Enabled: true, MaxAllowed: 1},
		},
	}
	out := EvaluationResult{
		EvalVersion: "noema_eval_v1",
		Results: []EvalResultItem{
			{ID: "a", Severity: 0},
		},
	}
	if err := validateEvaluationResult(out, cfg); err == nil {
		t.Fatalf("expected validation error for missing ids")
	}
}
