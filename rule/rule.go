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

func (o Operator) NeedsExpectedValue() bool {
	switch o {
	case OpEquals, OpNotEquals, OpArrayExcludes:
		return true
	default:
		return false
	}
}

type Rule struct {
	ID            string   `json:"id"`
	RuleID        string   `json:"ruleId"`
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	Severity      Severity `json:"severity"`
	ResourceType  string   `json:"resourceType"`
	FieldPath     string   `json:"fieldPath"`
	Operator      Operator `json:"operator"`
	ExpectedValue string   `json:"expectedValue"`
	ImportID      *string  `json:"importId"`
	CreatedAt     string   `json:"createdAt"`
	UpdatedAt     string   `json:"updatedAt"`
}

type RuleInput struct {
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	Severity      Severity `json:"severity"`
	ResourceType  string   `json:"resourceType"`
	FieldPath     string   `json:"fieldPath"`
	Operator      Operator `json:"operator"`
	ExpectedValue string   `json:"expectedValue"`
}

type preparedInput struct {
	title         string
	description   string
	severity      Severity
	resourceType  string
	fieldPath     string
	operator      Operator
	expectedValue string
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
	fieldPath := strings.TrimSpace(input.FieldPath)
	if title == "" {
		return preparedInput{}, fmt.Errorf("title must not be empty")
	}
	if resourceType == "" {
		return preparedInput{}, fmt.Errorf("resource_type must not be empty")
	}
	if fieldPath == "" {
		return preparedInput{}, fmt.Errorf("field_path must not be empty")
	}
	if input.Operator == "" {
		input.Operator = OpEquals
	}
	if _, err := ParseOperator(string(input.Operator)); err != nil {
		return preparedInput{}, err
	}
	if _, err := ParseSeverity(string(input.Severity)); err != nil {
		return preparedInput{}, err
	}
	if input.Operator.NeedsExpectedValue() && strings.TrimSpace(input.ExpectedValue) == "" {
		return preparedInput{}, fmt.Errorf("expected_value is required for this operator")
	}
	return preparedInput{
		title:         title,
		description:   strings.TrimSpace(input.Description),
		severity:      input.Severity,
		resourceType:  resourceType,
		fieldPath:     fieldPath,
		operator:      input.Operator,
		expectedValue: input.ExpectedValue,
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
