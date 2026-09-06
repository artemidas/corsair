package trivy

import (
	"os"
	"path/filepath"
)

var userConfigDir = os.UserConfigDir

func resolveCacheDir() string {
	ladon := ladonCacheDir()
	if dbReady(ladon) {
		return ladon
	}
	for _, dir := range systemCacheCandidates() {
		if dbReady(dir) {
			return dir
		}
	}
	return ladon
}

func ladonCacheDir() string {
	dir, err := userConfigDir()
	if err != nil {
		return ""
	}
	cache := filepath.Join(dir, "ladon", "trivy-cache")
	if err := os.MkdirAll(cache, 0o755); err != nil {
		return ""
	}
	return cache
}

func systemCacheCandidates() []string {
	var dirs []string
	if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
		dirs = append(dirs, filepath.Join(xdg, "trivy"))
	}
	if home, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs,
			filepath.Join(home, "Library", "Caches", "trivy"),
			filepath.Join(home, ".cache", "trivy"),
		)
	}
	return dirs
}

func dbReady(cacheDir string) bool {
	if cacheDir == "" {
		return false
	}
	st, err := os.Stat(filepath.Join(cacheDir, "db", "trivy.db"))
	return err == nil && !st.IsDir()
}
