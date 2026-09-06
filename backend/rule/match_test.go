package rule

import (
	"encoding/json"
	"testing"
)

func mustJSON(t *testing.T, raw string) any {
	t.Helper()
	var v any
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		t.Fatal(err)
	}
	return v
}

func twoContainerPod(t *testing.T) any {
	t.Helper()
	return mustJSON(t, `{
		"spec": {
			"containers": [
				{
					"resources": { "limits": { "cpu": "500m" } },
					"securityContext": { "capabilities": { "drop": ["ALL"] } }
				},
				{ "name": "bare" }
			]
		}
	}`)
}

func TestParsePath(t *testing.T) {
	t.Parallel()
	got := parsePath("spec.containers[*].foo")
	if len(got) != 4 || got[0].field != "spec" || got[1].field != "containers" || !got[2].iterate || got[3].field != "foo" {
		t.Fatalf("attached = %+v", got)
	}
	got = parsePath("spec.containers.[*].foo")
	if len(got) != 4 || got[1].field != "containers" || !got[2].iterate {
		t.Fatalf("standalone = %+v", got)
	}
}

func TestEvaluateFieldPath(t *testing.T) {
	t.Parallel()
	leaves := EvaluateFieldPath(twoContainerPod(t), "spec.containers[*].resources.limits.cpu")
	if len(leaves) != 2 || leaves[0] != "500m" || leaves[1] != nil {
		t.Fatalf("missing intermediate = %#v", leaves)
	}

	v := mustJSON(t, `{"spec":{"containers":[{"name":"a"},{"name":"b"}]}}`)
	leaves = EvaluateFieldPath(v, "spec.containers")
	if len(leaves) != 1 {
		t.Fatalf("array leaf count = %d", len(leaves))
	}
	arr, ok := leaves[0].([]any)
	if !ok || len(arr) != 2 {
		t.Fatalf("array leaf = %#v", leaves[0])
	}
}

func TestEvaluateOperator(t *testing.T) {
	t.Parallel()
	if !EvaluateOperator([]any{true, false}, OpEquals, "true") {
		t.Fatal("equals true")
	}
	if EvaluateOperator([]any{true, false}, OpEquals, "maybe") {
		t.Fatal("equals maybe")
	}

	if !EvaluateOperator([]any{true, nil}, OpNotEquals, "true") {
		t.Fatal("notEquals mixed")
	}
	if EvaluateOperator([]any{true}, OpNotEquals, "true") {
		t.Fatal("notEquals all match")
	}
	if EvaluateOperator(nil, OpNotEquals, "true") {
		t.Fatal("notEquals empty")
	}

	if !EvaluateOperator([]any{"x"}, OpPresent, "") {
		t.Fatal("present string")
	}
	if EvaluateOperator([]any{nil}, OpPresent, "") {
		t.Fatal("present null")
	}
	if EvaluateOperator([]any{""}, OpPresent, "") {
		t.Fatal("present empty string")
	}
	if EvaluateOperator([]any{[]any{}}, OpPresent, "") {
		t.Fatal("present empty array")
	}

	leaves := EvaluateFieldPath(twoContainerPod(t), "spec.containers[*].resources.limits.cpu")
	if !EvaluateOperator(leaves, OpAbsent, "") {
		t.Fatal("absent two containers")
	}
	if EvaluateOperator([]any{"500m"}, OpAbsent, "") {
		t.Fatal("absent present value")
	}

	leaves = EvaluateFieldPath(twoContainerPod(t), "spec.containers[*].securityContext.capabilities.drop")
	if !EvaluateOperator(leaves, OpArrayExcludes, "ALL") {
		t.Fatal("arrayExcludes two containers")
	}
	if EvaluateOperator([]any{[]any{"ALL"}}, OpArrayExcludes, "ALL") {
		t.Fatal("arrayExcludes contained")
	}

	pod := mustJSON(t, `{
		"spec": {
			"hostNetwork": true,
			"containers": [{"securityContext": {"privileged": true}}]
		}
	}`)
	leaves = EvaluateFieldPath(pod, "spec.containers[*].securityContext.privileged")
	if !EvaluateOperator(leaves, OpEquals, "true") {
		t.Fatal("privileged equals")
	}
}
