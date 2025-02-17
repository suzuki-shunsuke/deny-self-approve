# deny-self-approve

[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/suzuki-shunsuke/deny-self-approve/main/LICENSE) | [Install](INSTALL.md) | [Usage](USAGE.md) | [GitHub Action](https://github.com/suzuki-shunsuke/deny-self-approve-action)

`deny-self-approve` is a CLI tool designed to validate self-approved GitHub Pull Requests.
It dismisses self-approvals and triggers a CI failure if no external approver — someone who did not push commits to the pull request — approves the changes.

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

To prevent this threat, this tool checks if a given pull request approvers and commiters.
It can dismiss self-approvals.
And it can also validate pull requests before doing something like deployment.

## GitHub Actions

[We provide a GitHub Actions to prevent self-approvals easily.](https://github.com/suzuki-shunsuke/deny-self-approve-action)

## :warning: Commit not linked to a GitHub User

deny-self-approve gets commits' users to deny self approvals.
But if a commit isn't linked to a GitHub User, deny-self-approve can't get a committer.
In that case, deny-self-approve fails by default.

- [You should configure Git properly](https://docs.github.com/en/pull-requests/committing-changes-to-your-project/troubleshooting-commits/why-are-my-commits-linked-to-the-wrong-user).
- [As a best practice, all commits should be verified](https://docs.github.com/en/authentication/managing-commit-signature-verification).

But if it's difficult, you can ignore these commits by `--ignore-not-linked-commit` option.

```sh
deny-self-approve --ignore-not-linked-commit validate # or dismiss
```

We don't recommend this because this isn't secure.
