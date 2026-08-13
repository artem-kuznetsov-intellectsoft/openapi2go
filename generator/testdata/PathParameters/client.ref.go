package generated

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

func NewClient(baseURL, apiKey string, httpClient *http.Client) *Client {
	return &Client{baseURL: baseURL, apiKey: apiKey, http: httpClient}
}

// GetUserById is generated for operationId get-user-by-id.
// It performs a get request against paths["/user/{userId}"] of the OpenAPI spec.
func (c *Client) GetUserById(ctx context.Context, params GetUserByIdParams) (*ResponseOK, error) {
	reqURL := fmt.Sprintf("%s/user/%v", c.baseURL, params.UserId)
	q := url.Values{}
	if params.MiddleName != nil {
		q.Set("middleName", fmt.Sprint(*params.MiddleName))
	}
	if len(q) > 0 {
		reqURL += "?" + q.Encode()
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	httpResp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	var resp ResponseOK
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
