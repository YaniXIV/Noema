package evaluate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

type EvalConstraintResult struct {
	ID        string `json:"id"`
	Severity  int    `json:"severity"`
	Rationale string `json:"rationale"`
}

type EvalOutput struct {
	SchemaVersion int                    `json:"schema_version"`
	Constraints   []EvalConstraintResult `json:"constraints"`
	MaxSeverity   int                    `json:"max_severity"`
	Confidence    *float64               `json:"confidence,omitempty"`
}

func parseEvalOutput(raw string) (EvalOutput, error) {
	dec := json.NewDecoder(bytes.NewBufferString(raw))
	dec.DisallowUnknownFields()
	var out EvalOutput
	if err := dec.Decode(&out); err != nil {
		return EvalOutput{}, fmt.Errorf("invalid gemini JSON output")
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return EvalOutput{}, fmt.Errorf("invalid gemini JSON output")
	}
	if out.SchemaVersion != 1 {
		return EvalOutput{}, fmt.Errorf("invalid gemini JSON output")
	}
	if len(out.Constraints) == 0 {
		return EvalOutput{}, fmt.Errorf("invalid gemini JSON output")
	}
	if out.MaxSeverity < 0 || out.MaxSeverity > 2 {
		return EvalOutput{}, fmt.Errorf("invalid gemini JSON output")
	}
	if out.Confidence != nil {
		if *out.Confidence < 0 || *out.Confidence > 1 {
			return EvalOutput{}, fmt.Errorf("invalid gemini JSON output")
		}
	}
	return out, nil
}

func validateEvalOutput(out EvalOutput, enabled map[string]ConstraintRule) error {
	seen := map[string]bool{}
	max := 0
	for _, c := range out.Constraints {
		if c.ID == "" || c.Rationale == "" {
			return fmt.Errorf("invalid gemini JSON output")
		}
		if c.Severity < 0 || c.Severity > 2 {
			return fmt.Errorf("invalid gemini JSON output")
		}
		if _, ok := enabled[c.ID]; !ok {
			return fmt.Errorf("invalid gemini JSON output")
		}
		if seen[c.ID] {
			return fmt.Errorf("invalid gemini JSON output")
		}
		seen[c.ID] = true
		if c.Severity > max {
			max = c.Severity
		}
	}
	if max != out.MaxSeverity {
		return fmt.Errorf("invalid gemini JSON output")
	}
	if len(seen) != len(enabled) {
		return fmt.Errorf("invalid gemini JSON output")
	}
	return nil
}
