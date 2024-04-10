package main

import (
	"context"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/google/go-github/v61/github"
)

type Gh struct {
	// The owner of the repository (ex: adore-me)
	// +private
	Owner string
	// The name of the repository (ex: daggerverse)
	// +private
	Repo string
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
	// The owner of the repository (ex: adore-me)
	// +required
	owner,
	// The name of the repository (ex: daggerverse)
	// +required
	repo,
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
		Owner:      owner,
		Repo:       repo,
		BaseBranch: baseBranch,
		Token:      token,
		RepoPath:   repoPath,
	}
}

// CreateLocalBranch creates a new branch in the provided repository.
//
// Example usage: dagger call --token=env:TOKEN --owner=adore-me --repo=daggerverse create-new-branch --branch=test-daggerverse --repo-path=/path/to/repo
func (m *Gh) CreateLocalBranch(
	// The new branch name to create the pull request from
	// +required
	branchName string,
) (string, error) {
	ctx := context.Background()
	_, err := m.RepoPath.Export(ctx, "./"+m.Repo)
	if err != nil {
		return "", fmt.Errorf("failed to export repository: %v", err)
	}

	// Open the existing checked-out repository
	gitRepo, err := git.PlainOpen("./" + m.Repo)
	if err != nil {
		return "", fmt.Errorf("failed to open repository: %v", err)
	}

	// Switch to the desired branch
	// Check if the branch already exists
	_, err = gitRepo.Reference(plumbing.NewBranchReferenceName(m.BaseBranch), true)
	if err != nil {
		return "", fmt.Errorf("failed to get reference: %v", err)
	}

	// Checkout the base branch
	wt, err := gitRepo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %v", err)
	}
	err = wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(m.BaseBranch),
	})

	// Create a new branch from the base branch
	headRef, err := gitRepo.Head()
	// Checkout specific branch
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %v", err)
	}

	ref := plumbing.NewHashReference(plumbing.NewBranchReferenceName(branchName), headRef.Hash())
	err = gitRepo.Storer.SetReference(ref)
	if err != nil {
		return "", fmt.Errorf("failed to set reference: %v", err)
	}

	return fmt.Sprintf("created local branch %s", branchName), nil
}

func (m *Gh) CommitChangesInNewBranch(
	// Title of the commit
	// +optional
	// +default="Update file"
	commitTitle,
	// The new branch name to commit changes to
	// +required
	branchName string,
) (string, error) {
	ctx := context.Background()
	_, err := m.RepoPath.Export(ctx, "./"+m.Repo)
	if err != nil {
		return "", fmt.Errorf("failed to export repository: %v", err)
	}

	// Open the existing checked-out repository
	gitRepo, err := git.PlainOpen("./" + m.Repo)
	if err != nil {
		return "", fmt.Errorf("failed to open repository: %v", err)
	}

	// Create a new branchName
	headRef, err := gitRepo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %v", err)
	}

	ref := plumbing.NewHashReference(plumbing.NewBranchReferenceName(branchName), headRef.Hash())
	err = gitRepo.Storer.SetReference(ref)
	if err != nil {
		return "", fmt.Errorf("failed to set reference: %v", err)
	}

	// Create a new worktree for the repository
	wt, err := gitRepo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %v", err)
	}

	// Add all changes to the worktree
	_, err = wt.Add(".")
	if err != nil {
		return "", fmt.Errorf("failed to add changes: %v", err)
	}

	// Commit the changes
	cOpts := &git.CommitOptions{
		All: true,
		Author: &object.Signature{
			Name:  "Dagger",
			Email: "mihai.g@adoreme.com",
		},
	}
	_, err = wt.Commit(commitTitle, cOpts)
	if err != nil {
		return "", fmt.Errorf("failed to commit changes: %v", err)
	}

	return "successfully committed changes", nil
}

// GetFileFromRemote retrieves the content of a file from remote repository and returns the content base64 encoded.
//
// Example usage: dagger call --token=env:TOKEN --owner=adore-me --repo=daggerverse update-file --branch=main --file-path=README.md
func (m *Gh) GetFileFromRemote(
	// The branch to update the file in
	// +required
	branch,
	// The path to the file you want to update
	// +required
	filePath string,
) (content string, err error) {
	ctx := context.Background()
	token, err := m.Token.Plaintext(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get token: %v", err)
	}

	client := github.NewClient(nil).WithAuthToken(token)

	// Get the reference of the branch you want to create from
	ref, _, err := client.Git.GetRef(ctx, m.Owner, m.Repo, "refs/heads/"+branch)
	if err != nil {
		return "", fmt.Errorf("failed to get ref: %v", err)
	}

	// Get the contents of the file you want to edit
	fileContent, _, _, err := client.Repositories.GetContents(ctx, m.Owner, m.Repo, filePath, &github.RepositoryContentGetOptions{
		Ref: *ref.Ref,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get contents: %v", err)
	}

	// Check if fileContent is nil before trying to access its Content property
	if fileContent == nil {
		return "", fmt.Errorf("fileContent is nil")
	}

	return *fileContent.Content, nil
}

// CreatePullRequest creates a new pull request in the provided repository.
//
// Example usage: dagger call --token=env:TOKEN --owner=adore-me --repo=daggerverse create-pull-request --title="New PR" --new-branch=test-daggerverse
func (m *Gh) CreatePullRequest(
	// The title of the pull request
	// +required
	title,
	// The new branch name to create the pull request from
	// +required
	newBranch,
	// The base branch to create the pull request against
	// +optional
	baseBranch string,
) (string, error) {
	ctx := context.Background()
	token, err := m.Token.Plaintext(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get token: %v", err)
	}

	client := github.NewClient(nil).WithAuthToken(token)

	if baseBranch == "" {
		baseBranch = m.BaseBranch
	}
	// Create a pull request
	newPR := &github.NewPullRequest{
		Title: github.String(title),
		Head:  github.String(newBranch),
		Base:  github.String(baseBranch),
	}
	_, rsp, err := client.PullRequests.Create(ctx, m.Owner, m.Repo, newPR)
	if err != nil {
		return "", fmt.Errorf("failed to create pull request with error: %v.\nMore details in response: %v", err, rsp.Body)
	}

	return "successfully created the pull request", nil
}
