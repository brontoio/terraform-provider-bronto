package internal

import (
	"encoding/json"
	"strings"
)

// UserGroup is the wire representation of the Groups API's UserGroup
// schema, returned by create/get/list/update.
type UserGroup struct {
	GroupId     string `json:"group_id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	CreatedAt   int64  `json:"created_at,omitempty"`
	ModifiedAt  int64  `json:"modified_at,omitempty"`
}

// GroupRequest is the body of both POST /groups (create) and
// PATCH /groups/{groupId} (update) - both accept the same shape.
type GroupRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// UserGroupList is the response body of GET /groups.
type UserGroupList struct {
	Groups []UserGroup `json:"groups"`
}

// GroupMemberRef identifies one member of a group in an add/remove request.
type GroupMemberRef struct {
	MemberType string `json:"member_type"`
	MemberId   string `json:"member_id"`
}

// GroupMember is one entry of the GET /groups/{groupId}/members response -
// the same as GroupMemberRef plus the group it belongs to and when it was
// added.
type GroupMember struct {
	GroupId    string `json:"group_id"`
	MemberId   string `json:"member_id"`
	MemberType string `json:"member_type"`
	ModifiedAt int64  `json:"modified_at"`
}

// GroupMembersList is the response body of GET /groups/{groupId}/members.
type GroupMembersList struct {
	Members []GroupMember `json:"members"`
}

// GroupMembersRequest is the body of both POST /groups/{groupId}/members
// (add) and DELETE /groups/{groupId}/members (remove).
type GroupMembersRequest struct {
	Members []GroupMemberRef `json:"members"`
}

// DecodeUserGroup decodes a single UserGroup from a JSON response body.
func DecodeUserGroup(body string) (UserGroup, error) {
	var g UserGroup
	err := json.NewDecoder(strings.NewReader(body)).Decode(&g)
	return g, err
}

// DecodeUserGroupList decodes a {"groups": [...]} response body.
func DecodeUserGroupList(body string) (UserGroupList, error) {
	var list UserGroupList
	err := json.NewDecoder(strings.NewReader(body)).Decode(&list)
	return list, err
}

// DecodeGroupMembersList decodes a {"members": [...]} response body.
func DecodeGroupMembersList(body string) (GroupMembersList, error) {
	var list GroupMembersList
	err := json.NewDecoder(strings.NewReader(body)).Decode(&list)
	return list, err
}
