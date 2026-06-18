# moonai

`moonai` is a small Go-based CLI that provides focused language utilities from the terminal.

The current commands are:

- `mo ts`: translate between Chinese and English
- `mo vb`: explain English words or phrases for programmers
- `mo ck`: check whether an English sentence is correct and idiomatic
- `mo voice` / `mo vo`: speak text with text-to-speech

## Features

- Streaming terminal output
- Single-shot mode with command arguments
- Interactive REPL-style mode when no text is passed
- Shared Anthropic-compatible configuration loaded from `~/.mooncli/settings.json`
- Optional Volcengine V3 pronunciation audio for vocabulary lookups
- Standalone Volcengine V3 text-to-speech playback
- Built with Go and Cobra

## Requirements

- Go `1.26.1`
- Access to an Anthropic-compatible API endpoint
- A local Moon CLI settings file at `~/.mooncli/settings.json`
- macOS `afplay` for vocabulary pronunciation audio
- macOS `afplay` for standalone text-to-speech playback

## Configuration

The CLI reads configuration from:

```json
~/.mooncli/settings.json
```

Expected structure:

```json
{
  "env": {
    "ANTHROPIC_AUTH_TOKEN": "your-api-key",
    "ANTHROPIC_BASE_URL": "https://your-api-endpoint",
    "ANTHROPIC_MODEL": "your-model-name",

    "VOLCENGINE_TTS_API_KEY": "your-volcengine-api-key",
    "VOLCENGINE_TTS_RESOURCE_ID": "seed-tts-2.0",
    "VOLCENGINE_TTS_VOICE_TYPE": "your-voice-type",
    "VOLCENGINE_TTS_ENDPOINT": "https://openspeech.bytedance.com/api/v3/tts/unidirectional"
  }
}
```

`VOLCENGINE_TTS_ENDPOINT` is optional. If omitted, TTS commands use the Volcengine V3 HTTP unidirectional endpoint. TTS requests synthesize MP3 audio at 24000 Hz and play it with `afplay`.

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
./mo vb hello --repeat 3
./mo vb hello -r 3
./mo vb resilient --no-speech
./mo vb resilient -q
```

Start interactive vocabulary mode:

```bash
./mo vb
./mo vb -r 3
./mo vb --no-speech
./mo vb -q
```

When speech is enabled, `mo vb` pronounces the original lookup text before printing the vocabulary explanation. Use `--repeat` or `-r` to replay the same synthesized pronunciation from 1 to 10 times; the vocabulary explanation is still printed once. Use `-q` or `--no-speech` to skip pronunciation audio. If pronunciation configuration, network access, or local playback fails, the command prints one warning and still prints the vocabulary explanation.

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

### Text To Speech

Speak text directly:

```bash
./mo voice "Hello from Moon CLI"
./mo vo "Hello from Moon CLI"
./mo vo "Replay this" --repeat 3 --interval 500ms
./mo vo "Replay this" -r 3 -i 2s
```

Start interactive text-to-speech mode:

```bash
./mo vo
./mo vo -r 3 -i 500ms
```

`mo voice` synthesizes the input text once, then plays the same audio 5 times by default with a 1 second interval between playbacks. Use `--repeat` or `-r` to set the replay count from 1 to 10. Use `--interval` or `-i` with Go duration syntax, such as `500ms`, `2s`, or `1m`. The interval is only applied between playbacks.

## Command Behavior

- If you pass text arguments, the CLI sends them as one prompt and prints the streamed response.
- If you do not pass arguments, the CLI enters interactive mode and reads from standard input.
- Retry logic is built in for transient connection errors such as unexpected EOF or connection reset.
- `mo vb` pronunciation is best-effort and never blocks the vocabulary explanation after a failure.
- `mo voice` exits with an error when configuration, synthesis, or playback fails; in interactive mode it prints the error and keeps reading input.

## Development

Run locally:

```bash
go run . ts "你好"
go run . vb concurrency
go run . ck "I very like Go."
go run . vo "hello"
```

Project structure:

```text
.
├── cmd/
│   ├── root.go
│   ├── check.go
│   ├── translate.go
│   ├── vocabulary.go
│   └── voice.go
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
- The vocabulary command pronounces only the original lookup text, not the generated explanation.
- The sentence-check command only suggests alternatives when the original sentence is incorrect or unnatural.
- The voice command speaks exactly the input text and does not call the LLM.
