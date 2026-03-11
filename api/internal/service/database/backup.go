package database

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/deployer/api/internal/model"
	"github.com/docker/docker/api/types/container"
	"github.com/google/uuid"
)

// CreateBackup creates a database backup by executing a dump command inside the container.
func (s *DatabaseService) CreateBackup(ctx context.Context, userID, dbID uuid.UUID) (*model.DatabaseBackup, error) {
	mdb, err := s.dbRepo.GetByID(ctx, dbID)
	if err != nil {
		return nil, err
	}
	if mdb.UserID != userID {
		return nil, fmt.Errorf("unauthorized")
	}
	if mdb.Status != model.DatabaseStatusRunning {
		return nil, fmt.Errorf("database is not running")
	}

	// Determine dump command
	var cmd []string
	var fileName string
	ts := time.Now().UTC().Format("20060102T150405")

	switch mdb.Engine {
	case model.EnginePostgres:
		cmd = []string{"pg_dump", "-U", mdb.Username, "-d", mdb.DBName, "-F", "c"}
		fileName = fmt.Sprintf("%s_%s.pgdump", mdb.Name, ts)
	case model.EngineMySQL:
		cmd = []string{"mysqldump", "-u", mdb.Username, fmt.Sprintf("-p%s", mdb.Password), mdb.DBName}
		fileName = fmt.Sprintf("%s_%s.sql", mdb.Name, ts)
	case model.EngineMongoDB:
		cmd = []string{"mongodump", "--username", mdb.Username, "--password", mdb.Password,
			"--db", mdb.DBName, "--authenticationDatabase", "admin", "--archive"}
		fileName = fmt.Sprintf("%s_%s.archive", mdb.Name, ts)
	case model.EngineRedis:
		return nil, fmt.Errorf("backup is not supported for Redis")
	default:
		return nil, fmt.Errorf("unsupported engine: %s", mdb.Engine)
	}

	// Create backup record
	backup := &model.DatabaseBackup{
		DatabaseID: mdb.ID,
		FilePath:   filepath.Join(s.cfg.BackupDir, fileName),
		Status:     "creating",
	}
	if err := s.dbRepo.CreateBackup(ctx, backup); err != nil {
		return nil, fmt.Errorf("failed to create backup record: %w", err)
	}

	// Execute dump inside container
	output, err := s.execInContainer(ctx, mdb.ContainerID, cmd)
	if err != nil {
		_ = s.dbRepo.UpdateBackupStatus(ctx, backup.ID, "failed", 0)
		return nil, fmt.Errorf("dump failed: %w", err)
	}

	// Ensure backup directory exists
	if err := os.MkdirAll(s.cfg.BackupDir, 0o755); err != nil {
		_ = s.dbRepo.UpdateBackupStatus(ctx, backup.ID, "failed", 0)
		return nil, fmt.Errorf("failed to create backup dir: %w", err)
	}

	// Write dump to file
	if err := os.WriteFile(backup.FilePath, []byte(output), 0o600); err != nil {
		_ = s.dbRepo.UpdateBackupStatus(ctx, backup.ID, "failed", 0)
		return nil, fmt.Errorf("failed to write backup file: %w", err)
	}

	info, _ := os.Stat(backup.FilePath)
	fileSize := int64(0)
	if info != nil {
		fileSize = info.Size()
	}

	if err := s.dbRepo.UpdateBackupStatus(ctx, backup.ID, "completed", fileSize); err != nil {
		return nil, err
	}

	backup.Status = "completed"
	backup.FileSize = fileSize
	return backup, nil
}

// RestoreBackup restores a database from a backup file.
func (s *DatabaseService) RestoreBackup(ctx context.Context, userID, dbID, backupID uuid.UUID) error {
	mdb, err := s.dbRepo.GetByID(ctx, dbID)
	if err != nil {
		return err
	}
	if mdb.UserID != userID {
		return fmt.Errorf("unauthorized")
	}
	if mdb.Status != model.DatabaseStatusRunning {
		return fmt.Errorf("database is not running")
	}

	backups, err := s.dbRepo.ListBackups(ctx, dbID)
	if err != nil {
		return err
	}

	var backup *model.DatabaseBackup
	for i := range backups {
		if backups[i].ID == backupID {
			backup = &backups[i]
			break
		}
	}
	if backup == nil {
		return fmt.Errorf("backup not found")
	}

	// Read backup file
	data, err := os.ReadFile(backup.FilePath)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %w", err)
	}

	// Determine restore command
	var cmd []string
	switch mdb.Engine {
	case model.EnginePostgres:
		cmd = []string{"pg_restore", "-U", mdb.Username, "-d", mdb.DBName, "--clean", "--if-exists"}
	case model.EngineMySQL:
		cmd = []string{"mysql", "-u", mdb.Username, fmt.Sprintf("-p%s", mdb.Password), mdb.DBName}
	case model.EngineMongoDB:
		cmd = []string{"mongorestore", "--username", mdb.Username, "--password", mdb.Password,
			"--db", mdb.DBName, "--authenticationDatabase", "admin", "--archive"}
	default:
		return fmt.Errorf("restore not supported for engine: %s", mdb.Engine)
	}

	// Execute restore with stdin
	if err := s.execInContainerWithStdin(ctx, mdb.ContainerID, cmd, data); err != nil {
		return fmt.Errorf("restore failed: %w", err)
	}

	return nil
}

// ListBackups returns all backups for a database.
func (s *DatabaseService) ListBackups(ctx context.Context, userID, dbID uuid.UUID) ([]model.DatabaseBackup, error) {
	mdb, err := s.dbRepo.GetByID(ctx, dbID)
	if err != nil {
		return nil, err
	}
	if mdb.UserID != userID {
		return nil, fmt.Errorf("unauthorized")
	}
	return s.dbRepo.ListBackups(ctx, dbID)
}

// execInContainer runs a command inside a container and returns the stdout output.
func (s *DatabaseService) execInContainer(ctx context.Context, containerID string, cmd []string) (string, error) {
	execCfg := container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
	}

	execID, err := s.docker.ContainerExecCreate(ctx, containerID, execCfg)
	if err != nil {
		return "", fmt.Errorf("exec create failed: %w", err)
	}

	resp, err := s.docker.ContainerExecAttach(ctx, execID.ID, container.ExecAttachOptions{})
	if err != nil {
		return "", fmt.Errorf("exec attach failed: %w", err)
	}
	defer resp.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, resp.Reader); err != nil {
		return "", fmt.Errorf("exec read failed: %w", err)
	}

	return buf.String(), nil
}

// execInContainerWithStdin runs a command inside a container, piping data to stdin.
func (s *DatabaseService) execInContainerWithStdin(ctx context.Context, containerID string, cmd []string, stdinData []byte) error {
	execCfg := container.ExecOptions{
		Cmd:          cmd,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
	}

	execID, err := s.docker.ContainerExecCreate(ctx, containerID, execCfg)
	if err != nil {
		return fmt.Errorf("exec create failed: %w", err)
	}

	resp, err := s.docker.ContainerExecAttach(ctx, execID.ID, container.ExecAttachOptions{})
	if err != nil {
		return fmt.Errorf("exec attach failed: %w", err)
	}
	defer resp.Close()

	// Write stdin data
	if _, err := resp.Conn.Write(stdinData); err != nil {
		return fmt.Errorf("exec write stdin failed: %w", err)
	}
	resp.CloseWrite()

	// Drain output
	_, _ = io.Copy(io.Discard, resp.Reader)

	return nil
}
