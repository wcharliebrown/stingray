package database

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

type Database struct {
	db       *sql.DB
	debugDB  *DebugDB
}

func NewDatabase(dsn string, debuggingMode bool) (*Database, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Create debug wrapper
	debugDB := NewDebugDB(db, debuggingMode)

	database := &Database{
		db:      db,
		debugDB: debugDB,
	}
	
	if err := database.initDatabase(); err != nil {
		return nil, err
	}

	return database, nil
}

func (d *Database) Close() error {
	if d.debugDB != nil {
		return d.debugDB.Close()
	}
	return d.db.Close()
} 

// GetDB returns the underlying sql.DB for testing purposes
func (d *Database) GetDB() *sql.DB {
	return d.db
}

// GetDebugDB returns the debug wrapper for database operations
func (d *Database) GetDebugDB() *DebugDB {
	return d.debugDB
}

// Helper methods that use debug wrapper when available

// Exec executes a query using debug wrapper if available
func (d *Database) Exec(query string, args ...interface{}) (sql.Result, error) {
	if d.debugDB != nil {
		return d.debugDB.Exec(query, args...)
	}
	return d.db.Exec(query, args...)
}

// Query executes a query using debug wrapper if available
func (d *Database) Query(query string, args ...interface{}) (*sql.Rows, error) {
	if d.debugDB != nil {
		return d.debugDB.Query(query, args...)
	}
	return d.db.Query(query, args...)
}

// QueryRow executes a query using debug wrapper if available
func (d *Database) QueryRow(query string, args ...interface{}) *sql.Row {
	if d.debugDB != nil {
		return d.debugDB.QueryRow(query, args...)
	}
	return d.db.QueryRow(query, args...)
}

// Prepare prepares a statement using debug wrapper if available
func (d *Database) Prepare(query string) (*sql.Stmt, error) {
	if d.debugDB != nil {
		return d.debugDB.Prepare(query)
	}
	return d.db.Prepare(query)
}

// Begin begins a transaction using debug wrapper if available
func (d *Database) Begin() (*sql.Tx, error) {
	if d.debugDB != nil {
		return d.debugDB.Begin()
	}
	return d.db.Begin()
}
