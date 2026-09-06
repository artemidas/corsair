package rule

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ladon/backend/appdb"
)

const samplePack = `
version: 1
rules:
  - id: POD010
    title: CPU limit not set
    description: A container without a CPU limit can consume unbounded CPU.
    severity: medium
    resourceType: Pod
    matcher: declarative
    fieldPath: spec.containers[*].resources.limits.cpu
    operator: absent
    expectedValue: ""
  - id: POD001
    title: Privileged container
    severity: critical
    resourceType: Pod
    matcher: native
    notes: implemented in rules.rs
  - id: POD002
    title: Host network
    severity: high
    resourceType: Pod
    fieldPath: spec.hostNetwork
    operator: equals
    expectedValue: true
  - title: No id always inserts
    severity: low
    resourceType: Pod
    fieldPath: spec.hostIPC
    operator: equals
    expectedValue: "true"
`

func TestParseRejectsUnsupportedVersion(t *testing.T) {
	t.Parallel()
	_, err := parseRulePack("version: 2\nrules: []")
	if err == nil || err.Error() != "unsupported rule pack version: 2" {
		t.Fatalf("err = %v", err)
	}
}

func TestPrepareSkipsNativeAndIncomplete(t *testing.T) {
	t.Parallel()
	pack, err := parseRulePack(samplePack)
	if err != nil {
		t.Fatal(err)
	}
	var skips []string
	ready := 0
	for _, entry := range pack.Rules {
		prepared := prepareEntry(entry)
		if prepared.skip != nil {
			skips = append(skips, prepared.skip.Reason)
			continue
		}
		ready++
	}
	if len(skips) != 1 {
		t.Fatalf("skips = %v", skips)
	}
	if !strings.Contains(skips[0], "native matcher") {
		t.Fatalf("skip reason = %q", skips[0])
	}
	if ready != 3 {
		t.Fatalf("ready = %d", ready)
	}
}

func TestExpectedValueAcceptsBool(t *testing.T) {
	t.Parallel()
	pack, err := parseRulePack(samplePack)
	if err != nil {
		t.Fatal(err)
	}
	var pod002 *rulePackEntry
	for i := range pack.Rules {
		if pack.Rules[i].ID != nil && *pack.Rules[i].ID == "POD002" {
			pod002 = &pack.Rules[i]
			break
		}
	}
	if pod002 == nil {
		t.Fatal("POD002 missing")
	}
	if string(pod002.ExpectedValue) != "true" {
		t.Fatalf("expectedValue = %q", pod002.ExpectedValue)
	}
}

func TestExportRoundTripPreservesImportID(t *testing.T) {
	t.Parallel()
	importID := "POD010"
	rule := Rule{
		ID:            "uuid-1",
		RuleID:        "POD10",
		Title:         "CPU limit not set",
		Severity:      SeverityMedium,
		ResourceType:  "Pod",
		FieldPath:     "spec.containers[*].resources.limits.cpu",
		Operator:      OpAbsent,
		ExpectedValue: "",
		ImportID:      &importID,
	}
	yamlText, err := serializeRulePack([]Rule{rule})
	if err != nil {
		t.Fatal(err)
	}
	pack, err := parseRulePack(yamlText)
	if err != nil {
		t.Fatal(err)
	}
	if pack.Version != 1 {
		t.Fatalf("version = %d", pack.Version)
	}
	if pack.Rules[0].ID == nil || *pack.Rules[0].ID != "POD010" {
		t.Fatalf("id = %v", pack.Rules[0].ID)
	}
	if pack.Rules[0].Operator == nil || *pack.Rules[0].Operator != OpAbsent {
		t.Fatalf("operator = %v", pack.Rules[0].Operator)
	}
}

func TestMergeReimportUpdatesInsteadOfDuplicating(t *testing.T) {
	svc := testRuleService(t)
	pack, err := parseRulePack(samplePack)
	if err != nil {
		t.Fatal(err)
	}
	first, err := svc.applyPack(pack)
	if err != nil {
		t.Fatal(err)
	}
	if first.Created != 3 || first.Updated != 0 || len(first.Skipped) != 1 {
		t.Fatalf("first = %+v", first)
	}

	second, err := svc.applyPack(pack)
	if err != nil {
		t.Fatal(err)
	}
	if second.Created != 1 || second.Updated != 2 || len(second.Skipped) != 1 {
		t.Fatalf("second = %+v", second)
	}

	rows, err := svc.ListRules()
	if err != nil {
		t.Fatal(err)
	}
	imported := 0
	for _, r := range rows {
		if r.ImportID != nil {
			imported++
		}
	}
	if imported != 2 {
		t.Fatalf("imported = %d", imported)
	}
}

func TestReplaceLeavesHandAuthoredRows(t *testing.T) {
	svc := testRuleService(t)
	if _, err := svc.CreateRule(RuleInput{
		Title:         "hand",
		Severity:      SeverityLow,
		ResourceType:  "Pod",
		FieldPath:     "spec.hostNetwork",
		Operator:      OpEquals,
		ExpectedValue: "true",
	}); err != nil {
		t.Fatal(err)
	}

	pack, err := parseRulePack(samplePack)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.applyPack(pack); err != nil {
		t.Fatal(err)
	}
	if err := svc.deleteImported(); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.applyPack(pack); err != nil {
		t.Fatal(err)
	}

	rows, err := svc.ListRules()
	if err != nil {
		t.Fatal(err)
	}
	var hasHand, hasPOD010 bool
	for _, r := range rows {
		if r.Title == "hand" && r.ImportID == nil {
			hasHand = true
		}
		if r.ImportID != nil && *r.ImportID == "POD010" {
			hasPOD010 = true
		}
	}
	if !hasHand || !hasPOD010 {
		t.Fatalf("hand=%v POD010=%v", hasHand, hasPOD010)
	}
}

func TestExportRulesWritesYAMLExtension(t *testing.T) {
	svc := testRuleService(t)
	pack, err := parseRulePack(samplePack)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.applyPack(pack); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "ladon-rules")
	if _, err := svc.ExportRules(path); err != nil {
		t.Fatal(err)
	}
	text, err := os.ReadFile(path + ".yaml")
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := parseRulePack(string(text))
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, entry := range parsed.Rules {
		if entry.ID != nil && *entry.ID == "POD010" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("exported pack missing POD010")
	}
}

func testRuleService(t *testing.T) *Service {
	t.Helper()
	db, err := appdb.Open(filepath.Join(t.TempDir(), "ladon.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = appdb.Close(db) })
	return New(db)
}
