package project

import (
	"fmt"
	"strings"
)

type ProjectKind string

const (
	KindKubernetesClusterReview ProjectKind = "kubernetesClusterReview"
	KindContainerImageReview    ProjectKind = "containerImageReview"
)

type ProjectConfig struct {
	Context *string  `json:"context"`
	Image   *string  `json:"image,omitempty"`
	Images  []string `json:"images"`
}

type Project struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	Kind      ProjectKind   `json:"kind"`
	Config    ProjectConfig `json:"config"`
	CreatedAt string        `json:"createdAt"`
	UpdatedAt string        `json:"updatedAt"`
}

type ProjectInput struct {
	Name   string        `json:"name"`
	Kind   ProjectKind   `json:"kind"`
	Config ProjectConfig `json:"config"`
}

func ParseKind(s string) (ProjectKind, error) {
	switch ProjectKind(s) {
	case KindKubernetesClusterReview, KindContainerImageReview:
		return ProjectKind(s), nil
	default:
		return "", fmt.Errorf("unknown project kind '%s' in db", s)
	}
}

func CollectImages(config ProjectConfig) []string {
	var images []string
	seen := make(map[string]struct{})
	add := func(raw string) {
		image := strings.TrimSpace(raw)
		if image == "" {
			return
		}
		if _, ok := seen[image]; ok {
			return
		}
		seen[image] = struct{}{}
		images = append(images, image)
	}
	for _, image := range config.Images {
		add(image)
	}
	if config.Image != nil {
		add(*config.Image)
	}
	if images == nil {
		return []string{}
	}
	return images
}

func NormalizeForKind(kind ProjectKind, config ProjectConfig) (ProjectConfig, error) {
	switch kind {
	case KindKubernetesClusterReview:
		if config.Context != nil {
			trimmed := strings.TrimSpace(*config.Context)
			if trimmed == "" {
				return ProjectConfig{}, fmt.Errorf("context must not be empty when provided")
			}
			config.Context = &trimmed
		}
		return ProjectConfig{
			Context: config.Context,
			Images:  []string{},
		}, nil
	case KindContainerImageReview:
		images := CollectImages(config)
		if len(images) == 0 {
			return ProjectConfig{}, fmt.Errorf("select at least one image")
		}
		return ProjectConfig{
			Images: images,
		}, nil
	default:
		return ProjectConfig{}, fmt.Errorf("unknown project kind '%s'", kind)
	}
}

func validateInput(input ProjectInput) (string, ProjectConfig, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return "", ProjectConfig{}, fmt.Errorf("name must not be empty")
	}
	config, err := NormalizeForKind(input.Kind, input.Config)
	if err != nil {
		return "", ProjectConfig{}, err
	}
	return name, config, nil
}
