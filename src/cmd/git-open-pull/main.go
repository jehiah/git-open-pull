package main

import (
	"context"
	"fmt"
	"log"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func main() {
	ctx := context.Background()
	settings, err := LoadSettings(ctx)
	if err != nil {
		log.Fatalf("%s", err)
	}
	fmt.Printf("%#v\n", settings)

	branch, err := GitFeatureBranch(ctx)
	if err != nil {
		log.Fatalf("%s", err)
	}
	fmt.Printf("current branch %s", branch)
	issueNumber := DetectIssueNumber(branch)
	if issueNumber != 0 {
		fmt.Printf("issue number %d\n", issueNumber)
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: settings.Token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	client.UserAgent = "git-open-pull/1.0 (+http://github.com/jehiah/git-open-pull)"

	// get labels
	labels, _, err := client.Issues.ListLabels(ctx, settings.BaseAccount, settings.BaseRepo, &github.ListOptions{PerPage: 100})
	if err != nil {
		log.Fatalf("%s", err)
	}
	for _, l := range labels {
		if l.Name != nil {
			fmt.Printf("label:%s\n", *l.Name)
		}
	}

}
