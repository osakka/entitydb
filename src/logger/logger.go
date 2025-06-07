// Package logger provides structured logging capabilities for EntityDB.
//
// The logger supports multiple log levels (TRACE, DEBUG, INFO, WARN, ERROR)
// and automatically includes contextual information such as file, function,
// and line numbers. It's designed for high-performance concurrent access
// with atomic operations for level checking.
//
// Log output format:
//   YYYY/MM/DD HH:MM:SS.ssssss [PID:GID] [LEVEL] Message (function.file:line)
//
// The logger is safe for concurrent use and provides minimal overhead when
// logging is disabled for a particular level.
package logger

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// LogLevel represents the severity of a log message.
// Higher values indicate more severe messages.
type LogLevel int32

const (
	TRACE LogLevel = iota
	DEBUG
	INFO
	WARN
	ERROR
)

var (
	currentLevel atomic.Int32
	levelNames = map[LogLevel]string{
		TRACE: "TRACE",
		DEBUG: "DEBUG",
		INFO:  "INFO",
		WARN:  "WARN",
		ERROR: "ERROR",
	}
	
	// Trace subsystems that are enabled
	traceSubsystems = make(map[string]bool)
	traceMutex      sync.RWMutex
	
	// Process and thread IDs
	processID = os.Getpid()
	
	// Logger instance
	logger *log.Logger
)

func init() {
	// Custom logger with no prefix (we'll format everything ourselves)
	logger = log.New(os.Stdout, "", 0)
	currentLevel.Store(int32(INFO))
}

// SetLogLevel sets the minimum log level
func SetLogLevel(level string) error {
	switch strings.ToUpper(level) {
	case "TRACE":
		currentLevel.Store(int32(TRACE))
		Info("log level changed to TRACE")
	case "DEBUG":
		currentLevel.Store(int32(DEBUG))
		Info("log level changed to DEBUG")
	case "INFO":
		currentLevel.Store(int32(INFO))
		Info("log level changed to INFO")
	case "WARN":
		currentLevel.Store(int32(WARN))
		Info("log level changed to WARN")
	case "ERROR":
		currentLevel.Store(int32(ERROR))
		Info("log level changed to ERROR")
	default:
		return fmt.Errorf("invalid log level: %s", level)
	}
	return nil
}

// GetLogLevel returns the current log level
func GetLogLevel() string {
	level := LogLevel(currentLevel.Load())
	return strings.TrimSpace(levelNames[level])
}

// EnableTrace enables trace logging for specific subsystems
func EnableTrace(subsystems ...string) {
	traceMutex.Lock()
	defer traceMutex.Unlock()
	for _, subsystem := range subsystems {
		traceSubsystems[subsystem] = true
	}
}

// DisableTrace disables trace logging for specific subsystems
func DisableTrace(subsystems ...string) {
	traceMutex.Lock()
	defer traceMutex.Unlock()
	for _, subsystem := range subsystems {
		delete(traceSubsystems, subsystem)
	}
}

// ClearTrace disables all trace subsystems
func ClearTrace() {
	traceMutex.Lock()
	defer traceMutex.Unlock()
	traceSubsystems = make(map[string]bool)
}

// GetTraceSubsystems returns currently enabled trace subsystems
func GetTraceSubsystems() []string {
	traceMutex.RLock()
	defer traceMutex.RUnlock()
	
	subsystems := make([]string, 0, len(traceSubsystems))
	for subsystem := range traceSubsystems {
		subsystems = append(subsystems, subsystem)
	}
	return subsystems
}

// isTraceEnabled checks if trace is enabled for a subsystem
func isTraceEnabled(subsystem string) bool {
	traceMutex.RLock()
	defer traceMutex.RUnlock()
	return traceSubsystems[subsystem]
}

// formatMessage formats a log message according to our standard
func formatMessage(level LogLevel, skip int, format string, args ...interface{}) string {
	// Get caller info
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		file = "unknown"
		line = 0
	}
	
	// Extract just the filename without extension
	if idx := strings.LastIndex(file, "/"); idx != -1 {
		file = file[idx+1:]
	}
	if idx := strings.LastIndex(file, ".go"); idx != -1 {
		file = file[:idx]
	}
	
	// Get function name
	funcName := "unknown"
	if fn := runtime.FuncForPC(pc); fn != nil {
		fullName := fn.Name()
		// Extract just the function name
		if idx := strings.LastIndex(fullName, "."); idx != -1 {
			funcName = fullName[idx+1:]
		}
	}
	
	// Format message
	msg := fmt.Sprintf(format, args...)
	
	// Get current goroutine ID (thread ID equivalent)
	threadID := getGoroutineID()
	
	// Format: timestamp [pid:tid] [LEVEL] function.filename:line: message
	timestamp := time.Now().Format("2006/01/02 15:04:05.000000")
	return fmt.Sprintf("%s [%d:%d] [%s] %s.%s:%d: %s",
		timestamp, processID, threadID, levelNames[level], funcName, file, line, msg)
}

// getGoroutineID extracts the current goroutine ID
func getGoroutineID() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(string(buf[:n]))[1]
	id := 0
	fmt.Sscanf(idField, "%d", &id)
	return id
}

// logMessage is the internal logging function
func logMessage(level LogLevel, skip int, format string, args ...interface{}) {
	// Quick check if we should log (atomic operation, very fast)
	if level < LogLevel(currentLevel.Load()) {
		return
	}
	
	// Format and output message
	msg := formatMessage(level, skip, format, args...)
	logger.Println(msg)
}

// TraceIf logs a trace message only if the subsystem is enabled
func TraceIf(subsystem string, format string, args ...interface{}) {
	// Double check: both trace level and subsystem must be enabled
	if LogLevel(currentLevel.Load()) > TRACE || !isTraceEnabled(subsystem) {
		return
	}
	logMessage(TRACE, 3, "[%s] %s", subsystem, fmt.Sprintf(format, args...))
}

// Trace logs a trace-level message
func Trace(format string, args ...interface{}) {
	logMessage(TRACE, 3, format, args...)
}

// Debug logs a debug message
func Debug(format string, args ...interface{}) {
	logMessage(DEBUG, 3, format, args...)
}

// Info logs an info message
func Info(format string, args ...interface{}) {
	logMessage(INFO, 3, format, args...)
}

// Warn logs a warning message
func Warn(format string, args ...interface{}) {
	logMessage(WARN, 3, format, args...)
}

// Error logs an error message
func Error(format string, args ...interface{}) {
	logMessage(ERROR, 3, format, args...)
}

// Fatal logs an error message and exits
func Fatal(format string, args ...interface{}) {
	msg := formatMessage(ERROR, 2, format, args...)
	logger.Println(msg)
	os.Exit(1)
}

// Panic logs an error message and panics
func Panic(format string, args ...interface{}) {
	msg := formatMessage(ERROR, 2, format, args...)
	logger.Println(msg)
	panic(fmt.Sprintf(format, args...))
}

// Configure sets up logging from environment variables
func Configure() {
	// Set log level from environment
	if level := os.Getenv("ENTITYDB_LOG_LEVEL"); level != "" {
		SetLogLevel(level)
	}
	
	// Set trace subsystems from environment
	if trace := os.Getenv("ENTITYDB_TRACE_SUBSYSTEMS"); trace != "" {
		subsystems := strings.Split(trace, ",")
		for i, s := range subsystems {
			subsystems[i] = strings.TrimSpace(s)
		}
		EnableTrace(subsystems...)
	}
}

// Aliases for backward compatibility
var (
	Tracef = Trace
	Debugf = Debug
	Infof  = Info
	Warnf  = Warn
	Errorf = Error
	Fatalf = Fatal
	Panicf = Panic
)