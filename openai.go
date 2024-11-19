package internal

import (
	"context"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

func (agent *TranslationAgent) getCompletion(prompt string, systemMessage string) (string, error) {
	config := openai.DefaultConfig(agent.ApiKey)
	config.BaseURL = agent.BaseURL
	client := openai.NewClientWithConfig(config)

	if systemMessage == "" {
		systemMessage = "You are a helpful assistant."
	}

	if agent.ModelName == "" {
		agent.ModelName = "gpt-4o-mini" // Equivalent to "gpt-4-turbo" if that's what you meant by default
	}

	// fmt.Printf("System message: %s\n", systemMessage)
	// fmt.Printf("Prompt: %s\n", prompt)

	request := openai.ChatCompletionRequest{
		Model:       agent.ModelName,
		Temperature: agent.Temperature,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemMessage,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
	}

	resp, err := client.CreateChatCompletion(context.Background(), request)
	if err != nil {
		return "", fmt.Errorf("ChatCompletion error: %v", err)
	}

	if len(resp.Choices) > 0 {
		return resp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no completion choices returned")
}
