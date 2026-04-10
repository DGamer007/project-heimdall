package infra

import (
	"database/sql"

	"github.com/redis/go-redis/v9"
)

type DataStore struct {
	Postgres *sql.DB
	Redis *redis.Client
}
