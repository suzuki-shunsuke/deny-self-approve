package controller

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/deny-self-approve/pkg/github"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

// Run denys self-approved GitHub Pull Requests.
// It gets pull request reviews and commiters via GitHub GraphQL API, and checks if people other than commiters approve the PR.
// If Dismiss is true, it dismisses the approval of commiters.
// If the PR isn't approved by people other than commiters, it returns an error.
func (c *Controller) Run(ctx context.Context, logE *logrus.Entry, input *Input) error {
	// Get a pull request reviews and commiters via GraphQL API
	pr, err := c.gh.GetPR(ctx, input.RepoOwner, input.RepoName, input.PR)
	if err != nil {
		return fmt.Errorf("get a pull request: %w", err)
	}
	// Checks if people other than commiters approve the PR
	commits := make([]*github.Commit, len(pr.Commits.Nodes))
	for i, commit := range pr.Commits.Nodes {
		commits[i] = commit.Commit
	}
	selfApprovals, ok := check(pr.HeadRefOID, pr.Reviews.Nodes, commits)
	if input.Dismiss {
		// If Dismiss is true, dismiss the approval of commiters
		for _, selfApproval := range selfApprovals {
			if err := c.gh.Dismiss(ctx, selfApproval.ID); err != nil {
				logerr.WithError(logE, err).Error("dismiss a self-approval")
			}
		}
	}
	if !ok {
		return errors.New("pull requests must be approved by people who don't push commits to them")
	}
	return nil
}

type Approval struct {
	Login string
	ID    string
}

// check checks if commiters approve the pull request themselves.
// This function returns a list of commiters doing self-approval and a boolean if others approve the pull request.
// The second return value is true if others approve the pull request.
func check(headRefOID string, reviews []*github.Review, commits []*github.Commit) ([]*Approval, bool) {
	commiters := map[string]struct{}{}
	for _, commit := range commits {
		commiters[commit.Login()] = struct{}{}
	}
	selfApprovals := []*Approval{}
	nonSelfApproved := false
	for _, review := range reviews {
		// TODO check CODEOWNERS
		if review.State != "APPROVED" || review.Commit.OID != headRefOID {
			// Ignore reviews other than APPROVED
			// Ignore reviews for non head commits
			continue
		}
		if strings.HasSuffix(review.Author.Login, "[bot]") {
			// Ignore approvals from bots
			continue
		}
		if _, ok := commiters[review.Author.Login]; ok {
			// self-approve
			selfApprovals = append(selfApprovals, &Approval{
				ID:    review.ID,
				Login: review.Author.Login,
			})
			continue
		}
		// Someone other than committers approved the PR, so this PR is not self-approved.
		// TODO dismiss approvals from bots
		nonSelfApproved = true
	}
	return selfApprovals, nonSelfApproved
}
