package main

import (
	"fmt"
	"testing"
)

func TestDetectIssueNumbert(t *testing.T) {
	type testCase struct {
		branch      string
		issueNumber int
	}
	tests := []testCase{
		{"1_branch", 1},
		{"branch_1", 1},
		{"branch_2_a", 0},
		{"branch_2_33", 33},
		{"1_branch_2", 2},
		{"1-branch", 1},
		{"branch", 0},
	}
	for i, tc := range tests {
		tc := tc
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Parallel()
			if got := DetectIssueNumber(tc.branch); got != tc.issueNumber {
				t.Errorf("got %d expected %d for %q", got, tc.issueNumber, tc.branch)
			}
		})
	}
}
