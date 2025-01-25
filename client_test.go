package openai //nolint:testpackage // testing private field

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"testing"

	"github.com/alejandrojnm/go-openai/internal/test"
	"github.com/alejandrojnm/go-openai/internal/test/checks"
)

var errTestRequestBuilderFailed = errors.New("test request builder failed")

type failingRequestBuilder struct{}

func (*failingRequestBuilder) Build(_ context.Context, _, _ string, _ any, _ http.Header) (*http.Request, error) {
	return nil, errTestRequestBuilderFailed
}

func TestClient(t *testing.T) {
	const mockToken = "mock token"
	client := NewClient(mockToken)
	if client.config.authToken != mockToken {
		t.Errorf("Client does not contain proper token")
	}

	const mockOrg = "mock org"
	client = NewOrgClient(mockToken, mockOrg)
	if client.config.authToken != mockToken {
		t.Errorf("Client does not contain proper token")
	}
	if client.config.OrgID != mockOrg {
		t.Errorf("Client does not contain proper orgID")
	}
}

func TestDecodeResponse(t *testing.T) {
	stringInput := ""

	testCases := []struct {
		name     string
		value    interface{}
		expected interface{}
		body     io.Reader
		hasError bool
	}{
		{
			name:     "nil input",
			value:    nil,
			body:     bytes.NewReader([]byte("")),
			expected: nil,
		},
		{
			name:     "string input",
			value:    &stringInput,
			body:     bytes.NewReader([]byte("test")),
			expected: "test",
		},
		{
			name:  "map input",
			value: &map[string]interface{}{},
			body:  bytes.NewReader([]byte(`{"test": "test"}`)),
			expected: map[string]interface{}{
				"test": "test",
			},
		},
		{
			name:     "reader return error",
			value:    &stringInput,
			body:     &errorReader{err: errors.New("dummy")},
			hasError: true,
		},
		{
			name:  "audio text input",
			value: &audioTextResponse{},
			body:  bytes.NewReader([]byte("test")),
			expected: audioTextResponse{
				Text: "test",
			},
		},
	}

	assertEqual := func(t *testing.T, expected, actual interface{}) {
		t.Helper()
		if expected == actual {
			return
		}
		v := reflect.ValueOf(actual).Elem().Interface()
		if !reflect.DeepEqual(v, expected) {
			t.Fatalf("Unexpected value: %v, expected: %v", v, expected)
		}
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := decodeResponse(tc.body, tc.value)
			if tc.hasError {
				checks.HasError(t, err, "Unexpected nil error")
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			assertEqual(t, tc.expected, tc.value)
		})
	}
}

type errorReader struct {
	err error
}

func (e *errorReader) Read(_ []byte) (n int, err error) {
	return 0, e.err
}

func TestHandleErrorResp(t *testing.T) {
	// var errRes *ErrorResponse
	var errRes ErrorResponse
	var reqErr RequestError
	t.Log(errRes, errRes.Error)
	if errRes.Error != nil {
		reqErr.Err = errRes.Error
	}
	t.Log(fmt.Errorf("error, %w", &reqErr))
	t.Log(errRes.Error, "nil pointer check Pass")

	const mockToken = "mock token"
	client := NewClient(mockToken)

	testCases := []struct {
		name        string
		httpCode    int
		httpStatus  string
		contentType string
		body        io.Reader
		expected    string
	}{
		{
			name:        "401 Invalid Authentication",
			httpCode:    http.StatusUnauthorized,
			contentType: "application/json",
			body: bytes.NewReader([]byte(
				`{
					"error":{
						"message":"You didn't provide an API key. ....",
						"type":"invalid_request_error",
						"param":null,
						"code":null
					}
				}`,
			)),
			expected: "error, status code: 401, status: , message: You didn't provide an API key. ....",
		},
		{
			name:        "401 Azure Access Denied",
			httpCode:    http.StatusUnauthorized,
			contentType: "application/json",
			body: bytes.NewReader([]byte(
				`{
					"error":{
						"code":"AccessDenied",
						"message":"Access denied due to Virtual Network/Firewall rules."
					}
				}`,
			)),
			expected: "error, status code: 401, status: , message: Access denied due to Virtual Network/Firewall rules.",
		},
		{
			name:        "503 Model Overloaded",
			httpCode:    http.StatusServiceUnavailable,
			contentType: "application/json",
			body: bytes.NewReader([]byte(`
				{
					"error":{
						"message":"That model...",
						"type":"server_error",
						"param":null,
						"code":null
					}
				}`)),
			expected: "error, status code: 503, status: , message: That model...",
		},
		{
			name:        "503 no message (Unknown response)",
			httpCode:    http.StatusServiceUnavailable,
			contentType: "application/json",
			body: bytes.NewReader([]byte(`
				{
					"error":{}
				}`)),
			expected: `error, status code: 503, status: , message: , body: 
				{
					"error":{}
				}`,
		},
		{
			name:        "413 Request Entity Too Large",
			httpCode:    http.StatusRequestEntityTooLarge,
			contentType: "text/html",
			body: bytes.NewReader([]byte(`
	<html>
	<head><title>413 Request Entity Too Large</title></head>
	<body>
	<center><h1>413 Request Entity Too Large</h1></center>
	<hr><center>nginx</center>
	</body>
	</html>`)),
			expected: `error, status code: 413, status: , message: invalid character '<' looking for beginning of value, body: 
	<html>
	<head><title>413 Request Entity Too Large</title></head>
	<body>
	<center><h1>413 Request Entity Too Large</h1></center>
	<hr><center>nginx</center>
	</body>
	</html>`,
		},
		{
			name:        "errorReader",
			httpCode:    http.StatusRequestEntityTooLarge,
			contentType: "text/html",
			body:        &errorReader{err: errors.New("errorReader")},
			expected:    "error, reading response body: errorReader",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testCase := &http.Response{
				Header: map[string][]string{
					"Content-Type": {tc.contentType},
				},
			}
			testCase.StatusCode = tc.httpCode
			testCase.Body = io.NopCloser(tc.body)
			err := client.handleErrorResp(testCase)
			t.Log(err.Error())
			if err.Error() != tc.expected {
				t.Errorf("Unexpected error: %v , expected: %s", err, tc.expected)
				t.Fail()
			}
		})
	}
}

func TestClientReturnsRequestBuilderErrors(t *testing.T) {
	config := DefaultConfig(test.GetTestToken())
	client := NewClientWithConfig(config)
	client.requestBuilder = &failingRequestBuilder{}
	ctx := context.Background()

	type TestCase struct {
		Name     string
		TestFunc func() (any, error)
	}

	testCases := []TestCase{
		{"CreateCompletion", func() (any, error) {
			return client.CreateCompletion(ctx, CompletionRequest{Prompt: "testing"})
		}},
		{"CreateCompletionStream", func() (any, error) {
			return client.CreateCompletionStream(ctx, CompletionRequest{Prompt: ""})
		}},
		{"CreateChatCompletion", func() (any, error) {
			return client.CreateChatCompletion(ctx, ChatCompletionRequest{Model: GPT3Dot5Turbo})
		}},
		{"CreateChatCompletionStream", func() (any, error) {
			return client.CreateChatCompletionStream(ctx, ChatCompletionRequest{Model: GPT3Dot5Turbo})
		}},
		{"CreateFineTune", func() (any, error) {
			return client.CreateFineTune(ctx, FineTuneRequest{})
		}},
		{"ListFineTunes", func() (any, error) {
			return client.ListFineTunes(ctx)
		}},
		{"CancelFineTune", func() (any, error) {
			return client.CancelFineTune(ctx, "")
		}},
		{"GetFineTune", func() (any, error) {
			return client.GetFineTune(ctx, "")
		}},
		{"DeleteFineTune", func() (any, error) {
			return client.DeleteFineTune(ctx, "")
		}},
		{"ListFineTuneEvents", func() (any, error) {
			return client.ListFineTuneEvents(ctx, "")
		}},
		{"CreateFineTuningJob", func() (any, error) {
			return client.CreateFineTuningJob(ctx, FineTuningJobRequest{})
		}},
		{"CancelFineTuningJob", func() (any, error) {
			return client.CancelFineTuningJob(ctx, "")
		}},
		{"RetrieveFineTuningJob", func() (any, error) {
			return client.RetrieveFineTuningJob(ctx, "")
		}},
		{"ListFineTuningJobEvents", func() (any, error) {
			return client.ListFineTuningJobEvents(ctx, "")
		}},
		{"Moderations", func() (any, error) {
			return client.Moderations(ctx, ModerationRequest{})
		}},
		{"Edits", func() (any, error) {
			return client.Edits(ctx, EditsRequest{})
		}},
		{"CreateEmbeddings", func() (any, error) {
			return client.CreateEmbeddings(ctx, EmbeddingRequest{})
		}},
		{"CreateImage", func() (any, error) {
			return client.CreateImage(ctx, ImageRequest{})
		}},
		{"CreateFileBytes", func() (any, error) {
			return client.CreateFileBytes(ctx, FileBytesRequest{})
		}},
		{"DeleteFile", func() (any, error) {
			return nil, client.DeleteFile(ctx, "")
		}},
		{"GetFile", func() (any, error) {
			return client.GetFile(ctx, "")
		}},
		{"GetFileContent", func() (any, error) {
			return client.GetFileContent(ctx, "")
		}},
		{"ListFiles", func() (any, error) {
			return client.ListFiles(ctx)
		}},
		{"ListEngines", func() (any, error) {
			return client.ListEngines(ctx)
		}},
		{"GetEngine", func() (any, error) {
			return client.GetEngine(ctx, "")
		}},
		{"ListModels", func() (any, error) {
			return client.ListModels(ctx)
		}},
		{"GetModel", func() (any, error) {
			return client.GetModel(ctx, "text-davinci-003")
		}},
		{"DeleteFineTuneModel", func() (any, error) {
			return client.DeleteFineTuneModel(ctx, "")
		}},
		{"CreateAssistant", func() (any, error) {
			return client.CreateAssistant(ctx, AssistantRequest{})
		}},
		{"RetrieveAssistant", func() (any, error) {
			return client.RetrieveAssistant(ctx, "")
		}},
		{"ModifyAssistant", func() (any, error) {
			return client.ModifyAssistant(ctx, "", AssistantRequest{})
		}},
		{"DeleteAssistant", func() (any, error) {
			return client.DeleteAssistant(ctx, "")
		}},
		{"ListAssistants", func() (any, error) {
			return client.ListAssistants(ctx, nil, nil, nil, nil)
		}},
		{"CreateAssistantFile", func() (any, error) {
			return client.CreateAssistantFile(ctx, "", AssistantFileRequest{})
		}},
		{"ListAssistantFiles", func() (any, error) {
			return client.ListAssistantFiles(ctx, "", nil, nil, nil, nil)
		}},
		{"RetrieveAssistantFile", func() (any, error) {
			return client.RetrieveAssistantFile(ctx, "", "")
		}},
		{"DeleteAssistantFile", func() (any, error) {
			return nil, client.DeleteAssistantFile(ctx, "", "")
		}},
		{"CreateMessage", func() (any, error) {
			return client.CreateMessage(ctx, "", MessageRequest{})
		}},
		{"ListMessage", func() (any, error) {
			return client.ListMessage(ctx, "", nil, nil, nil, nil, nil)
		}},
		{"RetrieveMessage", func() (any, error) {
			return client.RetrieveMessage(ctx, "", "")
		}},
		{"ModifyMessage", func() (any, error) {
			return client.ModifyMessage(ctx, "", "", nil)
		}},
		{"DeleteMessage", func() (any, error) {
			return client.DeleteMessage(ctx, "", "")
		}},
		{"RetrieveMessageFile", func() (any, error) {
			return client.RetrieveMessageFile(ctx, "", "", "")
		}},
		{"ListMessageFiles", func() (any, error) {
			return client.ListMessageFiles(ctx, "", "")
		}},
		{"CreateThread", func() (any, error) {
			return client.CreateThread(ctx, ThreadRequest{})
		}},
		{"RetrieveThread", func() (any, error) {
			return client.RetrieveThread(ctx, "")
		}},
		{"ModifyThread", func() (any, error) {
			return client.ModifyThread(ctx, "", ModifyThreadRequest{})
		}},
		{"DeleteThread", func() (any, error) {
			return client.DeleteThread(ctx, "")
		}},
		{"CreateRun", func() (any, error) {
			return client.CreateRun(ctx, "", RunRequest{})
		}},
		{"RetrieveRun", func() (any, error) {
			return client.RetrieveRun(ctx, "", "")
		}},
		{"ModifyRun", func() (any, error) {
			return client.ModifyRun(ctx, "", "", RunModifyRequest{})
		}},
		{"ListRuns", func() (any, error) {
			return client.ListRuns(ctx, "", Pagination{})
		}},
		{"SubmitToolOutputs", func() (any, error) {
			return client.SubmitToolOutputs(ctx, "", "", SubmitToolOutputsRequest{})
		}},
		{"CancelRun", func() (any, error) {
			return client.CancelRun(ctx, "", "")
		}},
		{"CreateThreadAndRun", func() (any, error) {
			return client.CreateThreadAndRun(ctx, CreateThreadAndRunRequest{})
		}},
		{"RetrieveRunStep", func() (any, error) {
			return client.RetrieveRunStep(ctx, "", "", "")
		}},
		{"ListRunSteps", func() (any, error) {
			return client.ListRunSteps(ctx, "", "", Pagination{})
		}},
		{"CreateSpeech", func() (any, error) {
			return client.CreateSpeech(ctx, CreateSpeechRequest{Model: TTSModel1, Voice: VoiceAlloy})
		}},
		{"CreateBatch", func() (any, error) {
			return client.CreateBatch(ctx, CreateBatchRequest{})
		}},
		{"CreateBatchWithUploadFile", func() (any, error) {
			return client.CreateBatchWithUploadFile(ctx, CreateBatchWithUploadFileRequest{})
		}},
		{"RetrieveBatch", func() (any, error) {
			return client.RetrieveBatch(ctx, "")
		}},
		{"CancelBatch", func() (any, error) { return client.CancelBatch(ctx, "") }},
		{"ListBatch", func() (any, error) { return client.ListBatch(ctx, nil, nil) }},
	}

	for _, testCase := range testCases {
		_, err := testCase.TestFunc()
		if !errors.Is(err, errTestRequestBuilderFailed) {
			t.Fatalf("%s did not return error when request builder failed: %v", testCase.Name, err)
		}
	}
}

func TestClientReturnsRequestBuilderErrorsAddition(t *testing.T) {
	config := DefaultConfig(test.GetTestToken())
	client := NewClientWithConfig(config)
	client.requestBuilder = &failingRequestBuilder{}
	ctx := context.Background()
	_, err := client.CreateCompletion(ctx, CompletionRequest{Prompt: 1})
	if !errors.Is(err, ErrCompletionRequestPromptTypeNotSupported) {
		t.Fatalf("Did not return error when request builder failed: %v", err)
	}
	_, err = client.CreateCompletionStream(ctx, CompletionRequest{Prompt: 1})
	if !errors.Is(err, ErrCompletionRequestPromptTypeNotSupported) {
		t.Fatalf("Did not return error when request builder failed: %v", err)
	}
}

func TestClient_suffixWithAPIVersion(t *testing.T) {
	type fields struct {
		apiVersion string
	}
	type args struct {
		suffix string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      string
		wantPanic string
	}{
		{
			"",
			fields{apiVersion: "2023-05"},
			args{suffix: "/assistants"},
			"/assistants?api-version=2023-05",
			"",
		},
		{
			"",
			fields{apiVersion: "2023-05"},
			args{suffix: "/assistants?limit=5"},
			"/assistants?api-version=2023-05&limit=5",
			"",
		},
		{
			"",
			fields{apiVersion: "2023-05"},
			args{suffix: "123:assistants?limit=5"},
			"/assistants?api-version=2023-05&limit=5",
			"failed to parse url suffix",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				config: ClientConfig{APIVersion: tt.fields.apiVersion},
			}
			defer func() {
				if r := recover(); r != nil {
					// Check if the panic message matches the expected panic message
					if rStr, ok := r.(string); ok {
						if rStr != tt.wantPanic {
							t.Errorf("suffixWithAPIVersion() = %v, want %v", rStr, tt.wantPanic)
						}
					} else {
						// If the panic is not a string, log it
						t.Errorf("suffixWithAPIVersion() panicked with non-string value: %v", r)
					}
				}
			}()
			if got := c.suffixWithAPIVersion(tt.args.suffix); got != tt.want {
				t.Errorf("suffixWithAPIVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_baseURLWithAzureDeployment(t *testing.T) {
	type args struct {
		baseURL string
		suffix  string
		model   string
	}
	tests := []struct {
		name           string
		args           args
		wantNewBaseURL string
	}{
		{
			"",
			args{baseURL: "https://test.openai.azure.com/", suffix: assistantsSuffix, model: GPT4oMini},
			"https://test.openai.azure.com/openai",
		},
		{
			"",
			args{baseURL: "https://test.openai.azure.com/", suffix: chatCompletionsSuffix, model: GPT4oMini},
			"https://test.openai.azure.com/openai/deployments/gpt-4o-mini",
		},
		{
			"",
			args{baseURL: "https://test.openai.azure.com/", suffix: chatCompletionsSuffix, model: ""},
			"https://test.openai.azure.com/openai/deployments/UNKNOWN",
		},
	}
	client := NewClient("")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotNewBaseURL := client.baseURLWithAzureDeployment(
				tt.args.baseURL,
				tt.args.suffix,
				tt.args.model,
			); gotNewBaseURL != tt.wantNewBaseURL {
				t.Errorf("baseURLWithAzureDeployment() = %v, want %v", gotNewBaseURL, tt.wantNewBaseURL)
			}
		})
	}
}
