package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const batchesSuffix = "/batches"

type BatchEndpoint string

const (
	BatchEndpointChatCompletions BatchEndpoint = "/v1/chat/completions"
	BatchEndpointCompletions     BatchEndpoint = "/v1/completions"
	BatchEndpointEmbeddings      BatchEndpoint = "/v1/embeddings"
)

type BatchStatus string

const (
	BatchStatusValidating BatchStatus = "validating"
	BatchStatusFailed     BatchStatus = "failed"
	BatchStatusInProgress BatchStatus = "in_progress"
	BatchStatusFinalizing BatchStatus = "finalizing"
	BatchStatusCompleted  BatchStatus = "completed"
	BatchStatusExpired    BatchStatus = "expired"
	BatchStatusCancelling BatchStatus = "cancelling"
	BatchStatusCancelled  BatchStatus = "cancelled"
)

type BatchLineItem interface {
	MarshalBatchLineItem() []byte
}

type BatchChatCompletionRequest struct {
	CustomID string                `json:"custom_id"`
	Body     ChatCompletionRequest `json:"body"`
	Method   string                `json:"method"`
	URL      BatchEndpoint         `json:"url"`
}

func (r BatchChatCompletionRequest) MarshalBatchLineItem() []byte {
	marshal, _ := json.Marshal(r)
	return marshal
}

type BatchCompletionRequest struct {
	CustomID string            `json:"custom_id"`
	Body     CompletionRequest `json:"body"`
	Method   string            `json:"method"`
	URL      BatchEndpoint     `json:"url"`
}

func (r BatchCompletionRequest) MarshalBatchLineItem() []byte {
	marshal, _ := json.Marshal(r)
	return marshal
}

type BatchEmbeddingRequest struct {
	CustomID string           `json:"custom_id"`
	Body     EmbeddingRequest `json:"body"`
	Method   string           `json:"method"`
	URL      BatchEndpoint    `json:"url"`
}

func (r BatchEmbeddingRequest) MarshalBatchLineItem() []byte {
	marshal, _ := json.Marshal(r)
	return marshal
}

type Batch struct {
	ID               string             `json:"id"`
	Object           string             `json:"object"`
	InputFileID      string             `json:"input_file_id"`
	CompletionWindow string             `json:"completion_window"`
	Endpoint         BatchEndpoint      `json:"endpoint"`
	Status           BatchStatus        `json:"status"`
	Errors           *BatchErrors       `json:"errors,omitempty"`
	OutputFileID     string             `json:"output_file_id,omitempty"`
	ErrorFileID      string             `json:"error_file_id,omitempty"`
	CreatedAt        int64              `json:"created_at"`
	InProgressAt     int64              `json:"in_progress_at,omitempty"`
	ExpiresAt        int64              `json:"expires_at,omitempty"`
	FinalizingAt     int64              `json:"finalizing_at,omitempty"`
	CompletedAt      int64              `json:"completed_at,omitempty"`
	FailedAt         int64              `json:"failed_at,omitempty"`
	ExpiredAt        int64              `json:"expired_at,omitempty"`
	CancellingAt     int64              `json:"cancelling_at,omitempty"`
	CancelledAt      int64              `json:"cancelled_at,omitempty"`
	RequestCounts    BatchRequestCounts `json:"request_counts,omitempty"`
	Metadata         map[string]any     `json:"metadata"`
}

type BatchErrors struct {
	Object string       `json:"object,omitempty"`
	Data   []BatchError `json:"data"`
}

type BatchError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Param   string `json:"param,omitempty"`
	Line    int    `json:"line,omitempty"`
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

type UploadBatchFileRequest struct {
	FileName string
	Lines    []BatchLineItem
}

func (r *UploadBatchFileRequest) MarshalJSONL() []byte {
	buff := bytes.Buffer{}
	for i, line := range r.Lines {
		if i != 0 {
			buff.Write([]byte("\n"))
		}
		buff.Write(line.MarshalBatchLineItem())
	}
	return buff.Bytes()
}

func (r *UploadBatchFileRequest) AddChatCompletion(customerID string, body ChatCompletionRequest) {
	r.Lines = append(r.Lines, BatchChatCompletionRequest{
		CustomID: customerID,
		Body:     body,
		Method:   "POST",
		URL:      BatchEndpointChatCompletions,
	})
}

func (r *UploadBatchFileRequest) AddCompletion(customerID string, body CompletionRequest) {
	r.Lines = append(r.Lines, BatchCompletionRequest{
		CustomID: customerID,
		Body:     body,
		Method:   "POST",
		URL:      BatchEndpointCompletions,
	})
}

func (r *UploadBatchFileRequest) AddEmbedding(customerID string, body EmbeddingRequest) {
	r.Lines = append(r.Lines, BatchEmbeddingRequest{
		CustomID: customerID,
		Body:     body,
		Method:   "POST",
		URL:      BatchEndpointEmbeddings,
	})
}

// UploadBatchFile — upload batch file.
func (c *Client) UploadBatchFile(ctx context.Context, request UploadBatchFileRequest) (File, error) {
	if request.FileName == "" {
		request.FileName = "@batchinput.jsonl"
	}
	return c.CreateFileBytes(ctx, FileBytesRequest{
		Name:    request.FileName,
		Bytes:   request.MarshalJSONL(),
		Purpose: PurposeBatch,
	})
}

type CreateBatchWithUploadFileRequest struct {
	Endpoint         BatchEndpoint  `json:"endpoint"`
	CompletionWindow string         `json:"completion_window"`
	Metadata         map[string]any `json:"metadata"`
	UploadBatchFileRequest
}

// CreateBatchWithUploadFile — API call to Create batch with upload file.
func (c *Client) CreateBatchWithUploadFile(
	ctx context.Context,
	request CreateBatchWithUploadFileRequest,
) (response BatchResponse, err error) {
	var file File
	file, err = c.UploadBatchFile(ctx, UploadBatchFileRequest{
		FileName: request.FileName,
		Lines:    request.Lines,
	})
	if err != nil {
		return
	}
	return c.CreateBatch(ctx, CreateBatchRequest{
		InputFileID:      file.ID,
		Endpoint:         request.Endpoint,
		CompletionWindow: request.CompletionWindow,
		Metadata:         request.Metadata,
	})
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
