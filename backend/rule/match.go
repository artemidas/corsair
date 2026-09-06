package rule

import (
	"encoding/json"
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

// EvaluateFieldPath walks a dotted path through JSON. "[*]" iterates arrays,
// either as its own segment or attached to a field (`containers[*]`).
// Missing objects contribute nil so cardinality is preserved across branches.
func EvaluateFieldPath(value any, path string) []any {
	current := []any{value}
	for _, seg := range parsePath(path) {
		next := make([]any, 0)
		for _, v := range current {
			if seg.iterate {
				arr, ok := v.([]any)
				if ok {
					next = append(next, arr...)
				}
				continue
			}
			obj, ok := v.(map[string]any)
			if !ok {
				next = append(next, nil)
				continue
			}
			if got, exists := obj[seg.field]; exists {
				next = append(next, got)
			} else {
				next = append(next, nil)
			}
		}
		current = next
	}
	return current
}

// ValueMatches coerces a JSON leaf to a string and compares it to expected.
func ValueMatches(value any, expected string) bool {
	switch v := value.(type) {
	case nil:
		return expected == "null"
	case string:
		return v == expected
	case bool:
		return strconv.FormatBool(v) == expected
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64) == expected
	case json.Number:
		return v.String() == expected
	default:
		return false
	}
}

func isPresent(v any) bool {
	switch x := v.(type) {
	case nil:
		return false
	case []any:
		return len(x) > 0
	case string:
		return x != ""
	default:
		return true
	}
}

func arrayContains(v any, expected string) bool {
	arr, ok := v.([]any)
	if !ok {
		return false
	}
	for _, el := range arr {
		if ValueMatches(el, expected) {
			return true
		}
	}
	return false
}

// EvaluateOperator returns true when the operator's "raise a finding"
// condition holds for the resolved leaves of one resource.
func EvaluateOperator(leaves []any, operator Operator, expected string) bool {
	switch operator {
	case OpEquals:
		for _, v := range leaves {
			if ValueMatches(v, expected) {
				return true
			}
		}
		return false
	case OpNotEquals:
		if len(leaves) == 0 {
			return false
		}
		for _, v := range leaves {
			if !ValueMatches(v, expected) {
				return true
			}
		}
		return false
	case OpPresent:
		for _, v := range leaves {
			if isPresent(v) {
				return true
			}
		}
		return false
	case OpAbsent:
		if len(leaves) == 0 {
			return true
		}
		for _, v := range leaves {
			if !isPresent(v) {
				return true
			}
		}
		return false
	case OpArrayExcludes:
		if len(leaves) == 0 {
			return true
		}
		for _, v := range leaves {
			if !arrayContains(v, expected) {
				return true
			}
		}
		return false
	default:
		return false
	}
}
