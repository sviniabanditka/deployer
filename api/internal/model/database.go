package model

import (
	"fmt"
	"net/url"
	"time"

	"github.com/google/uuid"
)

type DatabaseEngine string

const (
	EnginePostgres DatabaseEngine = "postgres"
	EngineMySQL    DatabaseEngine = "mysql"
	EngineMongoDB  DatabaseEngine = "mongodb"
	EngineRedis    DatabaseEngine = "redis"
)

// DefaultVersions maps each engine to its default version.
var DefaultVersions = map[DatabaseEngine]string{
	EnginePostgres: "16",
	EngineMySQL:    "8",
	EngineMongoDB:  "7",
	EngineRedis:    "7",
}

type DatabaseStatus string

const (
	DatabaseStatusCreating DatabaseStatus = "creating"
	DatabaseStatusRunning  DatabaseStatus = "running"
	DatabaseStatusStopped  DatabaseStatus = "stopped"
	DatabaseStatusFailed   DatabaseStatus = "failed"
	DatabaseStatusDeleted  DatabaseStatus = "deleted"
)

type ManagedDatabase struct {
	ID              uuid.UUID      `json:"id"`
	UserID          uuid.UUID      `json:"user_id"`
	AppID           *uuid.UUID     `json:"app_id,omitempty"`
	Name            string         `json:"name"`
	Engine          DatabaseEngine `json:"engine"`
	Version         string         `json:"version"`
	Status          DatabaseStatus `json:"status"`
	Host            string         `json:"host"`
	Port            int            `json:"port"`
	DBName          string         `json:"db_name"`
	Username        string         `json:"username"`
	Password        string         `json:"-"`
	ConnectionURL   string         `json:"-"`
	ContainerID     string         `json:"container_id,omitempty"`
	MemoryLimit     int64          `json:"memory_limit"`
	StorageUsed     int64          `json:"storage_used"`
	StorageLimit    int64          `json:"storage_limit"`
	AutoBackup      bool           `json:"auto_backup"`
	BackupRetention int            `json:"backup_retention"`
	BackupSchedule  string         `json:"backup_schedule"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

// SafeConnectionURL returns the connection URL with the password masked.
func (d *ManagedDatabase) SafeConnectionURL() string {
	if d.ConnectionURL == "" {
		return ""
	}
	parsed, err := url.Parse(d.ConnectionURL)
	if err != nil {
		return ""
	}
	if parsed.User != nil {
		parsed.User = url.UserPassword(parsed.User.Username(), "****")
	}
	return parsed.String()
}

// FullConnectionInfo returns a copy with sensitive fields included in JSON.
type DatabaseConnectionInfo struct {
	ID            uuid.UUID      `json:"id"`
	Name          string         `json:"name"`
	Engine        DatabaseEngine `json:"engine"`
	Version       string         `json:"version"`
	Status        DatabaseStatus `json:"status"`
	Host          string         `json:"host"`
	Port          int            `json:"port"`
	DBName        string         `json:"db_name"`
	Username      string         `json:"username"`
	Password      string         `json:"password"`
	ConnectionURL string         `json:"connection_url"`
}

// ConnectionInfo returns the full connection details including secrets.
func (d *ManagedDatabase) ConnectionInfo() *DatabaseConnectionInfo {
	return &DatabaseConnectionInfo{
		ID:            d.ID,
		Name:          d.Name,
		Engine:        d.Engine,
		Version:       d.Version,
		Status:        d.Status,
		Host:          d.Host,
		Port:          d.Port,
		DBName:        d.DBName,
		Username:      d.Username,
		Password:      d.Password,
		ConnectionURL: d.ConnectionURL,
	}
}

// ValidEngine returns true if the engine is supported.
func ValidEngine(engine string) bool {
	switch DatabaseEngine(engine) {
	case EnginePostgres, EngineMySQL, EngineMongoDB, EngineRedis:
		return true
	}
	return false
}

// DockerImage returns the Docker image reference for the engine and version.
func DockerImage(engine DatabaseEngine, version string) string {
	switch engine {
	case EnginePostgres:
		return fmt.Sprintf("postgres:%s-alpine", version)
	case EngineMySQL:
		return fmt.Sprintf("mysql:%s", version)
	case EngineMongoDB:
		return fmt.Sprintf("mongo:%s", version)
	case EngineRedis:
		return fmt.Sprintf("redis:%s-alpine", version)
	default:
		return ""
	}
}

// DefaultPort returns the default port for the engine.
func DefaultPort(engine DatabaseEngine) int {
	switch engine {
	case EnginePostgres:
		return 5432
	case EngineMySQL:
		return 3306
	case EngineMongoDB:
		return 27017
	case EngineRedis:
		return 6379
	default:
		return 0
	}
}

type DatabaseBackup struct {
	ID         uuid.UUID `json:"id"`
	DatabaseID uuid.UUID `json:"database_id"`
	FilePath   string    `json:"file_path"`
	FileSize   int64     `json:"file_size"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}
