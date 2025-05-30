package models

import (
	"sync"
)

// StringIntern provides a global string interning pool to reduce memory usage
// for frequently repeated strings like tags
type StringIntern struct {
	mu      sync.RWMutex
	strings map[string]string
}

// globalIntern is the singleton instance
var globalIntern = &StringIntern{
	strings: make(map[string]string),
}

// Intern returns an interned version of the string
// If the string already exists in the pool, it returns the pooled version
// Otherwise, it adds the string to the pool
func Intern(s string) string {
	// Fast path - check if already interned
	globalIntern.mu.RLock()
	if interned, ok := globalIntern.strings[s]; ok {
		globalIntern.mu.RUnlock()
		return interned
	}
	globalIntern.mu.RUnlock()
	
	// Slow path - add to pool
	globalIntern.mu.Lock()
	defer globalIntern.mu.Unlock()
	
	// Double check in case another goroutine added it
	if interned, ok := globalIntern.strings[s]; ok {
		return interned
	}
	
	// Add to pool
	globalIntern.strings[s] = s
	return s
}

// InternSlice interns all strings in a slice
func InternSlice(strings []string) []string {
	for i, s := range strings {
		strings[i] = Intern(s)
	}
	return strings
}

// Size returns the number of interned strings
func Size() int {
	globalIntern.mu.RLock()
	defer globalIntern.mu.RUnlock()
	return len(globalIntern.strings)
}

// Clear removes all interned strings (use with caution)
func Clear() {
	globalIntern.mu.Lock()
	defer globalIntern.mu.Unlock()
	globalIntern.strings = make(map[string]string)
}