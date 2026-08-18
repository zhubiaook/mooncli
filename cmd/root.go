package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zhubiaook/moonai/internal/llm"
)

var rootCmd = &cobra.Command{
	Use:   "mo",
	Short: "A CLI tool for language and speech utilities",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

type promptOptions struct {
	BeforePrompt func(context.Context, string) error
}

func runWithPrompt(args []string, systemPrompt string) error {
	return runWithPromptOptions(args, systemPrompt, promptOptions{})
}

func runWithPromptOptions(args []string, systemPrompt string, opts promptOptions) error {
	ctx := context.Background()
	var client *llm.Client
	stream := func(prompt string) error {
		if client == nil {
			var err error
			client, err = llm.NewClient(systemPrompt)
			if err != nil {
				return err
			}
		}
		return client.Stream(ctx, prompt, func(text string) {
			fmt.Print(text)
		})
	}

	if len(args) > 0 {
		question := strings.Join(args, " ")
		if opts.BeforePrompt != nil {
			if err := opts.BeforePrompt(ctx, question); err != nil {
				return err
			}
		}
		if err := stream(question); err != nil {
			return err
		}
		fmt.Println()
		return nil
	}

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
		if opts.BeforePrompt != nil {
			if err := opts.BeforePrompt(ctx, input); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				continue
			}
		}
		if err := stream(input); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		fmt.Println()
	}
}
