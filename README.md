# deny-self-approve

[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/suzuki-shunsuke/deny-self-approve/main/LICENSE) | [Install](docs/install.md)

deny-self-approve is a CLI to deny self-approved GitHub Pull Requests.

This tool is useful to prevent unreviewed pull requests from being merged.

Even if you configure the branch rule set or branch protection rule properly, people can bypass the restriction.
For instance, people can approve pull requests using GitHub Actions token, GitHub App, or Machine Users.
And people can also push commits to pull requests created by others (other users, GitHub Actions token, GitHub App, or Machine Users) and approve them.

To prevent this threat, this tool checks if a given pull request approvers and commiters.
It can dismiss self-approvals, and it validates if the pull request is approved by someone other than committers.
