package main

import (
	"strconv"
	"strings"
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
