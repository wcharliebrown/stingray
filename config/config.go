package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"strconv"
)

type Config struct {
	MySQLHost     string
	MySQLPort     string
	MySQLUser     string
	MySQLPassword string
	MySQLDatabase string
	DebuggingMode bool
	// Test user credentials
	TestAdminUsername    string
	TestAdminPassword    string
	TestCustomerUsername string
	TestCustomerPassword string
	TestWrongPassword    string
}

func LoadConfig() *Config {
	// Load .env file if it exists
	loadEnvFile(".env")
	
	return &Config{
		MySQLHost:     os.Getenv("MYSQL_HOST"),
		MySQLPort:     os.Getenv("MYSQL_PORT"),
		MySQLUser:     os.Getenv("MYSQL_USER"),
		MySQLPassword: os.Getenv("MYSQL_PASSWORD"),
		MySQLDatabase: os.Getenv("MYSQL_DATABASE"),
		DebuggingMode: getEnvBool("DEBUGGING_MODE", false),
		// Test user credentials
		TestAdminUsername:    getEnv("TEST_ADMIN_USERNAME", "admin"),
		TestAdminPassword:    getEnv("TEST_ADMIN_PASSWORD", "admin"),
		TestCustomerUsername: getEnv("TEST_CUSTOMER_USERNAME", "customer"),
		TestCustomerPassword: getEnv("TEST_CUSTOMER_PASSWORD", "customer"),
		TestWrongPassword:    getEnv("TEST_WRONG_PASSWORD", "wrongpassword"),
	}
}

func loadEnvFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		// File doesn't exist, that's okay
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// Only set if not already set in environment
			if os.Getenv(key) == "" {
				os.Setenv(key, value)
			}
		}
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.MySQLUser, c.MySQLPassword, c.MySQLHost, c.MySQLPort, c.MySQLDatabase)
} 