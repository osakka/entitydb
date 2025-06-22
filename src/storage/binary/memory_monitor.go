// Package binary provides comprehensive memory monitoring and automatic pressure relief.
package binary

import (
	"entitydb/logger"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// MemoryMonitor provides comprehensive memory monitoring and automatic pressure relief
type MemoryMonitor struct {
	// Configuration
	highPressureThreshold  float64 // 0.8 = 80% memory usage
	criticalThreshold      float64 // 0.9 = 90% memory usage
	monitorInterval        time.Duration
	
	// State
	running                int64 // Atomic flag
	currentPressure        int64 // Atomic float64 bits
	lastGCTime             int64 // Atomic timestamp
	
	// Callbacks for pressure relief
	pressureCallbacks      []PressureReliefCallback
	callbackMutex          sync.RWMutex
	
	// Statistics
	gcTriggerCount         int64
	pressureReliefCount    int64
	maxPressureObserved    int64 // Atomic float64 bits
}

// PressureReliefCallback is called when memory pressure is detected
type PressureReliefCallback func(pressure float64, level PressureLevel)

// PressureLevel indicates the severity of memory pressure
type PressureLevel int

const (
	PressureLow PressureLevel = iota
	PressureMedium
	PressureHigh
	PressureCritical
)

func (p PressureLevel) String() string {
	switch p {
	case PressureLow:
		return "low"
	case PressureMedium:
		return "medium"
	case PressureHigh:
		return "high"
	case PressureCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// NewMemoryMonitor creates a new memory monitor with automatic pressure relief
func NewMemoryMonitor() *MemoryMonitor {
	return &MemoryMonitor{
		highPressureThreshold: 0.8,
		criticalThreshold:     0.9,
		monitorInterval:       30 * time.Second, // Check every 30 seconds
		pressureCallbacks:     make([]PressureReliefCallback, 0),
	}
}

// Start begins memory monitoring in the background
func (mm *MemoryMonitor) Start() {
	if atomic.CompareAndSwapInt64(&mm.running, 0, 1) {
		logger.Info("Memory monitor started with thresholds: high=%.1f%%, critical=%.1f%%", 
			mm.highPressureThreshold*100, mm.criticalThreshold*100)
		go mm.monitorLoop()
	}
}

// Stop stops the memory monitor
func (mm *MemoryMonitor) Stop() {
	atomic.StoreInt64(&mm.running, 0)
	logger.Info("Memory monitor stopped")
}

// AddPressureCallback registers a callback for memory pressure events
func (mm *MemoryMonitor) AddPressureCallback(callback PressureReliefCallback) {
	mm.callbackMutex.Lock()
	defer mm.callbackMutex.Unlock()
	mm.pressureCallbacks = append(mm.pressureCallbacks, callback)
}

// GetCurrentPressure returns the current memory pressure (0.0 to 1.0)
func (mm *MemoryMonitor) GetCurrentPressure() float64 {
	bits := atomic.LoadInt64(&mm.currentPressure)
	return float64FromBits(bits)
}

// GetStats returns memory monitoring statistics
type MemoryStats struct {
	CurrentPressure     float64
	MaxPressureObserved float64
	GCTriggerCount      int64
	PressureReliefCount int64
	LastGCTime          time.Time
	IsRunning           bool
}

func (mm *MemoryMonitor) GetStats() MemoryStats {
	return MemoryStats{
		CurrentPressure:     mm.GetCurrentPressure(),
		MaxPressureObserved: float64FromBits(atomic.LoadInt64(&mm.maxPressureObserved)),
		GCTriggerCount:      atomic.LoadInt64(&mm.gcTriggerCount),
		PressureReliefCount: atomic.LoadInt64(&mm.pressureReliefCount),
		LastGCTime:          time.Unix(0, atomic.LoadInt64(&mm.lastGCTime)),
		IsRunning:           atomic.LoadInt64(&mm.running) == 1,
	}
}

// monitorLoop runs the memory monitoring loop
func (mm *MemoryMonitor) monitorLoop() {
	ticker := time.NewTicker(mm.monitorInterval)
	defer ticker.Stop()
	
	for atomic.LoadInt64(&mm.running) == 1 {
		select {
		case <-ticker.C:
			mm.checkMemoryPressure()
		}
	}
}

// checkMemoryPressure checks current memory pressure and takes action if needed
func (mm *MemoryMonitor) checkMemoryPressure() {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	
	// Calculate memory pressure
	pressure := mm.calculatePressure(&mem)
	atomic.StoreInt64(&mm.currentPressure, float64ToBits(pressure))
	
	// Update max pressure observed
	for {
		currentMax := atomic.LoadInt64(&mm.maxPressureObserved)
		if pressure <= float64FromBits(currentMax) {
			break
		}
		if atomic.CompareAndSwapInt64(&mm.maxPressureObserved, currentMax, float64ToBits(pressure)) {
			break
		}
	}
	
	// Determine pressure level
	level := mm.getPressureLevel(pressure)
	
	// Log pressure if significant
	if pressure > 0.5 {
		logger.Debug("Memory pressure: %.1f%% (%s) - Heap: %d MB, Sys: %d MB", 
			pressure*100, level.String(), 
			mem.HeapInuse/(1024*1024), mem.Sys/(1024*1024))
	}
	
	// Take action based on pressure level
	switch level {
	case PressureHigh, PressureCritical:
		mm.handleHighPressure(pressure, level, &mem)
	case PressureMedium:
		mm.handleMediumPressure(pressure, level, &mem)
	}
	
	// Notify callbacks
	if level >= PressureMedium {
		mm.notifyPressureCallbacks(pressure, level)
	}
}

// calculatePressure calculates memory pressure from MemStats
func (mm *MemoryMonitor) calculatePressure(mem *runtime.MemStats) float64 {
	// Use multiple metrics to calculate pressure
	heapPressure := float64(mem.HeapInuse) / float64(mem.Sys)
	gcPressure := float64(mem.NumGC) / 1000.0 // Normalize GC count
	if gcPressure > 1.0 {
		gcPressure = 1.0
	}
	
	// Weighted combination
	pressure := (heapPressure * 0.8) + (gcPressure * 0.2)
	if pressure > 1.0 {
		pressure = 1.0
	}
	
	return pressure
}

// getPressureLevel determines the pressure level from pressure value
func (mm *MemoryMonitor) getPressureLevel(pressure float64) PressureLevel {
	if pressure >= mm.criticalThreshold {
		return PressureCritical
	} else if pressure >= mm.highPressureThreshold {
		return PressureHigh
	} else if pressure >= 0.6 {
		return PressureMedium
	}
	return PressureLow
}

// handleHighPressure handles high/critical memory pressure
func (mm *MemoryMonitor) handleHighPressure(pressure float64, level PressureLevel, mem *runtime.MemStats) {
	logger.Warn("High memory pressure detected: %.1f%% (%s)", pressure*100, level.String())
	
	// Force garbage collection
	lastGC := time.Unix(0, atomic.LoadInt64(&mm.lastGCTime))
	if time.Since(lastGC) > 10*time.Second { // Don't GC too frequently
		logger.Info("Triggering garbage collection due to memory pressure")
		runtime.GC()
		atomic.StoreInt64(&mm.lastGCTime, time.Now().UnixNano())
		atomic.AddInt64(&mm.gcTriggerCount, 1)
	}
	
	// For critical pressure, take emergency measures
	if level == PressureCritical {
		logger.Error("CRITICAL memory pressure: %.1f%% - Taking emergency measures", pressure*100)
		
		// Disable metrics globally as emergency measure
		DisableMetricsGlobally()
		logger.Warn("Metrics collection disabled due to critical memory pressure")
		
		// Force a more aggressive GC
		runtime.GC()
		runtime.GC() // Double GC for critical situations
	}
	
	atomic.AddInt64(&mm.pressureReliefCount, 1)
}

// handleMediumPressure handles medium memory pressure
func (mm *MemoryMonitor) handleMediumPressure(pressure float64, level PressureLevel, mem *runtime.MemStats) {
	logger.Debug("Medium memory pressure: %.1f%% - Preparing for cleanup", pressure*100)
	
	// Re-enable metrics if they were disabled and pressure has reduced
	if pressure < 0.7 && atomic.LoadInt64(&metricsDisabledGlobally) > 0 {
		EnableMetricsGlobally()
		logger.Info("Metrics collection re-enabled - memory pressure reduced to %.1f%%", pressure*100)
	}
}

// notifyPressureCallbacks notifies all registered callbacks
func (mm *MemoryMonitor) notifyPressureCallbacks(pressure float64, level PressureLevel) {
	mm.callbackMutex.RLock()
	callbacks := make([]PressureReliefCallback, len(mm.pressureCallbacks))
	copy(callbacks, mm.pressureCallbacks)
	mm.callbackMutex.RUnlock()
	
	for _, callback := range callbacks {
		func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("Memory pressure callback panicked: %v", r)
				}
			}()
			callback(pressure, level)
		}()
	}
}

// float64ToBits converts float64 to int64 bits for atomic operations
func float64ToBits(f float64) int64 {
	return int64(f * 1e9) // Store as nanoseconds for precision
}

// float64FromBits converts int64 bits back to float64
func float64FromBits(bits int64) float64 {
	return float64(bits) / 1e9
}

// Global memory monitor instance
var globalMemoryMonitor *MemoryMonitor

// InitializeMemoryMonitor initializes the global memory monitor
func InitializeMemoryMonitor() *MemoryMonitor {
	if globalMemoryMonitor == nil {
		globalMemoryMonitor = NewMemoryMonitor()
	}
	return globalMemoryMonitor
}

// GetGlobalMemoryMonitor returns the global memory monitor instance
func GetGlobalMemoryMonitor() *MemoryMonitor {
	return globalMemoryMonitor
}