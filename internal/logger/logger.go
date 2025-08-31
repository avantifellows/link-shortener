package logger

import (
	"log"
	"os"
	"strings"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

var currentLogLevel LogLevel

func init() {
	// Set log level from environment variable
	logLevel := strings.ToUpper(os.Getenv("LOG_LEVEL"))
	switch logLevel {
	case "DEBUG":
		currentLogLevel = DEBUG
	case "INFO":
		currentLogLevel = INFO
	case "WARN":
		currentLogLevel = WARN
	case "ERROR":
		currentLogLevel = ERROR
	default:
		// Default to INFO for production
		currentLogLevel = INFO
	}
}

func Debug(msg string, args ...interface{}) {
	if currentLogLevel <= DEBUG {
		log.Printf("[DEBUG] "+msg, args...)
	}
}

func Info(msg string, args ...interface{}) {
	if currentLogLevel <= INFO {
		log.Printf("[INFO] "+msg, args...)
	}
}

func Warn(msg string, args ...interface{}) {
	if currentLogLevel <= WARN {
		log.Printf("[WARN] "+msg, args...)
	}
}

func Error(msg string, args ...interface{}) {
	if currentLogLevel <= ERROR {
		log.Printf("[ERROR] "+msg, args...)
	}
}