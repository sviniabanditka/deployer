package repository

import (
	"context"
	"errors"
	"time"

	"github.com/deployer/api/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DatabaseRepository interface {
	Create(ctx context.Context, db *model.ManagedDatabase) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.ManagedDatabase, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]model.ManagedDatabase, error)
	Update(ctx context.Context, db *model.ManagedDatabase) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByAppID(ctx context.Context, appID uuid.UUID) ([]model.ManagedDatabase, error)
	CreateBackup(ctx context.Context, backup *model.DatabaseBackup) error
	ListBackups(ctx context.Context, databaseID uuid.UUID) ([]model.DatabaseBackup, error)
	UpdateBackupStatus(ctx context.Context, backupID uuid.UUID, status string, fileSize int64) error
	DeleteBackup(ctx context.Context, backupID uuid.UUID) error
	ListAutoBackupDatabases(ctx context.Context) ([]model.ManagedDatabase, error)
}

type databaseRepo struct {
	db *pgxpool.Pool
}

func NewDatabaseRepository(db *pgxpool.Pool) DatabaseRepository {
	return &databaseRepo{db: db}
}

const dbSelectColumns = `id, user_id, app_id, name, engine, version, status, host, port, db_name,
	username, password, connection_url, container_id, memory_limit, storage_used,
	storage_limit, auto_backup, backup_retention, backup_schedule, created_at, updated_at`

func scanManagedDatabase(row pgx.Row) (*model.ManagedDatabase, error) {
	var d model.ManagedDatabase
	err := row.Scan(
		&d.ID, &d.UserID, &d.AppID, &d.Name, &d.Engine, &d.Version, &d.Status,
		&d.Host, &d.Port, &d.DBName, &d.Username, &d.Password, &d.ConnectionURL,
		&d.ContainerID, &d.MemoryLimit, &d.StorageUsed, &d.StorageLimit,
		&d.AutoBackup, &d.BackupRetention, &d.BackupSchedule,
		&d.CreatedAt, &d.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &d, err
}

func scanManagedDatabaseRows(rows pgx.Rows) ([]model.ManagedDatabase, error) {
	var dbs []model.ManagedDatabase
	for rows.Next() {
		var d model.ManagedDatabase
		if err := rows.Scan(
			&d.ID, &d.UserID, &d.AppID, &d.Name, &d.Engine, &d.Version, &d.Status,
			&d.Host, &d.Port, &d.DBName, &d.Username, &d.Password, &d.ConnectionURL,
			&d.ContainerID, &d.MemoryLimit, &d.StorageUsed, &d.StorageLimit,
			&d.AutoBackup, &d.BackupRetention, &d.BackupSchedule,
			&d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, err
		}
		dbs = append(dbs, d)
	}
	return dbs, rows.Err()
}

func (r *databaseRepo) Create(ctx context.Context, mdb *model.ManagedDatabase) error {
	mdb.ID = uuid.New()
	now := time.Now().UTC()
	mdb.CreatedAt = now
	mdb.UpdatedAt = now

	// Set defaults for new fields if not already set.
	if mdb.BackupRetention == 0 {
		mdb.BackupRetention = 7
	}
	if mdb.BackupSchedule == "" {
		mdb.BackupSchedule = "0 2 * * *"
	}

	_, err := r.db.Exec(ctx,
		`INSERT INTO managed_databases
		 (id, user_id, app_id, name, engine, version, status, host, port, db_name,
		  username, password, connection_url, container_id, memory_limit, storage_used, storage_limit,
		  auto_backup, backup_retention, backup_schedule,
		  created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22)`,
		mdb.ID, mdb.UserID, mdb.AppID, mdb.Name, mdb.Engine, mdb.Version, mdb.Status,
		mdb.Host, mdb.Port, mdb.DBName, mdb.Username, mdb.Password, mdb.ConnectionURL,
		mdb.ContainerID, mdb.MemoryLimit, mdb.StorageUsed, mdb.StorageLimit,
		mdb.AutoBackup, mdb.BackupRetention, mdb.BackupSchedule,
		mdb.CreatedAt, mdb.UpdatedAt,
	)
	return err
}

func (r *databaseRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.ManagedDatabase, error) {
	row := r.db.QueryRow(ctx,
		`SELECT `+dbSelectColumns+` FROM managed_databases WHERE id = $1`, id,
	)
	return scanManagedDatabase(row)
}

func (r *databaseRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]model.ManagedDatabase, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+dbSelectColumns+`
		 FROM managed_databases WHERE user_id = $1 AND status != 'deleted'
		 ORDER BY created_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanManagedDatabaseRows(rows)
}

func (r *databaseRepo) Update(ctx context.Context, mdb *model.ManagedDatabase) error {
	mdb.UpdatedAt = time.Now().UTC()
	_, err := r.db.Exec(ctx,
		`UPDATE managed_databases SET
		 app_id=$1, name=$2, status=$3, host=$4, port=$5, db_name=$6,
		 username=$7, password=$8, connection_url=$9, container_id=$10,
		 memory_limit=$11, storage_used=$12, storage_limit=$13,
		 auto_backup=$14, backup_retention=$15, backup_schedule=$16,
		 updated_at=$17
		 WHERE id=$18`,
		mdb.AppID, mdb.Name, mdb.Status, mdb.Host, mdb.Port, mdb.DBName,
		mdb.Username, mdb.Password, mdb.ConnectionURL, mdb.ContainerID,
		mdb.MemoryLimit, mdb.StorageUsed, mdb.StorageLimit,
		mdb.AutoBackup, mdb.BackupRetention, mdb.BackupSchedule,
		mdb.UpdatedAt,
		mdb.ID,
	)
	return err
}

func (r *databaseRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM managed_databases WHERE id = $1`, id)
	return err
}

func (r *databaseRepo) GetByAppID(ctx context.Context, appID uuid.UUID) ([]model.ManagedDatabase, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+dbSelectColumns+`
		 FROM managed_databases WHERE app_id = $1 AND status != 'deleted'
		 ORDER BY created_at DESC`, appID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanManagedDatabaseRows(rows)
}

func (r *databaseRepo) CreateBackup(ctx context.Context, backup *model.DatabaseBackup) error {
	backup.ID = uuid.New()
	backup.CreatedAt = time.Now().UTC()

	_, err := r.db.Exec(ctx,
		`INSERT INTO database_backups (id, database_id, file_path, file_size, status, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		backup.ID, backup.DatabaseID, backup.FilePath, backup.FileSize, backup.Status, backup.CreatedAt,
	)
	return err
}

func (r *databaseRepo) ListBackups(ctx context.Context, databaseID uuid.UUID) ([]model.DatabaseBackup, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, database_id, file_path, file_size, status, created_at
		 FROM database_backups WHERE database_id = $1 ORDER BY created_at DESC`, databaseID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var backups []model.DatabaseBackup
	for rows.Next() {
		var b model.DatabaseBackup
		if err := rows.Scan(&b.ID, &b.DatabaseID, &b.FilePath, &b.FileSize, &b.Status, &b.CreatedAt); err != nil {
			return nil, err
		}
		backups = append(backups, b)
	}
	return backups, rows.Err()
}

func (r *databaseRepo) UpdateBackupStatus(ctx context.Context, backupID uuid.UUID, status string, fileSize int64) error {
	_, err := r.db.Exec(ctx,
		`UPDATE database_backups SET status = $1, file_size = $2 WHERE id = $3`,
		status, fileSize, backupID,
	)
	return err
}

func (r *databaseRepo) DeleteBackup(ctx context.Context, backupID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM database_backups WHERE id = $1`, backupID,
	)
	return err
}

func (r *databaseRepo) ListAutoBackupDatabases(ctx context.Context) ([]model.ManagedDatabase, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+dbSelectColumns+`
		 FROM managed_databases WHERE auto_backup = true AND status = 'running'
		 ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanManagedDatabaseRows(rows)
}
