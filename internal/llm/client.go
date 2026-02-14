// Package llm provides integration with Anthropic's Claude for generating sticker alt-text.
package llm

import (
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// Client wraps the Anthropic client for generating alt-text
type Client struct {
	client    anthropic.Client
	model     string
	maxTokens int64
}

// NewClient creates a new LLM client for alt-text generation
func NewClient(apiKey string, model string, maxTokens int) *Client {
	client := anthropic.NewClient(
		option.WithAPIKey(apiKey),
	)

	return &Client{
		client:    client,
		model:     model,
		maxTokens: int64(maxTokens),
	}
}

// Model returns the configured model name
func (c *Client) Model() string {
	return c.model
}

// MaxTokens returns the configured max tokens
func (c *Client) MaxTokens() int64 {
	return c.maxTokens
}
