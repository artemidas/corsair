package project

import (
	"path/filepath"
	"testing"

	"ladon/appdb"
)

func ptr(s string) *string { return &s }

func TestNormalizeForKind(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		kind     ProjectKind
		in       ProjectConfig
		wantCtx  *string
		wantImgs []string
		wantErr  string
	}{
		{
			name:     "cluster keeps context",
			kind:     KindKubernetesClusterReview,
			in:       ProjectConfig{Context: ptr("  minikube  "), Images: []string{"drop/me:latest"}},
			wantCtx:  ptr("minikube"),
			wantImgs: []string{},
		},
		{
			name:    "cluster empty context errors",
			kind:    KindKubernetesClusterReview,
			in:      ProjectConfig{Context: ptr("   ")},
			wantErr: "context must not be empty when provided",
		},
		{
			name:     "image folds legacy field",
			kind:     KindContainerImageReview,
			in:       ProjectConfig{Image: ptr(" alpine:3 "), Images: []string{"nginx:latest", "alpine:3"}},
			wantImgs: []string{"nginx:latest", "alpine:3"},
		},
		{
			name:    "image requires one",
			kind:    KindContainerImageReview,
			in:      ProjectConfig{},
			wantErr: "select at least one image",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NormalizeForKind(tt.kind, tt.in)
			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("error = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if tt.wantCtx == nil && got.Context != nil {
				t.Fatalf("context = %v, want nil", *got.Context)
			}
			if tt.wantCtx != nil && (got.Context == nil || *got.Context != *tt.wantCtx) {
				t.Fatalf("context = %v, want %q", got.Context, *tt.wantCtx)
			}
			if len(got.Images) != len(tt.wantImgs) {
				t.Fatalf("images = %v, want %v", got.Images, tt.wantImgs)
			}
			for i, image := range tt.wantImgs {
				if got.Images[i] != image {
					t.Fatalf("images = %v, want %v", got.Images, tt.wantImgs)
				}
			}
		})
	}
}

func TestParseKind(t *testing.T) {
	t.Parallel()
	if _, err := ParseKind("nope"); err == nil {
		t.Fatal("expected error")
	}
	kind, err := ParseKind("kubernetesClusterReview")
	if err != nil || kind != KindKubernetesClusterReview {
		t.Fatalf("got %q %v", kind, err)
	}
}

func TestProjectCRUD(t *testing.T) {
	dir := t.TempDir()
	db, err := appdb.Open(filepath.Join(dir, "ladon.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = appdb.Close(db) })
	svc := New(db)

	created, err := svc.CreateProject(ProjectInput{
		Name:   "  demo  ",
		Kind:   KindKubernetesClusterReview,
		Config: ProjectConfig{Context: ptr("kind-dev")},
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.Name != "demo" || created.Kind != KindKubernetesClusterReview {
		t.Fatalf("created = %+v", created)
	}

	listed, err := svc.ListProjects()
	if err != nil || len(listed) != 1 {
		t.Fatalf("list = %v %v", listed, err)
	}

	got, err := svc.GetProject(created.ID)
	if err != nil || got == nil || got.ID != created.ID {
		t.Fatalf("get = %+v %v", got, err)
	}

	updated, err := svc.UpdateProject(created.ID, ProjectInput{
		Name:   "renamed",
		Kind:   KindKubernetesClusterReview,
		Config: ProjectConfig{},
	})
	if err != nil || updated.Name != "renamed" {
		t.Fatalf("update = %+v %v", updated, err)
	}

	if err := svc.DeleteProject(created.ID); err != nil {
		t.Fatal(err)
	}
	if err := svc.DeleteProject(created.ID); err == nil {
		t.Fatal("expected missing delete error")
	}
}
