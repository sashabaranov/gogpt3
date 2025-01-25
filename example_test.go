package openai_test

import (
	"bufio"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/alejandrojnm/go-openai"
)

func Example() {
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Hello!",
				},
			},
		},
	)
	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return
	}

	fmt.Println(resp.Choices[0].Message.Content)
}

func ExampleClient_CreateChatCompletionStream() {
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	stream, err := client.CreateChatCompletionStream(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:     openai.GPT3Dot5Turbo,
			MaxTokens: 20,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Lorem ipsum",
				},
			},
			Stream: true,
		},
	)
	if err != nil {
		fmt.Printf("ChatCompletionStream error: %v\n", err)
		return
	}
	defer stream.Close()

	fmt.Print("Stream response: ")
	for {
		var response openai.ChatCompletionStreamResponse
		response, err = stream.Recv()
		if errors.Is(err, io.EOF) {
			fmt.Println("\nStream finished")
			return
		}

		if err != nil {
			fmt.Printf("\nStream error: %v\n", err)
			return
		}

		fmt.Println(response.Choices[0].Delta.Content)
	}
}

func ExampleClient_CreateCompletion() {
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

func ExampleClient_CreateCompletionStream() {
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	stream, err := client.CreateCompletionStream(
		context.Background(),
		openai.CompletionRequest{
			Model:     openai.GPT3Babbage002,
			MaxTokens: 5,
			Prompt:    "Lorem ipsum",
			Stream:    true,
		},
	)
	if err != nil {
		fmt.Printf("CompletionStream error: %v\n", err)
		return
	}
	defer stream.Close()

	for {
		var response openai.CompletionResponse
		response, err = stream.Recv()
		if errors.Is(err, io.EOF) {
			fmt.Println("Stream finished")
			return
		}

		if err != nil {
			fmt.Printf("Stream error: %v\n", err)
			return
		}

		fmt.Printf("Stream response: %#v\n", response)
	}
}

func ExampleClient_CreateTranscription() {
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	resp, err := client.CreateTranscription(
		context.Background(),
		openai.AudioRequest{
			Model:    openai.Whisper1,
			FilePath: "recording.mp3",
		},
	)
	if err != nil {
		fmt.Printf("Transcription error: %v\n", err)
		return
	}
	fmt.Println(resp.Text)
}

func ExampleClient_CreateTranscription_captions() {
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	resp, err := client.CreateTranscription(
		context.Background(),
		openai.AudioRequest{
			Model:    openai.Whisper1,
			FilePath: os.Args[1],
			Format:   openai.AudioResponseFormatSRT,
		},
	)
	if err != nil {
		fmt.Printf("Transcription error: %v\n", err)
		return
	}
	f, err := os.Create(os.Args[1] + ".srt")
	if err != nil {
		fmt.Printf("Could not open file: %v\n", err)
		return
	}
	defer f.Close()
	if _, err = f.WriteString(resp.Text); err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
		return
	}
}

func ExampleClient_CreateTranslation() {
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	resp, err := client.CreateTranslation(
		context.Background(),
		openai.AudioRequest{
			Model:    openai.Whisper1,
			FilePath: "recording.mp3",
		},
	)
	if err != nil {
		fmt.Printf("Translation error: %v\n", err)
		return
	}
	fmt.Println(resp.Text)
}

func ExampleClient_CreateImage() {
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	respURL, err := client.CreateImage(
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
	fmt.Println(respURL.Data[0].URL)
}

func ExampleClient_CreateImage_base64() {
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	resp, err := client.CreateImage(
		context.Background(),
		openai.ImageRequest{
			Prompt:         "Portrait of a humanoid parrot in a classic costume, high detail, realistic light, unreal engine",
			Size:           openai.CreateImageSize512x512,
			ResponseFormat: openai.CreateImageResponseFormatB64JSON,
			N:              1,
		},
	)
	if err != nil {
		fmt.Printf("Image creation error: %v\n", err)
		return
	}

	b, err := base64.StdEncoding.DecodeString(resp.Data[0].B64JSON)
	if err != nil {
		fmt.Printf("Base64 decode error: %v\n", err)
		return
	}

	f, err := os.Create("example.png")
	if err != nil {
		fmt.Printf("File creation error: %v\n", err)
		return
	}
	defer f.Close()

	_, err = f.Write(b)
	if err != nil {
		fmt.Printf("File write error: %v\n", err)
		return
	}

	fmt.Println("The image was saved as example.png")
}

func ExampleClientConfig_clientWithProxy() {
	config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	port := os.Getenv("OPENAI_PROXY_PORT")
	proxyURL, err := url.Parse(fmt.Sprintf("http://localhost:%s", port))
	if err != nil {
		panic(err)
	}
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}
	config.HTTPClient = &http.Client{
		Transport: transport,
	}

	client := openai.NewClientWithConfig(config)

	client.CreateChatCompletion( //nolint:errcheck // outside of the scope of this example.
		context.Background(),
		openai.ChatCompletionRequest{
			// etc...
		},
	)
}

func Example_chatbot() {
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	req := openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "you are a helpful chatbot",
			},
		},
	}
	fmt.Println("Conversation")
	fmt.Println("---------------------")
	fmt.Print("> ")
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		req.Messages = append(req.Messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: s.Text(),
		})
		resp, err := client.CreateChatCompletion(context.Background(), req)
		if err != nil {
			fmt.Printf("ChatCompletion error: %v\n", err)
			continue
		}
		fmt.Printf("%s\n\n", resp.Choices[0].Message.Content)
		req.Messages = append(req.Messages, resp.Choices[0].Message)
		fmt.Print("> ")
	}
}

func ExampleDefaultAzureConfig() {
	azureKey := os.Getenv("AZURE_OPENAI_API_KEY")       // Your azure API key
	azureEndpoint := os.Getenv("AZURE_OPENAI_ENDPOINT") // Your azure OpenAI endpoint
	config := openai.DefaultAzureConfig(azureKey, azureEndpoint)
	client := openai.NewClientWithConfig(config)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Hello Azure OpenAI!",
				},
			},
		},
	)
	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return
	}

	fmt.Println(resp.Choices[0].Message.Content)
}

// Open-AI maintains clear documentation on how to handle API errors.
//
// see: https://platform.openai.com/docs/guides/error-codes/api-errors
func ExampleAPIError() {
	var err error // Assume this is the error you are checking.
	e := &openai.APIError{}
	if errors.As(err, &e) {
		switch e.HTTPStatusCode {
		case 401:
		// invalid auth or key (do not retry)
		case 429:
		// rate limiting or engine overload (wait and retry)
		case 500:
		// openai server error (retry)
		default:
			// unhandled
		}
	}
}
