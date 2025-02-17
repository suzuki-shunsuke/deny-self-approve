package github_test

import (
	"testing"

	"github.com/suzuki-shunsuke/deny-self-approve/pkg/github"
)

func TestCommit_Login(t *testing.T) {
	t.Parallel()
	data := []struct {
		name   string
		commit *github.Commit
		exp    string
	}{
		{
			name: "author",
			commit: &github.Commit{
				Committer: &github.Committer{},
				Author: &github.Committer{
					User: &github.User{
						Login: "foo",
					},
				},
			},
			exp: "foo",
		},
		{
			name: "committer",
			commit: &github.Commit{
				Author: &github.Committer{},
				Committer: &github.Committer{
					User: &github.User{
						Login: "bar",
					},
				},
			},
			exp: "bar",
		},
		{
			name: "not linked commit",
			commit: &github.Commit{
				Author:    &github.Committer{},
				Committer: &github.Committer{},
			},
			exp: "",
		},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			login := d.commit.Login()
			if login != d.exp {
				t.Fatalf("wanted: %s, got: %s", d.exp, login)
			}
		})
	}
}
