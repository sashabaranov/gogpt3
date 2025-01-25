package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/alejandrojnm/go-openai"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("please provide a filename to convert to text")
		return
	}
	if _, err := os.Stat(os.Args[1]); errors.Is(err, os.ErrNotExist) {
		fmt.Printf("file %s does not exist\n", os.Args[1])
		return
	}

	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	resp, err := client.CreateTranscription(
		context.Background(),
		openai.AudioRequest{
			Model:    openai.Whisper1,
			FilePath: os.Args[1],
		},
	)
	if err != nil {
		fmt.Printf("Transcription error: %v\n", err)
		return
	}
	fmt.Println(resp.Text)
}
