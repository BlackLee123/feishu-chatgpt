package openai

import (
	"context"
	"errors"
	"fmt"
	"io"

	go_openai "github.com/sashabaranov/go-openai"
)

func (c *ChatGPT) StreamChat(ctx context.Context, msg []go_openai.ChatCompletionMessage, mode AIMode, responseStream chan<- string, done chan struct{}) error {
	//change msg type from Messages to openai.ChatCompletionMessage
	chatMsgs := make([]go_openai.ChatCompletionMessage, len(msg))
	for i, m := range msg {
		chatMsgs[i] = go_openai.ChatCompletionMessage{
			Role:    m.Role,
			Content: m.Content,
		}
	}
	req := go_openai.ChatCompletionRequest{
		Model:     go_openai.GPT4o,
		Messages:  chatMsgs,
		MaxTokens: 2000,
		Stream:    true,
	}
	stream, err := c.Client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		fmt.Printf("ChatCompletionStream error: %v\n", err)
		return err
	}

	defer stream.Close()
	defer close(done)
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			fmt.Println("Stream closed")
			return nil
		}
		if err != nil {
			fmt.Printf("Stream error: %v\n", err)
			return err
		}
		fmt.Println("answer: ", response.Choices[0].Delta.Content)
		responseStream <- response.Choices[0].Delta.Content
	}
}
