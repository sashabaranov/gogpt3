package openai

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type Run struct {
	ID             string             `json:"id"`
	Object         string             `json:"object"`
	CreatedAt      int64              `json:"created_at"`
	ThreadID       string             `json:"thread_id"`
	AssistantID    string             `json:"assistant_id"`
	Status         RunStatus          `json:"status"`
	RequiredAction *RunRequiredAction `json:"required_action,omitempty"`
	LastError      *RunLastError      `json:"last_error,omitempty"`
	ExpiresAt      int64              `json:"expires_at"`
	StartedAt      *int64             `json:"started_at,omitempty"`
	CancelledAt    *int64             `json:"cancelled_at,omitempty"`
	FailedAt       *int64             `json:"failed_at,omitempty"`
	CompletedAt    *int64             `json:"completed_at,omitempty"`
	Model          string             `json:"model"`
	Instructions   string             `json:"instructions,omitempty"`
	Tools          []Tool             `json:"tools"`
	FileIDS        []string           `json:"file_ids"` //nolint:revive // backwards-compatibility
	Metadata       map[string]any     `json:"metadata"`
	Usage          Usage              `json:"usage,omitempty"`

	Temperature *float32 `json:"temperature,omitempty"`
	// The maximum number of prompt tokens that may be used over the course of the run.
	// If the run exceeds the number of prompt tokens specified, the run will end with status 'incomplete'.
	MaxPromptTokens int `json:"max_prompt_tokens,omitempty"`
	// The maximum number of completion tokens that may be used over the course of the run.
	// If the run exceeds the number of completion tokens specified, the run will end with status 'incomplete'.
	MaxCompletionTokens int `json:"max_completion_tokens,omitempty"`
	// ThreadTruncationStrategy defines the truncation strategy to use for the thread.
	TruncationStrategy *ThreadTruncationStrategy `json:"truncation_strategy,omitempty"`

	httpHeader
}

type RunStatus string

const (
	RunStatusQueued         RunStatus = "queued"
	RunStatusInProgress     RunStatus = "in_progress"
	RunStatusRequiresAction RunStatus = "requires_action"
	RunStatusCancelling     RunStatus = "cancelling"
	RunStatusFailed         RunStatus = "failed"
	RunStatusCompleted      RunStatus = "completed"
	RunStatusIncomplete     RunStatus = "incomplete"
	RunStatusExpired        RunStatus = "expired"
	RunStatusCancelled      RunStatus = "cancelled"
)

type RunRequiredAction struct {
	Type              RequiredActionType `json:"type"`
	SubmitToolOutputs *SubmitToolOutputs `json:"submit_tool_outputs,omitempty"`
}

type RequiredActionType string

const (
	RequiredActionTypeSubmitToolOutputs RequiredActionType = "submit_tool_outputs"
)

type SubmitToolOutputs struct {
	ToolCalls []ToolCall `json:"tool_calls"`
}

type RunLastError struct {
	Code    RunError `json:"code"`
	Message string   `json:"message"`
}

type RunError string

const (
	RunErrorServerError       RunError = "server_error"
	RunErrorRateLimitExceeded RunError = "rate_limit_exceeded"
)

type RunRequest struct {
	AssistantID            string          `json:"assistant_id"`
	Model                  string          `json:"model,omitempty"`
	Instructions           string          `json:"instructions,omitempty"`
	AdditionalInstructions string          `json:"additional_instructions,omitempty"`
	AdditionalMessages     []ThreadMessage `json:"additional_messages,omitempty"`
	Tools                  []Tool          `json:"tools,omitempty"`
	Metadata               map[string]any  `json:"metadata,omitempty"`

	// Sampling temperature between 0 and 2. Higher values like 0.8 are  more random.
	// lower values are more focused and deterministic.
	Temperature *float32 `json:"temperature,omitempty"`
	TopP        *float32 `json:"top_p,omitempty"`

	// The maximum number of prompt tokens that may be used over the course of the run.
	// If the run exceeds the number of prompt tokens specified, the run will end with status 'incomplete'.
	MaxPromptTokens int `json:"max_prompt_tokens,omitempty"`

	// The maximum number of completion tokens that may be used over the course of the run.
	// If the run exceeds the number of completion tokens specified, the run will end with status 'incomplete'.
	MaxCompletionTokens int `json:"max_completion_tokens,omitempty"`

	// ThreadTruncationStrategy defines the truncation strategy to use for the thread.
	TruncationStrategy *ThreadTruncationStrategy `json:"truncation_strategy,omitempty"`

	// This can be either a string or a ToolChoice object.
	ToolChoice any `json:"tool_choice,omitempty"`
	// This can be either a string or a ResponseFormat object.
	ResponseFormat any `json:"response_format,omitempty"`
	// Disable the default behavior of parallel tool calls by setting it: false.
	ParallelToolCalls any `json:"parallel_tool_calls,omitempty"`
}

// ThreadTruncationStrategy defines the truncation strategy to use for the thread.
// https://platform.openai.com/docs/assistants/how-it-works/truncation-strategy.
type ThreadTruncationStrategy struct {
	// default 'auto'.
	Type TruncationStrategy `json:"type,omitempty"`
	// this field should be set if the truncation strategy is set to LastMessages.
	LastMessages *int `json:"last_messages,omitempty"`
}

// TruncationStrategy defines the existing truncation strategies existing for thread management in an assistant.
type TruncationStrategy string

const (
	// TruncationStrategyAuto messages in the middle of the thread will be dropped to fit the context length of the model.
	TruncationStrategyAuto = TruncationStrategy("auto")
	// TruncationStrategyLastMessages the thread will be truncated to the n most recent messages in the thread.
	TruncationStrategyLastMessages = TruncationStrategy("last_messages")
)

type RunRequestStreaming struct {
	RunRequest
	Stream bool `json:"stream"`
}

// ReponseFormat specifies the format the model must output.
// https://platform.openai.com/docs/api-reference/runs/createRun#runs-createrun-response_format.
// Type can either be text or json_object.
type ReponseFormat struct {
	Type string `json:"type"`
}

type RunModifyRequest struct {
	Metadata map[string]any `json:"metadata,omitempty"`
}

// RunList is a list of runs.
type RunList struct {
	Runs []Run `json:"data"`

	httpHeader
}

type SubmitToolOutputsRequest struct {
	ToolOutputs []ToolOutput `json:"tool_outputs"`
}

type ToolOutput struct {
	ToolCallID string `json:"tool_call_id"`
	Output     any    `json:"output"`
}

type CreateThreadAndRunRequest struct {
	RunRequest
	Thread ThreadRequest `json:"thread"`
}

type CreateThreadAndStreamRequest struct {
	CreateThreadAndRunRequest
	Stream bool `json:"stream"`
}

type RunStep struct {
	ID          string         `json:"id"`
	Object      string         `json:"object"`
	CreatedAt   int64          `json:"created_at"`
	AssistantID string         `json:"assistant_id"`
	ThreadID    string         `json:"thread_id"`
	RunID       string         `json:"run_id"`
	Type        RunStepType    `json:"type"`
	Status      RunStepStatus  `json:"status"`
	StepDetails StepDetails    `json:"step_details"`
	LastError   *RunLastError  `json:"last_error,omitempty"`
	ExpiredAt   *int64         `json:"expired_at,omitempty"`
	CancelledAt *int64         `json:"cancelled_at,omitempty"`
	FailedAt    *int64         `json:"failed_at,omitempty"`
	CompletedAt *int64         `json:"completed_at,omitempty"`
	Metadata    map[string]any `json:"metadata"`

	httpHeader
}

type RunStepStatus string

const (
	RunStepStatusInProgress RunStepStatus = "in_progress"
	RunStepStatusCancelling RunStepStatus = "cancelled"
	RunStepStatusFailed     RunStepStatus = "failed"
	RunStepStatusCompleted  RunStepStatus = "completed"
	RunStepStatusExpired    RunStepStatus = "expired"
)

type RunStepType string

const (
	RunStepTypeMessageCreation RunStepType = "message_creation"
	RunStepTypeToolCalls       RunStepType = "tool_calls"
)

type StepDetails struct {
	Type            RunStepType                 `json:"type"`
	MessageCreation *StepDetailsMessageCreation `json:"message_creation,omitempty"`
	ToolCalls       []ToolCall                  `json:"tool_calls,omitempty"`
}

type StepDetailsMessageCreation struct {
	MessageID string `json:"message_id"`
}

// RunStepList is a list of steps.
type RunStepList struct {
	RunSteps []RunStep `json:"data"`

	FirstID string `json:"first_id"`
	LastID  string `json:"last_id"`
	HasMore bool   `json:"has_more"`

	httpHeader
}

type Pagination struct {
	Limit  *int
	Order  *string
	After  *string
	Before *string
}

// CreateRun creates a new run.
func (c *Client) CreateRun(
	ctx context.Context,
	threadID string,
	request RunRequest,
) (response Run, err error) {
	urlSuffix := fmt.Sprintf("/threads/%s/runs", threadID)
	req, err := c.newRequest(
		ctx,
		http.MethodPost,
		c.fullURL(urlSuffix),
		withBody(request),
		withBetaAssistantVersion(c.config.AssistantVersion))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// RetrieveRun retrieves a run.
func (c *Client) RetrieveRun(
	ctx context.Context,
	threadID string,
	runID string,
) (response Run, err error) {
	urlSuffix := fmt.Sprintf("/threads/%s/runs/%s", threadID, runID)
	req, err := c.newRequest(
		ctx,
		http.MethodGet,
		c.fullURL(urlSuffix),
		withBetaAssistantVersion(c.config.AssistantVersion))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// ModifyRun modifies a run.
func (c *Client) ModifyRun(
	ctx context.Context,
	threadID string,
	runID string,
	request RunModifyRequest,
) (response Run, err error) {
	urlSuffix := fmt.Sprintf("/threads/%s/runs/%s", threadID, runID)
	req, err := c.newRequest(
		ctx,
		http.MethodPost,
		c.fullURL(urlSuffix),
		withBody(request),
		withBetaAssistantVersion(c.config.AssistantVersion))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// ListRuns lists runs.
func (c *Client) ListRuns(
	ctx context.Context,
	threadID string,
	pagination Pagination,
) (response RunList, err error) {
	urlValues := url.Values{}
	if pagination.Limit != nil {
		urlValues.Add("limit", fmt.Sprintf("%d", *pagination.Limit))
	}
	if pagination.Order != nil {
		urlValues.Add("order", *pagination.Order)
	}
	if pagination.After != nil {
		urlValues.Add("after", *pagination.After)
	}
	if pagination.Before != nil {
		urlValues.Add("before", *pagination.Before)
	}

	encodedValues := ""
	if len(urlValues) > 0 {
		encodedValues = "?" + urlValues.Encode()
	}

	urlSuffix := fmt.Sprintf("/threads/%s/runs%s", threadID, encodedValues)
	req, err := c.newRequest(
		ctx,
		http.MethodGet,
		c.fullURL(urlSuffix),
		withBetaAssistantVersion(c.config.AssistantVersion))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// SubmitToolOutputs submits tool outputs.
func (c *Client) SubmitToolOutputs(
	ctx context.Context,
	threadID string,
	runID string,
	request SubmitToolOutputsRequest) (response Run, err error) {
	urlSuffix := fmt.Sprintf("/threads/%s/runs/%s/submit_tool_outputs", threadID, runID)
	req, err := c.newRequest(
		ctx,
		http.MethodPost,
		c.fullURL(urlSuffix),
		withBody(request),
		withBetaAssistantVersion(c.config.AssistantVersion))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

type SubmitToolOutputsStreamRequest struct {
	SubmitToolOutputsRequest
	Stream bool `json:"stream"`
}

func (c *Client) SubmitToolOutputsStream(
	ctx context.Context,
	threadID string,
	runID string,
	request SubmitToolOutputsRequest,
) (stream *AssistantStream, err error) {
	urlSuffix := fmt.Sprintf("/threads/%s/runs/%s/submit_tool_outputs", threadID, runID)
	r := SubmitToolOutputsStreamRequest{
		SubmitToolOutputsRequest: request,
		Stream:                   true,
	}
	req, err := c.newRequest(
		ctx,
		http.MethodPost,
		c.fullURL(urlSuffix),
		withBody(r),
		withBetaAssistantVersion(c.config.AssistantVersion),
	)
	if err != nil {
		return
	}

	resp, err := sendRequestStream[AssistantStreamEvent](c, req)
	if err != nil {
		return
	}
	stream = &AssistantStream{
		streamReader: resp,
	}
	return
}

// CancelRun cancels a run.
func (c *Client) CancelRun(
	ctx context.Context,
	threadID string,
	runID string) (response Run, err error) {
	urlSuffix := fmt.Sprintf("/threads/%s/runs/%s/cancel", threadID, runID)
	req, err := c.newRequest(
		ctx,
		http.MethodPost,
		c.fullURL(urlSuffix),
		withBetaAssistantVersion(c.config.AssistantVersion))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// CreateThreadAndRun submits tool outputs.
func (c *Client) CreateThreadAndRun(
	ctx context.Context,
	request CreateThreadAndRunRequest) (response Run, err error) {
	urlSuffix := "/threads/runs"
	req, err := c.newRequest(
		ctx,
		http.MethodPost,
		c.fullURL(urlSuffix),
		withBody(request),
		withBetaAssistantVersion(c.config.AssistantVersion))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

type StreamMessageDelta struct {
	Role    string           `json:"role"`
	Content []MessageContent `json:"content"`
	FileIDs []string         `json:"file_ids"`
}

type AssistantStreamEvent struct {
	ID     string             `json:"id"`
	Object string             `json:"object"`
	Delta  StreamMessageDelta `json:"delta,omitempty"`

	// Run
	CreatedAt      int64              `json:"created_at,omitempty"`
	ThreadID       string             `json:"thread_id,omitempty"`
	AssistantID    string             `json:"assistant_id,omitempty"`
	Status         RunStatus          `json:"status,omitempty"`
	RequiredAction *RunRequiredAction `json:"required_action,omitempty"`
	LastError      *RunLastError      `json:"last_error,omitempty"`
	ExpiresAt      int64              `json:"expires_at,omitempty"`
	StartedAt      *int64             `json:"started_at,omitempty"`
	CancelledAt    *int64             `json:"cancelled_at,omitempty"`
	FailedAt       *int64             `json:"failed_at,omitempty"`
	CompletedAt    *int64             `json:"completed_at,omitempty"`
	Model          string             `json:"model,omitempty"`
	Instructions   string             `json:"instructions,omitempty"`
	Tools          []Tool             `json:"tools,omitempty"`
	FileIDS        []string           `json:"file_ids"` //nolint:revive // backwards-compatibility
	Metadata       map[string]any     `json:"metadata,omitempty"`
	Usage          Usage              `json:"usage,omitempty"`

	// ThreadMessage.Completed
	Role    string           `json:"role,omitempty"`
	Content []MessageContent `json:"content,omitempty"`
	// IncompleteDetails
	// IncompleteAt

	// Run steps
	RunID       string      `json:"run_id"`
	Type        RunStepType `json:"type"`
	StepDetails StepDetails `json:"step_details"`
	ExpiredAt   *int64      `json:"expired_at,omitempty"`
}

type AssistantStream struct {
	*streamReader[AssistantStreamEvent]
}

func (c *Client) CreateThreadAndStream(
	ctx context.Context,
	request CreateThreadAndRunRequest) (stream *AssistantStream, err error) {
	urlSuffix := "/threads/runs"
	sr := CreateThreadAndStreamRequest{
		CreateThreadAndRunRequest: request,
		Stream:                    true,
	}
	req, err := c.newRequest(
		ctx,
		http.MethodPost,
		c.fullURL(urlSuffix),
		withBody(sr),
		withBetaAssistantVersion(c.config.AssistantVersion),
	)
	if err != nil {
		return
	}

	resp, err := sendRequestStream[AssistantStreamEvent](c, req)
	if err != nil {
		return
	}
	stream = &AssistantStream{
		streamReader: resp,
	}
	return
}

func (c *Client) CreateRunStreaming(
	ctx context.Context,
	threadID string,
	request RunRequest) (stream *AssistantStream, err error) {
	urlSuffix := fmt.Sprintf("/threads/%s/runs", threadID)

	r := RunRequestStreaming{
		RunRequest: request,
		Stream:     true,
	}

	req, err := c.newRequest(
		ctx,
		http.MethodPost,
		c.fullURL(urlSuffix),
		withBody(r),
		withBetaAssistantVersion(c.config.AssistantVersion),
	)
	if err != nil {
		return
	}

	resp, err := sendRequestStream[AssistantStreamEvent](c, req)
	if err != nil {
		return
	}
	stream = &AssistantStream{
		streamReader: resp,
	}
	return
}

// RetrieveRunStep retrieves a run step.
func (c *Client) RetrieveRunStep(
	ctx context.Context,
	threadID string,
	runID string,
	stepID string,
) (response RunStep, err error) {
	urlSuffix := fmt.Sprintf("/threads/%s/runs/%s/steps/%s", threadID, runID, stepID)
	req, err := c.newRequest(
		ctx,
		http.MethodGet,
		c.fullURL(urlSuffix),
		withBetaAssistantVersion(c.config.AssistantVersion))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// ListRunSteps lists run steps.
func (c *Client) ListRunSteps(
	ctx context.Context,
	threadID string,
	runID string,
	pagination Pagination,
) (response RunStepList, err error) {
	urlValues := url.Values{}
	if pagination.Limit != nil {
		urlValues.Add("limit", fmt.Sprintf("%d", *pagination.Limit))
	}
	if pagination.Order != nil {
		urlValues.Add("order", *pagination.Order)
	}
	if pagination.After != nil {
		urlValues.Add("after", *pagination.After)
	}
	if pagination.Before != nil {
		urlValues.Add("before", *pagination.Before)
	}

	encodedValues := ""
	if len(urlValues) > 0 {
		encodedValues = "?" + urlValues.Encode()
	}

	urlSuffix := fmt.Sprintf("/threads/%s/runs/%s/steps%s", threadID, runID, encodedValues)
	req, err := c.newRequest(
		ctx,
		http.MethodGet,
		c.fullURL(urlSuffix),
		withBetaAssistantVersion(c.config.AssistantVersion))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}
