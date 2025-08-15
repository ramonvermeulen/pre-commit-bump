# pre-commit-bump
[![Go Report Card](https://goreportcard.com/badge/github.com/ramonvermeulen/pre-commit-bump)](https://goreportcard.com/report/github.com/ramonvermeulen/pre-commit-bump)
[![Go Reference](https://pkg.go.dev/badge/github.com/ramonvermeulen/pre-commit-bump.svg)](https://pkg.go.dev/github.com/ramonvermeulen/pre-commit-bump)

Dev tool to automatically bump the pre-commit hook versions in your `.pre-commit-config.yaml` file.
Mainly build for learning purposes, but can be useful for automating the process of updating pre-commit hooks.

> [!WARNING]  
> The pre-commit-bump tool is still in early development and not yet intended to be used.

## Installation
You can install `pre-commit-bump` using `go install`:

```bash
go install github.com/ramonvermeulen/pre-commit-bump@latest
```

## Use with GitHub Actions

t.b.d.

## Basic Usage

To use `pre-commit-bump`, simply run the command in the root of your repository:
```bash
pre-commit-bump --help
```

## Contributing
Contributions are welcome! Please create an issue or a pull request if you have any suggestions or improvements.

## License
This project is licensed under the Apache License 2.0. See the [LICENSE](LICENSE) file for details.