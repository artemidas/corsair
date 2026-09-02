package images

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
)

type LocalImage struct {
	ID         string  `json:"id"`
	Reference  string  `json:"reference"`
	Repository string  `json:"repository"`
	Tag        string  `json:"tag"`
	Size       string  `json:"size"`
	SizeBytes  *uint64 `json:"sizeBytes"`
}

type LocalImageList struct {
	Runtime string       `json:"runtime"`
	Images  []LocalImage `json:"images"`
}

type rawImage struct {
	Repository string
	Tag        string
	ID         string
	Size       json.RawMessage
	Names      []string
	RepoTags   []string
}

func (r *rawImage) UnmarshalJSON(b []byte) error {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	r.Repository = unmarshalString(m, "repository", "Repository")
	r.Tag = unmarshalString(m, "tag", "Tag")
	r.ID = unmarshalString(m, "id", "ID", "Id")
	r.Size = firstRaw(m, "size", "Size")
	r.Names = unmarshalStringSlice(m, "names", "Names")
	r.RepoTags = unmarshalStringSlice(m, "repoTags", "RepoTags")
	return nil
}

func firstRaw(m map[string]json.RawMessage, keys ...string) json.RawMessage {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			return v
		}
	}
	return nil
}

func unmarshalString(m map[string]json.RawMessage, keys ...string) string {
	raw := firstRaw(m, keys...)
	if len(raw) == 0 {
		return ""
	}
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	return ""
}

func unmarshalStringSlice(m map[string]json.RawMessage, keys ...string) []string {
	raw := firstRaw(m, keys...)
	if len(raw) == 0 {
		return nil
	}
	var out []string
	if json.Unmarshal(raw, &out) == nil {
		return out
	}
	return nil
}

func parseImages(stdout string) []LocalImage {
	trimmed := strings.TrimSpace(stdout)
	if trimmed == "" {
		return []LocalImage{}
	}
	var raw []rawImage
	if strings.HasPrefix(trimmed, "[") {
		if err := json.Unmarshal([]byte(trimmed), &raw); err != nil {
			return []LocalImage{}
		}
	} else {
		for _, line := range strings.Split(trimmed, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			var img rawImage
			if err := json.Unmarshal([]byte(line), &img); err != nil {
				continue
			}
			raw = append(raw, img)
		}
	}
	out := make([]LocalImage, 0, len(raw))
	for _, r := range raw {
		if img, ok := intoLocal(r); ok {
			out = append(out, img)
		}
	}
	return out
}

func intoLocal(raw rawImage) (LocalImage, bool) {
	ref, ok := namedRef(raw.Names)
	if !ok {
		ref, ok = namedRef(raw.RepoTags)
	}
	if !ok {
		ref, ok = composeRef(raw.Repository, raw.Tag)
	}
	if !ok {
		return LocalImage{}, false
	}
	size, bytes := sizeParts(raw.Size)
	return LocalImage{
		ID:         raw.ID,
		Reference:  ref,
		Repository: raw.Repository,
		Tag:        raw.Tag,
		Size:       size,
		SizeBytes:  bytes,
	}, true
}

func namedRef(names []string) (string, bool) {
	for _, name := range names {
		name = strings.TrimSpace(name)
		if !noneish(name) {
			return name, true
		}
	}
	return "", false
}

func composeRef(repository, tag string) (string, bool) {
	if noneish(repository) {
		return "", false
	}
	if noneish(tag) {
		return repository, true
	}
	return repository + ":" + tag, true
}

func noneish(s string) bool {
	return s == "" || s == "<none>"
}

func sizeParts(raw json.RawMessage) (string, *uint64) {
	bytes := parseSizeBytes(raw)
	if bytes != nil {
		return formatBytes(*bytes), bytes
	}
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s, nil
	}
	return "", nil
}

func parseSizeBytes(raw json.RawMessage) *uint64 {
	if len(raw) == 0 {
		return nil
	}
	var n uint64
	if json.Unmarshal(raw, &n) == nil {
		return &n
	}
	var s string
	if json.Unmarshal(raw, &s) != nil {
		return nil
	}
	return parseHumanSize(s)
}

func parseHumanSize(raw string) *uint64 {
	compact := strings.ReplaceAll(strings.TrimSpace(raw), " ", "")
	if compact == "" {
		return nil
	}
	split := -1
	for i, c := range compact {
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') {
			split = i
			break
		}
	}
	if split < 0 {
		return nil
	}
	n, err := strconv.ParseFloat(compact[:split], 64)
	if err != nil {
		return nil
	}
	var mul float64
	switch strings.ToUpper(compact[split:]) {
	case "B":
		mul = 1
	case "KB", "KIB", "K":
		mul = 1024
	case "MB", "MIB", "M":
		mul = 1024 * 1024
	case "GB", "GIB", "G":
		mul = 1024 * 1024 * 1024
	case "TB", "TIB", "T":
		mul = 1024 * 1024 * 1024 * 1024
	default:
		return nil
	}
	v := uint64(math.Round(n * mul))
	return &v
}

func formatBytes(n uint64) string {
	const (
		kb = 1024.0
		mb = 1024.0 * kb
		gb = 1024.0 * mb
	)
	f := float64(n)
	switch {
	case f >= gb:
		return fmt.Sprintf("%.1f GB", f/gb)
	case f >= mb:
		return fmt.Sprintf("%.1f MB", f/mb)
	case f >= kb:
		return fmt.Sprintf("%.1f KB", f/kb)
	default:
		return fmt.Sprintf("%.0f B", f)
	}
}
