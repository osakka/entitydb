package binary

import (
	"encoding/binary"
	"time"
	"fmt"
	"strconv"
)

// TimestampFormat defines our optimized temporal format
// Using Unix nanoseconds as int64 for fast comparisons
type TimestampFormat struct {
	// We'll use a compact binary format for storage
	// and string format for API compatibility
}

// CompactTimestamp converts RFC3339Nano to compact int64 nanoseconds
func CompactTimestamp(rfc3339 string) (int64, error) {
	t, err := time.Parse(time.RFC3339Nano, rfc3339)
	if err != nil {
		return 0, err
	}
	return t.UnixNano(), nil
}

// ExpandTimestamp converts int64 nanoseconds back to RFC3339Nano
func ExpandTimestamp(nanos int64) string {
	return time.Unix(0, nanos).Format(time.RFC3339Nano)
}

// BinaryTimestamp encodes a timestamp as 8 bytes
func BinaryTimestamp(nanos int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(nanos))
	return b
}

// ParseBinaryTimestamp decodes a timestamp from 8 bytes
func ParseBinaryTimestamp(b []byte) int64 {
	return int64(binary.BigEndian.Uint64(b))
}

// FormatTagWithTimestamp creates a temporal tag with optimized format
// Format: NANOS|tag where NANOS is the int64 nanosecond timestamp
func FormatTagWithTimestamp(tag string, timestamp int64) string {
	return fmt.Sprintf("%d|%s", timestamp, tag)
}

// ParseTemporalTag extracts timestamp and tag from temporal format
func ParseTemporalTag(temporalTag string) (int64, string, error) {
	// Find the delimiter
	for i, ch := range temporalTag {
		if ch == '|' {
			// Parse the timestamp
			nanos, err := strconv.ParseInt(temporalTag[:i], 10, 64)
			if err != nil {
				return 0, "", err
			}
			return nanos, temporalTag[i+1:], nil
		}
	}
	return 0, "", fmt.Errorf("invalid temporal tag format")
}

// CompareTimestamps provides fast timestamp comparison
func CompareTimestamps(t1, t2 int64) int {
	if t1 < t2 {
		return -1
	} else if t1 > t2 {
		return 1
	}
	return 0
}

// TimestampRange represents a time range for queries
type TimestampRange struct {
	Start int64
	End   int64
}

// Contains checks if a timestamp is within the range
func (tr TimestampRange) Contains(timestamp int64) bool {
	return timestamp >= tr.Start && timestamp <= tr.End
}

// TimeBucket provides efficient time bucketing for indexes
type TimeBucket struct {
	BucketSize int64 // Size in nanoseconds
}

// Standard bucket sizes
var (
	SecondBucket = TimeBucket{1e9}
	MinuteBucket = TimeBucket{60 * 1e9}
	HourBucket   = TimeBucket{3600 * 1e9}
	DayBucket    = TimeBucket{86400 * 1e9}
)

// GetBucket returns the bucket ID for a timestamp
func (tb TimeBucket) GetBucket(timestamp int64) int64 {
	return (timestamp / tb.BucketSize) * tb.BucketSize
}

// GetBucketRange returns the range of a bucket
func (tb TimeBucket) GetBucketRange(bucketID int64) TimestampRange {
	return TimestampRange{
		Start: bucketID,
		End:   bucketID + tb.BucketSize - 1,
	}
}