package openai_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test/checks"
)

func TestFileBytesUpload(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/files", handleCreateFile)
	req := openai.FileBytesRequest{
		Name:    "foo",
		Bytes:   []byte("foo"),
		Purpose: openai.PurposeFineTune,
	}
	_, err := client.CreateFileBytes(context.Background(), req)
	checks.NoError(t, err, "CreateFile error")
}

func TestFileUpload(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/files", handleCreateFile)
	req := openai.FileRequest{
		FileName: "test.go",
		FilePath: "client.go",
		Purpose:  "fine-tune",
	}
	_, err := client.CreateFile(context.Background(), req)
	checks.NoError(t, err, "CreateFile error")
}

// handleCreateFile Handles the images endpoint by the test server.
func handleCreateFile(w http.ResponseWriter, r *http.Request) {
	var err error
	var resBytes []byte

	// edits only accepts POST requests
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
	err = r.ParseMultipartForm(1024 * 1024 * 1024)
	if err != nil {
		http.Error(w, "file is more than 1GB", http.StatusInternalServerError)
		return
	}

	values := r.Form
	var purpose string
	for key, value := range values {
		if key == "purpose" {
			purpose = value[0]
		}
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		return
	}
	defer file.Close()

	fileReq := openai.File{
		Bytes:     int(header.Size),
		ID:        strconv.Itoa(int(time.Now().Unix())),
		FileName:  header.Filename,
		Purpose:   purpose,
		CreatedAt: time.Now().Unix(),
		Object:    "test-objecct",
	}

	resBytes, _ = json.Marshal(fileReq)
	fmt.Fprint(w, string(resBytes))
}

func TestDeleteFile(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/files/deadbeef", func(http.ResponseWriter, *http.Request) {})
	err := client.DeleteFile(context.Background(), "deadbeef")
	checks.NoError(t, err, "DeleteFile error")
}

func TestListFile(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/files", func(w http.ResponseWriter, _ *http.Request) {
		resBytes, _ := json.Marshal(openai.FilesList{})
		fmt.Fprintln(w, string(resBytes))
	})
	_, err := client.ListFiles(context.Background())
	checks.NoError(t, err, "ListFiles error")
}

func TestListFileWithOpts(t *testing.T) {
	limit := 10
	order := "desc"
	afterID := "deadbeef"

	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/files", func(w http.ResponseWriter, _ *http.Request) {
		resBytes, _ := json.Marshal(openai.FilesList{})
		fmt.Fprintln(w, string(resBytes))
	})
	_, err := client.ListFilesWithOpts(context.Background(), &limit, &order, &afterID)
	checks.NoError(t, err, "ListFiles error")
}

func TestGetFile(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/files/deadbeef", func(w http.ResponseWriter, _ *http.Request) {
		resBytes, _ := json.Marshal(openai.File{})
		fmt.Fprintln(w, string(resBytes))
	})
	_, err := client.GetFile(context.Background(), "deadbeef")
	checks.NoError(t, err, "GetFile error")
}

func TestGetFileContent(t *testing.T) {
	wantRespJsonl := `{"prompt": "foo", "completion": "foo"}
{"prompt": "bar", "completion": "bar"}
{"prompt": "baz", "completion": "baz"}
`
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/files/deadbeef/content", func(w http.ResponseWriter, r *http.Request) {
		// edits only accepts GET requests
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		fmt.Fprint(w, wantRespJsonl)
	})

	content, err := client.GetFileContent(context.Background(), "deadbeef")
	checks.NoError(t, err, "GetFileContent error")
	defer content.Close()

	actual, _ := io.ReadAll(content)
	if string(actual) != wantRespJsonl {
		t.Errorf("Expected %s, got %s", wantRespJsonl, string(actual))
	}
}

func TestGetFileContentReturnError(t *testing.T) {
	wantMessage := "To help mitigate abuse, downloading of fine-tune training files is disabled for free accounts."
	wantType := "invalid_request_error"
	wantErrorResp := `{
  "error": {
    "message": "` + wantMessage + `",
    "type": "` + wantType + `",
    "param": null,
    "code": null
  }
}`
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/files/deadbeef/content", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, wantErrorResp)
	})

	_, err := client.GetFileContent(context.Background(), "deadbeef")
	if err == nil {
		t.Fatal("Did not return error")
	}

	apiErr := &openai.APIError{}
	if !errors.As(err, &apiErr) {
		t.Fatalf("Did not return APIError: %+v\n", apiErr)
	}
	if apiErr.Message != wantMessage {
		t.Fatalf("Expected %s Message, got = %s\n", wantMessage, apiErr.Message)
		return
	}
	if apiErr.Type != wantType {
		t.Fatalf("Expected %s Type, got = %s\n", wantType, apiErr.Type)
		return
	}
}

func TestGetFileContentReturnTimeoutError(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/files/deadbeef/content", func(http.ResponseWriter, *http.Request) {
		time.Sleep(10 * time.Nanosecond)
	})
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Nanosecond)
	defer cancel()

	_, err := client.GetFileContent(ctx, "deadbeef")
	if err == nil {
		t.Fatal("Did not return error")
	}
	if !os.IsTimeout(err) {
		t.Fatal("Did not return timeout error")
	}
}
