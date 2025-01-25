package openai_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/alejandrojnm/go-openai"
	"github.com/alejandrojnm/go-openai/internal/test"
	"github.com/alejandrojnm/go-openai/internal/test/checks"
)

func TestSpeechIntegration(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()

	server.RegisterHandler("/v1/audio/speech", func(w http.ResponseWriter, r *http.Request) {
		dir, cleanup := test.CreateTestDirectory(t)
		path := filepath.Join(dir, "fake.mp3")
		test.CreateTestFile(t, path)
		defer cleanup()

		// audio endpoints only accept POST requests
		if r.Method != "POST" {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil {
			http.Error(w, "failed to parse media type", http.StatusBadRequest)
			return
		}

		if mediaType != "application/json" {
			http.Error(w, "request is not json", http.StatusBadRequest)
			return
		}

		// Parse the JSON body of the request
		var params map[string]interface{}
		err = json.NewDecoder(r.Body).Decode(&params)
		if err != nil {
			http.Error(w, "failed to parse request body", http.StatusBadRequest)
			return
		}

		// Check if each required field is present in the parsed JSON object
		reqParams := []string{"model", "input", "voice"}
		for _, param := range reqParams {
			_, ok := params[param]
			if !ok {
				http.Error(w, fmt.Sprintf("no %s in params", param), http.StatusBadRequest)
				return
			}
		}

		// read audio file content
		audioFile, err := os.ReadFile(path)
		if err != nil {
			http.Error(w, "failed to read audio file", http.StatusInternalServerError)
			return
		}

		// write audio file content to response
		w.Header().Set("Content-Type", "audio/mpeg")
		w.Header().Set("Transfer-Encoding", "chunked")
		w.Header().Set("Connection", "keep-alive")
		_, err = w.Write(audioFile)
		if err != nil {
			http.Error(w, "failed to write body", http.StatusInternalServerError)
			return
		}
	})

	t.Run("happy path", func(t *testing.T) {
		res, err := client.CreateSpeech(context.Background(), openai.CreateSpeechRequest{
			Model: openai.TTSModel1,
			Input: "Hello!",
			Voice: openai.VoiceAlloy,
		})
		checks.NoError(t, err, "CreateSpeech error")
		defer res.Close()

		buf, err := io.ReadAll(res)
		checks.NoError(t, err, "ReadAll error")

		// save buf to file as mp3
		err = os.WriteFile("test.mp3", buf, 0644)
		checks.NoError(t, err, "Create error")
	})
}
