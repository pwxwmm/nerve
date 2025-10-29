// Package storage provides storage configuration and connection management.
//
// Author: mmwei3 (2025-10-28)
// Wethers: cloudWays
package storage

import (
	"time"
)

// Config holds storage configuration
type Config struct {
	Type     string              `yaml:"type"`
	MongoDB  *MongoDBConfig      `yaml:"mongodb,omitempty"`
	Redis    *RedisConfig        `yaml:"redis,omitempty"`
	Postgres *PostgresConfig     `yaml:"postgres,omitempty"`
}

// MongoDBConfig contains MongoDB connection configuration
type MongoDBConfig struct {
	URI      string        `yaml:"uri"`
	Database string        `yaml:"database"`
	Timeout  time.Duration `yaml:"timeout,omitempty"`
}

// RedisConfig contains Redis connection configuration
type RedisConfig struct {
	Host     string        `yaml:"host"`
	Port     int           `yaml:"port"`
	Password string        `yaml:"password"`
	Database int           `yaml:"database"`
	Timeout  time.Duration `yaml:"timeout,omitempty"`
}

// PostgresConfig contains PostgreSQL connection configuration
type PostgresConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	SSLMode  string `yaml:"sslmode"`
}

// NewFromConfig creates a storage instance from configuration
func NewFromConfig(cfg Config) (Storage, error) {
	switch cfg.Type {
	case "mongodb":
		if cfg.MongoDB == nil {
			return nil, ErrNotFound
		}
		return NewMongoDB(*cfg.MongoDB)
	case "postgres":
		if cfg.Postgres == nil {
			return nil, ErrNotFound
		}
		return NewPostgres(*cfg.Postgres)
	case "redis":
		if cfg.Redis == nil {
			return nil, ErrNotFound
		}
		return NewRedis(*cfg.Redis)
	case "memory", "":
		return NewInMemory(), nil
	default:
		return NewInMemory(), nil
	}
}

