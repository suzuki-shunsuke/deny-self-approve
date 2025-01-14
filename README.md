# deny-self-approve

[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/suzuki-shunsuke/deny-self-approve/main/LICENSE) | [Usage](USAGE.md)

`deny-self-approve` is a CLI tool designed to validate self-approved GitHub Pull Requests.
It dismisses self-approvals and triggers a CI failure if no external approver — someone who did not push commits to the pull request — approves the changes.

We assume it's run in CI.
The following GitHub Repository Branch Rulesets is useful to protect branches like default branches:

- `Require a pull request before merging`
- `Require status checks to pass`
- `Require approval of the most recent reviewable push`
- `Require review from Code Owners`

But even if you configure these rulesets properly, people can still bypass the restriction.
For instance, people can approve pull requests using GitHub Actions token, GitHub App, or Machine Users.
And people can also push commits to pull requests created by others (other users, GitHub Actions token, GitHub App, or Machine Users) and approve them.

To prevent this threat, this tool checks if a given pull request approvers and committers.
It can dismiss self-approvals, and it validates if the pull request is approved by someone other than committers.

## How To Use

### Dismiss approvals by pull_request_review events

<img width="964" alt="image" src="https://github.com/user-attachments/assets/fc5bbd3d-6b04-495d-8b72-d14a81a93dc0" />

```yaml
name: Dismiss self approvals
on:
  pull_request_review:
    types:
      - submitted
      - edited
permissions: {}
jobs:
  dismiss:
    timeout-minutes: 10
    runs-on: ubuntu-latest
    permissions:
      pull-requests: write
    steps:
      - uses: actions/checkout@11bd71901bbe5b1630ceea73d27597364c9af683 # v4.2.2
        with:
          persist-credentials: false
      - uses: aquaproj/aqua-installer@f13c5d2f0357708d85477aabe50fd3f725528745 # v3.1.0
        with:
          aqua_version: v2.42.2
      - run: deny-self-approve -d
        env:
          GITHUB_TOKEN: ${{github.token}}
```

### Check if pull requests are self-approved before deploying

<img width="933" alt="image" src="https://github.com/user-attachments/assets/05a441e7-99a2-4a2f-a5f5-9b04401992b8" />

```yaml
name: Deploy
on:
  push:
    branches:
      - main
permissions: {}
jobs:
  deploy:
    timeout-minutes: 10
    runs-on: ubuntu-latest
    permissions:
      pull-requests: write
    steps:
      - uses: actions/checkout@11bd71901bbe5b1630ceea73d27597364c9af683 # v4.2.2
        with:
          persist-credentials: false
      - uses: aquaproj/aqua-installer@f13c5d2f0357708d85477aabe50fd3f725528745 # v3.1.0
        with:
          aqua_version: v2.42.2
      - run: |
          env=$(ci-info run)
          eval "$env"
          echo "PR_NUMBER=$CI_INFO_PR_NUMBER" >> "$GITHUB_ENV"
      - run: |
          if ! deny-self-approve --pr "$PR_NUMBER"; then
            github-comment post -k deny-self-approve -pr "$PR_NUMBER"
            exit 1
          fi
        env:
          GITHUB_TOKEN: ${{github.token}}
      - run: echo "Deploy"
```
