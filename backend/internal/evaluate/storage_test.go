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
