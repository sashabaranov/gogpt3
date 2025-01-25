package main

import (
	"context"
	"fmt"
	"os"

	"github.com/alejandrojnm/go-openai"
)

func main() {
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	respUrl, err := client.CreateImage(
		context.Background(),
		openai.ImageRequest{
			Prompt:         "Parrot on a skateboard performs a trick, cartoon style, natural light, high detail",
			Size:           openai.CreateImageSize256x256,
			ResponseFormat: openai.CreateImageResponseFormatURL,
			N:              1,
		},
	)
	if err != nil {
		fmt.Printf("Image creation error: %v\n", err)
		return
	}
	fmt.Println(respUrl.Data[0].URL)
}
