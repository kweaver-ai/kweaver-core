# Release Guidelines

[中文](RELEASE.zh.md) | English

This document defines the version management, branching strategy, and release process for the KWeaver project.

---

## 📋 Table of Contents

- [Branching Strategy](#-branching-strategy)
- [Versioning](#-versioning)
- [Release Process](#-release-process)
- [Changelog Guidelines](#-changelog-guidelines)
- [Backport Strategy](#-backport-strategy)

---

## 🌿 Branching Strategy

### Trunk-based Development

KWeaver follows the **Trunk-based Development** model with these core principles:

| Principle | Description |
| --- | --- |
| **main is always releasable** | The `main` branch is always in a releasable state; a release can be created from `main` at any time |
| **Short-lived branches** | Feature and fix branches should be short-lived, typically no more than 2-3 days |
| **Small PRs** | Each PR should focus on a single responsibility for easier review and rollback |

### Branch Naming Convention

| Branch Type | Format | Description | Example |
| --- | --- | --- | --- |
| Main branch | `main` | Always-releasable trunk | `main` |
| Feature branch | `feature/*` | New feature development | `feature/add-oauth-support` |
| Fix branch | `fix/*` | Bug fixes | `fix/memory-leak-in-loader` |
| Release branch | `release/x.x.x` | Release preparation | `release/1.2.0` |
| Support branch | `support/x.y` | Maintenance branch for LTS versions | `support/1.2` |

### Branch Lifecycle

```text
main ─────────────────────────────────────────────────────────►
       │                    │                    │
       │ feature/foo        │ fix/bar            │
       └──────┬─────────────┴──────┬─────────────┘
              │                    │
              ▼ (merge & delete)   ▼ (merge & delete)
```

**Best Practices:**

- ✅ Create branches from `main`
- ✅ Frequently rebase to stay in sync with `main`
- ✅ Delete branches immediately after merging
- ❌ Avoid long-lived feature branches
- ❌ Avoid merging between feature branches

---

## 🏷️ Versioning

### Semantic Versioning

KWeaver follows [Semantic Versioning 2.0.0](https://semver.org/):

```
vMAJOR.MINOR.PATCH[-PRERELEASE]
```

| Version | Meaning | When to Increment |
| --- | --- | --- |
| **MAJOR** | Major version | Incompatible API changes |
| **MINOR** | Minor version | Backward-compatible new features |
| **PATCH** | Patch version | Backward-compatible bug fixes |

### Pre-release Versions

| Tag Format | Description | Example |
| --- | --- | --- |
| `-alpha.N` | Internal testing, incomplete features | `v1.2.0-alpha.1` |
| `-beta.N` | Public testing, complete but may have bugs | `v1.2.0-beta.1` |
| `-rc.N` | Release candidate, ready for release | `v1.2.0-rc.1` |

### Tag Conventions

| Rule | Description |
| --- | --- |
| Prefix | Must start with `v` |
| Format | `vX.Y.Z` or `vX.Y.Z-rc.N` |
| Signing | GPG-signed tags are recommended |

**Correct Examples:**

```bash
# Release version
git tag -a v1.0.0 -m "Release v1.0.0"

# Pre-release version
git tag -a v1.1.0-rc.1 -m "Release candidate 1 for v1.1.0"

# Signed tag
git tag -s v1.0.0 -m "Release v1.0.0"
```

**Incorrect Examples:**

```bash
# ❌ Missing v prefix
git tag 1.0.0

# ❌ Non-standard pre-release format
git tag v1.0.0-RC1
git tag v1.0.0.rc.1
```

---

## 🚀 Release Process

### Automated Releases

KWeaver uses a **tag-triggered** automated release process:

```text
Developer pushes tag  →  CI detects tag  →  Run tests  →  Build artifacts  →  Publish Release
```

CI will automatically perform the following:

1. ✅ Run full test suite
2. ✅ Build binaries for all platforms
3. ✅ Build Docker images
4. ✅ Generate Release Notes
5. ✅ Publish to GitHub Releases
6. ✅ Push images to Container Registry

### Release Process (Release With Freeze / RC Flow)

KWeaver uses the **Release With Freeze** model. All releases require an RC (Release Candidate) validation cycle.

```text
main ─────────┬─────────────────────────────────────────────────►
              │                                        ▲
              │ create release branch                  │ merge back to main
              ▼                                        │
         release/1.2.0 ──► rc.1 ──► rc.2 ──► v1.2.0 ──┘
              │              │        │         │
              │          (fix issues) (fix issues)
              │              │        │         │
              └──────────────┴────────┴─────────┘
                   only bug fixes and release-related changes
```

### Steps

#### 1. Create Release Branch

```bash
# Create release branch from main
git checkout main
git pull origin main
git checkout -b release/1.2.0

# Push release branch
git push origin release/1.2.0
```

#### 2. Code Freeze

Once the release branch is created, it enters **code freeze** state:

| Allowed ✅ | Forbidden ❌ |
| --- | --- |
| Bug fixes | New features |
| Documentation updates | Refactoring |
| Version number updates | Performance optimizations (unless fixing issues) |
| Configuration adjustments | Dependency upgrades (unless security fixes) |

#### 3. Publish RC Versions

```bash
# Publish first RC on release branch
git checkout release/1.2.0
git tag -a v1.2.0-rc.1 -m "Release candidate 1 for v1.2.0"
git push origin v1.2.0-rc.1

# If issues are fixed, publish subsequent RCs
git tag -a v1.2.0-rc.2 -m "Release candidate 2 for v1.2.0"
git push origin v1.2.0-rc.2
```

#### 4. RC Validation

- Deploy RC version to test/staging environment
- Execute full test suite
- Perform User Acceptance Testing (UAT)
- Collect feedback and fix discovered issues

#### 5. Publish Final Release

```bash
# After RC validation passes, publish final release
git checkout release/1.2.0
git tag -a v1.2.0 -m "Release v1.2.0"
git push origin v1.2.0
```

#### 6. Merge Back to main

```bash
# Merge release branch back to main
git checkout main
git pull origin main
git merge release/1.2.0 --no-ff -m "Merge release/1.2.0 into main"
git push origin main

# Delete release branch (optional)
git branch -d release/1.2.0
git push origin --delete release/1.2.0
```

---

### Verify Release

- Check the [GitHub Releases](https://github.com/kweaver-ai/kweaver/releases) page
- Verify artifact downloads and integrity
- Confirm Docker images are pullable

### Release Checklist

Before creating a release tag, confirm:

- [ ] All planned features are merged
- [ ] All tests pass
- [ ] CHANGELOG.md is updated
- [ ] Version number follows semantic versioning
- [ ] Documentation is synchronized
- [ ] Breaking changes are documented
- [ ] All RC versions have been validated
- [ ] Release branch has been merged back to main

---

## 📄 Changelog Guidelines

### Keep a Changelog

KWeaver follows the [Keep a Changelog](https://keepachangelog.com/) format.

### File Format

```markdown
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added
- New features

### Changed
- Changes to existing functionality

### Deprecated
- Features to be removed in future versions

### Removed
- Removed features

### Fixed
- Bug fixes

### Security
- Security-related fixes

## [1.1.0] - 2025-01-09

### Added
- Add OAuth 2.0 authentication support (#123)
- Add batch import functionality (#456)

### Fixed
- Fix memory leak issue (#789)

## [1.0.0] - 2024-12-01

### Added
- Initial release
```

### Change Types

| Type | Description |
| --- | --- |
| **Added** | New features |
| **Changed** | Changes to existing functionality |
| **Deprecated** | Features to be removed in next major version |
| **Removed** | Removed features |
| **Fixed** | Bug fixes |
| **Security** | Security vulnerability fixes |

### Writing Guidelines

- ✅ Link each entry to a PR or Issue number
- ✅ Categorize by change type
- ✅ Use user-friendly language
- ✅ Highlight breaking changes in bold
- ❌ Don't include internal refactoring details (unless user-facing)
- ❌ Don't include CI/test-related changes

---

## 🔄 Backport Strategy

### When to Backport

Backports only apply to **Long-Term Support (LTS)** version maintenance branches:

| Scenario | Backport? |
| --- | --- |
| Security vulnerability fixes | ✅ Required |
| Critical bug fixes | ✅ Recommended |
| General bug fixes | ⚠️ Case by case |
| New features | ❌ No backport |
| Refactoring | ❌ No backport |

### Support Branch Naming

```
support/x.y
```

Example: `support/1.2` for maintaining the 1.2.x version series.

### Backport Process

#### 1. Fix on main First

```bash
# Create fix branch on main
git checkout main
git checkout -b fix/security-issue-123

# Commit the fix
git commit -m "fix(auth): patch security vulnerability CVE-2025-XXXX"

# Merge to main
# Record the merged commit hash and PR number
```

#### 2. Cherry-pick to Support Branch

```bash
# Switch to support branch
git checkout support/1.2

# Cherry-pick the fix commit
git cherry-pick -x <commit-hash>

# The -x flag automatically adds cherry-pick source information
```

#### 3. Record Backport Information

Add the original PR link in the commit message:

```bash
git commit --amend

# Add to commit message:
# (cherry picked from commit abc1234)
# Backport of #123
```

#### 4. Create Patch Version

```bash
# Create patch version on support branch
git tag -a v1.2.1 -m "Release v1.2.1 (security patch)"
git push origin v1.2.1
```

### Backport Checklist

- [ ] Fix has been verified on `main` branch
- [ ] Used `cherry-pick -x` to preserve source information
- [ ] Commit message includes original PR link
- [ ] Support branch CHANGELOG is updated
- [ ] Patch version number is correctly incremented

---

## 📚 Resources

- [Semantic Versioning](https://semver.org/)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [Keep a Changelog](https://keepachangelog.com/)
- [Trunk Based Development](https://trunkbaseddevelopment.com/)

---

*Last updated: 2025-01-09*
