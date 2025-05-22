package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// LogLevel defines the logging level
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	return [...]string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}[l]
}

// LogFormat defines the format for log output
type LogFormat int

const (
	JSONFormat LogFormat = iota // JSON format for structured logging
	TextFormat                  // Human-readable text format
)

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp string      `json:"timestamp"`
	Level     string      `json:"level"`
	Process   string      `json:"process"` // Process or system component
	Action    string      `json:"action"`
	Message   string      `json:"message"`
	Details   interface{} `json:"details,omitempty"` // Can contain any structured data
}

// Logger is the main logging object
type Logger struct {
	baseLogDir string // Base directory for all logs
	logFiles   map[string]*os.File
	mutex      sync.Mutex
	minLevel   LogLevel
	format     LogFormat
}

var (
	defaultLogger *Logger
	once          sync.Once
)

// InitDefaultLogger initializes the default logger instance
func InitDefaultLogger(baseLogDir string, minLevel LogLevel, format LogFormat) (*Logger, error) {
	var err error
	once.Do(func() {
		defaultLogger, err = NewLogger(baseLogDir, minLevel, format)
	})
	return defaultLogger, err
}

// GetDefaultLogger returns the default logger instance
func GetDefaultLogger() *Logger {
	if defaultLogger == nil {
		panic("DefaultLogger not initialized. Call InitDefaultLogger first.")
	}
	return defaultLogger
}

// NewLogger creates a new logger instance
func NewLogger(baseLogDir string, minLevel LogLevel, format LogFormat) (*Logger, error) {
	// Create base log directory
	if err := os.MkdirAll(baseLogDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base log directory: %w", err)
	}

	return &Logger{
		baseLogDir: baseLogDir,
		logFiles:   make(map[string]*os.File),
		minLevel:   minLevel,
		format:     format,
	}, nil
}

// getLogFile returns or creates a log file for the specified process and category
func (l *Logger) getLogFile(processDir, category string) (*os.File, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// Get current date in YYYY-MM-DD format
	currentDate := time.Now().Format("2006-01-02")

	// Create unique key for caching open file
	// Key includes date so a new file is created each day
	fileKey := filepath.Join(processDir, category, currentDate)

	// If file is already open, return it
	if file, exists := l.logFiles[fileKey]; exists {
		return file, nil
	}

	// Form full path to process log directory
	processLogDir := filepath.Join(l.baseLogDir, processDir)

	// Create directory for process logs
	if err := os.MkdirAll(processLogDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create process log directory: %w", err)
	}

	// Determine log file name with date
	var filename string
	if category == "" {
		// If no category specified, use process name with date
		filename = filepath.Join(processLogDir, fmt.Sprintf("%s.log", currentDate))
	} else {
		// Otherwise use category with date
		filename = filepath.Join(processLogDir, fmt.Sprintf("%s_%s.log", category, currentDate))
	}

	// Open log file
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file %s: %w", filename, err)
	}

	// Cache the opened file
	l.logFiles[fileKey] = file
	return file, nil
}

// formatDetails formats details into human-readable text
func formatDetails(details interface{}) string {
	if details == nil {
		return ""
	}

	switch v := details.(type) {
	case string:
		return v
	case map[string]interface{}:
		var parts []string
		for key, val := range v {
			parts = append(parts, fmt.Sprintf("%s=%v", key, val))
		}
		return strings.Join(parts, " ")
	default:
		// For other types, try using JSON
		jsonData, err := json.Marshal(details)
		if err != nil {
			return fmt.Sprintf("%v", details)
		}
		// Remove curly braces for cleaner output
		jsonStr := string(jsonData)
		jsonStr = strings.TrimPrefix(jsonStr, "{")
		jsonStr = strings.TrimSuffix(jsonStr, "}")
		// Replace JSON separators with spaces
		jsonStr = strings.ReplaceAll(jsonStr, "\":", "=")
		jsonStr = strings.ReplaceAll(jsonStr, "\",", " ")
		jsonStr = strings.ReplaceAll(jsonStr, "\"", "")
		return jsonStr
	}
}

// LogToProcess writes a message to the log for the specified process and category
func (l *Logger) LogToProcess(level LogLevel, processDir, category, action, message string, details interface{}) error {
	// Check logging level
	if level < l.minLevel {
		return nil
	}

	// Get file for writing
	file, err := l.getLogFile(processDir, category)
	if err != nil {
		return err
	}

	timestamp := time.Now().Format(time.RFC3339)
	var logLine []byte

	if l.format == JSONFormat {
		// Standard JSON format
		entry := LogEntry{
			Timestamp: timestamp,
			Level:     level.String(),
			Process:   processDir,
			Action:    action,
			Message:   message,
			Details:   details,
		}

		// Serialize to JSON
		jsonData, err := json.Marshal(entry)
		if err != nil {
			return fmt.Errorf("failed to marshal log entry: %w", err)
		}
		logLine = jsonData
	} else {
		// Human-readable text format
		detailsStr := ""
		if details != nil {
			detailsStr = " | " + formatDetails(details)
		}

		// Format: [TIMESTAMP] LEVEL | PROCESS | ACTION | MESSAGE | details
		logLine = []byte(fmt.Sprintf("[%s] %s | %s | %s | %s%s",
			timestamp,
			level.String(),
			processDir,
			action,
			message,
			detailsStr))
	}

	// Write to file
	logLine = append(logLine, '\n')
	if _, err := file.Write(logLine); err != nil {
		return fmt.Errorf("failed to write to log file: %w", err)
	}

	// For FATAL level, also output to stderr
	if level == FATAL {
		fmt.Fprintf(os.Stderr, "%s\n", string(logLine))
	}

	return nil
}

// SetFormat changes the logging format
func (l *Logger) SetFormat(format LogFormat) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.format = format
}

// Convenience wrappers for process-specific logging
func (l *Logger) DebugProcess(processDir, category, action, message string, details interface{}) error {
	return l.LogToProcess(DEBUG, processDir, category, action, message, details)
}

func (l *Logger) InfoProcess(processDir, category, action, message string, details interface{}) error {
	return l.LogToProcess(INFO, processDir, category, action, message, details)
}

func (l *Logger) WarnProcess(processDir, category, action, message string, details interface{}) error {
	return l.LogToProcess(WARN, processDir, category, action, message, details)
}

func (l *Logger) ErrorProcess(processDir, category, action, message string, details interface{}) error {
	return l.LogToProcess(ERROR, processDir, category, action, message, details)
}

func (l *Logger) FatalProcess(processDir, category, action, message string, details interface{}) {
	l.LogToProcess(FATAL, processDir, category, action, message, details)
	os.Exit(1)
}

// For backward compatibility, old methods log to general directory
func (l *Logger) Debug(source, action, message string, details interface{}) error {
	return l.LogToProcess(DEBUG, "general", source, action, message, details)
}

func (l *Logger) Info(source, action, message string, details interface{}) error {
	return l.LogToProcess(INFO, "general", source, action, message, details)
}

func (l *Logger) Warn(source, action, message string, details interface{}) error {
	return l.LogToProcess(WARN, "general", source, action, message, details)
}

func (l *Logger) Error(source, action, message string, details interface{}) error {
	return l.LogToProcess(ERROR, "general", source, action, message, details)
}

func (l *Logger) Fatal(source, action, message string, details interface{}) {
	l.LogToProcess(FATAL, "general", source, action, message, details)
	os.Exit(1)
}

// Close closes all open log files
func (l *Logger) Close() {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	for _, file := range l.logFiles {
		file.Close()
	}
}

// SetFormat changes the format for the default logger
func SetFormat(format LogFormat) {
	GetDefaultLogger().SetFormat(format)
}

// Helper functions for working with the default logger

// DebugProcess logs a debug message for a specific process
func DebugProcess(processDir, category, action, message string, details interface{}) error {
	return GetDefaultLogger().DebugProcess(processDir, category, action, message, details)
}

// InfoProcess logs an info message for a specific process
func InfoProcess(processDir, category, action, message string, details interface{}) error {
	return GetDefaultLogger().InfoProcess(processDir, category, action, message, details)
}

// WarnProcess logs a warning message for a specific process
func WarnProcess(processDir, category, action, message string, details interface{}) error {
	return GetDefaultLogger().WarnProcess(processDir, category, action, message, details)
}

// ErrorProcess logs an error message for a specific process
func ErrorProcess(processDir, category, action, message string, details interface{}) error {
	return GetDefaultLogger().ErrorProcess(processDir, category, action, message, details)
}

// FatalProcess logs a fatal message for a specific process and exits
func FatalProcess(processDir, category, action, message string, details interface{}) {
	GetDefaultLogger().FatalProcess(processDir, category, action, message, details)
}

// For backward compatibility

// Debug logs a debug message
func Debug(source, action, message string, details interface{}) error {
	return GetDefaultLogger().Debug(source, action, message, details)
}

// Info logs an info message
func Info(source, action, message string, details interface{}) error {
	return GetDefaultLogger().Info(source, action, message, details)
}

// Warn logs a warning message
func Warn(source, action, message string, details interface{}) error {
	return GetDefaultLogger().Warn(source, action, message, details)
}

// Error logs an error message
func Error(source, action, message string, details interface{}) error {
	return GetDefaultLogger().Error(source, action, message, details)
}

// Fatal logs a fatal message and exits
func Fatal(source, action, message string, details interface{}) {
	GetDefaultLogger().Fatal(source, action, message, details)
}

// Close closes the default logger
func Close() {
	if defaultLogger != nil {
		defaultLogger.Close()
	}
}
