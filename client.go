package openai

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	utils "github.com/sashabaranov/go-openai/internal"
)

// Client is OpenAI GPT-3 API client.
type Client struct {
	config ClientConfig

	requestBuilder    utils.RequestBuilder
	createFormBuilder func(io.Writer) utils.FormBuilder
}

type Response interface {
	SetHeader(http.Header)
}

type httpHeader http.Header

func (h *httpHeader) SetHeader(header http.Header) {
	*h = httpHeader(header)
}

func (h *httpHeader) Header() http.Header {
	return http.Header(*h)
}

func (h *httpHeader) GetRateLimitHeaders() RateLimitHeaders {
	return newRateLimitHeaders(h.Header())
}

type RawResponse struct {
	io.ReadCloser

	httpHeader
}

// NewClient creates new OpenAI API client.
func NewClient(authToken string) *Client {
	config := DefaultConfig(authToken)
	return NewClientWithConfig(config)
}

// NewClientWithConfig creates new OpenAI API client for specified config.
func NewClientWithConfig(config ClientConfig) *Client {
	return &Client{
		config:         config,
		requestBuilder: utils.NewRequestBuilder(),
		createFormBuilder: func(body io.Writer) utils.FormBuilder {
			return utils.NewFormBuilder(body)
		},
	}
}

// NewOrgClient creates new OpenAI API client for specified Organization ID.
//
// Deprecated: Please use NewClientWithConfig.
func NewOrgClient(authToken, org string) *Client {
	config := DefaultConfig(authToken)
	config.OrgID = org
	return NewClientWithConfig(config)
}

type requestOptions struct {
	body   any
	header http.Header
}

type requestOption func(*requestOptions)

func withBody(body any) requestOption {
	return func(args *requestOptions) {
		args.body = body
	}
}

func withExtraBody(extraBody map[string]any) requestOption {
	return func(args *requestOptions) {
		// Assert that args.body is a map[string]any.
		bodyMap, ok := args.body.(map[string]any)
		if ok {
			// If it's a map[string]any then only add extraBody
			// fields to args.body otherwise keep only fields in request struct.
			for key, value := range extraBody {
				bodyMap[key] = value
			}
		}
	}
}

func withContentType(contentType string) requestOption {
	return func(args *requestOptions) {
		args.header.Set("Content-Type", contentType)
	}
}

func withBetaAssistantVersion(version string) requestOption {
	return func(args *requestOptions) {
		args.header.Set("OpenAI-Beta", fmt.Sprintf("assistants=%s", version))
	}
}

func (c *Client) newRequest(ctx context.Context, method, url string, setters ...requestOption) (*http.Request, error) {
	// Default Options
	args := &requestOptions{
		body:   nil,
		header: make(http.Header),
	}
	for _, setter := range setters {
		setter(args)
	}
	req, err := c.requestBuilder.Build(ctx, method, url, args.body, args.header)
	if err != nil {
		return nil, err
	}
	c.setCommonHeaders(req)
	return req, nil
}

func (c *Client) sendRequest(req *http.Request, v Response) error {
	req.Header.Set("Accept", "application/json")

	// Check whether Content-Type is already set, Upload Files API requires
	// Content-Type == multipart/form-data
	contentType := req.Header.Get("Content-Type")
	if contentType == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := c.config.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if v != nil {
		v.SetHeader(res.Header)
	}

	if isFailureStatusCode(res) {
		return c.handleErrorResp(res)
	}

	return decodeResponse(res.Body, v)
}

func (c *Client) sendRequestRaw(req *http.Request) (response RawResponse, err error) {
	resp, err := c.config.HTTPClient.Do(req) //nolint:bodyclose // body should be closed by outer function
	if err != nil {
		return
	}

	if isFailureStatusCode(resp) {
		err = c.handleErrorResp(resp)
		return
	}

	response.SetHeader(resp.Header)
	response.ReadCloser = resp.Body
	return
}

func sendRequestStream[T streamable](client *Client, req *http.Request) (*streamReader[T], error) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	resp, err := client.config.HTTPClient.Do(req) //nolint:bodyclose // body is closed in stream.Close()
	if err != nil {
		return new(streamReader[T]), err
	}
	if isFailureStatusCode(resp) {
		return new(streamReader[T]), client.handleErrorResp(resp)
	}
	return &streamReader[T]{
		emptyMessagesLimit: client.config.EmptyMessagesLimit,
		reader:             bufio.NewReader(resp.Body),
		response:           resp,
		errAccumulator:     utils.NewErrorAccumulator(),
		unmarshaler:        &utils.JSONUnmarshaler{},
		httpHeader:         httpHeader(resp.Header),
	}, nil
}

func (c *Client) setCommonHeaders(req *http.Request) {
	// https://learn.microsoft.com/en-us/azure/cognitive-services/openai/reference#authentication
	switch c.config.APIType {
	case APITypeAzure, APITypeCloudflareAzure:
		// Azure API Key authentication
		req.Header.Set(AzureAPIKeyHeader, c.config.authToken)
	case APITypeAnthropic:
		// https://docs.anthropic.com/en/api/versioning
		req.Header.Set("anthropic-version", c.config.APIVersion)
	case APITypeOpenAI, APITypeAzureAD:
		fallthrough
	default:
		if c.config.authToken != "" {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.authToken))
		}
	}

	if c.config.OrgID != "" {
		req.Header.Set("OpenAI-Organization", c.config.OrgID)
	}
}

func isFailureStatusCode(resp *http.Response) bool {
	return resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest
}

func decodeResponse(body io.Reader, v any) error {
	if v == nil {
		return nil
	}

	switch o := v.(type) {
	case *string:
		return decodeString(body, o)
	case *audioTextResponse:
		return decodeString(body, &o.Text)
	default:
		return json.NewDecoder(body).Decode(v)
	}
}

func decodeString(body io.Reader, output *string) error {
	b, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	*output = string(b)
	return nil
}

type fullURLOptions struct {
	model string
}

type fullURLOption func(*fullURLOptions)

func withModel(model string) fullURLOption {
	return func(args *fullURLOptions) {
		args.model = model
	}
}

var azureDeploymentsEndpoints = []string{
	"/completions",
	"/embeddings",
	"/chat/completions",
	"/audio/transcriptions",
	"/audio/translations",
	"/audio/speech",
	"/images/generations",
}

// fullURL returns full URL for request.
func (c *Client) fullURL(suffix string, setters ...fullURLOption) string {
	baseURL := strings.TrimRight(c.config.BaseURL, "/")
	args := fullURLOptions{}
	for _, setter := range setters {
		setter(&args)
	}

	if c.config.APIType == APITypeAzure || c.config.APIType == APITypeAzureAD {
		baseURL = c.baseURLWithAzureDeployment(baseURL, suffix, args.model)
	}

	if c.config.APIVersion != "" {
		suffix = c.suffixWithAPIVersion(suffix)
	}
	return fmt.Sprintf("%s%s", baseURL, suffix)
}

func (c *Client) suffixWithAPIVersion(suffix string) string {
	parsedSuffix, err := url.Parse(suffix)
	if err != nil {
		panic("failed to parse url suffix")
	}
	query := parsedSuffix.Query()
	query.Add("api-version", c.config.APIVersion)
	return fmt.Sprintf("%s?%s", parsedSuffix.Path, query.Encode())
}

func (c *Client) baseURLWithAzureDeployment(baseURL, suffix, model string) (newBaseURL string) {
	baseURL = fmt.Sprintf("%s/%s", strings.TrimRight(baseURL, "/"), azureAPIPrefix)
	if containsSubstr(azureDeploymentsEndpoints, suffix) {
		azureDeploymentName := c.config.GetAzureDeploymentByModel(model)
		if azureDeploymentName == "" {
			azureDeploymentName = "UNKNOWN"
		}
		baseURL = fmt.Sprintf("%s/%s/%s", baseURL, azureDeploymentsPrefix, azureDeploymentName)
	}
	return baseURL
}

func (c *Client) handleErrorResp(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error, reading response body: %w", err)
	}
	var errRes ErrorResponse
	err = json.Unmarshal(body, &errRes)
	if err != nil || errRes.Error == nil {
		reqErr := &RequestError{
			HTTPStatus:     resp.Status,
			HTTPStatusCode: resp.StatusCode,
			Err:            err,
			Body:           body,
		}
		if errRes.Error != nil {
			reqErr.Err = errRes.Error
		}
		return reqErr
	}

	errRes.Error.HTTPStatus = resp.Status
	errRes.Error.HTTPStatusCode = resp.StatusCode
	return errRes.Error
}

func containsSubstr(s []string, e string) bool {
	for _, v := range s {
		if strings.Contains(e, v) {
			return true
		}
	}
	return false
}
