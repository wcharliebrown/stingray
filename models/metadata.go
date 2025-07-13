package models

import (
	"time"
)

// TableMetadata represents metadata for database tables
type TableMetadata struct {
	ID          int
	TableName   string
	DisplayName string
	Description string
	ReadGroups  string // JSON array of group names that can read this table
	WriteGroups string // JSON array of group names that can write to this table
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// FieldMetadata represents metadata for database table fields
type FieldMetadata struct {
	ID              int
	TableName       string
	FieldName       string
	DisplayName     string
	Description     string
	DBType          string // MySQL data type (e.g., VARCHAR, INT, TEXT)
	HTMLInputType   string // HTML input type (e.g., text, email, password, select)
	FormPosition    int    // Position in edit form (0-based)
	ListPosition    int    // Position in table listing (0-based)
	IsRequired      bool   // Whether field is required
	IsReadOnly      bool   // Whether field is read-only
	DefaultValue    string // Default value for the field
	ValidationRules string // JSON string with validation rules
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// TableRow represents a generic row from any table
type TableRow struct {
	ID   int                    `json:"id"`
	Data map[string]interface{} `json:"data"`
}

// TableData represents the structure for displaying table data
type TableData struct {
	TableName    string
	DisplayName  string
	Fields       []FieldMetadata
	Rows         []TableRow
	TotalRows    int
	CurrentPage  int
	PageSize     int
	CanEdit      bool
	CanDelete    bool
	CanCreate    bool
}

// FormData represents the structure for editing/creating table rows
type FormData struct {
	TableName    string
	DisplayName  string
	Fields       []FieldMetadata
	Row          TableRow
	IsNew        bool
	EngineerMode bool
} 