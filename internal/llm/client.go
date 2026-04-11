package llm

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"charm.land/fantasy"
	"charm.land/fantasy/providers/anthropic"
	"github.com/zhubiaook/moonai/internal/config"
)

const maxRetries = 3

type Client struct {
	agent fantasy.Agent
}

func NewClient(systemPrompt string) (*Client, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	provider, err := anthropic.New(
		anthropic.WithBaseURL(cfg.BaseURL),
		anthropic.WithAPIKey(cfg.APIKey),
	)
	if err != nil {
		return nil, fmt.Errorf("create provider: %w", err)
	}

	ctx := context.Background()
	model, err := provider.LanguageModel(ctx, cfg.Model)
	if err != nil {
		return nil, fmt.Errorf("create language model: %w", err)
	}

	agent := fantasy.NewAgent(model, fantasy.WithSystemPrompt(systemPrompt))

	return &Client{agent: agent}, nil
}

func isRetryable(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.EOF) {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "EOF") || strings.Contains(msg, "connection reset")
}

func (c *Client) Stream(ctx context.Context, prompt string, onText func(string)) error {
	var lastErr error
	for attempt := range maxRetries {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
		}
		_, err := c.agent.Stream(ctx, fantasy.AgentStreamCall{
			Prompt: prompt,
			OnTextDelta: func(id, text string) error {
				onText(text)
				return nil
			},
		})
		if err == nil {
			return nil
		}
		lastErr = err
		if !isRetryable(err) {
			return err
		}
	}
	return lastErr
}

func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	var lastErr error
	for attempt := range maxRetries {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
		}
		result, err := c.agent.Generate(ctx, fantasy.AgentCall{Prompt: prompt})
		if err == nil {
			return result.Response.Content.Text(), nil
		}
		lastErr = err
		if !isRetryable(err) {
			return "", err
		}
	}
	return "", lastErr
}
