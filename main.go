package main

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"os/exec"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}

func run(args []string) error {
	var sc *bufio.Scanner
	switch len(args) {
	case 0:
		sc = bufio.NewScanner(os.Stdin)
	case 2:
		cmd := exec.Command("diff", "-u", args[0], args[1])
		cmd.Stderr = os.Stderr
		b, err := cmd.Output()
		if err != nil && len(b) == 0 { // diff exits with 0 only on empty output
			return err
		}
		sc = bufio.NewScanner(bytes.NewReader(b))
	default:
		return errors.New("usage: colordiff file1 file2\nor: diff -u file1 file2 | colordiff")
	}
	w := os.Stdout
	var text []byte
	for sc.Scan() {
		text = text[:0]
		switch line := sc.Bytes(); {
		case len(line) == 0:
		case bytes.HasPrefix(line, []byte("--- ")):
			fallthrough
		case bytes.HasPrefix(line, []byte("+++ ")):
			text = append(text, colorFileName...)
			text = append(text, line...)
			text = append(text, reset...)
		case line[0] == '-':
			text = append(text, colorLineRemoved...)
			text = append(text, line...)
			text = append(text, reset...)
		case line[0] == '+':
			text = append(text, colorLineAdded...)
			text = append(text, line...)
			text = append(text, reset...)
		case line[0] == '@':
			if i := bytes.Index(line, []byte(" @@")); i != -1 {
				text = append(text, colorContextRange...)
				text = append(text, line[:i+3]...)
				text = append(text, reset...)
				text = append(text, line[i+3:]...)
			} else {
				text = append(text, line...)
			}
		case line[0] == ' ':
			text = append(text, line...)
		default:
			text = append(text, colorMeta...)
			text = append(text, line...)
			text = append(text, reset...)
		}
		text = append(text, '\n')
		w.Write(text)
	}
	if err := sc.Err(); err != nil {
		return err
	}
	if cap(text) != 0 {
		os.Exit(1) // mimic diff behavior
	}
	return nil
}

const (
	colorFileName     = "\x1b[1m" // "\x1b[1;34m"
	colorMeta         = "\x1b[1m"
	colorContextRange = "\x1b[36m"
	colorLineAdded    = "\x1b[32m"
	colorLineRemoved  = "\x1b[33m"
	reset             = "\x1b[0m"
)
