package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpsertAndGetLifecycleWatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			var body LifecycleWatchUpsertRequest
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body.Vendor != "openai" {
				t.Errorf("expected vendor openai, got %q", body.Vendor)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"watched": true, "id": 42})
			return
		}
		// GET list
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{"id": 42, "vendor": "openai", "resource_type": "ai_model", "notify_90d": true, "notify_7d": false},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	id, err := c.UpsertLifecycleWatch(context.Background(), 1, LifecycleWatchUpsertRequest{Vendor: "openai", ResourceType: "ai_model"})
	if err != nil || id != 42 {
		t.Fatalf("upsert: id=%d err=%v", id, err)
	}
	w2, err := c.GetLifecycleWatch(context.Background(), 1, 42)
	if err != nil || w2 == nil || w2.Vendor != "openai" || w2.Notify7d != false {
		t.Fatalf("get: %+v err=%v", w2, err)
	}
	gone, _ := c.GetLifecycleWatch(context.Background(), 1, 999)
	if gone != nil {
		t.Errorf("expected nil for missing watch")
	}
}
