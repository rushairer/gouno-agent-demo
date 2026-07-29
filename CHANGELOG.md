# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [0.1.0] - 2026-07-29

### Added

- Initialize the gouno-based secure Agent API demo project.
- Add a constrained enterprise IT service-desk Agent with local FAQ retrieval, fixed system instructions, prompt-injection checks, out-of-scope refusal, and output safety checks.
- Add unified synchronous JSON and SSE endpoints backed by OpenAI Responses API and Anthropic Messages API adapters.
- Add gateway API-key authentication with bcrypt hashes, per-principal permissions and rate limits, bounded global IP rate limiting, and safe upstream host validation.
- Add provider configuration through server-side YAML and environment variables, health/readiness endpoints, OpenAPI documentation, API/provider/security tests, and a quick-start README.

### Changed

- Upgrade gouno from v1.0.0 to v1.0.2 and use immutable error-response constructors.

## [1.0.1] - 2026-06-13

### Changed

- Include complete module requirements and checksums so rendered projects can run Go tooling immediately.
- Return configuration load and validation errors from `ConfigManager` instead of exiting inside the config package.
- Add baseline configuration validation for generated projects.
- Strengthen template verification to cover downloaded module checksums.

## [1.0.0] - 2026-05-31

### Added

- Complete DDD project scaffold: `cmd/`, `config/`, `internal/` (domain, repository, service, task), `router/`, `middleware/`, `utility/`.
- Cobra CLI with `web` and `generator` commands.
- Viper multi-environment configuration (`development.yaml`, `test.yaml`, `production.yaml`).
- `ConfigManager` thread-safe configuration singleton.
- Gin web server with graceful shutdown.
- `Makefile` with build, run, dev, test targets.
- `.air.toml` for hot-reload development.
- Code generation templates (`domain.tmpl`, `repository.tmpl`, `service.tmpl`, `controller.tmpl`, `task.tmpl`).
- Bilingual README (English / Chinese).
