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
	LoggingLevel  int
	// Test user credentials
	TestAdminUsername    string
	TestAdminPassword    string
	TestCustomerUsername string
	TestCustomerPassword string
	TestWrongPassword    string
	// Email configuration
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
	// DKIM configuration
	DKIMPrivateKeyFile string
	DKIMSelector       string
	DKIMDomain         string
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
		LoggingLevel:  getEnvInt("LOGGING_LEVEL", 1),
		// Test user credentials
		TestAdminUsername:    getEnv("TEST_ADMIN_USERNAME", "admin"),
		TestAdminPassword:    getEnv("TEST_ADMIN_PASSWORD", "admin"),
		TestCustomerUsername: getEnv("TEST_CUSTOMER_USERNAME", "customer"),
		TestCustomerPassword: getEnv("TEST_CUSTOMER_PASSWORD", "customer"),
		TestWrongPassword:    getEnv("TEST_WRONG_PASSWORD", "wrongpassword"),
		// Email configuration
		SMTPHost:     getEnv("SMTP_HOST", "localhost"),
		SMTPPort:     getEnv("SMTP_PORT", "25"),
		SMTPUsername: getEnv("SMTP_USERNAME", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		FromEmail:    getEnv("FROM_EMAIL", "noreply@yourdomain.com"),
		FromName:     getEnv("FROM_NAME", "Sting Ray CMS"),
		// DKIM configuration
		DKIMPrivateKeyFile: getEnv("DKIM_PRIVATE_KEY_FILE", ".DKIM_KEY.txt"),
		DKIMSelector:       getEnv("DKIM_SELECTOR", "default"),
		DKIMDomain:         getEnv("DKIM_DOMAIN", "yourdomain.com"),
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
	var currentKey string
	var currentValue strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		
		// Skip empty lines and comments
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}

		// Check if this line starts a new key=value pair
		if strings.Contains(line, "=") {
			// Save previous key-value pair if exists
			if currentKey != "" {
				value := strings.TrimSpace(currentValue.String())
				if os.Getenv(currentKey) == "" {
					os.Setenv(currentKey, value)
				}
			}

			// Start new key-value pair
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				currentKey = strings.TrimSpace(parts[0])
				currentValue.Reset()
				currentValue.WriteString(strings.TrimSpace(parts[1]))
			}
		} else if currentKey != "" {
			// Continue multi-line value (indented or continuation)
			currentValue.WriteString(line)
			currentValue.WriteString("\n")
		}
	}

	// Save the last key-value pair
	if currentKey != "" {
		value := strings.TrimSpace(currentValue.String())
		if os.Getenv(currentKey) == "" {
			os.Setenv(currentKey, value)
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

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.MySQLUser, c.MySQLPassword, c.MySQLHost, c.MySQLPort, c.MySQLDatabase)
} 