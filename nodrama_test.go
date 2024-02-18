package nodrama

import (
	"os"
	"strings"
	"testing"
)

func TestRunWithNonGitDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "nodrama-test")
	if err != nil {
		t.Fatalf("failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	options := Options{
		SourcePath: tmpDir,
	}

	err = run(options)

	if err == nil {
		t.Error("expected an error, got nil")
		return
	}

	expectedErrorMsg := "repository does not exist"
	if !strings.Contains(err.Error(), expectedErrorMsg) {
		t.Errorf("expected error message to contain '%s', got: %s", expectedErrorMsg, err.Error())
		return
	}
}
