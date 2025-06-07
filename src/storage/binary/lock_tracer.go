package binary

import (
	"fmt"
	"runtime"
	"sync"
	"time"
	"entitydb/logger"
)

// Global lock tracer for debugging
var lockTracer = &LockTracer{
	active: make(map[string]*LockInfo),
}

type LockInfo struct {
	ID        string
	Type      LockType
	Holder    string
	Stack     string
	Timestamp time.Time
}

type LockTracer struct {
	mu     sync.Mutex
	active map[string]*LockInfo
}

func (lt *LockTracer) RecordAcquisition(entityID string, lockType LockType) {
	lt.mu.Lock()
	defer lt.mu.Unlock()
	
	// Get caller info
	pc, file, line, _ := runtime.Caller(2)
	fn := runtime.FuncForPC(pc)
	
	info := &LockInfo{
		ID:        entityID,
		Type:      lockType,
		Holder:    fmt.Sprintf("%s:%d %s", file, line, fn.Name()),
		Stack:     string(getStack()),
		Timestamp: time.Now(),
	}
	
	// Check if lock is already held
	if existing, exists := lt.active[entityID]; exists {
		logger.Warn("[LOCK_CONFLICT] Entity %s already locked by %s, new request from %s",
			entityID, existing.Holder, info.Holder)
		logger.Warn("[LOCK_CONFLICT] Existing stack:\n%s", existing.Stack)
		logger.Warn("[LOCK_CONFLICT] New stack:\n%s", info.Stack)
	}
	
	lt.active[entityID] = info
	logger.Debug("[LOCK_ACQUIRE] Entity: %s, Type: %v, Holder: %s", entityID, lockType, info.Holder)
}

func (lt *LockTracer) RecordRelease(entityID string, lockType LockType) {
	lt.mu.Lock()
	defer lt.mu.Unlock()
	
	if info, exists := lt.active[entityID]; exists {
		duration := time.Since(info.Timestamp)
		logger.Debug("[LOCK_RELEASE] Entity: %s, Type: %v, Held for: %v", entityID, lockType, duration)
		delete(lt.active, entityID)
		
		// Warn about long-held locks
		if duration > 100*time.Millisecond {
			logger.Warn("[LOCK_SLOW] Entity %s held for %v by %s", entityID, duration, info.Holder)
		}
	}
}

func (lt *LockTracer) GetActiveLocks() []LockInfo {
	lt.mu.Lock()
	defer lt.mu.Unlock()
	
	var locks []LockInfo
	for _, info := range lt.active {
		locks = append(locks, *info)
	}
	return locks
}

func getStack() []byte {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return buf[:n]
}

// TracedLockManager wraps LockManager with tracing
type TracedLockManager struct {
	*LockManager
	traceEnabled bool
}

func NewTracedLockManager(lm *LockManager) *TracedLockManager {
	return &TracedLockManager{
		LockManager:  lm,
		traceEnabled: true,
	}
}

func (tlm *TracedLockManager) AcquireEntityLock(entityID string, lockType LockType) {
	if tlm.traceEnabled {
		lockTracer.RecordAcquisition(entityID, lockType)
	}
	tlm.LockManager.AcquireEntityLock(entityID, lockType)
}

func (tlm *TracedLockManager) ReleaseEntityLock(entityID string, lockType LockType) {
	tlm.LockManager.ReleaseEntityLock(entityID, lockType)
	if tlm.traceEnabled {
		lockTracer.RecordRelease(entityID, lockType)
	}
}

// PrintDeadlockDiagnostics prints current lock state for debugging
func PrintDeadlockDiagnostics() {
	locks := lockTracer.GetActiveLocks()
	if len(locks) == 0 {
		logger.Info("[DEADLOCK_DIAG] No active locks")
		return
	}
	
	logger.Warn("[DEADLOCK_DIAG] Active locks: %d", len(locks))
	for _, lock := range locks {
		age := time.Since(lock.Timestamp)
		logger.Warn("[DEADLOCK_DIAG] Entity: %s, Type: %v, Age: %v, Holder: %s",
			lock.ID, lock.Type, age, lock.Holder)
		if age > 5*time.Second {
			logger.Error("[DEADLOCK_DIAG] POTENTIAL DEADLOCK - Lock held for %v", age)
		}
	}
}