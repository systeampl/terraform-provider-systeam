package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateAgentRegistrationToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/agents/registration-token" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		var body AgentRegistrationTokenCreateRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.Name != "laptop" {
			t.Errorf("expected name 'laptop', got %q", body.Name)
		}
		if body.Mode != "private" {
			t.Errorf("expected mode 'private', got %q", body.Mode)
		}
		if body.OrganizationID != 1 {
			t.Errorf("expected org 1, got %d", body.OrganizationID)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"token":      "sekret-enrollment-token",
			"expires_at": "2026-07-23T18:23:13Z",
			"mode":       "private",
			"message":    "Use this token to register a new agent. Valid for 1 hour.",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	tok, err := c.CreateAgentRegistrationToken(context.Background(), AgentRegistrationTokenCreateRequest{
		Name:           "laptop",
		Mode:           "private",
		OrganizationID: 1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok.Token != "sekret-enrollment-token" {
		t.Errorf("token not captured, got %q", tok.Token)
	}
	if tok.ExpiresAt != "2026-07-23T18:23:13Z" {
		t.Errorf("expires_at not captured, got %q", tok.ExpiresAt)
	}
	if tok.Mode != "private" {
		t.Errorf("mode not captured, got %q", tok.Mode)
	}
}
