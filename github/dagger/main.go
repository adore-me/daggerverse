// This module handles the GitHub API interactions.
//
// The CreatePullRequest function creates a new pull request in the provided repository.
package main

import "github.com/google/go-github/github"

type GitHub struct {
	// +private
	Client *github.Client
}

func test() {
	gh := &GitHub{}
	gh.Client = github.NewClient(nil)
}

// New creates a new GitHub module with the provided token
//
// Example usage: dagger call --token=env:TOKEN create-pull-request
//func New(
//	token *Secret,
//) (*GitHub, error) {
//	gh := &GitHub{}
//
//	// Init GitHub client
//	ctx := context.Background()
//	t, err := token.Plaintext(ctx)
//	if err != nil {
//		return gh, fmt.Errorf("failed to get token: %v", err)
//	}
//	gh.Client = github.NewClient(nil).WithAuthToken(t)
//
//	return gh, nil
//}

//func (m *GitHub) CreatePullRequest(
//	// +default="New PR"
//	title,
//	// +optional
//	// +default="main"
//	baseBranch,
//	// +optional
//	// +default="upgrade-istio"
//	newBranch,
//	// +required
//	owner,
//	// +required
//	repo string,
//) error {
//	ctx := context.Background()
//
//	// Get the reference of the branch you want to create from
//	ref, _, err := m.Client.Git.GetRef(ctx, owner, repo, "refs/heads/"+baseBranch)
//	if err != nil {
//		return fmt.Errorf("failed to get ref: %v", err)
//	}
//
//	// Create a new reference for the new branch
//	newRef := &github.Reference{
//		Ref: github.String("refs/heads/" + newBranch),
//		Object: &github.GitObject{
//			SHA: ref.Object.SHA,
//		},
//	}
//	_, _, err = m.Client.Git.CreateRef(ctx, owner, repo, newRef)
//	if err != nil {
//		return fmt.Errorf("failed to create ref: %v", err)
//	}
//
//	return nil
//}
