package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateService(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/organizations/1/services" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		var body ServiceCreateRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.Name != "Billing API" || body.Tier != "P1" {
			t.Errorf("unexpected body %+v", body)
		}
		if len(body.NotificationChannelIDs) != 2 {
			t.Errorf("expected 2 channel ids, got %v", body.NotificationChannelIDs)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": 5, "organization_id": 1, "name": "Billing API", "slug": "billing-api",
			"tier": "P1", "is_active": true,
			"notification_channels": []map[string]any{{"id": 7}, {"id": 9}},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	svc, err := c.CreateService(context.Background(), 1, ServiceCreateRequest{
		Name: "Billing API", Tier: "P1", NotificationChannelIDs: []int{7, 9},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if svc.ID != 5 || svc.Slug != "billing-api" {
		t.Errorf("unexpected service %+v", svc)
	}
	ids := svc.NotificationChannelIDs()
	if len(ids) != 2 || ids[0] != 7 || ids[1] != 9 {
		t.Errorf("channel ids not parsed from response objects: %v", ids)
	}
}

func TestDeleteService(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.Method != http.MethodDelete || r.URL.Path != "/api/organizations/1/services/5" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	if err := c.DeleteService(context.Background(), 1, 5); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("delete not called")
	}
}
