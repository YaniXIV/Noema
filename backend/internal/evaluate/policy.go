package evaluate

import "fmt"

type ConstraintRule struct {
	ID                 string
	AllowedMaxSeverity int
}

func enabledConstraints(spec Spec) (map[string]ConstraintRule, error) {
	out := make(map[string]ConstraintRule)
	for _, c := range spec.Constraints {
		if !c.Enabled {
			continue
		}
		if _, exists := out[c.ID]; exists {
			return nil, fmt.Errorf("duplicate constraint id: %s", c.ID)
		}
		out[c.ID] = ConstraintRule{ID: c.ID, AllowedMaxSeverity: c.AllowedMaxSeverity}
	}
	for _, c := range spec.CustomConstraints {
		if !c.Enabled {
			continue
		}
		if _, exists := out[c.ID]; exists {
			return nil, fmt.Errorf("duplicate constraint id: %s", c.ID)
		}
		out[c.ID] = ConstraintRule{ID: c.ID, AllowedMaxSeverity: c.AllowedMaxSeverity}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("at least one constraint must be enabled")
	}
	return out, nil
}

func computePolicyResult(out EvalOutput, enabled map[string]ConstraintRule) (overallPass bool, maxSeverity int, policyThreshold int) {
	policyThreshold = 2
	for _, rule := range enabled {
		if rule.AllowedMaxSeverity < policyThreshold {
			policyThreshold = rule.AllowedMaxSeverity
		}
	}
	overallPass = true
	for _, c := range out.Constraints {
		if c.Severity > maxSeverity {
			maxSeverity = c.Severity
		}
		rule, ok := enabled[c.ID]
		if ok && c.Severity > rule.AllowedMaxSeverity {
			overallPass = false
		}
	}
	return overallPass, maxSeverity, policyThreshold
}
