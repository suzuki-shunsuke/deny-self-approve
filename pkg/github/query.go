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
	HasNextPage bool   `json:"hasNextPage"`
	EndCursor   string `json:"endCursor"`
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
	HeadRefOID string   `json:"headRefOid"`
	Reviews    *Reviews `json:"reviews" graphql:"reviews(first:30)"`
	Commits    *Commits `json:"commits" graphql:"commits(first:30)"`
}

type Reviews struct {
	TotalCount int       `json:"totalCount"`
	PageInfo   *PageInfo `json:"pageInfo"`
	Nodes      []*Review `json:"nodes"`
}

type Review struct {
	ID     string        `json:"id"`
	Author *User         `json:"author"`
	State  string        `json:"state"`
	Commit *ReviewCommit `json:"commit"`
}

type ReviewCommit struct {
	OID string `json:"oid"`
}

type Commits struct {
	TotalCount int                  `json:"totalCount"`
	PageInfo   *PageInfo            `json:"pageInfo"`
	Nodes      []*PullRequestCommit `json:"nodes"`
}

type PullRequestCommit struct {
	Commit *Commit `json:"commit"`
}

func (c *Commit) Login() string {
	if c.Committer.User != nil {
		return c.Committer.User.Login
	}
	return c.Author.User.Login
}

type Commit struct {
	Committer *Committer `json:"committer"`
	Author    *Committer `json:"author"`
}

type Committer struct {
	User *User `json:"user"`
}

type User struct {
	Login string `json:"login"`
}
