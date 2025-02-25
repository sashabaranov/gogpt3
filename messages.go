package openai

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

const (
	messagesSuffix = "messages"
)

type Message struct {
	ID          string           `json:"id"`
	Object      string           `json:"object"`
	CreatedAt   int              `json:"created_at"`
	ThreadID    string           `json:"thread_id"`
	Role        string           `json:"role"`
	Content     []MessageContent `json:"content"`
	FileIds     []string         `json:"file_ids"` //nolint:revive //backwards-compatibility
	AssistantID *string          `json:"assistant_id,omitempty"`
	RunID       *string          `json:"run_id,omitempty"`
	Metadata    map[string]any   `json:"metadata"`

	httpHeader
}

type MessagesList struct {
	Messages []Message `json:"data"`

	Object  string  `json:"object"`
	FirstID *string `json:"first_id"`
	LastID  *string `json:"last_id"`
	HasMore bool    `json:"has_more"`

	httpHeader
}

type MessageContent struct {
	Type      string       `json:"type"`
	Text      *MessageText `json:"text,omitempty"`
	ImageFile *ImageFile   `json:"image_file,omitempty"`
	ImageURL  *ImageURL    `json:"image_url,omitempty"`
}
type MessageText struct {
	Value       string `json:"value"`
	Annotations []any  `json:"annotations"`
}

type ImageFile struct {
	FileID string `json:"file_id"`
}

type ImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail"`
}

type MessageRequest struct {
	Role        string             `json:"role"`
	Content     string             `json:"content"`
	FileIds     []string           `json:"file_ids,omitempty"` //nolint:revive // backwards-compatibility
	Metadata    map[string]any     `json:"metadata,omitempty"`
	Attachments []ThreadAttachment `json:"attachments,omitempty"`
}

type MessageFile struct {
	ID        string `json:"id"`
	Object    string `json:"object"`
	CreatedAt int    `json:"created_at"`
	MessageID string `json:"message_id"`

	httpHeader
}

type MessageFilesList struct {
	MessageFiles []MessageFile `json:"data"`

	httpHeader
}

type MessageDeletionStatus struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Deleted bool   `json:"deleted"`

	httpHeader
}

// CreateMessage creates a new message.
func (c *Client) CreateMessage(ctx context.Context, threadID string, request MessageRequest) (msg Message, err error) {
	urlSuffix := fmt.Sprintf("/threads/%s/%s", threadID, messagesSuffix)
	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL(urlSuffix), withBody(request),
		withBetaAssistantVersion(c.config.AssistantVersion))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &msg)
	return
}

// ListMessage fetches all messages in the thread.
func (c *Client) ListMessage(ctx context.Context, threadID string,
	limit *int,
	order *string,
	after *string,
	before *string,
	runID *string,
) (messages MessagesList, err error) {
	urlValues := url.Values{}
	if limit != nil {
		urlValues.Add("limit", fmt.Sprintf("%d", *limit))
	}
	if order != nil {
		urlValues.Add("order", *order)
	}
	if after != nil {
		urlValues.Add("after", *after)
	}
	if before != nil {
		urlValues.Add("before", *before)
	}
	if runID != nil {
		urlValues.Add("run_id", *runID)
	}

	encodedValues := ""
	if len(urlValues) > 0 {
		encodedValues = "?" + urlValues.Encode()
	}

	urlSuffix := fmt.Sprintf("/threads/%s/%s%s", threadID, messagesSuffix, encodedValues)
	req, err := c.newRequest(ctx, http.MethodGet, c.fullURL(urlSuffix),
		withBetaAssistantVersion(c.config.AssistantVersion))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &messages)
	return
}

// RetrieveMessage retrieves a Message.
func (c *Client) RetrieveMessage(
	ctx context.Context,
	threadID, messageID string,
) (msg Message, err error) {
	urlSuffix := fmt.Sprintf("/threads/%s/%s/%s", threadID, messagesSuffix, messageID)
	req, err := c.newRequest(ctx, http.MethodGet, c.fullURL(urlSuffix),
		withBetaAssistantVersion(c.config.AssistantVersion))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &msg)
	return
}

// ModifyMessage modifies a message.
func (c *Client) ModifyMessage(
	ctx context.Context,
	threadID, messageID string,
	metadata map[string]any,
) (msg Message, err error) {
	urlSuffix := fmt.Sprintf("/threads/%s/%s/%s", threadID, messagesSuffix, messageID)
	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL(urlSuffix),
		withBody(map[string]any{"metadata": metadata}), withBetaAssistantVersion(c.config.AssistantVersion))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &msg)
	return
}

// RetrieveMessageFile fetches a message file.
func (c *Client) RetrieveMessageFile(
	ctx context.Context,
	threadID, messageID, fileID string,
) (file MessageFile, err error) {
	urlSuffix := fmt.Sprintf("/threads/%s/%s/%s/files/%s", threadID, messagesSuffix, messageID, fileID)
	req, err := c.newRequest(ctx, http.MethodGet, c.fullURL(urlSuffix),
		withBetaAssistantVersion(c.config.AssistantVersion))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &file)
	return
}

// ListMessageFiles fetches all files attached to a message.
func (c *Client) ListMessageFiles(
	ctx context.Context,
	threadID, messageID string,
) (files MessageFilesList, err error) {
	urlSuffix := fmt.Sprintf("/threads/%s/%s/%s/files", threadID, messagesSuffix, messageID)
	req, err := c.newRequest(ctx, http.MethodGet, c.fullURL(urlSuffix),
		withBetaAssistantVersion(c.config.AssistantVersion))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &files)
	return
}

// DeleteMessage deletes a message..
func (c *Client) DeleteMessage(
	ctx context.Context,
	threadID, messageID string,
) (status MessageDeletionStatus, err error) {
	urlSuffix := fmt.Sprintf("/threads/%s/%s/%s", threadID, messagesSuffix, messageID)
	req, err := c.newRequest(ctx, http.MethodDelete, c.fullURL(urlSuffix),
		withBetaAssistantVersion(c.config.AssistantVersion))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &status)
	return
}
