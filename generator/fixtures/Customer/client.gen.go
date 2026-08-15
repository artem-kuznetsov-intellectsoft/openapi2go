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

// CustomerControllerGetAllCustomers is generated for operationId CustomerController_getAllCustomers.
// It performs a get request against paths["/customer"] of the OpenAPI spec.
func (c *Client) CustomerControllerGetAllCustomers(ctx context.Context, params CustomerControllerGetAllCustomersParams) (*CustomerControllerGetAllCustomersResponse200, error) {
	reqURL := fmt.Sprintf("%s/customer", c.baseURL)
	q := url.Values{}
	q.Set("limit", fmt.Sprint(params.Limit))
	if params.Page != nil {
		q.Set("page", fmt.Sprint(*params.Page))
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

	switch httpResp.StatusCode {
	case 401:
		return nil, Response401{}
	}

	var resp CustomerControllerGetAllCustomersResponse200
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
