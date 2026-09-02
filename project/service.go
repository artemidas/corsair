package project

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type Service struct {
	db *sql.DB
}

func DefaultDBPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir = filepath.Join(dir, "ladon")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "ladon.db"), nil
}

func Open(dbPath string) (*Service, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}
	if err := migrate(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Service{db: db}, nil
}

func (s *Service) close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

// Close releases the database. Not a bound method (unexported receiver helper).
func Close(s *Service) error {
	return s.close()
}
