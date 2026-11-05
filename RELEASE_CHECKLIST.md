# Release Checklist

- [ ] Latest-patch Go 1.26.x and 1.27.x CI quality checks pass with no reachable vulnerabilities.
- [ ] CI did not use real AI provider credentials or external production services.
- [ ] `CHANGELOG.md` has a dated release entry and target tag/release do not exist.
- [ ] The worktree is clean and `git diff --check` passes.

- [ ] Agent JSON API, SSE, OpenAI Responses and Anthropic Messages provider tests pass without production credentials.
- [ ] Core v1.4.0 and the patched x/net v0.60.0 dependency baseline are verified through the public module proxy.
