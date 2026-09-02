package scan

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"ladon/cluster"
	"ladon/rule"

	"k8s.io/client-go/kubernetes"
)

type Service struct {
	db      *sql.DB
	session *cluster.Session
	rules   *rule.Service
}

func New(db *sql.DB, session *cluster.Session, rules *rule.Service) *Service {
	return &Service{db: db, session: session, rules: rules}
}

func (s *Service) PreviewScan() ([]Finding, error) {
	client, _, err := s.session.Client()
	if err != nil {
		return nil, err
	}
	return s.evaluate(client)
}

func (s *Service) RunScan(projectID string) (ScanResult, error) {
	ok, err := s.projectExists(projectID)
	if err != nil {
		return ScanResult{}, err
	}
	if !ok {
		return ScanResult{}, fmt.Errorf("project '%s' not found", projectID)
	}
	client, contextName, err := s.session.Client()
	if err != nil {
		return ScanResult{}, err
	}
	findings, evalErr := s.evaluate(client)
	return s.persist(projectID, nonemptyPtr(contextName), findings, evalErr)
}

func (s *Service) ListScans(projectID string) ([]Scan, error) {
	return s.listScans(projectID)
}

func (s *Service) GetScan(id string) (*Scan, error) {
	return s.getScan(id)
}

func (s *Service) ListScanFindings(scanID string) ([]Finding, error) {
	return s.listFindings(scanID)
}

func (s *Service) projectExists(id string) (bool, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM projects WHERE id = ?`, id).Scan(&n)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (s *Service) evaluate(client kubernetes.Interface) ([]Finding, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	hardcoded, err := runHardcoded(ctx, client)
	if err != nil {
		return nil, err
	}
	stored, err := s.rules.ListRules()
	if err != nil {
		return nil, err
	}
	out := append([]Finding{}, hardcoded...)
	out = append(out, evaluateStored(ctx, client, stored)...)
	return out, nil
}

func nonemptyPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
