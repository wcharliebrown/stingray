package models

import (
	"time"
)

type Session struct {
	ID        int
	SessionID string
	UserID    string
	Username  string
	CreatedAt time.Time
	ExpiresAt time.Time
	IsActive  bool
} 