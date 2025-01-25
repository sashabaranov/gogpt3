package openai_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alejandrojnm/go-openai"
	"github.com/alejandrojnm/go-openai/internal/test"
	"github.com/alejandrojnm/go-openai/internal/test/checks"
)

// TestAudio Tests the transcription and translation endpoints of the API using the mocked server.
func TestAudio(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/audio/transcriptions", handleAudioEndpoint)
	server.RegisterHandler("/v1/audio/translations", handleAudioEndpoint)

	testcases := []struct {
		name     string
		createFn func(context.Context, openai.AudioRequest) (openai.AudioResponse, error)
	}{
		{
			"transcribe",
			client.CreateTranscription,
		},
		{
			"translate",
			client.CreateTranslation,
		},
	}

	ctx := context.Background()

	dir, cleanup := test.CreateTestDirectory(t)
	defer cleanup()

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(dir, "fake.mp3")
			test.CreateTestFile(t, path)

			req := openai.AudioRequest{
				FilePath: path,
				Model:    "whisper-3",
			}
			_, err := tc.createFn(ctx, req)
			checks.NoError(t, err, "audio API error")
		})

		t.Run(tc.name+" (with reader)", func(t *testing.T) {
			req := openai.AudioRequest{
				FilePath: "fake.webm",
				Reader:   bytes.NewBuffer([]byte(`some webm binary data`)),
				Model:    "whisper-3",
			}
			_, err := tc.createFn(ctx, req)
			checks.NoError(t, err, "audio API error")
		})
	}
}

func TestAudioWithOptionalArgs(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()
	server.RegisterHandler("/v1/audio/transcriptions", handleAudioEndpoint)
	server.RegisterHandler("/v1/audio/translations", handleAudioEndpoint)

	testcases := []struct {
		name     string
		createFn func(context.Context, openai.AudioRequest) (openai.AudioResponse, error)
	}{
		{
			"transcribe",
			client.CreateTranscription,
		},
		{
			"translate",
			client.CreateTranslation,
		},
	}

	ctx := context.Background()

	dir, cleanup := test.CreateTestDirectory(t)
	defer cleanup()

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(dir, "fake.mp3")
			test.CreateTestFile(t, path)

			req := openai.AudioRequest{
				FilePath:    path,
				Model:       "whisper-3",
				Prompt:      "用简体中文",
				Temperature: 0.5,
				Language:    "zh",
				Format:      openai.AudioResponseFormatSRT,
				TimestampGranularities: []openai.TranscriptionTimestampGranularity{
					openai.TranscriptionTimestampGranularitySegment,
					openai.TranscriptionTimestampGranularityWord,
				},
			}
			_, err := tc.createFn(ctx, req)
			checks.NoError(t, err, "audio API error")
		})
	}
}

// handleAudioEndpoint Handles the completion endpoint by the test server.
func handleAudioEndpoint(w http.ResponseWriter, r *http.Request) {
	var err error

	// audio endpoints only accept POST requests
	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}

	mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		http.Error(w, "failed to parse media type", http.StatusBadRequest)
		return
	}

	if !strings.HasPrefix(mediaType, "multipart") {
		http.Error(w, "request is not multipart", http.StatusBadRequest)
	}

	boundary, ok := params["boundary"]
	if !ok {
		http.Error(w, "no boundary in params", http.StatusBadRequest)
		return
	}

	fileData := &bytes.Buffer{}
	mr := multipart.NewReader(r.Body, boundary)
	part, err := mr.NextPart()
	if err != nil && errors.Is(err, io.EOF) {
		http.Error(w, "error accessing file", http.StatusBadRequest)
		return
	}
	if _, err = io.Copy(fileData, part); err != nil {
		http.Error(w, "failed to copy file", http.StatusInternalServerError)
		return
	}

	if len(fileData.Bytes()) == 0 {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, "received empty file data", http.StatusBadRequest)
		return
	}

	if _, err = w.Write([]byte(`{"body": "hello"}`)); err != nil {
		http.Error(w, "failed to write body", http.StatusInternalServerError)
		return
	}
}
