package logger

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	TRACE LogLevel = iota // Most detailed level for tracing data flow
	DEBUG
	INFO
	WARN
	ERROR
)

var (
	currentLevel LogLevel = INFO
	levelNames = map[LogLevel]string{
		TRACE: "TRACE",
		DEBUG: "DEBUG",
		INFO:  "INFO",
		WARN:  "WARN",
		ERROR: "ERROR",
	}
)

// Logger is the main logger instance
var Logger *log.Logger

func init() {
	Logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lmicroseconds)
}

// SetLogLevel sets the minimum log level
func SetLogLevel(level string) error {
	switch strings.ToLower(level) {
	case "trace":
		currentLevel = TRACE
	case "debug":
		currentLevel = DEBUG
	case "info":
		currentLevel = INFO
	case "warn":
		currentLevel = WARN
	case "error":
		currentLevel = ERROR
	default:
		return fmt.Errorf("invalid log level: %s", level)
	}
	return nil
}

// GetLogLevel returns the current log level
func GetLogLevel() string {
	return levelNames[currentLevel]
}


// logf is the internal logging function
func logf(level LogLevel, format string, args ...interface{}) {
	if level >= currentLevel {
		// Get caller info (skip 3: logf -> Debug/Info/etc -> actual caller)
		pc, file, line, ok := runtime.Caller(3)
		if !ok {
			file = "unknown"
			line = 0
		}
		
		// Extract just the filename (not the full path)
		if idx := strings.LastIndex(file, "/"); idx != -1 {
			file = file[idx+1:]
		}
		
		// Get function name
		funcName := "unknown"
		if fn := runtime.FuncForPC(pc); fn != nil {
			fullName := fn.Name()
			// Extract just the function name (remove package path)
			parts := strings.Split(fullName, "/")
			lastPart := parts[len(parts)-1]
			// Remove package name if present
			if idx := strings.LastIndex(lastPart, "."); idx != -1 {
				funcName = lastPart[idx+1:]
			} else {
				funcName = lastPart
			}
		}
		
		msg := fmt.Sprintf(format, args...)
		// Format: timestamp [LEVEL] [file] [function] [line] message
		Logger.Printf("[%s] [%s] [%s] [%d] %s", levelNames[level], file, funcName, line, msg)
	}
}

// Trace logs a trace-level message for detailed data flow tracing
func Trace(format string, args ...interface{}) {
	logf(TRACE, format, args...)
}

// Debug logs a debug message
func Debug(format string, args ...interface{}) {
	logf(DEBUG, format, args...)
}

// Info logs an info message
func Info(format string, args ...interface{}) {
	logf(INFO, format, args...)
}

// Warn logs a warning message
func Warn(format string, args ...interface{}) {
	logf(WARN, format, args...)
}

// Error logs an error message
func Error(format string, args ...interface{}) {
	logf(ERROR, format, args...)
}

// Tracef is an alias for Trace
func Tracef(format string, args ...interface{}) {
	Trace(format, args...)
}

// Debugf is an alias for Debug
func Debugf(format string, args ...interface{}) {
	Debug(format, args...)
}

// Infof is an alias for Info
func Infof(format string, args ...interface{}) {
	Info(format, args...)
}

// Warnf is an alias for Warn
func Warnf(format string, args ...interface{}) {
	Warn(format, args...)
}

// Errorf is an alias for Error
func Errorf(format string, args ...interface{}) {
	Error(format, args...)
}

// Fatal logs an error message and exits the program
func Fatal(format string, args ...interface{}) {
	// Get caller info (skip 2: Fatal -> actual caller)
	pc, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "unknown"
		line = 0
	}
	
	// Extract just the filename (not the full path)
	if idx := strings.LastIndex(file, "/"); idx != -1 {
		file = file[idx+1:]
	}
	
	// Get function name
	funcName := "unknown"
	if fn := runtime.FuncForPC(pc); fn != nil {
		fullName := fn.Name()
		// Extract just the function name (remove package path)
		parts := strings.Split(fullName, "/")
		lastPart := parts[len(parts)-1]
		// Remove package name if present
		if idx := strings.LastIndex(lastPart, "."); idx != -1 {
			funcName = lastPart[idx+1:]
		} else {
			funcName = lastPart
		}
	}
	
	msg := fmt.Sprintf(format, args...)
	Logger.Printf("[FATAL] [%s] [%s] [%d] %s", file, funcName, line, msg)
	os.Exit(1)
}

// Fatalf is an alias for Fatal
func Fatalf(format string, args ...interface{}) {
	Fatal(format, args...)
}

// Panic logs an error message and panics
func Panic(format string, args ...interface{}) {
	// Get caller info (skip 2: Panic -> actual caller)
	pc, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "unknown"
		line = 0
	}
	
	// Extract just the filename (not the full path)
	if idx := strings.LastIndex(file, "/"); idx != -1 {
		file = file[idx+1:]
	}
	
	// Get function name
	funcName := "unknown"
	if fn := runtime.FuncForPC(pc); fn != nil {
		fullName := fn.Name()
		// Extract just the function name (remove package path)
		parts := strings.Split(fullName, "/")
		lastPart := parts[len(parts)-1]
		// Remove package name if present
		if idx := strings.LastIndex(lastPart, "."); idx != -1 {
			funcName = lastPart[idx+1:]
		} else {
			funcName = lastPart
		}
	}
	
	msg := fmt.Sprintf(format, args...)
	Logger.Printf("[PANIC] [%s] [%s] [%d] %s", file, funcName, line, msg)
	panic(msg)
}

// Panicf is an alias for Panic
func Panicf(format string, args ...interface{}) {
	Panic(format, args...)
}