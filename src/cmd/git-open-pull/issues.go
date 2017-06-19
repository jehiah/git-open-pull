package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
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
	for _, l := range labels {
		fmt.Fprintf(tempFile, "# Label: %s\n", l)
	}

	io.WriteString(tempFile, `
# Please enter a title and description for your new issue. The first
# line will be used as the issue title, and any subsequent lines will
# be used as the issue description.
#
# Lines starting with '#' will be ignored.`)

	tempFile.Sync()
	tempFile.Close()
	cmd := exec.CommandContext(ctx, settings.Editor, tempFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		tempFile.Close()
		// os.Remove(tempFile.Name())
		return 0, err
	}
	tempFile, err = os.Open(tempFile.Name())
	if err != nil {
		return 0, err
	}

	var title string
	var descriptions, selectedLabels []string
	scanner := bufio.NewScanner(tempFile)
	for scanner.Scan() {
		log.Printf("line %#v", scanner.Text())
		line := strings.TrimSpace(scanner.Text())
		switch {
		case strings.HasPrefix(line, "Label:"):
			label := strings.TrimSpace(line[len("Label:"):])
			if label != "" {
				selectedLabels = append(selectedLabels, label)
			}
		case strings.HasPrefix(line, "#"):
		case title == "" && line != "":
			title = line
		default:
			descriptions = append(descriptions, strings.TrimRight(scanner.Text(), " \t\r\n"))
		}

		if err := scanner.Err(); err != nil {
			return 0, err
		}
	}
	description := strings.TrimSpace(strings.Join(descriptions, "\n"))

	log.Printf("title:%#v", title)
	log.Printf("description:%#v", description)
	log.Printf("labels:%#v", selectedLabels)

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
