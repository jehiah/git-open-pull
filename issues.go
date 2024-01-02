package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/google/go-github/v57/github"
)

// DetectIssueNumber parses out an existing issue from passed in branch name.
// Issue numbers appear at the end of branch names are are separated from the
// rest of the branch name by an underscore.
// i.e: somebranch_1234
func DetectIssueNumber(branch string) int {
	if branch == "" {
		return 0
	}

	sub := strings.Split(branch, "_")
	if len(sub) <= 1 {
		return 0
	}

	issueNumber, err := strconv.Atoi(sub[len(sub)-1])
	if err != nil {
		return 0
	}

	return issueNumber
}

func NewIssue(ctx context.Context, client *github.Client, settings *Settings, interactive bool, title, description string, labels []string) (issueNumber int, err error) {
	var gir *github.IssueRequest
	if interactive {
		gir, err = PopulateIssueInteractive(ctx, client, settings, title, description, labels)
		if err != nil {
			log.Fatalf("Interactive issue creation failed: %v", err)
		}

	} else {
		if title == "" {
			log.Fatal("title cannot be empty")
		}

		gir = &github.IssueRequest{
			Title:    &title,
			Body:     &description,
			Assignee: &settings.User,
		}

		if labels != nil {
			gir.Labels = &labels
		}
	}

	i, _, err := client.Issues.Create(ctx, settings.BaseAccount, settings.BaseRepo, gir)
	if err != nil {
		return 0, err
	}

	if interactive {
		fmt.Printf("Created issue %d (%s)\n", *i.Number, *i.Title)
	}

	return *i.Number, nil
}

// PopulateIssueInteractive creates a template, parses the template and returns the Issue number if the user is in interactive mode
func PopulateIssueInteractive(ctx context.Context, client *github.Client, settings *Settings, inputTitle, inputDescription string, labelSlice []string) (ir *github.IssueRequest, err error) {
	labels, err := Labels(ctx, client, settings)
	if err != nil {
		return nil, err
	}

	labelSet := make(map[string]bool)
	for _, l := range labelSlice {
		labelSet[l] = true
	}

	tempFile, err := os.CreateTemp("", "git-open-pull")
	if err != nil {
		return nil, err
	}
	// fmt.Printf("drafting %s\n", tempFile.Name())
	defer os.Remove(tempFile.Name())

	if inputTitle != "" {
		fmt.Fprintf(tempFile, "%s\n", inputTitle)
	}
	if inputDescription != "" {
		fmt.Fprintf(tempFile, "%s\n", inputDescription)
	}

	// seed template with commit history
	mergeBase, err := MergeBase(ctx, settings)
	if err != nil {
		log.Printf("error getting merge base %s", err)
	} else {
		// fmt.Printf("merge base is %s\n", mergeBase)
		if mergeBase != "" {
			commits, err := Commits(ctx, mergeBase)
			if err != nil {
				log.Printf("error getting commits %s", err)
			}
			for i, c := range commits {
				// log.Printf("[%d] commit %s", i, c)
				t, b, err := CommitDetails(ctx, c)
				if err != nil {
					return nil, err
				}
				if t == "" {
					continue
				}
				switch i {
				case 0:
					fmt.Fprintf(tempFile, "%s\n", t)
				case 1:
					fmt.Fprintf(tempFile, "\n * %s\n", t)
				default:
					fmt.Fprintf(tempFile, " * %s\n", t)
				}
				if b != "" {
					fmt.Fprintf(tempFile, "%s\n", b)
				}
			}
		}
	}
	io.WriteString(tempFile, "\n# Uncomment to assign labels\n")
	for _, l := range labels {
		// if labels are passed as command line input, uncomment them
		if labelSet[l] {
			fmt.Fprintf(tempFile, "Label: %s\n", l)
			continue
		}
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

	// pre process template
	if settings.PreProcess != "" {
		cmd := exec.CommandContext(ctx, settings.PreProcess, tempFile.Name())
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("error running pre process template: %s:\n  %s", settings.PreProcess, out)
			return nil, err
		}
	}

	cmd := exec.CommandContext(ctx, settings.Editor, tempFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		tempFile.Close()
		// os.Remove(tempFile.Name())
		return nil, err
	}
	if cmd.ProcessState != nil && !cmd.ProcessState.Success() {
		return nil, fmt.Errorf("non-zero exit code from editor")
	}

	// post process template
	if settings.PostProcess != "" {
		cmd = exec.CommandContext(ctx, settings.PostProcess, tempFile.Name())
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("error running post process template: %s:\n  %s", settings.PostProcess, out)
			return nil, err
		}
	}

	// re-open the temp file
	tempFile, err = os.Open(tempFile.Name())
	if err != nil {
		return nil, err
	}

	var title string
	var descriptions, selectedLabels []string
	scanner := bufio.NewScanner(tempFile)
	for scanner.Scan() {
		// log.Printf("line %#v", scanner.Text())
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
			return nil, err
		}
	}

	description := strings.TrimSpace(strings.Join(descriptions, "\n"))

	if title == "" {
		return nil, fmt.Errorf("missing title")
	}

	issue := &github.IssueRequest{
		Title:    &title,
		Assignee: &settings.User,
	}
	if description != "" {
		issue.Body = &description
	}
	if len(selectedLabels) > 0 {
		issue.Labels = &selectedLabels
	}

	return issue, nil
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
