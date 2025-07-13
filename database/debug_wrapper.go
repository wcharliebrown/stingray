package database

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"time"
)

// DebugDB wraps a sql.DB to log all queries when debugging is enabled
type DebugDB struct {
	db            *sql.DB
	debuggingMode bool
	logger        *log.Logger
	logFile       *os.File
}

// NewDebugDB creates a new debug database wrapper
func NewDebugDB(db *sql.DB, debuggingMode bool) *DebugDB {
	debugDB := &DebugDB{
		db:            db,
		debuggingMode: debuggingMode,
	}

	if debuggingMode {
		debugDB.setupQueryLogger()
	}

	return debugDB
}

// setupQueryLogger initializes the query logger
func (d *DebugDB) setupQueryLogger() {
	logDir := "logs"
	logFile := filepath.Join(logDir, "db_query.log")
	
	// Create logs directory if it doesn't exist
	if _, statErr := os.Stat(logDir); os.IsNotExist(statErr) {
		os.MkdirAll(logDir, 0755)
	}

	// Open log file
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Failed to open query log file: %v", err)
		return
	}

	d.logFile = file
	d.logger = log.New(file, "", log.LstdFlags)
}

// logQuery logs a database query with timestamp
func (d *DebugDB) logQuery(query string, args ...interface{}) {
	if !d.debuggingMode || d.logger == nil {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	d.logger.Printf("[%s] QUERY: %s", timestamp, query)
	if len(args) > 0 {
		d.logger.Printf("[%s] ARGS: %v", timestamp, args)
	}
	d.logger.Println() // Empty line for readability
}

// Exec logs and executes the query
func (d *DebugDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	d.logQuery(query, args...)
	return d.db.Exec(query, args...)
}

// Query logs and executes the query
func (d *DebugDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	d.logQuery(query, args...)
	return d.db.Query(query, args...)
}

// QueryRow logs and executes the query
func (d *DebugDB) QueryRow(query string, args ...interface{}) *sql.Row {
	d.logQuery(query, args...)
	return d.db.QueryRow(query, args...)
}

// Prepare logs and prepares the statement
func (d *DebugDB) Prepare(query string) (*sql.Stmt, error) {
	d.logQuery("PREPARE: " + query)
	return d.db.Prepare(query)
}

// Begin logs and begins a transaction
func (d *DebugDB) Begin() (*sql.Tx, error) {
	if d.debuggingMode && d.logger != nil {
		timestamp := time.Now().Format("2006-01-02 15:04:05.000")
		d.logger.Printf("[%s] BEGIN TRANSACTION", timestamp)
		d.logger.Println()
	}
	return d.db.Begin()
}

// Close closes the database connection and log file
func (d *DebugDB) Close() error {
	if d.logFile != nil {
		d.logFile.Close()
	}
	return d.db.Close()
}

// Ping pings the database
func (d *DebugDB) Ping() error {
	return d.db.Ping()
}

// GetDB returns the underlying sql.DB for testing purposes
func (d *DebugDB) GetDB() *sql.DB {
	return d.db
} 