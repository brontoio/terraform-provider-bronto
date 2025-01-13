package internal

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client -
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	ApkKey     string
}

func NewClient(region *string, apiKey *string) (*Client, error) {
	c := Client{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}

	// If region or apiKey not provided, return empty client
	if region == nil || apiKey == nil {
		return &c, fmt.Errorf("region and api key are required")
	}

	if strings.EqualFold("EU", *region) {
		c.BaseURL = "https://api.eu.brontobytes.io"
	} else if strings.EqualFold("US", *region) {
		c.BaseURL = "https://api.us.brontobytes.io"
	} else if strings.EqualFold("Staging", *region) {
		c.BaseURL = "https://api.staging.brontobytes.io"
	} else {
		return nil, fmt.Errorf("region %s is not valid", *region)
	}
	c.ApkKey = *apiKey
	return &c, nil
}

func (c *Client) CreateRequest(ctx context.Context, httpReqBody []byte, path string) (string, error) {
	endpoint := c.BaseURL + path
	httpReq := c.createHttpRequest(ctx, httpReqBody, http.MethodPost, endpoint)
	return c.doRequest(httpReq, http.StatusCreated)
}

func (c *Client) ReadRequest(ctx context.Context, path string) (string, error) {
	endpoint := c.BaseURL + path
	httpReq := c.createHttpRequest(ctx, nil, http.MethodGet, endpoint)
	return c.doRequest(httpReq, http.StatusOK)
}

func (c *Client) UpdateRequest(ctx context.Context, httpReqBody []byte, path string) error {
	endpoint := c.BaseURL + path
	httpReq := c.createHttpRequest(ctx, httpReqBody, http.MethodPatch, endpoint)
	_, err := c.doRequest(httpReq, http.StatusOK)
	return err
}

func (c *Client) DeleteRequest(ctx context.Context, path string) error {
	endpoint := c.BaseURL + path
	httpReq := c.createHttpRequest(ctx, nil, http.MethodDelete, endpoint)
	_, err := c.doRequest(httpReq, http.StatusNoContent)
	return err
}

func (c *Client) createHttpRequest(ctx context.Context, httpReqBody []byte, method, endpoint string) *http.Request {
	httpReq, _ := http.NewRequestWithContext(
		ctx,
		method,
		endpoint,
		bytes.NewBuffer(httpReqBody),
	)
	c.addHeaders(httpReq)
	return httpReq
}

func (c *Client) addHeaders(httpReq *http.Request) {
	httpReq.Header.Add("X-BRONTO-API-KEY", c.ApkKey)
	httpReq.Header.Add("Content-Type", "application/json")
}

func (c *Client) doRequest(httpReq *http.Request, expectedResult int) (string, error) {
	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != expectedResult {
		return "", fmt.Errorf("received non-%d status code: %d", expectedResult, resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	// Return the body as a string
	return string(body), nil
}
