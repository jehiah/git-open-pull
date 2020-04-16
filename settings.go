package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/jehiah/git-open-pull/internal/input"
)

type Settings struct {
	// config: github.user
	User string
	// config: gitOpenPull.token
	Token string
	// config: gitOpenPull.baseAccount
	BaseAccount string
	// config: gitOpenPull.baseRepo
	BaseRepo string
	// config: gitOpenPull.base
	BaseBranch string
	// Editor to use for draft PR description (default: vi)
	// config: core.editor
	Editor string
	// Allow maintainers of the upstream repo to modify this branch
	// https://help.github.com/articles/allowing-changes-to-a-pull-request-branch-created-from-a-fork/
	// config: gitOpenPull.maintainersCanModify
	MaintainersCanModify bool

	// command to pre or post process the commit template
	// It is run with the first argument as the template name
	PreProcess  string
	PostProcess string
	// callback is called after a PR is created. It's first argument is a filename that contains the PR json
	Callback string
}

// this function tries to get settings from the environment variables
func GetEnvSettings(s *Settings) error {
	token := os.Getenv("GITOPENPULL_TOKEN")
	if token != "" {
		s.Token = token
	}
	user := os.Getenv("GITOPENPULL_USER")
	if user != "" {
		s.User = user
	}
	baseAccount := os.Getenv("GITOPENPULL_BASE_ACCOUNT")
	if baseAccount != "" {
		s.BaseAccount = baseAccount
	}
	baseRepo := os.Getenv("GITOPENPULL_BASE_REPO")
	if baseRepo != "" {
		s.BaseRepo = baseRepo
	}
	preProcess := os.Getenv("GITOPENPULL_PRE_PROCESS")
	if preProcess != "" {
		s.PreProcess = preProcess
	}
	postProcess := os.Getenv("GITOPENPULL_POST_PROCESS")
	if postProcess != "" {
		s.PostProcess = postProcess
	}
	callback := os.Getenv("GITOPENPULL_CALLBACK")
	if callback != "" {
		s.Callback = callback
	}

	baseBranch := os.Getenv("GITOPENPULL_BASE_BRANCH")
	if baseBranch != "" {
		s.BaseBranch = baseBranch
	}

	mcmStr := os.Getenv("GITOPENPULL_MAINTAINERS_CAN_MODIFY")
	if mcmStr != "" {
		mcm, err := strconv.ParseBool(mcmStr)
		if err != nil {
			return err
		}
		s.MaintainersCanModify = mcm
	}

	editor := os.Getenv("GITOPENPULL_EDITOR")
	if editor != "" {
		s.Editor = editor
	}

	return nil
}

// LoadSettings extracts the git and gitOpenPull sections from $HOME/.gitconfig and .git/config
func LoadSettings(ctx context.Context) (*Settings, error) {

	body, err := RunGit(ctx, "config", "--list")
	if err != nil {
		return nil, err
	}
	s := Settings{
		BaseBranch: "master",
		Editor:     "/usr/bin/vi",
	}
	var defaultBaseRepo, maintainersCanModify string
	scanner := bufio.NewScanner(bytes.NewBuffer(body))
	for scanner.Scan() {
		line := strings.SplitN(strings.TrimSpace(scanner.Text()), "=", 2)
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
		case "gitopenpull.maintainerscanmodify":
			maintainersCanModify = line[1]
			switch strings.ToLower(line[1]) {
			case "true":
				s.MaintainersCanModify = true
			}
		case "gitopenpull.preprocess":
			s.PreProcess = line[1]
		case "gitopenpull.postprocess":
			s.PostProcess = line[1]
		case "gitopenpull.callback":
			s.Callback = line[1]
		case "core.editor":
			s.Editor = line[1]
		default:
			if strings.HasSuffix(line[0], ".url") && strings.HasSuffix(line[1], ".git") && defaultBaseRepo == "" {
				chunks := strings.Split(line[1], "/")
				defaultBaseRepo = chunks[len(chunks)-1]
				defaultBaseRepo = defaultBaseRepo[:len(defaultBaseRepo)-4]
			}
		}
	}
	if maintainersCanModify == "" {
		s.MaintainersCanModify = true
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	err = GetEnvSettings(&s)
	if err != nil {
		return nil, err
	}

	// https://github.com/settings/tokens
	if s.User == "" {
		s.User, err = input.Ask("GitHub username", "")
		if err != nil {
			return nil, err
		}
		if s.User == "" {
			return nil, errors.New("GitHub username required. Set `git config --global github.user $USER`")
		}
		_, err = RunGit(ctx, "config", "--global", "github.user", s.User)
		if err != nil {
			return nil, err
		}
	}

	if s.BaseAccount == "" {
		s.BaseAccount, err = input.Ask("destination GitHub username (account to pull code into)", "")
		if err != nil {
			return nil, err
		}
		if s.BaseAccount == "" {
			return nil, fmt.Errorf("Destination GitHub username required. Set `git config gitOpenPull.baseAccount $USER`")
		}
		_, err = RunGit(ctx, "config", "gitOpenPull.baseAccount", s.BaseAccount)
		if err != nil {
			return nil, err
		}
	}

	if s.BaseRepo == "" {
		s.BaseRepo, err = input.Ask(fmt.Sprintf("GitHub repository name (ie: github.com/%s/___)", s.BaseAccount), defaultBaseRepo)
		if err != nil {
			return nil, err
		}
		if s.BaseRepo == "" {
			return nil, fmt.Errorf("GitHub repository name required. Set `git config gitOpenPull.baseAccount $PROJECT`")
		}
		_, err = RunGit(ctx, "config", "gitOpenPull.baseRepo", s.BaseRepo)
		if err != nil {
			return nil, err
		}
	}

	if s.Token == "" {
		s.Token, err = input.Ask("Github access token (You can genrate a token from https://github.com/settings/tokens)", "")
		if err != nil {
			return nil, err
		}
		if s.Token == "" {
			return nil, fmt.Errorf("Github token required. Set `git config --global gitOpenPull.token $TOKEN`")
		}
		_, err = RunGit(ctx, "config", "--global", "gitOpenPull.token", s.Token)
		if err != nil {
			return nil, err
		}
	}

	return &s, nil

}
