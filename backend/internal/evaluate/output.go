package evaluate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

type EvalResultItem struct {
	ID         string   `json:"id"`
	Severity   int      `json:"severity"`
	Confidence *float64 `json:"confidence,omitempty"`
	Rationale  string   `json:"rationale,omitempty"`
}

type EvaluationResult struct {
	EvalVersion string           `json:"eval_version"`
	Results     []EvalResultItem `json:"results"`
}

func parseEvaluationResult(raw string) (EvaluationResult, error) {
	dec := json.NewDecoder(bytes.NewBufferString(raw))
	dec.DisallowUnknownFields()
	var out EvaluationResult
	if err := dec.Decode(&out); err != nil {
		return EvaluationResult{}, fmt.Errorf("invalid gemini JSON output")
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return EvaluationResult{}, fmt.Errorf("invalid gemini JSON output")
	}
	if out.EvalVersion != "noema_eval_v1" {
		return EvaluationResult{}, fmt.Errorf("invalid gemini JSON output")
	}
	if len(out.Results) == 0 {
		return EvaluationResult{}, fmt.Errorf("invalid gemini JSON output")
	}
	return out, nil
}

func validateEvaluationResult(out EvaluationResult, cfg PolicyConfig) error {
	byID := make(map[string]PolicyConstraint, len(cfg.Constraints))
	for _, c := range cfg.Constraints {
		byID[c.ID] = c
	}
	seen := map[string]bool{}
	for _, r := range out.Results {
		if r.ID == "" {
			return fmt.Errorf("invalid gemini JSON output")
		}
		if r.Severity < 0 || r.Severity > 2 {
			return fmt.Errorf("invalid gemini JSON output")
		}
		if r.Confidence != nil {
			if *r.Confidence < 0 || *r.Confidence > 1 {
				return fmt.Errorf("invalid gemini JSON output")
			}
		}
		if _, ok := byID[r.ID]; !ok {
			return fmt.Errorf("invalid gemini JSON output")
		}
		if seen[r.ID] {
			return fmt.Errorf("invalid gemini JSON output")
		}
		seen[r.ID] = true
	}
	if len(seen) != len(cfg.Constraints) {
		return fmt.Errorf("invalid gemini JSON output")
	}
	return nil
}
