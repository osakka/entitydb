package binary

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
	
	"entitydb/logger"
	"entitydb/models"
)

// TransactionManager handles atomic multi-file operations
type TransactionManager struct {
	mu           sync.Mutex
	transactions map[string]*Transaction
	dataPath     string
}

// Transaction represents an atomic operation across multiple files
type Transaction struct {
	ID            string
	StartTime     time.Time
	State         TransactionState
	Operations    []TransactionOp
	Backups       map[string]string // original file -> backup path
	TempFiles     map[string]string // final file -> temp path
	mu            sync.Mutex
}

// TransactionState represents the state of a transaction
type TransactionState int

const (
	TxStateActive TransactionState = iota
	TxStatePrepared
	TxStateCommitted
	TxStateAborted
)

// TransactionOp represents a single operation in a transaction
type TransactionOp struct {
	Type     string
	File     string
	Data     []byte
	Callback func() error
}

// NewTransactionManager creates a new transaction manager
func NewTransactionManager(dataPath string) *TransactionManager {
	return &TransactionManager{
		transactions: make(map[string]*Transaction),
		dataPath:     dataPath,
	}
}

// Begin starts a new transaction
func (tm *TransactionManager) Begin() *Transaction {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	
	tx := &Transaction{
		ID:        string(models.GenerateOperationID()),
		StartTime: time.Now(),
		State:     TxStateActive,
		Backups:   make(map[string]string),
		TempFiles: make(map[string]string),
	}
	
	tm.transactions[tx.ID] = tx
	
	op := models.StartOperation(models.OpTypeTransaction, tx.ID, map[string]interface{}{
		"transaction_id": tx.ID,
		"start_time":     tx.StartTime,
	})
	op.Complete()
	
	logger.Info("[Transaction] Started transaction %s", tx.ID)
	return tx
}

// AddOperation adds an operation to the transaction
func (tx *Transaction) AddOperation(op TransactionOp) {
	tx.mu.Lock()
	defer tx.mu.Unlock()
	
	if tx.State != TxStateActive {
		logger.Error("[Transaction] Cannot add operation to transaction %s in state %v", tx.ID, tx.State)
		return
	}
	
	tx.Operations = append(tx.Operations, op)
	logger.Debug("[Transaction] Added operation to transaction %s: %s on %s", tx.ID, op.Type, op.File)
}

// Prepare prepares the transaction for commit (two-phase commit)
func (tx *Transaction) Prepare() error {
	tx.mu.Lock()
	defer tx.mu.Unlock()
	
	if tx.State != TxStateActive {
		return fmt.Errorf("transaction %s is not active", tx.ID)
	}
	
	op := models.StartOperation(models.OpTypeTransaction, tx.ID, map[string]interface{}{
		"phase": "prepare",
		"operations": len(tx.Operations),
	})
	defer op.Complete()
	
	logger.Info("[Transaction] Preparing transaction %s with %d operations", tx.ID, len(tx.Operations))
	
	// Create backups of all files that will be modified
	for _, txOp := range tx.Operations {
		if txOp.Type == "write" || txOp.Type == "update" {
			backupPath := fmt.Sprintf("%s.tx_%s.backup", txOp.File, tx.ID)
			
			// Check if file exists
			if _, err := os.Stat(txOp.File); err == nil {
				// Create backup
				if err := copyFile(txOp.File, backupPath); err != nil {
					op.Fail(err)
					tx.Rollback()
					return fmt.Errorf("failed to backup %s: %v", txOp.File, err)
				}
				tx.Backups[txOp.File] = backupPath
				logger.Debug("[Transaction] Created backup: %s -> %s", txOp.File, backupPath)
			}
		}
	}
	
	// Prepare all operations (write to temp files)
	for _, txOp := range tx.Operations {
		if txOp.Type == "write" {
			tempPath := fmt.Sprintf("%s.tx_%s.tmp", txOp.File, tx.ID)
			
			// Write to temp file
			if err := os.WriteFile(tempPath, txOp.Data, 0644); err != nil {
				op.Fail(err)
				tx.Rollback()
				return fmt.Errorf("failed to write temp file %s: %v", tempPath, err)
			}
			
			tx.TempFiles[txOp.File] = tempPath
			logger.Debug("[Transaction] Wrote temp file: %s", tempPath)
		} else if txOp.Callback != nil {
			// For callbacks, we can't prepare them, just note they exist
			logger.Debug("[Transaction] Callback operation prepared for %s", txOp.File)
		}
	}
	
	tx.State = TxStatePrepared
	logger.Info("[Transaction] Transaction %s prepared successfully", tx.ID)
	return nil
}

// Commit commits the transaction
func (tx *Transaction) Commit() error {
	tx.mu.Lock()
	defer tx.mu.Unlock()
	
	if tx.State != TxStatePrepared {
		return fmt.Errorf("transaction %s is not prepared", tx.ID)
	}
	
	op := models.StartOperation(models.OpTypeTransaction, tx.ID, map[string]interface{}{
		"phase": "commit",
	})
	defer op.Complete()
	
	logger.Info("[Transaction] Committing transaction %s", tx.ID)
	
	// Move all temp files to final locations
	for finalPath, tempPath := range tx.TempFiles {
		if err := os.Rename(tempPath, finalPath); err != nil {
			op.Fail(err)
			logger.Error("[Transaction] Failed to rename %s to %s: %v", tempPath, finalPath, err)
			// Try to rollback what we can
			tx.Rollback()
			return fmt.Errorf("failed to commit file %s: %v", finalPath, err)
		}
		logger.Debug("[Transaction] Committed: %s", finalPath)
	}
	
	// Execute callbacks
	for _, txOp := range tx.Operations {
		if txOp.Callback != nil {
			if err := txOp.Callback(); err != nil {
				op.Fail(err)
				logger.Error("[Transaction] Callback failed: %v", err)
				// At this point we can't rollback file changes
				return fmt.Errorf("callback failed: %v", err)
			}
		}
	}
	
	// Clean up backups
	for _, backupPath := range tx.Backups {
		os.Remove(backupPath)
		logger.Debug("[Transaction] Removed backup: %s", backupPath)
	}
	
	tx.State = TxStateCommitted
	logger.Info("[Transaction] Transaction %s committed successfully", tx.ID)
	return nil
}

// Rollback rolls back the transaction
func (tx *Transaction) Rollback() error {
	tx.mu.Lock()
	defer tx.mu.Unlock()
	
	if tx.State == TxStateCommitted || tx.State == TxStateAborted {
		return fmt.Errorf("transaction %s already finalized", tx.ID)
	}
	
	op := models.StartOperation(models.OpTypeTransaction, tx.ID, map[string]interface{}{
		"phase": "rollback",
	})
	defer op.Complete()
	
	logger.Info("[Transaction] Rolling back transaction %s", tx.ID)
	
	// Remove temp files
	for _, tempPath := range tx.TempFiles {
		os.Remove(tempPath)
		logger.Debug("[Transaction] Removed temp file: %s", tempPath)
	}
	
	// Restore backups
	for originalPath, backupPath := range tx.Backups {
		if err := os.Rename(backupPath, originalPath); err != nil {
			logger.Error("[Transaction] Failed to restore backup %s to %s: %v", backupPath, originalPath, err)
		} else {
			logger.Debug("[Transaction] Restored backup: %s -> %s", backupPath, originalPath)
		}
	}
	
	tx.State = TxStateAborted
	logger.Info("[Transaction] Transaction %s rolled back", tx.ID)
	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

// WithTransaction executes a function within a transaction
func (tm *TransactionManager) WithTransaction(fn func(*Transaction) error) error {
	tx := tm.Begin()
	
	// Prepare phase
	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}
	
	// Prepare
	if err := tx.Prepare(); err != nil {
		return err
	}
	
	// Commit
	return tx.Commit()
}

// CleanupOldTransactions cleans up old transaction files
func (tm *TransactionManager) CleanupOldTransactions() error {
	pattern := filepath.Join(tm.dataPath, "*.tx_*.backup")
	backups, _ := filepath.Glob(pattern)
	
	pattern = filepath.Join(tm.dataPath, "*.tx_*.tmp")
	temps, _ := filepath.Glob(pattern)
	
	cleaned := 0
	for _, file := range append(backups, temps...) {
		// Check if file is older than 1 hour
		info, err := os.Stat(file)
		if err != nil {
			continue
		}
		
		if time.Since(info.ModTime()) > time.Hour {
			os.Remove(file)
			cleaned++
		}
	}
	
	if cleaned > 0 {
		logger.Info("[Transaction] Cleaned up %d old transaction files", cleaned)
	}
	
	return nil
}