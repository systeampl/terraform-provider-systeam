package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateCheck(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/checks" {
			t.Errorf("expected path /api/checks, got %s", r.URL.Path)
		}

		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "My HTTP Check" {
			t.Errorf("expected name 'My HTTP Check', got '%v'", body["name"])
		}
		if body["type"] != "uptime" {
			t.Errorf("expected type 'uptime', got '%v'", body["type"])
		}

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Check{
			ID:        1,
			Name:      "My HTTP Check",
			Type:      "uptime",
			ProjectID: 10,
			IsActive:  true,
			URL:       "https://example.com",
			Interval:  60,
			Timeout:   10,
			SSLVerify: true,
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	check, err := c.CreateCheck(context.Background(), CheckCreateRequest{
		Name:      "My HTTP Check",
		Type:      "uptime",
		ProjectID: 10,
		IsActive:  true,
		URL:       "https://example.com",
		Interval:  60,
		Timeout:   10,
		SSLVerify: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if check.ID != 1 {
		t.Errorf("expected ID 1, got %d", check.ID)
	}
	if check.Name != "My HTTP Check" {
		t.Errorf("expected name 'My HTTP Check', got '%s'", check.Name)
	}
	if check.Type != "uptime" {
		t.Errorf("expected type 'uptime', got '%s'", check.Type)
	}
}

func TestCreateDatabaseCheck(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["type"] != "database" {
			t.Errorf("expected type 'database', got '%v'", body["type"])
		}
		if body["db_type"] != "postgresql" {
			t.Errorf("expected db_type 'postgresql', got '%v'", body["db_type"])
		}
		if body["db_host"] != "db.internal" {
			t.Errorf("expected db_host 'db.internal', got '%v'", body["db_host"])
		}
		if body["db_port"].(float64) != 5432 {
			t.Errorf("expected db_port 5432, got '%v'", body["db_port"])
		}
		// Write-only secret must be sent on create...
		if body["db_password"] != "s3cr3t" {
			t.Errorf("expected db_password to be sent, got '%v'", body["db_password"])
		}
		if body["runbook_url"] != "https://wiki/runbook" {
			t.Errorf("expected runbook_url, got '%v'", body["runbook_url"])
		}

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		// ...but the API never returns it: response omits db_password.
		_ = json.NewEncoder(w).Encode(Check{
			ID:         3,
			Name:       "PG Check",
			Type:       "database",
			ProjectID:  10,
			IsActive:   true,
			DBType:     "postgresql",
			DBHost:     "db.internal",
			DBPort:     5432,
			DBName:     "app",
			DBUsername: "monitor",
			DBQuery:    "SELECT 1",
			RunbookURL: "https://wiki/runbook",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	check, err := c.CreateCheck(context.Background(), CheckCreateRequest{
		Name:       "PG Check",
		Type:       "database",
		ProjectID:  10,
		IsActive:   true,
		DBType:     "postgresql",
		DBHost:     "db.internal",
		DBPort:     5432,
		DBName:     "app",
		DBUsername: "monitor",
		DBPassword: "s3cr3t",
		DBQuery:    "SELECT 1",
		RunbookURL: "https://wiki/runbook",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if check.DBType != "postgresql" || check.DBHost != "db.internal" || check.DBPort != 5432 {
		t.Errorf("db fields not mapped back: %+v", check)
	}
	if check.RunbookURL != "https://wiki/runbook" {
		t.Errorf("expected runbook_url mapped back, got '%s'", check.RunbookURL)
	}
}

func TestGetCheck(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/checks/42" {
			t.Errorf("expected path /api/checks/42, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Check{
			ID:                   42,
			Name:                 "DNS Check",
			Type:                 "dns",
			ProjectID:            5,
			IsActive:             true,
			Host:                 "example.com",
			DNSServer:            "8.8.8.8",
			DNSRecordType:        "A",
			DNSExpectedValue:     "93.184.216.34",
			GeoMonitoringEnabled: true,
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	check, err := c.GetCheck(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if check.ID != 42 {
		t.Errorf("expected ID 42, got %d", check.ID)
	}
	if check.Type != "dns" {
		t.Errorf("expected type 'dns', got '%s'", check.Type)
	}
	if check.DNSServer != "8.8.8.8" {
		t.Errorf("expected dns_server '8.8.8.8', got '%s'", check.DNSServer)
	}
	if !check.GeoMonitoringEnabled {
		t.Error("expected geo_monitoring_enabled to be true")
	}
}

func TestDeleteCheck(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/checks/7" {
			t.Errorf("expected path /api/checks/7, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	err := c.DeleteCheck(context.Background(), 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
