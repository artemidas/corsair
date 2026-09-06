package rule

import (
	"context"
	"fmt"
	"strings"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"
)

const (
	policyFilename = "policy.rego"
	policyPackage  = "ladon"
	policyQuery    = "data.ladon.violation"
	violationRule  = "violation"
)

// DefaultRego is the stub policy shown for a new rule. It compiles and never fires.
const DefaultRego = `package ladon

# input is one Kubernetes resource of the selected kind.
# Example:
#   some c in input.spec.containers
#   c.securityContext.privileged == true
violation if {
	false
}
`

// Prepared is a compiled Rego module ready to evaluate against resource JSON.
type Prepared struct {
	query rego.PreparedEvalQuery
}

func parsePolicy(src string) (*ast.Module, ast.RegoVersion, error) {
	mod, err := ast.ParseModuleWithOpts(policyFilename, src, ast.ParserOptions{RegoVersion: ast.RegoV1})
	if err == nil {
		return mod, ast.RegoV1, nil
	}
	mod, errV0 := ast.ParseModuleWithOpts(policyFilename, src, ast.ParserOptions{RegoVersion: ast.RegoV0})
	if errV0 == nil {
		return mod, ast.RegoV0, nil
	}
	return nil, ast.RegoV1, fmt.Errorf("invalid Rego: %w", err)
}

func normalizeSource(src string) string {
	trimmed := strings.TrimSpace(src)
	for _, line := range strings.Split(trimmed, "\n") {
		t := strings.TrimSpace(line)
		if t == "" || strings.HasPrefix(t, "#") {
			continue
		}
		if strings.HasPrefix(t, "package ") {
			return trimmed
		}
		break
	}
	if trimmed == "" {
		return strings.TrimSpace(DefaultRego)
	}
	return "package ladon\n\n" + trimmed
}

func packageName(mod *ast.Module) string {
	if mod == nil || mod.Package == nil {
		return ""
	}
	path := mod.Package.Path.String()
	return strings.TrimPrefix(path, "data.")
}

func declaresViolation(mod *ast.Module) bool {
	if mod == nil {
		return false
	}
	for _, r := range mod.Rules {
		if r == nil || r.Head == nil {
			continue
		}
		if r.Head.Name.String() == violationRule {
			return true
		}
		if len(r.Head.Reference) > 0 && r.Head.Reference[0] != nil {
			if r.Head.Reference[0].Value.Compare(ast.Var(violationRule)) == 0 {
				return true
			}
		}
	}
	return false
}

func policyCapabilities() *ast.Capabilities {
	caps := ast.CapabilitiesForThisVersion()
	deny := map[string]struct{}{
		"http.send":          {},
		"net.lookup_ip_addr": {},
		"opa.runtime":        {},
	}
	kept := make([]*ast.Builtin, 0, len(caps.Builtins))
	for _, b := range caps.Builtins {
		if _, skip := deny[b.Name]; skip {
			continue
		}
		kept = append(kept, b)
	}
	caps.Builtins = kept
	return caps
}

// Validate compiles src and returns the normalized module text.
func Validate(src string) (string, error) {
	normalized := normalizeSource(src)
	mod, version, err := parsePolicy(normalized)
	if err != nil {
		return "", err
	}
	if pkg := packageName(mod); pkg != policyPackage {
		return "", fmt.Errorf("Rego package must be %s, got %s", policyPackage, pkg)
	}
	if !declaresViolation(mod) {
		return "", fmt.Errorf("Rego module must declare a %s rule", violationRule)
	}
	r := rego.New(
		rego.Query(policyQuery),
		rego.Module(policyFilename, normalized),
		rego.SetRegoVersion(version),
		rego.Capabilities(policyCapabilities()),
	)
	if _, err := r.PrepareForEval(context.Background()); err != nil {
		return "", fmt.Errorf("invalid Rego: %w", err)
	}
	return normalized, nil
}

// Prepare compiles src for repeated evaluation.
func Prepare(src string) (*Prepared, error) {
	normalized, err := Validate(src)
	if err != nil {
		return nil, err
	}
	_, version, err := parsePolicy(normalized)
	if err != nil {
		return nil, err
	}
	r := rego.New(
		rego.Query(policyQuery),
		rego.Module(policyFilename, normalized),
		rego.SetRegoVersion(version),
		rego.Capabilities(policyCapabilities()),
	)
	q, err := r.PrepareForEval(context.Background())
	if err != nil {
		return nil, fmt.Errorf("invalid Rego: %w", err)
	}
	return &Prepared{query: q}, nil
}

// Eval returns finding messages for one resource. An empty slice means no finding.
// A boolean violation contributes a single empty message so the caller can use the rule title.
func (p *Prepared) Eval(ctx context.Context, input any) ([]string, error) {
	if p == nil {
		return nil, fmt.Errorf("rego policy is not prepared")
	}
	rs, err := p.query.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return nil, err
	}
	if len(rs) == 0 {
		return nil, nil
	}
	var out []string
	for _, result := range rs {
		for _, expr := range result.Expressions {
			out = append(out, messagesFromValue(expr.Value)...)
		}
	}
	return out, nil
}

func messagesFromValue(v any) []string {
	switch x := v.(type) {
	case nil:
		return nil
	case bool:
		if x {
			return []string{""}
		}
		return nil
	case string:
		return []string{x}
	case []any:
		if len(x) == 0 {
			return nil
		}
		out := make([]string, 0, len(x))
		for _, el := range x {
			out = append(out, messagesFromValue(el)...)
		}
		return out
	case map[string]any:
		if len(x) == 0 {
			return nil
		}
		out := make([]string, 0, len(x))
		for k, el := range x {
			switch ev := el.(type) {
			case bool:
				if ev {
					out = append(out, k)
				}
			default:
				out = append(out, messagesFromValue(el)...)
			}
		}
		if len(out) == 0 {
			return []string{""}
		}
		return out
	default:
		return []string{fmt.Sprint(x)}
	}
}
