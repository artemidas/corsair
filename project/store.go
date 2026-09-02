package project

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const schemaV1 = `CREATE TABLE IF NOT EXISTS projects (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	kind TEXT NOT NULL,
	config TEXT NOT NULL,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
)`

func migrate(db *sql.DB) error {
	var version int
	if err := db.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		return fmt.Errorf("read schema version: %w", err)
	}
	if version < 1 {
		if _, err := db.Exec(schemaV1); err != nil {
			return fmt.Errorf("create projects table: %w", err)
		}
		if _, err := db.Exec("PRAGMA user_version = 1"); err != nil {
			return fmt.Errorf("set schema version: %w", err)
		}
	}
	return nil
}

func (s *Service) ListProjects() ([]Project, error) {
	rows, err := s.db.Query(
		`SELECT id, name, kind, config, created_at, updated_at
		 FROM projects ORDER BY created_at ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []Project{}
	for rows.Next() {
		project, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, project)
	}
	return out, rows.Err()
}

func (s *Service) GetProject(id string) (*Project, error) {
	row := s.db.QueryRow(
		`SELECT id, name, kind, config, created_at, updated_at
		 FROM projects WHERE id = ?`,
		id,
	)
	project, err := scanProject(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (s *Service) CreateProject(input ProjectInput) (Project, error) {
	name, config, err := validateInput(input)
	if err != nil {
		return Project{}, err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	project := Project{
		ID:        uuid.NewString(),
		Name:      name,
		Kind:      input.Kind,
		Config:    config,
		CreatedAt: now,
		UpdatedAt: now,
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		return Project{}, err
	}
	_, err = s.db.Exec(
		`INSERT INTO projects (id, name, kind, config, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		project.ID, project.Name, string(project.Kind), string(configJSON),
		project.CreatedAt, project.UpdatedAt,
	)
	if err != nil {
		return Project{}, err
	}
	return project, nil
}

func (s *Service) UpdateProject(id string, input ProjectInput) (Project, error) {
	name, config, err := validateInput(input)
	if err != nil {
		return Project{}, err
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		return Project{}, err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.Exec(
		`UPDATE projects SET name = ?, kind = ?, config = ?, updated_at = ?
		 WHERE id = ?`,
		name, string(input.Kind), string(configJSON), now, id,
	)
	if err != nil {
		return Project{}, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return Project{}, err
	}
	if n == 0 {
		return Project{}, fmt.Errorf("project '%s' not found", id)
	}
	updated, err := s.GetProject(id)
	if err != nil {
		return Project{}, err
	}
	if updated == nil {
		return Project{}, fmt.Errorf("project '%s' disappeared after update", id)
	}
	return *updated, nil
}

func (s *Service) DeleteProject(id string) error {
	if _, err := s.db.Exec(
		`DELETE FROM findings WHERE scan_id IN (SELECT id FROM scans WHERE project_id = ?)`, id,
	); err != nil {
		return err
	}
	if _, err := s.db.Exec(`DELETE FROM scans WHERE project_id = ?`, id); err != nil {
		return err
	}
	res, err := s.db.Exec(`DELETE FROM projects WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("project '%s' not found", id)
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanProject(row rowScanner) (Project, error) {
	var (
		id, name, kind, configJSON, createdAt, updatedAt string
	)
	if err := row.Scan(&id, &name, &kind, &configJSON, &createdAt, &updatedAt); err != nil {
		return Project{}, err
	}
	parsedKind, err := ParseKind(kind)
	if err != nil {
		return Project{}, err
	}
	var config ProjectConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return Project{}, fmt.Errorf("invalid config in db: %w", err)
	}
	config.Images = CollectImages(config)
	config.Image = nil
	return Project{
		ID:        id,
		Name:      name,
		Kind:      parsedKind,
		Config:    config,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}
