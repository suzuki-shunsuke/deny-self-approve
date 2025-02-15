# deny-self-approve

[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/suzuki-shunsuke/deny-self-approve/main/LICENSE) | [Usage](USAGE.md)

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

We provide a GitHub Actions: https://github.com/suzuki-shunsuke/deny-self-approve-action
This action is useful to prevent self-approvals.
