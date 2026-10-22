package internal

import (
	"context"
	"encoding/json"
	"fmt"
)

// TagEntity identifies the resource a tag is attached to, per the Tags
// API's Entity schema.
type TagEntity struct {
	Id   string `json:"id"`
	Type string `json:"type"`
}

// Tag is the body of both POST /tags (create) and PUT /tags (update the
// value of an existing tag - name+entity is the identity, value is what
// changes).
type Tag struct {
	Name   string    `json:"name"`
	Value  string    `json:"value"`
	Entity TagEntity `json:"entity"`
}

// DeleteTagRequest is the body of DELETE /tags.
type DeleteTagRequest struct {
	Name   string    `json:"name"`
	Entity TagEntity `json:"entity"`
}

// CreateTag creates a single tag on an entity via POST /tags.
func (c *Client) CreateTag(ctx context.Context, entityType, entityId, name, value string) error {
	body, err := json.Marshal(Tag{Name: name, Value: value, Entity: TagEntity{Id: entityId, Type: entityType}})
	if err != nil {
		return err
	}
	_, err = c.CreateRequest(ctx, body, "/tags")
	return err
}

// UpdateTag updates the value of an existing tag via PUT /tags.
func (c *Client) UpdateTag(ctx context.Context, entityType, entityId, name, value string) error {
	body, err := json.Marshal(Tag{Name: name, Value: value, Entity: TagEntity{Id: entityId, Type: entityType}})
	if err != nil {
		return err
	}
	_, err = c.PutRequestWithBody(ctx, body, "/tags")
	return err
}

// DeleteTag deletes a single tag from an entity via DELETE /tags.
func (c *Client) DeleteTag(ctx context.Context, entityType, entityId, name string) error {
	body, err := json.Marshal(DeleteTagRequest{Name: name, Entity: TagEntity{Id: entityId, Type: entityType}})
	if err != nil {
		return err
	}
	return c.DeleteRequestWithBody(ctx, body, "/tags")
}

// MergeTags merges the provider's default_tags with a resource's own tags,
// with the resource's own tags winning on key conflicts. Either map may be
// nil.
func MergeTags(defaultTags, specific map[string]string) map[string]string {
	merged := make(map[string]string, len(defaultTags)+len(specific))
	for k, v := range defaultTags {
		merged[k] = v
	}
	for k, v := range specific {
		merged[k] = v
	}
	return merged
}

// OwnTags is the inverse of MergeTags: given the full tag set currently on
// an entity, it returns the subset not explained by defaultTags - i.e. it
// excludes any key whose value exactly matches the current default_tags
// value for that key. This is a best-effort heuristic, only used when
// reconstructing "what's mine" from a bare GET (on Read, with no plan to
// compare against - Create/Update already know the resource's own tags
// directly from the plan and don't need this). If a resource's own tag
// coincidentally has both the same key and value as a default, it will be
// misclassified as default-derived and omitted here; that's an inherent
// limitation of the approach (the same one the AWS/GCP Terraform providers
// document for their own default_tags/tags_all split), not something a
// smarter algorithm can fully resolve without the API itself tracking
// tag provenance.
func OwnTags(full, defaultTags map[string]string) map[string]string {
	own := make(map[string]string, len(full))
	for k, v := range full {
		if defaultValue, isDefault := defaultTags[k]; isDefault && defaultValue == v {
			continue
		}
		own[k] = v
	}
	return own
}

// diffTags compares the currently-applied tags against the desired tags and
// returns which keys need to be created, which need their value updated,
// and which need to be removed, to reconcile the two. Tag name is unique
// per entity (per the Tags API), so it's the natural diff key.
func diffTags(current, desired map[string]string) (toCreate, toUpdate map[string]string, toDelete []string) {
	toCreate = map[string]string{}
	toUpdate = map[string]string{}

	for name, value := range desired {
		oldValue, existed := current[name]
		if !existed {
			toCreate[name] = value
		} else if oldValue != value {
			toUpdate[name] = value
		}
	}

	for name := range current {
		if _, stillWanted := desired[name]; !stillWanted {
			toDelete = append(toDelete, name)
		}
	}

	return toCreate, toUpdate, toDelete
}

// SyncTags reconciles the tags on an entity from current (as last read) to
// desired (as configured), issuing one Create/Update/Delete /tags call per
// changed key. There is no bulk "replace all tags" endpoint, so this is the
// only way to apply a tag map - shared by every resource that supports
// tags (dashboards, widgets, users; metric definitions do not support tags
// at all, per the API's Entity.type enum).
func (c *Client) SyncTags(ctx context.Context, entityType, entityId string, current, desired map[string]string) error {
	toCreate, toUpdate, toDelete := diffTags(current, desired)

	for name, value := range toCreate {
		if err := c.CreateTag(ctx, entityType, entityId, name, value); err != nil {
			return fmt.Errorf("creating tag %q: %w", name, err)
		}
	}

	for name, value := range toUpdate {
		if err := c.UpdateTag(ctx, entityType, entityId, name, value); err != nil {
			return fmt.Errorf("updating tag %q: %w", name, err)
		}
	}

	for _, name := range toDelete {
		if err := c.DeleteTag(ctx, entityType, entityId, name); err != nil {
			return fmt.Errorf("deleting tag %q: %w", name, err)
		}
	}

	return nil
}
