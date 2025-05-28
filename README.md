# Go Process Logger

A flexible and efficient structured logging library for Go applications with support for process-specific logging, multiple output formats, and daily log rotation.

## Features

- **Process-specific logging** - Organize logs by process/service
- **Multiple output formats** - JSON and human-readable text formats
- **Daily log rotation** - Automatic daily file rotation with date-based naming
- **Structured logging** - Support for structured data in log entries
- **Thread-safe** - Concurrent logging support with mutex protection
- **Configurable log levels** - DEBUG, INFO, WARN, ERROR, FATAL
- **Backward compatibility** - Support for both new process-specific and legacy logging methods

## Installation

```bash
go get github.com/yourusername/go-process-logger
```

## Quick Start

```go
package main

import (
    "log"
    
    "github.com/cryptogral/logger"
)

func main() {
    // Initialize the default logger
    _, err := logger.InitDefaultLogger("./logs", logger.INFO, logger.TextFormat)
    if err != nil {
        log.Fatal("Failed to initialize logger:", err)
    }
    defer logger.Close()

    // Log to a specific process
    logger.InfoProcess("api-server", "requests", "user_login", "User logged in successfully", map[string]interface{}{
        "user_id": 12345,
        "ip":      "192.168.1.100",
    })

    // Log to general category (backward compatibility)
    logger.Info("database", "connection", "Connected to database", nil)
}
```

## Usage Examples

### Basic Logging

```go
// Initialize logger with JSON format
logger, err := logger.NewLogger("./logs", logger.DEBUG, logger.JSONFormat)
if err != nil {
    panic(err)
}

// Log with different levels
logger.Debug("app", "debug", "Debugging information", nil)
logger.Info("app", "startup", "Application started", map[string]interface{}{
    "version": "1.0.0",
    "port":    8080,
})
logger.Warn("app", "warning", "Low memory warning", map[string]interface{}{
    "available": "100MB",
})
logger.Error("app", "error", "Database connection failed", map[string]interface{}{
    "error": "connection timeout",
    "host":  "localhost:5432",
})
```

### Process-Specific Logging

```go
// Different processes can have separate log files
logger.InfoProcess("api-server", "requests", "http_request", "GET /users", map[string]interface{}{
    "method":      "GET",
    "path":        "/users",
    "status_code": 200,
    "duration_ms": 45,
})

logger.InfoProcess("worker", "jobs", "job_completed", "Email sent successfully", map[string]interface{}{
    "job_id":    "job-123",
    "recipient": "user@example.com",
})

logger.InfoProcess("scheduler", "", "task_scheduled", "Backup task scheduled", map[string]interface{}{
    "task": "daily_backup",
    "time": "02:00",
})
```

### Switching Log Formats

```go
// Start with JSON format
logger.SetFormat(logger.JSONFormat)
logger.Info("app", "test", "This will be in JSON format", nil)

// Switch to text format
logger.SetFormat(logger.TextFormat)
logger.Info("app", "test", "This will be in text format", nil)
```

## Log File Organization

The logger organizes log files in the following structure:

```
logs/
├── api-server/
│   ├── requests_2024-01-15.log
│   ├── requests_2024-01-16.log
│   └── errors_2024-01-15.log
├── worker/
│   ├── jobs_2024-01-15.log
│   └── jobs_2024-01-16.log
└── general/
    ├── database_2024-01-15.log
    └── app_2024-01-15.log
```

## Log Formats

### JSON Format
```json
{"timestamp":"2024-01-15T10:30:45Z","level":"INFO","process":"api-server","action":"user_login","message":"User logged in successfully","details":{"user_id":12345,"ip":"192.168.1.100"}}
```

### Text Format
```
[2024-01-15T10:30:45Z] INFO | api-server | user_login | User logged in successfully | user_id=12345 ip=192.168.1.100
```

## Configuration

### Log Levels

- `DEBUG` - Detailed information for debugging
- `INFO` - General information about application flow
- `WARN` - Warning messages for potentially harmful situations
- `ERROR` - Error messages for error conditions
- `FATAL` - Critical errors that cause application termination

### Usage with Different Log Levels

```go
// Only log INFO and above
logger, _ := logger.NewLogger("./logs", logger.INFO, logger.TextFormat)

// This will be logged
logger.Info("app", "info", "This will appear", nil)

// This will be ignored
logger.Debug("app", "debug", "This will be ignored", nil)
```

## API Reference

### Core Functions

- `NewLogger(baseLogDir, minLevel, format)` - Create a new logger instance
- `InitDefaultLogger(baseLogDir, minLevel, format)` - Initialize default logger
- `GetDefaultLogger()` - Get the default logger instance

### Logging Methods

**Process-specific logging:**
- `DebugProcess(processDir, category, action, message, details)`
- `InfoProcess(processDir, category, action, message, details)`
- `WarnProcess(processDir, category, action, message, details)`
- `ErrorProcess(processDir, category, action, message, details)`
- `FatalProcess(processDir, category, action, message, details)`

**General logging (backward compatibility):**
- `Debug(source, action, message, details)`
- `Info(source, action, message, details)`
- `Warn(source, action, message, details)`
- `Error(source, action, message, details)`
- `Fatal(source, action, message, details)`

## Thread Safety

All logging operations are thread-safe and can be called concurrently from multiple goroutines.

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Changelog

### v1.0.0
- Initial release
- Process-specific logging
- Multiple output formats (JSON/Text)
- Daily log rotation
- Thread-safe operations
- Backward compatibility support
