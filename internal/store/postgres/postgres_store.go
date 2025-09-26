package postgres

import "database/sql"

type PostgresStore struct{ db *sql.DB }

func NewPostgresStore(dsn string) (*PostgresStore, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) Init() error { return s.db.Ping() }
