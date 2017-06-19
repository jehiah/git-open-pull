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
