package main

import (
	"context"
	"fmt"
	"log"
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

}
