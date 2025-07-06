package main

import (
	"fmt"
	"os"
)

// DatabaseConfig holds MySQL connection configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

// GetDatabaseConfig returns the database configuration from environment variables or defaults
func GetDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Host:     getEnv("MYSQL_HOST", "localhost"),
		Port:     getEnv("MYSQL_PORT", "3306"),
		User:     getEnv("MYSQL_USER", "root"),
		Password: getEnv("MYSQL_PASSWORD", "password"),
		Database: getEnv("MYSQL_DATABASE", "stingray"),
	}
}

// GetDSN returns the MySQL connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		c.User, c.Password, c.Host, c.Port, c.Database)
}

// GetDSNWithoutDB returns the MySQL connection string without database name
func (c *DatabaseConfig) GetDSNWithoutDB() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=true&loc=Local",
		c.User, c.Password, c.Host, c.Port)
}

// getEnv gets an environment variable or returns the default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
} 