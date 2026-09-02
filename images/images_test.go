package images

import (
	"context"
	"strings"
	"testing"
)

func TestParseImagesNDJSON(t *testing.T) {
	t.Parallel()
	got := parseImages(`
{"Repository":"nginx","Tag":"latest","ID":"sha256:abc","Size":"12.3MB"}
{"Repository":"<none>","Tag":"<none>","ID":"sha256:def","Size":"1B"}
`)
	if len(got) != 1 || got[0].Reference != "nginx:latest" {
		t.Fatalf("got %+v", got)
	}
	if got[0].SizeBytes == nil || *got[0].SizeBytes != 12897485 {
		t.Fatalf("size = %v %v", got[0].Size, got[0].SizeBytes)
	}
}

func TestParseImagesArrayAndNames(t *testing.T) {
	t.Parallel()
	got := parseImages(`[{"Id":"abc","Names":["alpine:3"],"Size":1048576}]`)
	if len(got) != 1 || got[0].Reference != "alpine:3" || got[0].ID != "abc" {
		t.Fatalf("got %+v", got)
	}
	if got[0].Size != "1.0 MB" || got[0].SizeBytes == nil || *got[0].SizeBytes != 1048576 {
		t.Fatalf("size = %+v", got[0])
	}
}

func TestComposeRef(t *testing.T) {
	t.Parallel()
	if ref, ok := composeRef("busybox", ""); !ok || ref != "busybox" {
		t.Fatalf("repo only = %q %v", ref, ok)
	}
	if _, ok := composeRef("<none>", "latest"); ok {
		t.Fatal("none repo should skip")
	}
}

func TestListLocalImagesUsesDockerThenPodman(t *testing.T) {
	svc := &Service{
		lookBin: func(name string) (string, bool) {
			if name == "docker" {
				return "", false
			}
			return "/usr/bin/" + name, true
		},
		run: func(_ context.Context, bin string, args ...string) (string, error) {
			if !strings.HasSuffix(bin, "podman") {
				t.Fatalf("bin = %s args = %v", bin, args)
			}
			return `{"Repository":"alpine","Tag":"3","ID":"1","Size":"2KB"}`, nil
		},
	}
	got, err := svc.ListLocalImages()
	if err != nil {
		t.Fatal(err)
	}
	if got.Runtime != "podman" || len(got.Images) != 1 || got.Images[0].Reference != "alpine:3" {
		t.Fatalf("got %+v", got)
	}
}

func TestListLocalImagesBothMissing(t *testing.T) {
	svc := &Service{
		lookBin: func(string) (string, bool) { return "", false },
		run:     func(context.Context, string, ...string) (string, error) { return "", nil },
	}
	if _, err := svc.ListLocalImages(); err == nil {
		t.Fatal("expected error")
	}
}
