package controller

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/deny-self-approve/pkg/github"
)

// Run denys self-approved GitHub Pull Requests.
func (c *Controller) Run(ctx context.Context, logE *logrus.Entry, input *Input) error {
	// Get a pull request reviews and commiters via GraphQL API
	pr, err := c.gh.GetPR(ctx, input.RepoOwner, input.RepoName, input.PR)
	if err != nil {
		return fmt.Errorf("get a pull request: %w", err)
	}
	// Checks if people other than commiters approve the PR
	selfApprovers, ok := check(pr.Reviews.Nodes, pr.Commits.Nodes)
	// If Dismiss is true, dismiss the approval of commiters
	if input.Dismiss {
		for selfApprover := range selfApprovers {
		}
	}
	return nil
}

// check checks if commiters approve the pull request themselves.
// This function returns a list of commiters doing self-approval and a boolean if others approve the pull request.
// The second return value is true if others approve the pull request.
func check(reviews []*github.Review, commits []*github.Commit) (map[string]struct{}, bool) {
	commiters := map[string]struct{}{}
	for _, commit := range commits {
		commiters[commit.Commiter.User.Login] = struct{}{}
	}
	selfApprovers := map[string]struct{}{}
	nonSelfApproved := false
	for _, review := range reviews {
		if review.State != "APPROVED" {
			// Ignore reviews other than APPROVED
			continue
		}
		if _, ok := commiters[review.Author.Login]; ok {
			// self-approve
			selfApprovers[review.Author.Login] = struct{}{}
			continue
		}
		// Someone other than the committer approved the PR, so this PR is not self-approved.
		nonSelfApproved = true
	}
	return selfApprovers, nonSelfApproved
}
