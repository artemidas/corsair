package appdb

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type pathSegment struct {
	field   string
	iterate bool
}

func parsePath(path string) []pathSegment {
	var segments []pathSegment
	for _, raw := range strings.Split(path, ".") {
		if raw == "[*]" {
			segments = append(segments, pathSegment{iterate: true})
			continue
		}
		if field, ok := strings.CutSuffix(raw, "[*]"); ok {
			segments = append(segments, pathSegment{field: field})
			segments = append(segments, pathSegment{iterate: true})
			continue
		}
		segments = append(segments, pathSegment{field: raw})
	}
	return segments
}

func regoLiteral(expected string) string {
	switch expected {
	case "true", "false", "null":
		return expected
	}
	if expected != "" {
		if _, err := strconv.ParseFloat(expected, 64); err == nil {
			return expected
		}
	}
	b, err := json.Marshal(expected)
	if err != nil {
		return strconv.Quote(expected)
	}
	return string(b)
}

func convertDeclarativeToRego(fieldPath, operator, expected string) (string, error) {
	fieldPath = strings.TrimSpace(fieldPath)
	if fieldPath == "" {
		return "", fmt.Errorf("field_path must not be empty")
	}
	if operator == "" {
		operator = "equals"
	}

	ref := "input"
	var body []string
	iter := 0
	var parent string
	var lastField string
	for _, seg := range parsePath(fieldPath) {
		if seg.iterate {
			name := fmt.Sprintf("v%d", iter)
			iter++
			body = append(body, fmt.Sprintf("\tsome %s in %s", name, ref))
			parent = ""
			lastField = ""
			ref = name
			continue
		}
		if seg.field == "" {
			continue
		}
		parent = ref
		lastField = seg.field
		ref = ref + "." + seg.field
	}

	lit := regoLiteral(expected)
	switch operator {
	case "equals":
		body = append(body, fmt.Sprintf("\t%s == %s", ref, lit))
	case "notEquals":
		body = append(body, fmt.Sprintf("\t%s != %s", ref, lit))
	case "present":
		body = append(body, fmt.Sprintf("\t%s != null", ref))
	case "absent":
		body = append(body, fmt.Sprintf("\tnot %s", ref))
	case "arrayExcludes":
		if parent != "" && lastField != "" {
			body = append(body, fmt.Sprintf("\tnot %s in object.get(%s, %s, [])", lit, parent, regoLiteral(lastField)))
		} else {
			body = append(body, fmt.Sprintf("\tnot %s in object.get(input, %s, [])", lit, regoLiteral(fieldPath)))
		}
	default:
		return "", fmt.Errorf("unknown operator %q", operator)
	}

	return fmt.Sprintf("package ladon\n\nviolation if {\n%s\n}\n", strings.Join(body, "\n")), nil
}
