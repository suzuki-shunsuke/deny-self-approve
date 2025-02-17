package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/deny-self-approve/pkg/github"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

// Run denys self-approved GitHub Pull Requests.
// It gets pull request reviews and committers via GitHub GraphQL API, and checks if people other than committers approve the PR.
// If Dismiss is true, it dismisses the approval of committers.
// If the PR isn't approved by people other than committers, it returns an error.
func (c *Controller) Run(ctx context.Context, logE *logrus.Entry, input *Input) error { //nolint:cyclop
	// Get a pull request reviews and committers via GraphQL API
	pr, err := c.gh.GetPR(ctx, input.RepoOwner, input.RepoName, input.PR)
	if err != nil {
		return fmt.Errorf("get a pull request: %w", err)
	}
	encoder := json.NewEncoder(c.stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(&Result{
		PullRequest: &PullRequest{
			Repo:       fmt.Sprintf("%s/%s", input.RepoOwner, input.RepoName),
			Number:     input.PR,
			HeadRefOID: pr.HeadRefOID,
			Reviews:    pr.Reviews,
			Commits:    pr.Commits,
		},
		Options: &ResultOptions{
			Dismiss: input.Dismiss,
		},
	}); err != nil {
		return fmt.Errorf("encode the pull request: %w", err)
	}
	// Convert []*PullRequestCommit to []*Commit
	commits := make([]*github.Commit, len(pr.Commits.Nodes))
	for i, commit := range pr.Commits.Nodes {
		commits[i] = commit.Commit
	}
	committers, err := getCommitters(commits)
	if err != nil {
		if !input.IgnoreUnknownCommit {
			return err
		}
		logerr.WithError(logE, err).Warn("some commits don't have committer and author")
	}
	// Checks if people other than committers approve the PR
	selfApprovals, ok := check(pr.HeadRefOID, pr.Reviews.Nodes, committers)
	if input.Dismiss {
		// If Dismiss is true, dismiss the approval of committers
		for _, selfApproval := range selfApprovals {
			if err := c.gh.Dismiss(ctx, selfApproval.ID); err != nil {
				logerr.WithError(logE, err).Error("dismiss a self-approval")
			}
			logE.WithFields(logrus.Fields{
				"review_id": selfApproval.ID,
				"approver":  selfApproval.Login,
			}).Info("dismiss a self-approval")
		}
	}
	if input.Command == "validate" && !ok {
		return errors.New("pull requests must be approved by people who don't push commits to them")
	}
	return nil
}

type Result struct {
	Options     *ResultOptions `json:"options"`
	PullRequest *PullRequest   `json:"pull_request"`
}

type ResultOptions struct {
	Dismiss bool `json:"dismiss"`
}

type PullRequest struct {
	Repo       string          `json:"repo"`
	Number     int             `json:"number"`
	HeadRefOID string          `json:"headRefOid"`
	Reviews    *github.Reviews `json:"reviews" graphql:"reviews(first:30)"`
	Commits    *github.Commits `json:"commits" graphql:"commits(first:30)"`
}

type Approval struct {
	Login string
	ID    string
}

func getCommitters(commits []*github.Commit) (map[string]struct{}, error) {
	committers := make(map[string]struct{}, len(commits))
	failed := false
	for _, commit := range commits {
		user, err := commit.Login()
		if err != nil {
			failed = true
			continue
		}
		committers[user] = struct{}{}
	}
	if failed {
		return committers, errors.New("both commiter and author are null")
	}
	return committers, nil
}

// check checks if committers approve the pull request themselves.
// This function returns a list of committers doing self-approval and a boolean if others approve the pull request.
// The second return value is true if others approve the pull request.
func check(headRefOID string, reviews []*github.Review, committers map[string]struct{}) ([]*Approval, bool) {
	selfApprovals := []*Approval{}
	nonSelfApproved := false
	for _, review := range reviews {
		// TODO check CODEOWNERS
		if review.State != "APPROVED" || review.Commit.OID != headRefOID {
			// Ignore reviews other than APPROVED
			// Ignore reviews for non head commits
			continue
		}
		if strings.HasPrefix(review.Author.ResourcePath, "/apps/") || strings.HasSuffix(review.Author.Login, "[bot]") {
			// Ignore approvals from bots
			continue
		}
		if _, ok := committers[review.Author.Login]; ok {
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
