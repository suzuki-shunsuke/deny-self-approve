package controller

import (
	"errors"
	"testing"

	"github.com/suzuki-shunsuke/deny-self-approve/pkg/github"
)

func Test_validatePR(t *testing.T) { //nolint:funlen,maintidx
	t.Parallel()
	normalUser := &github.User{
		Login: "suzuki-shunsuke",
	}
	normalUser2 := &github.User{
		Login: "suzuki-shunsuke-2",
	}
	untrustedMachineUser := &github.User{
		Login: "suzuki-shunsuke-3",
	}
	renovate := &github.User{
		Login: "renovate[bot]",
	}
	untrustedBot := &github.User{
		Login: "untrusted[bot]",
	}

	normalCommit := &github.Commit{
		Committer: &github.Committer{
			User: normalUser,
		},
		Author: &github.Committer{
			User: normalUser,
		},
	}
	normalPRCommit := &github.PullRequestCommit{
		Commit: normalCommit,
	}
	appPRCommit := &github.PullRequestCommit{
		Commit: &github.Commit{
			Author: &github.Committer{
				User: untrustedBot,
			},
		},
	}
	untrustedMachineUserPRCommit := &github.PullRequestCommit{
		Commit: &github.Commit{
			Author: &github.Committer{
				User: untrustedMachineUser,
			},
			Committer: &github.Committer{
				User: untrustedMachineUser,
			},
		},
	}
	notLinkedCommit := &github.Commit{}
	notLinkedPRCommit := &github.PullRequestCommit{
		Commit: notLinkedCommit,
	}

	octocat := &github.User{
		Login: "octocat",
	}
	renovateApprove := &github.User{
		Login: "renovate-approve[bot]",
	}
	const headRefOID = "1234567890abcdef"
	headReviewCommit := &github.ReviewCommit{
		OID: headRefOID,
	}
	oldReviewCommit := &github.ReviewCommit{
		OID: "1234567890ffffff",
	}
	tests := []struct {
		name  string
		input *Input
		pr    *github.PullRequest
		exp   error
	}{
		{
			name:  "normal",
			exp:   nil,
			input: &Input{},
			pr: &github.PullRequest{
				Author:     normalUser,
				HeadRefOID: headRefOID,
				Reviews: &github.Reviews{
					Nodes: []*github.Review{
						{
							Author: octocat,
							State:  "APPROVED",
							Commit: headReviewCommit,
						},
					},
				},
				Commits: &github.Commits{
					Nodes: []*github.PullRequestCommit{
						normalPRCommit,
					},
				},
			},
		},
		{
			// At least one approval is required
			name: "approval is required",
			exp:  errApproval,
			input: &Input{
				UntrustedMachineUsers: map[string]struct{}{
					untrustedBot.Login: {},
				},
			},
			pr: &github.PullRequest{
				Author:     normalUser,
				HeadRefOID: headRefOID,
				Reviews: &github.Reviews{
					Nodes: []*github.Review{
						{
							// ignore reviews other than approval
							Author: octocat,
							State:  "REQUEST_CHANGES",
							Commit: headReviewCommit,
						},
						{
							// ignore reviews of non head commits
							Author: octocat,
							State:  "APPROVED",
							Commit: oldReviewCommit,
						},
						{
							// ignore approvals from bots
							Author: renovateApprove,
							State:  "APPROVED",
							Commit: headReviewCommit,
						},
						{
							// ignore approvals from untrusted machine users
							Author: untrustedBot,
							State:  "APPROVED",
							Commit: headReviewCommit,
						},
					},
				},
				Commits: &github.Commits{
					Nodes: []*github.PullRequestCommit{
						{
							Commit: normalCommit,
						},
					},
				},
			},
		},
		{
			name:  "over one aprovals are okay",
			exp:   nil,
			input: &Input{},
			pr: &github.PullRequest{
				Author:     normalUser,
				HeadRefOID: headRefOID,
				Reviews: &github.Reviews{
					Nodes: []*github.Review{
						{
							Author: octocat,
							State:  "APPROVED",
							Commit: headReviewCommit,
						},
						{
							Author: normalUser2,
							State:  "APPROVED",
							Commit: headReviewCommit,
						},
					},
				},
				Commits: &github.Commits{
					Nodes: []*github.PullRequestCommit{
						notLinkedPRCommit,
					},
				},
			},
		},
		{
			name: "ignore approvals from committers",
			exp:  errApproval,
			input: &Input{
				TrustedApps: map[string]struct{}{
					"renovate[bot]": {},
				},
			},
			pr: &github.PullRequest{
				Author:     renovate,
				HeadRefOID: headRefOID,
				Reviews: &github.Reviews{
					Nodes: []*github.Review{
						// This review is ignored because the reviwer is a committer
						{
							Author: normalUser,
							State:  "APPROVED",
							Commit: headReviewCommit,
						},
					},
				},
				Commits: &github.Commits{
					Nodes: []*github.PullRequestCommit{
						{
							Commit: normalCommit,
						},
						{
							Commit: &github.Commit{
								Committer: &github.Committer{
									User: renovate,
								},
							},
						},
					},
				},
			},
		},
		{
			// Require two approvals for PRs created by apps, excluding trusted apps
			name: errTwoApproval.Error(),
			exp:  errTwoApproval,
			input: &Input{
				TrustedApps: map[string]struct{}{
					"renovate[bot]": {},
				},
			},
			pr: &github.PullRequest{
				Author:     untrustedBot,
				HeadRefOID: headRefOID,
				Reviews: &github.Reviews{
					Nodes: []*github.Review{
						{
							Author: normalUser,
							State:  "APPROVED",
							Commit: headReviewCommit,
						},
					},
				},
				Commits: &github.Commits{
					Nodes: []*github.PullRequestCommit{
						{
							Commit: &github.Commit{
								Committer: &github.Committer{
									User: untrustedBot,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Require two approvals for PRs created by untrusted machine users",
			exp:  errTwoApproval,
			input: &Input{
				UntrustedMachineUsers: map[string]struct{}{
					untrustedMachineUser.Login: {},
				},
			},
			pr: &github.PullRequest{
				Author:     untrustedMachineUser,
				HeadRefOID: headRefOID,
				Reviews: &github.Reviews{
					Nodes: []*github.Review{
						{
							Author: normalUser,
							State:  "APPROVED",
							Commit: headReviewCommit,
						},
					},
				},
				Commits: &github.Commits{
					Nodes: []*github.PullRequestCommit{
						{
							Commit: &github.Commit{
								Committer: &github.Committer{
									User: untrustedMachineUser,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Require two approvals if the pull request has commits not linked to any user",
			exp:  errTwoApproval,
			input: &Input{
				UntrustedMachineUsers: map[string]struct{}{
					untrustedMachineUser.Login: {},
				},
			},
			pr: &github.PullRequest{
				Author:     normalUser,
				HeadRefOID: headRefOID,
				Reviews: &github.Reviews{
					Nodes: []*github.Review{
						{
							Author: octocat,
							State:  "APPROVED",
							Commit: headReviewCommit,
						},
					},
				},
				Commits: &github.Commits{
					Nodes: []*github.PullRequestCommit{
						{
							Commit: notLinkedCommit,
						},
					},
				},
			},
		},
		{
			name:  "If the pull request has commits from apps, require two approvals",
			exp:   errTwoApproval,
			input: &Input{},
			pr: &github.PullRequest{
				Author:     normalUser,
				HeadRefOID: headRefOID,
				Reviews: &github.Reviews{
					Nodes: []*github.Review{
						{
							Author: octocat,
							State:  "APPROVED",
							Commit: headReviewCommit,
						},
					},
				},
				Commits: &github.Commits{
					Nodes: []*github.PullRequestCommit{
						{
							Commit: normalCommit,
						},
						appPRCommit,
					},
				},
			},
		},
		{
			name: "If the pull request has commits from untrusted machine users, require two approvals",
			exp:  errTwoApproval,
			input: &Input{
				UntrustedMachineUsers: map[string]struct{}{
					untrustedMachineUser.Login: {},
				},
			},
			pr: &github.PullRequest{
				Author:     normalUser,
				HeadRefOID: headRefOID,
				Reviews: &github.Reviews{
					Nodes: []*github.Review{
						{
							Author: octocat,
							State:  "APPROVED",
							Commit: headReviewCommit,
						},
					},
				},
				Commits: &github.Commits{
					Nodes: []*github.PullRequestCommit{
						normalPRCommit,
						untrustedMachineUserPRCommit,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := validatePR(tt.input, tt.pr); err != nil {
				if tt.exp == nil {
					t.Fatal(err)
				}
				if errors.Is(err, tt.exp) {
					return
				}
				t.Fatalf("expected %v, got %v", tt.exp, err)
			}
			if tt.exp != nil {
				t.Fatalf("expected %v, got nil", tt.exp)
			}
		})
	}
}
