package models

import (
	"time"
)

type Session struct {
	ID        int
	SessionID string
	UserID    int
	Username  string
	CreatedAt time.Time
	ExpiresAt time.Time
	IsActive  bool
} 