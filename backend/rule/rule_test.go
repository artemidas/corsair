package rule

import (
	"path/filepath"
	"testing"

	"ladon/backend/appdb"
)

func TestPrepareInput(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		in      RuleInput
		wantErr string
	}{
		{
			name: "ok",
			in: RuleInput{
				Title: "  Privileged  ", ResourceType: " Pod ", FieldPath: " spec.hostNetwork ",
				Severity: SeverityHigh, Operator: OpEquals, ExpectedValue: "true",
			},
		},
		{name: "empty title", in: RuleInput{ResourceType: "Pod", FieldPath: "x", Severity: SeverityLow, Operator: OpPresent}, wantErr: "title must not be empty"},
		{name: "equals needs value", in: RuleInput{Title: "t", ResourceType: "Pod", FieldPath: "x", Severity: SeverityLow, Operator: OpEquals}, wantErr: "expected_value is required for this operator"},
		{name: "present skips value", in: RuleInput{Title: "t", ResourceType: "Pod", FieldPath: "x", Severity: SeverityLow, Operator: OpPresent}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := prepareInput(tt.in)
			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("error = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if tt.name == "ok" && got.title != "Privileged" {
				t.Fatalf("title = %q", got.title)
			}
		})
	}
}

func TestAllocateRuleID(t *testing.T) {
	t.Parallel()
	if formatRuleID("POD", 1) != "POD01" {
		t.Fatalf("format = %s", formatRuleID("POD", 1))
	}
	n, ok := parseRuleIDSeq("POD03", "POD")
	if !ok || n != 3 {
		t.Fatalf("parse = %d %v", n, ok)
	}
	if _, ok := parseRuleIDSeq("SA01", "POD"); ok {
		t.Fatal("prefix should not match")
	}
}

func TestRuleCRUD(t *testing.T) {
	db, err := appdb.Open(filepath.Join(t.TempDir(), "ladon.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = appdb.Close(db) })

	svc := New(db)
	listed, err := svc.ListRules()
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) < 9 {
		t.Fatalf("expected seeded rules, got %d", len(listed))
	}
	for _, r := range listed {
		if r.RuleID == "" {
			t.Fatalf("seeded rule %s missing ruleId", r.ID)
		}
	}

	created, err := svc.CreateRule(RuleInput{
		Title: "  extra  ", Description: "d", Severity: SeverityMedium,
		ResourceType: "Pod", FieldPath: "spec.hostIPC", Operator: OpEquals, ExpectedValue: "true",
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.Title != "extra" || created.RuleID == "" {
		t.Fatalf("created = %+v", created)
	}

	updated, err := svc.UpdateRule(created.ID, RuleInput{
		Title: "renamed", Description: "d", Severity: SeverityLow,
		ResourceType: "Pod", FieldPath: "spec.hostIPC", Operator: OpPresent,
	})
	if err != nil || updated.Title != "renamed" || updated.Severity != SeverityLow {
		t.Fatalf("update = %+v %v", updated, err)
	}

	if err := svc.DeleteRule(created.ID); err != nil {
		t.Fatal(err)
	}
	if err := svc.DeleteRule(created.ID); err == nil {
		t.Fatal("expected missing delete")
	}
}
