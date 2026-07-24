package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateTeam(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/organizations/1/teams" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		var body TeamCreateRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.Name != "SRE" {
			t.Errorf("expected name 'SRE', got %q", body.Name)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": 3, "organization_id": 1, "name": "SRE", "slug": "sre",
			"description": "Site reliability", "is_active": true, "member_count": 1,
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	team, err := c.CreateTeam(context.Background(), 1, TeamCreateRequest{Name: "SRE", Description: "Site reliability"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if team.ID != 3 || team.Slug != "sre" {
		t.Errorf("unexpected team %+v", team)
	}
}

func TestUpdateTeam(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/api/organizations/1/teams/3" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		var body TeamUpdateRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": 3, "organization_id": 1, "name": body.Name, "slug": "sre",
			"description": body.Description, "is_active": true, "member_count": 2,
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	team, err := c.UpdateTeam(context.Background(), 1, 3, TeamUpdateRequest{Name: "SRE Team", Description: "updated"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if team.Name != "SRE Team" || team.Description != "updated" {
		t.Errorf("update not reflected: %+v", team)
	}
}

func TestDeleteTeam(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.Method != http.MethodDelete || r.URL.Path != "/api/organizations/1/teams/3" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	if err := c.DeleteTeam(context.Background(), 1, 3); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("delete not called")
	}
}
