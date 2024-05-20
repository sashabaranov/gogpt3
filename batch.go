package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

const batchesSuffix = "/batches"

type BatchEndpoint string

const BatchEndpointChatCompletions BatchEndpoint = "/v1/chat/completions"
const BatchEndpointCompletions BatchEndpoint = "/v1/completions"
const BatchEndpointEmbeddings BatchEndpoint = "/v1/embeddings"

type BatchRequestFile interface {
	MarshalBatchFile() []byte
}

type BatchRequestFiles []BatchRequestFile

func (r BatchRequestFiles) Marshal() []byte {
	buff := bytes.Buffer{}
	for i, request := range r {
		marshal := request.MarshalBatchFile()
		if i != 0 {
			buff.Write([]byte("\n"))
		}
		buff.Write(marshal)
	}
	return buff.Bytes()
}

type BatchChatCompletionRequest struct {
	CustomID string                `json:"custom_id"`
	Body     ChatCompletionRequest `json:"body"`
	Method   string                `json:"method"`
	URL      BatchEndpoint         `json:"url"`
}

func (r BatchChatCompletionRequest) MarshalBatchFile() []byte {
	marshal, _ := json.Marshal(r)
	return marshal
}

type BatchCompletionRequest struct {
	CustomID string            `json:"custom_id"`
	Body     CompletionRequest `json:"body"`
	Method   string            `json:"method"`
	URL      BatchEndpoint     `json:"url"`
}

func (r BatchCompletionRequest) MarshalBatchFile() []byte {
	marshal, _ := json.Marshal(r)
	return marshal
}

type BatchEmbeddingRequest struct {
	CustomID string           `json:"custom_id"`
	Body     EmbeddingRequest `json:"body"`
	Method   string           `json:"method"`
	URL      BatchEndpoint    `json:"url"`
}

func (r BatchEmbeddingRequest) MarshalBatchFile() []byte {
	marshal, _ := json.Marshal(r)
	return marshal
}

type Batch struct {
	ID       string `json:"id"`
	Object   string `json:"object"`
	Endpoint string `json:"endpoint"`
	Errors   *struct {
		Object string `json:"object,omitempty"`
		Data   struct {
			Code    string  `json:"code,omitempty"`
			Message string  `json:"message,omitempty"`
			Param   *string `json:"param,omitempty"`
			Line    *int    `json:"line,omitempty"`
		} `json:"data"`
	} `json:"errors"`
	InputFileID      string             `json:"input_file_id"`
	CompletionWindow string             `json:"completion_window"`
	Status           string             `json:"status"`
	OutputFileID     *string            `json:"output_file_id"`
	ErrorFileID      *string            `json:"error_file_id"`
	CreatedAt        int                `json:"created_at"`
	InProgressAt     *int               `json:"in_progress_at"`
	ExpiresAt        *int               `json:"expires_at"`
	FinalizingAt     *int               `json:"finalizing_at"`
	CompletedAt      *int               `json:"completed_at"`
	FailedAt         *int               `json:"failed_at"`
	ExpiredAt        *int               `json:"expired_at"`
	CancellingAt     *int               `json:"cancelling_at"`
	CancelledAt      *int               `json:"cancelled_at"`
	RequestCounts    BatchRequestCounts `json:"request_counts"`
	Metadata         map[string]any     `json:"metadata"`
}

type BatchRequestCounts struct {
	Total     int `json:"total"`
	Completed int `json:"completed"`
	Failed    int `json:"failed"`
}

type CreateBatchRequest struct {
	InputFileID      string         `json:"input_file_id"`
	Endpoint         BatchEndpoint  `json:"endpoint"`
	CompletionWindow string         `json:"completion_window"`
	Metadata         map[string]any `json:"metadata"`
}

type BatchResponse struct {
	httpHeader
	Batch
}

var ErrUploadBatchFileFailed = errors.New("upload batch file failed")

// CreateBatch — API call to Create batch.
func (c *Client) CreateBatch(
	ctx context.Context,
	request CreateBatchRequest,
) (response BatchResponse, err error) {
	if request.CompletionWindow == "" {
		request.CompletionWindow = "24h"
	}

	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL(batchesSuffix), withBody(request))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

type CreateBatchWithUploadFileRequest struct {
	FileName         string            `json:"file_name"`
	Endpoint         BatchEndpoint     `json:"endpoint"`
	CompletionWindow string            `json:"completion_window"`
	Metadata         map[string]any    `json:"metadata"`
	Requests         BatchRequestFiles `json:"requests"`
}

func (r *CreateBatchWithUploadFileRequest) AddChatCompletion(customerID string, body ChatCompletionRequest) {
	r.Requests = append(r.Requests, BatchChatCompletionRequest{
		CustomID: customerID,
		Body:     body,
		Method:   "POST",
		URL:      BatchEndpointChatCompletions,
	})
}

func (r *CreateBatchWithUploadFileRequest) AddCompletion(customerID string, body CompletionRequest) {
	r.Requests = append(r.Requests, BatchCompletionRequest{
		CustomID: customerID,
		Body:     body,
		Method:   "POST",
		URL:      BatchEndpointCompletions,
	})
}

func (r *CreateBatchWithUploadFileRequest) AddEmbedding(customerID string, body EmbeddingRequest) {
	r.Requests = append(r.Requests, BatchEmbeddingRequest{
		CustomID: customerID,
		Body:     body,
		Method:   "POST",
		URL:      BatchEndpointEmbeddings,
	})
}

// CreateBatchWithUploadFile — API call to Create batch with upload file.
func (c *Client) CreateBatchWithUploadFile(
	ctx context.Context,
	request CreateBatchWithUploadFileRequest,
) (response BatchResponse, err error) {
	if request.FileName == "" {
		request.FileName = "@batchinput.jsonl"
	}
	var file File
	file, err = c.CreateFileBytes(ctx, FileBytesRequest{
		Name:    request.FileName,
		Bytes:   request.Requests.Marshal(),
		Purpose: PurposeBatch,
	})
	if err != nil {
		err = errors.Join(ErrUploadBatchFileFailed, err)
		return
	}
	response, err = c.CreateBatch(ctx, CreateBatchRequest{
		InputFileID:      file.ID,
		Endpoint:         request.Endpoint,
		CompletionWindow: request.CompletionWindow,
		Metadata:         request.Metadata,
	})
	return
}

// RetrieveBatch — API call to Retrieve batch.
func (c *Client) RetrieveBatch(
	ctx context.Context,
	batchID string,
) (response BatchResponse, err error) {
	urlSuffix := fmt.Sprintf("%s/%s", batchesSuffix, batchID)
	req, err := c.newRequest(ctx, http.MethodGet, c.fullURL(urlSuffix))
	if err != nil {
		return
	}
	err = c.sendRequest(req, &response)
	return
}

// CancelBatch — API call to Cancel batch.
func (c *Client) CancelBatch(
	ctx context.Context,
	batchID string,
) (response BatchResponse, err error) {
	urlSuffix := fmt.Sprintf("%s/%s/cancel", batchesSuffix, batchID)
	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL(urlSuffix))
	if err != nil {
		return
	}
	err = c.sendRequest(req, &response)
	return
}

type ListBatchResponse struct {
	httpHeader
	Object  string  `json:"object"`
	Data    []Batch `json:"data"`
	FirstID string  `json:"first_id"`
	LastID  string  `json:"last_id"`
	HasMore bool    `json:"has_more"`
}

// ListBatch API call to List batch.
func (c *Client) ListBatch(ctx context.Context, after *string, limit *int) (response ListBatchResponse, err error) {
	urlValues := url.Values{}
	if limit != nil {
		urlValues.Add("limit", fmt.Sprintf("%d", *limit))
	}
	if after != nil {
		urlValues.Add("after", *after)
	}
	encodedValues := ""
	if len(urlValues) > 0 {
		encodedValues = "?" + urlValues.Encode()
	}

	urlSuffix := fmt.Sprintf("%s%s", batchesSuffix, encodedValues)
	req, err := c.newRequest(ctx, http.MethodGet, c.fullURL(urlSuffix))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}
