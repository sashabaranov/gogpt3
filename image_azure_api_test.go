package openai_test

import (
	"bytes"
	"strings"

	. "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test/checks"

	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestAzureImages(t *testing.T) {
	client, server, teardown := setupAzureTestServer()
	defer teardown()
	server.RegisterHandler("/openai/images/generations:submit", handleAzureImageEndpoint)
	server.RegisterHandler("/openai/operations/images/request-id", handleImageCallbackEndpoint)

	_, err := client.CreateAzureImage(context.Background(), ImageRequest{
		Prompt:         "Lorem ipsum",
		ResponseFormat: CreateImageResponseFormatURL,
		N:              2,
	})
	checks.NoError(t, err, "Azure CreateImage error")
}

// handleImageEndpoint Handles the images endpoint by the test server.
func handleAzureImageEndpoint(w http.ResponseWriter, r *http.Request) {
	var err error
	var resBytes []byte

	// imagess only accepts POST requests
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
	// Azure Image Generation request - respond with callback Header only & HTTP accepted status.
	if strings.Contains(r.RequestURI, "/openai/images/generations:submit") {
		w.Header().Add("Operation-Location", "http://"+r.Host+"/openai/operations/images/request-id")
		w.WriteHeader(http.StatusAccepted)
		return
	}
	var imageReq ImageRequest
	if imageReq, err = getImageBody(r); err != nil {
		http.Error(w, "could not read request", http.StatusInternalServerError)
		return
	}
	res := ImageResponse{
		Created: time.Now().Unix(),
	}
	for i := 0; i < imageReq.N; i++ {
		imageData := ImageResponseDataInner{}
		switch imageReq.ResponseFormat {
		case CreateImageResponseFormatURL, "":
			imageData.URL = "https://example.com/image.png"
		case CreateImageResponseFormatB64JSON:
			// This decodes to "{}" in base64.
			imageData.B64JSON = "e30K"
		default:
			http.Error(w, "invalid response format", http.StatusBadRequest)
			return
		}
		res.Data = append(res.Data, imageData)
	}
	resBytes, _ = json.Marshal(res)
	fmt.Fprintln(w, string(resBytes))
}

// handleImageCallbackEndpoint Handles the callback endpoint by the test server.
func handleImageCallbackEndpoint(w http.ResponseWriter, r *http.Request) {
	var err error

	// image callback only accepts GET requests
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Set the status to succeeded if this is a retry request.
	status := "running"
	if r.Header.Get("Retry") == "true" {
		status = "succeeded"
	}

	cbResponse := CallBackResponse{
		Created: time.Now().Unix(),
		Status:  status,
		Result: CBResult{
			Data: CBData{
				{URL: "http://example.com/image1"},
				{URL: "http://example.com/image2"},
			},
		},
	}
	cbResponseBytes := new(bytes.Buffer)
	err = json.NewEncoder(cbResponseBytes).Encode(cbResponse)
	if err != nil {
		http.Error(w, "could not write repsonse", http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, cbResponseBytes.String())
}
