package internal

import (
	"encoding/json"
	"strings"
)

// ProviderIntegration is the wire representation of the Provider
// Integrations API's ProviderIntegrationResource schema. The secret used to
// authenticate with the downstream provider is write-only: the API never
// returns it, only whether one is currently stored (HasSecret).
type ProviderIntegration struct {
	Id        string            `json:"id"`
	Name      string            `json:"name"`
	Provider  string            `json:"provider"`
	Url       string            `json:"url"`
	HasSecret bool              `json:"has_secret"`
	AuthType  string            `json:"auth_type"`
	Tags      map[string]string `json:"tags,omitempty"`
}

// CreateProviderIntegrationRequest is the body of POST /integrations/providers.
type CreateProviderIntegrationRequest struct {
	Name     string `json:"name,omitempty"`
	Provider string `json:"provider"`
	Url      string `json:"url"`
	Secret   string `json:"secret"`
	AuthType string `json:"auth_type"`
}

// UpdateProviderIntegrationRequest is the body of PUT
// /integrations/providers/{provider_id}. Provider can't be changed after
// creation (it's absent from this request's schema); Secret is omitted
// (left empty) to keep whatever secret is already stored.
type UpdateProviderIntegrationRequest struct {
	Name     string `json:"name,omitempty"`
	Url      string `json:"url"`
	Secret   string `json:"secret,omitempty"`
	AuthType string `json:"auth_type"`
}

// ProviderIntegrationsList is the response body of GET /integrations/providers.
type ProviderIntegrationsList struct {
	Integrations []ProviderIntegration `json:"integrations"`
}

// DecodeProviderIntegration decodes a single ProviderIntegration from a
// JSON response body.
func DecodeProviderIntegration(body string) (ProviderIntegration, error) {
	var p ProviderIntegration
	err := json.NewDecoder(strings.NewReader(body)).Decode(&p)
	return p, err
}

// DecodeProviderIntegrationsList decodes a {"integrations": [...]} response
// body.
func DecodeProviderIntegrationsList(body string) (ProviderIntegrationsList, error) {
	var list ProviderIntegrationsList
	err := json.NewDecoder(strings.NewReader(body)).Decode(&list)
	return list, err
}
