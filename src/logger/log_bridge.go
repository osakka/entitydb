package logger

import (
	"log"
	"strings"
)

// logWriter implements io.Writer to redirect standard library log output to our logger
type logWriter struct{}

func (lw *logWriter) Write(p []byte) (n int, err error) {
	// Convert bytes to string and trim whitespace
	message := strings.TrimSpace(string(p))
	
	// Skip empty messages
	if message == "" {
		return len(p), nil
	}
	
	// Check if it's a TLS or HTTP error
	if strings.Contains(message, "TLS") || strings.Contains(message, "tls") {
		Warn("HTTP Server: %s", message)
	} else if strings.Contains(message, "error") || strings.Contains(message, "Error") {
		Error("HTTP Server: %s", message)
	} else {
		Info("HTTP Server: %s", message)
	}
	
	return len(p), nil
}

// InitLogBridge redirects standard library log output to our logger
func InitLogBridge() {
	// Create a new log writer
	writer := &logWriter{}
	
	// Set the standard library logger to use our writer
	log.SetOutput(writer)
	
	// Remove the default timestamp since our logger adds it
	log.SetFlags(0)
	
	Debug("Standard library log output redirected to EntityDB logger")
}

// SetHTTPServerErrorLog returns a logger that can be used for http.Server.ErrorLog
func SetHTTPServerErrorLog() *log.Logger {
	writer := &logWriter{}
	return log.New(writer, "", 0)
}