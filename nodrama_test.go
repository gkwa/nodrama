package nodrama

import (
	"os"
	"strings"
	"testing"

	"github.com/go-git/go-git/v5"
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

func TestRunWithNonExistantDirectory(t *testing.T) {
	nonExistentDir := "/path/to/nonexistent/directory"

	options := Options{
		SourcePath: nonExistentDir,
	}

	err := run(options)

	if err == nil {
		t.Error("expected an error, got nil")
		return
	}

	expectedErrorMsg := "cannot read or not a directory"
	if !strings.Contains(err.Error(), expectedErrorMsg) {
		t.Errorf("expected error message to contain '%s', got: %s", expectedErrorMsg, err.Error())
		return
	}
}

func TestGitRepoForDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "nodrama-test")
	if err != nil {
		t.Fatalf("failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize a git repository in the temporary directory
	_, err = git.PlainInit(tmpDir, false)
	if err != nil {
		t.Fatalf("failed to initialize git repository: %v", err)
	}

	repo, err := GitRepoForDir(tmpDir)

	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	if repo == nil {
		t.Error("expected a non-nil repository")
	}
}

func TestCountCommits(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "nodrama-test")
	if err != nil {
		t.Fatalf("failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize a git repository in the temporary directory
	repo, err := git.PlainInit(tmpDir, false)
	if err != nil {
		t.Fatalf("failed to initialize git repository: %v", err)
	}

	// Make two test commits
	w, err := repo.Worktree()
	if err != nil {
		t.Fatalf("failed to get worktree: %v", err)
	}

	// Create and stage test.txt
	testFile, err := os.Create(tmpDir + "/test.txt")
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	testFile.Close() // Close the file before adding to the index
	_, err = w.Add("test.txt")
	if err != nil {
		t.Fatalf("failed to stage test file: %v", err)
	}

	// Commit the changes
	_, err = w.Commit("Initial commit", &git.CommitOptions{})
	if err != nil {
		t.Fatalf("failed to commit initial commit: %v", err)
	}

	// Create and stage test2.txt
	test2File, err := os.Create(tmpDir + "/test2.txt")
	if err != nil {
		t.Fatalf("failed to create test2 file: %v", err)
	}
	test2File.Close() // Close the file before adding to the index
	_, err = w.Add("test2.txt")
	if err != nil {
		t.Fatalf("failed to stage test2 file: %v", err)
	}

	// Commit the changes
	_, err = w.Commit("Second commit", &git.CommitOptions{})
	if err != nil {
		t.Fatalf("failed to commit second commit: %v", err)
	}

	// Get the count of commits
	count, err := CountCommits(repo)

	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	expectedCount := 2
	if count != expectedCount {
		t.Errorf("expected %d commits, got: %d", expectedCount, count)
	}
}
