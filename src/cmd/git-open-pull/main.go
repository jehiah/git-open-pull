package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"input"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
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
	fmt.Printf("current branch %s", branch)
	issueNumber := DetectIssueNumber(branch)
	if issueNumber != 0 {
		fmt.Printf("issue number %d\n", issueNumber)
	}

	client := SetupClient(ctx, settings)

	// prompt for new issue
	// read -p "enter issue number (or 'c' to create): " ISSUE_NUMBER
	// or, confirm auto-detected issue number
	// read -p "issue number [$ISSUE_NUMBER]: " temp

	if issueNumber != 0 {
		n, err := input.Ask(fmt.Sprintf("issue number [%d]", issueNumber), "")
		if err != nil {
			log.Fatal(err)
		}
		log.Fatalf("got %#v", n)
	} else {
		n, err := input.Ask("enter issue number (or 'c' to create)", "")
		if err != nil {
			log.Fatal(err)
		}
		if strings.ToLower(n) == "c" {
			_, err = NewIssue(ctx, client, settings)
		}
		if err != nil {
			log.Fatal(err)
		}
	}

}
