package nodrama

import (
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func countCommits(r *git.Repository) (int, error) {
	cIter, err := r.Log(&git.LogOptions{})
	if err != nil {
		return 0, fmt.Errorf("error getting commit iterator: %w", err)
	}

	count := 0
	err = cIter.ForEach(func(c *object.Commit) error {
		count++
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("error counting commits: %w", err)
	}

	return count, nil
}

func isRepoClean(repo *git.Repository) (bool, error) {
	wt, err := repo.Worktree()
	if err != nil {
		return false, fmt.Errorf("error getting worktree: %w", err)
	}

	status, err := wt.Status()
	if err != nil {
		return false, fmt.Errorf("error getting worktree status: %w", err)
	}

	return status.IsClean(), nil
}

func isDirectoryReadable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir() && (info.Mode()&os.ModePerm)&0o400 != 0
}
