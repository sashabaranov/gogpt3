package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Client is OpenAI GPT-3 API client.
type Client struct {
	config ClientConfig

	requestBuilder    requestBuilder
	createFormBuilder func(io.Writer) formBuilder
}

// NewClient creates new OpenAI API client.
func NewClient(authToken string, options ...Option) *Client {
	config := DefaultConfig(authToken)

	for _, opt := range options {
		opt(&config)
	}

	return newClient(config)
}

// NewAzureClient create new openAI API from Azure client
func NewAzureClient(authToken string, baseUrl string, engine string, options ...Option) *Client {
	config := DefaultAzureConfig(authToken, baseUrl, engine)

	for _, opt := range options {
		opt(&config)
	}

	return newClient(config)
}

func newClient(config ClientConfig) *Client {
	return &Client{
		config:         config,
		requestBuilder: newRequestBuilder(),
		createFormBuilder: func(body io.Writer) formBuilder {
			return newFormBuilder(body)
		},
	}
}

func (c *Client) sendRequest(req *http.Request, v any) error {
	req.Header.Set("Accept", "application/json; charset=utf-8")
	// Azure API Key authentication
	if c.config.APIType == APITypeAzure {
		req.Header.Set(AzureAPIKeyHeader, c.config.authToken)
	} else {
		// OpenAI or Azure AD authentication
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.authToken))
	}

	// Check whether Content-Type is already set, Upload Files API requires
	// Content-Type == multipart/form-data
	contentType := req.Header.Get("Content-Type")
	if contentType == "" {
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
	}

	if len(c.config.OrgID) > 0 {
		req.Header.Set("OpenAI-Organization", c.config.OrgID)
	}

	res, err := c.config.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		return c.handleErrorResp(res)
	}

	return decodeResponse(res.Body, v)
}

func decodeResponse(body io.Reader, v any) error {
	if v == nil {
		return nil
	}

	if result, ok := v.(*string); ok {
		return decodeString(body, result)
	}
	return json.NewDecoder(body).Decode(v)
}

func decodeString(body io.Reader, output *string) error {
	b, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	*output = string(b)
	return nil
}

func (c *Client) fullURL(suffix string) string {
	// /openai/deployments/{engine}/chat/completions?api-version={api_version}
	if c.config.APIType == APITypeAzure || c.config.APIType == APITypeAzureAD {
		baseURL := c.config.BaseURL
		baseURL = strings.TrimRight(baseURL, "/")
		return fmt.Sprintf("%s/%s/%s/%s%s?api-version=%s",
			baseURL, azureAPIPrefix, azureDeploymentsPrefix, c.config.Engine, suffix, c.config.APIVersion)
	}

	// c.config.APIType == APITypeOpenAI || c.config.APIType == ""
	return fmt.Sprintf("%s%s", c.config.BaseURL, suffix)
}

func (c *Client) newStreamRequest(
	ctx context.Context,
	method string,
	urlSuffix string,
	body any) (*http.Request, error) {
	req, err := c.requestBuilder.build(ctx, method, c.fullURL(urlSuffix), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	// https://learn.microsoft.com/en-us/azure/cognitive-services/openai/reference#authentication
	// Azure API Key authentication
	if c.config.APIType == APITypeAzure {
		req.Header.Set(AzureAPIKeyHeader, c.config.authToken)
	} else {
		// OpenAI or Azure AD authentication
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.authToken))
	}
	if c.config.OrgID != "" {
		req.Header.Set("OpenAI-Organization", c.config.OrgID)
	}
	return req, nil
}

func (c *Client) handleErrorResp(resp *http.Response) error {
	var errRes ErrorResponse
	err := json.NewDecoder(resp.Body).Decode(&errRes)
	if err != nil || errRes.Error == nil {
		reqErr := RequestError{
			HTTPStatusCode: resp.StatusCode,
			Err:            err,
		}
		return fmt.Errorf("error, %w", &reqErr)
	}
	errRes.Error.HTTPStatusCode = resp.StatusCode
	return fmt.Errorf("error, status code: %d, message: %w", resp.StatusCode, errRes.Error)
}
