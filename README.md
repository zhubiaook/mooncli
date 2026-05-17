# moonai

`moonai` is a small Go-based CLI that uses an LLM to provide focused language utilities from the terminal.

The current commands are:

- `mo ts`: translate between Chinese and English
- `mo vb`: explain English words or phrases for programmers
- `mo ck`: check whether an English sentence is correct and idiomatic

## Features

- Streaming terminal output
- Single-shot mode with command arguments
- Interactive REPL-style mode when no text is passed
- Shared Anthropic-compatible configuration loaded from `~/.claude/settings.json`
- Built with Go and Cobra

## Requirements

- Go `1.26.1`
- Access to an Anthropic-compatible API endpoint
- A local Claude settings file at `~/.claude/settings.json`

## Configuration

The CLI reads configuration from:

```json
~/.claude/settings.json
```

Expected structure:

```json
{
  "env": {
    "ANTHROPIC_AUTH_TOKEN": "your-api-key",
    "ANTHROPIC_BASE_URL": "https://your-api-endpoint",
    "ANTHROPIC_MODEL": "your-model-name"
  }
}
```

## Installation

Build the CLI:

```bash
go build -o mo .
```

Or run it directly:

```bash
go run . --help
```

## Usage

Show help:

```bash
./mo --help
```

### Translation

Translate text directly:

```bash
./mo ts "你好，欢迎使用 moonai"
./mo ts "Please make this sentence more natural."
```

Start interactive translation mode:

```bash
./mo ts
```

Exit interactive mode with:

```text
q
exit
quit
```

### Vocabulary Tutor

Explain a word or phrase directly:

```bash
./mo vb resilient
./mo vb "technical debt"
```

Start interactive vocabulary mode:

```bash
./mo vb
```

### English Sentence Check

Evaluate an English sentence directly:

```bash
./mo ck "I very like this feature."
./mo ck "This sentence sounds natural."
```

Start interactive sentence-check mode:

```bash
./mo ck
```

## Command Behavior

- If you pass text arguments, the CLI sends them as one prompt and prints the streamed response.
- If you do not pass arguments, the CLI enters interactive mode and reads from standard input.
- Retry logic is built in for transient connection errors such as unexpected EOF or connection reset.

## Development

Run locally:

```bash
go run . ts "你好"
go run . vb concurrency
go run . ck "I very like Go."
```

Project structure:

```text
.
├── cmd/
│   ├── root.go
│   ├── check.go
│   ├── translate.go
│   └── vocabulary.go
├── internal/
│   ├── config/
│   │   └── config.go
│   └── llm/
│       └── client.go
├── main.go
└── go.mod
```

## Notes

- The executable command name is `mo`.
- The translation command is intentionally restricted to Chinese-English translation only.
- The vocabulary command is tuned for concise explanations with programmer-friendly examples.
- The sentence-check command only suggests alternatives when the original sentence is incorrect or unnatural.
