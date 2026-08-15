package generated

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

func NewClient(baseURL, apiKey string, httpClient *http.Client) *Client {
	return &Client{baseURL: baseURL, apiKey: apiKey, http: httpClient}
}

// DocumentControllerRequestUpload is generated for operationId DocumentController_requestUpload.
// It performs a post request against paths["/document/upload-request"] of the OpenAPI spec.
func (c *Client) DocumentControllerRequestUpload(ctx context.Context, req DocumentControllerRequestUploadRequest) error {
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	reqURL := fmt.Sprintf("%s/document/upload-request", c.baseURL)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := c.http.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	return nil
}
