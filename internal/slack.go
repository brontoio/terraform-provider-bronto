package internal

import (
	"encoding/json"
	"strings"
)

// SlackChannel is one entry in a Slack integration's channels list. The API
// also accepts a deprecated bare-string shorthand for a channel with
// default settings; this provider only ever sends/reads the object form.
type SlackChannel struct {
	Name              string `json:"name"`
	IncludeGraphImage bool   `json:"include_graph_image"`
}

// SlackIntegration is the wire representation of the Slack Integrations
// API's SlackIntegrationResource schema.
type SlackIntegration struct {
	Id          string         `json:"id"`
	Name        string         `json:"name"`
	Channels    []SlackChannel `json:"channels"`
	WorkspaceId string         `json:"workspace_id"`
}

// SlackIntegrationRequest is the body of both POST /integrations/slack
// (create) and PUT /integrations/slack/{id} (full replace) - both accept
// the same shape.
type SlackIntegrationRequest struct {
	Name        string         `json:"name"`
	Channels    []SlackChannel `json:"channels"`
	WorkspaceId string         `json:"workspace_id"`
}

// SlackWorkspace is the wire representation of the Slack Integrations API's
// SlackWorkspaceResource schema. Workspaces are connected via an OAuth flow
// in the Bronto UI, not created through this API - the id here is an
// internal identifier, not Slack's own workspace/team id.
type SlackWorkspace struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

// SlackWorkspacesList is the response body of
// GET /integrations/slack/workspaces.
type SlackWorkspacesList struct {
	SlackWorkspaces []SlackWorkspace `json:"slack_workspaces"`
}

// DecodeSlackIntegration decodes a single SlackIntegration from a JSON
// response body.
func DecodeSlackIntegration(body string) (SlackIntegration, error) {
	var s SlackIntegration
	err := json.NewDecoder(strings.NewReader(body)).Decode(&s)
	return s, err
}

// DecodeSlackWorkspacesList decodes a {"slack_workspaces": [...]} response
// body.
func DecodeSlackWorkspacesList(body string) (SlackWorkspacesList, error) {
	var list SlackWorkspacesList
	err := json.NewDecoder(strings.NewReader(body)).Decode(&list)
	return list, err
}
