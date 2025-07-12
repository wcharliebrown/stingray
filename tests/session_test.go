package tests

import (
	"testing"
	"time"
	"stingray/models"
	"stingray/database"
)

func TestSessionOperations(t *testing.T) {
	// This test requires a database connection
	// You would need to set up a test database or mock the database operations
	
	// Test session creation
	t.Run("CreateSession", func(t *testing.T) {
		// This would test session creation
		// For now, we'll just test the session model structure
		session := &models.Session{
			ID:        1,
			SessionID: "test_session_id",
			UserID:    "admin",
			Username:  "admin",
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(24 * time.Hour),
			IsActive:  true,
		}

		if session.SessionID != "test_session_id" {
			t.Errorf("Expected session ID to be 'test_session_id', got %s", session.SessionID)
		}

		if session.Username != "admin" {
			t.Errorf("Expected username to be 'admin', got %s", session.Username)
		}

		if !session.IsActive {
			t.Error("Expected session to be active")
		}
	})

	t.Run("SessionExpiration", func(t *testing.T) {
		// Test session expiration logic
		now := time.Now()
		expiredSession := &models.Session{
			ExpiresAt: now.Add(-1 * time.Hour), // Expired 1 hour ago
		}

		if !expiredSession.ExpiresAt.Before(now) {
			t.Error("Expected session to be expired")
		}

		activeSession := &models.Session{
			ExpiresAt: now.Add(1 * time.Hour), // Expires in 1 hour
		}

		if activeSession.ExpiresAt.Before(now) {
			t.Error("Expected session to be active")
		}
	})
}

func TestSessionIDGeneration(t *testing.T) {
	id1, err1 := database.GenerateSessionIDForTest()
	if err1 != nil {
		t.Fatalf("Error generating session ID: %v", err1)
	}
	id2, err2 := database.GenerateSessionIDForTest()
	if err2 != nil {
		t.Fatalf("Error generating session ID: %v", err2)
	}

	if len(id1) != 64 {
		t.Errorf("Expected session ID length to be 64, got %d", len(id1))
	}
	if len(id2) != 64 {
		t.Errorf("Expected session ID length to be 64, got %d", len(id2))
	}
	if id1 == id2 {
		t.Error("Expected different session IDs to be different")
	}
} 