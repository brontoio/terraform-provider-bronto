package internal

import (
	"encoding/json"
	"strings"
)

// LoginMethod mirrors the UserLoginMethod schema, used both in the User
// response and in the login_methods field of an update request.
type LoginMethod struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

// User is the wire representation of the Bronto Users API's User schema,
// shared by the bronto_user resource and data source.
type User struct {
	Id           string            `json:"id"`
	FirstName    string            `json:"first_name"`
	LastName     string            `json:"last_name"`
	Email        string            `json:"email"`
	Roles        []string          `json:"roles"`
	Status       string            `json:"status"`
	Tags         map[string]string `json:"tags"`
	LastLogins   map[string]int64  `json:"last_logins"`
	LoginMethods []LoginMethod     `json:"login_methods"`
	InvitedAt    *int64            `json:"invited_at"`
	InvitedBy    *string           `json:"invited_by"`
}

// UserStatusActive is the status value the API reports for an active user.
const UserStatusActive = "ACTIVE"

// UsersList is the response body of GET /users and GET /users/all.
type UsersList struct {
	Users []User `json:"users"`
}

// SetActiveResponse is the response body of the deactivate/reactivate action
// endpoints. Despite what the OpenAPI spec documents, these do NOT return a
// full User object in practice - just a success flag and the new status
// (confirmed against the real staging API).
type SetActiveResponse struct {
	Success bool   `json:"success"`
	Status  string `json:"status"`
}

// DecodeSetActiveResponse decodes a deactivate/reactivate response body.
func DecodeSetActiveResponse(body string) (SetActiveResponse, error) {
	var r SetActiveResponse
	err := json.NewDecoder(strings.NewReader(body)).Decode(&r)
	return r, err
}

// DecodeUser decodes a single User from a JSON response body.
func DecodeUser(body string) (User, error) {
	var user User
	err := json.NewDecoder(strings.NewReader(body)).Decode(&user)
	return user, err
}

// DecodeUsersList decodes a {"users": [...]} response body.
func DecodeUsersList(body string) (UsersList, error) {
	var list UsersList
	err := json.NewDecoder(strings.NewReader(body)).Decode(&list)
	return list, err
}
