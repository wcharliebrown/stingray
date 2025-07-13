package models

import (
	"time"
)

type User struct {
	ID        int
	Username  string
	Email     string
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time // This will map to 'modified' in the database
}

type Group struct {
	ID          int
	Name        string
	Description string
	CreatedAt   time.Time // This will map to 'created' in the database
}

type UserGroup struct {
	ID      int
	UserID  int
	GroupID int
} 