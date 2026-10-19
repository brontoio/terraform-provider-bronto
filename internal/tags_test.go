package internal

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDiffTagsCreateUpdateDeleteNoOp(t *testing.T) {
	current := map[string]string{
		"unchanged": "same",
		"changed":   "old",
		"removed":   "gone",
	}
	desired := map[string]string{
		"unchanged": "same",
		"changed":   "new",
		"added":     "value",
	}

	toCreate, toUpdate, toDelete := diffTags(current, desired)

	if len(toCreate) != 1 || toCreate["added"] != "value" {
		t.Errorf("toCreate = %v, want {added: value}", toCreate)
	}
	if len(toUpdate) != 1 || toUpdate["changed"] != "new" {
		t.Errorf("toUpdate = %v, want {changed: new}", toUpdate)
	}
	if len(toDelete) != 1 || toDelete[0] != "removed" {
		t.Errorf("toDelete = %v, want [removed]", toDelete)
	}
}

func TestDiffTagsEmptyBoth(t *testing.T) {
	toCreate, toUpdate, toDelete := diffTags(nil, nil)
	if len(toCreate) != 0 || len(toUpdate) != 0 || len(toDelete) != 0 {
		t.Errorf("expected no changes, got create=%v update=%v delete=%v", toCreate, toUpdate, toDelete)
	}
}

func TestCreateTagSendsExpectedBody(t *testing.T) {
	var gotBody Tag
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/tags" {
			t.Errorf("request = %s %s, want POST /tags", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := testClient(t, srv)
	if err := c.CreateTag(context.Background(), "dashboard", "dash-1", "team", "observability"); err != nil {
		t.Fatalf("CreateTag: %v", err)
	}

	want := Tag{Name: "team", Value: "observability", Entity: TagEntity{Id: "dash-1", Type: "dashboard"}}
	if gotBody != want {
		t.Errorf("request body = %+v, want %+v", gotBody, want)
	}
}

func TestUpdateTagUsesPutAndExpects200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method = %s, want PUT", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := testClient(t, srv)
	if err := c.UpdateTag(context.Background(), "widget", "widget-1", "team", "platform"); err != nil {
		t.Fatalf("UpdateTag: %v", err)
	}
}

func TestDeleteTagSendsBodyAndExpects204(t *testing.T) {
	var gotBody DeleteTagRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := testClient(t, srv)
	if err := c.DeleteTag(context.Background(), "user", "user-1", "team"); err != nil {
		t.Fatalf("DeleteTag: %v", err)
	}

	want := DeleteTagRequest{Name: "team", Entity: TagEntity{Id: "user-1", Type: "user"}}
	if gotBody != want {
		t.Errorf("request body = %+v, want %+v", gotBody, want)
	}
}

func TestSyncTagsIssuesOneCallPerChange(t *testing.T) {
	var creates, updates, deletes int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			creates++
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{}`))
		case http.MethodPut:
			updates++
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{}`))
		case http.MethodDelete:
			deletes++
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer srv.Close()

	c := testClient(t, srv)
	current := map[string]string{"changed": "old", "removed": "gone"}
	desired := map[string]string{"changed": "new", "added": "value"}

	if err := c.SyncTags(context.Background(), "dashboard", "dash-1", current, desired); err != nil {
		t.Fatalf("SyncTags: %v", err)
	}

	if creates != 1 || updates != 1 || deletes != 1 {
		t.Errorf("creates=%d updates=%d deletes=%d, want 1 each", creates, updates, deletes)
	}
}
