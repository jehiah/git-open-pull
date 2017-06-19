package input

import (
	"bytes"
	"fmt"
)

func (i *UI) Ask(query, defaultVal string) (string, error) {
	var buf bytes.Buffer
	buf.WriteString(query)
	if defaultVal != "" {
		buf.WriteString(fmt.Sprintf(" (Default is %s)", defaultVal))
	}

	buf.WriteString(": ")
	i.print(buf.String())

	line, err := i.readline()
	if line == "" {
		line = defaultVal
	}
	i.print("\n")
	return line, err
}

func Ask(query, defaultVal string) (string, error) {
	return Default.Ask(query, defaultVal)
}

// ValidateFunc is function to validate the user input.
//
// The following example shows validating the user input is
// 'Y' or 'n' when asking yes or no question.
type ValidateFunc func(string) error

func (i *UI) AskValidate(query, defaultVal string, v ValidateFunc) (string, error) {
	for {
		l, err := i.Ask(query, defaultVal)
		if err != nil {
			return "", err
		}
		if v != nil {
			err = v(l)
			if err != nil {
				i.print(fmt.Sprintf("Failed to validate input string: %s\n\n", err))
				continue
			}
		}
		return l, err
	}
}

func AskValidate(query, defaultVal string, v ValidateFunc) (string, error) {
	return Default.AskValidate(query, defaultVal, v)
}
