package trivy

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCommandEnvSetsDockerHostOnMacOS(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("docker host detection is macOS-specific")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	sock := filepath.Join(home, ".docker", "run", "docker.sock")
	if _, err := os.Stat(sock); err != nil {
		t.Skip("docker desktop socket not present")
	}

	t.Setenv("DOCKER_HOST", "")
	env := commandEnv()
	got := envValue(env, "DOCKER_HOST")
	want := "unix://" + sock
	if got != want {
		t.Fatalf("DOCKER_HOST = %q, want %q", got, want)
	}
}

func TestAugmentedPathIncludesHomebrew(t *testing.T) {
	got := augmentedPath("/usr/bin")
	if !strings.Contains(got, "/opt/homebrew/bin") {
		t.Fatalf("path = %q", got)
	}
}

func envValue(env []string, key string) string {
	prefix := key + "="
	for _, entry := range env {
		if strings.HasPrefix(entry, prefix) {
			return strings.TrimPrefix(entry, prefix)
		}
	}
	return ""
}
