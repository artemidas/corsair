package trivy

import (
	"ladon/backend/rule"
)

type Finding struct {
	RuleID       string
	RuleTitle    string
	Severity     rule.Severity
	ResourceKind string
	ResourceName string
	Message      string
}
