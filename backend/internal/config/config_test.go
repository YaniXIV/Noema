package config

import "testing"

func TestSampleItemsLimit(t *testing.T) {
	t.Setenv("NOEMA_SAMPLE_ITEMS", "")
	if got := SampleItemsLimit(); got != 100 {
		t.Fatalf("expected default 100, got %d", got)
	}

	t.Setenv("NOEMA_SAMPLE_ITEMS", "25")
	if got := SampleItemsLimit(); got != 25 {
		t.Fatalf("expected 25, got %d", got)
	}

	t.Setenv("NOEMA_SAMPLE_ITEMS", "0")
	if got := SampleItemsLimit(); got != 100 {
		t.Fatalf("expected default 100 for 0, got %d", got)
	}

	t.Setenv("NOEMA_SAMPLE_ITEMS", "-5")
	if got := SampleItemsLimit(); got != 100 {
		t.Fatalf("expected default 100 for negative, got %d", got)
	}

	t.Setenv("NOEMA_SAMPLE_ITEMS", "nope")
	if got := SampleItemsLimit(); got != 100 {
		t.Fatalf("expected default 100 for invalid, got %d", got)
	}
}

func TestRunsIndexLimit(t *testing.T) {
	t.Setenv("NOEMA_RUNS_INDEX_LIMIT", "")
	if got := RunsIndexLimit(); got != 50 {
		t.Fatalf("expected default 50, got %d", got)
	}

	t.Setenv("NOEMA_RUNS_INDEX_LIMIT", "12")
	if got := RunsIndexLimit(); got != 12 {
		t.Fatalf("expected 12, got %d", got)
	}

	t.Setenv("NOEMA_RUNS_INDEX_LIMIT", "0")
	if got := RunsIndexLimit(); got != 50 {
		t.Fatalf("expected default 50 for 0, got %d", got)
	}

	t.Setenv("NOEMA_RUNS_INDEX_LIMIT", "-1")
	if got := RunsIndexLimit(); got != 50 {
		t.Fatalf("expected default 50 for negative, got %d", got)
	}
}

func TestRunsMax(t *testing.T) {
	t.Setenv("NOEMA_RUNS_MAX", "")
	if got := RunsMax(); got != 50 {
		t.Fatalf("expected default 50, got %d", got)
	}

	t.Setenv("NOEMA_RUNS_MAX", "0")
	if got := RunsMax(); got != 0 {
		t.Fatalf("expected 0 to disable pruning, got %d", got)
	}

	t.Setenv("NOEMA_RUNS_MAX", "200")
	if got := RunsMax(); got != 200 {
		t.Fatalf("expected 200, got %d", got)
	}

	t.Setenv("NOEMA_RUNS_MAX", "-2")
	if got := RunsMax(); got != 50 {
		t.Fatalf("expected default 50 for negative, got %d", got)
	}

	t.Setenv("NOEMA_RUNS_MAX", "nope")
	if got := RunsMax(); got != 50 {
		t.Fatalf("expected default 50 for invalid, got %d", got)
	}
}
