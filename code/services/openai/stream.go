package openai

import (
	"context"
	"errors"
	"fmt"
	"io"

	go_openai "github.com/sashabaranov/go-openai"
)

func (c *ChatGPT) StreamChat(ctx context.Context,
	msg []go_openai.ChatCompletionMessage, mode AIMode,
	responseStream chan string) error {
	//change msg type from Messages to openai.ChatCompletionMessage
	chatMsgs := make([]go_openai.ChatCompletionMessage, len(msg))
	for i, m := range msg {
		chatMsgs[i] = go_openai.ChatCompletionMessage{
			Role:    m.Role,
			Content: m.Content,
		}
	}
	return c.StreamChatWithHistory(ctx, chatMsgs, 2000, mode,
		responseStream)
}

func (c *ChatGPT) StreamChatWithHistory(ctx context.Context,
	msg []go_openai.ChatCompletionMessage, maxTokens int,
	aiMode AIMode,
	responseStream chan string,
) error {

	//pp.Printf("client: %v", client)
	//turn aimode to float64()
	var temperature float32
	temperature = float32(aiMode)
	req := go_openai.ChatCompletionRequest{
		Model:       c.Model,
		Messages:    msg,
		N:           1,
		Temperature: temperature,
		MaxTokens:   maxTokens,
		//TopP:        1,
		//Moderation:     true,
		//ModerationStop: true,
	}
	stream, err := c.Client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		fmt.Errorf("CreateCompletionStream returned error: %v", err)
	}

	defer stream.Close()
	for {
		response, err := stream.Recv()
		// fmt.Println("response: ", response)
		if errors.Is(err, io.EOF) {
			//fmt.Println("Stream finished")
			return nil
		}
		if err != nil {
			fmt.Printf("Stream error: %v\n", err)
			return err
		}
		responseStream <- response.Choices[0].Delta.Content
	}
	return nil

}
