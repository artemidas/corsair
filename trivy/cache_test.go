package trivy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDBReady(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if dbReady(dir) {
		t.Fatal("empty dir should not be ready")
	}
	dbDir := filepath.Join(dir, "db")
	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dbDir, "trivy.db"), []byte("db"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !dbReady(dir) {
		t.Fatal("expected ready cache")
	}
}

func TestBuildArgsSkipsDBUpdateOnlyWhenReady(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	opts := ScanOptions{SkipDBUpdate: true, Scanners: []string{"vuln"}}

	args, err := opts.buildArgs("nginx:1.27", dir)
	if err != nil {
		t.Fatal(err)
	}
	if contains(args, "--skip-db-update") {
		t.Fatalf("should not skip on first run: %v", args)
	}

	dbDir := filepath.Join(dir, "db")
	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dbDir, "trivy.db"), []byte("db"), 0o644); err != nil {
		t.Fatal(err)
	}

	args, err = opts.buildArgs("nginx:1.27", dir)
	if err != nil {
		t.Fatal(err)
	}
	if !contains(args, "--skip-db-update") {
		t.Fatalf("should skip when db exists: %v", args)
	}
}

func TestResolveCacheDirPrefersExistingSystemCache(t *testing.T) {
	system := t.TempDir()
	ladonRoot := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", system)

	systemCache := filepath.Join(system, "trivy", "db")
	if err := os.MkdirAll(systemCache, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(systemCache, "trivy.db"), []byte("db"), 0o644); err != nil {
		t.Fatal(err)
	}

	origUserConfigDir := userConfigDir
	userConfigDir = func() (string, error) {
		return ladonRoot, nil
	}
	t.Cleanup(func() { userConfigDir = origUserConfigDir })

	got := resolveCacheDir()
	want := filepath.Join(system, "trivy")
	if got != want {
		t.Fatalf("cache dir = %q, want %q", got, want)
	}
}
