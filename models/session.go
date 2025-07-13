package models

import (
	"time"
)

type Session struct {
	ID          int
	SessionID   string
	UserID      int
	Username    string
	ReadGroups  string
	WriteGroups string
	CreatedAt   time.Time // This will map to 'created' in the database
	ExpiresAt   time.Time
	IsActive    bool
} 