//go:build !windows

package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/creack/pty"
)

func TestEveryInteractiveCommandEditsLookupTextWithLeftArrow(t *testing.T) {
	binary := buildMoonCLI(t)
	for _, commandName := range []string{"ts", "vb", "ck", "q", "voice"} {
		t.Run(commandName, func(t *testing.T) {
			output := runInTerminal(t, binary, commandName, []byte("xq"), []byte("\x1b[D"), []byte("\x7f\r"))
			if hasTerminalFlag(output, "-icanon") || hasTerminalFlag(output, "-echo") ||
				!hasTerminalFlag(output, "icanon") || !hasTerminalFlag(output, "echo") {
				t.Fatalf("terminal settings were not restored after interactive mode:\n%s", output)
			}
		})
	}
}

func TestInteractiveModeSupportsCursorEditingAndControls(t *testing.T) {
	binary := buildMoonCLI(t)
	tests := []struct {
		name       string
		inputSteps [][]byte
	}{
		{
			name:       "right arrow",
			inputSteps: [][]byte{[]byte("qx"), []byte("\x1b[D"), []byte("\x1b[C"), []byte("\x7f\r")},
		},
		{
			name:       "insertion at cursor",
			inputSteps: [][]byte{[]byte("x"), []byte("\x1b[D"), []byte("q"), []byte("\x04\r")},
		},
		{
			name:       "Unicode lookup text",
			inputSteps: [][]byte{[]byte("你q"), []byte("\x1b[D"), []byte("\x7f\r")},
		},
		{
			name:       "Ctrl-D exits",
			inputSteps: [][]byte{[]byte("\x04")},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runInTerminal(t, binary, "ts", test.inputSteps...)
		})
	}

	t.Run("Ctrl-C cancels the current line", func(t *testing.T) {
		terminal, beforeInput := startInTerminal(t, binary, "ts")
		writeTerminalStep(t, terminal, []byte("discard"), beforeInput)
		writeTerminalStep(t, terminal, []byte("\x03"), beforeInput)
		freshPrompt := readTerminalUntil(t, terminal, []byte("> "))
		writeTerminalStep(t, terminal, []byte("q\r"), freshPrompt)
		finishTerminal(t, terminal)
	})
}

func TestInteractiveModeKeepsHistoryForTheCurrentSession(t *testing.T) {
	command := exec.Command(os.Args[0], "-test.run=^TestInteractiveHistoryHelper$")
	command.Env = append(os.Environ(), "MOONCLI_TEST_HISTORY=1")
	terminal, err := pty.StartWithSize(command, &pty.Winsize{Rows: 24, Cols: 80})
	if err != nil {
		t.Fatalf("start history helper in terminal: %v", err)
	}
	defer terminal.Close()
	defer func() { _ = command.Process.Kill() }()

	readTerminalUntil(t, terminal, []byte("> "))
	if _, err := terminal.Write([]byte("hello\r")); err != nil {
		t.Fatalf("write first history entry: %v", err)
	}
	readTerminalUntil(t, terminal, []byte("READ:hello"))
	time.Sleep(100 * time.Millisecond)
	if _, err := terminal.Write([]byte("\x1b[A")); err != nil {
		t.Fatalf("recall history entry: %v", err)
	}
	time.Sleep(100 * time.Millisecond)
	if _, err := terminal.Write([]byte("\r")); err != nil {
		t.Fatalf("submit recalled history entry: %v", err)
	}
	output := readTerminalUntil(t, terminal, []byte("READ:hello"))
	if !strings.Contains(string(output), "READ:hello") {
		t.Fatalf("recalled line was not submitted: %q", output)
	}
	time.Sleep(100 * time.Millisecond)
	if _, err := terminal.Write([]byte("\x1b[A")); err != nil {
		t.Fatalf("recall history entry before moving down: %v", err)
	}
	time.Sleep(100 * time.Millisecond)
	if _, err := terminal.Write([]byte("\x1b[B")); err != nil {
		t.Fatalf("move down from recalled history entry: %v", err)
	}
	time.Sleep(100 * time.Millisecond)
	if _, err := terminal.Write([]byte("world\r")); err != nil {
		t.Fatalf("submit line after moving down: %v", err)
	}
	output = readTerminalUntil(t, terminal, []byte("READ:world"))
	if !strings.Contains(string(output), "READ:world") {
		t.Fatalf("Down did not restore the current line: %q", output)
	}
	if !bytes.Contains(output, []byte("PASS")) {
		readTerminalUntil(t, terminal, []byte("PASS"))
	}
}

func TestInteractiveHistoryHelper(t *testing.T) {
	if os.Getenv("MOONCLI_TEST_HISTORY") != "1" {
		return
	}

	reader, err := newInteractiveReader()
	if err != nil {
		t.Fatalf("create interactive reader: %v", err)
	}
	defer reader.Close()
	for range 3 {
		line, err := reader.Readline()
		if err != nil {
			t.Fatalf("read interactive line: %v", err)
		}
		fmt.Printf("\nREAD:%s\n", line)
	}
}

func TestPipedInputKeepsLineByLineBehavior(t *testing.T) {
	binary := buildMoonCLI(t)
	for _, commandName := range []string{"ts", "voice"} {
		t.Run(commandName, func(t *testing.T) {
			command := exec.Command(binary, commandName)
			command.Stdin = strings.NewReader("q\n")
			output, err := command.CombinedOutput()
			if err != nil {
				t.Fatalf("run with piped input: %v\n%s", err, output)
			}
			if string(output) != "> " {
				t.Fatalf("piped output = %q, want %q", output, "> ")
			}
		})
	}
}

func hasTerminalFlag(output, flag string) bool {
	for _, field := range strings.Fields(output) {
		if strings.TrimSuffix(field, ";") == flag {
			return true
		}
	}
	return false
}

func buildMoonCLI(t *testing.T) string {
	t.Helper()

	binary := filepath.Join(t.TempDir(), "mo")
	command := exec.Command("go", "build", "-o", binary, "..")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build Moon CLI: %v\n%s", err, output)
	}
	return binary
}

func runInTerminal(t *testing.T, binary, commandName string, inputSteps ...[]byte) string {
	t.Helper()

	terminal, beforeInput := startInTerminal(t, binary, commandName)
	for _, input := range inputSteps {
		writeTerminalStep(t, terminal, input, beforeInput)
	}
	afterInput := finishTerminal(t, terminal)
	output := string(beforeInput) + string(afterInput)
	if strings.Contains(output, "^[[D") || strings.Contains(output, "^[[C") {
		t.Fatalf("cursor escape sequence was displayed as lookup text: %q", output)
	}
	return output
}

func startInTerminal(t *testing.T, binary, commandName string) (*os.File, []byte) {
	t.Helper()

	command := exec.Command("sh", "-c", `"$MOONCLI_TEST_BINARY" "$MOONCLI_TEST_COMMAND"; mooncli_status=$?; stty -a; printf '\nMOONCLI-STATUS:%d\nSHELL-READY\n' "$mooncli_status"`)
	command.Env = append(os.Environ(),
		"MOONCLI_TEST_BINARY="+binary,
		"MOONCLI_TEST_COMMAND="+commandName,
	)
	terminal, err := pty.StartWithSize(command, &pty.Winsize{Rows: 24, Cols: 80})
	if err != nil {
		t.Fatalf("start Moon CLI in terminal: %v", err)
	}
	t.Cleanup(func() {
		_ = terminal.Close()
		_ = command.Process.Kill()
	})

	return terminal, readTerminalUntil(t, terminal, []byte("> "))
}

func writeTerminalStep(t *testing.T, terminal *os.File, input, previousOutput []byte) {
	t.Helper()

	if _, err := terminal.Write(input); err != nil {
		t.Fatalf("write terminal input after %q: %v", previousOutput, err)
	}
	time.Sleep(100 * time.Millisecond)
}

func finishTerminal(t *testing.T, terminal *os.File) []byte {
	t.Helper()

	output := readTerminalUntil(t, terminal, []byte("SHELL-READY"))
	if !bytes.Contains(output, []byte("MOONCLI-STATUS:0")) {
		t.Fatalf("Moon CLI did not exit cleanly: %q", output)
	}
	return output
}

func readTerminalUntil(t *testing.T, terminal *os.File, marker []byte) []byte {
	t.Helper()

	type result struct {
		output []byte
		err    error
	}
	results := make(chan result, 1)
	var output bytes.Buffer
	var outputMu sync.Mutex
	go func() {
		buffer := make([]byte, 256)
		for {
			count, err := terminal.Read(buffer)
			if count > 0 {
				outputMu.Lock()
				output.Write(buffer[:count])
				if bytes.Contains(output.Bytes(), marker) {
					data := bytes.Clone(output.Bytes())
					outputMu.Unlock()
					results <- result{output: data}
					return
				}
				outputMu.Unlock()
			}
			if err != nil {
				outputMu.Lock()
				data := bytes.Clone(output.Bytes())
				outputMu.Unlock()
				results <- result{output: data, err: err}
				return
			}
		}
	}()

	select {
	case result := <-results:
		if result.err != nil {
			t.Fatalf("read terminal until %q: %v\n%s", marker, result.err, fmt.Sprintf("output: %q", string(result.output)))
		}
		return result.output
	case <-time.After(5 * time.Second):
		outputMu.Lock()
		data := bytes.Clone(output.Bytes())
		outputMu.Unlock()
		t.Fatalf("timed out reading terminal until %q\n%s", marker, fmt.Sprintf("output: %q", string(data)))
		return nil
	}
}
