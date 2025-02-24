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
// If the PR isn't approved by people other than committers, it returns an error.
func (c *Controller) Run(ctx context.Context, _ *logrus.Entry, input *Input) error {
	// Get a pull request reviews and committers via GraphQL API
	pr, err := c.gh.GetPR(ctx, input.RepoOwner, input.RepoName, input.PR)
	if err != nil {
		return fmt.Errorf("get a pull request: %w", err)
	}
	if err := c.output(input, pr); err != nil {
		return err
	}
	reviews := filterReviews(pr.Reviews.Nodes, pr.HeadRefOID)

	if len(reviews) > 1 {
		// Allow multiple approvals
		return nil
	}

	if len(reviews) == 0 {
		// Approval is required
		return errApproval
	}

	committers, err := getCommitters(convertCommits(pr.Commits.Nodes))
	if err != nil {
		return err
	}
	// Checks if people other than committers approve the PR
	return validate(reviews, committers)
}

// convertCommits converts []*PullRequestCommit to []*Commit
func convertCommits(commits []*github.PullRequestCommit) []*github.Commit {
	arr := make([]*github.Commit, len(commits))
	for i, commit := range commits {
		arr[i] = commit.Commit
	}
	return arr
}

func (c *Controller) output(input *Input, pr *github.PullRequest) error {
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
	}); err != nil {
		return fmt.Errorf("encode the pull request: %w", err)
	}
	return nil
}

type Result struct {
	PullRequest *PullRequest `json:"pull_request"`
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
	for _, commit := range commits {
		login := commit.Login()
		if login == "" {
			return committers, logerr.WithFields(errors.New("commit isn't linked to a GitHub User"), logrus.Fields{ //nolint:wrapcheck
				"docs": "https://github.com/suzuki-shunsuke/deny-self-approve/tree/main/docs/001.md",
			})
		}
		committers[login] = struct{}{}
	}
	return committers, nil
}

func filterReviews(reviews []*github.Review, headRefOID string) []*github.Review {
	arr := make([]*github.Review, 0, len(reviews))
	for _, review := range reviews {
		if review.State != "APPROVED" || review.Commit.OID != headRefOID {
			// Ignore reviews other than APPROVED
			// Ignore reviews for non head commits
			continue
		}
		if strings.HasPrefix(review.Author.ResourcePath, "/apps/") || strings.HasSuffix(review.Author.Login, "[bot]") {
			// Ignore approvals from bots
			continue
		}
		arr = append(arr, review)
	}
	return arr
}

var errApproval = errors.New("pull requests must be approved by people who don't push commits to them")

// validate validates if committers approve the pull request themselves.
func validate(reviews []*github.Review, committers map[string]struct{}) error {
	for _, review := range reviews {
		// TODO check CODEOWNERS
		if _, ok := committers[review.Author.Login]; ok {
			// self-approve
			continue
		}
		// Someone other than committers approved the PR, so this PR is not self-approved.
		return nil
	}
	return errApproval
}
