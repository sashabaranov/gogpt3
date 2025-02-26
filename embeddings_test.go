package openai_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"reflect"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/internal/test/checks"
)

func TestEmbedding(t *testing.T) {
	embeddedModels := []openai.EmbeddingModel{
		openai.AdaSimilarity,
		openai.BabbageSimilarity,
		openai.CurieSimilarity,
		openai.DavinciSimilarity,
		openai.AdaSearchDocument,
		openai.AdaSearchQuery,
		openai.BabbageSearchDocument,
		openai.BabbageSearchQuery,
		openai.CurieSearchDocument,
		openai.CurieSearchQuery,
		openai.DavinciSearchDocument,
		openai.DavinciSearchQuery,
		openai.AdaCodeSearchCode,
		openai.AdaCodeSearchText,
		openai.BabbageCodeSearchCode,
		openai.BabbageCodeSearchText,
	}
	for _, model := range embeddedModels {
		// test embedding request with strings (simple embedding request)
		embeddingReq := openai.EmbeddingRequest{
			Input: []string{
				"The food was delicious and the waiter",
				"Other examples of embedding request",
			},
			Model: model,
		}
		// marshal embeddingReq to JSON and confirm that the model field equals
		// the AdaSearchQuery type
		marshaled, err := json.Marshal(embeddingReq)
		checks.NoError(t, err, "Could not marshal embedding request")
		if !bytes.Contains(marshaled, []byte(`"model":"`+model+`"`)) {
			t.Fatalf("Expected embedding request to contain model field")
		}

		// test embedding request with strings and extra_body param
		embeddingReqWithExtraBody := openai.EmbeddingRequest{
			Input: []string{
				"The food was delicious and the waiter",
				"Other examples of embedding request",
			},
			Model: model,
			ExtraBody: map[string]any{
				"input_type": "query",
				"truncate":   "NONE",
			},
		}
		marshaled, err = json.Marshal(embeddingReqWithExtraBody)
		checks.NoError(t, err, "Could not marshal embedding request")
		if !bytes.Contains(marshaled, []byte(`"model":"`+model+`"`)) {
			t.Fatalf("Expected embedding request to contain model field")
		}

		// test embedding request with strings
		embeddingReqStrings := openai.EmbeddingRequestStrings{
			Input: []string{
				"The food was delicious and the waiter",
				"Other examples of embedding request",
			},
			Model: model,
		}
		marshaled, err = json.Marshal(embeddingReqStrings)
		checks.NoError(t, err, "Could not marshal embedding request")
		if !bytes.Contains(marshaled, []byte(`"model":"`+model+`"`)) {
			t.Fatalf("Expected embedding request to contain model field")
		}

		// test embedding request with tokens
		embeddingReqTokens := openai.EmbeddingRequestTokens{
			Input: [][]int{
				{464, 2057, 373, 12625, 290, 262, 46612},
				{6395, 6096, 286, 11525, 12083, 2581},
			},
			Model: model,
		}
		marshaled, err = json.Marshal(embeddingReqTokens)
		checks.NoError(t, err, "Could not marshal embedding request")
		if !bytes.Contains(marshaled, []byte(`"model":"`+model+`"`)) {
			t.Fatalf("Expected embedding request to contain model field")
		}
	}
}

func TestEmbeddingEndpoint(t *testing.T) {
	client, server, teardown := setupOpenAITestServer()
	defer teardown()

	sampleEmbeddings := []openai.Embedding{
		{Embedding: []float32{1.23, 4.56, 7.89}},
		{Embedding: []float32{-0.006968617, -0.0052718227, 0.011901081}},
	}

	sampleBase64Embeddings := []openai.Base64Embedding{
		{Embedding: "pHCdP4XrkUDhevxA"},
		{Embedding: "/1jku0G/rLvA/EI8"},
	}

	server.RegisterHandler(
		"/v1/embeddings",
		func(w http.ResponseWriter, r *http.Request) {
			var req struct {
				EncodingFormat openai.EmbeddingEncodingFormat `json:"encoding_format"`
				User           string                         `json:"user"`
			}
			_ = json.NewDecoder(r.Body).Decode(&req)

			var resBytes []byte
			switch {
			case req.User == "invalid":
				w.WriteHeader(http.StatusBadRequest)
				return
			case req.EncodingFormat == openai.EmbeddingEncodingFormatBase64:
				resBytes, _ = json.Marshal(openai.EmbeddingResponseBase64{Data: sampleBase64Embeddings})
			default:
				resBytes, _ = json.Marshal(openai.EmbeddingResponse{Data: sampleEmbeddings})
			}
			fmt.Fprintln(w, string(resBytes))
		},
	)
	// test create embeddings with strings (simple embedding request)
	res, err := client.CreateEmbeddings(context.Background(), openai.EmbeddingRequest{})
	checks.NoError(t, err, "CreateEmbeddings error")
	if !reflect.DeepEqual(res.Data, sampleEmbeddings) {
		t.Errorf("Expected %#v embeddings, got %#v", sampleEmbeddings, res.Data)
	}

	// test create embeddings with strings (ExtraBody in request)
	res, err = client.CreateEmbeddings(
		context.Background(),
		openai.EmbeddingRequest{
			ExtraBody: map[string]any{
				"input_type": "query",
				"truncate":   "NONE",
			},
			Dimensions: 1,
		},
	)
	checks.NoError(t, err, "CreateEmbeddings error")
	if !reflect.DeepEqual(res.Data, sampleEmbeddings) {
		t.Errorf("Expected %#v embeddings, got %#v", sampleEmbeddings, res.Data)
	}

	// test create embeddings with strings (ExtraBody in request and )
	_, err = client.CreateEmbeddings(
		context.Background(),
		openai.EmbeddingRequest{
			Input: make(chan int), // Channels are not serializable
			Model: "example_model",
		},
	)
	checks.HasError(t, err, "CreateEmbeddings error")

	// test failed (Serialize JSON error)
	res, err = client.CreateEmbeddings(
		context.Background(),
		openai.EmbeddingRequest{
			EncodingFormat: openai.EmbeddingEncodingFormatBase64,
		},
	)
	checks.NoError(t, err, "CreateEmbeddings error")
	if !reflect.DeepEqual(res.Data, sampleEmbeddings) {
		t.Errorf("Expected %#v embeddings, got %#v", sampleEmbeddings, res.Data)
	}

	// test create embeddings with strings
	res, err = client.CreateEmbeddings(context.Background(), openai.EmbeddingRequestStrings{})
	checks.NoError(t, err, "CreateEmbeddings strings error")
	if !reflect.DeepEqual(res.Data, sampleEmbeddings) {
		t.Errorf("Expected %#v embeddings, got %#v", sampleEmbeddings, res.Data)
	}

	// test create embeddings with tokens
	res, err = client.CreateEmbeddings(context.Background(), openai.EmbeddingRequestTokens{})
	checks.NoError(t, err, "CreateEmbeddings tokens error")
	if !reflect.DeepEqual(res.Data, sampleEmbeddings) {
		t.Errorf("Expected %#v embeddings, got %#v", sampleEmbeddings, res.Data)
	}

	// test failed sendRequest
	_, err = client.CreateEmbeddings(context.Background(), openai.EmbeddingRequest{
		User:           "invalid",
		EncodingFormat: openai.EmbeddingEncodingFormatBase64,
	})
	checks.HasError(t, err, "CreateEmbeddings error")
}

func TestAzureEmbeddingEndpoint(t *testing.T) {
	client, server, teardown := setupAzureTestServer()
	defer teardown()

	sampleEmbeddings := []openai.Embedding{
		{Embedding: []float32{1.23, 4.56, 7.89}},
		{Embedding: []float32{-0.006968617, -0.0052718227, 0.011901081}},
	}

	server.RegisterHandler(
		"/openai/deployments/text-embedding-ada-002/embeddings",
		func(w http.ResponseWriter, _ *http.Request) {
			resBytes, _ := json.Marshal(openai.EmbeddingResponse{Data: sampleEmbeddings})
			fmt.Fprintln(w, string(resBytes))
		},
	)
	// test create embeddings with strings (simple embedding request)
	res, err := client.CreateEmbeddings(context.Background(), openai.EmbeddingRequest{
		Model: openai.AdaEmbeddingV2,
	})
	checks.NoError(t, err, "CreateEmbeddings error")
	if !reflect.DeepEqual(res.Data, sampleEmbeddings) {
		t.Errorf("Expected %#v embeddings, got %#v", sampleEmbeddings, res.Data)
	}
}

func TestEmbeddingResponseBase64_ToEmbeddingResponse(t *testing.T) {
	type fields struct {
		Object string
		Data   []openai.Base64Embedding
		Model  openai.EmbeddingModel
		Usage  openai.Usage
	}
	tests := []struct {
		name    string
		fields  fields
		want    openai.EmbeddingResponse
		wantErr bool
	}{
		{
			name: "test embedding response base64 to embedding response",
			fields: fields{
				Data: []openai.Base64Embedding{
					{Embedding: "pHCdP4XrkUDhevxA"},
					{Embedding: "/1jku0G/rLvA/EI8"},
				},
			},
			want: openai.EmbeddingResponse{
				Data: []openai.Embedding{
					{Embedding: []float32{1.23, 4.56, 7.89}},
					{Embedding: []float32{-0.006968617, -0.0052718227, 0.011901081}},
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid embedding",
			fields: fields{
				Data: []openai.Base64Embedding{
					{
						Embedding: "----",
					},
				},
			},
			want:    openai.EmbeddingResponse{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &openai.EmbeddingResponseBase64{
				Object: tt.fields.Object,
				Data:   tt.fields.Data,
				Model:  tt.fields.Model,
				Usage:  tt.fields.Usage,
			}
			got, err := r.ToEmbeddingResponse()
			if (err != nil) != tt.wantErr {
				t.Errorf("EmbeddingResponseBase64.ToEmbeddingResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EmbeddingResponseBase64.ToEmbeddingResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDotProduct(t *testing.T) {
	v1 := &openai.Embedding{Embedding: []float32{1, 2, 3}}
	v2 := &openai.Embedding{Embedding: []float32{2, 4, 6}}
	expected := float32(28.0)

	result, err := v1.DotProduct(v2)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if math.Abs(float64(result-expected)) > 1e-12 {
		t.Errorf("Unexpected result. Expected: %v, but got %v", expected, result)
	}

	v1 = &openai.Embedding{Embedding: []float32{1, 0, 0}}
	v2 = &openai.Embedding{Embedding: []float32{0, 1, 0}}
	expected = float32(0.0)

	result, err = v1.DotProduct(v2)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if math.Abs(float64(result-expected)) > 1e-12 {
		t.Errorf("Unexpected result. Expected: %v, but got %v", expected, result)
	}

	// Test for VectorLengthMismatchError
	v1 = &openai.Embedding{Embedding: []float32{1, 0, 0}}
	v2 = &openai.Embedding{Embedding: []float32{0, 1}}
	_, err = v1.DotProduct(v2)
	if !errors.Is(err, openai.ErrVectorLengthMismatch) {
		t.Errorf("Expected Vector Length Mismatch Error, but got: %v", err)
	}
}
