package main

import (
	"context"
	"fmt"
	"strings"
)

type Gh struct {
	// The base branch of the repository (ex: main, master)
	// +private
	BaseBranch string
	// The token to authenticate with GitHub
	// +private
	Token *Secret
}

// New creates a new GitHub module with the provided inputs
func New(
	// The base branch of the repository (ex: main, master)
	// +optional
	// +default="master"
	baseBranch string,
	// The token to authenticate with GitHub
	// +required
	token *Secret,
) *Gh {
	return &Gh{
		BaseBranch: baseBranch,
		Token:      token,
	}
}

// RunGit runs a command using the git CLI.
//
// Example usage: dagger call --token=env:TOKEN --base-branch=main run-git --cmd="status" --repo-path="/workspace/repo"
func (m *Gh) RunGit(
	ctx context.Context,
	// RepoDir of the GitHub repo
	// +required
	repoPath *Directory,
	// command to run
	// +required
	cmd string,
	// version of the Github CLI
	// +optional
	// +default="2.43.0"
	version string,
) (string, error) {
	c := dag.Container().
		From("alpine/git:"+version).
		WithDirectory("/workspace", repoPath, ContainerWithDirectoryOpts{}).
		WithSecretVariable("GITHUB_TOKEN", m.Token).
		WithWorkdir("/workspace").
		WithExec(
			[]string{"sh", "-c", strings.Join([]string{"git", cmd}, " ")},
			ContainerWithExecOpts{SkipEntrypoint: true},
		)
	//.Sync(ctx)
	//if err != nil {
	//	return "", fmt.Errorf("failed to run git command: %w", err)
	//}

	return c.Stdout(ctx)
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
