# Go OpenAI
[![Go Reference](https://pkg.go.dev/badge/github.com/sashabaranov/go-openai.svg)](https://pkg.go.dev/github.com/sashabaranov/go-openai)
[![Go Report Card](https://goreportcard.com/badge/github.com/sashabaranov/go-openai)](https://goreportcard.com/report/github.com/sashabaranov/go-openai)
[![codecov](https://codecov.io/gh/sashabaranov/go-openai/branch/master/graph/badge.svg?token=bCbIfHLIsW)](https://codecov.io/gh/sashabaranov/go-openai)

This library provides unofficial Go clients for [OpenAI API](https://platform.openai.com/). We support: 

* ChatGPT 4o, o1
* GPT-3, GPT-4
* DALL·E 2, DALL·E 3
* Whisper

## Installation

```
go get github.com/sashabaranov/go-openai
```
Currently, go-openai requires Go version 1.18 or greater.


## Usage

### ChatGPT example usage:

```go
package main

import (
	"context"
	"fmt"
	openai "github.com/sashabaranov/go-openai"
)

func main() {
	client := openai.NewClient("your token")
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

```

### Getting an OpenAI API Key:

1. Visit the OpenAI website at [https://platform.openai.com/account/api-keys](https://platform.openai.com/account/api-keys).
2. If you don't have an account, click on "Sign Up" to create one. If you do, click "Log In".
3. Once logged in, navigate to your API key management page.
4. Click on "Create new secret key".
5. Enter a name for your new key, then click "Create secret key".
6. Your new API key will be displayed. Use this key to interact with the OpenAI API.

**Note:** Your API key is sensitive information. Do not share it with anyone.

### Other examples:

<details>
<summary>ChatGPT streaming completion</summary>

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	openai "github.com/sashabaranov/go-openai"
)

func main() {
	c := openai.NewClient("your token")
	ctx := context.Background()

	req := openai.ChatCompletionRequest{
		Model:     openai.GPT3Dot5Turbo,
		MaxTokens: 20,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "Lorem ipsum",
			},
		},
		Stream: true,
	}
	stream, err := c.CreateChatCompletionStream(ctx, req)
	if err != nil {
		fmt.Printf("ChatCompletionStream error: %v\n", err)
		return
	}
	defer stream.Close()

	fmt.Printf("Stream response: ")
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			fmt.Println("\nStream finished")
			return
		}

		if err != nil {
			fmt.Printf("\nStream error: %v\n", err)
			return
		}

		fmt.Printf(response.Choices[0].Delta.Content)
	}
}
```
</details>

<details>
<summary>GPT-3 completion</summary>

```go
package main

import (
	"context"
	"fmt"
	openai "github.com/sashabaranov/go-openai"
)

func main() {
	c := openai.NewClient("your token")
	ctx := context.Background()

	req := openai.CompletionRequest{
		Model:     openai.GPT3Babbage002,
		MaxTokens: 5,
		Prompt:    "Lorem ipsum",
	}
	resp, err := c.CreateCompletion(ctx, req)
	if err != nil {
		fmt.Printf("Completion error: %v\n", err)
		return
	}
	fmt.Println(resp.Choices[0].Text)
}
```
</details>

<details>
<summary>GPT-3 streaming completion</summary>

```go
package main

import (
	"errors"
	"context"
	"fmt"
	"io"
	openai "github.com/sashabaranov/go-openai"
)

func main() {
	c := openai.NewClient("your token")
	ctx := context.Background()

	req := openai.CompletionRequest{
		Model:     openai.GPT3Babbage002,
		MaxTokens: 5,
		Prompt:    "Lorem ipsum",
		Stream:    true,
	}
	stream, err := c.CreateCompletionStream(ctx, req)
	if err != nil {
		fmt.Printf("CompletionStream error: %v\n", err)
		return
	}
	defer stream.Close()

	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			fmt.Println("Stream finished")
			return
		}

		if err != nil {
			fmt.Printf("Stream error: %v\n", err)
			return
		}


		fmt.Printf("Stream response: %v\n", response)
	}
}
```
</details>

<details>
<summary>Audio Speech-To-Text</summary>

```go
package main

import (
	"context"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

func main() {
	c := openai.NewClient("your token")
	ctx := context.Background()

	req := openai.AudioRequest{
		Model:    openai.Whisper1,
		FilePath: "recording.mp3",
	}
	resp, err := c.CreateTranscription(ctx, req)
	if err != nil {
		fmt.Printf("Transcription error: %v\n", err)
		return
	}
	fmt.Println(resp.Text)
}
```
</details>

<details>
<summary>Audio Captions</summary>

```go
package main

import (
	"context"
	"fmt"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

func main() {
	c := openai.NewClient(os.Getenv("OPENAI_KEY"))

	req := openai.AudioRequest{
		Model:    openai.Whisper1,
		FilePath: os.Args[1],
		Format:   openai.AudioResponseFormatSRT,
	}
	resp, err := c.CreateTranscription(context.Background(), req)
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
	if _, err := f.WriteString(resp.Text); err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
		return
	}
}
```
</details>

<details>
<summary>DALL-E 2 image generation</summary>

```go
package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	openai "github.com/sashabaranov/go-openai"
	"image/png"
	"os"
)

func main() {
	c := openai.NewClient("your token")
	ctx := context.Background()

	// Sample image by link
	reqUrl := openai.ImageRequest{
		Prompt:         "Parrot on a skateboard performs a trick, cartoon style, natural light, high detail",
		Size:           openai.CreateImageSize256x256,
		ResponseFormat: openai.CreateImageResponseFormatURL,
		N:              1,
	}

	respUrl, err := c.CreateImage(ctx, reqUrl)
	if err != nil {
		fmt.Printf("Image creation error: %v\n", err)
		return
	}
	fmt.Println(respUrl.Data[0].URL)

	// Example image as base64
	reqBase64 := openai.ImageRequest{
		Prompt:         "Portrait of a humanoid parrot in a classic costume, high detail, realistic light, unreal engine",
		Size:           openai.CreateImageSize256x256,
		ResponseFormat: openai.CreateImageResponseFormatB64JSON,
		N:              1,
	}

	respBase64, err := c.CreateImage(ctx, reqBase64)
	if err != nil {
		fmt.Printf("Image creation error: %v\n", err)
		return
	}

	imgBytes, err := base64.StdEncoding.DecodeString(respBase64.Data[0].B64JSON)
	if err != nil {
		fmt.Printf("Base64 decode error: %v\n", err)
		return
	}

	r := bytes.NewReader(imgBytes)
	imgData, err := png.Decode(r)
	if err != nil {
		fmt.Printf("PNG decode error: %v\n", err)
		return
	}

	file, err := os.Create("example.png")
	if err != nil {
		fmt.Printf("File creation error: %v\n", err)
		return
	}
	defer file.Close()

	if err := png.Encode(file, imgData); err != nil {
		fmt.Printf("PNG encode error: %v\n", err)
		return
	}

	fmt.Println("The image was saved as example.png")
}

```
</details>

<details>
<summary>Configuring proxy</summary>

```go
config := openai.DefaultConfig("token")
proxyUrl, err := url.Parse("http://localhost:{port}")
if err != nil {
	panic(err)
}
transport := &http.Transport{
	Proxy: http.ProxyURL(proxyUrl),
}
config.HTTPClient = &http.Client{
	Transport: transport,
}

c := openai.NewClientWithConfig(config)
```

See also: https://pkg.go.dev/github.com/sashabaranov/go-openai#ClientConfig
</details>

<details>
<summary>ChatGPT support context</summary>

```go
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sashabaranov/go-openai"
)

func main() {
	client := openai.NewClient("your token")
	messages := make([]openai.ChatCompletionMessage, 0)
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Conversation")
	fmt.Println("---------------------")

	for {
		fmt.Print("-> ")
		text, _ := reader.ReadString('\n')
		// convert CRLF to LF
		text = strings.Replace(text, "\n", "", -1)
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: text,
		})

		resp, err := client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model:    openai.GPT3Dot5Turbo,
				Messages: messages,
			},
		)

		if err != nil {
			fmt.Printf("ChatCompletion error: %v\n", err)
			continue
		}

		content := resp.Choices[0].Message.Content
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: content,
		})
		fmt.Println(content)
	}
}
```
</details>

<details>
<summary>Azure OpenAI ChatGPT</summary>

```go
package main

import (
	"context"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

func main() {
	config := openai.DefaultAzureConfig("your Azure OpenAI Key", "https://your Azure OpenAI Endpoint")
	// If you use a deployment name different from the model name, you can customize the AzureModelMapperFunc function
	// config.AzureModelMapperFunc = func(model string) string {
	// 	azureModelMapping := map[string]string{
	// 		"gpt-3.5-turbo": "your gpt-3.5-turbo deployment name",
	// 	}
	// 	return azureModelMapping[model]
	// }

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

```
</details>

<details>
<summary>Embedding Semantic Similarity</summary>

```go
package main

import (
	"context"
	"log"
	openai "github.com/sashabaranov/go-openai"

)

func main() {
	client := openai.NewClient("your-token")

	// Create an EmbeddingRequest for the user query
	queryReq := openai.EmbeddingRequest{
		Input: []string{"How many chucks would a woodchuck chuck"},
		Model: openai.AdaEmbeddingV2,
	}

	// Create an embedding for the user query
	queryResponse, err := client.CreateEmbeddings(context.Background(), queryReq)
	if err != nil {
		log.Fatal("Error creating query embedding:", err)
	}

	// Create an EmbeddingRequest for the target text
	targetReq := openai.EmbeddingRequest{
		Input: []string{"How many chucks would a woodchuck chuck if the woodchuck could chuck wood"},
		Model: openai.AdaEmbeddingV2,
	}

	// Create an embedding for the target text
	targetResponse, err := client.CreateEmbeddings(context.Background(), targetReq)
	if err != nil {
		log.Fatal("Error creating target embedding:", err)
	}

	// Now that we have the embeddings for the user query and the target text, we
	// can calculate their similarity.
	queryEmbedding := queryResponse.Data[0]
	targetEmbedding := targetResponse.Data[0]

	similarity, err := queryEmbedding.DotProduct(&targetEmbedding)
	if err != nil {
		log.Fatal("Error calculating dot product:", err)
	}

	log.Printf("The similarity score between the query and the target is %f", similarity)
}

```
</details>

<details>
<summary>Azure OpenAI Embeddings</summary>

```go
package main

import (
	"context"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

func main() {

	config := openai.DefaultAzureConfig("your Azure OpenAI Key", "https://your Azure OpenAI Endpoint")
	config.APIVersion = "2023-05-15" // optional update to latest API version

	//If you use a deployment name different from the model name, you can customize the AzureModelMapperFunc function
	//config.AzureModelMapperFunc = func(model string) string {
	//    azureModelMapping := map[string]string{
	//        "gpt-3.5-turbo":"your gpt-3.5-turbo deployment name",
	//    }
	//    return azureModelMapping[model]
	//}

	input := "Text to vectorize"

	client := openai.NewClientWithConfig(config)
	resp, err := client.CreateEmbeddings(
		context.Background(),
		openai.EmbeddingRequest{
			Input: []string{input},
			Model: openai.AdaEmbeddingV2,
		})

	if err != nil {
		fmt.Printf("CreateEmbeddings error: %v\n", err)
		return
	}

	vectors := resp.Data[0].Embedding // []float32 with 1536 dimensions

	fmt.Println(vectors[:10], "...", vectors[len(vectors)-10:])
}
```
</details>

<details>
<summary>JSON Schema for function calling</summary>

It is now possible for chat completion to choose to call a function for more information ([see developer docs here](https://platform.openai.com/docs/guides/gpt/function-calling)).

In order to describe the type of functions that can be called, a JSON schema must be provided. Many JSON schema libraries exist and are more advanced than what we can offer in this library, however we have included a simple `jsonschema` package for those who want to use this feature without formatting their own JSON schema payload.

The developer documents give this JSON schema definition as an example:

```json
{
  "name":"get_current_weather",
  "description":"Get the current weather in a given location",
  "parameters":{
    "type":"object",
    "properties":{
        "location":{
          "type":"string",
          "description":"The city and state, e.g. San Francisco, CA"
        },
        "unit":{
          "type":"string",
          "enum":[
              "celsius",
              "fahrenheit"
          ]
        }
    },
    "required":[
        "location"
    ]
  }
}
```

Using the `jsonschema` package, this schema could be created using structs as such:

```go
FunctionDefinition{
  Name: "get_current_weather",
  Parameters: jsonschema.Definition{
    Type: jsonschema.Object,
    Properties: map[string]jsonschema.Definition{
      "location": {
        Type: jsonschema.String,
        Description: "The city and state, e.g. San Francisco, CA",
      },
      "unit": {
        Type: jsonschema.String,
        Enum: []string{"celsius", "fahrenheit"},
      },
    },
    Required: []string{"location"},
  },
}
```

The `Parameters` field of a `FunctionDefinition` can accept either of the above styles, or even a nested struct from another library (as long as it can be marshalled into JSON).
</details>

<details>
<summary>Error handling</summary>

Open-AI maintains clear documentation on how to [handle API errors](https://platform.openai.com/docs/guides/error-codes/api-errors)

example:
```
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

```
</details>

<details>
<summary>Fine Tune Model</summary>

```go
package main

import (
	"context"
	"fmt"
	"github.com/sashabaranov/go-openai"
)

func main() {
	client := openai.NewClient("your token")
	ctx := context.Background()

	// create a .jsonl file with your training data for conversational model
	// {"prompt": "<prompt text>", "completion": "<ideal generated text>"}
	// {"prompt": "<prompt text>", "completion": "<ideal generated text>"}
	// {"prompt": "<prompt text>", "completion": "<ideal generated text>"}

	// chat models are trained using the following file format:
	// {"messages": [{"role": "system", "content": "Marv is a factual chatbot that is also sarcastic."}, {"role": "user", "content": "What's the capital of France?"}, {"role": "assistant", "content": "Paris, as if everyone doesn't know that already."}]}
	// {"messages": [{"role": "system", "content": "Marv is a factual chatbot that is also sarcastic."}, {"role": "user", "content": "Who wrote 'Romeo and Juliet'?"}, {"role": "assistant", "content": "Oh, just some guy named William Shakespeare. Ever heard of him?"}]}
	// {"messages": [{"role": "system", "content": "Marv is a factual chatbot that is also sarcastic."}, {"role": "user", "content": "How far is the Moon from Earth?"}, {"role": "assistant", "content": "Around 384,400 kilometers. Give or take a few, like that really matters."}]}

	// you can use openai cli tool to validate the data
	// For more info - https://platform.openai.com/docs/guides/fine-tuning

	file, err := client.CreateFile(ctx, openai.FileRequest{
		FilePath: "training_prepared.jsonl",
		Purpose:  "fine-tune",
	})
	if err != nil {
		fmt.Printf("Upload JSONL file error: %v\n", err)
		return
	}

	// create a fine tuning job
	// Streams events until the job is done (this often takes minutes, but can take hours if there are many jobs in the queue or your dataset is large)
	// use below get method to know the status of your model
	fineTuningJob, err := client.CreateFineTuningJob(ctx, openai.FineTuningJobRequest{
		TrainingFile: file.ID,
		Model:        "davinci-002", // gpt-3.5-turbo-0613, babbage-002.
	})
	if err != nil {
		fmt.Printf("Creating new fine tune model error: %v\n", err)
		return
	}

	fineTuningJob, err = client.RetrieveFineTuningJob(ctx, fineTuningJob.ID)
	if err != nil {
		fmt.Printf("Getting fine tune model error: %v\n", err)
		return
	}
	fmt.Println(fineTuningJob.FineTunedModel)

	// once the status of fineTuningJob is `succeeded`, you can use your fine tune model in Completion Request or Chat Completion Request

	// resp, err := client.CreateCompletion(ctx, openai.CompletionRequest{
	//	 Model:  fineTuningJob.FineTunedModel,
	//	 Prompt: "your prompt",
	// })
	// if err != nil {
	//	 fmt.Printf("Create completion error %v\n", err)
	//	 return
	// }
	//
	// fmt.Println(resp.Choices[0].Text)
}
```
</details>

<details>
<summary>Structured Outputs</summary>

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

func main() {
	client := openai.NewClient("your token")
	ctx := context.Background()

	type Result struct {
		Steps []struct {
			Explanation string `json:"explanation"`
			Output      string `json:"output"`
		} `json:"steps"`
		FinalAnswer string `json:"final_answer"`
	}
	var result Result
	schema, err := jsonschema.GenerateSchemaForType(result)
	if err != nil {
		log.Fatalf("GenerateSchemaForType error: %v", err)
	}
	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4oMini,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are a helpful math tutor. Guide the user through the solution step by step.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "how can I solve 8x + 7 = -23",
			},
		},
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
			JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
				Name:   "math_reasoning",
				Schema: schema,
				Strict: true,
			},
		},
	})
	if err != nil {
		log.Fatalf("CreateChatCompletion error: %v", err)
	}
	err = schema.Unmarshal(resp.Choices[0].Message.Content, &result)
	if err != nil {
		log.Fatalf("Unmarshal schema error: %v", err)
	}
	fmt.Println(result)
}
```
</details>

<details>
<summary>Assistant API example</summary>

```go
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
```
</details>

See the `examples/` folder for more.

## Frequently Asked Questions

### Why don't we get the same answer when specifying a temperature field of 0 and asking the same question?

Even when specifying a temperature field of 0, it doesn't guarantee that you'll always get the same response. Several factors come into play.

1. Go OpenAI Behavior: When you specify a temperature field of 0 in Go OpenAI, the omitempty tag causes that field to be removed from the request. Consequently, the OpenAI API applies the default value of 1.
2. Token Count for Input/Output: If there's a large number of tokens in the input and output, setting the temperature to 0 can still result in non-deterministic behavior. In particular, when using around 32k tokens, the likelihood of non-deterministic behavior becomes highest even with a temperature of 0.

Due to the factors mentioned above, different answers may be returned even for the same question.

**Workarounds:**
1. As of November 2023, use [the new `seed` parameter](https://platform.openai.com/docs/guides/text-generation/reproducible-outputs) in conjunction with the `system_fingerprint` response field, alongside Temperature management.
2. Try using `math.SmallestNonzeroFloat32`: By specifying `math.SmallestNonzeroFloat32` in the temperature field instead of 0, you can mimic the behavior of setting it to 0.
3. Limiting Token Count: By limiting the number of tokens in the input and output and especially avoiding large requests close to 32k tokens, you can reduce the risk of non-deterministic behavior.

By adopting these strategies, you can expect more consistent results.

**Related Issues:**  
[omitempty option of request struct will generate incorrect request when parameter is 0.](https://github.com/sashabaranov/go-openai/issues/9)

### Does Go OpenAI provide a method to count tokens?

No, Go OpenAI does not offer a feature to count tokens, and there are no plans to provide such a feature in the future. However, if there's a way to implement a token counting feature with zero dependencies, it might be possible to merge that feature into Go OpenAI. Otherwise, it would be more appropriate to implement it in a dedicated library or repository.

For counting tokens, you might find the following links helpful:  
- [Counting Tokens For Chat API Calls](https://github.com/pkoukk/tiktoken-go#counting-tokens-for-chat-api-calls)
- [How to count tokens with tiktoken](https://github.com/openai/openai-cookbook/blob/main/examples/How_to_count_tokens_with_tiktoken.ipynb)

**Related Issues:**  
[Is it possible to join the implementation of GPT3 Tokenizer](https://github.com/sashabaranov/go-openai/issues/62)

## Contributing

By following [Contributing Guidelines](https://github.com/sashabaranov/go-openai/blob/master/CONTRIBUTING.md), we hope to ensure that your contributions are made smoothly and efficiently.

## Thank you

We want to take a moment to express our deepest gratitude to the [contributors](https://github.com/sashabaranov/go-openai/graphs/contributors) and sponsors of this project:
- [Carson Kahn](https://carsonkahn.com) of [Spindle AI](https://spindleai.com)

To all of you: thank you. You've helped us achieve more than we ever imagined possible. Can't wait to see where we go next, together!
