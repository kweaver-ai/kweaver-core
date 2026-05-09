# Contributing Guide

[中文](CONTRIBUTING.zh.md) | English

Thank you for your interest in contributing to KWeaver! We welcome all forms of contributions, including bug fixes, feature proposals, documentation improvements, answering questions, and more.

Please read this guide before submitting contributions to ensure consistent processes and standardized submissions.

---

## 🏗 Repository Layout

KWeaver Core is a **monorepo** ([`kweaver-ai/kweaver-core`](https://github.com/kweaver-ai/kweaver-core)) that ships the platform's backend modules together. Pick the directory matching the component you want to change:

| Module | Path | Description |
| --- | --- | --- |
| **AI Data Platform (ADP)** | [`adp/`](../adp) | BKN Engine (`adp/bkn`), Context Loader (`adp/context-loader`), Dataflow (`adp/dataflow`), Execution Factory (`adp/execution-factory`), VEGA virtualization (`adp/vega`) |
| **Decision Agent** | [`decision-agent/`](../decision-agent) | Agent Executor / Factory / Memory under `agent-backend/` |
| **Trace AI** | [`trace-ai/`](../trace-ai) | Agent observability and OpenTelemetry collector chart |
| **Infra** | [`infra/`](../infra) | `mf-model-manager` (model registry), `oss-gateway-backend`, `sandbox` runtime |
| **BKN samples** | [`bkn/`](../bkn) | Reference Business Knowledge Networks (e.g. `smart_home_supply_chain`) |
| **Examples** | [`examples/`](../examples) | End-to-end CLI walkthroughs (DB / CSV / actions) |
| **Help / Website** | [`help/`](../help), [`website/`](../website) | Bilingual product docs and the public docs site |
| **Deploy** | [`deploy/`](../deploy) | One-click `deploy.sh` for Kubernetes (k8s + KWeaver Core charts) |

External CLI/SDKs that interact with this backend live in their own repositories:

| Repository | Purpose |
| --- | --- |
| [`kweaver-ai/kweaver-sdk`](https://github.com/kweaver-ai/kweaver-sdk) | `kweaver` CLI + TypeScript / Python SDK + AI agent skills (`kweaver-core`, `create-bkn`) |
| [`kweaver-ai/kweaver-admin`](https://github.com/kweaver-ai/kweaver-admin) | `kweaver-admin` platform-administrator CLI + agent skill |

> **Note**: Each module has its own README (and often `AGENTS.md` / `CLAUDE.md`) with build, test, and dev-loop instructions in the language it is written in. Read them before making module-local changes.

---

## 📖 DOCUMENTATION

Use the location that matches **audience and scope**:

1. **[`docs/`](../docs) (repository root)** — System architecture, repo-wide design, and technical decisions that span multiple subsystems or the whole Core platform.
2. **`<module>/docs/` (inside each subtree)** — Design notes and technical decisions **scoped to that module only** (for example [`adp/bkn/docs/`](../adp/bkn/docs), [`decision-agent/`](../decision-agent) design folders under each service).
3. **[`help/{en,zh}/manual/`](../help/en/manual)** — Product **manuals / reference**: how subsystems work from a user or operator perspective (install stays next to the manual tree as `help/{lang}/install.md`).
4. **[`help/{en,zh}/cookbook/`](../help/en/cookbook)** — **Cookbooks**: short, runnable task recipes. Start a new one by copying [`_TEMPLATE.md`](../help/en/cookbook/_TEMPLATE.md), use the worked-out [`cookbook_example.md`](../help/en/cookbook/cookbook_example.md) as reference, and add a row to [`cookbook/README.md`](../help/en/cookbook/README.md) (Chinese lives under `help/zh/cookbook/`).

Keep **English** in `help/en/` and **Chinese** in `help/zh/` for user-facing help.

---

## 🧩 Types of Contributions

You can contribute in the following ways:

- 🐛 **Report Bugs**: Help us identify and fix issues
- 🌟 **Propose Features**: Suggest new functionality or improvements
- 📚 **Improve Documentation**: Enhance docs, examples, or tutorials
- 🔧 **Fix Bugs**: Submit patches for existing issues
- 🚀 **Implement Features**: Build new functionality
- 🧪 **Add Tests**: Improve test coverage
- 🎨 **Refactor Code**: Optimize code structure and improve maintainability

---

## 🗂 Issue Guidelines (Bug & Feature)

### 1. Bug Report Format

When reporting a bug, please provide the following information:

- **Version/Environment**:
  - KWeaver Core version (`git describe --tags` or `VERSION` file, e.g. `v0.6.0`)
  - Module affected (e.g. `adp/bkn`, `decision-agent/agent-backend/agent-executor`, `infra/sandbox`)
  - Runtime (Java / Go / Python / Node — and version, e.g. JDK 17, Go 1.23, Python 3.11)
  - OS (Linux distro + kernel, macOS, Windows)
  - Cluster (single-node K3s / kubeadm / managed K8s) and how it was installed (`deploy.sh kweaver-core install [--minimum]`)
  - Storage backends as relevant (MariaDB / DM8, OpenSearch, Redis, Kafka)

- **Reproduction Steps**: Clear, step-by-step instructions to reproduce the issue

- **Expected vs Actual Behavior**: What should happen vs what actually happens

- **Error Logs/Screenshots**: Include relevant error messages, stack traces, or screenshots

- **Minimal Reproducible Code (MRC)**: A minimal code example that demonstrates the issue

**Example Bug Report Template:**

```markdown
**Environment:**
- KWeaver Core: v0.6.0
- Module: adp/bkn
- Runtime: JDK 17
- OS: Linux Ubuntu 22.04
- Cluster: single-node K3s (deploy.sh kweaver-core install)
- Database: MariaDB 11.4

**Steps to Reproduce:**
1. Start the service
2. Perform the action
3. Error occurs

**Expected Behavior:**
Action should complete successfully

**Actual Behavior:**
Error: "unexpected error"

**Error Log:**
[Paste error log here]
```

### 2. Feature Request Format

When proposing a feature, please describe:

- **Background/Purpose**: Why is this feature needed? What problem does it solve?

- **Feature Description**: Detailed description of the proposed functionality

- **API Design** (if applicable): Proposed API changes or new endpoints

- **Backward Compatibility**: Potential impact on existing functionality

- **Implementation Direction** (optional): Suggestions on how to implement it

> **Note**: All major features should be discussed in an Issue first before submitting a Pull Request.

**Example Feature Request Template:**

```markdown
**Background:**
Currently, users need to manually refresh the knowledge network after updates.
This feature would automate the refresh process.

**Feature Description:**
Add an auto-refresh mechanism that updates the knowledge network when
underlying data changes.

**Proposed API:**
POST /api/v1/networks/{id}/auto-refresh
{
  "enabled": true,
  "interval": 300
}

**Backward Compatibility:**
This is a new feature and does not affect existing functionality.
```

---

## 🔀 Pull Request (PR) Process

### 1. Fork the Repository

Fork the repository to your GitHub account.

### 2. Create a Branch

Create a new branch from `main` (or the appropriate base branch):

```bash
git checkout -b feature/my-feature
# or
git checkout -b fix/bug-description
# or (module scope + linked Issue: second segment often starts with "<number>-<slug>")
git checkout -b feature/agent-web/123-add-login
git checkout -b fix/bkn-backend/456-query-timeout
```

**Branch Naming Convention:**

Branch names are validated automatically by CI on every Pull Request (see `.github/workflows/lint-branch-name.yml`).

**Formats (at most two path segments after the type prefix):**

- `<type>/<description>`, e.g. `feature/add-oauth-support`
- `<type>/<issue-number>-<description>` (common when linking an Issue), e.g. `fix/123-memory-leak`
- `<type>/<module>/<description>` (optional, for scoped work); the **description** segment may include an issue number, e.g. `feature/agent-web/123-add-login`, `fix/bkn-backend/456-query-timeout`

**Do not** use three or more segments after the type, e.g. `feature/foo/bar/baz` will fail CI.

**Examples (module + Issue)**:

| Branch name | Notes |
| --- | --- |
| `feature/studio/789-export-pipeline` | Feature scoped to `studio`, Issue `#789` |
| `fix/ontology-query/404-id-not-exist` | Fix scoped to `ontology-query`, Issue `#404` |
| `docs/rules/120-contributing-branch` | Docs under `rules/`, Issue `#120` |

| Branch Type | Format | Description | Example |
| --- | --- | --- | --- |
| Feature | `feature/*` or `feat/*` | New feature development | `feature/add-oauth-support` |
| Fix | `fix/*` | Bug fixes | `fix/123-memory-leak-in-loader` |
| Hotfix | `hotfix/*` | Urgent production fixes | `hotfix/critical-auth-bypass` |
| Docs | `docs/*` | Documentation changes | `docs/update-api-reference` |
| Refactor | `refactor/*` | Code refactoring | `refactor/simplify-auth-flow` |
| Test | `test/*` | Adding or updating tests | `test/add-unit-tests-for-loader` |
| Chore | `chore/*` | Maintenance tasks | `chore/upgrade-dependencies` |
| CI | `ci/*` | CI/CD configuration changes | `ci/add-branch-name-lint` |
| Performance | `perf/*` | Performance improvements | `perf/optimize-query-execution` |
| Build | `build/*` | Build system or dependencies | `build/update-go-module` |
| Style | `style/*` | Code style / formatting | `style/fix-linter-warnings` |
| Revert | `revert/*` | Reverting previous changes | `revert/rollback-auth-change` |
| Release | `release/x.y.z` (optional prerelease suffix) | Release preparation | `release/1.2.0`, `release/1.2.0-rc.1` |

Rules:
- If the branch is linked to an Issue, a common pattern is `<type>/<N>-<description>` (e.g. `fix/123-memory-leak`); you can also put the issue number in a two-segment path (e.g. `feature/agent-web/123-add-login`)
- Each segment (between `/`) must start with a lowercase letter or digit; the rest may use hyphens (`-`), dots (`.`), or underscores (`_`), matching CI
- Branch names must start with a valid type prefix followed by `/`
- Bot branches (`dependabot/*`, `renovate/*`) are automatically exempted

> **Note**: For branching strategy, versioning rules, and release process, see [Release Guidelines](RELEASE.md).

### 3. Make Your Changes

- Write clean, maintainable code
- Follow the project's code structure and architecture patterns
- Add appropriate comments and documentation
- Include standard file headers (see [Source Code Header Guidelines](#-source-code-header-guidelines) below)

### 4. Write Tests

- Add unit tests for new functionality
- Ensure existing tests still pass
- Aim for good test coverage
- Use the test runner that fits the module's language. Each module's README / `AGENTS.md` describes the canonical command — for example:
  - Java (Maven) modules: `mvn test`
  - Go modules: `go test ./...`
  - Python modules: `pytest`
  - Node / TypeScript modules: `npm test`

### 5. Update Documentation

- Update relevant documentation if your changes affect user-facing features
- Update API documentation if you modify endpoints
- Add examples if introducing new functionality
- Update CHANGELOG.md if applicable

#### README Guidelines

When updating README files, please follow these guidelines:

- **Default Language**: `README.md` should be in English (default)
- **Chinese Version**: Chinese documentation should be in `README.zh.md`
- **Keep in Sync**: If you update `README.md`, please also update `README.zh.md` accordingly
- **Structure**: Maintain consistent structure between English and Chinese versions
- **Links**: Update language switcher links at the top of each README file:
  - English: `[中文](README.zh.md) | English`
  - Chinese: `[中文](README.zh.md) | [English](README.md)`

**Example README Structure:**

```markdown
# Project Name

[中文](README.zh.md) | English

[![License](...)](LICENSE.txt)
[![Go Version](...)](...)

Brief description...

## 📚 Quick Links

- Links to documentation, contributing guide, etc.

## Main Content

...
```

### 6. Commit Your Changes

Write clear, descriptive commit messages:

```bash
git commit -m "feat: add auto-refresh for knowledge networks

- Add auto-refresh configuration endpoint
- Implement background refresh worker
- Add tests for refresh functionality

Closes #123"
```

**Commit Message Format:**

Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

| Type | Description |
| --- | --- |
| `feat` | A new feature |
| `fix` | A bug fix |
| `docs` | Documentation only changes |
| `style` | Code style changes (formatting, etc.) |
| `refactor` | Code refactoring (not feat/fix) |
| `perf` | Performance improvements |
| `test` | Adding or updating tests |
| `build` | Build system or dependencies |
| `ci` | CI configuration changes |
| `chore` | Other maintenance tasks |
| `revert` | Revert a previous commit |

> **Note**: For detailed commit conventions and versioning rules, see [Release Guidelines](RELEASE.md).

### 7. Keep Your Branch Up to Date

Since this project requires linear history, please rebase your branch on the latest `main` branch before pushing:

```bash
# Make sure you're on your feature branch
git checkout feature/my-feature

# Ensure all changes are committed
git status  # Check for uncommitted changes

# If you have uncommitted changes, commit them first:
# git add .
# git commit -m "your commit message"

# Option 1: If you have upstream configured, fetch and rebase on upstream/main
# git fetch upstream
# git rebase upstream/main

# Option 2: Fetch latest changes from origin and rebase on origin/main
git fetch origin
git rebase origin/main

# If there are conflicts, resolve them and continue:
# 1. Fix conflicts in the affected files
# 2. git add <resolved-files>
# 3. git rebase --continue

# If you want to abort the rebase:
# git rebase --abort

# Force push (required after rebase)
git push origin feature/my-feature --force-with-lease
```

> **Note**:
>
> - Use `--force-with-lease` instead of `--force` to avoid overwriting others' work.
> - Make sure you're on your feature branch before rebasing.
> - If you prefer to track the upstream repository, you can add it: `git remote add upstream https://github.com/kweaver-ai/kweaver-core.git`

### 8. Push to Your Fork

```bash
git push origin feature/my-feature
```

### 9. Create a Pull Request

1. Go to the original repository on GitHub
2. Click "New Pull Request"
3. Select your fork and branch
4. Fill out the PR template with:
   - Description of changes
   - Related issue number (if applicable)
   - Testing instructions
   - Screenshots (if UI changes)

**PR Checklist:**

- [ ] Self-review completed
- [ ] Comments added for complex code
- [ ] Documentation updated
- [ ] Tests added/updated
- [ ] All tests pass
- [ ] Changes are backward compatible (or migration guide provided)

---

## 📋 Code Review Process

1. **Automated Checks**: PRs will be checked by CI/CD pipelines
   - Unit tests
   - Build verification

2. **Review**: Maintainers will review your PR
   - Address review comments promptly
   - Make requested changes
   - Keep discussions constructive

3. **Approval**: Once approved, a maintainer will merge your PR
   - PRs will be merged using squash merge or rebase merge to maintain linear history
   - Please ensure your branch is up to date before requesting review

---

## ⚙️ CI Workflow Guidelines

All GitHub Actions workflow files live in `.github/workflows/`. GitHub does **not** support subdirectories — files in nested folders will be silently ignored.

### File Naming Convention

Use a **category prefix** to group related workflows, so they sort together in the file list:

| Prefix | Purpose | Example |
| --- | --- | --- |
| `lint-` | Code / commit / branch linting | `lint-branch-name.yml`, `lint-commit.yml`, `lint-workflow-files.yml` |
| `ci-` | Build, test, typecheck, integration on PR or push | `ci-backend.yml`, `ci-website.yml` |
| `release-` | Build & publish releases | `release-agent-observability.yml` |
| `deploy-` | Deployment tasks | `deploy-pages.yml` |
| `security-` | Supply chain / application security scanning | `security-codeql.yml`, `security-dependency-review.yml` |
| `automation-` | Repo bots and scheduled housekeeping | `automation-stale.yml`, `automation-labeler.yml` |
| `reusable-` | Callable-only workflows (`on.workflow_call`) | `reusable-ci-go.yml` |

Rules:
- File names must be lowercase kebab-case with a `.yml` extension
- Always use a category prefix to keep related workflows visually grouped
- If a workflow references its own file path (e.g. in `on.push.paths`), rename both the file **and** the internal path reference together

### Reusable Workflows & Composite Actions

When workflow count grows or shared logic emerges:
- **Reusable workflows** can be placed in `.github/workflows/` and called via `uses: ./.github/workflows/xxx.yml`
- **Composite actions** can be placed in `.github/actions/<name>/action.yml` for step-level reuse

---

## 📝 Source Code Header Guidelines

This section defines the standard source code file header used across **kweaver.ai** open-source projects.

The goal is to ensure:

- clear copyright ownership
- clear licensing (Apache License 2.0)
- consistent and readable file documentation

> **Note**: We use "The kweaver.ai Authors" instead of individual author names.
> Git history already tracks all contributors, and this approach is easier to maintain.

### Standard Header (Go / C / Java)

Use the following header for all core source files:

```go
// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.
```

### Language-Specific Variants

#### Python

```python
# Copyright The kweaver.ai Authors.
#
# Licensed under the Apache License, Version 2.0.
# See the LICENSE file in the project root for details.
```

#### JavaScript / TypeScript

```ts
/**
 * Copyright The kweaver.ai Authors.
 *
 * Licensed under the Apache License, Version 2.0.
 * See the LICENSE file in the project root for details.
 */
```

#### Shell

```bash
#!/usr/bin/env bash
# Copyright The kweaver.ai Authors.
# Licensed under the Apache License, Version 2.0.
# See the LICENSE file in the project root for details.
```

#### HTML / XML

```html
<!--
  Copyright The kweaver.ai Authors.
  Licensed under the Apache License, Version 2.0.
  See the LICENSE file in the project root for details.
-->
```

### Derived or Forked Files (Optional)

If a file was originally derived from another project, you may add an origin note
after the license header (for key files only):

```go
// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.
//
// This file is derived from [original-project](https://github.com/org/repo)
```

This is optional but recommended for transparency and community trust.

### Scope

Headers are **recommended** for:

- core logic and business code
- public APIs and interfaces
- libraries and SDKs
- CLI tools and utilities

Headers are **optional** for:

- unit tests and test fixtures
- examples and demos
- generated files (protobuf, OpenAPI, etc.)
- configuration files (YAML, JSON, TOML)
- documentation files (Markdown, etc.)

### Why No Individual Author Names?

Following the practice of major open-source projects (Kubernetes, TensorFlow, etc.):

- **Git history** already provides a complete and accurate record of all contributors
- Individual author lists are **hard to maintain** and often become outdated
- Using "The kweaver.ai Authors" ensures **consistent attribution** across all files
- Contributors are recognized through the project's **CONTRIBUTORS** file and git log

### License Requirement

All repositories **must** include a `LICENSE` file containing the full text of
the Apache License, Version 2.0.

### Guiding Principle

> If a file is expected to be reused, forked, or maintained long-term,
> it deserves a clear and explicit header.

---

## 🏗 Development Setup

### Prerequisites

KWeaver Core is polyglot. You only need the toolchain(s) for the module(s) you touch:

- **Git** (always)
- **Java** (JDK 17+) and Maven for most ADP / decision-agent backend modules
- **Go** (1.23+) for `infra/oss-gateway-backend`, several CLIs and small services
- **Python** (3.11+) for `infra/mf-model-manager`, model / data utilities
- **Node.js** — kweaver CLIs require **22+** (per [`@kweaver-ai/kweaver-sdk` on npm](https://www.npmjs.com/package/@kweaver-ai/kweaver-sdk)). For `website/`, use `package.json` `engines` (e.g. >= 20).
- **Docker** + a Kubernetes (single-node K3s / kubeadm / Docker Desktop) for end-to-end testing
- **MariaDB 11.4+** (or DM8), **OpenSearch 2.x**, **Redis**, **Kafka** when running services that need them — usually provided by `deploy.sh`

Each module's `README.md` / `AGENTS.md` lists the exact prerequisites, build commands and dev-loop for that module — always read them first.

### Local Development

1. **Clone your fork of `kweaver-core`:**

   ```bash
   git clone https://github.com/YOUR_USERNAME/kweaver-core.git
   cd kweaver-core
   ```

2. **Add upstream remote:**

   ```bash
   git remote add upstream https://github.com/kweaver-ai/kweaver-core.git
   ```

3. **Pick the module you want to work on and follow its README.** For example:

   ```bash
   # Java module (Maven)
   cd adp/bkn/bkn-backend && mvn -DskipTests package

   # Go module
   cd infra/oss-gateway-backend && make build && make test

   # Python module
   cd infra/mf-model-manager && pip install -r requirements.txt && pytest

   # Docs site
   cd website && npm install && npm run start
   ```

4. **(Optional) Spin up a full cluster** to test end-to-end via `deploy/` (see [`help/en/install.md`](../help/en/install.md)):

   ```bash
   cd deploy
   ./deploy.sh kweaver-core install --minimum    # quick try
   # or full install: ./deploy.sh kweaver-core install
   ```

---

## 🐛 Reporting Security Issues

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, please open a private report via GitHub's built-in security advisory flow:

- [Report a vulnerability — kweaver-ai/kweaver-core](https://github.com/kweaver-ai/kweaver-core/security/advisories/new)

We will acknowledge receipt and work with you to address the issue. Please include reproduction steps, affected version (`git describe --tags`), and the impact you observed.

---

## ❓ Getting Help

- **Documentation**: Check the [README](README.md) and module-specific docs
- **Issues**: Search existing issues before creating a new one
- **Discussions**: Use GitHub Discussions for questions and ideas

---

## 📜 License

By contributing to KWeaver, you agree that your contributions will be licensed under the Apache License 2.0.

---

Thank you for contributing to KWeaver! 🎉
