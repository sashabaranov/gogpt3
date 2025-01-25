package openai_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/alejandrojnm/go-openai"
	"github.com/alejandrojnm/go-openai/internal/test/checks"
)

// TestEdits Tests the edits endpoint of the API using the mocked server.
func TestEdits(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/edits", handleEditEndpoint)
	// create an edit request
	model := "ada"
	editReq := openai.EditsRequest{
		Model: &model,
		Input: "Lorem ipsum dolor sit amet, consectetur adipiscing elit, " +
			"sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim" +
			" ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip" +
			" ex ea commodo consequat. Duis aute irure dolor in reprehe",
		Instruction: "test instruction",
		N:           3,
	}
	response, err := client.Edits(context.Background(), editReq)
	checks.NoError(t, err, "Edits error")
	if len(response.Choices) != editReq.N {
		t.Fatalf("edits does not properly return the correct number of choices")
	}
}

// handleEditEndpoint Handles the edit endpoint by the test server.
func handleEditEndpoint(w http.ResponseWriter, r *http.Request) {
	var err error
	var resBytes []byte

	// edits only accepts POST requests
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
	var editReq openai.EditsRequest
	editReq, err = getEditBody(r)
	if err != nil {
		http.Error(w, "could not read request", http.StatusInternalServerError)
		return
	}
	// create a response
	res := openai.EditsResponse{
		Object:  "test-object",
		Created: time.Now().Unix(),
	}
	// edit and calculate token usage
	editString := "edited by mocked OpenAI server :)"
	inputTokens := numTokens(editReq.Input+editReq.Instruction) * editReq.N
	completionTokens := int(float32(len(editString))/4) * editReq.N
	for i := 0; i < editReq.N; i++ {
		// instruction will be hidden and only seen by OpenAI
		res.Choices = append(res.Choices, openai.EditsChoice{
			Text:  editReq.Input + editString,
			Index: i,
		})
	}
	res.Usage = openai.Usage{
		PromptTokens:     inputTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      inputTokens + completionTokens,
	}
	resBytes, _ = json.Marshal(res)
	fmt.Fprint(w, string(resBytes))
}

// getEditBody Returns the body of the request to create an edit.
func getEditBody(r *http.Request) (openai.EditsRequest, error) {
	edit := openai.EditsRequest{}
	// read the request body
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		return openai.EditsRequest{}, err
	}
	err = json.Unmarshal(reqBody, &edit)
	if err != nil {
		return openai.EditsRequest{}, err
	}
	return edit, nil
}
