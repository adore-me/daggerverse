// This module handles the Istio version management.
package main

import (
	"context"
	"dagger/istio/internal/dagger"
	"encoding/json"
	"fmt"
	"github.com/Masterminds/semver"
	"gopkg.in/yaml.v2"
	"io"
	"net/http"
)

type Istio struct {
	LatestVersion string
	LocalVersion  string
	// +private
	ConfigMap *File
}

// New creates a new Istio module with the provided ConfigMap file and Directory
//
// Example usage: dagger call --cm-path=clusters/dev/istio-version.yaml --dir=. is-new-version
func New(
	// ConfigMap (that stores istio current version) file path. Should be relative to the dir parameter.
	// +required
	ConfigMap *File,
) *Istio {
	i := &Istio{}
	i.ConfigMap = ConfigMap
	if err := i.setLocalVersion(); err != nil {
		panic(err)
	}
	if err := i.setLatestVersion(); err != nil {
		panic(err)
	}

	return i
}

type Release struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
}

type IstioVersionCm struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name        string `yaml:"name"`
		Namespace   string `yaml:"namespace"`
		Annotations struct {
			KustomizeToolkitFluxcdIoSsa string `yaml:"kustomize.toolkit.fluxcd.io/ssa"`
		} `yaml:"annotations"`
	} `yaml:"metadata"`
	Data struct {
		Version string `yaml:"version"`
	} `yaml:"data"`
}

// setLatestVersion Get the latest Istio version from GitHub
func (m *Istio) setLatestVersion() error {
	owner := "istio" // Replace with the repository owner's username
	repo := "istio"  // Replace with the repository name
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to get latest version: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var release Release
	if err := json.Unmarshal(body, &release); err != nil {
		return fmt.Errorf("failed to unmarshal json: %w", err)
	}

	m.LatestVersion = release.TagName

	return nil
}

// setLocalVersion Get the local Istio version from the provided ConfigMap file
func (m *Istio) setLocalVersion() error {
	ctx := context.Background()

	content, err := m.ConfigMap.Contents(ctx)
	if err != nil {
		return fmt.Errorf("failed to read file contents: %w", err)
	}

	iVersion := &IstioVersionCm{}
	if err := yaml.Unmarshal([]byte(content), iVersion); err != nil {
		return fmt.Errorf("failed to unmarshal yaml: %w", err)
	}

	m.LocalVersion = iVersion.Data.Version

	return nil
}

// IsNewerVersion Check if the latest Istio version is newer than the local version
//
// Example usage: dagger call --config-map=clusters/dev/istio-version.yaml is-new-version
func (m *Istio) IsNewerVersion() (bool, error) {
	latestVersion, err := semver.NewVersion(m.LatestVersion)
	if err != nil {
		return false, fmt.Errorf("failed to parse latest version: %w", err)
	}

	localVersion, err := semver.NewVersion(m.LocalVersion)
	if err != nil {
		return false, fmt.Errorf("failed to parse local version: %w", err)
	}

	result := latestVersion.Compare(localVersion)

	if result > 0 {
		return true, nil
	}

	return false, nil
}

// ReturnUpdatedCm Update the version in the ConfigMap file
//
// Example usage: dagger call --config-map=./clusters/dev/istio-version.yaml update-version-cm
func (m *Istio) ReturnUpdatedCm() (string, error) {
	isNewerVersion, err := m.IsNewerVersion()
	if err != nil {
		return "", fmt.Errorf("failed to check if newer version: %w", err)
	}
	if !isNewerVersion {
		return fmt.Sprintf("No update needed. Latest version is %s", m.LatestVersion), nil
	}

	ctx := context.Background()
	content, err := m.ConfigMap.Contents(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to read file contents: %w", err)
	}

	iVersion := &IstioVersionCm{}
	if err := yaml.Unmarshal([]byte(content), iVersion); err != nil {
		return "", fmt.Errorf("failed to unmarshal yaml: %w", err)
	}

	iVersion.Data.Version = m.LatestVersion

	newContent, err := yaml.Marshal(iVersion)
	if err != nil {
		return "", fmt.Errorf("failed to marshal yaml: %w", err)
	}

	m.ConfigMap = m.ConfigMap.With(func(r *dagger.File) *dagger.File {
		nc := string(newContent)
		r.contents = &nc
		return r
	})

	return string(newContent), nil
}
