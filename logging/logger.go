package logging

import (
	"fmt"
	"log"
	"os"
	"time"
)

type Logger struct {
	accessLog *log.Logger
	errorLog  *log.Logger
	level     int
}

const (
	LevelErrors = 1
	LevelLogin  = 2
	LevelVerbose = 3
)

func NewLogger(level int) *Logger {
	// Open access log file
	accessFile, err := os.OpenFile("logs/access-log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Failed to open access log file: %v", err)
		accessFile = os.Stdout
	}

	// Open error log file
	errorFile, err := os.OpenFile("logs/error-log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Failed to open error log file: %v", err)
		errorFile = os.Stderr
	}

	return &Logger{
		accessLog: log.New(accessFile, "", 0),
		errorLog:  log.New(errorFile, "", 0),
		level:     level,
	}
}

// LogAccess logs page requests in Apache-like format
func (l *Logger) LogAccess(method, path, remoteAddr, userAgent string, statusCode int, responseTime time.Duration) {
	if l.level >= LevelVerbose {
		timestamp := time.Now().Format("02/Jan/2006:15:04:05 -0700")
		l.accessLog.Printf(`%s - - [%s] "%s %s" %d %d "%s" "%s"`,
			remoteAddr, timestamp, method, path, statusCode, responseTime.Milliseconds(), "-", userAgent)
	}
}

// LogError logs error messages
func (l *Logger) LogError(format string, args ...interface{}) {
	if l.level >= LevelErrors {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		l.errorLog.Printf("[%s] ERROR: %s", timestamp, fmt.Sprintf(format, args...))
	}
}

// LogLogin logs user login attempts
func (l *Logger) LogLogin(username, remoteAddr string, success bool) {
	if l.level >= LevelLogin {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		status := "FAILED"
		if success {
			status = "SUCCESS"
		}
		l.accessLog.Printf("[%s] LOGIN: %s from %s - %s", timestamp, username, remoteAddr, status)
	}
}

// LogVerbose logs verbose information
func (l *Logger) LogVerbose(format string, args ...interface{}) {
	if l.level >= LevelVerbose {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		l.accessLog.Printf("[%s] VERBOSE: %s", timestamp, fmt.Sprintf(format, args...))
	}
} 