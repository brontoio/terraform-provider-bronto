package internal

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	return &Client{
		BaseURL:    srv.URL,
		HTTPClient: srv.Client(),
		ApiKey:     "test-key",
	}
}

func TestCreateRequestSendsApiKeyHeaderAndReturnsBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-BRONTO-API-KEY"); got != "test-key" {
			t.Errorf("X-BRONTO-API-KEY header = %q, want %q", got, "test-key")
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type header = %q, want application/json", got)
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":"123"}`))
	}))
	defer srv.Close()

	c := testClient(t, srv)
	body, err := c.CreateRequest(context.Background(), []byte(`{}`), "/users")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if body != `{"id":"123"}` {
		t.Errorf("body = %q, want %q", body, `{"id":"123"}`)
	}
}

func TestReadRequestNotFoundReturnsAPIErrorDetectableByIsNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"not found"}`))
	}))
	defer srv.Close()

	c := testClient(t, srv)
	_, err := c.ReadRequest(context.Background(), "/users/does-not-exist")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("IsNotFound(err) = false, want true for err: %v", err)
	}
}

func TestReadRequestUnexpectedStatusSurfacesResponseBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"boom"}`))
	}))
	defer srv.Close()

	c := testClient(t, srv)
	_, err := c.ReadRequest(context.Background(), "/users/1")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if IsNotFound(err) {
		t.Errorf("IsNotFound(err) = true, want false for a 500")
	}
	if got := err.Error(); got == "" || !strings.Contains(got, "boom") {
		t.Errorf("error %q does not surface the response body", got)
	}
}

func TestUpdateRequestReturnsResponseBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method = %s, want PATCH", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"123","first_name":"Updated"}`))
	}))
	defer srv.Close()

	c := testClient(t, srv)
	body, err := c.UpdateRequest(context.Background(), []byte(`{}`), "/users/123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if body != `{"id":"123","first_name":"Updated"}` {
		t.Errorf("body = %q", body)
	}
}

func TestPostRequestExpects200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"123","status":"DEACTIVATED"}`))
	}))
	defer srv.Close()

	c := testClient(t, srv)
	body, err := c.PostRequest(context.Background(), nil, "/users/123/deactivate")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if body != `{"id":"123","status":"DEACTIVATED"}` {
		t.Errorf("body = %q", body)
	}
}

func TestDeleteRequestExpectsNoContent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := testClient(t, srv)
	if err := c.DeleteRequest(context.Background(), "/users/123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewClientRejectsMissingArgs(t *testing.T) {
	if _, err := NewClient(nil, nil); err == nil {
		t.Error("expected an error for nil endpoint and api key")
	}

	key := "key"
	for _, endpoint := range []string{"not-a-url", "ftp://api.example.com", "https://", "https:///no-host"} {
		endpoint := endpoint
		if _, err := NewClient(&endpoint, &key); err == nil {
			t.Errorf("expected an error for invalid endpoint %q", endpoint)
		}
	}
}

func TestNewClientUsesEndpointAsBaseURL(t *testing.T) {
	cases := map[string]string{
		"https://api.eu.bronto.io":      "https://api.eu.bronto.io",
		"https://api.us.bronto.io/":     "https://api.us.bronto.io",
		"http://localhost:8080":         "http://localhost:8080",
		"https://api.staging.bronto.io": "https://api.staging.bronto.io",
	}
	key := "key"
	for endpoint, wantURL := range cases {
		endpoint := endpoint
		c, err := NewClient(&endpoint, &key)
		if err != nil {
			t.Fatalf("endpoint %s: unexpected error: %v", endpoint, err)
		}
		if c.BaseURL != wantURL {
			t.Errorf("endpoint %s: BaseURL = %s, want %s", endpoint, c.BaseURL, wantURL)
		}
		if c.ApiKey != key {
			t.Errorf("endpoint %s: ApiKey = %s, want %s", endpoint, c.ApiKey, key)
		}
	}
}
