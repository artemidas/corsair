package trivy

import (
	"ladon/rule"
)

type Finding struct {
	RuleID       string
	RuleTitle    string
	Severity     rule.Severity
	ResourceKind string
	ResourceName string
	Message      string
}
