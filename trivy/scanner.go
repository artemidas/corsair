package trivy

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

)

const (
	scanTimeout  = 10 * time.Minute
	binaryName   = "trivy"
)

type runner func(ctx context.Context, bin string, args ...string) (string, error)
type looker func(name string) (string, bool)

type Scanner struct {
	run      runner
	lookBin  looker
	cacheDir string
}

func New() *Scanner {
	return &Scanner{
		run:      defaultRun,
		lookBin:  findBinary,
		cacheDir: resolveCacheDir(),
	}
}

func (s *Scanner) ScanImages(images []string, opts ScanOptions) ([]Finding, error) {
	if len(images) == 0 {
		return nil, fmt.Errorf("no images to scan")
	}
	bin, ok := s.lookBin(binaryName)
	if !ok {
		return nil, fmt.Errorf("trivy not found in PATH; install with: brew install trivy")
	}
	var out []Finding
	for _, image := range images {
		image = strings.TrimSpace(image)
		if image == "" {
			continue
		}
		findings, err := s.scanOne(bin, image, opts)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", image, err)
		}
		out = append(out, findings...)
	}
	return out, nil
}

func (s *Scanner) scanOne(bin, image string, opts ScanOptions) ([]Finding, error) {
	args, err := opts.buildArgs(image, s.cacheDir)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), scanTimeout)
	defer cancel()
	stdout, err := s.run(ctx, bin, args...)
	if err != nil {
		return nil, err
	}
	return parseReport(image, stdout)
}

func defaultRun(ctx context.Context, bin string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Stdin = nil
	cmd.Env = commandEnv()
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("timed out running trivy")
		}
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = strings.TrimSpace(stdout.String())
		}
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("%s", msg)
	}
	return stdout.String(), nil
}

func findBinary(name string) (string, bool) {
	if p, err := exec.LookPath(name); err == nil {
		return p, true
	}
	exe := name
	if runtime.GOOS == "windows" {
		exe = name + ".exe"
	}
	for _, dir := range extraBinDirs() {
		p := filepath.Join(dir, exe)
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p, true
		}
	}
	return "", false
}

func extraBinDirs() []string {
	dirs := []string{
		"/opt/homebrew/bin",
		"/usr/local/bin",
		"/usr/bin",
	}
	if home, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs, filepath.Join(home, ".local", "bin"))
	}
	return dirs
}
