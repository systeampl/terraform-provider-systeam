package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateEscalationPolicy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/organizations/1/escalation-policies" {
			t.Errorf("expected path '/api/organizations/1/escalation-policies', got '%s'", r.URL.Path)
		}

		var body EscalationPolicyCreateRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.Name != "critical-policy" {
			t.Errorf("expected name 'critical-policy', got '%s'", body.Name)
		}
		if len(body.Steps) != 1 {
			t.Fatalf("expected 1 step, got %d", len(body.Steps))
		}
		if body.Steps[0].TargetType != "user" {
			t.Errorf("expected target_type 'user', got '%s'", body.Steps[0].TargetType)
		}

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		userID := 10
		_ = json.NewEncoder(w).Encode(EscalationPolicy{
			ID:             5,
			Name:           "critical-policy",
			OrganizationID: 1,
			IsActive:       true,
			Steps: []EscalationStep{
				{StepOrder: 1, DelayMinutes: 5, TargetType: "user", TargetUserID: &userID},
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	userID := 10
	policy, err := c.CreateEscalationPolicy(context.Background(), 1, EscalationPolicyCreateRequest{
		Name:     "critical-policy",
		IsActive: true,
		Steps: []EscalationStep{
			{StepOrder: 1, DelayMinutes: 5, TargetType: "user", TargetUserID: &userID},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.ID != 5 {
		t.Errorf("expected ID 5, got %d", policy.ID)
	}
	if policy.Name != "critical-policy" {
		t.Errorf("expected name 'critical-policy', got '%s'", policy.Name)
	}
	if len(policy.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(policy.Steps))
	}
	if policy.Steps[0].TargetUserID == nil || *policy.Steps[0].TargetUserID != 10 {
		t.Error("expected step target_user_id to be 10")
	}
}

func TestGetEscalationPolicy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/organizations/1/escalation-policies/5" {
			t.Errorf("expected path '/api/organizations/1/escalation-policies/5', got '%s'", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		channelID := 3
		_ = json.NewEncoder(w).Encode(EscalationPolicy{
			ID:             5,
			Name:           "critical-policy",
			OrganizationID: 1,
			IsActive:       true,
			Steps: []EscalationStep{
				{StepOrder: 1, DelayMinutes: 0, TargetType: "channel", TargetChannelID: &channelID},
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	policy, err := c.GetEscalationPolicy(context.Background(), 1, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.ID != 5 {
		t.Errorf("expected ID 5, got %d", policy.ID)
	}
	if !policy.IsActive {
		t.Error("expected IsActive to be true")
	}
	if len(policy.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(policy.Steps))
	}
	if policy.Steps[0].TargetChannelID == nil || *policy.Steps[0].TargetChannelID != 3 {
		t.Error("expected step target_channel_id to be 3")
	}
}

func TestDeleteEscalationPolicy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/organizations/1/escalation-policies/5" {
			t.Errorf("expected path '/api/organizations/1/escalation-policies/5', got '%s'", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	err := c.DeleteEscalationPolicy(context.Background(), 1, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
