package rule

import (
	"fmt"
	"strings"
)

type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
)

type Operator string

const (
	OpEquals        Operator = "equals"
	OpNotEquals     Operator = "notEquals"
	OpPresent       Operator = "present"
	OpAbsent        Operator = "absent"
	OpArrayExcludes Operator = "arrayExcludes"
)

type Rule struct {
	ID           string   `json:"id"`
	RuleID       string   `json:"ruleId"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Severity     Severity `json:"severity"`
	ResourceType string   `json:"resourceType"`
	Rego         string   `json:"rego"`
	ImportID     *string  `json:"importId"`
	CreatedAt    string   `json:"createdAt"`
	UpdatedAt    string   `json:"updatedAt"`
}

type RuleInput struct {
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Severity     Severity `json:"severity"`
	ResourceType string   `json:"resourceType"`
	Rego         string   `json:"rego"`
}

type preparedInput struct {
	title        string
	description  string
	severity     Severity
	resourceType string
	rego         string
}

func ParseSeverity(s string) (Severity, error) {
	switch Severity(s) {
	case SeverityCritical, SeverityHigh, SeverityMedium, SeverityLow:
		return Severity(s), nil
	default:
		return "", fmt.Errorf("unknown severity '%s' in db", s)
	}
}

func ParseOperator(s string) (Operator, error) {
	switch Operator(s) {
	case OpEquals, OpNotEquals, OpPresent, OpAbsent, OpArrayExcludes:
		return Operator(s), nil
	default:
		return "", fmt.Errorf("unknown operator '%s' in db", s)
	}
}

func prepareInput(input RuleInput) (preparedInput, error) {
	title := strings.TrimSpace(input.Title)
	resourceType := strings.TrimSpace(input.ResourceType)
	if title == "" {
		return preparedInput{}, fmt.Errorf("title must not be empty")
	}
	if resourceType == "" {
		return preparedInput{}, fmt.Errorf("resource_type must not be empty")
	}
	if _, err := ParseSeverity(string(input.Severity)); err != nil {
		return preparedInput{}, err
	}
	rego, err := Validate(input.Rego)
	if err != nil {
		return preparedInput{}, err
	}
	return preparedInput{
		title:        title,
		description:  strings.TrimSpace(input.Description),
		severity:     input.Severity,
		resourceType: resourceType,
		rego:         rego,
	}, nil
}

func ruleIDPrefix(resourceType string) string {
	switch resourceType {
	case "Pod":
		return "POD"
	case "ServiceAccount":
		return "SA"
	case "Role":
		return "ROLE"
	case "ClusterRole":
		return "CR"
	case "RoleBinding":
		return "RB"
	case "ClusterRoleBinding":
		return "CRB"
	default:
		return "RULE"
	}
}

func parseRuleIDSeq(ruleID, prefix string) (uint32, bool) {
	if !strings.HasPrefix(ruleID, prefix) {
		return 0, false
	}
	rest := strings.TrimPrefix(ruleID, prefix)
	if rest == "" {
		return 0, false
	}
	for _, c := range rest {
		if c < '0' || c > '9' {
			return 0, false
		}
	}
	var n uint32
	for _, c := range rest {
		n = n*10 + uint32(c-'0')
	}
	return n, true
}

func formatRuleID(prefix string, seq uint32) string {
	return fmt.Sprintf("%s%02d", prefix, seq)
}
