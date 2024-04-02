package main

import (
	"context"
	"fmt"
	"github.com/google/go-github/v60/github"
)

type GitHub struct {
	// +private
	Client *github.Client
}

func New(
	// +optional
	token *Secret,
) (*GitHub, error) {
	gh := &GitHub{}

	// Init GitHub client
	ctx := context.Background()
	t, err := token.Plaintext(ctx)
	if err != nil {
		return gh, fmt.Errorf("failed to get token: %v", err)
	}
	gh.Client = github.NewClient(nil).WithAuthToken(t)

	return gh, nil
}

func (m *GitHub) CreatePullRequest(
	// +default="New PR"
	title string,
	// +optional
	// +default="main"
	baseBranch string,
) {
	ctx := context.Background()

	// Get the reference of the branch you want to create from
	ref, _, err := m.Client.Git.GetRef(ctx, "owner", "repo", "refs/heads/main")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Create a new reference for the new branch
	newRef := &github.Reference{
		Ref: github.String("refs/heads/new-branch"),
		Object: &github.GitObject{
			SHA: ref.Object.SHA,
		},
	}
	_, _, err = m.Client.Git.CreateRef(ctx, "owner", "repo", newRef)
	if err != nil {
		fmt.Println(err)
		return
	}
}
