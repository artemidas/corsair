package trivy

import (
	"fmt"
	"strings"
)

type ScanOptions struct {
	Scanners      []string `json:"scanners"`
	Severity      []string `json:"severity"`
	IgnoreUnfixed bool     `json:"ignoreUnfixed"`
	SkipDBUpdate  bool     `json:"skipDbUpdate"`
	ExtraArgs     []string `json:"extraArgs"`
}

func (o ScanOptions) buildArgs(image, cacheDir string) ([]string, error) {
	args := []string{"image", "--format", "json", "--quiet"}
	if cacheDir != "" {
		args = append(args, "--cache-dir", cacheDir)
	}
	if o.SkipDBUpdate && dbReady(cacheDir) {
		args = append(args, "--skip-db-update", "--skip-java-db-update")
	}
	if scanners := joinCSV(o.Scanners); scanners != "" {
		args = append(args, "--scanners", scanners)
	}
	if severity := joinCSV(o.Severity); severity != "" {
		args = append(args, "--severity", severity)
	}
	if o.IgnoreUnfixed {
		args = append(args, "--ignore-unfixed")
	}
	extra, err := validateExtraArgs(o.ExtraArgs)
	if err != nil {
		return nil, err
	}
	args = append(args, extra...)
	args = append(args, image)
	return args, nil
}

func joinCSV(values []string) string {
	var out []string
	for _, v := range values {
		trimmed := strings.TrimSpace(v)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return strings.Join(out, ",")
}

func validateExtraArgs(args []string) ([]string, error) {
	out := make([]string, 0, len(args))
	for _, raw := range args {
		arg := strings.TrimSpace(raw)
		if arg == "" {
			continue
		}
		if strings.ContainsAny(arg, ";&|`$") {
			return nil, fmt.Errorf("extra trivy arg %q contains invalid characters", arg)
		}
		if strings.HasPrefix(arg, "-") {
			out = append(out, arg)
			continue
		}
		if len(out) == 0 {
			return nil, fmt.Errorf("extra trivy arg %q must start with '-'", arg)
		}
		out = append(out, arg)
	}
	return out, nil
}
