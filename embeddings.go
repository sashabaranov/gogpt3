package openai

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"math"
	"net/http"
)

var ErrVectorLengthMismatch = errors.New("vector length mismatch")

// EmbeddingModel enumerates the models which can be used
// to generate Embedding vectors.
type EmbeddingModel string

const (
	// Deprecated: The following block is shut down. Use text-embedding-ada-002 instead.
	AdaSimilarity         EmbeddingModel = "text-similarity-ada-001"
	BabbageSimilarity     EmbeddingModel = "text-similarity-babbage-001"
	CurieSimilarity       EmbeddingModel = "text-similarity-curie-001"
	DavinciSimilarity     EmbeddingModel = "text-similarity-davinci-001"
	AdaSearchDocument     EmbeddingModel = "text-search-ada-doc-001"
	AdaSearchQuery        EmbeddingModel = "text-search-ada-query-001"
	BabbageSearchDocument EmbeddingModel = "text-search-babbage-doc-001"
	BabbageSearchQuery    EmbeddingModel = "text-search-babbage-query-001"
	CurieSearchDocument   EmbeddingModel = "text-search-curie-doc-001"
	CurieSearchQuery      EmbeddingModel = "text-search-curie-query-001"
	DavinciSearchDocument EmbeddingModel = "text-search-davinci-doc-001"
	DavinciSearchQuery    EmbeddingModel = "text-search-davinci-query-001"
	AdaCodeSearchCode     EmbeddingModel = "code-search-ada-code-001"
	AdaCodeSearchText     EmbeddingModel = "code-search-ada-text-001"
	BabbageCodeSearchCode EmbeddingModel = "code-search-babbage-code-001"
	BabbageCodeSearchText EmbeddingModel = "code-search-babbage-text-001"

	AdaEmbeddingV2  EmbeddingModel = "text-embedding-ada-002"
	SmallEmbedding3 EmbeddingModel = "text-embedding-3-small"
	LargeEmbedding3 EmbeddingModel = "text-embedding-3-large"
)

// Embedding is a special format of data representation that can be easily utilized by machine
// learning models and algorithms. The embedding is an information dense representation of the
// semantic meaning of a piece of text. Each embedding is a vector of floating point numbers,
// such that the distance between two embeddings in the vector space is correlated with semantic similarity
// between two inputs in the original format. For example, if two texts are similar,
// then their vector representations should also be similar.
type Embedding struct {
	Object    string    `json:"object"`
	Embedding []float32 `json:"embedding"`
	Index     int       `json:"index"`
}

// DotProduct calculates the dot product of the embedding vector with another
// embedding vector. Both vectors must have the same length; otherwise, an
// ErrVectorLengthMismatch is returned. The method returns the calculated dot
// product as a float32 value.
func (e *Embedding) DotProduct(other *Embedding) (float32, error) {
	if len(e.Embedding) != len(other.Embedding) {
		return 0, ErrVectorLengthMismatch
	}

	var dotProduct float32
	for i := range e.Embedding {
		dotProduct += e.Embedding[i] * other.Embedding[i]
	}

	return dotProduct, nil
}

// EmbeddingResponse is the response from a Create embeddings request.
type EmbeddingResponse struct {
	Object string         `json:"object"`
	Data   []Embedding    `json:"data"`
	Model  EmbeddingModel `json:"model"`
	Usage  Usage          `json:"usage"`

	httpHeader
}

type base64String string

func (b base64String) Decode() ([]float32, error) {
	decodedData, err := base64.StdEncoding.DecodeString(string(b))
	if err != nil {
		return nil, err
	}

	const sizeOfFloat32 = 4
	floats := make([]float32, len(decodedData)/sizeOfFloat32)
	for i := 0; i < len(floats); i++ {
		floats[i] = math.Float32frombits(binary.LittleEndian.Uint32(decodedData[i*4 : (i+1)*4]))
	}

	return floats, nil
}

// Base64Embedding is a container for base64 encoded embeddings.
type Base64Embedding struct {
	Object    string       `json:"object"`
	Embedding base64String `json:"embedding"`
	Index     int          `json:"index"`
}

// EmbeddingResponseBase64 is the response from a Create embeddings request with base64 encoding format.
type EmbeddingResponseBase64 struct {
	Object string            `json:"object"`
	Data   []Base64Embedding `json:"data"`
	Model  EmbeddingModel    `json:"model"`
	Usage  Usage             `json:"usage"`

	httpHeader
}

// ToEmbeddingResponse converts an embeddingResponseBase64 to an EmbeddingResponse.
func (r *EmbeddingResponseBase64) ToEmbeddingResponse() (EmbeddingResponse, error) {
	data := make([]Embedding, len(r.Data))

	for i, base64Embedding := range r.Data {
		embedding, err := base64Embedding.Embedding.Decode()
		if err != nil {
			return EmbeddingResponse{}, err
		}

		data[i] = Embedding{
			Object:    base64Embedding.Object,
			Embedding: embedding,
			Index:     base64Embedding.Index,
		}
	}

	return EmbeddingResponse{
		Object: r.Object,
		Model:  r.Model,
		Data:   data,
		Usage:  r.Usage,
	}, nil
}

type EmbeddingRequestConverter interface {
	// Needs to be of type EmbeddingRequestStrings or EmbeddingRequestTokens
	Convert() EmbeddingRequest
}

// EmbeddingEncodingFormat is the format of the embeddings data.
// Currently, only "float" and "base64" are supported, however, "base64" is not officially documented.
// If not specified OpenAI will use "float".
type EmbeddingEncodingFormat string

const (
	EmbeddingEncodingFormatFloat  EmbeddingEncodingFormat = "float"
	EmbeddingEncodingFormatBase64 EmbeddingEncodingFormat = "base64"
)

type EmbeddingRequest struct {
	Input          any                     `json:"input"`
	Model          EmbeddingModel          `json:"model"`
	User           string                  `json:"user,omitempty"`
	EncodingFormat EmbeddingEncodingFormat `json:"encoding_format,omitempty"`
	// Dimensions The number of dimensions the resulting output embeddings should have.
	// Only supported in text-embedding-3 and later models.
	Dimensions int `json:"dimensions,omitempty"`
	// The ExtraBody field allows for the inclusion of arbitrary key-value pairs
	// in the request body that may not be explicitly defined in this struct.
	ExtraBody map[string]any `json:"extra_body,omitempty"`
}

func (r EmbeddingRequest) Convert() EmbeddingRequest {
	return r
}

// EmbeddingRequestStrings is the input to a create embeddings request with a slice of strings.
type EmbeddingRequestStrings struct {
	// Input is a slice of strings for which you want to generate an Embedding vector.
	// Each input must not exceed 8192 tokens in length.
	// OpenAPI suggests replacing newlines (\n) in your input with a single space, as they
	// have observed inferior results when newlines are present.
	// E.g.
	//	"The food was delicious and the waiter..."
	Input []string `json:"input"`
	// ID of the model to use. You can use the List models API to see all of your available models,
	// or see our Model overview for descriptions of them.
	Model EmbeddingModel `json:"model"`
	// A unique identifier representing your end-user, which will help OpenAI to monitor and detect abuse.
	User string `json:"user"`
	// EmbeddingEncodingFormat is the format of the embeddings data.
	// Currently, only "float" and "base64" are supported, however, "base64" is not officially documented.
	// If not specified OpenAI will use "float".
	EncodingFormat EmbeddingEncodingFormat `json:"encoding_format,omitempty"`
	// Dimensions The number of dimensions the resulting output embeddings should have.
	// Only supported in text-embedding-3 and later models.
	Dimensions int `json:"dimensions,omitempty"`
	// The ExtraBody field allows for the inclusion of arbitrary key-value pairs
	// in the request body that may not be explicitly defined in this struct.
	ExtraBody map[string]any `json:"extra_body,omitempty"`
}

func (r EmbeddingRequestStrings) Convert() EmbeddingRequest {
	return EmbeddingRequest{
		Input:          r.Input,
		Model:          r.Model,
		User:           r.User,
		EncodingFormat: r.EncodingFormat,
		Dimensions:     r.Dimensions,
		ExtraBody:      r.ExtraBody,
	}
}

type EmbeddingRequestTokens struct {
	// Input is a slice of slices of ints ([][]int) for which you want to generate an Embedding vector.
	// Each input must not exceed 8192 tokens in length.
	// OpenAPI suggests replacing newlines (\n) in your input with a single space, as they
	// have observed inferior results when newlines are present.
	// E.g.
	//	"The food was delicious and the waiter..."
	Input [][]int `json:"input"`
	// ID of the model to use. You can use the List models API to see all of your available models,
	// or see our Model overview for descriptions of them.
	Model EmbeddingModel `json:"model"`
	// A unique identifier representing your end-user, which will help OpenAI to monitor and detect abuse.
	User string `json:"user"`
	// EmbeddingEncodingFormat is the format of the embeddings data.
	// Currently, only "float" and "base64" are supported, however, "base64" is not officially documented.
	// If not specified OpenAI will use "float".
	EncodingFormat EmbeddingEncodingFormat `json:"encoding_format,omitempty"`
	// Dimensions The number of dimensions the resulting output embeddings should have.
	// Only supported in text-embedding-3 and later models.
	Dimensions int `json:"dimensions,omitempty"`
	// The ExtraBody field allows for the inclusion of arbitrary key-value pairs
	// in the request body that may not be explicitly defined in this struct.
	ExtraBody map[string]any `json:"extra_body,omitempty"`
}

func (r EmbeddingRequestTokens) Convert() EmbeddingRequest {
	return EmbeddingRequest{
		Input:          r.Input,
		Model:          r.Model,
		User:           r.User,
		EncodingFormat: r.EncodingFormat,
		Dimensions:     r.Dimensions,
		ExtraBody:      r.ExtraBody,
	}
}

// CreateEmbeddings returns an EmbeddingResponse which will contain an Embedding for every item in |body.Input|.
// https://beta.openai.com/docs/api-reference/embeddings/create
//
// Body should be of type EmbeddingRequestStrings for embedding strings or EmbeddingRequestTokens
// for embedding groups of text already converted to tokens.
func (c *Client) CreateEmbeddings(
	ctx context.Context,
	conv EmbeddingRequestConverter,
) (res EmbeddingResponse, err error) {
	baseReq := conv.Convert()

	// The body map is used to dynamically construct the request payload for the embedding API.
	// Instead of relying on a fixed struct, the body map allows for flexible inclusion of fields
	// based on their presence, avoiding unnecessary or empty fields in the request.
	extraBody := baseReq.ExtraBody
	baseReq.ExtraBody = nil

	// Serialize baseReq to JSON
	jsonData, err := json.Marshal(baseReq)
	if err != nil {
		return
	}

	// Deserialize JSON to map[string]any
	var body map[string]any
	_ = json.Unmarshal(jsonData, &body)

	req, err := c.newRequest(
		ctx,
		http.MethodPost,
		c.fullURL("/embeddings", withModel(string(baseReq.Model))),
		withBody(body),           // Main request body.
		withExtraBody(extraBody), // Merge ExtraBody fields.
	)
	if err != nil {
		return
	}

	if baseReq.EncodingFormat != EmbeddingEncodingFormatBase64 {
		err = c.sendRequest(req, &res)
		return
	}

	base64Response := &EmbeddingResponseBase64{}
	err = c.sendRequest(req, base64Response)
	if err != nil {
		return
	}

	res, err = base64Response.ToEmbeddingResponse()
	return
}
