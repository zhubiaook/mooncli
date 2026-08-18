package cmd

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/chzyer/readline"
)

func TestReadLookupTextContinuesAfterInterrupt(t *testing.T) {
	reader := &scriptedReader{results: []readerResult{
		{err: readline.ErrInterrupt},
		{line: "hello"},
	}}

	lookupText, ok, err := readLookupText(reader)
	if err != nil {
		t.Fatalf("read lookup text: %v", err)
	}
	if !ok || lookupText != "hello" {
		t.Fatalf("readLookupText() = %q, %t, want %q, true", lookupText, ok, "hello")
	}
}

func TestReadLookupTextStopsAtEOF(t *testing.T) {
	lookupText, ok, err := readLookupText(&scriptedReader{results: []readerResult{{err: io.EOF}}})
	if err != nil {
		t.Fatalf("read lookup text: %v", err)
	}
	if ok || lookupText != "" {
		t.Fatalf("readLookupText() = %q, %t, want empty lookup text and false", lookupText, ok)
	}
}

func TestScannerReaderReadsPipedInputLineByLine(t *testing.T) {
	var output bytes.Buffer
	reader := &scannerReader{
		scanner: bufio.NewScanner(strings.NewReader("hello\n")),
		output:  &output,
	}

	lookupText, ok, err := readLookupText(reader)
	if err != nil {
		t.Fatalf("read lookup text: %v", err)
	}
	if !ok || lookupText != "hello" {
		t.Fatalf("readLookupText() = %q, %t, want %q, true", lookupText, ok, "hello")
	}
	if output.String() != "> " {
		t.Fatalf("prompt = %q, want %q", output.String(), "> ")
	}
}

type readerResult struct {
	line string
	err  error
}

type scriptedReader struct {
	results []readerResult
}

func (r *scriptedReader) Readline() (string, error) {
	result := r.results[0]
	r.results = r.results[1:]
	return result.line, result.err
}

func (r *scriptedReader) Close() error {
	return nil
}
