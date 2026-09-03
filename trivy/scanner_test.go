package trivy

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"ladon/rule"
)

func TestBuildArgs(t *testing.T) {
	t.Parallel()
	opts := ScanOptions{
		Scanners:      []string{"vuln", "misconfig"},
		Severity:      []string{"CRITICAL", "HIGH"},
		IgnoreUnfixed: true,
		SkipDBUpdate:  true,
		ExtraArgs:     []string{"--timeout", "5m"},
	}
	cacheDir := t.TempDir()
	dbDir := filepath.Join(cacheDir, "db")
	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dbDir, "trivy.db"), []byte("db"), 0o644); err != nil {
		t.Fatal(err)
	}
	args, err := opts.buildArgs("nginx:1.27", cacheDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(args) < 2 || args[0] != "image" || args[len(args)-1] != "nginx:1.27" {
		t.Fatalf("args = %v", args)
	}
	wantParts := []string{
		"--format", "json", "--quiet",
		"--cache-dir", cacheDir,
		"--skip-db-update", "--skip-java-db-update",
		"--scanners", "vuln,misconfig",
		"--severity", "CRITICAL,HIGH",
		"--ignore-unfixed",
		"--timeout", "5m",
	}
	for _, part := range wantParts {
		if !contains(args, part) {
			t.Fatalf("args %v missing %q", args, part)
		}
	}
}

func contains(values []string, target string) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}
	return false
}

func TestValidateExtraArgs(t *testing.T) {
	t.Parallel()
	if _, err := validateExtraArgs([]string{"bad"}); err == nil {
		t.Fatal("expected invalid arg")
	}
	if _, err := validateExtraArgs([]string{"--timeout", "5m"}); err != nil {
		t.Fatal(err)
	}
}

func TestParseReport(t *testing.T) {
	t.Parallel()
	stdout := `{
		"Results": [{
			"Target": "nginx:1.27",
			"Vulnerabilities": [{
				"VulnerabilityID": "CVE-2024-1234",
				"PkgName": "openssl",
				"InstalledVersion": "1.1.1",
				"FixedVersion": "1.1.2",
				"Severity": "HIGH",
				"Title": "OpenSSL issue"
			}],
			"Misconfigurations": [{
				"ID": "DS002",
				"Title": "Root user",
				"Severity": "MEDIUM",
				"Message": "Specify a non-root user"
			}],
			"Secrets": [{
				"RuleID": "aws-access-key-id",
				"Title": "AWS Access Key",
				"Severity": "CRITICAL",
				"Match": "AKIA..."
			}]
		}]
	}`
	got, err := parseReport("nginx:1.27", stdout)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("got %d findings", len(got))
	}
	if got[0].RuleID != "CVE-2024-1234" || got[0].Severity != rule.SeverityHigh {
		t.Fatalf("vuln = %+v", got[0])
	}
	if got[1].RuleID != "DS002" || got[1].Severity != rule.SeverityMedium {
		t.Fatalf("misconfig = %+v", got[1])
	}
	if got[2].RuleID != "aws-access-key-id" || got[2].Severity != rule.SeverityCritical {
		t.Fatalf("secret = %+v", got[2])
	}
}

func TestScanImagesUsesTrivy(t *testing.T) {
	scanner := &Scanner{
		lookBin: func(name string) (string, bool) {
			if name == "trivy" {
				return "/usr/bin/trivy", true
			}
			return "", false
		},
		run: func(_ context.Context, bin string, args ...string) (string, error) {
			if bin != "/usr/bin/trivy" {
				t.Fatalf("bin = %s", bin)
			}
			if args[len(args)-1] != "alpine:3" {
				t.Fatalf("args = %v", args)
			}
			return `{"Results":[{"Target":"alpine:3","Vulnerabilities":[]}]}`, nil
		},
	}
	got, err := scanner.ScanImages([]string{"alpine:3"}, ScanOptions{Scanners: []string{"vuln"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("got %+v", got)
	}
}
