package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// PostgresStorage implements Storage using PostgreSQL
type PostgresStorage struct {
	db *sql.DB
}

// NewPostgres creates a new PostgreSQL storage instance
func NewPostgres(cfg PostgresConfig) (*PostgresStorage, error) {
	dsn := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Database, cfg.User, cfg.Password, cfg.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	storage := &PostgresStorage{db: db}
	
	// Create tables
	if err := storage.createTables(); err != nil {
		return nil, err
	}

	return storage, nil
}

// createTables creates necessary tables for Nerve data
func (p *PostgresStorage) createTables() error {
	query := `
	-- Agents table
	CREATE TABLE IF NOT EXISTS agents (
		id SERIAL PRIMARY KEY,
		hostname VARCHAR(255) UNIQUE NOT NULL,
		system_info JSONB NOT NULL,
		cluster_id INTEGER,
		status VARCHAR(50),
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW(),
		last_seen TIMESTAMP
	);

	-- Create indexes
	CREATE INDEX IF NOT EXISTS idx_agents_hostname ON agents(hostname);
	CREATE INDEX IF NOT EXISTS idx_agents_status ON agents(status);
	CREATE INDEX IF NOT EXISTS idx_agents_last_seen ON agents(last_seen);
	CREATE INDEX IF NOT EXISTS idx_agents_system_info ON agents USING GIN (system_info);
	CREATE INDEX IF NOT EXISTS idx_agents_cluster ON agents(cluster_id);

	-- Heartbeats table with time partitioning
	CREATE TABLE IF NOT EXISTS heartbeats (
		id SERIAL PRIMARY KEY,
		agent_id INTEGER REFERENCES agents(id),
		timestamp TIMESTAMP DEFAULT NOW(),
		metrics JSONB,
		UNIQUE(agent_id, timestamp)
	);

	CREATE INDEX IF NOT EXISTS idx_heartbeats_agent_timestamp ON heartbeats(agent_id, timestamp DESC);

	-- Tasks table
	CREATE TABLE IF NOT EXISTS tasks (
		id SERIAL PRIMARY KEY,
		task_id VARCHAR(255) UNIQUE NOT NULL,
		agent_id INTEGER REFERENCES agents(id),
		action VARCHAR(255),
		params JSONB,
		status VARCHAR(50),
		result JSONB,
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_tasks_agent_status ON tasks(agent_id, status);
	CREATE INDEX IF NOT EXISTS idx_tasks_created ON tasks(created_at DESC);

	-- Retention policy (cleanup old data)
	CREATE OR REPLACE FUNCTION cleanup_old_heartbeats()
	RETURNS void AS $$
	BEGIN
		DELETE FROM heartbeats WHERE timestamp < NOW() - INTERVAL '7 days';
	END;
	$$ LANGUAGE plpgsql;
	`

	_, err := p.db.Exec(query)
	return err
}

// Get retrieves a value from storage
func (p *PostgresStorage) Get(key string) (interface{}, error) {
	var value string
	err := p.db.QueryRow("SELECT value FROM storage WHERE key = $1", key).Scan(&value)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	
	var result interface{}
	if err := json.Unmarshal([]byte(value), &result); err != nil {
		return nil, err
	}
	
	return result, err
}

// Set stores a value in storage
func (p *PostgresStorage) Set(key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO storage (key, value, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (key)
		DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()
	`
	_, err = p.db.Exec(query, key, string(data))
	return err
}

// Delete removes a value from storage
func (p *PostgresStorage) Delete(key string) error {
	_, err := p.db.Exec("DELETE FROM storage WHERE key = $1", key)
	return err
}

// List returns all key-value pairs
func (p *PostgresStorage) List() map[string]interface{} {
	rows, err := p.db.Query("SELECT key, value FROM storage")
	if err != nil {
		return make(map[string]interface{})
	}
	defer rows.Close()

	result := make(map[string]interface{})
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			continue
		}
		var data interface{}
		if err := json.Unmarshal([]byte(value), &data); err != nil {
			continue
		}
		result[key] = data
	}
	
	return result
}

// SaveAgent saves agent information
func (p *PostgresStorage) SaveAgent(agent interface{}) error {
	data, err := json.Marshal(agent)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO agents (hostname, system_info, updated_at, last_seen)
		VALUES (
			(SELECT hostname FROM jsonb_to_record($1::jsonb) AS x(hostname TEXT)),
			$1::jsonb,
			NOW(),
			NOW()
		)
		ON CONFLICT (hostname)
		DO UPDATE SET
			system_info = EXCLUDED.system_info,
			updated_at = NOW(),
			last_seen = NOW()
	`
	_, err = p.db.Exec(query, string(data))
	return err
}

// SaveHeartbeat saves heartbeat data
func (p *PostgresStorage) SaveHeartbeat(agentID string, heartbeat interface{}) error {
	data, err := json.Marshal(heartbeat)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO heartbeats (agent_id, timestamp, metrics)
		VALUES (
			(SELECT id FROM agents WHERE hostname = $1),
			NOW(),
			$2::jsonb
		)
		ON CONFLICT (agent_id, timestamp)
		DO UPDATE SET metrics = EXCLUDED.metrics
	`
	_, err = p.db.Exec(query, agentID, string(data))
	return err
}

// GetAgents retrieves all agents
func (p *PostgresStorage) GetAgents(filter map[string]interface{}) ([]interface{}, error) {
	query := "SELECT * FROM agents"
	args := []interface{}{}
	argIndex := 1

	if len(filter) > 0 {
		query += " WHERE "
		conditions := []string{}
		for key, value := range filter {
			conditions = append(conditions, fmt.Sprintf("%s = $%d", key, argIndex))
			args = append(args, value)
			argIndex++
		}
		query += fmt.Sprintf("%s", conditions[0])
	}

	rows, err := p.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []interface{}
	for rows.Next() {
		var id int
		var hostname, status string
		var systemInfo []byte
		var clusterID sql.NullInt64
		var createdAt, updatedAt, lastSeen sql.NullTime

		if err := rows.Scan(&id, &hostname, &systemInfo, &clusterID, &status, &createdAt, &updatedAt, &lastSeen); err != nil {
			continue
		}

		var info interface{}
		if err := json.Unmarshal(systemInfo, &info); err != nil {
			continue
		}

		result := map[string]interface{}{
			"id":       id,
			"hostname": hostname,
			"system_info": info,
			"status":   status,
		}
		results = append(results, result)
	}

	return results, nil
}

// Close closes the PostgreSQL connection
func (p *PostgresStorage) Close() error {
	return p.db.Close()
}

// RunCleanup runs cleanup tasks (e.g., old heartbeats)
func (p *PostgresStorage) RunCleanup() error {
	ctx := context.Background()
	_, err := p.db.ExecContext(ctx, "SELECT cleanup_old_heartbeats()")
	return err
}
