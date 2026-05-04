package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v60/github"
	"github.com/jehiah/agentdetection"
	"github.com/jehiah/git-open-pull/internal/input"
	"golang.org/x/oauth2"
)

//go:embed SKILL.md
var skillDoc string

func RenameBranch(ctx context.Context, branch string, issueNumber int) (string, error) {
	branch = fmt.Sprintf("%s_%d", branch, issueNumber)
	_, err := RunGit(ctx, "branch", "-m", branch)
	if err != nil {
		return "", err
	}
	return branch, nil
}

func SetupClient(ctx context.Context, s *Settings) *github.Client {
	if s == nil {
		panic("missing settings")
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: s.Token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	client.UserAgent = fmt.Sprintf("git-open-pull/%s (+http://github.com/jehiah/git-open-pull)", Version)
	return client
}

// GetIssueNumber prompts to create a new issue, or confirmation of auto-detected issue number
func GetIssueNumber(ctx context.Context, client *github.Client, settings *Settings, detected int, interactive bool, title, description string, labels []string) (int, error) {
	var issue int
	if detected == 0 {
		var err error
		n := "c"
		if interactive {
			n, err = input.Ask("enter issue number (or 'c' to create)", "")
			if err != nil {
				return issue, err
			}
		}
		switch n {
		case "", "c", "C":
			return NewIssue(ctx, client, settings, interactive, title, description, labels)
		default:
			return strconv.Atoi(n)
		}
	}

	if interactive {
		n, err := input.Ask(fmt.Sprintf("issue number [%d]", detected), "")
		if err != nil {
			log.Fatal(err)
		}
		if n == "" {
			return detected, nil
		}
		return strconv.Atoi(n)
	}

	return detected, nil
}

func printUsage(settings *Settings) {
	out := flag.CommandLine.Output()
	fmt.Fprintln(out, "git-open-pull creates an issue, renames the local branch to include that issue number, pushes the renamed branch and finally converts the issue into a pull request against the renamed branch.")
	fmt.Fprintln(out, "Functionally similar to 'gh pr create'.")
	if settings != nil && settings.User != "" && settings.BaseAccount != "" && settings.BaseRepo != "" {
		fmt.Fprintf(out, "By default, code is pushed to %s/%s and the pull request targets %s/%s branch %s.\n", settings.User, settings.BaseRepo, settings.BaseAccount, settings.BaseRepo, settings.BaseBranch)
	}
	fmt.Fprintln(out)

	if settings != nil {
		if hints := settings.RequiredHints(); len(hints) > 0 {
			fmt.Fprintln(out, "Required configuration missing:")
			for _, h := range hints {
				fmt.Fprintf(out, "  - %s\n", h)
			}
			fmt.Fprintln(out)
		}
	}

	if agentdetection.IsAgent() {
		fmt.Fprintln(out, "Agent Hint: run --list-labels first to inspect the valid repository labels before passing --labels.")
		fmt.Fprintln(out, "Do not push code prior to running, as the branch will be renamed to include the issue number.")
		fmt.Fprintln(out, "Run --skill for full agent usage documentation, or redirect to a file: git-open-pull --skill > SKILL.md")
		fmt.Fprintln(out)
	}
	fmt.Fprintf(out, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	ctx := context.Background()
	preSettings, _ := readSettingsConfig(ctx)

	flag.Usage = func() {
		printUsage(preSettings)
	}

	description := flag.String("description-file", "", "Path to PR description file")
	listLabels := flag.Bool("list-labels", false, "List available issue labels and exit")
	skill := flag.Bool("skill", false, "Print agent skill documentation and exit")
	labels := flag.String("labels", "", "Comma separated PR Labels")
	title := flag.String("title", "", "PR Title")
	interactive := flag.Bool("interactive", true, "Toggles interactive mode")
	version := flag.Bool("version", false, "Prints current version")
	draft := flag.Bool("draft", true, "Open PR in draft mode (non-interactive only)")

	flag.Parse()

	if *version {
		fmt.Printf("git-open-pull v%s %s\n", Version, runtime.Version())
		os.Exit(0)
	}

	if *skill {
		fmt.Print(skillDoc)
		return
	}

	var settings *Settings
	var err error
	if *interactive {
		settings, err = LoadSettings(ctx)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		settings, err = readSettingsConfig(ctx)
		if err != nil {
			log.Fatal(err)
		}
		if hints := settings.RequiredHints(); len(hints) > 0 {
			for _, h := range hints {
				fmt.Fprintln(os.Stderr, h)
			}
			os.Exit(1)
		}
	}

	client := SetupClient(ctx, settings)

	if *listLabels {
		labels, err := Labels(ctx, client, settings)
		if err != nil {
			log.Fatal(err)
		}
		for _, label := range labels {
			fmt.Println(label)
		}
		return
	}

	var labelSlice []string
	if *labels != "" {
		labelSlice = strings.Split(*labels, ",")
		for idx := range labelSlice {
			labelSlice[idx] = strings.TrimSpace(labelSlice[idx])
		}
	}

	var descriptionString string
	if *description != "" {
		fileContent, err := os.ReadFile(*description)
		if err != nil {
			log.Fatalf("error reading description file: %v", err)
		}
		descriptionString = string(fileContent)
	}

	// Validate flag combinations
	if !*interactive && *description != "" && *title == "" {
		log.Fatal("--title is required when using --description-file with --interactive=false")
	}

	branch, err := GitFeatureBranch(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("current branch %s\n", branch)
	switch branch {
	case "main", "master":
		yn, err := input.Ask(fmt.Sprintf("Are you sure you want to make a pull request from %s? [y/N]", branch), "")
		if err != nil {
			log.Fatal(err)
		}
		if yn != "y" && yn != "Y" {
			os.Exit(1)
		}
	}
	detectedIssueNumber := DetectIssueNumber(branch)

	// create issue if needed
	issueNumber, err := GetIssueNumber(ctx, client, settings, detectedIssueNumber, *interactive, *title, descriptionString, labelSlice)
	if err != nil {
		log.Fatal(err)
	}
	if issueNumber == 0 {
		log.Fatal("expected issue number")
	}

	// Do we need/want to rename the branch?
	if issueNumber != detectedIssueNumber {
		if *interactive {
			yn, err := input.Ask(fmt.Sprintf("rename branch to %s_%d [Y/n]", branch, issueNumber), "")
			if err != nil {
				log.Fatal(err)
			}
			switch yn {
			case "", "y", "Y":
				branch, err = RenameBranch(ctx, branch, issueNumber)
				if err != nil {
					log.Fatal(err)
				}
			case "n", "N":
			default:
				log.Fatalf("unknown response %q", yn)
			}
		} else {
			branch, err = RenameBranch(ctx, branch, issueNumber)
			if err != nil {
				log.Fatal(err)
			}
		}

	}

	// confirm issue number is valid and issue is open
	issue, _, err := client.Issues.Get(ctx, settings.BaseAccount, settings.BaseRepo, issueNumber)
	if err != nil {
		log.Fatalf("error verifying issue %d %s", issueNumber, err)
	}
	if *issue.State != "open" {
		log.Fatalf("error: Issue %s/%s#%d is %s (%s)", settings.BaseAccount, settings.BaseRepo, issueNumber, *issue.State, *issue.Title)
	}

	fmt.Printf("pushing branch %s to %s\n", branch, settings.User)
	_, err = RunGit(ctx, "push", "-u", settings.User, branch)
	if err != nil {
		log.Fatal(err)
	}

	// GitHub needs a variable amount of time before a new branch
	// can be used to open a pull request. This is usually enough.
	time.Sleep(2 * time.Second)

	// check branch exists on remote
	branches, _, err := client.Repositories.ListBranches(ctx, settings.User, settings.BaseRepo, &github.BranchListOptions{ListOptions: github.ListOptions{PerPage: 100}})
	if err != nil {
		log.Fatal(err)
	}
	var foundBranch bool
	for _, b := range branches {
		if *b.Name == branch {
			foundBranch = true
		}
	}
	if !foundBranch {
		fmt.Printf("Error: branch %s does not exist in %s/%s\n", branch, settings.User, settings.BaseRepo)
		if len(branches) > 1 {
			fmt.Printf("valid branches are:")
			for i, b := range branches {
				if i > 0 {
					fmt.Print(", ")
				}
				fmt.Printf("%s", *b.Name)
			}
		}
		os.Exit(1)
	}

	fmt.Printf("Issue: %d (%s)\n", issueNumber, *issue.Title)
	head := fmt.Sprintf("%s:%s", settings.User, branch)
	fmt.Printf("pulling from %s into %s/%s branch %s\n", head, settings.BaseAccount, settings.BaseRepo, settings.BaseBranch)
	if *interactive {
		yn, err := input.Ask("confirm [y/n]", "")
		if err != nil {
			log.Fatal(err)
		}
		if strings.ToLower(yn) != "y" {
			log.Fatal("exiting")
		}
	}

	if *interactive {
		yn, err := input.Ask("Open as draft [Y/n]", "")
		if err != nil {
			log.Fatal(err)
		}
		switch yn {
		case "", "Y", "y":
			*draft = true
		case "n", "N":
			*draft = false
		}
	}

	// convert Issue to PR
	params := &github.NewPullRequest{
		Issue:               &issueNumber,
		Base:                &settings.BaseBranch,
		Head:                &head,
		MaintainerCanModify: &settings.MaintainersCanModify,
		Draft:               draft,
	}
	_, _, err = client.PullRequests.Create(ctx, settings.BaseAccount, settings.BaseRepo, params)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", issue.GetHTMLURL())
	// set asignee (if needed) ?

	if settings.Callback != "" {
		// fetch the json of the current issue
		tempFile, err := os.CreateTemp("", fmt.Sprintf("issue-%d", issueNumber))
		if err != nil {
			log.Fatal(err)
		}
		defer os.Remove(tempFile.Name())
		req, err := client.NewRequest("GET", fmt.Sprintf("/repos/%s/%s/pulls/%d", settings.BaseAccount, settings.BaseRepo, issueNumber), nil)
		if err != nil {
			log.Fatal(err)
		}

		resp, err := client.Do(ctx, req, tempFile)
		tempFile.Sync()
		tempFile.Close()
		if err != nil {
			log.Fatal(err)
		}
		if resp.StatusCode != 200 {
			log.Fatalf("got unexpected response code %d", resp.StatusCode)
		}

		cmd := exec.CommandContext(ctx, settings.Callback, tempFile.Name())
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("error on callback: %s:\n   %s", settings.Callback, out)
			log.Fatal(err)
		}
	}

}
