package trivy

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func commandEnv() []string {
	base := os.Environ()
	seen := make(map[string]string, len(base)+4)
	for _, entry := range base {
		key, value, ok := strings.Cut(entry, "=")
		if ok {
			seen[key] = value
		}
	}

	if home, err := os.UserHomeDir(); err == nil {
		seen["HOME"] = home
	}
	seen["PATH"] = augmentedPath(seen["PATH"])
	if host := dockerHost(seen["DOCKER_HOST"]); host != "" {
		seen["DOCKER_HOST"] = host
	}

	out := make([]string, 0, len(seen))
	for key, value := range seen {
		out = append(out, key+"="+value)
	}
	return out
}

func augmentedPath(current string) string {
	parts := []string{}
	if current != "" {
		parts = strings.Split(current, string(os.PathListSeparator))
	}
	for _, dir := range []string{
		"/opt/homebrew/bin",
		"/usr/local/bin",
		"/usr/bin",
		"/bin",
	} {
		parts = append(parts, dir)
	}
	if home, err := os.UserHomeDir(); err == nil {
		parts = append(parts,
			filepath.Join(home, ".local", "bin"),
			filepath.Join(home, ".docker", "bin"),
		)
	}
	return joinUniquePath(parts)
}

func joinUniquePath(parts []string) string {
	seen := make(map[string]struct{}, len(parts))
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if _, ok := seen[part]; ok {
			continue
		}
		seen[part] = struct{}{}
		out = append(out, part)
	}
	return strings.Join(out, string(os.PathListSeparator))
}

func dockerHost(current string) string {
	if strings.TrimSpace(current) != "" {
		return current
	}
	if runtime.GOOS != "darwin" {
		return ""
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	candidates := []string{
		filepath.Join(home, ".docker", "run", "docker.sock"),
		"/var/run/docker.sock",
	}
	for _, sock := range candidates {
		if st, err := os.Stat(sock); err == nil && st.Mode()&os.ModeSocket != 0 {
			return "unix://" + sock
		}
	}
	return ""
}
