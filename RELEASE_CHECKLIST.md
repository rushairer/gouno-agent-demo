# Release Checklist

- [x] Latest-patch Go 1.26.x and 1.27.x CI quality checks pass with no reachable vulnerabilities.
- [x] CI did not use real AI provider credentials or external production services.
- [x] `CHANGELOG.md` has a dated release entry and target tag/release do not exist.
- [ ] The worktree is clean and `git diff --check` passes.

- [x] Agent JSON API, SSE, OpenAI Responses and Anthropic Messages provider tests pass without production credentials.
- [x] Core v1.4.0 and the patched x/net v0.60.0 dependency baseline are verified through the public module proxy.

Evidence: hosted CI run 42010715395 passed on the dependency-upgrade commit d6485f225ae57008ca785b4a61f6e60ee197cf4b (Go 1.26.x / 1.27.x, lint, race, vet, module verification and govulncheck). Local Go 1.26.10 / 1.27.3 race/vet and provider/Agent tests passed. The final release-metadata commit must pass the same hosted gates before v0.3.0 is created. The historical v0.2.0 changelog entry is not represented as an existing tag/release.
