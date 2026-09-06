package rule

import (
	"testing"
)

func TestValidateAndEval(t *testing.T) {
	t.Parallel()

	pod := map[string]any{
		"spec": map[string]any{
			"hostNetwork": true,
			"containers": []any{
				map[string]any{
					"name": "main",
					"securityContext": map[string]any{
						"privileged": true,
					},
				},
			},
		},
	}
	safe := map[string]any{
		"spec": map[string]any{
			"hostNetwork": false,
			"containers": []any{
				map[string]any{"name": "main"},
			},
		},
	}

	tests := []struct {
		name    string
		src     string
		input   any
		want    []string
		wantErr bool
	}{
		{
			name: "boolean violation",
			src: `package ladon
violation if { input.spec.hostNetwork == true }`,
			input: pod,
			want:  []string{""},
		},
		{
			name: "no match",
			src: `package ladon
violation if { input.spec.hostNetwork == true }`,
			input: safe,
		},
		{
			name: "set messages",
			src: `package ladon
violation contains msg if {
	some c in input.spec.containers
	c.securityContext.privileged == true
	msg := sprintf("container %q is privileged", [c.name])
}`,
			input: pod,
			want:  []string{`container "main" is privileged`},
		},
		{
			name: "missing package is added",
			src:  "violation if { input.spec.hostNetwork == true }",
			input: pod,
			want:  []string{""},
		},
		{
			name:    "wrong package",
			src:     "package other\nviolation if { true }",
			wantErr: true,
		},
		{
			name:    "missing violation",
			src:     "package ladon\nallow if { true }",
			wantErr: true,
		},
		{
			name:    "syntax error",
			src:     "package ladon\nviolation if {",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			prepared, err := Prepare(tt.src)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			got, err := prepared.Eval(t.Context(), tt.input)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("messages = %#v, want %#v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("messages = %#v, want %#v", got, tt.want)
				}
			}
		})
	}
}

func TestConvertDeclarative(t *testing.T) {
	t.Parallel()
	src, err := ConvertDeclarative("spec.hostNetwork", "equals", "true")
	if err != nil {
		t.Fatal(err)
	}
	prepared, err := Prepare(src)
	if err != nil {
		t.Fatal(err)
	}
	hit, err := prepared.Eval(t.Context(), map[string]any{"spec": map[string]any{"hostNetwork": true}})
	if err != nil || len(hit) != 1 {
		t.Fatalf("hit = %#v %v", hit, err)
	}
	miss, err := prepared.Eval(t.Context(), map[string]any{"spec": map[string]any{"hostNetwork": false}})
	if err != nil || len(miss) != 0 {
		t.Fatalf("miss = %#v %v", miss, err)
	}

	src, err = ConvertDeclarative("spec.containers[*].securityContext.privileged", "equals", "true")
	if err != nil {
		t.Fatal(err)
	}
	prepared, err = Prepare(src)
	if err != nil {
		t.Fatal(err)
	}
	pod := map[string]any{
		"spec": map[string]any{
			"containers": []any{
				map[string]any{"securityContext": map[string]any{"privileged": true}},
			},
		},
	}
	got, err := prepared.Eval(t.Context(), pod)
	if err != nil || len(got) != 1 {
		t.Fatalf("privileged = %#v %v", got, err)
	}
}
