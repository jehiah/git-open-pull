package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type Settings struct {
	User        string
	Token       string
	BaseAccount string
	BaseRepo    string
	BaseBranch  string
	Editor      string
}

func RunGit(ctx context.Context, arg ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "git", arg...)
	return cmd.Output()
}

// LoadSettings extracts the git and gitOpenPull sections from $HOME/.gitconfig and .git/config
func LoadSettings(ctx context.Context) (*Settings, error) {

	body, err := RunGit(ctx, "config", "-l")
	if err != nil {
		return nil, err
	}
	s := Settings{
		BaseBranch: "master",
		Editor:     "/usr/bin/vi",
	}
	scanner := bufio.NewScanner(bytes.NewBuffer(body))
	for scanner.Scan() {
		line := strings.SplitN(strings.TrimSpace(scanner.Text()), " ", 2)
		if len(line) != 2 {
			return nil, fmt.Errorf("Invalid line %#v", line)
		}
		switch line[0] {
		case "github.user":
			s.User = line[1]
		case "gitopenpull.token":
			s.Token = line[1]
		case "gitopenpull.baseaccount":
			s.BaseAccount = line[1]
		case "gitopenpull.baserepo":
			s.BaseRepo = line[1]
		case "gitopenpull.base":
			s.BaseBranch = line[1]
		case "core.editor":
			s.Editor = line[1]
		}
	}

	return &s, scanner.Err()

}

func GitFeatureBranch(ctx context.Context) (string, error) {
	body, err := RunGit(ctx, "rev-parse", "--abbrev-ref", "HEAD")
	return string(body), err
}
