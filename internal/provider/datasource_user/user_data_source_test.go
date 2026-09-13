package datasource_user

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	bronto "terraform-provider-bronto/internal"
)

func testDataSource(t *testing.T, srv *httptest.Server) *userDataSource {
	t.Helper()
	client := &bronto.Client{
		BaseURL:    srv.URL,
		HTTPClient: srv.Client(),
		ApiKey:     "test-key",
	}
	return &userDataSource{client: client}
}

func TestFindByEmailMatches(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users" {
			t.Errorf("path = %s, want /users", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"users":[
			{"id":"1","email":"alice@bronto.io"},
			{"id":"2","email":"Bob@Bronto.io"}
		]}`))
	}))
	defer srv.Close()

	d := testDataSource(t, srv)

	user, err := d.findByEmail(context.Background(), "bob@bronto.io")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Id != "2" {
		t.Errorf("Id = %s, want 2 (case-insensitive email match)", user.Id)
	}
}

func TestFindByEmailNoMatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"users":[{"id":"1","email":"alice@bronto.io"}]}`))
	}))
	defer srv.Close()

	d := testDataSource(t, srv)

	_, err := d.findByEmail(context.Background(), "missing@bronto.io")
	if err == nil || !strings.Contains(err.Error(), "no user found") {
		t.Fatalf("expected a not-found error, got: %v", err)
	}
}

func TestFindByEmailMultipleMatches(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"users":[
			{"id":"1","email":"dup@bronto.io"},
			{"id":"2","email":"dup@bronto.io"}
		]}`))
	}))
	defer srv.Close()

	d := testDataSource(t, srv)

	_, err := d.findByEmail(context.Background(), "dup@bronto.io")
	if err == nil || !strings.Contains(err.Error(), "multiple users found") {
		t.Fatalf("expected a multiple-matches error, got: %v", err)
	}
}

func TestFindById(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/abc-123" {
			t.Errorf("path = %s, want /users/abc-123", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"abc-123","email":"alice@bronto.io"}`))
	}))
	defer srv.Close()

	d := testDataSource(t, srv)

	user, err := d.findById(context.Background(), "abc-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Email != "alice@bronto.io" {
		t.Errorf("Email = %s", user.Email)
	}
}

func TestFindByIdNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	d := testDataSource(t, srv)

	_, err := d.findById(context.Background(), "missing")
	if !bronto.IsNotFound(err) {
		t.Fatalf("expected a not-found error, got: %v", err)
	}
}

func TestToLoginMethodsList(t *testing.T) {
	list, diags := toLoginMethodsList([]bronto.LoginMethod{
		{Name: "Password", Enabled: true},
		{Name: "Okta", Enabled: false},
	})
	if diags.HasError() {
		t.Fatalf("toLoginMethodsList: %v", diags)
	}
	if list.IsNull() || list.IsUnknown() {
		t.Fatalf("expected a known, non-null list")
	}
	if len(list.Elements()) != 2 {
		t.Fatalf("got %d elements, want 2", len(list.Elements()))
	}
}

func TestToLoginMethodsListEmpty(t *testing.T) {
	list, diags := toLoginMethodsList(nil)
	if diags.HasError() {
		t.Fatalf("toLoginMethodsList: %v", diags)
	}
	if list.IsNull() {
		t.Fatalf("expected an empty but non-null list for no login methods")
	}
}

func TestPopulateModel(t *testing.T) {
	ctx := context.Background()
	invitedAt := int64(1731370728)

	var data UserModel
	diags := populateModel(ctx, &data, bronto.User{
		Id:           "1",
		Email:        "alice@bronto.io",
		FirstName:    "Alice",
		LastName:     "Smith",
		Status:       bronto.UserStatusActive,
		Roles:        []string{"Admin"},
		Tags:         map[string]string{"region": "eu"},
		LastLogins:   map[string]int64{"Password": 1731338939},
		LoginMethods: []bronto.LoginMethod{{Name: "Password", Enabled: true}},
		InvitedAt:    &invitedAt,
	})
	if diags.HasError() {
		t.Fatalf("populateModel: %v", diags)
	}

	if !data.Active.ValueBool() {
		t.Errorf("Active = false, want true for status ACTIVE")
	}
	if data.InvitedAt.ValueInt64() != invitedAt {
		t.Errorf("InvitedAt = %d, want %d", data.InvitedAt.ValueInt64(), invitedAt)
	}
	if !data.InvitedBy.IsNull() {
		t.Errorf("InvitedBy should be null when the API omits it")
	}
	if data.LoginMethods.IsNull() {
		t.Errorf("LoginMethods should not be null")
	}
}
