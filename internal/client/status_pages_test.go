package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateStatusPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/status-pages" {
			t.Errorf("expected path /api/status-pages, got %s", r.URL.Path)
		}

		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "My Status Page" {
			t.Errorf("expected name 'My Status Page', got '%v'", body["name"])
		}
		if body["slug"] != "my-status" {
			t.Errorf("expected slug 'my-status', got '%v'", body["slug"])
		}

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(StatusPage{
			ID:       1,
			Name:     "My Status Page",
			Slug:     "my-status",
			IsPublic: true,
			IsActive: true,
			CheckIDs: []int{10, 20},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	page, err := c.CreateStatusPage(context.Background(), StatusPageCreateRequest{
		Name:     "My Status Page",
		Slug:     "my-status",
		IsPublic: true,
		CheckIDs: []int{10, 20},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.ID != 1 {
		t.Errorf("expected ID 1, got %d", page.ID)
	}
	if page.Name != "My Status Page" {
		t.Errorf("expected name 'My Status Page', got '%s'", page.Name)
	}
	if len(page.CheckIDs) != 2 {
		t.Errorf("expected 2 check IDs, got %d", len(page.CheckIDs))
	}
}

func TestGetStatusPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/status-pages/5" {
			t.Errorf("expected path /api/status-pages/5, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(StatusPage{
			ID:           5,
			Name:         "Production Status",
			Slug:         "prod-status",
			Description:  "Production services",
			IsPublic:     true,
			IsActive:     true,
			CustomDomain: "status.example.com",
			CheckIDs:     []int{1, 2, 3},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	page, err := c.GetStatusPage(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.ID != 5 {
		t.Errorf("expected ID 5, got %d", page.ID)
	}
	if page.Slug != "prod-status" {
		t.Errorf("expected slug 'prod-status', got '%s'", page.Slug)
	}
	if page.CustomDomain != "status.example.com" {
		t.Errorf("expected custom_domain 'status.example.com', got '%s'", page.CustomDomain)
	}
	if len(page.CheckIDs) != 3 {
		t.Errorf("expected 3 check IDs, got %d", len(page.CheckIDs))
	}
}

func TestDeleteStatusPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/status-pages/3" {
			t.Errorf("expected path /api/status-pages/3, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	err := c.DeleteStatusPage(context.Background(), 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
