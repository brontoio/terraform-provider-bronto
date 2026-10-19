package internal

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client -
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	ApiKey     string
}

type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	if e.Body == "" {
		return fmt.Sprintf("received unexpected status code: %d", e.StatusCode)
	}
	return fmt.Sprintf("received unexpected status code: %d, body: %s", e.StatusCode, e.Body)
}

// IsNotFound returns true if err is an APIError representing a 404 response.
func IsNotFound(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound
}

func NewClient(endpoint *string, apiKey *string) (*Client, error) {
	if endpoint == nil || apiKey == nil {
		return nil, fmt.Errorf("endpoint and api key are required")
	}

	baseURL, err := parseBaseURL(*endpoint)
	if err != nil {
		return nil, err
	}

	c := Client{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		BaseURL:    baseURL,
		ApiKey:     *apiKey,
	}
	return &c, nil
}

func parseBaseURL(endpoint string) (string, error) {
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("endpoint %q is not a valid URL: %w", endpoint, err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("endpoint %q must be an absolute URL with an http or https scheme", endpoint)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("endpoint %q must include a host", endpoint)
	}
	return strings.TrimSuffix(endpoint, "/"), nil
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

func (c *Client) UpdateRequest(ctx context.Context, httpReqBody []byte, path string) (string, error) {
	endpoint := c.BaseURL + path
	httpReq := c.createHttpRequest(ctx, httpReqBody, http.MethodPatch, endpoint)
	return c.doRequest(httpReq, http.StatusOK)
}

func (c *Client) PostRequest(ctx context.Context, httpReqBody []byte, path string) (string, error) {
	endpoint := c.BaseURL + path
	httpReq := c.createHttpRequest(ctx, httpReqBody, http.MethodPost, endpoint)
	return c.doRequest(httpReq, http.StatusOK)
}

// PutRequest performs a full-replace update via HTTP PUT, for endpoints
// that respond 204 No Content instead of echoing the updated resource (as
// opposed to UpdateRequest, which PATCHes and expects a 200 with a body).
func (c *Client) PutRequest(ctx context.Context, httpReqBody []byte, path string) error {
	endpoint := c.BaseURL + path
	httpReq := c.createHttpRequest(ctx, httpReqBody, http.MethodPut, endpoint)
	_, err := c.doRequest(httpReq, http.StatusNoContent)
	return err
}

// PostNoContentRequest performs an HTTP POST against an endpoint that
// responds 204 No Content instead of echoing a body (as opposed to
// PostRequest, which expects a 200 with a body).
func (c *Client) PostNoContentRequest(ctx context.Context, httpReqBody []byte, path string) error {
	endpoint := c.BaseURL + path
	httpReq := c.createHttpRequest(ctx, httpReqBody, http.MethodPost, endpoint)
	_, err := c.doRequest(httpReq, http.StatusNoContent)
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
	httpReq.Header.Add("X-BRONTO-API-KEY", c.ApiKey)
	httpReq.Header.Add("Content-Type", "application/json")
}

func (c *Client) doRequest(httpReq *http.Request, expectedResult int) (string, error) {
	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != expectedResult {
		return "", &APIError{StatusCode: resp.StatusCode, Body: string(body)}
	}

	// Return the body as a string
	return string(body), nil
}
