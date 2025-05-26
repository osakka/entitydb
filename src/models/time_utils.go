package models

import (
	"fmt"
	"strconv"
	"time"
)

// Centralized time utilities for EntityDB
// All timestamp operations should use these functions for consistency

// Now returns the current time as nanoseconds since Unix epoch
// This is the primary timestamp function for EntityDB
func Now() int64 {
	return time.Now().UnixNano()
}

// NowString returns the current time as a nanosecond epoch string
// Used for compatibility with existing string-based timestamp fields
func NowString() string {
	return FormatNanosToString(Now())
}

// FormatNanosToString converts nanosecond epoch to string
func FormatNanosToString(nanos int64) string {
	return fmt.Sprintf("%d", nanos)
}

// ParseStringToNanos converts string nanosecond epoch back to int64
func ParseStringToNanos(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// ToRFC3339 converts nanosecond epoch to RFC3339 for API responses
// Only use this for external API compatibility, not internal storage
func ToRFC3339(nanos int64) string {
	return time.Unix(0, nanos).Format(time.RFC3339Nano)
}

// FromRFC3339 converts RFC3339 string to nanosecond epoch
// Only use this for external API input parsing
func FromRFC3339(rfc3339 string) (int64, error) {
	t, err := time.Parse(time.RFC3339Nano, rfc3339)
	if err != nil {
		return 0, err
	}
	return t.UnixNano(), nil
}

// TimeAgo returns how many nanoseconds ago a timestamp was
func TimeAgo(nanos int64) int64 {
	return Now() - nanos
}

// IsRecent checks if a timestamp is within the last N nanoseconds
func IsRecent(nanos int64, withinNanos int64) bool {
	return TimeAgo(nanos) <= withinNanos
}

// FormatTemporalTag creates a tag with nanosecond timestamp prefix
// Format: "NANOS|tag" where NANOS is int64 nanoseconds since epoch
func FormatTemporalTag(tag string) string {
	return fmt.Sprintf("%d|%s", Now(), tag)
}

// FormatTemporalTagAt creates a tag with specific timestamp
func FormatTemporalTagAt(tag string, nanos int64) string {
	return fmt.Sprintf("%d|%s", nanos, tag)
}

// ParseTemporalTag extracts timestamp and tag from temporal format
func ParseTemporalTag(temporalTag string) (int64, string, error) {
	for i, ch := range temporalTag {
		if ch == '|' {
			nanos, err := strconv.ParseInt(temporalTag[:i], 10, 64)
			if err != nil {
				return 0, "", err
			}
			return nanos, temporalTag[i+1:], nil
		}
	}
	return 0, "", fmt.Errorf("invalid temporal tag format: %s", temporalTag)
}

// Time constants for convenience
const (
	Nanosecond  = int64(1)
	Microsecond = 1000 * Nanosecond
	Millisecond = 1000 * Microsecond
	Second      = 1000 * Millisecond
	Minute      = 60 * Second
	Hour        = 60 * Minute
	Day         = 24 * Hour
	Week        = 7 * Day
)

// MinutesAgo returns timestamp from N minutes ago
func MinutesAgo(minutes int) int64 {
	return Now() - int64(minutes)*Minute
}

// HoursAgo returns timestamp from N hours ago  
func HoursAgo(hours int) int64 {
	return Now() - int64(hours)*Hour
}

// DaysAgo returns timestamp from N days ago
func DaysAgo(days int) int64 {
	return Now() - int64(days)*Day
}

// Future time helpers
func InMinutes(minutes int) int64 {
	return Now() + int64(minutes)*Minute
}

func InHours(hours int) int64 {
	return Now() + int64(hours)*Hour
}

func InDays(days int) int64 {
	return Now() + int64(days)*Day
}