package openapi

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
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		http:    httpClient,
	}
}

// Hint for AI: will be named as PascalCase(operationId) 
type CreateCustomerParams struct{}

type CreateCustomerRequest struct{}

type CreateCustomerResponse struct{}

type Response400 struct {
	Message string `json:"message"`
}

func (r Response400) Error() string { return "" }

type Response404 struct{}

func (r Response404) Error() string { return "" }

type Response500 struct{}

func (r Response500) Error() string { return "" }

// Hint for AI: 
// CreateCustomer is a method accessing an endpoint
// described by paths["/customer"].post of openapi spec.
// Arguments `params` and `req` are present in a signature if openapi has them.
func (c *Client) CreateCustomer(ctx context.Context, params CreateCustomerParams, req CreateCustomerRequest) (*CreateCustomerResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/customer", c.baseURL), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	// Hint for AI: place here some code for adding path and query params from `params`

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("api-key-config", c.apiKey)

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
		// Hint for AI: since Response400 is structured, we expect the body carrying the error payload.
		var resp Response400
		if err := json.Unmarshal(respBody, &resp); err != nil {
			return nil, err
		}
		return nil, resp
	case 404:
		return nil, Response404{}
	case 500:
		return nil, Response500{}
	}

	var resp CreateCustomerResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
