package openai_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"testing"

	"github.com/alejandrojnm/go-openai"
	"github.com/alejandrojnm/go-openai/internal/test/checks"
)

func TestChatCompletionsStreamWrongModel(t *testing.T) {
	config := openai.DefaultConfig("whatever")
	config.BaseURL = "http://localhost/v1"
	client := openai.NewClientWithConfig(config)
	ctx := context.Background()

	req := openai.ChatCompletionRequest{
		MaxTokens: 5,
		Model:     "ada",
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "Hello!",
			},
		},
	}
	_, err := client.CreateChatCompletionStream(ctx, req)
	if !errors.Is(err, openai.ErrChatCompletionInvalidModel) {
		t.Fatalf("CreateChatCompletion should return ErrChatCompletionInvalidModel, but returned: %v", err)
	}
}

func TestCreateChatCompletionStream(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/chat/completions", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")

		// Send test responses
		dataBytes := []byte{}
		dataBytes = append(dataBytes, []byte("event: message\n")...)
		//nolint:lll
		data := `{"id":"1","object":"completion","created":1598069254,"model":"gpt-3.5-turbo","system_fingerprint": "fp_d9767fc5b9","choices":[{"index":0,"delta":{"content":"response1"},"finish_reason":"max_tokens"}]}`
		dataBytes = append(dataBytes, []byte("data: "+data+"\n\n")...)

		dataBytes = append(dataBytes, []byte("event: message\n")...)
		//nolint:lll
		data = `{"id":"2","object":"completion","created":1598069255,"model":"gpt-3.5-turbo","system_fingerprint": "fp_d9767fc5b9","choices":[{"index":0,"delta":{"content":"response2"},"finish_reason":"max_tokens"}]}`
		dataBytes = append(dataBytes, []byte("data: "+data+"\n\n")...)

		dataBytes = append(dataBytes, []byte("event: done\n")...)
		dataBytes = append(dataBytes, []byte("data: [DONE]\n\n")...)

		_, err := w.Write(dataBytes)
		checks.NoError(t, err, "Write error")
	})

	stream, err := client.CreateChatCompletionStream(context.Background(), openai.ChatCompletionRequest{
		MaxTokens: 5,
		Model:     openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "Hello!",
			},
		},
		Stream: true,
	})
	checks.NoError(t, err, "CreateCompletionStream returned error")
	defer stream.Close()

	expectedResponses := []openai.ChatCompletionStreamResponse{
		{
			ID:                "1",
			Object:            "completion",
			Created:           1598069254,
			Model:             openai.GPT3Dot5Turbo,
			SystemFingerprint: "fp_d9767fc5b9",
			Choices: []openai.ChatCompletionStreamChoice{
				{
					Delta: openai.ChatCompletionStreamChoiceDelta{
						Content: "response1",
					},
					FinishReason: "max_tokens",
				},
			},
		},
		{
			ID:                "2",
			Object:            "completion",
			Created:           1598069255,
			Model:             openai.GPT3Dot5Turbo,
			SystemFingerprint: "fp_d9767fc5b9",
			Choices: []openai.ChatCompletionStreamChoice{
				{
					Delta: openai.ChatCompletionStreamChoiceDelta{
						Content: "response2",
					},
					FinishReason: "max_tokens",
				},
			},
		},
	}

	for ix, expectedResponse := range expectedResponses {
		b, _ := json.Marshal(expectedResponse)
		t.Logf("%d: %s", ix, string(b))

		receivedResponse, streamErr := stream.Recv()
		checks.NoError(t, streamErr, "stream.Recv() failed")
		if !compareChatResponses(expectedResponse, receivedResponse) {
			t.Errorf("Stream response %v is %v, expected %v", ix, receivedResponse, expectedResponse)
		}
	}

	_, streamErr := stream.Recv()
	if !errors.Is(streamErr, io.EOF) {
		t.Errorf("stream.Recv() did not return EOF in the end: %v", streamErr)
	}

	_, streamErr = stream.Recv()

	checks.ErrorIs(t, streamErr, io.EOF, "stream.Recv() did not return EOF when the stream is finished")
	if !errors.Is(streamErr, io.EOF) {
		t.Errorf("stream.Recv() did not return EOF when the stream is finished: %v", streamErr)
	}
}

func TestCreateChatCompletionStreamError(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/chat/completions", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")

		// Send test responses
		dataBytes := []byte{}
		dataStr := []string{
			`{`,
			`"error": {`,
			`"message": "Incorrect API key provided: sk-***************************************",`,
			`"type": "invalid_request_error",`,
			`"param": null,`,
			`"code": "invalid_api_key"`,
			`}`,
			`}`,
		}
		for _, str := range dataStr {
			dataBytes = append(dataBytes, []byte(str+"\n")...)
		}

		_, err := w.Write(dataBytes)
		checks.NoError(t, err, "Write error")
	})

	stream, err := client.CreateChatCompletionStream(context.Background(), openai.ChatCompletionRequest{
		MaxTokens: 5,
		Model:     openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "Hello!",
			},
		},
		Stream: true,
	})
	checks.NoError(t, err, "CreateCompletionStream returned error")
	defer stream.Close()

	_, streamErr := stream.Recv()
	checks.HasError(t, streamErr, "stream.Recv() did not return error")

	var apiErr *openai.APIError
	if !errors.As(streamErr, &apiErr) {
		t.Errorf("stream.Recv() did not return APIError")
	}
	t.Logf("%+v\n", apiErr)
}

func TestCreateChatCompletionStreamWithHeaders(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/chat/completions", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set(xCustomHeader, xCustomHeaderValue)

		// Send test responses
		//nolint:lll
		dataBytes := []byte(`data: {"error":{"message":"The server had an error while processing your request. Sorry about that!", "type":"server_ error", "param":null,"code":null}}`)
		dataBytes = append(dataBytes, []byte("\n\ndata: [DONE]\n\n")...)

		_, err := w.Write(dataBytes)
		checks.NoError(t, err, "Write error")
	})

	stream, err := client.CreateChatCompletionStream(context.Background(), openai.ChatCompletionRequest{
		MaxTokens: 5,
		Model:     openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "Hello!",
			},
		},
		Stream: true,
	})
	checks.NoError(t, err, "CreateCompletionStream returned error")
	defer stream.Close()

	value := stream.Header().Get(xCustomHeader)
	if value != xCustomHeaderValue {
		t.Errorf("expected %s to be %s", xCustomHeaderValue, value)
	}
}

func TestCreateChatCompletionStreamWithRatelimitHeaders(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/chat/completions", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		for k, v := range rateLimitHeaders {
			switch val := v.(type) {
			case int:
				w.Header().Set(k, strconv.Itoa(val))
			default:
				w.Header().Set(k, fmt.Sprintf("%s", v))
			}
		}

		// Send test responses
		//nolint:lll
		dataBytes := []byte(`data: {"error":{"message":"The server had an error while processing your request. Sorry about that!", "type":"server_ error", "param":null,"code":null}}`)
		dataBytes = append(dataBytes, []byte("\n\ndata: [DONE]\n\n")...)

		_, err := w.Write(dataBytes)
		checks.NoError(t, err, "Write error")
	})

	stream, err := client.CreateChatCompletionStream(context.Background(), openai.ChatCompletionRequest{
		MaxTokens: 5,
		Model:     openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "Hello!",
			},
		},
		Stream: true,
	})
	checks.NoError(t, err, "CreateCompletionStream returned error")
	defer stream.Close()

	headers := stream.GetRateLimitHeaders()
	bs1, _ := json.Marshal(headers)
	bs2, _ := json.Marshal(rateLimitHeaders)
	if string(bs1) != string(bs2) {
		t.Errorf("expected rate limit header %s to be %s", bs2, bs1)
	}
}

func TestCreateChatCompletionStreamErrorWithDataPrefix(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/chat/completions", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")

		// Send test responses
		//nolint:lll
		dataBytes := []byte(`data: {"error":{"message":"The server had an error while processing your request. Sorry about that!", "type":"server_ error", "param":null,"code":null}}`)
		dataBytes = append(dataBytes, []byte("\n\ndata: [DONE]\n\n")...)

		_, err := w.Write(dataBytes)
		checks.NoError(t, err, "Write error")
	})

	stream, err := client.CreateChatCompletionStream(context.Background(), openai.ChatCompletionRequest{
		MaxTokens: 5,
		Model:     openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "Hello!",
			},
		},
		Stream: true,
	})
	checks.NoError(t, err, "CreateCompletionStream returned error")
	defer stream.Close()

	_, streamErr := stream.Recv()
	checks.HasError(t, streamErr, "stream.Recv() did not return error")

	var apiErr *openai.APIError
	if !errors.As(streamErr, &apiErr) {
		t.Errorf("stream.Recv() did not return APIError")
	}
	t.Logf("%+v\n", apiErr)
}

func TestCreateChatCompletionStreamRateLimitError(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/chat/completions", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(429)

		// Send test responses
		dataBytes := []byte(`{"error":{` +
			`"message": "You are sending requests too quickly.",` +
			`"type":"rate_limit_reached",` +
			`"param":null,` +
			`"code":"rate_limit_reached"}}`)

		_, err := w.Write(dataBytes)
		checks.NoError(t, err, "Write error")
	})
	_, err := client.CreateChatCompletionStream(context.Background(), openai.ChatCompletionRequest{
		MaxTokens: 5,
		Model:     openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "Hello!",
			},
		},
		Stream: true,
	})
	var apiErr *openai.APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("TestCreateChatCompletionStreamRateLimitError did not return APIError")
	}
	t.Logf("%+v\n", apiErr)
}

func TestCreateChatCompletionStreamWithRefusal(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/chat/completions", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")

		dataBytes := []byte{}

		//nolint:lll
		dataBytes = append(dataBytes, []byte(`data: {"id":"1","object":"chat.completion.chunk","created":1729585728,"model":"gpt-4o-mini-2024-07-18","system_fingerprint":"fp_d9767fc5b9","choices":[{"index":0,"delta":{"role":"assistant","content":"","refusal":null},"finish_reason":null}]}`)...)
		dataBytes = append(dataBytes, []byte("\n\n")...)

		//nolint:lll
		dataBytes = append(dataBytes, []byte(`data: {"id":"2","object":"chat.completion.chunk","created":1729585728,"model":"gpt-4o-mini-2024-07-18","system_fingerprint":"fp_d9767fc5b9","choices":[{"index":0,"delta":{"refusal":"Hello"},"finish_reason":null}]}`)...)
		dataBytes = append(dataBytes, []byte("\n\n")...)

		//nolint:lll
		dataBytes = append(dataBytes, []byte(`data: {"id":"3","object":"chat.completion.chunk","created":1729585728,"model":"gpt-4o-mini-2024-07-18","system_fingerprint":"fp_d9767fc5b9","choices":[{"index":0,"delta":{"refusal":" World"},"finish_reason":null}]}`)...)
		dataBytes = append(dataBytes, []byte("\n\n")...)

		//nolint:lll
		dataBytes = append(dataBytes, []byte(`data: {"id":"4","object":"chat.completion.chunk","created":1729585728,"model":"gpt-4o-mini-2024-07-18","system_fingerprint":"fp_d9767fc5b9","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}`)...)
		dataBytes = append(dataBytes, []byte("\n\n")...)

		dataBytes = append(dataBytes, []byte("data: [DONE]\n\n")...)

		_, err := w.Write(dataBytes)
		checks.NoError(t, err, "Write error")
	})

	stream, err := client.CreateChatCompletionStream(context.Background(), openai.ChatCompletionRequest{
		MaxTokens: 2000,
		Model:     openai.GPT4oMini20240718,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "Hello!",
			},
		},
		Stream: true,
	})
	checks.NoError(t, err, "CreateCompletionStream returned error")
	defer stream.Close()

	expectedResponses := []openai.ChatCompletionStreamResponse{
		{
			ID:                "1",
			Object:            "chat.completion.chunk",
			Created:           1729585728,
			Model:             openai.GPT4oMini20240718,
			SystemFingerprint: "fp_d9767fc5b9",
			Choices: []openai.ChatCompletionStreamChoice{
				{
					Index: 0,
					Delta: openai.ChatCompletionStreamChoiceDelta{},
				},
			},
		},
		{
			ID:                "2",
			Object:            "chat.completion.chunk",
			Created:           1729585728,
			Model:             openai.GPT4oMini20240718,
			SystemFingerprint: "fp_d9767fc5b9",
			Choices: []openai.ChatCompletionStreamChoice{
				{
					Index: 0,
					Delta: openai.ChatCompletionStreamChoiceDelta{
						Refusal: "Hello",
					},
				},
			},
		},
		{
			ID:                "3",
			Object:            "chat.completion.chunk",
			Created:           1729585728,
			Model:             openai.GPT4oMini20240718,
			SystemFingerprint: "fp_d9767fc5b9",
			Choices: []openai.ChatCompletionStreamChoice{
				{
					Index: 0,
					Delta: openai.ChatCompletionStreamChoiceDelta{
						Refusal: " World",
					},
				},
			},
		},
		{
			ID:                "4",
			Object:            "chat.completion.chunk",
			Created:           1729585728,
			Model:             openai.GPT4oMini20240718,
			SystemFingerprint: "fp_d9767fc5b9",
			Choices: []openai.ChatCompletionStreamChoice{
				{
					Index:        0,
					FinishReason: "stop",
				},
			},
		},
	}

	for ix, expectedResponse := range expectedResponses {
		b, _ := json.Marshal(expectedResponse)
		t.Logf("%d: %s", ix, string(b))

		receivedResponse, streamErr := stream.Recv()
		checks.NoError(t, streamErr, "stream.Recv() failed")
		if !compareChatResponses(expectedResponse, receivedResponse) {
			t.Errorf("Stream response %v is %v, expected %v", ix, receivedResponse, expectedResponse)
		}
	}

	_, streamErr := stream.Recv()
	if !errors.Is(streamErr, io.EOF) {
		t.Errorf("stream.Recv() did not return EOF in the end: %v", streamErr)
	}
}

func TestCreateChatCompletionStreamWithLogprobs(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/chat/completions", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")

		// Send test responses
		dataBytes := []byte{}

		//nolint:lll
		dataBytes = append(dataBytes, []byte(`data: {"id":"1","object":"chat.completion.chunk","created":1729585728,"model":"gpt-4o-mini-2024-07-18","system_fingerprint":"fp_d9767fc5b9","choices":[{"index":0,"delta":{"role":"assistant","content":"","refusal":null},"logprobs":{"content":[],"refusal":null},"finish_reason":null}]}`)...)
		dataBytes = append(dataBytes, []byte("\n\n")...)

		//nolint:lll
		dataBytes = append(dataBytes, []byte(`data: {"id":"2","object":"chat.completion.chunk","created":1729585728,"model":"gpt-4o-mini-2024-07-18","system_fingerprint":"fp_d9767fc5b9","choices":[{"index":0,"delta":{"content":"Hello"},"logprobs":{"content":[{"token":"Hello","logprob":-0.000020458236,"bytes":[72,101,108,108,111],"top_logprobs":[]}],"refusal":null},"finish_reason":null}]}`)...)
		dataBytes = append(dataBytes, []byte("\n\n")...)

		//nolint:lll
		dataBytes = append(dataBytes, []byte(`data: {"id":"3","object":"chat.completion.chunk","created":1729585728,"model":"gpt-4o-mini-2024-07-18","system_fingerprint":"fp_d9767fc5b9","choices":[{"index":0,"delta":{"content":" World"},"logprobs":{"content":[{"token":" World","logprob":-0.00055303273,"bytes":[32,87,111,114,108,100],"top_logprobs":[]}],"refusal":null},"finish_reason":null}]}`)...)
		dataBytes = append(dataBytes, []byte("\n\n")...)

		//nolint:lll
		dataBytes = append(dataBytes, []byte(`data: {"id":"4","object":"chat.completion.chunk","created":1729585728,"model":"gpt-4o-mini-2024-07-18","system_fingerprint":"fp_d9767fc5b9","choices":[{"index":0,"delta":{},"logprobs":null,"finish_reason":"stop"}]}`)...)
		dataBytes = append(dataBytes, []byte("\n\n")...)

		dataBytes = append(dataBytes, []byte("data: [DONE]\n\n")...)

		_, err := w.Write(dataBytes)
		checks.NoError(t, err, "Write error")
	})

	stream, err := client.CreateChatCompletionStream(context.Background(), openai.ChatCompletionRequest{
		MaxTokens: 2000,
		Model:     openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "Hello!",
			},
		},
		Stream: true,
	})
	checks.NoError(t, err, "CreateCompletionStream returned error")
	defer stream.Close()

	expectedResponses := []openai.ChatCompletionStreamResponse{
		{
			ID:                "1",
			Object:            "chat.completion.chunk",
			Created:           1729585728,
			Model:             openai.GPT4oMini20240718,
			SystemFingerprint: "fp_d9767fc5b9",
			Choices: []openai.ChatCompletionStreamChoice{
				{
					Index: 0,
					Delta: openai.ChatCompletionStreamChoiceDelta{},
					Logprobs: &openai.ChatCompletionStreamChoiceLogprobs{
						Content: []openai.ChatCompletionTokenLogprob{},
					},
				},
			},
		},
		{
			ID:                "2",
			Object:            "chat.completion.chunk",
			Created:           1729585728,
			Model:             openai.GPT4oMini20240718,
			SystemFingerprint: "fp_d9767fc5b9",
			Choices: []openai.ChatCompletionStreamChoice{
				{
					Index: 0,
					Delta: openai.ChatCompletionStreamChoiceDelta{
						Content: "Hello",
					},
					Logprobs: &openai.ChatCompletionStreamChoiceLogprobs{
						Content: []openai.ChatCompletionTokenLogprob{
							{
								Token:       "Hello",
								Logprob:     -0.000020458236,
								Bytes:       []int64{72, 101, 108, 108, 111},
								TopLogprobs: []openai.ChatCompletionTokenLogprobTopLogprob{},
							},
						},
					},
				},
			},
		},
		{
			ID:                "3",
			Object:            "chat.completion.chunk",
			Created:           1729585728,
			Model:             openai.GPT4oMini20240718,
			SystemFingerprint: "fp_d9767fc5b9",
			Choices: []openai.ChatCompletionStreamChoice{
				{
					Index: 0,
					Delta: openai.ChatCompletionStreamChoiceDelta{
						Content: " World",
					},
					Logprobs: &openai.ChatCompletionStreamChoiceLogprobs{
						Content: []openai.ChatCompletionTokenLogprob{
							{
								Token:       " World",
								Logprob:     -0.00055303273,
								Bytes:       []int64{32, 87, 111, 114, 108, 100},
								TopLogprobs: []openai.ChatCompletionTokenLogprobTopLogprob{},
							},
						},
					},
				},
			},
		},
		{
			ID:                "4",
			Object:            "chat.completion.chunk",
			Created:           1729585728,
			Model:             openai.GPT4oMini20240718,
			SystemFingerprint: "fp_d9767fc5b9",
			Choices: []openai.ChatCompletionStreamChoice{
				{
					Index:        0,
					Delta:        openai.ChatCompletionStreamChoiceDelta{},
					FinishReason: "stop",
				},
			},
		},
	}

	for ix, expectedResponse := range expectedResponses {
		b, _ := json.Marshal(expectedResponse)
		t.Logf("%d: %s", ix, string(b))

		receivedResponse, streamErr := stream.Recv()
		checks.NoError(t, streamErr, "stream.Recv() failed")
		if !compareChatResponses(expectedResponse, receivedResponse) {
			t.Errorf("Stream response %v is %v, expected %v", ix, receivedResponse, expectedResponse)
		}
	}

	_, streamErr := stream.Recv()
	if !errors.Is(streamErr, io.EOF) {
		t.Errorf("stream.Recv() did not return EOF in the end: %v", streamErr)
	}
}

func TestAzureCreateChatCompletionStreamRateLimitError(t *testing.T) {
	wantCode := "429"
	wantMessage := "Requests to the Creates a completion for the chat message Operation under Azure OpenAI API " +
		"version 2023-03-15-preview have exceeded token rate limit of your current OpenAI S0 pricing tier. " +
		"Please retry after 20 seconds. " +
		"Please go here: https://aka.ms/oai/quotaincrease if you would like to further increase the default rate limit."

	client, server, teardown := setupAzureTestServer()
	defer teardown()
	server.RegisterHandler("/openai/deployments/gpt-35-turbo/chat/completions",
		func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			// Send test responses
			dataBytes := []byte(`{"error": { "code": "` + wantCode + `", "message": "` + wantMessage + `"}}`)
			_, err := w.Write(dataBytes)

			checks.NoError(t, err, "Write error")
		})

	apiErr := &openai.APIError{}
	_, err := client.CreateChatCompletionStream(context.Background(), openai.ChatCompletionRequest{
		MaxTokens: 5,
		Model:     openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "Hello!",
			},
		},
		Stream: true,
	})
	if !errors.As(err, &apiErr) {
		t.Errorf("Did not return APIError: %+v\n", apiErr)
		return
	}
	if apiErr.HTTPStatusCode != http.StatusTooManyRequests {
		t.Errorf("Did not return HTTPStatusCode got = %d, want = %d\n", apiErr.HTTPStatusCode, http.StatusTooManyRequests)
		return
	}
	code, ok := apiErr.Code.(string)
	if !ok || code != wantCode {
		t.Errorf("Did not return Code. got = %v, want = %s\n", apiErr.Code, wantCode)
		return
	}
	if apiErr.Message != wantMessage {
		t.Errorf("Did not return Message. got = %s, want = %s\n", apiErr.Message, wantMessage)
		return
	}
}

func TestCreateChatCompletionStreamStreamOptions(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()

	server.RegisterHandler("/v1/chat/completions", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")

		// Send test responses
		var dataBytes []byte
		//nolint:lll
		data := `{"id":"1","object":"completion","created":1598069254,"model":"gpt-3.5-turbo","system_fingerprint": "fp_d9767fc5b9","choices":[{"index":0,"delta":{"content":"response1"},"finish_reason":"max_tokens"}],"usage":null}`
		dataBytes = append(dataBytes, []byte("data: "+data+"\n\n")...)

		//nolint:lll
		data = `{"id":"2","object":"completion","created":1598069255,"model":"gpt-3.5-turbo","system_fingerprint": "fp_d9767fc5b9","choices":[{"index":0,"delta":{"content":"response2"},"finish_reason":"max_tokens"}],"usage":null}`
		dataBytes = append(dataBytes, []byte("data: "+data+"\n\n")...)

		//nolint:lll
		data = `{"id":"3","object":"completion","created":1598069256,"model":"gpt-3.5-turbo","system_fingerprint": "fp_d9767fc5b9","choices":[],"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`
		dataBytes = append(dataBytes, []byte("data: "+data+"\n\n")...)

		dataBytes = append(dataBytes, []byte("data: [DONE]\n\n")...)

		_, err := w.Write(dataBytes)
		checks.NoError(t, err, "Write error")
	})

	stream, err := client.CreateChatCompletionStream(context.Background(), openai.ChatCompletionRequest{
		MaxTokens: 5,
		Model:     openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "Hello!",
			},
		},
		Stream: true,
		StreamOptions: &openai.StreamOptions{
			IncludeUsage: true,
		},
	})
	checks.NoError(t, err, "CreateCompletionStream returned error")
	defer stream.Close()

	expectedResponses := []openai.ChatCompletionStreamResponse{
		{
			ID:                "1",
			Object:            "completion",
			Created:           1598069254,
			Model:             openai.GPT3Dot5Turbo,
			SystemFingerprint: "fp_d9767fc5b9",
			Choices: []openai.ChatCompletionStreamChoice{
				{
					Delta: openai.ChatCompletionStreamChoiceDelta{
						Content: "response1",
					},
					FinishReason: "max_tokens",
				},
			},
		},
		{
			ID:                "2",
			Object:            "completion",
			Created:           1598069255,
			Model:             openai.GPT3Dot5Turbo,
			SystemFingerprint: "fp_d9767fc5b9",
			Choices: []openai.ChatCompletionStreamChoice{
				{
					Delta: openai.ChatCompletionStreamChoiceDelta{
						Content: "response2",
					},
					FinishReason: "max_tokens",
				},
			},
		},
		{
			ID:                "3",
			Object:            "completion",
			Created:           1598069256,
			Model:             openai.GPT3Dot5Turbo,
			SystemFingerprint: "fp_d9767fc5b9",
			Choices:           []openai.ChatCompletionStreamChoice{},
			Usage: &openai.Usage{
				PromptTokens:     1,
				CompletionTokens: 1,
				TotalTokens:      2,
			},
		},
	}

	for ix, expectedResponse := range expectedResponses {
		b, _ := json.Marshal(expectedResponse)
		t.Logf("%d: %s", ix, string(b))

		receivedResponse, streamErr := stream.Recv()
		checks.NoError(t, streamErr, "stream.Recv() failed")
		if !compareChatResponses(expectedResponse, receivedResponse) {
			t.Errorf("Stream response %v is %v, expected %v", ix, receivedResponse, expectedResponse)
		}
	}

	_, streamErr := stream.Recv()
	if !errors.Is(streamErr, io.EOF) {
		t.Errorf("stream.Recv() did not return EOF in the end: %v", streamErr)
	}

	_, streamErr = stream.Recv()

	checks.ErrorIs(t, streamErr, io.EOF, "stream.Recv() did not return EOF when the stream is finished")
	if !errors.Is(streamErr, io.EOF) {
		t.Errorf("stream.Recv() did not return EOF when the stream is finished: %v", streamErr)
	}
}

// Helper funcs.
func compareChatResponses(r1, r2 openai.ChatCompletionStreamResponse) bool {
	if r1.ID != r2.ID || r1.Object != r2.Object || r1.Created != r2.Created || r1.Model != r2.Model {
		return false
	}
	if len(r1.Choices) != len(r2.Choices) {
		return false
	}
	for i := range r1.Choices {
		if !compareChatStreamResponseChoices(r1.Choices[i], r2.Choices[i]) {
			return false
		}
	}
	if r1.Usage != nil || r2.Usage != nil {
		if r1.Usage == nil || r2.Usage == nil {
			return false
		}
		if r1.Usage.PromptTokens != r2.Usage.PromptTokens || r1.Usage.CompletionTokens != r2.Usage.CompletionTokens ||
			r1.Usage.TotalTokens != r2.Usage.TotalTokens {
			return false
		}
	}
	return true
}

func compareChatStreamResponseChoices(c1, c2 openai.ChatCompletionStreamChoice) bool {
	if c1.Index != c2.Index {
		return false
	}
	if c1.Delta.Content != c2.Delta.Content {
		return false
	}
	if c1.FinishReason != c2.FinishReason {
		return false
	}
	return true
}
