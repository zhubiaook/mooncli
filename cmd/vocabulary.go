package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/zhubiaook/moonai/internal/config"
	"github.com/zhubiaook/moonai/internal/tts"
)

const vbSystemPrompt = `You are an English vocabulary tutor for a programmer. Your ONLY job is to explain English words and phrases.

CRITICAL RULES:
1. NEVER greet the user or say "Hello", "Hi", "I'm ready", etc.
2. NEVER ask for confirmation or say "send me a word"
3. NEVER output anything except the vocabulary explanation
4. IMMEDIATELY explain any input using the exact format below
5. Treat EVERY input as a word or phrase to explain - no exceptions

OUTPUT FORMAT (use exactly):

📘 [WORD/PHRASE]  (/IPA/)  [part of speech]

1️⃣  [meaning] - brief clarification
   • "Example sentence with WORD in caps"
   • "Another example with WORD in caps"

2️⃣  [if more meanings, continue...]

🧠 TIP: [one simple memory trick]

🔁 RELATED: word1 • word2 • word3 • word4 • word5

STYLE:
- The user is a programmer. If the word has a specific meaning or usage in Computer Science/Programming, you MUST include an example related to coding/technology.
- Use simple, beginner-level English
- Give common, everyday examples
- Put the target word in UPPERCASE in examples
- Keep output compact and easy to read
- NEVER use Markdown formatting (no **, no *, no #, no backticks). Output plain text only. Emoji are OK.`

func init() {
	noSpeech := false
	repeat := 1
	vbCmd := &cobra.Command{
		Use:   "vb [word or phrase]",
		Short: "English vocabulary tutor",
		Long:  "Explain English words and phrases with examples, tips, and related words.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVocabulary(args, noSpeech, repeat)
		},
	}
	vbCmd.Flags().BoolVarP(&noSpeech, "no-speech", "q", false, "skip pronunciation audio")
	vbCmd.Flags().IntVarP(&repeat, "repeat", "r", 1, "pronunciation replay count (1-10)")
	rootCmd.AddCommand(vbCmd)
}

func runVocabulary(args []string, noSpeech bool, repeat int) error {
	if err := validateVocabularyOptions(noSpeech, repeat); err != nil {
		return err
	}

	if noSpeech {
		return runWithPrompt(args, vbSystemPrompt)
	}

	var client *tts.Client
	warned := false
	disabled := false
	warn := func(format string, a ...any) {
		if warned {
			return
		}
		warned = true
		fmt.Fprintf(os.Stderr, "Warning: pronunciation disabled: "+format+"\n", a...)
	}

	return runWithPromptOptions(args, vbSystemPrompt, promptOptions{
		BeforePrompt: func(ctx context.Context, lookupText string) error {
			if disabled {
				return nil
			}
			if client == nil {
				cfg, err := config.Load()
				if err != nil {
					warn("%v", err)
					disabled = true
					return nil
				}
				client, err = tts.NewClient(cfg.TTS)
				if err != nil {
					warn("%v", err)
					disabled = true
					return nil
				}
			}
			if err := client.SpeakRepeat(ctx, lookupText, repeat); err != nil {
				warn("%v", err)
				disabled = true
			}
			return nil
		},
	})
}

func validateVocabularyOptions(noSpeech bool, repeat int) error {
	if noSpeech {
		return nil
	}
	return validatePronunciationRepeat(repeat)
}

func validatePronunciationRepeat(repeat int) error {
	if repeat < 1 || repeat > 10 {
		return fmt.Errorf("repeat must be between 1 and 10")
	}
	return nil
}
