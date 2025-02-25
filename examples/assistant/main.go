package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/sashabaranov/go-openai"
)

func main() {
	ctx := context.Background()
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	assistantName := "my assistant"
	assistantRq := openai.AssistantRequest{
		Model: openai.GPT4o,
		Name:  &assistantName,
	}
	assistant, err := client.CreateAssistant(ctx, assistantRq)
	if err != nil {
		log.Fatal(err)
	}

	threadRq := openai.ThreadRequest{}
	threadRs, err := client.CreateThread(ctx, threadRq)
	if err != nil {
		log.Fatal(err)
	}

	content := "Hello! Can you help me?"

	messageReq := openai.MessageRequest{
		Role:    openai.ChatMessageRoleUser,
		Content: content,
	}
	_, err = client.CreateMessage(ctx, threadRs.ID, messageReq)
	if err != nil {
		log.Fatal(err)
	}

	runReq := openai.RunRequest{AssistantID: assistant.ID}
	run, err := client.CreateRun(ctx, threadRs.ID, runReq)
	if err != nil {
		log.Fatal(err)
	}

	isCompleted := false

	for !isCompleted {
		resp, err := client.RetrieveRun(ctx, threadRs.ID, run.ID)
		if err != nil {
			log.Fatal(err)
		}

		if resp.Status == openai.RunStatusFailed {
			isCompleted = true
			log.Fatal("Something went wrong. Openai response is failed:", resp.LastError.Message)
		}
		if resp.Status == openai.RunStatusCompleted {
			isCompleted = true
			asc := "asc" // get messages in ascending order (openai api doc)
			limit := 10
			messages, err := client.ListMessage(ctx, threadRs.ID, &limit, &asc, nil, nil, &run.ID)
			if err != nil {
				log.Fatal(err)
			}

			answer := messages.Messages[0].Content[0].Text.Value

			fmt.Print(answer)
		} else {
			time.Sleep(5 * time.Second) // wait 5 second before next request
		}
	}
}
