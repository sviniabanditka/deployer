package database

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const TaskScheduledBackup = "backup:scheduled"

// ScheduledBackupPayload contains the data needed to process a scheduled backup.
type ScheduledBackupPayload struct {
	DatabaseID string `json:"database_id"`
	UserID     string `json:"user_id"`
}

// NewScheduledBackupTask creates a new Asynq task for a scheduled backup.
func NewScheduledBackupTask(databaseID, userID uuid.UUID) (*asynq.Task, error) {
	payload, err := json.Marshal(ScheduledBackupPayload{
		DatabaseID: databaseID.String(),
		UserID:     userID.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal scheduled backup payload: %w", err)
	}
	return asynq.NewTask(TaskScheduledBackup, payload), nil
}

// ParseScheduledBackupPayload extracts the payload from an Asynq task.
func ParseScheduledBackupPayload(task *asynq.Task) (*ScheduledBackupPayload, error) {
	var p ScheduledBackupPayload
	if err := json.Unmarshal(task.Payload(), &p); err != nil {
		return nil, fmt.Errorf("failed to unmarshal scheduled backup payload: %w", err)
	}
	return &p, nil
}

// ProcessScheduledBackup is the Asynq task handler for scheduled database backups.
func (s *DatabaseService) ProcessScheduledBackup(ctx context.Context, task *asynq.Task) error {
	payload, err := ParseScheduledBackupPayload(task)
	if err != nil {
		return fmt.Errorf("failed to parse payload: %w", err)
	}

	dbID, err := uuid.Parse(payload.DatabaseID)
	if err != nil {
		return fmt.Errorf("invalid database ID: %w", err)
	}
	userID, err := uuid.Parse(payload.UserID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	log.Printf("starting scheduled backup for database %s", dbID)

	// 1. Create the backup.
	backup, err := s.CreateBackup(ctx, userID, dbID)
	if err != nil {
		return fmt.Errorf("scheduled backup failed: %w", err)
	}

	log.Printf("scheduled backup completed: %s (size: %d bytes)", backup.ID, backup.FileSize)

	// 2. Get the database to check retention policy.
	mdb, err := s.dbRepo.GetByID(ctx, dbID)
	if err != nil {
		return fmt.Errorf("failed to get database for retention check: %w", err)
	}

	retention := mdb.BackupRetention
	if retention <= 0 {
		retention = 7
	}

	// 3. Clean up old backups beyond retention limit.
	backups, err := s.dbRepo.ListBackups(ctx, dbID)
	if err != nil {
		return fmt.Errorf("failed to list backups for cleanup: %w", err)
	}

	if len(backups) > retention {
		// Backups are ordered by created_at DESC, so the oldest are at the end.
		for _, old := range backups[retention:] {
			// Remove the backup file from disk.
			if old.FilePath != "" {
				_ = os.Remove(old.FilePath)
			}
			// Remove the backup record from the database.
			if err := s.dbRepo.DeleteBackup(ctx, old.ID); err != nil {
				log.Printf("warning: failed to delete old backup record %s: %v", old.ID, err)
			}
		}
		log.Printf("cleaned up %d old backups for database %s", len(backups)-retention, dbID)
	}

	// 4. Update storage_used on the managed database.
	var totalSize int64
	currentBackups, err := s.dbRepo.ListBackups(ctx, dbID)
	if err == nil {
		for _, b := range currentBackups {
			totalSize += b.FileSize
		}
	}
	mdb.StorageUsed = totalSize
	_ = s.dbRepo.Update(ctx, mdb)

	return nil
}
