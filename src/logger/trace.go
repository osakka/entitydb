package logger

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// TraceContext represents a traced operation context
type TraceContext struct {
	TraceID      string
	Operation    string
	StartTime    time.Time
	GoroutineID  int
	ThreadID     int
	SocketInfo   string
	mu           sync.Mutex
	spans        []TraceSpan
	isActive     bool
}

// TraceSpan represents a span within a trace
type TraceSpan struct {
	Name         string
	StartTime    time.Time
	EndTime      time.Time
	GoroutineID  int
	ThreadID     int
	Attributes   map[string]string
}

var (
	// Global trace storage
	activeTraces   = make(map[string]*TraceContext)
	activeTracesMu sync.RWMutex
	
	// Trace ID counter
	traceCounter uint64
	
	// Enable/disable tracing
	tracingEnabled atomic.Bool
)

// EnableTracing enables HTTP and thread tracing
func EnableTracing(enabled bool) {
	tracingEnabled.Store(enabled)
	if enabled {
		Info("HTTP and thread tracing enabled")
	} else {
		Info("HTTP and thread tracing disabled")
	}
}

// IsTracingEnabled returns whether tracing is enabled
func IsTracingEnabled() bool {
	return tracingEnabled.Load()
}

// StartTrace begins a new trace context
func StartTrace(operation string, socketInfo string) *TraceContext {
	if !IsTracingEnabled() {
		return nil
	}
	
	traceID := fmt.Sprintf("trace_%d_%d", time.Now().UnixNano(), atomic.AddUint64(&traceCounter, 1))
	
	ctx := &TraceContext{
		TraceID:     traceID,
		Operation:   operation,
		StartTime:   time.Now(),
		GoroutineID: getGoroutineID(),
		ThreadID:    getThreadID(),
		SocketInfo:  socketInfo,
		spans:       make([]TraceSpan, 0),
		isActive:    true,
	}
	
	activeTracesMu.Lock()
	activeTraces[traceID] = ctx
	activeTracesMu.Unlock()
	
	Trace("[TRACE_START] ID=%s Op=%s Socket=%s Goroutine=%d Thread=%d",
		traceID, operation, socketInfo, ctx.GoroutineID, ctx.ThreadID)
	
	return ctx
}

// StartSpan begins a new span within a trace
func (tc *TraceContext) StartSpan(name string, attributes ...string) {
	if tc == nil || !tc.isActive {
		return
	}
	
	tc.mu.Lock()
	defer tc.mu.Unlock()
	
	span := TraceSpan{
		Name:        name,
		StartTime:   time.Now(),
		GoroutineID: getGoroutineID(),
		ThreadID:    getThreadID(),
		Attributes:  make(map[string]string),
	}
	
	// Parse attributes (key=value pairs)
	for _, attr := range attributes {
		parts := strings.SplitN(attr, "=", 2)
		if len(parts) == 2 {
			span.Attributes[parts[0]] = parts[1]
		}
	}
	
	tc.spans = append(tc.spans, span)
	
	elapsed := time.Since(tc.StartTime)
	Trace("[SPAN_START] Trace=%s Span=%s Elapsed=%v Goroutine=%d Thread=%d Attrs=%v",
		tc.TraceID, name, elapsed, span.GoroutineID, span.ThreadID, span.Attributes)
}

// EndSpan completes the most recent span
func (tc *TraceContext) EndSpan(name string) {
	if tc == nil || !tc.isActive {
		return
	}
	
	tc.mu.Lock()
	defer tc.mu.Unlock()
	
	// Find the most recent matching span
	for i := len(tc.spans) - 1; i >= 0; i-- {
		if tc.spans[i].Name == name && tc.spans[i].EndTime.IsZero() {
			tc.spans[i].EndTime = time.Now()
			duration := tc.spans[i].EndTime.Sub(tc.spans[i].StartTime)
			elapsed := time.Since(tc.StartTime)
			
			Trace("[SPAN_END] Trace=%s Span=%s Duration=%v Elapsed=%v Goroutine=%d",
				tc.TraceID, name, duration, elapsed, getGoroutineID())
			break
		}
	}
}

// EndTrace completes the trace
func (tc *TraceContext) EndTrace() {
	if tc == nil || !tc.isActive {
		return
	}
	
	tc.mu.Lock()
	tc.isActive = false
	duration := time.Since(tc.StartTime)
	tc.mu.Unlock()
	
	activeTracesMu.Lock()
	delete(activeTraces, tc.TraceID)
	activeTracesMu.Unlock()
	
	// Log trace summary
	tc.mu.Lock()
	defer tc.mu.Unlock()
	
	Trace("[TRACE_END] ID=%s Op=%s Duration=%v Spans=%d",
		tc.TraceID, tc.Operation, duration, len(tc.spans))
	
	// Log any unclosed spans (potential hang points)
	for _, span := range tc.spans {
		if span.EndTime.IsZero() {
			Warn("[UNCLOSED_SPAN] Trace=%s Span=%s Started=%v Goroutine=%d",
				tc.TraceID, span.Name, span.StartTime, span.GoroutineID)
		}
	}
}

// LogLockOperation logs mutex/lock operations for deadlock detection
func LogLockOperation(traceID, lockType, lockName, operation string) {
	if !IsTracingEnabled() {
		return
	}
	
	goroutineID := getGoroutineID()
	threadID := getThreadID()
	
	// Get stack trace for lock operations
	buf := make([]byte, 1024)
	n := runtime.Stack(buf, false)
	stack := string(buf[:n])
	
	// Extract calling function
	frames := strings.Split(stack, "\n")
	caller := "unknown"
	if len(frames) > 5 {
		// Skip runtime frames to get actual caller
		for i := 4; i < len(frames); i += 2 {
			if !strings.Contains(frames[i], "logger.LogLockOperation") &&
			   !strings.Contains(frames[i], "runtime.") {
				caller = strings.TrimSpace(frames[i])
				break
			}
		}
	}
	
	Trace("[LOCK_%s] Type=%s Name=%s Goroutine=%d Thread=%d Caller=%s TraceID=%s",
		strings.ToUpper(operation), lockType, lockName, goroutineID, threadID, caller, traceID)
}

// LogHTTPAccept logs when a connection is accepted
func LogHTTPAccept(localAddr, remoteAddr string) {
	if !IsTracingEnabled() {
		return
	}
	
	Trace("[HTTP_ACCEPT] Local=%s Remote=%s Goroutine=%d Thread=%d",
		localAddr, remoteAddr, getGoroutineID(), getThreadID())
}

// LogHTTPHandler logs when a handler starts/ends
func LogHTTPHandler(traceID, method, path, phase string) {
	if !IsTracingEnabled() {
		return
	}
	
	Trace("[HTTP_HANDLER_%s] Method=%s Path=%s Goroutine=%d TraceID=%s",
		strings.ToUpper(phase), method, path, getGoroutineID(), traceID)
}

// GetActiveTraces returns currently active traces (for debugging)
func GetActiveTraces() []string {
	activeTracesMu.RLock()
	defer activeTracesMu.RUnlock()
	
	traces := make([]string, 0, len(activeTraces))
	for traceID, ctx := range activeTraces {
		duration := time.Since(ctx.StartTime)
		traces = append(traces, fmt.Sprintf("%s: %s (duration: %v)", traceID, ctx.Operation, duration))
	}
	return traces
}

// getThreadID gets the OS thread ID (Linux-specific, returns -1 on other platforms)
func getThreadID() int {
	// This is a simplified version - real thread ID would require syscall
	// For now, we'll use a hash of the goroutine ID as a proxy
	return getGoroutineID() % 1000
}