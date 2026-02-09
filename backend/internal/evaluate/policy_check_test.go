package evaluate

import "testing"

func TestComputePolicyResult_AllPass(t *testing.T) {
	cfg := PolicyConfig{
		PolicyVersion: "noema_policy_v1",
		Constraints: []PolicyConstraint{
			{ID: "a", Enabled: true, MaxAllowed: 2},
			{ID: "b", Enabled: true, MaxAllowed: 1},
		},
	}
	out := EvaluationResult{
		EvalVersion: "noema_eval_v1",
		Results: []EvalResultItem{
			{ID: "a", Severity: 2},
			{ID: "b", Severity: 1},
		},
	}
	overall, maxSeverity, threshold := computePolicyResult(out, cfg)
	if !overall {
		t.Fatalf("expected overall pass true")
	}
	if maxSeverity != 2 {
		t.Fatalf("expected max severity 2, got %d", maxSeverity)
	}
	if threshold != 1 {
		t.Fatalf("expected policy threshold 1, got %d", threshold)
	}
}

func TestComputePolicyResult_FailsWhenExceeded(t *testing.T) {
	cfg := PolicyConfig{
		PolicyVersion: "noema_policy_v1",
		Constraints: []PolicyConstraint{
			{ID: "a", Enabled: true, MaxAllowed: 0},
			{ID: "b", Enabled: true, MaxAllowed: 2},
		},
	}
	out := EvaluationResult{
		EvalVersion: "noema_eval_v1",
		Results: []EvalResultItem{
			{ID: "a", Severity: 1},
			{ID: "b", Severity: 2},
		},
	}
	overall, maxSeverity, _ := computePolicyResult(out, cfg)
	if overall {
		t.Fatalf("expected overall pass false")
	}
	if maxSeverity != 2 {
		t.Fatalf("expected max severity 2, got %d", maxSeverity)
	}
}

func TestComputePolicyResult_IgnoresDisabledConstraints(t *testing.T) {
	cfg := PolicyConfig{
		PolicyVersion: "noema_policy_v1",
		Constraints: []PolicyConstraint{
			{ID: "a", Enabled: true, MaxAllowed: 2},
			{ID: "b", Enabled: false, MaxAllowed: 0},
		},
	}
	out := EvaluationResult{
		EvalVersion: "noema_eval_v1",
		Results: []EvalResultItem{
			{ID: "a", Severity: 1},
			{ID: "b", Severity: 2},
		},
	}
	overall, maxSeverity, _ := computePolicyResult(out, cfg)
	if !overall {
		t.Fatalf("expected overall pass true")
	}
	if maxSeverity != 1 {
		t.Fatalf("expected max severity 1 from enabled constraints, got %d", maxSeverity)
	}
}

func TestComputePolicyResult_NoEnabledConstraints(t *testing.T) {
	cfg := PolicyConfig{
		PolicyVersion: "noema_policy_v1",
		Constraints: []PolicyConstraint{
			{ID: "a", Enabled: false, MaxAllowed: 2},
		},
	}
	out := EvaluationResult{
		EvalVersion: "noema_eval_v1",
		Results: []EvalResultItem{
			{ID: "a", Severity: 2},
		},
	}
	overall, maxSeverity, _ := computePolicyResult(out, cfg)
	if !overall {
		t.Fatalf("expected overall pass true when no enabled constraints")
	}
	if maxSeverity != 0 {
		t.Fatalf("expected max severity 0 when no enabled constraints, got %d", maxSeverity)
	}
}

func TestComputePolicyResult_PolicyThresholdMin(t *testing.T) {
	cfg := PolicyConfig{
		PolicyVersion: "noema_policy_v1",
		Constraints: []PolicyConstraint{
			{ID: "a", Enabled: true, MaxAllowed: 2},
			{ID: "b", Enabled: true, MaxAllowed: 0},
			{ID: "c", Enabled: false, MaxAllowed: 1},
		},
	}
	out := EvaluationResult{
		EvalVersion: "noema_eval_v1",
		Results: []EvalResultItem{
			{ID: "a", Severity: 0},
			{ID: "b", Severity: 0},
			{ID: "c", Severity: 2},
		},
	}
	_, _, threshold := computePolicyResult(out, cfg)
	if threshold != 0 {
		t.Fatalf("expected policy threshold 0, got %d", threshold)
	}
}
