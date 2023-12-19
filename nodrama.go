package nodrama

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type Options struct {
	LogFormat  string
	LogLevel   string
	SourcePath string
}

func Execute() int {
	options := parseArgs()

	logger, err := getLogger(options.LogLevel, options.LogFormat)
	if err != nil {
		slog.Error("getLogger", "error", err)
		return 1
	}

	slog.SetDefault(logger)

	run(options)

	return 0
}

func parseArgs() Options {
	options := Options{}

	flag.StringVar(&options.LogLevel, "log-level", "info", "Log level (debug, info, warn, error), defult: info")
	flag.StringVar(&options.LogFormat, "log-format", "", "Log format (text or json)")
	flag.StringVar(&options.SourcePath, "src", "", "Local path to git repo")

	flag.Parse()

	return options
}

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

func run(options Options) {
	srcPath := options.SourcePath

	if srcPath == "" {
		slog.Error("source path flag -src is empty")
		return
	}

	srcFlag := flag.Lookup("src")
	if srcFlag == nil {
		panic("src flag has been removed")
	}

	if srcFlag == nil {
		slog.Error("source Path Flag empty", "value", srcFlag.Value.String())
		return
	}

	if !isDirectoryReadable(srcPath) {
		slog.Error("cannot read or not a directory", "path", srcPath)
		return
	}

	repo, err := git.PlainOpen(srcPath)
	if err != nil {
		slog.Error("error opening git repository", "path", srcPath, "error", err)
		return
	}

	isClean, err := isRepoClean(repo)
	if err != nil {
		slog.Error("error checking repository cleanliness", "repo", repo, "error", err)
		return
	}

	if !isClean {
		slog.Error("repository is dirty. Please commit or discard changes")
		return
	}

	commitCount, err := countCommits(repo)
	if err != nil {
		slog.Error("repo stats failed", "error", err)
		return
	}

	slog.Debug("repository stats", "clean", isClean, "repo", srcPath, "commits", commitCount)
}
