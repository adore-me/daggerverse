package main

import (
	"context"
	"fmt"
	"gopkg.in/ini.v1"
	"strings"
)

type Gh struct {
	// RepoDir of the GitHub repo
	// +private
	RepoPath *Directory
	// The base branch of the repository (ex: main, master)
	// +private
	BaseBranch string
	// The token to authenticate with GitHub
	// +private
	Token *Secret
}

// New creates a new GitHub module with the provided inputs
func New(
	// RepoDir of the GitHub repo
	// +required
	repoPath *Directory,
	// The base branch of the repository (ex: main, master)
	// +optional
	// +default="master"
	baseBranch string,
	// The token to authenticate with GitHub
	// +required
	token *Secret,
) *Gh {
	return &Gh{
		RepoPath:   repoPath,
		BaseBranch: baseBranch,
		Token:      token,
	}
}

// RunGit runs a command using the git CLI.
//
// Example usage: dagger call --token=env:TOKEN --repo-path="/workspace/repo" run-git --cmd=status
func (m *Gh) RunGit(
	ctx context.Context,
	// command to run
	// +required
	cmd string,
	// version of the Github CLI
	// +optional
	// +default="2.43.0"
	version string,
	// user email
	// +optional
	// +default="action@github.com"
	userEmail string,
	// user name
	// +optional
	// +default="GitHub Action"
	userName string,
) (*Container, error) {
	tk, err := m.Token.Plaintext(ctx)
	if err != nil {
		return &Container{}, fmt.Errorf("failed to get auth token: %w", err)
	}

	owner, repo, err := m.extractRepoOwnerAndName(ctx)
	if err != nil {
		return &Container{}, fmt.Errorf("failed to extract repo owner and name: %w", err)
	}

	c, err := dag.Container().
		From("alpine/git:"+version).
		WithDirectory("/workspace", m.RepoPath, ContainerWithDirectoryOpts{}).
		WithSecretVariable("GITHUB_TOKEN", m.Token).
		WithWorkdir("/workspace").
		WithExec(
			[]string{"sh", "-c", strings.Join([]string{
				"git", "config --global user.email '" + userEmail + "'",
			}, " ")},
			ContainerWithExecOpts{SkipEntrypoint: true},
		).
		WithExec(
			[]string{"sh", "-c", strings.Join([]string{
				"git", "config --global user.name '" + userName + "'",
			}, " ")},
			ContainerWithExecOpts{SkipEntrypoint: true},
		).
		WithExec(
			[]string{"sh", "-c", strings.Join([]string{
				"git", "remote set-url origin https://" + tk + ":x-oauth-basic@github.com/" + owner + "/" + repo + ".git",
			}, " ")},
			ContainerWithExecOpts{SkipEntrypoint: true},
		).
		WithExec(
			[]string{"sh", "-c", strings.Join([]string{"git", cmd}, " ")},
			ContainerWithExecOpts{SkipEntrypoint: true},
		).Sync(ctx)
	if err != nil {
		return &Container{}, fmt.Errorf("failed to run git command: %w", err)
	}

	return c, nil
}

func (m *Gh) extractRepoOwnerAndName(ctx context.Context) (owner string, repo string, err error) {
	if _, err := m.RepoPath.File(".git/config").Export(ctx, "/workspace/git-config"); err != nil {
		return "", "", fmt.Errorf("failed to export git config: %w", err)
	}

	// Load the .git/config file
	cfg, err := ini.Load("/workspace/git-config")
	if err != nil {
		return "", "", fmt.Errorf("failed to load git config: %w", err)
	}

	url := ""
	for _, section := range cfg.Sections() {
		if section.HasKey("url") && section.Name() != "DEFAULT" {
			url = section.Key("url").String()
		}
	}

	// Check if the URL is an ssh or https URL
	if strings.HasPrefix(url, "git@") {
		owner, repo = extractRepoOwnerAndNameSSH(url)
	} else {
		owner, repo = extractRepoOwnerAndNameHTTPS(url)
	}

	return owner, repo, nil
}

// RunGh runs a command using the git CLI.
//
// Example usage: dagger call --token=env:TOKEN --base-branch=main run-gh --cmd="status" --repo-path="/workspace/repo"
func (m *Gh) RunGh(
	ctx context.Context,
	// RepoDir of the GitHub repo
	// +required
	repoPath *Directory,
	// command to run
	// +required
	cmd string,
	// version of the Github CLI
	// +optional
	// +default="2.47.0"
	version string,
) (string, error) {
	c, err := dag.Container().
		From("maniator/gh:v"+version).
		WithDirectory("/workspace", repoPath, ContainerWithDirectoryOpts{}).
		WithSecretVariable("GITHUB_TOKEN", m.Token).
		WithWorkdir("/workspace").
		WithExec(
			[]string{"sh", "-c", strings.Join([]string{"gh", cmd}, " ")},
			ContainerWithExecOpts{SkipEntrypoint: true},
		).Sync(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to run git command: %w", err)
	}

	return c.Stdout(ctx)
}

func extractRepoOwnerAndNameSSH(url string) (string, string) {
	// Remove the .git extension
	url = strings.TrimSuffix(url, ".git")
	// Split the URL by the colon
	parts := strings.Split(url, ":")
	// Split the second part by the slash
	parts = strings.Split(parts[1], "/")
	// Return the owner and name
	return parts[0], parts[1]
}

func extractRepoOwnerAndNameHTTPS(url string) (string, string) {
	// Remove the .git extension
	url = strings.TrimSuffix(url, ".git")
	// Split the URL by the slash
	parts := strings.Split(url, "/")
	// Return the owner and name
	return parts[len(parts)-2], parts[len(parts)-1]
}
