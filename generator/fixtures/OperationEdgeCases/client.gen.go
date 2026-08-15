package generated

import (
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

// BrokenTemplate is generated for operationId BrokenTemplate.
// It performs a get request against paths["/broken{template"] of the OpenAPI spec.
func (c *Client) BrokenTemplate(ctx context.Context) error {
	reqURL := fmt.Sprintf("%s/broken{template", c.baseURL)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
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

// PostNoData is generated for operationId PostNoData.
// It performs a post request against paths["/nodata"] of the OpenAPI spec.
func (c *Client) PostNoData(ctx context.Context) error {
	reqURL := fmt.Sprintf("%s/nodata", c.baseURL)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, nil)
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

// GetResource is generated for operationId GetResource.
// It performs a get request against paths["/resource/{id}"] of the OpenAPI spec.
func (c *Client) GetResource(ctx context.Context, params GetResourceParams) (*GetResourceResponse200, error) {
	reqURL := fmt.Sprintf("%s/resource/%v", c.baseURL, params.Id)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("X-Api-Key", fmt.Sprint(params.XApiKey))
	if params.XTraceId != nil {
		httpReq.Header.Set("X-Trace-Id", fmt.Sprint(*params.XTraceId))
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
	case 404:
		return nil, Response404{}
	}

	var resp GetResourceResponse200
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// ReplaceResource is generated for operationId ReplaceResource.
// It performs a put request against paths["/resource/{id}"] of the OpenAPI spec.
func (c *Client) ReplaceResource(ctx context.Context, params ReplaceResourceParams) error {
	reqURL := fmt.Sprintf("%s/resource/%v", c.baseURL, params.Id)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPut, reqURL, nil)
	if err != nil {
		return err
	}

	httpReq.Header.Set("X-Api-Key", fmt.Sprint(params.XApiKey))
	httpResp, err := c.http.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	return nil
}

// DeleteResource is generated for operationId DeleteResource.
// It performs a delete request against paths["/resource/{id}"] of the OpenAPI spec.
func (c *Client) DeleteResource(ctx context.Context, params DeleteResourceParams) error {
	reqURL := fmt.Sprintf("%s/resource/%v", c.baseURL, params.Id)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodDelete, reqURL, nil)
	if err != nil {
		return err
	}

	httpReq.Header.Set("X-Api-Key", fmt.Sprint(params.XApiKey))
	httpResp, err := c.http.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	switch httpResp.StatusCode {
	case 404:
		return Response404{}
	}

	return nil
}

// OptionsResource is generated for operationId OptionsResource.
// It performs a options request against paths["/resource/{id}"] of the OpenAPI spec.
func (c *Client) OptionsResource(ctx context.Context, params OptionsResourceParams) error {
	reqURL := fmt.Sprintf("%s/resource/%v", c.baseURL, params.Id)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodOptions, reqURL, nil)
	if err != nil {
		return err
	}

	httpReq.Header.Set("X-Api-Key", fmt.Sprint(params.XApiKey))
	httpResp, err := c.http.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	return nil
}

// HeadResource is generated for operationId HeadResource.
// It performs a head request against paths["/resource/{id}"] of the OpenAPI spec.
func (c *Client) HeadResource(ctx context.Context, params HeadResourceParams) error {
	reqURL := fmt.Sprintf("%s/resource/%v", c.baseURL, params.Id)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodHead, reqURL, nil)
	if err != nil {
		return err
	}

	httpReq.Header.Set("X-Api-Key", fmt.Sprint(params.XApiKey))
	httpResp, err := c.http.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	return nil
}

// PatchResource is generated for operationId PatchResource.
// It performs a patch request against paths["/resource/{id}"] of the OpenAPI spec.
func (c *Client) PatchResource(ctx context.Context, params PatchResourceParams) error {
	reqURL := fmt.Sprintf("%s/resource/%v", c.baseURL, params.Id)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPatch, reqURL, nil)
	if err != nil {
		return err
	}

	httpReq.Header.Set("X-Api-Key", fmt.Sprint(params.XApiKey))
	httpResp, err := c.http.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	return nil
}

// TraceResource is generated for operationId TraceResource.
// It performs a trace request against paths["/resource/{id}"] of the OpenAPI spec.
func (c *Client) TraceResource(ctx context.Context, params TraceResourceParams) error {
	reqURL := fmt.Sprintf("%s/resource/%v", c.baseURL, params.Id)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodTrace, reqURL, nil)
	if err != nil {
		return err
	}

	httpReq.Header.Set("X-Api-Key", fmt.Sprint(params.XApiKey))
	httpResp, err := c.http.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	return nil
}
