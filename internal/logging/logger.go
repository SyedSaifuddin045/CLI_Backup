package logging

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

var (
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
	DebugLogger *log.Logger
)

// Init initializes the logger with custom or default values.
func Init(logToFile bool, debug bool, logFilePath string) {
	// Ensure that the directory for the log file exists
	dir := filepath.Dir(logFilePath) // Get the directory part of the log file path
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0755) // Create directory if it doesn't exist
		if err != nil {
			log.Fatalf("Failed to create log directory: %v", err)
		}
	}

	// Set output destination
	var output io.Writer = os.Stdout

	if logToFile {
		f, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatalf("Failed to open log file: %v", err)
		}
		output = io.MultiWriter(os.Stdout, f)
	}

	InfoLogger = log.New(output, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(output, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	if debug {
		DebugLogger = log.New(output, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
	} else {
		DebugLogger = log.New(io.Discard, "", 0) // disables debug logs
	}
}
