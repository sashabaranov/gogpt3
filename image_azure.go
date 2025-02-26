package openai

import (
	"context"
	"net/http"
)

// CreateImage - API call to create an image. This is the main endpoint of the DALL-E API.
func (c *AzureClient) CreateAzureImage(ctx context.Context, request ImageRequest) (response ImageResponse, err error) {
	urlSuffix := "/images/generations"
	req, err := c.newRequest(ctx, http.MethodPost, c.fullAzureURL(urlSuffix), withBody(request))
	if err != nil {
		return
	}

	err = c.sendAzureRequest(req, &response)
	return
}
