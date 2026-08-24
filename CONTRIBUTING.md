# Contributing

Use Go 1.25 or newer. Before opening a pull request, run formatting, `go test -race ./...`, `go vet ./...`, `govulncheck`, and `golangci-lint`. CI never calls a real AI provider; tests must use fake transports and local fixtures.

Use Conventional Commit PR titles and document user-visible changes in `CHANGELOG.md` under `Unreleased`.
