package models

import (
	"time"
)

type PasswordResetToken struct {
	ID        int
	UserID    int
	Token     string
	Email     string
	ExpiresAt time.Time
	Used      bool
	CreatedAt time.Time
	UpdatedAt time.Time
} 