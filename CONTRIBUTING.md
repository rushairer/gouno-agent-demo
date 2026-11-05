# Contributing

Use Go 1.26 or newer; CI verifies the latest Go 1.26.x and 1.27.x patches. Before opening a pull request, run formatting, `go test -race ./...`, `go vet ./...`, `govulncheck`, and `golangci-lint`. CI never calls a real AI provider; tests must use fake transports and local fixtures.

Use Conventional Commit PR titles and document user-visible changes in `CHANGELOG.md` under `Unreleased`.
