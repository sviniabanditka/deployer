package database

import (
	"context"
	"fmt"
	"log"

	"github.com/deployer/api/internal/repository"
	"github.com/hibiken/asynq"
)

// BackupScheduler manages periodic backup tasks for databases with auto_backup enabled.
type BackupScheduler struct {
	dbRepo    repository.DatabaseRepository
	scheduler *asynq.Scheduler
}

// NewBackupScheduler creates a new BackupScheduler.
func NewBackupScheduler(
	dbRepo repository.DatabaseRepository,
	redisOpt asynq.RedisClientOpt,
) *BackupScheduler {
	scheduler := asynq.NewScheduler(redisOpt, nil)
	return &BackupScheduler{
		dbRepo:    dbRepo,
		scheduler: scheduler,
	}
}

// StartScheduler registers periodic backup tasks for all databases with auto_backup enabled
// and starts the Asynq scheduler.
func (bs *BackupScheduler) StartScheduler(ctx context.Context) error {
	// Fetch all databases with auto_backup enabled.
	databases, err := bs.dbRepo.ListAutoBackupDatabases(ctx)
	if err != nil {
		return fmt.Errorf("failed to list auto-backup databases: %w", err)
	}

	for _, mdb := range databases {
		schedule := mdb.BackupSchedule
		if schedule == "" {
			schedule = "0 2 * * *" // default: 2 AM daily
		}

		task, err := NewScheduledBackupTask(mdb.ID, mdb.UserID)
		if err != nil {
			log.Printf("warning: failed to create backup task for database %s: %v", mdb.ID, err)
			continue
		}

		entryID, err := bs.scheduler.Register(schedule, task,
			asynq.Queue("default"),
			asynq.TaskID(fmt.Sprintf("backup:%s", mdb.ID.String())),
		)
		if err != nil {
			log.Printf("warning: failed to register backup schedule for database %s: %v", mdb.ID, err)
			continue
		}

		log.Printf("registered backup schedule for database %s (entry: %s, cron: %s)", mdb.ID, entryID, schedule)
	}

	log.Printf("backup scheduler starting with %d databases", len(databases))

	// Run the scheduler (blocks until shutdown).
	return bs.scheduler.Run()
}

// Shutdown gracefully shuts down the backup scheduler.
func (bs *BackupScheduler) Shutdown() {
	bs.scheduler.Shutdown()
}
