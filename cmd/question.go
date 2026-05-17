package cmd

import "github.com/spf13/cobra"

const qSystemPrompt = `You answer any question from the user in clear, concise language.

RULES:
- Answer directly.
- Keep the response brief and easy to understand.
- Use the same language as the user's question unless they ask otherwise.
- Do not greet the user.
- Do not ask follow-up questions unless the question cannot be answered safely without clarification.
- NEVER use Markdown formatting (no **, no *, no #, no backticks). Output plain text only.`

func init() {
	qCmd := &cobra.Command{
		Use:   "q [question]",
		Short: "Ask any question",
		Long:  "Answer any question in clear and concise language.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWithPrompt(args, qSystemPrompt)
		},
	}
	rootCmd.AddCommand(qCmd)
}
