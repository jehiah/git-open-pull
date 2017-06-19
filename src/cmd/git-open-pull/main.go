package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"input"
)

func SetupClient(ctx context.Context, s *Settings) *github.Client {
	if s == nil {
		panic("missing settings")
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: s.Token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	client.UserAgent = "git-open-pull/1.0 (+http://github.com/jehiah/git-open-pull)"
	return client
}

// GetIssueNumber prompts to create a new issue, or confirmation of auto-detected issue number
func GetIssueNumber(ctx context.Context, client *github.Client, settings *Settings, detected int) (int, error) {
	var issue int
	if detected == 0 {
		n, err := input.Ask("enter issue number (or 'c' to create)", "")
		if err != nil {
			return issue, err
		}
		switch n {
		case "", "c", "C":
			return NewIssue(ctx, client, settings)
		default:
			return strconv.Atoi(n)
		}
	}
	n, err := input.Ask(fmt.Sprintf("issue number [%d]", detected), "")
	if err != nil {
		log.Fatal(err)
	}
	if n == "" {
		return detected, nil
	}
	return strconv.Atoi(n)
}

func main() {
	ctx := context.Background()
	settings, err := LoadSettings(ctx)
	if err != nil {
		log.Fatal(err)
	}

	branch, err := GitFeatureBranch(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("current branch %s\n", branch)
	detectedIssueNumber := DetectIssueNumber(branch)
	if detectedIssueNumber != 0 {
		fmt.Printf("issue number %d\n", detectedIssueNumber)
	}

	client := SetupClient(ctx, settings)

	issueNumber, err := GetIssueNumber(ctx, client, settings, detectedIssueNumber)
	if err != nil {
		log.Fatal(err)
	}
	if issueNumber == 0 {
		log.Fatal("expected issue number")
	}

	// # Do we need/want to rename the branch?
	if issueNumber != detectedIssueNumber {
		yn, err := input.Ask(fmt.Sprintf("rename branch to %s_%d [Y/n]", branch, issueNumber), "")
		if err != nil {
			log.Fatal(err)
		}
		switch yn {
		case "", "y", "Y":
			fmt.Printf("renaming local branch %s to %s_%d\n", branch, branch, issueNumber)
			branch = fmt.Sprintf("%s_%d", branch, issueNumber)
			_, err = RunGit(ctx, "branch", "-m", branch)
			if err != nil {
				log.Fatal(err)
			}
		case "n", "N":
		default:
			log.Fatalf("unknown response %q", yn)
		}
	}

	fmt.Printf("pushing branch %s to %s", branch, settings.User)
	_, err = RunGit(ctx, "push", "-u", settings.User, branch)
	if err != nil {
		log.Fatal(err)
	}

	log.Fatalf("not implemented")
}
