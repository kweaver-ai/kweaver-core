# Contributing Guide

[中文](CONTRIBUTING.zh.md) | English

Thank you for your interest in contributing to KWeaver! We welcome all forms of contributions, including bug fixes, feature proposals, documentation improvements, answering questions, and more.

Please read this guide before submitting contributions to ensure consistent processes and standardized submissions.

---

## 🏗 Sub-Projects

KWeaver is an open-source ecosystem consisting of multiple sub-projects. Please navigate to the corresponding repository based on the component you want to contribute to:

| Sub-Project | Description | Repository |
| --- | --- | --- |
| **AI Store** | AI application and component marketplace | [kweaver-ai/dip](https://github.com/kweaver-ai/dip) |
| **Studio** | DIP Studio - Visual development and management interface | [kweaver-ai/studio](https://github.com/kweaver-ai/studio) |
| **Decision Agent** | Decision Agent - Intelligent decision agent | [kweaver-ai/decision-agent](https://github.com/kweaver-ai/decision-agent) |
| **ADP** | AI Data Platform - Including Ontology Engine, ContextLoader, and VEGA data virtualization engine | [kweaver-ai/adp](https://github.com/kweaver-ai/adp) |
| **Operator Hub** | Operator Platform - Operator management and orchestration | [kweaver-ai/operator-hub](https://github.com/kweaver-ai/operator-hub) |
| **Sandbox** | Sandbox runtime environment | [kweaver-ai/sandbox](https://github.com/kweaver-ai/sandbox) |

> **Note**: Each sub-project has its own README and contribution guidelines. Please refer to the specific repository for detailed setup and development instructions.

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
  - Go version (e.g., Go 1.23.0)
  - OS (Windows/Linux/macOS)
  - Database version (MariaDB 11.4+ / DM8)
  - OpenSearch version (if applicable)
  - Module affected (e.g., ADP, Decision Agent, DIP Studio)

- **Reproduction Steps**: Clear, step-by-step instructions to reproduce the issue

- **Expected vs Actual Behavior**: What should happen vs what actually happens

- **Error Logs/Screenshots**: Include relevant error messages, stack traces, or screenshots

- **Minimal Reproducible Code (MRC)**: A minimal code example that demonstrates the issue

**Example Bug Report Template:**

```markdown
**Environment:**
- Go: 1.23.0
- OS: Linux Ubuntu 22.04
- Module: ADP
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
```

**Branch Naming Convention:**

| Branch Type | Format | Description | Example |
| --- | --- | --- | --- |
| Feature | `feature/*` | New feature development | `feature/add-oauth-support` |
| Fix | `fix/*` | Bug fixes | `fix/memory-leak-in-loader` |
| Release | `release/x.x.x` | Release preparation | `release/1.2.0` |
| Docs | `docs/*` | Documentation changes | `docs/update-api-reference` |
| Refactor | `refactor/*` | Code refactoring | `refactor/simplify-auth-flow` |
| Test | `test/*` | Adding or updating tests | `test/add-unit-tests-for-loader` |

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

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...
```

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
> - If you prefer to track the upstream repository, you can add it: `git remote add upstream https://github.com/kweaver-ai/kweaver.git`

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

- Go 1.23.0 or higher
- MariaDB 11.4+ or DM8
- OpenSearch 2.x (optional, for full functionality)
- Git

### Local Development

1. **Clone your fork:**

```bash
git clone https://github.com/YOUR_USERNAME/kweaver.git
cd kweaver
```

1. **Add upstream remote:**

```bash
git remote add upstream https://github.com/kweaver-ai/kweaver.git
```

1. **Set up the development environment:**

```bash
# Navigate to the module you want to work on
cd <module-directory>/server

# Download dependencies
go mod download

# Run the service
go run main.go
```

1. **Run tests:**

```bash
go test ./...
```

---

## 🐛 Reporting Security Issues

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, please report them via:

- Email: [Security contact email]
- Internal security reporting system

We will acknowledge receipt and work with you to address the issue.

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
