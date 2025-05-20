package binary

import (
	"encoding/binary"
	"entitydb/logger"
	"fmt"
)

// RepairIndex attempts to fix corrupted index entries
func (w *Writer) RepairIndex() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	logger.Info("Starting index repair...")
	
	repaired := 0
	for id, entry := range w.index {
		// Check if offset looks corrupted (too large)
		if entry.Offset > uint64(w.header.FileSize) {
			// Attempt to fix by using only the low 32 bits
			oldOffset := entry.Offset
			newOffset := entry.Offset & 0xFFFFFFFF
			
			// Verify the new offset is reasonable
			if newOffset < uint64(w.header.FileSize) {
				logger.Info("Repairing index entry for %s: offset %d -> %d", id, oldOffset, newOffset)
				entry.Offset = newOffset
				repaired++
			} else {
				logger.Warn("Cannot repair index entry for %s: offset %d still invalid", id, oldOffset)
			}
		}
	}
	
	if repaired > 0 {
		logger.Info("Repaired %d index entries", repaired)
		
		// Rewrite the index
		if err := w.rewriteIndex(); err != nil {
			return fmt.Errorf("failed to rewrite index: %w", err)
		}
		
		// Sync to disk
		if err := w.file.Sync(); err != nil {
			return fmt.Errorf("failed to sync: %w", err)
		}
	} else {
		logger.Info("No index entries needed repair")
	}
	
	return nil
}

// rewriteIndex rewrites the entire index section
func (w *Writer) rewriteIndex() error {
	// Seek to index offset
	if _, err := w.file.Seek(int64(w.header.EntityIndexOffset), 0); err != nil {
		return fmt.Errorf("failed to seek to index: %w", err)
	}
	
	// Write all index entries
	for id, entry := range w.index {
		logger.Debug("Writing index entry for %s: offset=%d, size=%d", id, entry.Offset, entry.Size)
		
		if err := binary.Write(w.file, binary.LittleEndian, entry.EntityID); err != nil {
			return err
		}
		if err := binary.Write(w.file, binary.LittleEndian, entry.Offset); err != nil {
			return err
		}
		if err := binary.Write(w.file, binary.LittleEndian, entry.Size); err != nil {
			return err
		}
		if err := binary.Write(w.file, binary.LittleEndian, entry.Flags); err != nil {
			return err
		}
	}
	
	return nil
}