package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateMaintenanceWindow(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/maintenance" {
			t.Errorf("expected path '/api/maintenance', got '%s'", r.URL.Path)
		}

		var body MaintenanceWindowCreateRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.Name != "nightly-deploy" {
			t.Errorf("expected name 'nightly-deploy', got '%s'", body.Name)
		}
		if len(body.CheckIDs) != 2 {
			t.Errorf("expected 2 check_ids, got %d", len(body.CheckIDs))
		}

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(MaintenanceWindow{
			ID:          1,
			Name:        "nightly-deploy",
			Description: "Nightly deployment window",
			StartTime:   "2026-03-11T02:00:00Z",
			EndTime:     "2026-03-11T04:00:00Z",
			Timezone:    "UTC",
			CheckIDs:    []int{10, 20},
			ProjectIDs:  []int{},
			IsRecurring: true,
			IsActive:    true,
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	mw, err := c.CreateMaintenanceWindow(context.Background(), MaintenanceWindowCreateRequest{
		Name:        "nightly-deploy",
		Description: "Nightly deployment window",
		StartTime:   "2026-03-11T02:00:00Z",
		EndTime:     "2026-03-11T04:00:00Z",
		Timezone:    "UTC",
		CheckIDs:    []int{10, 20},
		IsRecurring: true,
		IsActive:    true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mw.ID != 1 {
		t.Errorf("expected ID 1, got %d", mw.ID)
	}
	if mw.Name != "nightly-deploy" {
		t.Errorf("expected name 'nightly-deploy', got '%s'", mw.Name)
	}
	if !mw.IsRecurring {
		t.Error("expected IsRecurring to be true")
	}
}

func TestGetMaintenanceWindow(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/maintenance/1" {
			t.Errorf("expected path '/api/maintenance/1', got '%s'", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(MaintenanceWindow{
			ID:          1,
			Name:        "nightly-deploy",
			Description: "Nightly deployment window",
			StartTime:   "2026-03-11T02:00:00Z",
			EndTime:     "2026-03-11T04:00:00Z",
			Timezone:    "UTC",
			CheckIDs:    []int{10, 20},
			ProjectIDs:  []int{5},
			IsRecurring: false,
			IsActive:    true,
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	mw, err := c.GetMaintenanceWindow(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mw.ID != 1 {
		t.Errorf("expected ID 1, got %d", mw.ID)
	}
	if len(mw.ProjectIDs) != 1 || mw.ProjectIDs[0] != 5 {
		t.Errorf("expected project_ids [5], got %v", mw.ProjectIDs)
	}
}

func TestDeleteMaintenanceWindow(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/maintenance/1" {
			t.Errorf("expected path '/api/maintenance/1', got '%s'", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	err := c.DeleteMaintenanceWindow(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
