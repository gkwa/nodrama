package nodrama

import (
	"flag"
	"fmt"
	"log/slog"
	"path/filepath"

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

func GitRepoForDir(dir string) (*git.Repository, error) {
	if !isDirectoryReadable(dir) {
		return nil, fmt.Errorf("cannot read or not a directory: %s", dir)
	}

	repo, err := git.PlainOpen(dir)
	if err != nil {
		return nil, fmt.Errorf("error opening git repository: %w", err)
	}

	return repo, nil
}

func run(options Options) error {
	srcPath := options.SourcePath

	if srcPath == "" {
		return fmt.Errorf("source path flag -src is empty")
	}

	repo, err := GitRepoForDir(srcPath)
	if err != nil {
		return fmt.Errorf("error opening git repository: %w", err)
	}

	isClean, err := IsRepoClean(repo)
	if err != nil {
		return fmt.Errorf("error checking repository cleanliness: %w", err)
	}

	if !isClean {
		return fmt.Errorf("repository is dirty. Please commit or discard changes")
	}

	commitCount, err := CountCommits(repo)
	if err != nil {
		return fmt.Errorf("error counting commits: %w", err)
	}

	slog.Debug("repository stats", "clean", isClean, "repo", srcPath, "commits", commitCount)
	return nil
}

// GitCommitAll commits all files in the git repository located at srcPath with the given message.
func GitCommitAll(repoRoot, msg string) error {
	repo, err := GitRepoForDir(repoRoot)
	if err != nil {
		return fmt.Errorf("error opening git repository: %w", err)
	}

	w, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("error getting worktree: %w", err)
	}

	_, err = w.Add(".")
	if err != nil {
		return fmt.Errorf("error adding files: %w", err)
	}

	_, err = w.Commit(msg, &git.CommitOptions{})
	if err != nil {
		return fmt.Errorf("error committing: %w", err)
	}

	return nil
}

func GetRecentCommitFiles(repo *git.Repository) ([]string, error) {
	ref, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD reference: %w", err)
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get commit object: %w", err)
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to get tree: %w", err)
	}

	var filepaths []string
	tree.Files().ForEach(func(f *object.File) error {
		filepaths = append(filepaths, f.Name)
		return nil
	})

	var absolutePaths []string
	worktree, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}
	workdir := worktree.Filesystem.Root()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree root: %w", err)
	}
	for _, path := range filepaths {
		absolutePath := filepath.Join(workdir, path)
		absolutePaths = append(absolutePaths, absolutePath)
	}

	return absolutePaths, nil
}
