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

// defaultStringInterner is the singleton instance for string interning
var defaultStringInterner = &StringIntern{
	strings: make(map[string]string),
}

// Intern returns an interned version of the string
// If the string already exists in the pool, it returns the pooled version
// Otherwise, it adds the string to the pool
func Intern(s string) string {
	// Fast path - check if already interned
	defaultStringInterner.mu.RLock()
	if interned, ok := defaultStringInterner.strings[s]; ok {
		defaultStringInterner.mu.RUnlock()
		return interned
	}
	defaultStringInterner.mu.RUnlock()
	
	// Slow path - add to pool
	defaultStringInterner.mu.Lock()
	defer defaultStringInterner.mu.Unlock()
	
	// Double check in case another goroutine added it
	if interned, ok := defaultStringInterner.strings[s]; ok {
		return interned
	}
	
	// Add to pool
	defaultStringInterner.strings[s] = s
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
	defaultStringInterner.mu.RLock()
	defer defaultStringInterner.mu.RUnlock()
	return len(defaultStringInterner.strings)
}

// Clear removes all interned strings (use with caution)
func Clear() {
	defaultStringInterner.mu.Lock()
	defer defaultStringInterner.mu.Unlock()
	defaultStringInterner.strings = make(map[string]string)
}