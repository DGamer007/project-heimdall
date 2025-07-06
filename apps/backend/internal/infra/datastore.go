package infra

import "database/sql"

type DataStore struct {
	Postgres *sql.DB
}
