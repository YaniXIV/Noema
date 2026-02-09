package evaluate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type PolicyConstraint struct {
	ID         string `json:"id"`
	Enabled    bool   `json:"enabled"`
	MaxAllowed int    `json:"max_allowed"`
}

type PolicyConfig struct {
	PolicyVersion string             `json:"policy_version"`
	Constraints   []PolicyConstraint `json:"constraints"`
}

func parsePolicyConfig(raw string) (PolicyConfig, error) {
	dec := json.NewDecoder(bytes.NewBufferString(raw))
	dec.DisallowUnknownFields()
	var cfg PolicyConfig
	if err := dec.Decode(&cfg); err != nil {
		return PolicyConfig{}, fmt.Errorf("invalid policy_config JSON")
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return PolicyConfig{}, fmt.Errorf("invalid policy_config JSON")
	}
	return cfg, nil
}

func validatePolicyConfig(cfg PolicyConfig) error {
	if cfg.PolicyVersion != "noema_policy_v1" {
		return fmt.Errorf("policy_version must be noema_policy_v1")
	}
	if len(cfg.Constraints) == 0 {
		return fmt.Errorf("constraints must be non-empty")
	}
	seen := make(map[string]struct{}, len(cfg.Constraints))
	for _, c := range cfg.Constraints {
		id := strings.TrimSpace(c.ID)
		if id == "" {
			return fmt.Errorf("constraint id must be non-empty")
		}
		if id != c.ID {
			return fmt.Errorf("constraint id must not include leading/trailing whitespace")
		}
		if _, exists := seen[id]; exists {
			return fmt.Errorf("duplicate constraint id: %s", id)
		}
		seen[id] = struct{}{}
		if !ValidateAllowedMaxSeverity(c.MaxAllowed) {
			return fmt.Errorf("constraint max_allowed must be 0, 1, or 2")
		}
	}
	return nil
}

func policyConfigFromSpec(spec Spec) PolicyConfig {
	cfg := PolicyConfig{
		PolicyVersion: "noema_policy_v1",
		Constraints:   make([]PolicyConstraint, 0, len(spec.Constraints)+len(spec.CustomConstraints)),
	}
	for _, c := range spec.Constraints {
		cfg.Constraints = append(cfg.Constraints, PolicyConstraint{
			ID:         c.ID,
			Enabled:    c.Enabled,
			MaxAllowed: c.AllowedMaxSeverity,
		})
	}
	for _, c := range spec.CustomConstraints {
		cfg.Constraints = append(cfg.Constraints, PolicyConstraint{
			ID:         c.ID,
			Enabled:    c.Enabled,
			MaxAllowed: c.AllowedMaxSeverity,
		})
	}
	return cfg
}
