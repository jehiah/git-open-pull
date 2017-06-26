package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func RunGit(ctx context.Context, arg ...string) ([]byte, error) {
	// log.Printf("run git %s", strings.Join(arg, " "))
	cmd := exec.CommandContext(ctx, "git", arg...)
	return cmd.Output()
}

func GitFeatureBranch(ctx context.Context) (string, error) {
	body, err := RunGit(ctx, "rev-parse", "--abbrev-ref", "HEAD")
	return strings.TrimSpace(string(body)), err
}

func MergeBase(ctx context.Context, settings *Settings) (string, error) {
	_, err := RunGit(ctx, "fetch", settings.BaseAccount, fmt.Sprintf("+refs/heads/%s", settings.BaseBranch))
	if err != nil {
		return "", err
	}
	base, err := RunGit(ctx, "merge-base", "FETCH_HEAD", "HEAD")
	return strings.TrimSpace(string(base)), err
}

// reverse an array of strings
func reverse(ss []string) {
	last := len(ss) - 1
	for i := 0; i < len(ss)/2; i++ {
		ss[i], ss[last-i] = ss[last-i], ss[i]
	}
}

// Commits returns a list of all commit sha's since merge base
// the oldest commit (first) is returned first
func Commits(ctx context.Context, base string) ([]string, error) {
	output, err := RunGit(ctx, "log", "--format=%H", fmt.Sprintf("%s...HEAD", base))
	commits := strings.Split(strings.TrimSpace(string(output)), "\n")
	reverse(commits)
	return commits, err
}

// CommitDetails gets the subject and body for a commit
func CommitDetails(ctx context.Context, hash string) (string, string, error) {
	title, err := RunGit(ctx, "show", "-s", "--format=%s", hash)
	if err != nil {
		return "", "", err
	}
	body, err := RunGit(ctx, "show", "-s", "--format=%b", hash)
	return strings.TrimSpace(string(title)), strings.TrimSpace(string(body)), err
}
