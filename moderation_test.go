package openai_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/alejandrojnm/go-openai"
	"github.com/alejandrojnm/go-openai/internal/test/checks"
)

// TestModeration Tests the moderations endpoint of the API using the mocked server.
func TestModerations(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/moderations", handleModerationEndpoint)
	_, err := client.Moderations(context.Background(), openai.ModerationRequest{
		Model: openai.ModerationTextStable,
		Input: "I want to kill them.",
	})
	checks.NoError(t, err, "Moderation error")
}

// TestModerationsWithIncorrectModel Tests passing valid and invalid models to moderations endpoint.
func TestModerationsWithDifferentModelOptions(t *testing.T) {
	var modelOptions []struct {
		model  string
		expect error
	}
	modelOptions = append(modelOptions,
		getModerationModelTestOption(openai.GPT3Dot5Turbo, openai.ErrModerationInvalidModel),
		getModerationModelTestOption(openai.ModerationTextStable, nil),
		getModerationModelTestOption(openai.ModerationTextLatest, nil),
		getModerationModelTestOption(openai.ModerationOmni20240926, nil),
		getModerationModelTestOption(openai.ModerationOmniLatest, nil),
		getModerationModelTestOption("", nil),
	)
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/moderations", handleModerationEndpoint)
	for _, modelTest := range modelOptions {
		_, err := client.Moderations(context.Background(), openai.ModerationRequest{
			Model: modelTest.model,
			Input: "I want to kill them.",
		})
		checks.ErrorIs(t, err, modelTest.expect,
			fmt.Sprintf("Moderations(..) expects err: %v, actual err:%v", modelTest.expect, err))
	}
}

func getModerationModelTestOption(model string, expect error) struct {
	model  string
	expect error
} {
	return struct {
		model  string
		expect error
	}{model: model, expect: expect}
}

// handleModerationEndpoint Handles the moderation endpoint by the test server.
func handleModerationEndpoint(w http.ResponseWriter, r *http.Request) {
	var err error
	var resBytes []byte

	// completions only accepts POST requests
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
	var moderationReq openai.ModerationRequest
	if moderationReq, err = getModerationBody(r); err != nil {
		http.Error(w, "could not read request", http.StatusInternalServerError)
		return
	}

	resCat := openai.ResultCategories{}
	resCatScore := openai.ResultCategoryScores{}
	switch {
	case strings.Contains(moderationReq.Input, "hate"):
		resCat = openai.ResultCategories{Hate: true}
		resCatScore = openai.ResultCategoryScores{Hate: 1}

	case strings.Contains(moderationReq.Input, "hate more"):
		resCat = openai.ResultCategories{HateThreatening: true}
		resCatScore = openai.ResultCategoryScores{HateThreatening: 1}

	case strings.Contains(moderationReq.Input, "harass"):
		resCat = openai.ResultCategories{Harassment: true}
		resCatScore = openai.ResultCategoryScores{Harassment: 1}

	case strings.Contains(moderationReq.Input, "harass hard"):
		resCat = openai.ResultCategories{Harassment: true}
		resCatScore = openai.ResultCategoryScores{HarassmentThreatening: 1}

	case strings.Contains(moderationReq.Input, "suicide"):
		resCat = openai.ResultCategories{SelfHarm: true}
		resCatScore = openai.ResultCategoryScores{SelfHarm: 1}

	case strings.Contains(moderationReq.Input, "wanna suicide"):
		resCat = openai.ResultCategories{SelfHarmIntent: true}
		resCatScore = openai.ResultCategoryScores{SelfHarm: 1}

	case strings.Contains(moderationReq.Input, "drink bleach"):
		resCat = openai.ResultCategories{SelfHarmInstructions: true}
		resCatScore = openai.ResultCategoryScores{SelfHarmInstructions: 1}

	case strings.Contains(moderationReq.Input, "porn"):
		resCat = openai.ResultCategories{Sexual: true}
		resCatScore = openai.ResultCategoryScores{Sexual: 1}

	case strings.Contains(moderationReq.Input, "child porn"):
		resCat = openai.ResultCategories{SexualMinors: true}
		resCatScore = openai.ResultCategoryScores{SexualMinors: 1}

	case strings.Contains(moderationReq.Input, "kill"):
		resCat = openai.ResultCategories{Violence: true}
		resCatScore = openai.ResultCategoryScores{Violence: 1}

	case strings.Contains(moderationReq.Input, "corpse"):
		resCat = openai.ResultCategories{ViolenceGraphic: true}
		resCatScore = openai.ResultCategoryScores{ViolenceGraphic: 1}
	}

	result := openai.Result{Categories: resCat, CategoryScores: resCatScore, Flagged: true}

	res := openai.ModerationResponse{
		ID:    strconv.Itoa(int(time.Now().Unix())),
		Model: moderationReq.Model,
	}
	res.Results = append(res.Results, result)

	resBytes, _ = json.Marshal(res)
	fmt.Fprintln(w, string(resBytes))
}

// getModerationBody Returns the body of the request to do a moderation.
func getModerationBody(r *http.Request) (openai.ModerationRequest, error) {
	moderation := openai.ModerationRequest{}
	// read the request body
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		return openai.ModerationRequest{}, err
	}
	err = json.Unmarshal(reqBody, &moderation)
	if err != nil {
		return openai.ModerationRequest{}, err
	}
	return moderation, nil
}
