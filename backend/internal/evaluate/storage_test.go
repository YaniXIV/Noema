package evaluate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateRunDir(t *testing.T) {
	runsDir := t.TempDir()
	runID, runPath, err := createRunDir(runsDir)
	if err != nil {
		t.Fatalf("createRunDir error: %v", err)
	}
	if !strings.HasPrefix(runID, "run_") {
		t.Fatalf("expected runID to start with run_, got %q", runID)
	}
	if runPath != filepath.Join(runsDir, runID) {
		t.Fatalf("unexpected runPath: %q", runPath)
	}
	info, err := os.Stat(runPath)
	if err != nil {
		t.Fatalf("stat runPath error: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("expected runPath to be a directory")
	}
}

func TestCreateRunDirCreatesBase(t *testing.T) {
	baseDir := t.TempDir()
	runsDir := filepath.Join(baseDir, "nested", "runs")
	if _, err := os.Stat(runsDir); !os.IsNotExist(err) {
		t.Fatalf("expected runsDir to not exist yet")
	}
	runID, runPath, err := createRunDir(runsDir)
	if err != nil {
		t.Fatalf("createRunDir error: %v", err)
	}
	if !strings.HasPrefix(runID, "run_") {
		t.Fatalf("expected runID to start with run_, got %q", runID)
	}
	info, err := os.Stat(runPath)
	if err != nil {
		t.Fatalf("stat runPath error: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("expected runPath to be a directory")
	}
}
