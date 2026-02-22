package llm

import (
	"context"
	"errors"
	"io"

	"ai-companion-cli-go/internal/models"
	"github.com/sashabaranov/go-openai"
)

// Client is a wrapper around go-openai for streaming chat completions
type Client struct {
	client       *openai.Client
	modelProfile models.ModelProfile
}

// NewClient creates a new configured wrapper
func NewClient(apiKey string, profile models.ModelProfile) *Client {
	var c *openai.Client
	if apiKey != "" {
		c = openai.NewClient(apiKey)
	}

	return &Client{
		client:       c,
		modelProfile: profile,
	}
}

// SetAPIKey dynamically updates the API key in the client (useful when user inputs in TUI)
func (c *Client) SetAPIKey(apiKey string) {
	c.client = openai.NewClient(apiKey)
}

// EnsureConfigured checks if an API key has been set
func (c *Client) EnsureConfigured() error {
	if c.client == nil {
		return errors.New("openAI API key has not been configured")
	}
	return nil
}

// StreamChat starts a streaming inference and returns a chan of tokens
func (c *Client) StreamChat(ctx context.Context, messages []openai.ChatCompletionMessage, preTemperature float32) (<-chan string, <-chan error) {
	tokenChan := make(chan string)
	errChan := make(chan error, 1)

	if err := c.EnsureConfigured(); err != nil {
		errChan <- err
		close(errChan)
		close(tokenChan)
		return tokenChan, errChan
	}

	req := openai.ChatCompletionRequest{
		Model:       c.modelProfile.PrimaryModel,
		Messages:    messages,
		Stream:      true,
		Temperature: preTemperature,
	}

	// Execute stream asynchronously
	go func() {
		defer close(tokenChan)
		defer close(errChan)

		stream, err := c.client.CreateChatCompletionStream(ctx, req)
		if err != nil {
			errChan <- err
			return
		}
		defer stream.Close()

		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				return
			}
			if err != nil {
				errChan <- err
				return
			}
			if len(response.Choices) > 0 {
				tokenChan <- response.Choices[0].Delta.Content
			}
		}
	}()

	return tokenChan, errChan
}

// GenerateSync does a synchronous (non-streaming) request for background tasks
func (c *Client) GenerateSync(ctx context.Context, systemPrompt string, userPrompt string) (string, error) {
	if err := c.EnsureConfigured(); err != nil {
		return "", err
	}

	req := openai.ChatCompletionRequest{
		Model: c.modelProfile.PrimaryModel,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: userPrompt},
		},
		Temperature: 0.7,
	}

	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) > 0 {
		return resp.Choices[0].Message.Content, nil
	}
	return "", nil
}
