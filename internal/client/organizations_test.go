package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetOrganizationBySlug(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/organizations/by-slug/systeam" {
			t.Errorf("expected path '/api/organizations/by-slug/systeam', got '%s'", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Organization{
			ID:          1,
			Name:        "SysTeam",
			Slug:        "systeam",
			Description: "Main organization",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	org, err := c.GetOrganizationBySlug(context.Background(), "systeam")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if org.ID != 1 {
		t.Errorf("expected ID 1, got %d", org.ID)
	}
	if org.Name != "SysTeam" {
		t.Errorf("expected name 'SysTeam', got '%s'", org.Name)
	}
	if org.Slug != "systeam" {
		t.Errorf("expected slug 'systeam', got '%s'", org.Slug)
	}
	if org.Description != "Main organization" {
		t.Errorf("expected description 'Main organization', got '%s'", org.Description)
	}
}
