package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zhubiaook/moonai/internal/llm"
)

var rootCmd = &cobra.Command{
	Use:   "mo",
	Short: "A CLI tool that uses LLM to answer questions",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runWithPrompt(args []string, systemPrompt string) error {
	client, err := llm.NewClient(systemPrompt)
	if err != nil {
		return err
	}

	ctx := context.Background()

	if len(args) > 0 {
		question := strings.Join(args, " ")
		if err := client.Stream(ctx, question, func(text string) {
			fmt.Print(text)
		}); err != nil {
			return err
		}
		fmt.Println()
		return nil
	}

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
		if err := client.Stream(ctx, input, func(text string) {
			fmt.Print(text)
		}); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		fmt.Println()
	}
	return scanner.Err()
}
