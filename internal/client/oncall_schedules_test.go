package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateOnCallSchedule(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/organizations/1/oncall-schedules" {
			t.Errorf("expected path '/api/organizations/1/oncall-schedules', got '%s'", r.URL.Path)
		}

		var body OnCallScheduleCreateRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.Name != "primary-rotation" {
			t.Errorf("expected name 'primary-rotation', got '%s'", body.Name)
		}
		if body.RotationType != "WEEKLY" {
			t.Errorf("expected rotation_type 'WEEKLY', got '%s'", body.RotationType)
		}
		if len(body.Participants) != 2 {
			t.Fatalf("expected 2 participants, got %d", len(body.Participants))
		}
		if body.Participants[0].UserID != 10 {
			t.Errorf("expected first participant user_id 10, got %d", body.Participants[0].UserID)
		}

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(OnCallSchedule{
			ID:             3,
			Name:           "primary-rotation",
			OrganizationID: 1,
			Timezone:       "Europe/Warsaw",
			RotationType:   "WEEKLY",
			IsActive:       true,
			Participants: []OnCallParticipant{
				{ID: 1, UserID: 10, Position: 0},
				{ID: 2, UserID: 20, Position: 1},
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	schedule, err := c.CreateOnCallSchedule(context.Background(), 1, OnCallScheduleCreateRequest{
		Name:         "primary-rotation",
		Timezone:     "Europe/Warsaw",
		RotationType: "WEEKLY",
		IsActive:     true,
		Participants: []OnCallParticipant{
			{UserID: 10, Position: 0},
			{UserID: 20, Position: 1},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if schedule.ID != 3 {
		t.Errorf("expected ID 3, got %d", schedule.ID)
	}
	if schedule.Name != "primary-rotation" {
		t.Errorf("expected name 'primary-rotation', got '%s'", schedule.Name)
	}
	if len(schedule.Participants) != 2 {
		t.Fatalf("expected 2 participants, got %d", len(schedule.Participants))
	}
	if schedule.Participants[1].UserID != 20 {
		t.Errorf("expected second participant user_id 20, got %d", schedule.Participants[1].UserID)
	}
}

func TestGetOnCallSchedule(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/organizations/1/oncall-schedules/3" {
			t.Errorf("expected path '/api/organizations/1/oncall-schedules/3', got '%s'", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(OnCallSchedule{
			ID:             3,
			Name:           "primary-rotation",
			OrganizationID: 1,
			Timezone:       "UTC",
			RotationType:   "DAILY",
			IsActive:       true,
			Participants: []OnCallParticipant{
				{ID: 1, UserID: 10, Position: 0},
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	schedule, err := c.GetOnCallSchedule(context.Background(), 1, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if schedule.ID != 3 {
		t.Errorf("expected ID 3, got %d", schedule.ID)
	}
	if schedule.RotationType != "DAILY" {
		t.Errorf("expected rotation_type 'DAILY', got '%s'", schedule.RotationType)
	}
	if !schedule.IsActive {
		t.Error("expected IsActive to be true")
	}
	if len(schedule.Participants) != 1 {
		t.Fatalf("expected 1 participant, got %d", len(schedule.Participants))
	}
	if schedule.Participants[0].UserID != 10 {
		t.Errorf("expected participant user_id 10, got %d", schedule.Participants[0].UserID)
	}
}

func TestDeleteOnCallSchedule(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/organizations/1/oncall-schedules/3" {
			t.Errorf("expected path '/api/organizations/1/oncall-schedules/3', got '%s'", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	err := c.DeleteOnCallSchedule(context.Background(), 1, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
