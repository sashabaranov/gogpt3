package openai_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/alejandrojnm/go-openai"
	"github.com/alejandrojnm/go-openai/internal/test/checks"
)

func TestCompletionsStreamWrongModel(t *testing.T) {
	config := openai.DefaultConfig("whatever")
	config.BaseURL = "http://localhost/v1"
	client := openai.NewClientWithConfig(config)

	_, err := client.CreateCompletionStream(
		context.Background(),
		openai.CompletionRequest{
			MaxTokens: 5,
			Model:     openai.GPT3Dot5Turbo,
		},
	)
	if !errors.Is(err, openai.ErrCompletionUnsupportedModel) {
		t.Fatalf("CreateCompletion should return ErrCompletionUnsupportedModel, but returned: %v", err)
	}
}

func TestCreateCompletionStream(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/completions", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")

		// Send test responses
		dataBytes := []byte{}
		dataBytes = append(dataBytes, []byte("event: message\n")...)
		//nolint:lll
		data := `{"id":"1","object":"completion","created":1598069254,"model":"text-davinci-002","choices":[{"text":"response1","finish_reason":"max_tokens"}]}`
		dataBytes = append(dataBytes, []byte("data: "+data+"\n\n")...)

		dataBytes = append(dataBytes, []byte("event: message\n")...)
		//nolint:lll
		data = `{"id":"2","object":"completion","created":1598069255,"model":"text-davinci-002","choices":[{"text":"response2","finish_reason":"max_tokens"}]}`
		dataBytes = append(dataBytes, []byte("data: "+data+"\n\n")...)

		dataBytes = append(dataBytes, []byte("event: done\n")...)
		dataBytes = append(dataBytes, []byte("data: [DONE]\n\n")...)

		_, err := w.Write(dataBytes)
		checks.NoError(t, err, "Write error")
	})

	stream, err := client.CreateCompletionStream(context.Background(), openai.CompletionRequest{
		Prompt:    "Ex falso quodlibet",
		Model:     "text-davinci-002",
		MaxTokens: 10,
		Stream:    true,
	})
	checks.NoError(t, err, "CreateCompletionStream returned error")
	defer stream.Close()

	expectedResponses := []openai.CompletionResponse{
		{
			ID:      "1",
			Object:  "completion",
			Created: 1598069254,
			Model:   "text-davinci-002",
			Choices: []openai.CompletionChoice{{Text: "response1", FinishReason: "max_tokens"}},
		},
		{
			ID:      "2",
			Object:  "completion",
			Created: 1598069255,
			Model:   "text-davinci-002",
			Choices: []openai.CompletionChoice{{Text: "response2", FinishReason: "max_tokens"}},
		},
	}

	for ix, expectedResponse := range expectedResponses {
		receivedResponse, streamErr := stream.Recv()
		if streamErr != nil {
			t.Errorf("stream.Recv() failed: %v", streamErr)
		}
		if !compareResponses(expectedResponse, receivedResponse) {
			t.Errorf("Stream response %v is %v, expected %v", ix, receivedResponse, expectedResponse)
		}
	}

	_, streamErr := stream.Recv()
	if !errors.Is(streamErr, io.EOF) {
		t.Errorf("stream.Recv() did not return EOF in the end: %v", streamErr)
	}

	_, streamErr = stream.Recv()
	if !errors.Is(streamErr, io.EOF) {
		t.Errorf("stream.Recv() did not return EOF when the stream is finished: %v", streamErr)
	}
}

func TestCreateCompletionStreamError(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/completions", func(w http.ResponseWriter, _ *http.Request) {
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

	stream, err := client.CreateCompletionStream(context.Background(), openai.CompletionRequest{
		MaxTokens: 5,
		Model:     openai.GPT3TextDavinci003,
		Prompt:    "Hello!",
		Stream:    true,
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

func TestCreateCompletionStreamRateLimitError(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/completions", func(w http.ResponseWriter, _ *http.Request) {
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

	var apiErr *openai.APIError
	_, err := client.CreateCompletionStream(context.Background(), openai.CompletionRequest{
		MaxTokens: 5,
		Model:     openai.GPT3Babbage002,
		Prompt:    "Hello!",
		Stream:    true,
	})
	if !errors.As(err, &apiErr) {
		t.Errorf("TestCreateCompletionStreamRateLimitError did not return APIError")
	}
	t.Logf("%+v\n", apiErr)
}

func TestCreateCompletionStreamTooManyEmptyStreamMessagesError(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/completions", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")

		// Send test responses
		dataBytes := []byte{}
		dataBytes = append(dataBytes, []byte("event: message\n")...)
		//nolint:lll
		data := `{"id":"1","object":"completion","created":1598069254,"model":"text-davinci-002","choices":[{"text":"response1","finish_reason":"max_tokens"}]}`
		dataBytes = append(dataBytes, []byte("data: "+data+"\n\n")...)

		// Totally 301 empty messages (300 is the limit)
		for i := 0; i < 299; i++ {
			dataBytes = append(dataBytes, '\n')
		}

		dataBytes = append(dataBytes, []byte("event: message\n")...)
		//nolint:lll
		data = `{"id":"2","object":"completion","created":1598069255,"model":"text-davinci-002","choices":[{"text":"response2","finish_reason":"max_tokens"}]}`
		dataBytes = append(dataBytes, []byte("data: "+data+"\n\n")...)

		dataBytes = append(dataBytes, []byte("event: done\n")...)
		dataBytes = append(dataBytes, []byte("data: [DONE]\n\n")...)

		_, err := w.Write(dataBytes)
		checks.NoError(t, err, "Write error")
	})

	stream, err := client.CreateCompletionStream(context.Background(), openai.CompletionRequest{
		Prompt:    "Ex falso quodlibet",
		Model:     "text-davinci-002",
		MaxTokens: 10,
		Stream:    true,
	})
	checks.NoError(t, err, "CreateCompletionStream returned error")
	defer stream.Close()

	_, _ = stream.Recv()
	_, streamErr := stream.Recv()
	if !errors.Is(streamErr, openai.ErrTooManyEmptyStreamMessages) {
		t.Errorf("TestCreateCompletionStreamTooManyEmptyStreamMessagesError did not return ErrTooManyEmptyStreamMessages")
	}
}

func TestCreateCompletionStreamUnexpectedTerminatedError(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/completions", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")

		// Send test responses
		dataBytes := []byte{}
		dataBytes = append(dataBytes, []byte("event: message\n")...)
		//nolint:lll
		data := `{"id":"1","object":"completion","created":1598069254,"model":"text-davinci-002","choices":[{"text":"response1","finish_reason":"max_tokens"}]}`
		dataBytes = append(dataBytes, []byte("data: "+data+"\n\n")...)

		// Stream is terminated without sending "done" message

		_, err := w.Write(dataBytes)
		checks.NoError(t, err, "Write error")
	})

	stream, err := client.CreateCompletionStream(context.Background(), openai.CompletionRequest{
		Prompt:    "Ex falso quodlibet",
		Model:     "text-davinci-002",
		MaxTokens: 10,
		Stream:    true,
	})
	checks.NoError(t, err, "CreateCompletionStream returned error")
	defer stream.Close()

	_, _ = stream.Recv()
	_, streamErr := stream.Recv()
	if !errors.Is(streamErr, io.EOF) {
		t.Errorf("TestCreateCompletionStreamUnexpectedTerminatedError did not return io.EOF")
	}
}

func TestCreateCompletionStreamBrokenJSONError(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/completions", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")

		// Send test responses
		dataBytes := []byte{}
		dataBytes = append(dataBytes, []byte("event: message\n")...)
		//nolint:lll
		data := `{"id":"1","object":"completion","created":1598069254,"model":"text-davinci-002","choices":[{"text":"response1","finish_reason":"max_tokens"}]}`
		dataBytes = append(dataBytes, []byte("data: "+data+"\n\n")...)

		// Send broken json
		dataBytes = append(dataBytes, []byte("event: message\n")...)
		data = `{"id":"2","object":"completion","created":1598069255,"model":`
		dataBytes = append(dataBytes, []byte("data: "+data+"\n\n")...)

		dataBytes = append(dataBytes, []byte("event: done\n")...)
		dataBytes = append(dataBytes, []byte("data: [DONE]\n\n")...)

		_, err := w.Write(dataBytes)
		checks.NoError(t, err, "Write error")
	})

	stream, err := client.CreateCompletionStream(context.Background(), openai.CompletionRequest{
		Prompt:    "Ex falso quodlibet",
		Model:     "text-davinci-002",
		MaxTokens: 10,
		Stream:    true,
	})
	checks.NoError(t, err, "CreateCompletionStream returned error")
	defer stream.Close()

	_, _ = stream.Recv()
	_, streamErr := stream.Recv()
	var syntaxError *json.SyntaxError
	if !errors.As(streamErr, &syntaxError) {
		t.Errorf("TestCreateCompletionStreamBrokenJSONError did not return json.SyntaxError")
	}
}

func TestCreateCompletionStreamReturnTimeoutError(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/completions", func(http.ResponseWriter, *http.Request) {
		time.Sleep(10 * time.Nanosecond)
	})
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Nanosecond)
	defer cancel()

	_, err := client.CreateCompletionStream(ctx, openai.CompletionRequest{
		Prompt:    "Ex falso quodlibet",
		Model:     "text-davinci-002",
		MaxTokens: 10,
		Stream:    true,
	})
	if err == nil {
		t.Fatal("Did not return error")
	}
	if !os.IsTimeout(err) {
		t.Fatal("Did not return timeout error")
	}
}

// Helper funcs.
func compareResponses(r1, r2 openai.CompletionResponse) bool {
	if r1.ID != r2.ID || r1.Object != r2.Object || r1.Created != r2.Created || r1.Model != r2.Model {
		return false
	}
	if len(r1.Choices) != len(r2.Choices) {
		return false
	}
	for i := range r1.Choices {
		if !compareResponseChoices(r1.Choices[i], r2.Choices[i]) {
			return false
		}
	}
	return true
}

func compareResponseChoices(c1, c2 openai.CompletionChoice) bool {
	if c1.Text != c2.Text || c1.FinishReason != c2.FinishReason {
		return false
	}
	return true
}
