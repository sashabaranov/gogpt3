package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/sashabaranov/go-openai"
)

func main() {
	// Initialize the OpenAI client with the API key from the environment variable
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	// Create a ChatCompletionRequest with the initial system message
	req := openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo, // Specifies the model to use
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "you are a helpful chatbot",// System message setting the chatbot's role	
			},
		},
	}

	// Start the conversation loop
	fmt.Println("Conversation")
	fmt.Println("---------------------")
	fmt.Print("> ")
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		
		// Append the user's input to the request messages
		req.Messages = append(req.Messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: s.Text(),
		})
		// Send the request to the OpenAI API
		resp, err := client.CreateChatCompletion(context.Background(), req)
		if err != nil {
			fmt.Printf("ChatCompletion error: %v\n", err)
			continue
		}
		fmt.Printf("%s\n\n", resp.Choices[0].Message.Content)
		// Append the chatbot's response to the conversation
		req.Messages = append(req.Messages, resp.Choices[0].Message)
		fmt.Print("> ")
	}
}
