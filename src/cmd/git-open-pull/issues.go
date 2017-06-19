package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/google/go-github/github"
)

func DetectIssueNumber(branch string) int {
	if branch == "" {
		return 0
	}
	b := strings.Replace(branch, "-", "_", -1)
	chunks := strings.Split(b, "_")

	for _, index := range []int{len(chunks) - 1, 0} {
		if n, err := strconv.Atoi(chunks[index]); err == nil && n > 0 {
			return n
		}
	}
	return 0
}

// NewIssue creates a template, parses the template and returns the Issue number
func NewIssue(ctx context.Context, client *github.Client, settings *Settings) (issueNumber int, err error) {
	labels, err := Labels(ctx, client, settings)
	if err != nil {
		return 0, err
	}

	tempFile, err := ioutil.TempFile("", "git-open-pull")
	if err != nil {
		return 0, err
	}
	fmt.Printf("drafting %s\n", tempFile.Name())
	defer os.Remove(tempFile.Name())

	// write commit history
	io.WriteString(tempFile, "\n# Uncomment to assign labels\n")
	for l, _ := range labels {
		fmt.Fprintf(tempFile, "# Label: %s\n", l)
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
		return 0, err
	}
	tempFile.Seek(0, 0)

	// TODO: read the template
	// parse out the comments
	// parse out the labels
	// parse the title and description
	// create issue

	return 0, fmt.Errorf("not implemented")

}

// Labels returns all of the labels for a given repo
func Labels(ctx context.Context, client *github.Client, settings *Settings) ([]string, error) {
	labels, _, err := client.Issues.ListLabels(ctx, settings.BaseAccount, settings.BaseRepo, &github.ListOptions{PerPage: 100})
	if err != nil {
		return nil, err
	}
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
	var o []string
	for _, l := range labels {
		if l.Name != nil {
			o = append(o, *l.Name)
		}
	}
	return o, nil
}
