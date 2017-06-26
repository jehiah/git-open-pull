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
