package scan

import (
	"fmt"

	"ladon/backend/rule"
)

type ScanStatus string

const (
	StatusCompleted ScanStatus = "completed"
	StatusFailed    ScanStatus = "failed"
)

type Scan struct {
	ID           string     `json:"id"`
	ProjectID    string     `json:"projectId"`
	Status       ScanStatus `json:"status"`
	Context      *string    `json:"context"`
	Error        *string    `json:"error"`
	FindingCount int64      `json:"findingCount"`
	StartedAt    string     `json:"startedAt"`
	FinishedAt   string     `json:"finishedAt"`
}

type Finding struct {
	ID           string        `json:"id"`
	ScanID       string        `json:"scanId,omitempty"`
	RuleID       string        `json:"ruleId"`
	RuleTitle    string        `json:"ruleTitle"`
	Severity     rule.Severity `json:"severity"`
	ResourceKind string        `json:"resourceKind"`
	ResourceName string        `json:"resourceName"`
	Namespace    *string       `json:"namespace"`
	Message      string        `json:"message"`
}

type ScanResult struct {
	Scan     Scan      `json:"scan"`
	Findings []Finding `json:"findings"`
}

func parseStatus(s string) (ScanStatus, error) {
	switch ScanStatus(s) {
	case StatusCompleted, StatusFailed:
		return ScanStatus(s), nil
	default:
		return "", fmt.Errorf("unknown scan status '%s' in db", s)
	}
}
