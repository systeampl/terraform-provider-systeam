package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateCheckSLO(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/checks/42/slo" {
			t.Errorf("expected path '/api/checks/42/slo', got '%s'", r.URL.Path)
		}

		var body CheckSLOCreateRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.SLIType != "availability" {
			t.Errorf("expected sli_type 'availability', got '%s'", body.SLIType)
		}
		if body.TargetPercentage != 99.9 {
			t.Errorf("expected target_percentage 99.9, got %f", body.TargetPercentage)
		}

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(CheckSLO{
			ID:                      1,
			CheckID:                 42,
			SLIType:                 "availability",
			TargetPercentage:        99.9,
			WindowDays:              30,
			IsActive:                true,
			NotifyOnBudgetWarn:      true,
			BudgetWarnPct:           20.0,
			NotifyOnBudgetExhausted: true,
			BurnRateAlertEnabled:    false,
			BurnRateThreshold:       14.4,
			BurnRateWindowMinutes:   60,
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	slo, err := c.CreateCheckSLO(context.Background(), 42, CheckSLOCreateRequest{
		SLIType:                 "availability",
		TargetPercentage:        99.9,
		WindowDays:              30,
		NotifyOnBudgetWarn:      true,
		BudgetWarnPct:           20.0,
		NotifyOnBudgetExhausted: true,
		BurnRateThreshold:       14.4,
		BurnRateWindowMinutes:   60,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if slo.ID != 1 {
		t.Errorf("expected ID 1, got %d", slo.ID)
	}
	if slo.CheckID != 42 {
		t.Errorf("expected CheckID 42, got %d", slo.CheckID)
	}
	if slo.TargetPercentage != 99.9 {
		t.Errorf("expected TargetPercentage 99.9, got %f", slo.TargetPercentage)
	}
}

func TestGetCheckSLO(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/checks/42/slo" {
			t.Errorf("expected path '/api/checks/42/slo', got '%s'", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(checkSLOGetResponse{
			CheckID: 42,
			SLO: &CheckSLO{
				ID:                      1,
				CheckID:                 42,
				SLIType:                 "availability",
				TargetPercentage:        99.9,
				WindowDays:              30,
				IsActive:                true,
				NotifyOnBudgetWarn:      true,
				BudgetWarnPct:           20.0,
				NotifyOnBudgetExhausted: true,
				BurnRateAlertEnabled:    false,
				BurnRateThreshold:       14.4,
				BurnRateWindowMinutes:   60,
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	slo, err := c.GetCheckSLO(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if slo.ID != 1 {
		t.Errorf("expected ID 1, got %d", slo.ID)
	}
	if slo.WindowDays != 30 {
		t.Errorf("expected WindowDays 30, got %d", slo.WindowDays)
	}
}

func TestDeleteCheckSLO(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/checks/42/slo/1" {
			t.Errorf("expected path '/api/checks/42/slo/1', got '%s'", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	err := c.DeleteCheckSLO(context.Background(), 42, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
