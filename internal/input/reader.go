package input

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

// UI is user-interface of input and output.
type UI struct {
	Interactive bool
	Reader      io.ReadCloser
	Writer      io.WriteCloser
	once        sync.Once
}

var Default = &UI{}

func (i *UI) init() {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err == nil {
		i.Interactive = true
		i.Reader = tty
		i.Writer = tty
	} else {
		i.Reader = os.Stdin
		i.Writer = os.Stdout
	}
}

func (i *UI) print(s string) error {
	i.once.Do(i.init)
	_, err := fmt.Fprint(i.Writer, s)
	return err
}

func (i *UI) readline() (string, error) {
	i.once.Do(i.init)
	var err error
	var buffer bytes.Buffer
	var b [1]byte
	for {
		var n int
		n, err = i.Reader.Read(b[:])
		if b[0] == '\n' {
			break
		}
		if n > 0 {
			buffer.WriteByte(b[0])
		}
		if n == 0 || err != nil {
			break
		}
	}

	if err != nil && err != io.EOF {
		return "", err
	}
	return strings.TrimSpace(buffer.String()), err
}
