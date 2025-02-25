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

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test/checks"
)

// TestModeration Tests the moderations endpoint of the API using the mocked server.
func TestModerations(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()

	server.RegisterHandler("/v1/moderations", handleModerationEndpoint)

	var requestInputs = []openai.ModerationRequestConverter{
		openai.ModerationRequest{
			Model: openai.ModerationTextStable,
			Input: "I want to kill them.",
		},
		openai.ModerationStrArrayRequest{
			Input: []string{
				"I want to kill them.",
				"Hello World",
			},
			Model: openai.ModerationTextStable,
		},
		openai.ModerationArrayRequest{
			Input: []openai.ModerationRequestItem{
				{
					Type: openai.ModerationItemTypeText,
					Text: "I want to kill them.",
				},
				{
					Type: openai.ModerationItemTypeImageURL,
					ImageURL: openai.ModerationImageURL{
						URL: "https://cdn.openai.com/API/images/guides/image_variation_original.webp",
					},
				},
			},
			Model: openai.ModerationOmniLatest,
		},
		openai.ModerationArrayRequest{
			Input: []openai.ModerationRequestItem{
				{
					Type: openai.ModerationItemTypeImageURL,
					ImageURL: openai.ModerationImageURL{
						URL: "https://cdn.openai.com/API/images/harass.png",
					},
				},
			},
			Model: openai.ModerationOmniLatest,
		},
	}

	for _, input := range requestInputs {
		_, err := client.Moderations(context.Background(), input)
		checks.NoError(t, err, "Moderation error")
	}
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
	var moderationReq openai.ModerationArrayRequest
	if moderationReq, err = getModerationBody(r); err != nil {
		http.Error(w, "could not read request", http.StatusInternalServerError)
		return
	}

	res := openai.ModerationResponse{
		ID:    strconv.Itoa(int(time.Now().Unix())),
		Model: moderationReq.Model,
	}

	for i := range moderationReq.Input {
		var (
			resCat        = openai.ResultCategories{}
			resCatScore   = openai.ResultCategoryScores{}
			resCatApplied = openai.CategoryAppliedInputType{}
		)

		switch {
		case strings.Contains(moderationReq.Input[i].Text, "hate"):
			resCat = openai.ResultCategories{Hate: true}
			resCatScore = openai.ResultCategoryScores{Hate: 1}
			resCatApplied = openai.CategoryAppliedInputType{
				Hate: []openai.ModerationItemType{openai.ModerationItemTypeText},
			}

		case strings.Contains(moderationReq.Input[i].Text, "hate more"):
			resCat = openai.ResultCategories{HateThreatening: true}
			resCatScore = openai.ResultCategoryScores{HateThreatening: 1}
			resCatApplied = openai.CategoryAppliedInputType{
				HarassmentThreatening: []openai.ModerationItemType{openai.ModerationItemTypeText},
			}

		case strings.Contains(moderationReq.Input[i].Text, "harass"):
			resCat = openai.ResultCategories{Harassment: true}
			resCatScore = openai.ResultCategoryScores{Harassment: 1}
			resCatApplied = openai.CategoryAppliedInputType{
				Harassment: []openai.ModerationItemType{openai.ModerationItemTypeText},
			}

		case strings.Contains(moderationReq.Input[i].Text, "harass hard"):
			resCat = openai.ResultCategories{Harassment: true}
			resCatScore = openai.ResultCategoryScores{HarassmentThreatening: 1}
			resCatApplied = openai.CategoryAppliedInputType{
				HarassmentThreatening: []openai.ModerationItemType{openai.ModerationItemTypeText},
			}

		case strings.Contains(moderationReq.Input[i].Text, "suicide"):
			resCat = openai.ResultCategories{SelfHarm: true}
			resCatScore = openai.ResultCategoryScores{SelfHarm: 1}
			resCatApplied = openai.CategoryAppliedInputType{
				SelfHarm: []openai.ModerationItemType{openai.ModerationItemTypeText},
			}

		case strings.Contains(moderationReq.Input[i].Text, "wanna suicide"):
			resCat = openai.ResultCategories{SelfHarmIntent: true}
			resCatScore = openai.ResultCategoryScores{SelfHarm: 1}
			resCatApplied = openai.CategoryAppliedInputType{
				SelfHarmIntent: []openai.ModerationItemType{openai.ModerationItemTypeText},
			}

		case strings.Contains(moderationReq.Input[i].Text, "drink bleach"):
			resCat = openai.ResultCategories{SelfHarmInstructions: true}
			resCatScore = openai.ResultCategoryScores{SelfHarmInstructions: 1}
			resCatApplied = openai.CategoryAppliedInputType{
				SelfHarmInstructions: []openai.ModerationItemType{openai.ModerationItemTypeText},
			}

		case strings.Contains(moderationReq.Input[i].Text, "porn"):
			resCat = openai.ResultCategories{Sexual: true}
			resCatScore = openai.ResultCategoryScores{Sexual: 1}
			resCatApplied = openai.CategoryAppliedInputType{
				Sexual: []openai.ModerationItemType{openai.ModerationItemTypeText},
			}

		case strings.Contains(moderationReq.Input[i].Text, "child porn"):
			resCat = openai.ResultCategories{SexualMinors: true}
			resCatScore = openai.ResultCategoryScores{SexualMinors: 1}
			resCatApplied = openai.CategoryAppliedInputType{
				SexualMinors: []openai.ModerationItemType{openai.ModerationItemTypeText},
			}

		case strings.Contains(moderationReq.Input[i].Text, "kill"):
			resCat = openai.ResultCategories{Violence: true}
			resCatScore = openai.ResultCategoryScores{Violence: 1}
			resCatApplied = openai.CategoryAppliedInputType{
				Violence: []openai.ModerationItemType{openai.ModerationItemTypeText},
			}

		case strings.Contains(moderationReq.Input[i].Text, "corpse"):
			resCat = openai.ResultCategories{ViolenceGraphic: true}
			resCatScore = openai.ResultCategoryScores{ViolenceGraphic: 1}
			resCatApplied = openai.CategoryAppliedInputType{
				ViolenceGraphic: []openai.ModerationItemType{openai.ModerationItemTypeText},
			}

		case strings.Contains(moderationReq.Input[i].Text, "how to shoplift"):
			resCat = openai.ResultCategories{Illicit: true}
			resCatScore = openai.ResultCategoryScores{Illicit: 1}
			resCatApplied = openai.CategoryAppliedInputType{
				Illicit: []openai.ModerationItemType{openai.ModerationItemTypeText},
			}

		case strings.Contains(moderationReq.Input[i].Text, "how to buy gun"):
			resCat = openai.ResultCategories{IllicitViolent: true}
			resCatScore = openai.ResultCategoryScores{IllicitViolent: 1}
			resCatApplied = openai.CategoryAppliedInputType{
				IllicitViolent: []openai.ModerationItemType{openai.ModerationItemTypeText},
			}
		case moderationReq.Input[i].Type == openai.ModerationItemTypeImageURL &&
			moderationReq.Input[i].ImageURL.URL == "https://cdn.openai.com/API/images/harass.png":
			resCat = openai.ResultCategories{Harassment: true}
			resCatScore = openai.ResultCategoryScores{Harassment: 1}
			resCatApplied = openai.CategoryAppliedInputType{
				Harassment: []openai.ModerationItemType{openai.ModerationItemTypeImageURL},
			}
		}

		result := openai.Result{
			Categories:                resCat,
			CategoryScores:            resCatScore,
			Flagged:                   true,
			CategoryAppliedInputTypes: resCatApplied,
		}
		res.Results = append(res.Results, result)
	}

	resBytes, _ = json.Marshal(res)
	fmt.Fprintln(w, string(resBytes))
}

// getModerationBody Returns the body of the request to do a moderation.
func getModerationBody(r *http.Request) (openai.ModerationArrayRequest, error) {
	var (
		moderation             = openai.ModerationRequest{}
		strArrayInput          = openai.ModerationStrArrayRequest{}
		moderationArrayRequest = openai.ModerationArrayRequest{}
	)
	// read the request body
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		return openai.ModerationArrayRequest{}, err
	}
	err = json.Unmarshal(reqBody, &moderation)
	if err == nil {
		return openai.ModerationArrayRequest{
			Input: []openai.ModerationRequestItem{
				{
					Type: openai.ModerationItemTypeText,
					Text: moderation.Input,
				},
			},
			Model: "",
		}, nil
	}
	err = json.Unmarshal(reqBody, &strArrayInput)
	if err == nil {
		moderationArrayRequest.Model = strArrayInput.Model
		for i := range strArrayInput.Input {
			moderationArrayRequest.Input = append(moderationArrayRequest.Input, openai.ModerationRequestItem{
				Type: openai.ModerationItemTypeText,
				Text: strArrayInput.Input[i],
			})
		}
		return moderationArrayRequest, nil
	}
	err = json.Unmarshal(reqBody, &moderationArrayRequest)
	if err != nil {
		return openai.ModerationArrayRequest{}, err
	}

	return moderationArrayRequest, nil
}
