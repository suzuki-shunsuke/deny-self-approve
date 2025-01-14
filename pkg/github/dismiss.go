package github

import (
	"context"
	"fmt"

	"github.com/shurcooL/githubv4"
)

// Dismiss dismiss a review via GitHub GraphQL API.
func (c *Client) Dismiss(ctx context.Context, reviewID string) error {
	var m struct{}

	input := githubv4.DismissPullRequestReviewInput{
		PullRequestReviewID: reviewID,
		Message:             "Dismiss a self-approval",
	}
	if err := c.v4Client.Mutate(ctx, &m, input, nil); err != nil {
		return fmt.Errorf("dismiss a review: %w", err)
	}
	return nil
}
