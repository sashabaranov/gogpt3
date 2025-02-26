package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/sashabaranov/go-openai"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: go run ./example/<target> <image_path>")
		return
	}

	ctx := context.Background()

	filePath := os.Args[1]

	c := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	fileContent, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	reqBase64 := base64.StdEncoding.EncodeToString(fileContent)

	messages := []openai.ChatCompletionMessage{
		{
			Role: openai.ChatMessageRoleUser,
			MultiContent: []openai.ChatMessagePart{
				{
					Type: openai.ChatMessagePartTypeText,
					Text: "What's in this image?",
				},
				{
					Type: openai.ChatMessagePartTypeImageURL,
					ImageURL: &openai.ChatMessageImageURL{
						URL:    fmt.Sprintf("data:image/jpeg;base64,%s", reqBase64),
						Detail: openai.ImageURLDetailLow,
					},
				},
			},
		},
	}

	maxTokens := 300

	request := openai.ChatCompletionRequest{
		Model:     openai.GPT4VisionPreview,
		Messages:  messages,
		Stream:    true,
		MaxTokens: maxTokens,
	}

	stream, err := c.CreateChatCompletionStream(ctx, request)
	if err != nil {
		fmt.Printf("Completion error: %v\n", err)
		return
	}
	defer stream.Close()

	fmt.Printf("Stream response: ")
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			fmt.Println("Stream finished")
			break
		}

		if err != nil {
			fmt.Printf("Stream error: %v\n", err)
			return
		}

		fmt.Printf(response.Choices[0].Delta.Content)
	}
}
