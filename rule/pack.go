package rule

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const packVersion uint32 = 1

type matcherKind string

const (
	matcherDeclarative matcherKind = "declarative"
	matcherNative      matcherKind = "native"
)

type ImportMode string

const (
	ImportMerge   ImportMode = "merge"
	ImportReplace ImportMode = "replace"
)

type SkippedRule struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Reason string `json:"reason"`
}

type ImportSummary struct {
	Created uint32        `json:"created"`
	Updated uint32        `json:"updated"`
	Skipped []SkippedRule `json:"skipped"`
}

type rulePack struct {
	Version uint32          `yaml:"version"`
	Rules   []rulePackEntry `yaml:"rules"`
}

type rulePackEntry struct {
	ID                *string      `yaml:"id"`
	Title             string       `yaml:"title"`
	Description       string       `yaml:"description"`
	Severity          Severity     `yaml:"severity"`
	ResourceType      string       `yaml:"resourceType"`
	Matcher           matcherKind  `yaml:"matcher"`
	FieldPath         *string      `yaml:"fieldPath"`
	Operator          *Operator    `yaml:"operator"`
	ExpectedValue     scalarString `yaml:"expectedValue"`
	Notes             *string      `yaml:"notes,omitempty"`
	DuplicatesBuiltin *string      `yaml:"duplicatesBuiltin,omitempty"`
}

// scalarString accepts YAML strings, bools, and numbers as a string.
type scalarString string

func (s *scalarString) UnmarshalYAML(value *yaml.Node) error {
	if value == nil || value.Tag == "!!null" {
		*s = ""
		return nil
	}
	if value.Kind == yaml.AliasNode && value.Alias != nil {
		return s.UnmarshalYAML(value.Alias)
	}
	if value.Kind != yaml.ScalarNode {
		return fmt.Errorf("expected a string, bool, or number")
	}
	*s = scalarString(value.Value)
	return nil
}

func (s *Service) ExportRules(path string) (int, error) {
	rules, err := s.ListRules()
	if err != nil {
		return 0, err
	}
	if len(rules) == 0 {
		return 0, fmt.Errorf("no rules to export")
	}
	return writePack(path, rules)
}

func (s *Service) ImportRules(path string, mode ImportMode) (ImportSummary, error) {
	text, err := os.ReadFile(path)
	if err != nil {
		return ImportSummary{}, fmt.Errorf("failed to read '%s': %w", path, err)
	}
	pack, err := parseRulePack(string(text))
	if err != nil {
		return ImportSummary{}, err
	}
	switch mode {
	case ImportReplace:
		if err := s.deleteImported(); err != nil {
			return ImportSummary{}, err
		}
	case ImportMerge, "":
		// merge is the default
	default:
		return ImportSummary{}, fmt.Errorf("unknown import mode %q", mode)
	}
	return s.applyPack(pack)
}

func parseRulePack(text string) (rulePack, error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return rulePack{}, fmt.Errorf("YAML file is empty")
	}
	var pack rulePack
	if err := yaml.Unmarshal([]byte(trimmed), &pack); err != nil {
		return rulePack{}, fmt.Errorf("invalid rule pack YAML: %w", err)
	}
	if pack.Version != packVersion {
		return rulePack{}, fmt.Errorf("unsupported rule pack version: %d", pack.Version)
	}
	for i := range pack.Rules {
		if pack.Rules[i].Matcher == "" {
			pack.Rules[i].Matcher = matcherDeclarative
		}
	}
	return pack, nil
}

func serializeRulePack(rules []Rule) (string, error) {
	pack := rulePack{
		Version: packVersion,
		Rules:   make([]rulePackEntry, 0, len(rules)),
	}
	for _, r := range rules {
		pack.Rules = append(pack.Rules, toPackEntry(r))
	}
	out, err := yaml.Marshal(&pack)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func toPackEntry(r Rule) rulePackEntry {
	id := r.ID
	if r.ImportID != nil && *r.ImportID != "" {
		id = *r.ImportID
	}
	op := r.Operator
	fieldPath := r.FieldPath
	return rulePackEntry{
		ID:            &id,
		Title:         r.Title,
		Description:   r.Description,
		Severity:      r.Severity,
		ResourceType:  r.ResourceType,
		Matcher:       matcherDeclarative,
		FieldPath:     &fieldPath,
		Operator:      &op,
		ExpectedValue: scalarString(r.ExpectedValue),
	}
}

type preparedEntry struct {
	skip     *SkippedRule
	importID *string
	input    RuleInput
}

func prepareEntry(entry rulePackEntry) preparedEntry {
	id := ""
	if entry.ID != nil {
		id = strings.TrimSpace(*entry.ID)
	}
	title := entry.Title

	if entry.Matcher == matcherNative {
		return preparedEntry{skip: &SkippedRule{
			ID: id, Title: title, Reason: "native matcher is not supported",
		}}
	}

	fieldPath := ""
	if entry.FieldPath != nil {
		fieldPath = strings.TrimSpace(*entry.FieldPath)
	}
	if fieldPath == "" {
		return preparedEntry{skip: &SkippedRule{
			ID: id, Title: title, Reason: "declarative rule is missing fieldPath",
		}}
	}

	if entry.Operator == nil {
		return preparedEntry{skip: &SkippedRule{
			ID: id, Title: title, Reason: "declarative rule is missing operator",
		}}
	}
	operator := *entry.Operator
	if _, err := ParseOperator(string(operator)); err != nil {
		return preparedEntry{skip: &SkippedRule{
			ID: id, Title: title, Reason: "declarative rule is missing operator",
		}}
	}

	expected := string(entry.ExpectedValue)
	if operator.NeedsExpectedValue() && strings.TrimSpace(expected) == "" {
		return preparedEntry{skip: &SkippedRule{
			ID:     id,
			Title:  title,
			Reason: fmt.Sprintf("operator %s requires a non-empty expectedValue", operator),
		}}
	}

	var importID *string
	if id != "" {
		importID = &id
	}
	return preparedEntry{
		importID: importID,
		input: RuleInput{
			Title:         entry.Title,
			Description:   entry.Description,
			Severity:      entry.Severity,
			ResourceType:  entry.ResourceType,
			FieldPath:     fieldPath,
			Operator:      operator,
			ExpectedValue: expected,
		},
	}
}

func ensureYAMLExtension(path string) string {
	lower := strings.ToLower(path)
	if strings.HasSuffix(lower, ".yaml") || strings.HasSuffix(lower, ".yml") {
		return path
	}
	return path + ".yaml"
}

func writePack(path string, rules []Rule) (int, error) {
	path = ensureYAMLExtension(path)
	yamlText, err := serializeRulePack(rules)
	if err != nil {
		return 0, err
	}
	if err := os.WriteFile(path, []byte(yamlText), 0o644); err != nil {
		return 0, fmt.Errorf("failed to write '%s': %w", path, err)
	}
	return len(rules), nil
}

func (s *Service) applyPack(pack rulePack) (ImportSummary, error) {
	summary := ImportSummary{Skipped: []SkippedRule{}}
	for _, entry := range pack.Rules {
		prepared := prepareEntry(entry)
		if prepared.skip != nil {
			summary.Skipped = append(summary.Skipped, *prepared.skip)
			continue
		}
		if prepared.importID != nil {
			existing, err := s.getByImportID(*prepared.importID)
			if err != nil {
				return ImportSummary{}, err
			}
			if existing != nil {
				if _, err := s.UpdateRule(existing.ID, prepared.input); err != nil {
					return ImportSummary{}, err
				}
				summary.Updated++
				continue
			}
		}
		if _, err := s.insertRule(prepared.input, prepared.importID); err != nil {
			return ImportSummary{}, err
		}
		summary.Created++
	}
	return summary, nil
}
