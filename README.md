# deny-self-approve

[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/suzuki-shunsuke/deny-self-approve/main/LICENSE) | [Install](INSTALL.md) | [Usage](USAGE.md) | [GitHub Action](https://github.com/suzuki-shunsuke/deny-self-approve-action)

`deny-self-approve` is a CLI tool designed to validate self-approved GitHub Pull Requests.

## :warning: Deprecated. Please use [validate-pr-review-action](https://github.com/suzuki-shunsuke/validate-pr-review-action) :warning:

validate-pr-review-action is a successor of deny-self-approve.

https://github.com/suzuki-shunsuke/validate-pr-review-action

---

```sh
deny-self-approve validate -r <repository full name> -pr <pr number>
```

The command fails if a given pull request isn't approved by someone who isn't a committer of the pull request.
It requires a approval from non committer of the pull request.

The exception is that multiple people approve the pull request.
The goal of this tool is to prevent a single person from merging a pull request without approvals from others by self-approval.
If multiple people approve the pull request, the goal is met.

We assume it's run in CI.
The following GitHub Repository Branch Rulesets are useful to protect branches like default branches:

- `Require a pull request before merging`
  - `Dismiss stale pull request approvals when new commits are pushed`
  - `Require review from Code Owners`
  - `Require approval of the most recent reviewable push`
- `Require status checks to pass`

But even if you configure these rulesets properly, people can still bypass the restriction.
For instance, people can approve pull requests using GitHub Actions token, GitHub App, or Machine Users.
And people can also push commits to pull requests created by others (other users, GitHub Actions token, GitHub App, or Machine Users) and approve them.

This tool prevents such a threat.

## GitHub Actions

[We provide a GitHub Actions to prevent self-approvals easily.](https://github.com/suzuki-shunsuke/deny-self-approve-action)

## :warning: Commit not linked to a GitHub User

[Please see the document.](docs/001.md)

## Get a repository and pull request number from CI environment

If you run this tool on your machine, you need to specify parameters `-repo` and `-pr`.

e.g.

```sh
deny-self-approve validate -r suzuki-shunsuke/deny-self-approve -pr 1
```

But in some CI platoforms such as GitHub Actions and CircleCI, you don't need to specify them because this tool gets these parameters automatically from environment variables and files.
This tool uses a library [go-ci-env](https://github.com/suzuki-shunsuke/go-ci-env).
