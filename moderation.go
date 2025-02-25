package openai

import (
	"context"
	"errors"
	"net/http"
)

// The moderation endpoint is a tool you can use to check whether content complies with OpenAI's usage policies.
// Developers can thus identify content that our usage policies prohibits and take action, for instance by filtering it.

// The default is text-moderation-latest which will be automatically upgraded over time.
// This ensures you are always using our most accurate model.
// If you use text-moderation-stable, we will provide advanced notice before updating the model.
// Accuracy of text-moderation-stable may be slightly lower than for text-moderation-latest.
const (
	ModerationOmniLatest   = "omni-moderation-latest"
	ModerationOmni20240926 = "omni-moderation-2024-09-26"
	ModerationTextStable   = "text-moderation-stable"
	ModerationTextLatest   = "text-moderation-latest"
	// Deprecated: use ModerationTextStable and ModerationTextLatest instead.
	ModerationText001 = "text-moderation-001"
)

var (
	ErrModerationInvalidModel = errors.New("this model is not supported with moderation, please use text-moderation-stable or text-moderation-latest instead") //nolint:lll
)

type ModerationItemType string

const (
	ModerationItemTypeText     ModerationItemType = "text"
	ModerationItemTypeImageURL ModerationItemType = "image_url"
)

var validModerationModel = map[string]struct{}{
	ModerationOmniLatest:   {},
	ModerationOmni20240926: {},
	ModerationTextStable:   {},
	ModerationTextLatest:   {},
}

// ModerationRequest represents a request structure for moderation API.
type ModerationRequest struct {
	Input string `json:"input,omitempty"`
	Model string `json:"model,omitempty"`
}

func (m ModerationRequest) Convert() ModerationRequestV2 {
	return ModerationRequestV2{
		Input: m.Input,
		Model: m.Model,
	}
}

type ModerationStrArrayRequest struct {
	Input []string `json:"input,omitempty"`
	Model string   `json:"model,omitempty"`
}

func (m ModerationStrArrayRequest) Convert() ModerationRequestV2 {
	return ModerationRequestV2{
		Input: m.Input,
		Model: m.Model,
	}
}

type ModerationArrayRequest struct {
	Input []ModerationRequestItem `json:"input,omitempty"`
	Model string                  `json:"model,omitempty"`
}

func (m ModerationArrayRequest) Convert() ModerationRequestV2 {
	return ModerationRequestV2{
		Input: m.Input,
		Model: m.Model,
	}
}

type ModerationRequestItem struct {
	Type ModerationItemType `json:"type"`

	ImageURL ModerationImageURL `json:"image_url,omitempty"`
	Text     string             `json:"text,omitempty"`
}

type ModerationImageURL struct {
	URL string `json:"url,omitempty"`
}

type ModerationRequestV2 struct {
	Input any    `json:"input,omitempty"`
	Model string `json:"model,omitempty"`
}

type ModerationRequestConverter interface {
	Convert() ModerationRequestV2
}

// Result represents one of possible moderation results.
type Result struct {
	Categories                ResultCategories         `json:"categories"`
	CategoryScores            ResultCategoryScores     `json:"category_scores"`
	Flagged                   bool                     `json:"flagged"`
	CategoryAppliedInputTypes CategoryAppliedInputType `json:"category_applied_input_types"`
}

// ResultCategories represents Categories of Result.
type ResultCategories struct {
	Hate                  bool `json:"hate"`
	HateThreatening       bool `json:"hate/threatening"`
	Harassment            bool `json:"harassment"`
	HarassmentThreatening bool `json:"harassment/threatening"`
	SelfHarm              bool `json:"self-harm"`
	SelfHarmIntent        bool `json:"self-harm/intent"`
	SelfHarmInstructions  bool `json:"self-harm/instructions"`
	Sexual                bool `json:"sexual"`
	SexualMinors          bool `json:"sexual/minors"`
	Violence              bool `json:"violence"`
	ViolenceGraphic       bool `json:"violence/graphic"`
	Illicit               bool `json:"illicit"`
	IllicitViolent        bool `json:"illicit/violent"`
}

// ResultCategoryScores represents CategoryScores of Result.
type ResultCategoryScores struct {
	Hate                  float32 `json:"hate"`
	HateThreatening       float32 `json:"hate/threatening"`
	Harassment            float32 `json:"harassment"`
	HarassmentThreatening float32 `json:"harassment/threatening"`
	SelfHarm              float32 `json:"self-harm"`
	SelfHarmIntent        float32 `json:"self-harm/intent"`
	SelfHarmInstructions  float32 `json:"self-harm/instructions"`
	Sexual                float32 `json:"sexual"`
	SexualMinors          float32 `json:"sexual/minors"`
	Violence              float32 `json:"violence"`
	ViolenceGraphic       float32 `json:"violence/graphic"`
	Illicit               float32 `json:"illicit"`
	IllicitViolent        float32 `json:"illicit/violent"`
}

type CategoryAppliedInputType struct {
	Harassment            []ModerationItemType `json:"harassment"`
	HarassmentThreatening []ModerationItemType `json:"harassment/threatening"`
	Sexual                []ModerationItemType `json:"sexual"`
	Hate                  []ModerationItemType `json:"hate"`
	HateThreatening       []ModerationItemType `json:"hate/threatening"`
	Illicit               []ModerationItemType `json:"illicit"`
	IllicitViolent        []ModerationItemType `json:"illicit/violent"`
	SelfHarmIntent        []ModerationItemType `json:"self-harm/intent"`
	SelfHarmInstructions  []ModerationItemType `json:"self-harm/instructions"`
	SelfHarm              []ModerationItemType `json:"self-harm"`
	SexualMinors          []ModerationItemType `json:"sexual/minors"`
	Violence              []ModerationItemType `json:"violence"`
	ViolenceGraphic       []ModerationItemType `json:"violence/graphic"`
}

// ModerationResponse represents a response structure for moderation API.
type ModerationResponse struct {
	ID      string   `json:"id"`
	Model   string   `json:"model"`
	Results []Result `json:"results"`

	httpHeader
}

// Moderations — perform a moderation api call over a string.
// Input can be an array or slice but a string will reduce the complexity.
func (c *Client) Moderations(ctx context.Context,
	request ModerationRequestConverter) (response ModerationResponse, err error) {
	realRequest := request.Convert()

	if _, ok := validModerationModel[realRequest.Model]; len(realRequest.Model) > 0 && !ok {
		err = ErrModerationInvalidModel
		return
	}
	req, err := c.newRequest(
		ctx,
		http.MethodPost,
		c.fullURL("/moderations", withModel(realRequest.Model)),
		withBody(&request),
	)
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}
