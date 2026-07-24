package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateNotificationChannel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/notification-channels" {
			t.Errorf("expected path /api/notification-channels, got %s", r.URL.Path)
		}

		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "Slack Alerts" {
			t.Errorf("expected name 'Slack Alerts', got '%v'", body["name"])
		}
		if body["channel_type"] != "slack" {
			t.Errorf("expected channel_type 'slack', got '%v'", body["channel_type"])
		}

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(NotificationChannel{
			ID:          1,
			Name:        "Slack Alerts",
			ChannelType: "slack",
			Config:      map[string]any{"webhook_url": "https://hooks.slack.com/test"},
			IsActive:    true,
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	channel, err := c.CreateNotificationChannel(context.Background(), NotificationChannelCreateRequest{
		Name:        "Slack Alerts",
		ChannelType: "slack",
		Config:      map[string]any{"webhook_url": "https://hooks.slack.com/test"},
		IsActive:    true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if channel.ID != 1 {
		t.Errorf("expected ID 1, got %d", channel.ID)
	}
	if channel.Name != "Slack Alerts" {
		t.Errorf("expected name 'Slack Alerts', got '%s'", channel.Name)
	}
	if channel.ChannelType != "slack" {
		t.Errorf("expected channel_type 'slack', got '%s'", channel.ChannelType)
	}
}

func TestGetNotificationChannel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/notification-channels/5" {
			t.Errorf("expected path /api/notification-channels/5, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(NotificationChannel{
			ID:          5,
			Name:        "Email Channel",
			ChannelType: "email",
			Config:      map[string]any{"email": "alerts@example.com"},
			IsActive:    true,
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	channel, err := c.GetNotificationChannel(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if channel.ID != 5 {
		t.Errorf("expected ID 5, got %d", channel.ID)
	}
	if channel.ChannelType != "email" {
		t.Errorf("expected channel_type 'email', got '%s'", channel.ChannelType)
	}
}

func TestDeleteNotificationChannel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/notification-channels/3" {
			t.Errorf("expected path /api/notification-channels/3, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	err := c.DeleteNotificationChannel(context.Background(), 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
