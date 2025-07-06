package postgres

import (
	"database/sql"
	"fmt"
	"time"

	internal_errors "heimdall/backend/pkg/errors"

	_ "github.com/lib/pq"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func NewConnection(config Config) (*sql.DB, error) {
	connectionURI := config.getConnectionURI()

	conn, err := sql.Open("postgres", connectionURI)
	if err != nil {
		return nil, internal_errors.NewInfrastructureError("Failed to open Postgres connection", err)
	}

	// Test the connection
	if err = conn.Ping(); err != nil {
		conn.Close() // Clean up the failed connection
		return nil, internal_errors.NewInfrastructureError("Failed to ping Postgres", err)
	}

	conn.SetMaxOpenConns(50)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxIdleTime(1 * time.Minute)
	conn.SetConnMaxLifetime(5 * time.Minute)

	return conn, nil
}

func (c *Config) getConnectionURI() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host,
		c.Port,
		c.User,
		c.Password,
		c.DBName,
		c.SSLMode,
	)
}
