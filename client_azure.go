package openai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var (
	ErrClientEmptyCallbackURL           = errors.New("Error retrieving callback URL (Operation-Location) for image request") //nolint:lll
	ErrClientRetrievingCallbackResponse = errors.New("Error retrieving callback response")                                   //nolint:lll
)

type AzureClient = Client

// Azure image request callback response struct.
type CBData []struct {
	URL string `json:"url"`
}
type CBResult struct {
	Data CBData `json:"data"`
}
type CallBackResponse struct {
	Created int64    `json:"created"`
	Expires int64    `json:"expires"`
	ID      string   `json:"id"`
	Result  CBResult `json:"result"`
	Status  string   `json:"status"`
}

func (c *AzureClient) sendAzureRequest(req *http.Request, v any) error {
	req.Header.Set("Accept", "application/json; charset=utf-8")

	// Check whether Content-Type is already set, Upload Files API requires
	// Content-Type == multipart/form-data
	contentType := req.Header.Get("Content-Type")
	if contentType == "" {
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
	}

	c.setCommonHeaders(req)

	res, err := c.config.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if isFailureStatusCode(res) {
		return c.handleErrorResp(res)
	}

	if c.config.APIType == APITypeAzure || c.config.APIType == APITypeAzureAD {
		// Special handling for initial call to Azure DALL-E API.
		if strings.Contains(req.URL.Path, "openai/images/generations") {
			return c.requestImage(res, v)
		}
		// Special handling for callBack to Azure DALL-E API.
		if strings.Contains(req.URL.Path, "openai/operations/images") {
			return c.imageRequestCallback(req, v, res)
		}
	}

	return decodeResponse(res.Body, v)
}

func (c *AzureClient) requestImage(res *http.Response, v any) error {
	_, _ = io.Copy(io.Discard, res.Body)
	callBackURL := res.Header.Get("Operation-Location")
	if callBackURL == "" {
		return ErrClientEmptyCallbackURL
	}
	newReq, err := http.NewRequest("GET", callBackURL, nil)
	if err != nil {
		return err
	}
	return c.sendAzureRequest(newReq, v)
}

// Handle image callback response from Azure DALL-E API.
func (c *AzureClient) imageRequestCallback(req *http.Request, v any, res *http.Response) error {
	// Retry Sleep seconds for Azure DALL-E 2 callback URL.
	var callBackWaitTime = 3

	// Wait for the callBack to complete
	var result *CallBackResponse
	err := json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return ErrClientRetrievingCallbackResponse
	}
	if result.Status == "" {
		return ErrClientRetrievingCallbackResponse
	}
	if result.Status != "succeeded" {
		time.Sleep(time.Duration(callBackWaitTime) * time.Second)
		req.Header.Add("Retry", "true")
		return c.sendAzureRequest(req, v)
	}

	// Convert the callBack response to the OpenAI ImageResponse
	var urlList []ImageResponseDataInner
	for _, data := range result.Result.Data {
		urlList = append(urlList, ImageResponseDataInner{URL: data.URL})
	}
	converted, _ := json.Marshal(ImageResponse{Created: result.Created, Data: urlList})
	return decodeResponse(bytes.NewReader(converted), v)
}

// fullURL returns full URL for request.
// args[0] is model name, if API type is Azure, model name is required to get deployment name.
func (c *AzureClient) fullAzureURL(suffix string, args ...any) string {
	baseURL := c.config.BaseURL
	baseURL = strings.TrimRight(baseURL, "/")
	switch {
	case strings.Contains(suffix, "/models"):
		// if suffix is /models change to {endpoint}/openai/models?api-version={api_version}
		// https://learn.microsoft.com/en-us/rest/api/cognitiveservices/azureopenaistable/models/list?tabs=HTTP
		return fmt.Sprintf("%s/%s%s?api-version=%s", baseURL, azureAPIPrefix, suffix, c.config.APIVersion)
	case strings.Contains(suffix, "/images"):
		// if suffix is /images change to {endpoint}openai/images/generations:submit?api-version={api_version}
		return fmt.Sprintf("%s/%s%s:submit?api-version=%s", baseURL, azureAPIPrefix, suffix, c.config.APIVersion)
	default:
		// /openai/deployments/{model}/chat/completions?api-version={api_version}
		azureDeploymentName := "UNKNOWN"
		if len(args) > 0 {
			model, ok := args[0].(string)
			if ok {
				azureDeploymentName = c.config.GetAzureDeploymentByModel(model)
			}
		}
		return fmt.Sprintf("%s/%s/%s/%s%s?api-version=%s", baseURL, azureAPIPrefix, azureDeploymentsPrefix,
			azureDeploymentName, suffix, c.config.APIVersion,
		)
	}
}
