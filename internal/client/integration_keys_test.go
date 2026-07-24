package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateIntegrationKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/organizations/1/integration-keys" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		var body IntegrationKeyCreateRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.Name != "grafana-prod" {
			t.Errorf("expected name 'grafana-prod', got %q", body.Name)
		}
		if body.EscalationPolicyID != 131 {
			t.Errorf("expected escalation_policy_id 131, got %d", body.EscalationPolicyID)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		// The create response is the only place the raw token appears.
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":                      7,
			"name":                    "grafana-prod",
			"token_prefix":            "ik_abcdefghij",
			"escalation_policy_id":    131,
			"is_active":               true,
			"grouping_type":           "none",
			"grouping_window_seconds": 300,
			"key":                     "ik_abcdefghijSECRETrest",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	key, err := c.CreateIntegrationKey(context.Background(), 1, IntegrationKeyCreateRequest{
		Name:               "grafana-prod",
		EscalationPolicyID: 131,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key.ID != 7 {
		t.Errorf("expected id 7, got %d", key.ID)
	}
	if key.Token != "ik_abcdefghijSECRETrest" {
		t.Errorf("token not captured from create response, got %q", key.Token)
	}
	if key.TokenPrefix != "ik_abcdefghij" {
		t.Errorf("expected token_prefix, got %q", key.TokenPrefix)
	}
}

func TestGetIntegrationKeyFindsByIDInList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{"id": 6, "name": "other", "escalation_policy_id": 1, "is_active": true},
			{"id": 7, "name": "grafana-prod", "token_prefix": "ik_abcdefghij", "escalation_policy_id": 131, "is_active": true, "grouping_type": "none", "grouping_window_seconds": 300},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	key, err := c.GetIntegrationKey(context.Background(), 1, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key == nil {
		t.Fatal("expected to find key 7, got nil")
	}
	if key.Name != "grafana-prod" {
		t.Errorf("wrong key returned: %q", key.Name)
	}
	// A key that isn't in the list reads as gone (nil, nil) — the resource drops it.
	gone, err := c.GetIntegrationKey(context.Background(), 1, 999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gone != nil {
		t.Errorf("expected nil for a missing key, got %+v", gone)
	}
}

func TestGetIntegrationKeyTreatsRevokedAsGone(t *testing.T) {
	// The backend soft-revokes (is_active=false) but still lists the key. A
	// revoked key is unusable, so GetIntegrationKey must report it as gone so the
	// resource recreates it instead of trusting a dead key.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{"id": 7, "name": "revoked-key", "token_prefix": "ik_abcdefghij", "escalation_policy_id": 131, "is_active": false},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	key, err := c.GetIntegrationKey(context.Background(), 1, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != nil {
		t.Errorf("expected a revoked key to read as gone (nil), got %+v", key)
	}
}

func TestDeleteIntegrationKey(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/organizations/1/integration-keys/7" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	if err := c.DeleteIntegrationKey(context.Background(), 1, 7); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("delete endpoint was not called")
	}
}
