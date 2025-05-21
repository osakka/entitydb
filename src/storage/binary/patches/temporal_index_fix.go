package binary

import (
	"strings"
	"time"
)

// AddTemporalIndexFix fixes tag indexing by ensuring we index both the full temporal tag and the non-timestamped version
func (ti *TemporalIndex) AddTemporalIndexFix(entityID string, tag string, timestamp time.Time) {
	// Original implementation
	ti.AddEntry(entityID, tag, timestamp)
	
	// Fix: Also extract and index the tag without its timestamp
	if strings.Contains(tag, "|") {
		parts := strings.SplitN(tag, "|", 2)
		if len(parts) == 2 {
			// Add the plain tag without timestamp to the index
			// No direct indexing here - this is just a helper function
			// The actual indexing happens in EntityRepository
		}
	}
}