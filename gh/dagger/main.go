package main

import (
	"context"
	"strings"
)

type Gh struct {
	// The base branch of the repository (ex: main, master)
	// +private
	BaseBranch string
	// The token to authenticate with GitHub
	// +private
	Token *Secret
	// RepoDir of the GitHub repo. Usually the root directory of the workdir.
	// +private
	RepoPath *Directory
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
	// RepoDir of the GitHub repo. Usually the root directory of the workdir.
	// +required
	repoPath *Directory,
) *Gh {
	return &Gh{
		BaseBranch: baseBranch,
		Token:      token,
		RepoPath:   repoPath,
	}
}

// RunGit runs a command using the git CLI.
//
// Example usage: dagger call --token=env:TOKEN --cmd="status' --base=main --head=test-daggerverse" --version="2.43.0"
func (m *Gh) RunGit(
	ctx context.Context,
	// command to run
	cmd string,
	// version of the Github CLI
	// +optional
	// +default="2.43.0"
	version string,
) (string, error) {
	return dag.Container().
		From("alpine/git:"+version).
		WithDirectory("/workspace", m.RepoPath, ContainerWithDirectoryOpts{}).
		WithSecretVariable("GITHUB_TOKEN", m.Token).
		WithWorkdir("/workspace").
		WithExec(
			[]string{"sh", "-c", strings.Join([]string{"git", cmd}, " ")},
			ContainerWithExecOpts{SkipEntrypoint: true},
		).
		Stdout(ctx)
}

// RunGh runs a command using the git CLI.
//
// Example usage: dagger call --token=env:TOKEN --cmd="status' --base=main --head=test-daggerverse" --version="2.43.0"
func (m *Gh) RunGh(
	ctx context.Context,
	// command to run
	cmd string,
	// version of the Github CLI
	// +optional
	// +default="2.47.0"
	version string,
) (string, error) {
	return dag.Container().
		From("maniator/gh:v"+version).
		WithDirectory("/workspace", m.RepoPath, ContainerWithDirectoryOpts{}).
		WithSecretVariable("GITHUB_TOKEN", m.Token).
		WithWorkdir("/workspace").
		WithExec(
			[]string{"sh", "-c", strings.Join([]string{"gh", cmd}, " ")},
			ContainerWithExecOpts{SkipEntrypoint: true},
		).
		Stdout(ctx)
}
