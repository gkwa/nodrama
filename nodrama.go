package nodrama

import (
	"flag"
	"fmt"
	"log/slog"

	"github.com/go-git/go-git/v5"
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
