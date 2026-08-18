package cmd

import (
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
	volume := tts.DefaultVolume
	voiceCmd := &cobra.Command{
		Use:     "voice [text]",
		Aliases: []string{"vo"},
		Short:   "Speak text with text-to-speech",
		Long:    "Synthesize text with Volcengine TTS and play it locally.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVoice(args, repeat, interval, volume, cmd.Flags().Changed("volume"))
		},
	}
	voiceCmd.Flags().IntVarP(&repeat, "repeat", "r", 5, "speech replay count (1-100)")
	voiceCmd.Flags().DurationVarP(&interval, "interval", "i", time.Second, "interval between replays")
	voiceCmd.Flags().Float64Var(&volume, "volume", tts.DefaultVolume, "playback volume from 0.0 to 10.0")
	rootCmd.AddCommand(voiceCmd)
}

func runVoice(args []string, repeat int, interval time.Duration, volume float64, volumeOverride bool) error {
	if err := validateVoiceOptions(repeat, interval, volume, volumeOverride); err != nil {
		return err
	}

	ctx := context.Background()
	if len(args) > 0 {
		client, err := newVoiceClient(volume, volumeOverride)
		if err != nil {
			return err
		}
		return client.SpeakRepeatInterval(ctx, strings.Join(args, " "), repeat, interval)
	}

	var client *tts.Client
	reader, err := newInteractiveReader()
	if err != nil {
		return err
	}
	defer reader.Close()

	for {
		input, ok, err := readLookupText(reader)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		if client == nil {
			var err error
			client, err = newVoiceClient(volume, volumeOverride)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				continue
			}
		}
		if err := client.SpeakRepeatInterval(ctx, input, repeat, interval); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
	}
}

func newVoiceClient(volume float64, volumeOverride bool) (*tts.Client, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	if volumeOverride {
		return tts.NewClientWithVolume(cfg.TTS, volume)
	}
	return tts.NewClient(cfg.TTS)
}

func validateVoiceOptions(repeat int, interval time.Duration, volume float64, volumeOverride bool) error {
	if err := validatePronunciationRepeat(repeat); err != nil {
		return err
	}
	if interval < 0 {
		return fmt.Errorf("interval must be at least 0")
	}
	if volumeOverride {
		return tts.ValidateVolume(volume)
	}
	return nil
}
