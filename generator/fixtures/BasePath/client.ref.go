package generated

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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

// GetUser is generated for operationId get-user.
// It performs a get request against paths["/user"] of the OpenAPI spec.
func (c *Client) GetUser(ctx context.Context, req GetUserRequest) (*ResponseOK, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	reqURL := fmt.Sprintf("%s/user", c.baseURL)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	switch httpResp.StatusCode {
	case 400:
		var errResp ResponseBadRequest
		if err := json.Unmarshal(respBody, &errResp); err != nil {
			return nil, err
		}
		return nil, errResp
	case 404:
		return nil, Response404{}
	}

	var resp ResponseOK
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
