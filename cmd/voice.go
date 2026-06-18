package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/zhubiaook/moonai/internal/config"
	"github.com/zhubiaook/moonai/internal/tts"
)

func init() {
	repeat := 5
	interval := time.Second
	voiceCmd := &cobra.Command{
		Use:     "voice [text]",
		Aliases: []string{"vo"},
		Short:   "Speak text with text-to-speech",
		Long:    "Synthesize text with Volcengine TTS and play it locally.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVoice(args, repeat, interval)
		},
	}
	voiceCmd.Flags().IntVarP(&repeat, "repeat", "r", 5, "speech replay count (1-10)")
	voiceCmd.Flags().DurationVarP(&interval, "interval", "i", time.Second, "interval between replays")
	rootCmd.AddCommand(voiceCmd)
}

func runVoice(args []string, repeat int, interval time.Duration) error {
	if err := validateVoiceOptions(repeat, interval); err != nil {
		return err
	}

	ctx := context.Background()
	if len(args) > 0 {
		client, err := newVoiceClient()
		if err != nil {
			return err
		}
		return client.SpeakRepeatInterval(ctx, strings.Join(args, " "), repeat, interval)
	}

	var client *tts.Client
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		switch input {
		case "q", "exit", "quit":
			return nil
		case "":
			continue
		}
		if client == nil {
			var err error
			client, err = newVoiceClient()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				continue
			}
		}
		if err := client.SpeakRepeatInterval(ctx, input, repeat, interval); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
	}
	return scanner.Err()
}

func newVoiceClient() (*tts.Client, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	return tts.NewClient(cfg.TTS)
}

func validateVoiceOptions(repeat int, interval time.Duration) error {
	if err := validatePronunciationRepeat(repeat); err != nil {
		return err
	}
	if interval < 0 {
		return fmt.Errorf("interval must be at least 0")
	}
	return nil
}
