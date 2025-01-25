//go:build integration

package openai_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/alejandrojnm/go-openai"
	"github.com/alejandrojnm/go-openai/internal/test/checks"
	"github.com/alejandrojnm/go-openai/jsonschema"
)

func TestAPI(t *testing.T) {
	apiToken := os.Getenv("OPENAI_TOKEN")
	if apiToken == "" {
		t.Skip("Skipping testing against production OpenAI API. Set OPENAI_TOKEN environment variable to enable it.")
	}

	var err error
	c := openai.NewClient(apiToken)
	ctx := context.Background()
	_, err = c.ListEngines(ctx)
	checks.NoError(t, err, "ListEngines error")

	_, err = c.GetEngine(ctx, openai.GPT3Davinci002)
	checks.NoError(t, err, "GetEngine error")

	fileRes, err := c.ListFiles(ctx)
	checks.NoError(t, err, "ListFiles error")

	if len(fileRes.Files) > 0 {
		_, err = c.GetFile(ctx, fileRes.Files[0].ID)
		checks.NoError(t, err, "GetFile error")
	} // else skip

	embeddingReq := openai.EmbeddingRequest{
		Input: []string{
			"The food was delicious and the waiter",
			"Other examples of embedding request",
		},
		Model: openai.AdaEmbeddingV2,
	}
	_, err = c.CreateEmbeddings(ctx, embeddingReq)
	checks.NoError(t, err, "Embedding error")

	_, err = c.CreateChatCompletion(
		ctx,
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

	checks.NoError(t, err, "CreateChatCompletion (without name) returned error")

	_, err = c.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Name:    "John_Doe",
					Content: "Hello!",
				},
			},
		},
	)
	checks.NoError(t, err, "CreateChatCompletion (with name) returned error")

	_, err = c.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "What is the weather like in Boston?",
				},
			},
			Functions: []openai.FunctionDefinition{{
				Name: "get_current_weather",
				Parameters: jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"location": {
							Type:        jsonschema.String,
							Description: "The city and state, e.g. San Francisco, CA",
						},
						"unit": {
							Type: jsonschema.String,
							Enum: []string{"celsius", "fahrenheit"},
						},
					},
					Required: []string{"location"},
				},
			}},
		},
	)
	checks.NoError(t, err, "CreateChatCompletion (with functions) returned error")
}

func TestCompletionStream(t *testing.T) {
	apiToken := os.Getenv("OPENAI_TOKEN")
	if apiToken == "" {
		t.Skip("Skipping testing against production OpenAI API. Set OPENAI_TOKEN environment variable to enable it.")
	}

	c := openai.NewClient(apiToken)
	ctx := context.Background()

	stream, err := c.CreateCompletionStream(ctx, openai.CompletionRequest{
		Prompt:    "Ex falso quodlibet",
		Model:     openai.GPT3Babbage002,
		MaxTokens: 5,
		Stream:    true,
	})
	checks.NoError(t, err, "CreateCompletionStream returned error")
	defer stream.Close()

	counter := 0
	for {
		_, err = stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			t.Errorf("Stream error: %v", err)
		} else {
			counter++
		}
	}
	if counter == 0 {
		t.Error("Stream did not return any responses")
	}
}

func TestAPIError(t *testing.T) {
	apiToken := os.Getenv("OPENAI_TOKEN")
	if apiToken == "" {
		t.Skip("Skipping testing against production OpenAI API. Set OPENAI_TOKEN environment variable to enable it.")
	}

	var err error
	c := openai.NewClient(apiToken + "_invalid")
	ctx := context.Background()
	_, err = c.ListEngines(ctx)
	checks.HasError(t, err, "ListEngines should fail with an invalid key")

	var apiErr *openai.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("Error is not an APIError: %+v", err)
	}

	if apiErr.HTTPStatusCode != 401 {
		t.Fatalf("Unexpected API error status code: %d", apiErr.HTTPStatusCode)
	}

	switch v := apiErr.Code.(type) {
	case string:
		if v != "invalid_api_key" {
			t.Fatalf("Unexpected API error code: %s", v)
		}
	default:
		t.Fatalf("Unexpected API error code type: %T", v)
	}

	if apiErr.Error() == "" {
		t.Fatal("Empty error message occurred")
	}
}

func TestChatCompletionResponseFormat_JSONSchema(t *testing.T) {
	apiToken := os.Getenv("OPENAI_TOKEN")
	if apiToken == "" {
		t.Skip("Skipping testing against production OpenAI API. Set OPENAI_TOKEN environment variable to enable it.")
	}

	var err error
	c := openai.NewClient(apiToken)
	ctx := context.Background()

	type MyStructuredResponse struct {
		PascalCase string `json:"pascal_case" required:"true" description:"PascalCase"`
		CamelCase  string `json:"camel_case" required:"true" description:"CamelCase"`
		KebabCase  string `json:"kebab_case" required:"true" description:"KebabCase"`
		SnakeCase  string `json:"snake_case" required:"true" description:"SnakeCase"`
	}
	var result MyStructuredResponse
	schema, err := jsonschema.GenerateSchemaForType(result)
	if err != nil {
		t.Fatal("CreateChatCompletion (use json_schema response) GenerateSchemaForType error")
	}
	resp, err := c.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4oMini,
			Messages: []openai.ChatCompletionMessage{
				{
					Role: openai.ChatMessageRoleSystem,
					Content: "Please enter a string, and we will convert it into the following naming conventions:" +
						"1. PascalCase: Each word starts with an uppercase letter, with no spaces or separators." +
						"2. CamelCase: The first word starts with a lowercase letter, " +
						"and subsequent words start with an uppercase letter, with no spaces or separators." +
						"3. KebabCase: All letters are lowercase, with words separated by hyphens `-`." +
						"4. SnakeCase: All letters are lowercase, with words separated by underscores `_`.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Hello World",
				},
			},
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
				JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
					Name:   "cases",
					Schema: schema,
					Strict: true,
				},
			},
		},
	)
	checks.NoError(t, err, "CreateChatCompletion (use json_schema response) returned error")
	if err == nil {
		err = schema.Unmarshal(resp.Choices[0].Message.Content, &result)
		checks.NoError(t, err, "CreateChatCompletion (use json_schema response) unmarshal error")
	}
}

func TestChatCompletionStructuredOutputsFunctionCalling(t *testing.T) {
	apiToken := os.Getenv("OPENAI_TOKEN")
	if apiToken == "" {
		t.Skip("Skipping testing against production OpenAI API. Set OPENAI_TOKEN environment variable to enable it.")
	}

	var err error
	c := openai.NewClient(apiToken)
	ctx := context.Background()

	resp, err := c.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4oMini,
			Messages: []openai.ChatCompletionMessage{
				{
					Role: openai.ChatMessageRoleSystem,
					Content: "Please enter a string, and we will convert it into the following naming conventions:" +
						"1. PascalCase: Each word starts with an uppercase letter, with no spaces or separators." +
						"2. CamelCase: The first word starts with a lowercase letter, " +
						"and subsequent words start with an uppercase letter, with no spaces or separators." +
						"3. KebabCase: All letters are lowercase, with words separated by hyphens `-`." +
						"4. SnakeCase: All letters are lowercase, with words separated by underscores `_`.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Hello World",
				},
			},
			Tools: []openai.Tool{
				{
					Type: openai.ToolTypeFunction,
					Function: &openai.FunctionDefinition{
						Name:   "display_cases",
						Strict: true,
						Parameters: &jsonschema.Definition{
							Type: jsonschema.Object,
							Properties: map[string]jsonschema.Definition{
								"PascalCase": {
									Type: jsonschema.String,
								},
								"CamelCase": {
									Type: jsonschema.String,
								},
								"KebabCase": {
									Type: jsonschema.String,
								},
								"SnakeCase": {
									Type: jsonschema.String,
								},
							},
							Required:             []string{"PascalCase", "CamelCase", "KebabCase", "SnakeCase"},
							AdditionalProperties: false,
						},
					},
				},
			},
			ToolChoice: openai.ToolChoice{
				Type: openai.ToolTypeFunction,
				Function: openai.ToolFunction{
					Name: "display_cases",
				},
			},
		},
	)
	checks.NoError(t, err, "CreateChatCompletion (use structured outputs response) returned error")
	var result = make(map[string]string)
	err = json.Unmarshal([]byte(resp.Choices[0].Message.ToolCalls[0].Function.Arguments), &result)
	checks.NoError(t, err, "CreateChatCompletion (use structured outputs response) unmarshal error")
	for _, key := range []string{"PascalCase", "CamelCase", "KebabCase", "SnakeCase"} {
		if _, ok := result[key]; !ok {
			t.Errorf("key:%s does not exist.", key)
		}
	}
}
