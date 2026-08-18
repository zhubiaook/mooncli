package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/chzyer/readline"
)

type interactiveReader interface {
	Readline() (string, error)
	Close() error
}

type scannerReader struct {
	scanner *bufio.Scanner
	output  io.Writer
}

type terminalReader struct {
	*readline.Instance
	input *readline.CancelableStdin
}

func newInteractiveReader() (interactiveReader, error) {
	if readline.DefaultIsTerminal() {
		input := readline.NewCancelableStdin(os.Stdin)
		instance, err := readline.NewEx(&readline.Config{
			Prompt:      "> ",
			HistoryFile: "",
			Stdin:       input,
		})
		if err != nil {
			input.Close()
			return nil, err
		}
		return &terminalReader{Instance: instance, input: input}, nil
	}
	return &scannerReader{
		scanner: bufio.NewScanner(os.Stdin),
		output:  os.Stdout,
	}, nil
}

func (r *terminalReader) Close() error {
	if err := r.input.Close(); err != nil {
		return err
	}
	return r.Instance.Close()
}

func (r *scannerReader) Readline() (string, error) {
	fmt.Fprint(r.output, "> ")
	if r.scanner.Scan() {
		return r.scanner.Text(), nil
	}
	if err := r.scanner.Err(); err != nil {
		return "", err
	}
	return "", io.EOF
}

func (r *scannerReader) Close() error {
	return nil
}

func readLookupText(reader interactiveReader) (string, bool, error) {
	for {
		line, err := reader.Readline()
		switch {
		case errors.Is(err, readline.ErrInterrupt):
			continue
		case errors.Is(err, io.EOF):
			return "", false, nil
		case err != nil:
			return "", false, err
		}

		lookupText := strings.TrimSpace(line)
		switch lookupText {
		case "q", "exit", "quit":
			return "", false, nil
		case "":
			continue
		default:
			return lookupText, true, nil
		}
	}
}
