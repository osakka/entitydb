package api

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"time"
	
	"entitydb/logger"
)

// TraceMiddleware provides detailed HTTP request tracing
type TraceMiddleware struct {
	enabled bool
}

// NewTraceMiddleware creates a new trace middleware
func NewTraceMiddleware() *TraceMiddleware {
	return &TraceMiddleware{
		enabled: true,
	}
}

// Middleware returns the trace middleware handler
func (tm *TraceMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !tm.enabled || !logger.IsTracingEnabled() {
			next.ServeHTTP(w, r)
			return
		}
		
		// Extract socket information
		remoteAddr := r.RemoteAddr
		localAddr := "unknown"
		if r.TLS != nil {
			localAddr = fmt.Sprintf("tls://%s", r.Host)
		} else {
			localAddr = fmt.Sprintf("http://%s", r.Host)
		}
		
		socketInfo := fmt.Sprintf("local=%s remote=%s", localAddr, remoteAddr)
		
		// Start trace
		trace := logger.StartTrace(fmt.Sprintf("HTTP_%s_%s", r.Method, r.URL.Path), socketInfo)
		if trace == nil {
			next.ServeHTTP(w, r)
			return
		}
		
		// Create traced response writer
		tw := &tracedResponseWriter{
			ResponseWriter: w,
			trace:          trace,
			statusCode:     200,
			startTime:      time.Now(),
		}
		
		// Add trace ID to request context for downstream use
		ctx := r.Context()
		ctx = withTraceID(ctx, trace.TraceID)
		r = r.WithContext(ctx)
		
		// Log request details
		trace.StartSpan("request_headers",
			fmt.Sprintf("content-length=%d", r.ContentLength),
			fmt.Sprintf("user-agent=%s", r.UserAgent()),
			fmt.Sprintf("connection=%s", r.Header.Get("Connection")),
		)
		
		// Check for problematic headers
		if teHeader := r.Header.Get("TE"); teHeader != "" {
			trace.StartSpan("te_header_detected", fmt.Sprintf("value=%s", teHeader))
			trace.EndSpan("te_header_detected")
		}
		
		trace.EndSpan("request_headers")
		
		// Trace handler execution
		trace.StartSpan("handler_execution")
		logger.LogHTTPHandler(trace.TraceID, r.Method, r.URL.Path, "start")
		
		// Call next handler
		next.ServeHTTP(tw, r)
		
		// End handler execution
		logger.LogHTTPHandler(trace.TraceID, r.Method, r.URL.Path, "end")
		trace.EndSpan("handler_execution")
		
		// Log response details
		duration := time.Since(tw.startTime)
		trace.StartSpan("response",
			fmt.Sprintf("status=%d", tw.statusCode),
			fmt.Sprintf("bytes_written=%d", tw.bytesWritten),
			fmt.Sprintf("duration_ms=%.2f", duration.Seconds()*1000),
		)
		trace.EndSpan("response")
		
		// End trace
		trace.EndTrace()
		
		// Warn on slow requests
		if duration > 5*time.Second {
			logger.Warn("[SLOW_REQUEST] Method=%s Path=%s Duration=%v Status=%d TraceID=%s",
				r.Method, r.URL.Path, duration, tw.statusCode, trace.TraceID)
		}
	})
}

// tracedResponseWriter wraps http.ResponseWriter to capture response details
type tracedResponseWriter struct {
	http.ResponseWriter
	trace        *logger.TraceContext
	statusCode   int
	bytesWritten int
	startTime    time.Time
	wroteHeader  bool
}

func (tw *tracedResponseWriter) WriteHeader(status int) {
	if !tw.wroteHeader {
		tw.statusCode = status
		tw.wroteHeader = true
		tw.ResponseWriter.WriteHeader(status)
		
		// Log when headers are written
		if tw.trace != nil {
			elapsed := time.Since(tw.startTime)
			logger.Trace("[HTTP_WRITE_HEADER] Status=%d Elapsed=%v TraceID=%s",
				status, elapsed, tw.trace.TraceID)
		}
	}
}

func (tw *tracedResponseWriter) Write(data []byte) (int, error) {
	if !tw.wroteHeader {
		tw.WriteHeader(200)
	}
	
	n, err := tw.ResponseWriter.Write(data)
	tw.bytesWritten += n
	
	// Log write operations
	if tw.trace != nil && logger.IsTracingEnabled() {
		elapsed := time.Since(tw.startTime)
		logger.Trace("[HTTP_WRITE_BODY] Bytes=%d Total=%d Elapsed=%v TraceID=%s",
			n, tw.bytesWritten, elapsed, tw.trace.TraceID)
	}
	
	return n, err
}

// Hijack implements http.Hijacker
func (tw *tracedResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := tw.ResponseWriter.(http.Hijacker); ok {
		conn, buf, err := hijacker.Hijack()
		if err == nil && tw.trace != nil {
			logger.Trace("[HTTP_HIJACK] TraceID=%s", tw.trace.TraceID)
		}
		return conn, buf, err
	}
	return nil, nil, fmt.Errorf("ResponseWriter does not implement http.Hijacker")
}