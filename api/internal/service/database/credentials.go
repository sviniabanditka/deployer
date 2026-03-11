package database

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/url"
	"strings"

	"github.com/deployer/api/internal/model"
)

const passwordChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// generatePassword creates a cryptographically random password of the given length.
func generatePassword(length int) string {
	b := make([]byte, length)
	for i := range b {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(passwordChars))))
		b[i] = passwordChars[idx.Int64()]
	}
	return string(b)
}

// generateUsername returns a default username for the engine.
func generateUsername(engine model.DatabaseEngine) string {
	switch engine {
	case model.EnginePostgres:
		return "pguser"
	case model.EngineMySQL:
		return "mysqluser"
	case model.EngineMongoDB:
		return "mongouser"
	case model.EngineRedis:
		return ""
	default:
		return "dbuser"
	}
}

// generateDBName returns a default database name based on the user-supplied name.
func generateDBName(name string) string {
	clean := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		if r >= 'A' && r <= 'Z' {
			return r + 32 // lowercase
		}
		return '_'
	}, name)
	if clean == "" {
		clean = "appdb"
	}
	return clean
}

// buildConnectionURL constructs a connection URL for the given engine.
func buildConnectionURL(engine model.DatabaseEngine, host string, port int, username, password, dbName string) string {
	switch engine {
	case model.EnginePostgres:
		return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable",
			url.PathEscape(username), url.PathEscape(password), host, port, dbName)
	case model.EngineMySQL:
		return fmt.Sprintf("mysql://%s:%s@%s:%d/%s",
			url.PathEscape(username), url.PathEscape(password), host, port, dbName)
	case model.EngineMongoDB:
		return fmt.Sprintf("mongodb://%s:%s@%s:%d/%s?authSource=admin",
			url.PathEscape(username), url.PathEscape(password), host, port, dbName)
	case model.EngineRedis:
		return fmt.Sprintf("redis://:%s@%s:%d/0",
			url.PathEscape(password), host, port)
	default:
		return ""
	}
}
