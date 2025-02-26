package openai

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
)

type FileRequest struct {
	// This parameter is unused.
	FileName string `json:"-"`

	FilePath string `json:"-"`
	Purpose  string `json:"purpose"`
}

// PurposeType represents the purpose of the file when uploading.
type PurposeType string

const (
	PurposeFineTune         PurposeType = "fine-tune"
	PurposeFineTuneResults  PurposeType = "fine-tune-results"
	PurposeAssistants       PurposeType = "assistants"
	PurposeAssistantsOutput PurposeType = "assistants_output"
	PurposeBatch            PurposeType = "batch"
)

// FileBytesRequest represents a file upload request.
type FileBytesRequest struct {
	// the name of the uploaded file in OpenAI
	Name string
	// the bytes of the file
	Bytes []byte
	// the purpose of the file
	Purpose PurposeType
}

// File struct represents an OpenAPI file.
type File struct {
	Bytes         int    `json:"bytes"`
	CreatedAt     int64  `json:"created_at"`
	ID            string `json:"id"`
	FileName      string `json:"filename"`
	Object        string `json:"object"`
	Status        string `json:"status"`
	Purpose       string `json:"purpose"`
	StatusDetails string `json:"status_details"`

	httpHeader
}

// FilesList is a list of files that belong to the user or organization.
type FilesList struct {
	Files []File `json:"data"`

	httpHeader
}

// CreateFileBytes uploads bytes directly to OpenAI without requiring a local file.
func (c *Client) CreateFileBytes(ctx context.Context, request FileBytesRequest) (file File, err error) {
	var b bytes.Buffer
	reader := bytes.NewReader(request.Bytes)
	builder := c.createFormBuilder(&b)

	err = builder.WriteField("purpose", string(request.Purpose))
	if err != nil {
		return
	}

	err = builder.CreateFormFileReader("file", reader, request.Name)
	if err != nil {
		return
	}

	err = builder.Close()
	if err != nil {
		return
	}

	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL("/files"),
		withBody(&b), withContentType(builder.FormDataContentType()))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &file)
	return
}

// CreateFile reads the contents at request.FilePath from disk, builds a jsonl
// file, and uploads it to OpenAI.
//
// FilePath must be a local file path.
func (c *Client) CreateFile(ctx context.Context, request FileRequest) (file File, err error) {
	var b bytes.Buffer
	builder := c.createFormBuilder(&b)

	err = builder.WriteField("purpose", request.Purpose)
	if err != nil {
		return
	}

	fileData, err := os.Open(request.FilePath)
	if err != nil {
		return
	}
	defer fileData.Close()

	err = builder.CreateFormFile("file", fileData)
	if err != nil {
		return
	}

	err = builder.Close()
	if err != nil {
		return
	}

	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL("/files"),
		withBody(&b), withContentType(builder.FormDataContentType()))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &file)
	return
}

// DeleteFile deletes an existing file.
func (c *Client) DeleteFile(ctx context.Context, fileID string) (err error) {
	req, err := c.newRequest(ctx, http.MethodDelete, c.fullURL("/files/"+fileID))
	if err != nil {
		return
	}

	err = c.sendRequest(req, nil)
	return
}

// ListFiles Lists the currently available files,
// and provides basic information about each file such as the file name and purpose.
func (c *Client) ListFiles(ctx context.Context) (files FilesList, err error) {
	req, err := c.newRequest(ctx, http.MethodGet, c.fullURL("/files"))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &files)
	return
}

// GetFile Retrieves a file instance, providing basic information about the file
// such as the file name and purpose.
func (c *Client) GetFile(ctx context.Context, fileID string) (file File, err error) {
	urlSuffix := fmt.Sprintf("/files/%s", fileID)
	req, err := c.newRequest(ctx, http.MethodGet, c.fullURL(urlSuffix))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &file)
	return
}

func (c *Client) GetFileContent(ctx context.Context, fileID string) (content RawResponse, err error) {
	urlSuffix := fmt.Sprintf("/files/%s/content", fileID)
	req, err := c.newRequest(ctx, http.MethodGet, c.fullURL(urlSuffix))
	if err != nil {
		return
	}

	return c.sendRequestRaw(req)
}
