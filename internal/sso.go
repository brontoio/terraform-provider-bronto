package internal

import (
	"encoding/json"
	"strings"
)

// AttributeMapping maps one identity provider attribute to a Bronto user
// attribute, per the SSO API's Idp.attribute_mappings schema.
type AttributeMapping struct {
	ProviderAttr string `json:"provider_attr"`
	UserAttr     string `json:"user_attr"`
}

// Idp is the wire representation of the SSO API's Idp schema. It has no
// separate generated id: name is the identifier used in the path
// (/sso/idps/{idp_name}) and can't be changed after creation.
type Idp struct {
	Name               string             `json:"name"`
	ProviderType       string             `json:"provider_type"`
	Identifiers        []string           `json:"identifiers,omitempty"`
	MetaCert           string             `json:"meta_cert,omitempty"`
	IssuerUrl          string             `json:"issuer_url,omitempty"`
	ClientId           string             `json:"client_id,omitempty"`
	ClientSecret       string             `json:"client_secret,omitempty"`
	AcceptIdpInitiated bool               `json:"accept_idp_initiated"`
	AppClientId        string             `json:"app_client_id,omitempty"`
	AttributeMappings  []AttributeMapping `json:"attribute_mappings"`
}

// CreateIdpRequest is the body of POST /sso/idps.
type CreateIdpRequest struct {
	Name               string             `json:"name"`
	ProviderType       string             `json:"provider_type"`
	Identifiers        []string           `json:"identifiers,omitempty"`
	MetaCert           string             `json:"meta_cert,omitempty"`
	AcceptIdpInitiated *bool              `json:"accept_idp_initiated,omitempty"`
	IssuerUrl          string             `json:"issuer_url,omitempty"`
	ClientId           string             `json:"client_id,omitempty"`
	ClientSecret       string             `json:"client_secret,omitempty"`
	AttributeMappings  []AttributeMapping `json:"attribute_mappings"`
}

// UpdateIdpRequest is the body of PATCH /sso/idps/{idp_name}. name isn't
// included: the spec's UpdateIdp schema has no name property, so the idp's
// name can't be changed after creation (only via destroy/recreate).
type UpdateIdpRequest struct {
	Identifiers        []string           `json:"identifiers,omitempty"`
	MetaCert           string             `json:"meta_cert,omitempty"`
	AcceptIdpInitiated *bool              `json:"accept_idp_initiated,omitempty"`
	AttributeMappings  []AttributeMapping `json:"attribute_mappings,omitempty"`
}

// IdpsList is the response body of GET /sso/idps.
type IdpsList struct {
	Idps []Idp `json:"idps"`
}

// SsoMapping maps one identity provider attribute value to a Bronto role or
// group, per the SSO API's SsoMapping schema. Mappings have no individual
// id - the whole set is created/read/replaced/deleted together via
// /sso/mappings.
type SsoMapping struct {
	ProviderKey   string `json:"provider_key"`
	ProviderValue string `json:"provider_value"`
	AssignAttr    string `json:"assign_attr"`
}

// SsoMappingsRequest is the body of POST and PUT /sso/mappings.
type SsoMappingsRequest struct {
	Mappings []SsoMapping `json:"mappings"`
}

// DecodeIdp decodes a single Idp from a JSON response body.
func DecodeIdp(body string) (Idp, error) {
	var idp Idp
	err := json.NewDecoder(strings.NewReader(body)).Decode(&idp)
	return idp, err
}

// DecodeIdpsList decodes a {"idps": [...]} response body.
func DecodeIdpsList(body string) (IdpsList, error) {
	var list IdpsList
	err := json.NewDecoder(strings.NewReader(body)).Decode(&list)
	return list, err
}

// DecodeSsoMappingsRequest decodes a {"mappings": [...]} response body -
// the same shape for GET, POST, and PUT /sso/mappings.
func DecodeSsoMappingsRequest(body string) (SsoMappingsRequest, error) {
	var m SsoMappingsRequest
	err := json.NewDecoder(strings.NewReader(body)).Decode(&m)
	return m, err
}
