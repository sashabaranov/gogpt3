package openai_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/alejandrojnm/go-openai"
	"github.com/alejandrojnm/go-openai/internal/test/checks"
)

const testFineTuninigJobID = "fine-tuning-job-id"

// TestFineTuningJob Tests the fine tuning job endpoint of the API using the mocked server.
func TestFineTuningJob(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler(
		"/v1/fine_tuning/jobs",
		func(w http.ResponseWriter, _ *http.Request) {
			resBytes, _ := json.Marshal(openai.FineTuningJob{
				Object:         "fine_tuning.job",
				ID:             testFineTuninigJobID,
				Model:          "davinci-002",
				CreatedAt:      1692661014,
				FinishedAt:     1692661190,
				FineTunedModel: "ft:davinci-002:my-org:custom_suffix:7q8mpxmy",
				OrganizationID: "org-123",
				ResultFiles:    []string{"file-abc123"},
				Status:         "succeeded",
				ValidationFile: "",
				TrainingFile:   "file-abc123",
				Hyperparameters: openai.Hyperparameters{
					Epochs:                 "auto",
					LearningRateMultiplier: "auto",
					BatchSize:              "auto",
				},
				TrainedTokens: 5768,
			})
			fmt.Fprintln(w, string(resBytes))
		},
	)

	server.RegisterHandler(
		"/v1/fine_tuning/jobs/"+testFineTuninigJobID+"/cancel",
		func(w http.ResponseWriter, _ *http.Request) {
			resBytes, _ := json.Marshal(openai.FineTuningJob{})
			fmt.Fprintln(w, string(resBytes))
		},
	)

	server.RegisterHandler(
		"/v1/fine_tuning/jobs/"+testFineTuninigJobID,
		func(w http.ResponseWriter, _ *http.Request) {
			var resBytes []byte
			resBytes, _ = json.Marshal(openai.FineTuningJob{})
			fmt.Fprintln(w, string(resBytes))
		},
	)

	server.RegisterHandler(
		"/v1/fine_tuning/jobs/"+testFineTuninigJobID+"/events",
		func(w http.ResponseWriter, _ *http.Request) {
			resBytes, _ := json.Marshal(openai.FineTuningJobEventList{})
			fmt.Fprintln(w, string(resBytes))
		},
	)

	ctx := context.Background()

	_, err := client.CreateFineTuningJob(ctx, openai.FineTuningJobRequest{})
	checks.NoError(t, err, "CreateFineTuningJob error")

	_, err = client.CancelFineTuningJob(ctx, testFineTuninigJobID)
	checks.NoError(t, err, "CancelFineTuningJob error")

	_, err = client.RetrieveFineTuningJob(ctx, testFineTuninigJobID)
	checks.NoError(t, err, "RetrieveFineTuningJob error")

	_, err = client.ListFineTuningJobEvents(ctx, testFineTuninigJobID)
	checks.NoError(t, err, "ListFineTuningJobEvents error")

	_, err = client.ListFineTuningJobEvents(
		ctx,
		testFineTuninigJobID,
		openai.ListFineTuningJobEventsWithAfter("last-event-id"),
	)
	checks.NoError(t, err, "ListFineTuningJobEvents error")

	_, err = client.ListFineTuningJobEvents(
		ctx,
		testFineTuninigJobID,
		openai.ListFineTuningJobEventsWithLimit(10),
	)
	checks.NoError(t, err, "ListFineTuningJobEvents error")

	_, err = client.ListFineTuningJobEvents(
		ctx,
		testFineTuninigJobID,
		openai.ListFineTuningJobEventsWithAfter("last-event-id"),
		openai.ListFineTuningJobEventsWithLimit(10),
	)
	checks.NoError(t, err, "ListFineTuningJobEvents error")
}
