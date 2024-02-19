package nodrama

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
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

	w, err := repo.Worktree()
	if err != nil {
		t.Fatalf("failed to get worktree: %v", err)
	}

	testFile, err := os.Create(tmpDir + "/test.txt")
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	testFile.Close() // Close the file before adding to the index
	_, err = w.Add("test.txt")
	if err != nil {
		t.Fatalf("failed to stage test file: %v", err)
	}

	_, err = w.Commit("Initial commit", &git.CommitOptions{})
	if err != nil {
		t.Fatalf("failed to commit initial commit: %v", err)
	}

	test2File, err := os.Create(tmpDir + "/test2.txt")
	if err != nil {
		t.Fatalf("failed to create test2 file: %v", err)
	}
	test2File.Close() // Close the file before adding to the index
	_, err = w.Add("test2.txt")
	if err != nil {
		t.Fatalf("failed to stage test2 file: %v", err)
	}

	_, err = w.Commit("Second commit", &git.CommitOptions{})
	if err != nil {
		t.Fatalf("failed to commit second commit: %v", err)
	}

	count, err := CountCommits(repo)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	expectedCount := 2
	if count != expectedCount {
		t.Errorf("expected %d commits, got: %d", expectedCount, count)
	}
}

func TestGitCommitAll(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "nodrama-test")
	if err != nil {
		t.Fatalf("failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	_, err = git.PlainInit(tmpDir, false)
	if err != nil {
		t.Fatalf("failed to initialize git repository: %v", err)
	}

	for i := 0; i < 10; i++ {
		fileName := fmt.Sprintf("test%d.txt", i)
		filePath := filepath.Join(tmpDir, fileName)
		file, err := os.Create(filePath)
		if err != nil {
			t.Fatalf("failed to create test file %s: %v", fileName, err)
		}
		file.Close()
	}

	err = GitCommitAll(tmpDir, "Add all files")
	if err != nil {
		t.Fatalf("failed to commit all files: %v", err)
	}

	repo, err := git.PlainOpen(tmpDir)
	if err != nil {
		t.Fatalf("failed to open git repository: %v", err)
	}

	count, err := CountCommits(repo)
	if err != nil {
		t.Errorf("error counting commits: %v", err)
	}

	expectedCount := 1 // We've only made one commit
	if count != expectedCount {
		t.Errorf("expected %d commits, got: %d", expectedCount, count)
	}

	w, err := repo.Worktree()
	if err != nil {
		t.Fatalf("failed to get worktree: %v", err)
	}

	files, err := w.Filesystem.ReadDir(".")
	if err != nil {
		t.Fatalf("failed to read directory %s: %v", tmpDir, err)
	}

	fileCount := 0
	for _, file := range files {
		if file.Name() != ".git" { // Exclude .git directory
			fileCount++
		}
	}

	expectedFileCount := 10 // We created 10 test files
	if fileCount != expectedFileCount {
		t.Errorf("expected %d files, got: %d", expectedFileCount, fileCount)
	}
}

func createTestFiles(dir string, count int) error {
	for i := 0; i < count; i++ {
		filePath := filepath.Join(dir, "test"+strconv.Itoa(i)+".txt")
		file, err := os.Create(filePath)
		if err != nil {
			return err
		}
		file.Close()
	}
	return nil
}

func TestGetRecentCommitFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "nodrama-test")
	if err != nil {
		t.Fatalf("failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	repo, err := git.PlainInit(tmpDir, false)
	if err != nil {
		t.Fatalf("failed to initialize git repository: %v", err)
	}

	numFiles := 3
	err = createTestFiles(tmpDir, numFiles)
	if err != nil {
		t.Fatalf("failed to create test files: %v", err)
	}

	w, err := repo.Worktree()
	if err != nil {
		t.Fatalf("failed to get worktree: %v", err)
	}
	_, err = w.Add(".")
	if err != nil {
		t.Fatalf("failed to add files: %v", err)
	}

	_, err = w.Commit("Add test files", &git.CommitOptions{})
	if err != nil {
		t.Fatalf("failed to commit changes: %v", err)
	}

	files, err := GetRecentCommitFiles(repo)
	if err != nil {
		t.Fatalf("failed to get recent commit files: %v", err)
	}

	if len(files) != numFiles {
		t.Errorf("expected %d files in the most recent commit, got %d", numFiles, len(files))
	}

	expectedPaths := []string{
		filepath.Join(tmpDir, "test0.txt"),
		filepath.Join(tmpDir, "test1.txt"),
		filepath.Join(tmpDir, "test2.txt"),
	}
	if !reflect.DeepEqual(files, expectedPaths) {
		t.Errorf("expected file paths %v, got %v", expectedPaths, files)
	}
}

func TestGetRecentCommitFilesWhenNoCommitsExistYet(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "nodrama-test")
	if err != nil {
		t.Fatalf("failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	repo, err := git.PlainInit(tmpDir, false)
	if err != nil {
		t.Fatalf("failed to initialize git repository: %v", err)
	}

	files, err := GetRecentCommitFiles(repo)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if len(files) != 0 {
		t.Errorf("expected empty files list, got: %v", files)
	}
}

func TestSumBytesOfFiles(t *testing.T) {
	repoDir, err := os.MkdirTemp("", "test-repo")
	if err != nil {
		t.Fatalf("failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(repoDir)

	repo, err := git.PlainInit(repoDir, false)
	if err != nil {
		t.Fatalf("failed to initialize git repository: %v", err)
	}

	filePath := filepath.Join(repoDir, "test.txt")
	file, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	defer file.Close()
	file.WriteString("aaa") // Write 3 bytes to the file

	w, err := repo.Worktree()
	if err != nil {
		t.Fatalf("failed to get worktree: %v", err)
	}
	_, err = w.Add("test.txt")
	if err != nil {
		t.Fatalf("failed to add file to git: %v", err)
	}

	_, err = w.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
		},
	})
	if err != nil {
		t.Fatalf("failed to commit changes: %v", err)
	}

	fileSize, err := SumBytesOfFiles([]string{filePath})
	if err != nil {
		t.Fatalf("SumBytesOfFiles failed: %v", err)
	}

	expectedSize := int64(3)
	if fileSize != expectedSize {
		t.Errorf("expected file size to be %d byte(s), got %d", expectedSize, fileSize)
	}
}
