package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDoRequest_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("expected Authorization header 'Bearer test-token', got '%s'", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type 'application/json', got '%s'", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("expected Accept 'application/json', got '%s'", r.Header.Get("Accept"))
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"name": "test-check"})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")

	var result map[string]string
	err := c.DoRequest(context.Background(), http.MethodGet, "/api/checks/1", nil, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["name"] != "test-check" {
		t.Errorf("expected name 'test-check', got '%s'", result["name"])
	}
}

func TestDoRequest_PostWithBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "new-check" {
			t.Errorf("expected body name 'new-check', got '%s'", body["name"])
		}

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"id": 1, "name": "new-check"})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")

	var result map[string]any
	err := c.DoRequest(context.Background(), http.MethodPost, "/api/checks", map[string]string{"name": "new-check"}, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["name"] != "new-check" {
		t.Errorf("expected name 'new-check', got '%v'", result["name"])
	}
}

func TestDoRequest_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"detail": "Check not found"})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")

	err := c.DoRequest(context.Background(), http.MethodGet, "/api/checks/999", nil, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if !apiErr.IsNotFound() {
		t.Errorf("expected IsNotFound() to be true, got false (status=%d)", apiErr.StatusCode)
	}
	if apiErr.Message != "Check not found" {
		t.Errorf("expected message 'Check not found', got '%s'", apiErr.Message)
	}
}

func TestDoRequest_ValidationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"detail": []map[string]any{
				{"loc": []string{"body", "name"}, "msg": "field required", "type": "value_error.missing"},
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")

	err := c.DoRequest(context.Background(), http.MethodPost, "/api/checks", map[string]string{}, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("expected status 422, got %d", apiErr.StatusCode)
	}
	if apiErr.Detail == nil {
		t.Error("expected Detail to be non-nil for structured validation error")
	}
}

func TestDoRequest_AuthError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"detail": "Invalid token"})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")

	err := c.DoRequest(context.Background(), http.MethodGet, "/api/checks", nil, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", apiErr.StatusCode)
	}
	if apiErr.Message != "Invalid token" {
		t.Errorf("expected message 'Invalid token', got '%s'", apiErr.Message)
	}
}

func TestDoRequest_NoResultPointer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")

	err := c.DoRequest(context.Background(), http.MethodDelete, "/api/checks/1", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
