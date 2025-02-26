package openai_test

import (
	"context"

	openai "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test/checks"

	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

// TestAssistant Tests the assistant endpoint of the API using the mocked server.
func TestRun(t *testing.T) {
	assistantID := "asst_abc123"
	threadID := "thread_abc123"
	runID := "run_abc123"
	stepID := "step_abc123"
	limit := 20
	order := "desc"
	after := "asst_abc122"
	before := "asst_abc124"

	client, server, teardown := setupOpenAITestServer()
	defer teardown()

	server.RegisterHandler(
		"/v1/threads/"+threadID+"/runs/"+runID+"/steps/"+stepID,
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				resBytes, _ := json.Marshal(openai.RunStep{
					ID:        runID,
					Object:    "run",
					CreatedAt: 1234567890,
					Status:    openai.RunStepStatusCompleted,
				})
				fmt.Fprintln(w, string(resBytes))
			}
		},
	)

	server.RegisterHandler(
		"/v1/threads/"+threadID+"/runs/"+runID+"/steps",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				resBytes, _ := json.Marshal(openai.RunStepList{
					RunSteps: []openai.RunStep{
						{
							ID:        runID,
							Object:    "run",
							CreatedAt: 1234567890,
							Status:    openai.RunStepStatusCompleted,
						},
					},
				})
				fmt.Fprintln(w, string(resBytes))
			}
		},
	)

	server.RegisterHandler(
		"/v1/threads/"+threadID+"/runs/"+runID+"/cancel",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost {
				resBytes, _ := json.Marshal(openai.Run{
					ID:        runID,
					Object:    "run",
					CreatedAt: 1234567890,
					Status:    openai.RunStatusCancelling,
				})
				fmt.Fprintln(w, string(resBytes))
			}
		},
	)

	server.RegisterHandler(
		"/v1/threads/"+threadID+"/runs/"+runID+"/submit_tool_outputs",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost {
				resBytes, _ := json.Marshal(openai.Run{
					ID:        runID,
					Object:    "run",
					CreatedAt: 1234567890,
					Status:    openai.RunStatusCancelling,
				})
				fmt.Fprintln(w, string(resBytes))
			}
		},
	)

	server.RegisterHandler(
		"/v1/threads/"+threadID+"/runs/"+runID,
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				resBytes, _ := json.Marshal(openai.Run{
					ID:        runID,
					Object:    "run",
					CreatedAt: 1234567890,
					Status:    openai.RunStatusQueued,
				})
				fmt.Fprintln(w, string(resBytes))
			} else if r.Method == http.MethodPost {
				var request openai.RunModifyRequest
				err := json.NewDecoder(r.Body).Decode(&request)
				checks.NoError(t, err, "Decode error")

				resBytes, _ := json.Marshal(openai.Run{
					ID:        runID,
					Object:    "run",
					CreatedAt: 1234567890,
					Status:    openai.RunStatusQueued,
					Metadata:  request.Metadata,
				})
				fmt.Fprintln(w, string(resBytes))
			}
		},
	)

	server.RegisterHandler(
		"/v1/threads/"+threadID+"/runs",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost {
				var request openai.RunRequest
				err := json.NewDecoder(r.Body).Decode(&request)
				checks.NoError(t, err, "Decode error")

				resBytes, _ := json.Marshal(openai.Run{
					ID:        runID,
					Object:    "run",
					CreatedAt: 1234567890,
					Status:    openai.RunStatusQueued,
				})
				fmt.Fprintln(w, string(resBytes))
			} else if r.Method == http.MethodGet {
				resBytes, _ := json.Marshal(openai.RunList{
					Runs: []openai.Run{
						{
							ID:        runID,
							Object:    "run",
							CreatedAt: 1234567890,
							Status:    openai.RunStatusQueued,
						},
					},
				})
				fmt.Fprintln(w, string(resBytes))
			}
		},
	)

	server.RegisterHandler(
		"/v1/threads/runs",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost {
				var request openai.CreateThreadAndRunRequest
				err := json.NewDecoder(r.Body).Decode(&request)
				checks.NoError(t, err, "Decode error")

				resBytes, _ := json.Marshal(openai.Run{
					ID:        runID,
					Object:    "run",
					CreatedAt: 1234567890,
					Status:    openai.RunStatusQueued,
				})
				fmt.Fprintln(w, string(resBytes))
			}
		},
	)

	ctx := context.Background()

	_, err := client.CreateRun(ctx, threadID, openai.RunRequest{
		AssistantID: assistantID,
	})
	checks.NoError(t, err, "CreateRun error")

	_, err = client.RetrieveRun(ctx, threadID, runID)
	checks.NoError(t, err, "RetrieveRun error")

	_, err = client.ModifyRun(ctx, threadID, runID, openai.RunModifyRequest{
		Metadata: map[string]any{
			"key": "value",
		},
	})
	checks.NoError(t, err, "ModifyRun error")

	_, err = client.ListRuns(
		ctx,
		threadID,
		openai.Pagination{
			Limit:  &limit,
			Order:  &order,
			After:  &after,
			Before: &before,
		},
	)
	checks.NoError(t, err, "ListRuns error")

	_, err = client.SubmitToolOutputs(ctx, threadID, runID,
		openai.SubmitToolOutputsRequest{})
	checks.NoError(t, err, "SubmitToolOutputs error")

	_, err = client.CancelRun(ctx, threadID, runID)
	checks.NoError(t, err, "CancelRun error")

	_, err = client.CreateThreadAndRun(ctx, openai.CreateThreadAndRunRequest{
		RunRequest: openai.RunRequest{
			AssistantID: assistantID,
		},
		Thread: openai.ThreadRequest{
			Messages: []openai.ThreadMessage{
				{
					Role:    openai.ThreadMessageRoleUser,
					Content: "Hello, World!",
				},
			},
		},
	})
	checks.NoError(t, err, "CreateThreadAndRun error")

	_, err = client.CreateThreadAndStream(ctx, openai.CreateThreadAndRunRequest{
		RunRequest: openai.RunRequest{
			AssistantID: assistantID,
		},
		Thread: openai.ThreadRequest{
			Messages: []openai.ThreadMessage{
				{
					Role:    openai.ThreadMessageRoleUser,
					Content: "Hello, World!",
				},
			},
		},
	})
	checks.NoError(t, err, "CreateThreadAndStream error")

	_, err = client.CreateRunStreaming(ctx, threadID, openai.RunRequest{
		AssistantID: assistantID,
	})
	checks.NoError(t, err, "CreateRunStreaming error")

	_, err = client.SubmitToolOutputsStream(ctx, threadID, runID, openai.SubmitToolOutputsRequest{
		ToolOutputs: nil,
	})
	checks.NoError(t, err, "SubmitToolOutputsStream error")

	_, err = client.RetrieveRunStep(ctx, threadID, runID, stepID)
	checks.NoError(t, err, "RetrieveRunStep error")

	_, err = client.ListRunSteps(
		ctx,
		threadID,
		runID,
		openai.Pagination{
			Limit:  &limit,
			Order:  &order,
			After:  &after,
			Before: &before,
		},
	)
	checks.NoError(t, err, "ListRunSteps error")
}
