package internal

import (
	"encoding/json"
	"strings"
)

// WebhookIntegration is the wire representation of the Webhook Integrations
// API's WebhookIntegrationResource schema.
type WebhookIntegration struct {
	Id            string            `json:"id"`
	Name          string            `json:"name"`
	Url           string            `json:"url"`
	CustomHeaders map[string]string `json:"custom_headers"`
	Payload       string            `json:"payload"`
}

// WebhookIntegrationRequest is the body of both POST /integrations/webhooks
// (create) and PUT /integrations/webhooks/{id} (full replace) - both
// accept the same shape.
type WebhookIntegrationRequest struct {
	Name          string            `json:"name"`
	Url           string            `json:"url"`
	CustomHeaders map[string]string `json:"custom_headers,omitempty"`
	Payload       string            `json:"payload,omitempty"`
}

// DecodeWebhookIntegration decodes a single WebhookIntegration from a JSON
// response body.
func DecodeWebhookIntegration(body string) (WebhookIntegration, error) {
	var w WebhookIntegration
	err := json.NewDecoder(strings.NewReader(body)).Decode(&w)
	return w, err
}
