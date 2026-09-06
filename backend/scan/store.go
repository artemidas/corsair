package scan

import (
	"database/sql"
	"time"

	"github.com/google/uuid"

	"ladon/backend/rule"
)

func (s *Service) persist(projectID string, contextName *string, findings []Finding, evalErr error) (ScanResult, error) {
	id := uuid.NewString()
	now := time.Now().UTC().Format(time.RFC3339)
	if evalErr != nil {
		msg := evalErr.Error()
		_, err := s.db.Exec(
			`INSERT INTO scans
			 (id, project_id, status, context, error, finding_count, started_at, finished_at)
			 VALUES (?, ?, ?, ?, ?, 0, ?, ?)`,
			id, projectID, string(StatusFailed), contextName, msg, now, now,
		)
		if err != nil {
			return ScanResult{}, err
		}
		return ScanResult{
			Scan: Scan{
				ID:           id,
				ProjectID:    projectID,
				Status:       StatusFailed,
				Context:      contextName,
				Error:        &msg,
				FindingCount: 0,
				StartedAt:    now,
				FinishedAt:   now,
			},
			Findings: []Finding{},
		}, nil
	}

	stored := make([]Finding, 0, len(findings))
	for _, f := range findings {
		f.ID = uuid.NewString()
		f.ScanID = id
		stored = append(stored, f)
	}

	tx, err := s.db.Begin()
	if err != nil {
		return ScanResult{}, err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(
		`INSERT INTO scans
		 (id, project_id, status, context, error, finding_count, started_at, finished_at)
		 VALUES (?, ?, ?, ?, NULL, ?, ?, ?)`,
		id, projectID, string(StatusCompleted), contextName, int64(len(stored)), now, now,
	); err != nil {
		return ScanResult{}, err
	}
	for _, f := range stored {
		if _, err := tx.Exec(
			`INSERT INTO findings
			 (id, scan_id, rule_id, rule_title, severity, resource_kind, resource_name, namespace, message)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			f.ID, f.ScanID, f.RuleID, f.RuleTitle, string(f.Severity),
			f.ResourceKind, f.ResourceName, f.Namespace, f.Message,
		); err != nil {
			return ScanResult{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return ScanResult{}, err
	}
	return ScanResult{
		Scan: Scan{
			ID:           id,
			ProjectID:    projectID,
			Status:       StatusCompleted,
			Context:      contextName,
			FindingCount: int64(len(stored)),
			StartedAt:    now,
			FinishedAt:   now,
		},
		Findings: stored,
	}, nil
}

func (s *Service) listScans(projectID string) ([]Scan, error) {
	rows, err := s.db.Query(
		`SELECT id, project_id, status, context, error, finding_count, started_at, finished_at
		 FROM scans WHERE project_id = ? ORDER BY started_at DESC`,
		projectID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Scan{}
	for rows.Next() {
		sc, err := scanScan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, sc)
	}
	return out, rows.Err()
}

func (s *Service) getScan(id string) (*Scan, error) {
	row := s.db.QueryRow(
		`SELECT id, project_id, status, context, error, finding_count, started_at, finished_at
		 FROM scans WHERE id = ?`,
		id,
	)
	sc, err := scanScan(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &sc, nil
}

func (s *Service) deleteScan(id string) error {
	_, err := s.db.Exec(
		`DELETE FROM scans WHERE id = ?`,
		id,
	)
	return err
}

func (s *Service) listFindings(scanID string) ([]Finding, error) {
	rows, err := s.db.Query(
		`SELECT id, scan_id, rule_id, rule_title, severity, resource_kind, resource_name, namespace, message
		 FROM findings WHERE scan_id = ? ORDER BY rule_id, resource_name`,
		scanID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Finding{}
	for rows.Next() {
		f, err := scanFinding(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanScan(row rowScanner) (Scan, error) {
	var (
		id, projectID, status, startedAt, finishedAt string
		contextName, errMsg                          sql.NullString
		findingCount                                 int64
	)
	if err := row.Scan(&id, &projectID, &status, &contextName, &errMsg, &findingCount, &startedAt, &finishedAt); err != nil {
		return Scan{}, err
	}
	st, err := parseStatus(status)
	if err != nil {
		return Scan{}, err
	}
	sc := Scan{
		ID:           id,
		ProjectID:    projectID,
		Status:       st,
		FindingCount: findingCount,
		StartedAt:    startedAt,
		FinishedAt:   finishedAt,
	}
	if contextName.Valid {
		sc.Context = &contextName.String
	}
	if errMsg.Valid {
		sc.Error = &errMsg.String
	}
	return sc, nil
}

func scanFinding(row rowScanner) (Finding, error) {
	var (
		id, scanID, ruleID, title, severity, kind, name, message string
		namespace                                                sql.NullString
	)
	if err := row.Scan(&id, &scanID, &ruleID, &title, &severity, &kind, &name, &namespace, &message); err != nil {
		return Finding{}, err
	}
	sev, err := rule.ParseSeverity(severity)
	if err != nil {
		return Finding{}, err
	}
	f := Finding{
		ID:           id,
		ScanID:       scanID,
		RuleID:       ruleID,
		RuleTitle:    title,
		Severity:     sev,
		ResourceKind: kind,
		ResourceName: name,
		Message:      message,
	}
	if namespace.Valid {
		f.Namespace = &namespace.String
	}
	return f, nil
}
