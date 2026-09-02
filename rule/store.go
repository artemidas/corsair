package rule

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const selectCols = `id, rule_id, title, description, severity, resource_type,
	field_path, operator, expected_value, import_id, created_at, updated_at`

func (s *Service) ListRules() ([]Rule, error) {
	ctx := context.Background()
	if err := s.backfillRuleIDs(ctx); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx,
		`SELECT `+selectCols+` FROM custom_rules ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []Rule{}
	for rows.Next() {
		r, err := scanRule(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Service) GetRule(id string) (*Rule, error) {
	ctx := context.Background()
	row := s.db.QueryRowContext(ctx,
		`SELECT `+selectCols+` FROM custom_rules WHERE id = ?`, id)
	r, err := scanRule(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *Service) CreateRule(input RuleInput) (Rule, error) {
	ctx := context.Background()
	prepared, err := prepareInput(input)
	if err != nil {
		return Rule{}, err
	}
	ruleID, err := s.allocateRuleID(ctx, prepared.resourceType, nil)
	if err != nil {
		return Rule{}, err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	r := Rule{
		ID:            uuid.NewString(),
		RuleID:        ruleID,
		Title:         prepared.title,
		Description:   prepared.description,
		Severity:      prepared.severity,
		ResourceType:  prepared.resourceType,
		FieldPath:     prepared.fieldPath,
		Operator:      prepared.operator,
		ExpectedValue: prepared.expectedValue,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO custom_rules
		 (id, rule_id, title, description, severity, resource_type, field_path, operator, expected_value, import_id, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NULL, ?, ?)`,
		r.ID, r.RuleID, r.Title, r.Description, string(r.Severity), r.ResourceType,
		r.FieldPath, string(r.Operator), r.ExpectedValue, r.CreatedAt, r.UpdatedAt,
	)
	if err != nil {
		return Rule{}, err
	}
	return r, nil
}

func (s *Service) UpdateRule(id string, input RuleInput) (Rule, error) {
	ctx := context.Background()
	prepared, err := prepareInput(input)
	if err != nil {
		return Rule{}, err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.ExecContext(ctx,
		`UPDATE custom_rules SET
		 title = ?, description = ?, severity = ?, resource_type = ?,
		 field_path = ?, operator = ?, expected_value = ?, updated_at = ?
		 WHERE id = ?`,
		prepared.title, prepared.description, string(prepared.severity),
		prepared.resourceType, prepared.fieldPath, string(prepared.operator),
		prepared.expectedValue, now, id,
	)
	if err != nil {
		return Rule{}, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return Rule{}, err
	}
	if n == 0 {
		return Rule{}, fmt.Errorf("rule '%s' not found", id)
	}
	updated, err := s.GetRule(id)
	if err != nil {
		return Rule{}, err
	}
	if updated == nil {
		return Rule{}, fmt.Errorf("rule '%s' disappeared after update", id)
	}
	return *updated, nil
}

func (s *Service) DeleteRule(id string) error {
	ctx := context.Background()
	res, err := s.db.ExecContext(ctx, `DELETE FROM custom_rules WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("rule '%s' not found", id)
	}
	return nil
}

func (s *Service) existingRuleIDs(ctx context.Context) ([]string, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT rule_id FROM custom_rules WHERE rule_id IS NOT NULL AND rule_id != ''`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (s *Service) allocateRuleID(ctx context.Context, resourceType string, preferred *string) (string, error) {
	existing, err := s.existingRuleIDs(ctx)
	if err != nil {
		return "", err
	}
	if preferred != nil {
		p := *preferred
		taken := false
		for _, id := range existing {
			if id == p {
				taken = true
				break
			}
		}
		if !taken {
			return p, nil
		}
	}
	prefix := ruleIDPrefix(resourceType)
	var max uint32
	for _, id := range existing {
		if n, ok := parseRuleIDSeq(id, prefix); ok && n > max {
			max = n
		}
	}
	return formatRuleID(prefix, max+1), nil
}

func (s *Service) backfillRuleIDs(ctx context.Context) error {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, resource_type, import_id FROM custom_rules
		 WHERE rule_id IS NULL OR rule_id = ''
		 ORDER BY created_at ASC, id ASC`)
	if err != nil {
		return err
	}
	defer rows.Close()

	type missing struct {
		id, resourceType string
		importID         *string
	}
	var todo []missing
	for rows.Next() {
		var m missing
		var importID sql.NullString
		if err := rows.Scan(&m.id, &m.resourceType, &importID); err != nil {
			return err
		}
		if importID.Valid && importID.String != "" {
			m.importID = &importID.String
		}
		todo = append(todo, m)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for _, m := range todo {
		ruleID, err := s.allocateRuleID(ctx, m.resourceType, m.importID)
		if err != nil {
			return err
		}
		if _, err := s.db.ExecContext(ctx,
			`UPDATE custom_rules SET rule_id = ? WHERE id = ?`, ruleID, m.id); err != nil {
			return err
		}
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanRule(row rowScanner) (Rule, error) {
	var (
		id, title, description, severity, resourceType, fieldPath, operator, expected, createdAt, updatedAt string
		ruleID, importID                                                                                    sql.NullString
	)
	if err := row.Scan(
		&id, &ruleID, &title, &description, &severity, &resourceType,
		&fieldPath, &operator, &expected, &importID, &createdAt, &updatedAt,
	); err != nil {
		return Rule{}, err
	}
	sev, err := ParseSeverity(severity)
	if err != nil {
		return Rule{}, err
	}
	op, err := ParseOperator(operator)
	if err != nil {
		return Rule{}, err
	}
	r := Rule{
		ID:            id,
		RuleID:        ruleID.String,
		Title:         title,
		Description:   description,
		Severity:      sev,
		ResourceType:  resourceType,
		FieldPath:     fieldPath,
		Operator:      op,
		ExpectedValue: expected,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	}
	if importID.Valid {
		r.ImportID = &importID.String
	}
	return r, nil
}
