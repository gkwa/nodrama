package nodrama

import (
	"flag"
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

func run(options Options) {
	srcPath := options.SourcePath

	if srcPath == "" {
		slog.Error("source path flag -src is empty")
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
