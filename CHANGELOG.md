# Changelog

本文件记录 `gouno-agent-demo` 的重要变更。

格式遵循 [Keep a Changelog](https://keepachangelog.com/)，版本遵循 [Semantic Versioning](https://semver.org/)。

## [Unreleased]

## [0.2.0] - 2026-08-24

### Changed

- 最低 Go 版本提升到 1.25.0，升级到 `gouno` v1.2.0，并在 Go 1.25.x 与 1.26.x 上验证。

### Security

- 增加 SHA 固定的共享 CI：格式、模块一致性、race、vet、可达漏洞扫描、lint 和 Conventional PR 标题门禁。

## [0.1.0] - 2026-07-29

### Added

- 初始化基于 gouno 的安全 Agent API Demo。
- 添加受限企业 IT 服务台 Agent：本地 FAQ 检索、固定系统提示词、提示词攻击检测、范围外拒答和输出安全检查。
- 添加统一的同步 JSON 与 SSE 接口，以及 OpenAI Responses API 和 Anthropic Messages API 适配。
- 添加 bcrypt 网关 API Key、调用方权限与限流、全局 IP 限流和上游地址安全校验。
- 添加服务端 YAML/环境变量配置、健康与就绪检查、OpenAPI 文档、自动化测试和快速启动文档。

### Changed

- 将 gouno 升级至 v1.0.2，并使用不可变错误响应构造函数。
