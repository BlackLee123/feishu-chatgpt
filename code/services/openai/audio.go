package openai

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

type AudioToTextRequestBody struct {
	File           string `json:"file"`
	Model          string `json:"model"`
	ResponseFormat string `json:"response_format"`
}

type AudioToTextResponseBody struct {
	Text string `json:"text"`
}

func audioMultipartForm(request AudioToTextRequestBody, w *multipart.Writer) error {
	f, err := os.Open(request.File)
	if err != nil {
		return fmt.Errorf("opening audio file: %w", err)
	}

	fw, err := w.CreateFormFile("file", f.Name())
	if err != nil {
		return fmt.Errorf("creating form file: %w", err)
	}

	if _, err = io.Copy(fw, f); err != nil {
		return fmt.Errorf("reading from opened audio file: %w", err)
	}

	fw, err = w.CreateFormField("model")
	if err != nil {
		return fmt.Errorf("creating form field: %w", err)
	}

	modelName := bytes.NewReader([]byte(request.Model))
	if _, err = io.Copy(fw, modelName); err != nil {
		return fmt.Errorf("writing model name: %w", err)
	}
	w.Close()

	return nil
}

func (gpt *ChatGPT) AudioToText(ctx context.Context, audio string) (string, error) {
	req := openai.AudioRequest{
		Model:    openai.Whisper1,
		FilePath: audio,
		Format:   openai.AudioResponseFormatText,
	}
	resp, err := gpt.Client.CreateTranscription(ctx, req)
	if err != nil {
		fmt.Printf("Transcription error: %v\n", err)
		return "", err
	}
	return resp.Text, nil
}

func (gpt *ChatGPT) TextToSpeech(ctx context.Context, text string, fileKey string) error {
	req := openai.CreateSpeechRequest{
		Model:          openai.TTSModel1,
		Input:          text,
		ResponseFormat: openai.SpeechResponseFormatOpus,
		Voice:          openai.VoiceNova,
	}
	res, err := gpt.Client.CreateSpeech(ctx, req)
	if err != nil {
		return err
	}
	defer res.Close()

	buf, err := io.ReadAll(res)
	if err != nil {
		return err
	}

	// save buf to file as mp3
	err = os.WriteFile(fileKey, buf, 0644)
	if err != nil {
		return err
	}
	return err
}
