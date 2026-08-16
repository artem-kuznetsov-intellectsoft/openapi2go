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
	opts    []RequestOption
	http    *http.Client
}

type RequestOption interface {
	Apply(*http.Request)
}

func NewClient(baseURL string, httpClient *http.Client, opts ...RequestOption) *Client {
	return &Client{
		baseURL: baseURL,
		opts:    opts,
		http:    httpClient,
	}
}

// DocumentControllerRequestUpload is generated for operationId DocumentController_requestUpload.
// It performs a post request against paths["/document/upload-request"] of the OpenAPI spec.
func (c *Client) DocumentControllerRequestUpload(ctx context.Context, req DocumentControllerRequestUploadRequest, opts ...RequestOption) error {
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

	for _, o := range c.opts {
		o.Apply(httpReq)
	}
	for _, o := range opts {
		o.Apply(httpReq)
	}

	httpResp, err := c.http.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	return nil
}
