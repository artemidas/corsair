package scan

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"ladon/backend/cluster"
	"ladon/backend/project"
	"ladon/backend/rule"
	"ladon/backend/trivy"

	"k8s.io/client-go/kubernetes"
)

type Service struct {
	db       *sql.DB
	session  *cluster.Session
	rules    *rule.Service
	projects *project.Service
	trivy    *trivy.Scanner
}

func New(db *sql.DB, session *cluster.Session, rules *rule.Service, projects *project.Service, scanner *trivy.Scanner) *Service {
	return &Service{db: db, session: session, rules: rules, projects: projects, trivy: scanner}
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

func (s *Service) RunImageScan(projectID string, opts trivy.ScanOptions) (ScanResult, error) {
	proj, err := s.projects.GetProject(projectID)
	if err != nil {
		return ScanResult{}, err
	}
	if proj == nil {
		return ScanResult{}, fmt.Errorf("project '%s' not found", projectID)
	}
	if proj.Kind != project.KindContainerImageReview {
		return ScanResult{}, fmt.Errorf("project '%s' is not a container image review", projectID)
	}
	images := project.CollectImages(proj.Config)
	raw, evalErr := s.trivy.ScanImages(images, opts)
	return s.persist(projectID, nil, toScanFindings(raw), evalErr)
}

func toScanFindings(raw []trivy.Finding) []Finding {
	out := make([]Finding, 0, len(raw))
	for _, f := range raw {
		out = append(out, Finding{
			RuleID:       f.RuleID,
			RuleTitle:    f.RuleTitle,
			Severity:     f.Severity,
			ResourceKind: f.ResourceKind,
			ResourceName: f.ResourceName,
			Message:      f.Message,
		})
	}
	return out
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

func (s *Service) DeleteScan(id string) error {
	return s.deleteScan(id)
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
