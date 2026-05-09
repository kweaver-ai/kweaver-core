# Development Rules

[中文](README.zh.md) | English

This directory contains the development standards for the KWeaver project. If you're new, read them in the order below.

## Reading Order

| # | Document | What you'll learn |
|---|----------|-------------------|
| 1 | [CONTRIBUTING](CONTRIBUTING.md) | How to fork, submit PRs, branch naming, code style |
| 2 | [WORKFLOW](WORKFLOW.md) | Issue → design doc → branch → PR → merge flow |
| 3 | [ARCHITECTURE](ARCHITECTURE.md) | System layering, module boundaries, dependency rules |
| 4 | [DEVELOPMENT](DEVELOPMENT.md) | API design, HTTP semantics, error handling, authentication |
| 5 | [TESTING](TESTING.md) | Test layers, Makefile targets, Agent-First testing |
| 6 | [RELEASE](RELEASE.md) | Versioning, branching strategy, release process |

## Quick Reference

- **Branch naming**: `feature/`, `fix/`, `refactor/`, `docs/`, `ci/`, `chore/` — see [CONTRIBUTING](CONTRIBUTING.md)
- **Commit format**: Conventional Commits (`type(scope): subject`) — see [CONTRIBUTING](CONTRIBUTING.md)
- **Documentation layout**: Root [`docs/`](../docs), per-module `docs/`, [`help/`](../help) manuals (`manual/`) and cookbooks (`cookbook/`) — see [CONTRIBUTING — DOCUMENTATION](CONTRIBUTING.md#documentation)
- **Design docs**: Located in each module's `docs/design/` directory — see [WORKFLOW](WORKFLOW.md)
- **API checklist**: Error format, pagination, status codes — see [DEVELOPMENT](DEVELOPMENT.md)
