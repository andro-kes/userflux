package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

const (
	logFilePath = "./logs/userflux.log"
	logFilePerm = 0644
	logDirPerm  = 0755
)

// Logger wraps the standard library logger with file and stderr output
type Logger struct {
	*log.Logger
	logFile *os.File
}

// NewLogger creates a new logger that writes to both a file and stderr
func NewLogger() (*Logger, error) {
	// Create logs directory if it doesn't exist
	logDir := filepath.Dir(logFilePath)
	if err := os.MkdirAll(logDir, logDirPerm); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open log file (truncating existing contents), create if it doesn't exist
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, logFilePerm)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Create multi-writer for both file and stderr
	multiWriter := io.MultiWriter(logFile, os.Stderr)

	// Create logger with timestamp and file/line info
	logger := log.New(multiWriter, "", log.LstdFlags|log.Lshortfile)

	return &Logger{
		Logger:  logger,
		logFile: logFile,
	}, nil
}

// Close closes the log file
func (l *Logger) Close() error {
	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}

// Info logs an informational message
func (l *Logger) Info(v ...interface{}) {
	msg := fmt.Sprint(v...)
	l.Logger.Output(2, "INFO: "+msg)
}

// Infof logs a formatted informational message
func (l *Logger) Infof(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	l.Logger.Output(2, "INFO: "+msg)
}

// Error logs an error message
func (l *Logger) Error(v ...interface{}) {
	msg := fmt.Sprint(v...)
	l.Logger.Output(2, "ERROR: "+msg)
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	l.Logger.Output(2, "ERROR: "+msg)
}
