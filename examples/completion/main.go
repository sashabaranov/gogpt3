package main

import (
	"context"
	"fmt"
	"os"

	"github.com/sashabaranov/go-openai"
)

func main() {
	// Initialize the OpenAI client with the API key from the environment variable
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	// Create a CompletionRequest with the desired model, max tokens, and prompt
	resp, err := client.CreateCompletion(
		context.Background(),
		openai.CompletionRequest{

			Model:     openai.GPT3Babbage002,	// Specifies the model to use (GPT-3 Ada in this case)
			MaxTokens: 5,	// Limits the number of tokens (words/characters) in the response
			Prompt:    "Lorem ipsum",	// The prompt or input text to generate a completion for

		},
	)
	// Handle any errors that occur during the API request
	if err != nil {
		fmt.Printf("Completion error: %v\n", err)
		return
	}

	// Print the generated text completion from the API response
	fmt.Println(resp.Choices[0].Text)
}
