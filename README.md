# pre-commit-bump
[![Go Report Card](https://goreportcard.com/badge/github.com/ramonvermeulen/pre-commit-bump)](https://goreportcard.com/report/github.com/ramonvermeulen/pre-commit-bump)
[![Go Reference](https://pkg.go.dev/badge/github.com/ramonvermeulen/pre-commit-bump.svg)](https://pkg.go.dev/github.com/ramonvermeulen/pre-commit-bump)

Dev tool to automatically bump the pre-commit hook versions in your `.pre-commit-config.yaml` file.
Mainly build for learning purposes, but can be useful for automating the process of updating pre-commit hooks.

## Installation
You can install `pre-commit-bump` using `go install`:

```bash
go install github.com/ramonvermeulen/pre-commit-bump@latest
```

## Basic Usage

To use `pre-commit-bump`, run the command in the root of your repository:
```bash
pre-commit-bump update
```
Or to only check for updates without applying them:
```bash
pre-commit-bump check
```

Use `pre-commit-bump --help` to see all available commands and options.

## GitHub Actions

There are two ways to use `pre-commit-bump` in your GitHub Actions workflow:

### 1) pre-commit-bump PR action

This action combines the `ramonvermeulen/pre-commit-bump` action with the `peter-evans/create-pull-request` [action](https://github.com/marketplace/actions/create-pull-request)
to automatically create a pull request with the updated pre-commit hook versions.
Unfortunately, __this will not work for private repositories on the GitHub free plan__.

_"Draft pull requests are available in public repositories with GitHub Free and GitHub Free for organizations, GitHub Pro, and legacy per-repository billing plans, and in public and private repositories with GitHub Team and GitHub Enterprise Cloud."_

Also ensure the following setting is enabled in your repository settings -> actions -> general:

![setting.png](.github/docs/setting.png)

```yaml
name: Bump pre-commit hooks

on:
  schedule:
    - cron: '0 0 * * *' # Every day at midnight
  workflow_dispatch:

permissions:
  contents: write
  pull-requests: write
  
jobs:
  pre-commit-bump:
    name: Run pre-commit-bump
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Update pre-commit hooks
        uses: ramonvermeulen/pre-commit-bump/gha/bot@v1
        with:
          command: update
          allow: major
          verbose: true
```

Example of a pull request created by the action:
![example_1.png](.github/docs/example_1.png)

![example_2.png](.github/docs/example_2.png)

### 2) pre-commit-bump standalone action
The standalone action is mostly used for checking the pre-commit hooks without creating a pull request.
This is useful for CI/CD pipelines to ensure that the pre-commit hooks are up-to-date, the action will fail if there 
are updates available. The potential updates will be logged in the GitHub actions log.

```yaml
name: Check pre-commit hooks

on:
  schedule:
    - cron: '0 0 * * *' # Every day at midnight
  workflow_dispatch:
  
jobs:
  lint:
    name: Pre-Commit Bump
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Pre-Commit Bump
        uses: ramonvermeulen/pre-commit-bump@v1
        with:
          command: check
          allow: major
          verbose: true
```

### Action inputs

All inputs are **optional**. If not set, sensible defaults will be used.

| Name         | Description                                                                                    | Default                  |
|--------------|------------------------------------------------------------------------------------------------|--------------------------|
| `command`    | Command to run, can be either `update` or `check`.                                             | `update`                 |
| `allow`      | Specific semantic versioning range to allow updates for (`major`, `minor`, or `patch`).        | `major`                  |
| `verbose`    | Whether to run in verbose mode.                                                                | `false`                  |
| `config`     | Path to the pre-commit configuration file, uses `.pre-commit-config.yaml` if not specified.    | `pre-commit-config.yaml` |
| `no-summary` | Whether to skip the summary output (generation of `summary.md` which is used as PR body).      | `false`                  |
| `dry-run`    | Whether to perform a dry run without making changes to the pre-commit yaml configuration file. | `false`                  |

## Contributing
Contributions are welcome! Please create an issue or a pull request if you have any suggestions or improvements.

## License
This project is licensed under the Apache License 2.0. See the [LICENSE](LICENSE) file for details.
