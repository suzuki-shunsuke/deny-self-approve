package github

/*
query($owner: String!, $repo: String!, $pr: Int!) {
  repository(owner: $owner, name: $repo) {
    pullRequest(number: $pr) {
      reviews(first: 30) {
        totalCount
        pageInfo {
          hasNextPage
          endCursor
        }
        nodes {
          author {
            login
          }
          state
        }
      }
      commits(first: 30) {
        totalCount
        pageInfo {
          hasNextPage
          endCursor
        }
        nodes {
          commit {
            committer {
              user {
                login
              }
            }
          }
        }
      }
    }
  }
}
*/

type PageInfo struct {
	HasNextPage bool
	EndCursor   string
}

type GetPRQuery struct {
	Repository *Repository `graphql:"repository(owner: $repoOwner, name: $repoName)"`
}

type ListCommitsQuery struct {
	Repository *CommitsRepository `graphql:"repository(owner: $repoOwner, name: $repoName)"`
}

func (q *ListCommitsQuery) PageInfo() *PageInfo {
	return q.Repository.PullRequest.Commits.PageInfo
}

func (q *ListCommitsQuery) Nodes() []*PullRequestCommit {
	return q.Repository.PullRequest.Commits.Nodes
}

type ListReviewsQuery struct {
	Repository *ReviewsRepository `graphql:"repository(owner: $repoOwner, name: $repoName)"`
}

func (q *ListReviewsQuery) Nodes() []*Review {
	return q.Repository.PullRequest.Reviews.Nodes
}

func (q *ListReviewsQuery) PageInfo() *PageInfo {
	return q.Repository.PullRequest.Reviews.PageInfo
}

type Repository struct {
	PullRequest *PullRequest `graphql:"pullRequest(number: $number)"`
}

type CommitsRepository struct {
	PullRequest *CommitsPullRequest `graphql:"pullRequest(number: $number)"`
}

type CommitsPullRequest struct {
	Commits *Commits `graphql:"commits(first:30)"`
}

type ReviewsRepository struct {
	PullRequest *ReviewsPullRequest `graphql:"pullRequest(number: $number)"`
}

type ReviewsPullRequest struct {
	Reviews *Reviews `graphql:"reviews(first:30)"`
}

type PullRequest struct {
	HeadRefOID string
	Reviews    *Reviews `graphql:"reviews(first:30)"`
	Commits    *Commits `graphql:"commits(first:30)"`
}

type Reviews struct {
	TotalCount int
	PageInfo   *PageInfo
	Nodes      []*Review
}

type Review struct {
	ID     string
	Author *User
	State  string
	Commit *ReviewCommit
}

type ReviewCommit struct {
	OID string
}

type Commits struct {
	TotalCount int
	PageInfo   *PageInfo
	Nodes      []*PullRequestCommit
}

type PullRequestCommit struct {
	Commit *Commit
}

type Commit struct {
	Commiter *Commiter
}

type Commiter struct {
	User *User
}

type User struct {
	Login string
}
