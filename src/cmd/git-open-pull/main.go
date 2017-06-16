package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sort"
	"time"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func main() {
	ctx := context.Background()
	settings, err := LoadSettings(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%#v\n", settings)

	branch, err := GitFeatureBranch(ctx)
	if err != nil {
		log.Fatal(err)
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
		log.Fatal(err)
	}
	fmt.Printf("got %d labels\n", len(labels))
	sort.Slice(labels, func(i, j int) bool {
		switch {
		case labels[i] == nil:
			return true
		case labels[j] == nil:
			return false
		default:
			return *labels[i].Name < *labels[j].Name
		}
	})

	tempFile, err := ioutil.TempFile("", "git-open-pull")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("drafting %s\n", tempFile.Name())
	defer os.Remove(tempFile.Name())

	// write commit history
	io.WriteString(tempFile, "\n# Uncomment to assign labels\n")
	for _, l := range labels {
		if l.Name != nil {
			fmt.Fprintf(tempFile, "# Label: %s\n", *l.Name)
		}
	}

	io.WriteString(tempFile, `
# Please enter a title and description for your new issue. The first
# line will be used as the issue title, and any subsequent lines will
# be used as the issue description.
#
# Lines starting with '#' will be ignored.`)

	cmd := exec.CommandContext(ctx, settings.Editor, tempFile.Name())
	err = cmd.Run()
	if err != nil {
		tempFile.Close()
		os.Remove(tempFile.Name())
		log.Fatal(err)
	}
	tempFile.Seek(0, 0)

}
