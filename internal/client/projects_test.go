package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateProject(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/organizations/1/projects" {
			t.Errorf("expected path '/api/organizations/1/projects', got '%s'", r.URL.Path)
		}

		var body ProjectCreateRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.Name != "my-project" {
			t.Errorf("expected name 'my-project', got '%s'", body.Name)
		}
		if body.OrganizationID != 1 {
			t.Errorf("expected organization_id 1, got %d", body.OrganizationID)
		}

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Project{
			ID:                   42,
			Name:                 "my-project",
			Description:          "A test project",
			OrganizationID:       1,
			AccessControlEnabled: false,
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	project, err := c.CreateProject(context.Background(), 1, ProjectCreateRequest{
		Name:           "my-project",
		Description:    "A test project",
		OrganizationID: 1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if project.ID != 42 {
		t.Errorf("expected ID 42, got %d", project.ID)
	}
	if project.Name != "my-project" {
		t.Errorf("expected name 'my-project', got '%s'", project.Name)
	}
}

func TestGetProject(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/organizations/1/projects/42" {
			t.Errorf("expected path '/api/organizations/1/projects/42', got '%s'", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Project{
			ID:                   42,
			Name:                 "my-project",
			Description:          "A test project",
			OrganizationID:       1,
			AccessControlEnabled: true,
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	project, err := c.GetProject(context.Background(), 1, 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if project.ID != 42 {
		t.Errorf("expected ID 42, got %d", project.ID)
	}
	if !project.AccessControlEnabled {
		t.Error("expected AccessControlEnabled to be true")
	}
}

func TestDeleteProject(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/organizations/1/projects/42" {
			t.Errorf("expected path '/api/organizations/1/projects/42', got '%s'", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	err := c.DeleteProject(context.Background(), 1, 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
