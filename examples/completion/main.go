package main

import (
	"context"
	"fmt"
	"os"

	"github.com/alejandrojnm/go-openai"
)

func main() {
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	resp, err := client.CreateCompletion(
		context.Background(),
		openai.CompletionRequest{
			Model:     openai.GPT3Babbage002,
			MaxTokens: 5,
			Prompt:    "Lorem ipsum",
		},
	)
	if err != nil {
		fmt.Printf("Completion error: %v\n", err)
		return
	}
	fmt.Println(resp.Choices[0].Text)
}
