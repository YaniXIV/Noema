package evaluate

import "fmt"

type ConstraintRule struct {
	ID                 string
	AllowedMaxSeverity int
}

func enabledConstraints(cfg PolicyConfig) (map[string]ConstraintRule, error) {
	out := make(map[string]ConstraintRule)
	for _, c := range cfg.Constraints {
		if !c.Enabled {
			continue
		}
		if _, exists := out[c.ID]; exists {
			return nil, fmt.Errorf("duplicate constraint id: %s", c.ID)
		}
		out[c.ID] = ConstraintRule{ID: c.ID, AllowedMaxSeverity: c.MaxAllowed}
	}
	return out, nil
}

func computePolicyResult(out EvaluationResult, cfg PolicyConfig) (overallPass bool, maxSeverity int, policyThreshold int) {
	enabled, _ := enabledConstraints(cfg)
	overallPass = true
	if len(enabled) == 0 {
		return true, 0, 0
	}
	policyThreshold = 2
	for _, rule := range enabled {
		if rule.AllowedMaxSeverity < policyThreshold {
			policyThreshold = rule.AllowedMaxSeverity
		}
	}
	byID := make(map[string]EvalResultItem, len(out.Results))
	for _, r := range out.Results {
		byID[r.ID] = r
	}
	for id, rule := range enabled {
		r := byID[id]
		if r.Severity > maxSeverity {
			maxSeverity = r.Severity
		}
		if r.Severity > rule.AllowedMaxSeverity {
			overallPass = false
		}
	}
	return overallPass, maxSeverity, policyThreshold
}
