package openai

import (
	"context"
	"net/http"
)

const (
	threadsSuffix = "/threads"
)

type Thread struct {
	ID            string         `json:"id"`
	Object        string         `json:"object"`
	CreatedAt     int64          `json:"created_at"`
	Metadata      map[string]any `json:"metadata"`
	ToolResources ToolResources  `json:"tool_resources,omitempty"`

	httpHeader
}

type ThreadRequest struct {
	Messages      []ThreadMessage       `json:"messages,omitempty"`
	Metadata      map[string]any        `json:"metadata,omitempty"`
	ToolResources *ToolResourcesRequest `json:"tool_resources,omitempty"`
}

type ToolResources struct {
	CodeInterpreter *CodeInterpreterToolResources `json:"code_interpreter,omitempty"`
	FileSearch      *FileSearchToolResources      `json:"file_search,omitempty"`
}

type CodeInterpreterToolResources struct {
	FileIDs []string `json:"file_ids,omitempty"`
}

type FileSearchToolResources struct {
	VectorStoreIDs []string `json:"vector_store_ids,omitempty"`
}

type ToolResourcesRequest struct {
	CodeInterpreter *CodeInterpreterToolResourcesRequest `json:"code_interpreter,omitempty"`
	FileSearch      *FileSearchToolResourcesRequest      `json:"file_search,omitempty"`
}

type CodeInterpreterToolResourcesRequest struct {
	FileIDs []string `json:"file_ids,omitempty"`
}

type FileSearchToolResourcesRequest struct {
	VectorStoreIDs []string                   `json:"vector_store_ids,omitempty"`
	VectorStores   []VectorStoreToolResources `json:"vector_stores,omitempty"`
}

type VectorStoreToolResources struct {
	FileIDs          []string          `json:"file_ids,omitempty"`
	ChunkingStrategy *ChunkingStrategy `json:"chunking_strategy,omitempty"`
	Metadata         map[string]any    `json:"metadata,omitempty"`
}

type ChunkingStrategy struct {
	Type   ChunkingStrategyType    `json:"type"`
	Static *StaticChunkingStrategy `json:"static,omitempty"`
}

type StaticChunkingStrategy struct {
	MaxChunkSizeTokens int `json:"max_chunk_size_tokens"`
	ChunkOverlapTokens int `json:"chunk_overlap_tokens"`
}

type ChunkingStrategyType string

const (
	ChunkingStrategyTypeAuto   ChunkingStrategyType = "auto"
	ChunkingStrategyTypeStatic ChunkingStrategyType = "static"
)

type ModifyThreadRequest struct {
	Metadata      map[string]any `json:"metadata"`
	ToolResources *ToolResources `json:"tool_resources,omitempty"`
}

type ThreadMessageRole string

const (
	ThreadMessageRoleAssistant ThreadMessageRole = "assistant"
	ThreadMessageRoleUser      ThreadMessageRole = "user"
)

type ThreadMessage struct {
	Role        ThreadMessageRole      `json:"role"`
	Content     []ThreadMessageContent `json:"content"`
	Attachments []ThreadAttachment     `json:"attachments,omitempty"`
	Metadata    map[string]any         `json:"metadata,omitempty"`
}

type ThreadMessageContent struct {
	Type      string                  `json:"type"`
	ImageFile *ThreadMessageImageFile `json:"image_file,omitempty"`
	Text      string                  `json:"text,omitempty"`
}

type ThreadMessageImageFile struct {
	FileID string  `json:"file_id"`
	Detail *string `json:"detail,omitempty"`
}

type ThreadAttachment struct {
	FileID string                 `json:"file_id"`
	Tools  []ThreadAttachmentTool `json:"tools"`
}

type ThreadAttachmentTool struct {
	Type string `json:"type"`
}

type ThreadDeleteResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Deleted bool   `json:"deleted"`

	httpHeader
}

// CreateThread creates a new thread.
func (c *Client) CreateThread(ctx context.Context, request ThreadRequest) (response Thread, err error) {
	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL(threadsSuffix), withBody(request),
		withBetaAssistantVersion(c.config.AssistantVersion))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// RetrieveThread retrieves a thread.
func (c *Client) RetrieveThread(ctx context.Context, threadID string) (response Thread, err error) {
	urlSuffix := threadsSuffix + "/" + threadID
	req, err := c.newRequest(ctx, http.MethodGet, c.fullURL(urlSuffix),
		withBetaAssistantVersion(c.config.AssistantVersion))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// ModifyThread modifies a thread.
func (c *Client) ModifyThread(
	ctx context.Context,
	threadID string,
	request ModifyThreadRequest,
) (response Thread, err error) {
	urlSuffix := threadsSuffix + "/" + threadID
	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL(urlSuffix), withBody(request),
		withBetaAssistantVersion(c.config.AssistantVersion))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// DeleteThread deletes a thread.
func (c *Client) DeleteThread(
	ctx context.Context,
	threadID string,
) (response ThreadDeleteResponse, err error) {
	urlSuffix := threadsSuffix + "/" + threadID
	req, err := c.newRequest(ctx, http.MethodDelete, c.fullURL(urlSuffix),
		withBetaAssistantVersion(c.config.AssistantVersion))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}
