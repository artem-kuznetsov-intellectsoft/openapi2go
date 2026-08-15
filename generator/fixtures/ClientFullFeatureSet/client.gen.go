package generated

import (
	"bytes"
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

// ListItems is generated for operationId ListItems.
// It performs a get request against paths["/items"] of the OpenAPI spec.
func (c *Client) ListItems(ctx context.Context, params ListItemsParams) (*ListItemsResponse200, error) {
	reqURL := fmt.Sprintf("%s/items", c.baseURL)
	q := url.Values{}
	if params.Page != nil {
		q.Set("page", fmt.Sprint(*params.Page))
	}
	q.Set("pageSize", fmt.Sprint(params.PageSize))
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

	var resp ListItemsResponse200
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// CreateItem is generated for operationId CreateItem.
// It performs a post request against paths["/items"] of the OpenAPI spec.
func (c *Client) CreateItem(ctx context.Context, req CreateItemRequest) (*CreateItemResponse201, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	reqURL := fmt.Sprintf("%s/items", c.baseURL)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(body))
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
		return nil, Response400{}
	case 401:
		return nil, Response401{}
	}

	var resp CreateItemResponse201
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetItem is generated for operationId GetItem.
// It performs a get request against paths["/items/{itemId}"] of the OpenAPI spec.
func (c *Client) GetItem(ctx context.Context, params GetItemParams) (*Item, error) {
	reqURL := fmt.Sprintf("%s/items/%v", c.baseURL, params.ItemId)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("X-Api-Key", fmt.Sprint(params.XApiKey))
	if params.XRequestId != nil {
		httpReq.Header.Set("X-Request-Id", fmt.Sprint(*params.XRequestId))
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
	case 400:
		var errResp ValidationError
		if err := json.Unmarshal(respBody, &errResp); err != nil {
			return nil, err
		}
		return nil, errResp
	case 404:
		return nil, Response404{}
	}

	var resp Item
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// ReplaceItem is generated for operationId ReplaceItem.
// It performs a put request against paths["/items/{itemId}"] of the OpenAPI spec.
func (c *Client) ReplaceItem(ctx context.Context, params ReplaceItemParams, req ReplaceItemRequest) (*ReplaceItemResponse200, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	reqURL := fmt.Sprintf("%s/items/%v", c.baseURL, params.ItemId)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPut, reqURL, bytes.NewReader(body))
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
	case 409:
		var errResp ConflictError
		if err := json.Unmarshal(respBody, &errResp); err != nil {
			return nil, err
		}
		return nil, errResp
	}

	var resp ReplaceItemResponse200
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// DeleteItem is generated for operationId DeleteItem.
// It performs a delete request against paths["/items/{itemId}"] of the OpenAPI spec.
func (c *Client) DeleteItem(ctx context.Context, params DeleteItemParams) error {
	reqURL := fmt.Sprintf("%s/items/%v", c.baseURL, params.ItemId)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodDelete, reqURL, nil)
	if err != nil {
		return err
	}

	httpResp, err := c.http.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return err
	}

	switch httpResp.StatusCode {
	case 404:
		var errResp NotFoundError
		if err := json.Unmarshal(respBody, &errResp); err != nil {
			return err
		}
		return errResp
	}

	return nil
}

// ArchiveItem is generated for operationId ArchiveItem.
// It performs a patch request against paths["/items/{itemId}/archive"] of the OpenAPI spec.
func (c *Client) ArchiveItem(ctx context.Context, params ArchiveItemParams) error {
	reqURL := fmt.Sprintf("%s/items/%v/archive", c.baseURL, params.ItemId)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPatch, reqURL, nil)
	if err != nil {
		return err
	}

	httpResp, err := c.http.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	return nil
}
