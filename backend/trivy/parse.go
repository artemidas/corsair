package trivy

import (
	"encoding/json"
	"fmt"
	"strings"

	"ladon/backend/rule"
)

type report struct {
	Results []result `json:"Results"`
}

type result struct {
	Target            string             `json:"Target"`
	Vulnerabilities   []vulnerability    `json:"Vulnerabilities"`
	Misconfigurations []misconfiguration `json:"Misconfigurations"`
	Secrets           []secret           `json:"Secrets"`
}

type vulnerability struct {
	VulnerabilityID  string `json:"VulnerabilityID"`
	PkgName          string `json:"PkgName"`
	InstalledVersion string `json:"InstalledVersion"`
	FixedVersion     string `json:"FixedVersion"`
	Severity         string `json:"Severity"`
	Title            string `json:"Title"`
	Description      string `json:"Description"`
}

type misconfiguration struct {
	ID          string `json:"ID"`
	Title       string `json:"Title"`
	Description string `json:"Description"`
	Severity    string `json:"Severity"`
	Message     string `json:"Message"`
}

type secret struct {
	RuleID   string `json:"RuleID"`
	Title    string `json:"Title"`
	Severity string `json:"Severity"`
	Category string `json:"Category"`
	Match    string `json:"Match"`
}

func parseReport(imageRef, stdout string) ([]Finding, error) {
	var rep report
	if err := json.Unmarshal([]byte(stdout), &rep); err != nil {
		return nil, fmt.Errorf("parse trivy json: %w", err)
	}
	var out []Finding
	for _, res := range rep.Results {
		target := firstNonEmpty(res.Target, imageRef)
		for _, vuln := range res.Vulnerabilities {
			out = append(out, vulnFinding(target, vuln))
		}
		for _, mis := range res.Misconfigurations {
			out = append(out, misconfigFinding(target, mis))
		}
		for _, sec := range res.Secrets {
			out = append(out, secretFinding(target, sec))
		}
	}
	return out, nil
}

func vulnFinding(image string, v vulnerability) Finding {
	ruleID := strings.TrimSpace(v.VulnerabilityID)
	if ruleID == "" {
		ruleID = "VULN"
	}
	title := strings.TrimSpace(v.Title)
	if title == "" {
		title = ruleID
	}
	msg := vulnMessage(v)
	return Finding{
		RuleID:       ruleID,
		RuleTitle:    title,
		Severity:     mapSeverity(v.Severity),
		ResourceKind: "ContainerImage",
		ResourceName: image,
		Message:      msg,
	}
}

func misconfigFinding(image string, m misconfiguration) Finding {
	ruleID := strings.TrimSpace(m.ID)
	if ruleID == "" {
		ruleID = "MISCONFIG"
	}
	title := strings.TrimSpace(m.Title)
	if title == "" {
		title = ruleID
	}
	msg := firstNonEmpty(strings.TrimSpace(m.Message), strings.TrimSpace(m.Description))
	return Finding{
		RuleID:       ruleID,
		RuleTitle:    title,
		Severity:     mapSeverity(m.Severity),
		ResourceKind: "ContainerImage",
		ResourceName: image,
		Message:      msg,
	}
}

func secretFinding(image string, s secret) Finding {
	ruleID := strings.TrimSpace(s.RuleID)
	if ruleID == "" {
		ruleID = "SECRET"
	}
	title := strings.TrimSpace(s.Title)
	if title == "" {
		title = ruleID
	}
	msg := strings.TrimSpace(s.Match)
	if msg == "" {
		msg = strings.TrimSpace(s.Category)
	}
	return Finding{
		RuleID:       ruleID,
		RuleTitle:    title,
		Severity:     mapSeverity(s.Severity),
		ResourceKind: "ContainerImage",
		ResourceName: image,
		Message:      msg,
	}
}

func vulnMessage(v vulnerability) string {
	parts := []string{}
	if pkg := strings.TrimSpace(v.PkgName); pkg != "" {
		parts = append(parts, pkg)
	}
	if installed := strings.TrimSpace(v.InstalledVersion); installed != "" {
		parts = append(parts, "installed "+installed)
	}
	if fixed := strings.TrimSpace(v.FixedVersion); fixed != "" {
		parts = append(parts, "fixed in "+fixed)
	}
	if len(parts) > 0 {
		return strings.Join(parts, ", ")
	}
	return firstNonEmpty(strings.TrimSpace(v.Description), strings.TrimSpace(v.Title))
}

func mapSeverity(raw string) rule.Severity {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case "CRITICAL":
		return rule.SeverityCritical
	case "HIGH":
		return rule.SeverityHigh
	case "MEDIUM":
		return rule.SeverityMedium
	default:
		return rule.SeverityLow
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
