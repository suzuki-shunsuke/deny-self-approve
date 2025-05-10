# deny-self-approve

[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/suzuki-shunsuke/deny-self-approve/main/LICENSE) | [Install](INSTALL.md) | [Usage](USAGE.md) | [GitHub Action](https://github.com/suzuki-shunsuke/deny-self-approve-action)

`deny-self-approve` is a CLI tool to validate reviews of GitHub Pull Requests.
It makes GitHub Actions secure.
It enforces the requirement for reviews and prevents pull requests from being merged without proper review.
While making reviews mandatory in branch rulesets helps, there are still loopholes that allow pull requests to be merged without a review.
This action helps close those loopholes.

When developing as a team, it's common to require that pull requests be reviewed by someone other than the author.
Code reviews help improve code quality, facilitate knowledge sharing among team members, and prevent any single person from making unauthorized changes without approval.
While you can enforce reviews using Branch Rulesets, there are still a few loopholes that allow pull requests to be merged without proper review.
This action was developed to close those loopholes.

First, you should enable the following branch ruleset on the default branch.

- `Require a pull request before merging`
  - `Require review from Code Owners`
  - `Require approval of the most recent reviewable push`
- `Require status checks to pass`

This rules require pull request reviews, but there are still several ways to improperly merge a pull request without a valid review:

1. Abusing a machine user with `CODEOWNER` privileges to approve the PR.
2. Adding commits to someone else’s PR and approving it yourself.
3. Using a machine user or bot to add commits to someone else’s PR, then approving it yourself.

To address these loopholes, you can run `deny-self-approve` using events like `pull_request_review` or `merge_group`, which are triggered right before a pull request is merged.
By adding the job running `deny-self-approve` to Required Checks, you can block invalid merges effectively.

`validate-pr-review-action` performs the following validations:

- The latest commit in the PR must be approved by someone who did **not** contribute commits to the PR.
- Approvals from GitHub Apps or untrusted machine users are ignored
  - This is because such accounts could be misused to provide unauthorized approvals.

In the following cases, **two or more approvals** are required:

- If any of the commits were made by an **untrusted** machine user or a GitHub App (excluding a few **trusted** ones).
  - This is to prevent abuse where someone uses such accounts to commit and then self-approves.
- If the pull request was created by an untrusted machine user or GitHub App (excluding a few trusted ones).
  - This helps prevent scenarios where a PR is created by such an account and then self-approved.
- If there are commits without a linked GitHub user.
  - This may indicate an attempt to mask the identity of the committer and self-approve.

## Trusted Apps and untrusted users

With the deny-self-approve option, you can specify lists of trusted GitHub Apps and untrusted machine users.

- --trusted-apps `<GitHub App 1>,<GitHub App 2>,...`
  - You should use this option carefully. You shouldn't specify GitHub Apps not managing securely
- --untrusted-machine-users `<Machine User 1>,<Machine User 2>`
  - You should specify all Machine Users except for Machine Users managing securely

Whether a GitHub App is considered trusted or a user is considered an untrusted machine user depends on how securely they are managed and whether they are susceptible to misuse.

For example, if a GitHub App is installed across all repositories in an organization and granted `contents:write` and `pull_requests:write` permissions, and if its App ID and private key are shared across all repositories via GitHub Organization Variables and Secrets, that App cannot be trusted.
Any organization member can exploit the App to create pull requests, make commits, or approve changes from any branch in any repository.

By default, only `renovate[bot]` and `dependabot[bot]` are treated as trusted GitHub Apps.
All others are considered untrusted unless explicitly specified.

### Steps to secure GitHub Apps and Machine Users

In many organizations, machine users and GitHub Apps are often not properly managed securely.
To address this, consider following these steps:

1. Create new machine users and GitHub Apps.
1. Apply strict access controls to these newly created accounts.
1. Gradually replace existing machine users and GitHub Apps with the newly secured ones.
1. Minimize permissions for any existing accounts that remain in use.
1. Decommission unused or insecure accounts.

### Client/Server Model Actions

Client/Server Model Actions allow you to manage GitHub Apps and Machine Users securely.
For more details, see:

👉 https://github.com/csm-actions/docs

## GitHub Actions

[We provide a GitHub Actions to prevent self-approvals easily.](https://github.com/suzuki-shunsuke/deny-self-approve-action)

## Get a repository and pull request number from CI environment

If you run this tool on your machine, you need to specify parameters `-repo` and `-pr`.

e.g.

```sh
deny-self-approve validate -r suzuki-shunsuke/deny-self-approve -pr 1
```

But in some CI platoforms such as GitHub Actions and CircleCI, you don't need to specify them because this tool gets these parameters automatically from environment variables and files.
This tool uses a library [go-ci-env](https://github.com/suzuki-shunsuke/go-ci-env).
