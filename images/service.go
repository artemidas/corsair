package images

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

const listTimeout = 10 * time.Second

type runner func(ctx context.Context, bin string, args ...string) (string, error)
type looker func(name string) (string, bool)

type Service struct {
	run     runner
	lookBin looker
}

func New() *Service {
	return &Service{run: defaultRun, lookBin: findBinary}
}

func (s *Service) ListLocalImages() (LocalImageList, error) {
	var errs []string
	for _, runtimeName := range []string{"docker", "podman"} {
		bin, ok := s.lookBin(runtimeName)
		if !ok {
			errs = append(errs, runtimeName+" not found")
			continue
		}
		images, err := s.listWith(bin)
		if err != nil {
			errs = append(errs, runtimeName+": "+err.Error())
			continue
		}
		return LocalImageList{Runtime: runtimeName, Images: images}, nil
	}
	return LocalImageList{}, fmt.Errorf(
		"Could not list container images (tried docker, then podman). %s",
		strings.Join(errs, ". "),
	)
}

func (s *Service) listWith(bin string) ([]LocalImage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), listTimeout)
	defer cancel()
	stdout, err := s.run(ctx, bin, "images", "--format", "{{json .}}")
	if err != nil {
		return nil, err
	}
	images := parseImages(stdout)
	sort.Slice(images, func(i, j int) bool {
		return images[i].Reference < images[j].Reference
	})
	images = dedupByRef(images)
	return images, nil
}

func dedupByRef(images []LocalImage) []LocalImage {
	out := images[:0]
	var prev string
	for _, img := range images {
		if img.Reference == prev {
			continue
		}
		out = append(out, img)
		prev = img.Reference
	}
	return out
}

func defaultRun(ctx context.Context, bin string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Stdin = nil
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("timed out running %s", bin)
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
		"/opt/podman/bin",
		"/usr/bin",
		"/snap/bin",
	}
	if home, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs, filepath.Join(home, ".local/bin"), filepath.Join(home, ".docker/bin"))
	}
	if runtime.GOOS == "windows" {
		if pf := os.Getenv("ProgramFiles"); pf != "" {
			dirs = append(dirs, filepath.Join(pf, `Docker\Docker\resources\bin`))
		}
		if pf86 := os.Getenv("ProgramFiles(x86)"); pf86 != "" {
			dirs = append(dirs, filepath.Join(pf86, `Docker\Docker\resources\bin`))
		}
	}
	return dirs
}
