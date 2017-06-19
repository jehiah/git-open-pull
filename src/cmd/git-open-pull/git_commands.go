package main

import (
	"context"
	"os/exec"
	"strings"
)

func RunGit(ctx context.Context, arg ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "git", arg...)
	return cmd.Output()
}

func GitFeatureBranch(ctx context.Context) (string, error) {
	body, err := RunGit(ctx, "rev-parse", "--abbrev-ref", "HEAD")
	return strings.TrimSpace(string(body)), err
}

func MergeBase(ctx context.Context, settings *Settings) (string, error) {
	base, err := RunGit(ctx, "merge-base", settings.BaseBranch, "HEAD")
	if err != nil {
		// fetch remote and re-try
		_, err = RunGit(ctx, "fetch", settings.BaseAccount, "+refs/heads/master:master")
		if err != nil {
			return "" ,err
		}
		base, err := RunGit(ctx, "merge-base", settings.BaseBranch, "HEAD")
	}
	return strings.TrimSpace(base), err
}

func Commits(ctx context.Context, base string) ([]string, error) {
	output, err := RunGit(ctx, "log", "--format=\"%h\"", fmt.Sprintf("%s...HEAD", base))
	return strings.Split(output, "\n"), err
}