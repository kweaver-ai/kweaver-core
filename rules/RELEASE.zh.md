# 版本发布规范

中文 | [English](RELEASE.md)

本文档定义了 KWeaver 项目的版本管理、分支策略和发布流程规范。

---

## 📋 目录

- [分支策略](#-分支策略)
- [版本号规范](#-版本号规范)
- [发布流程](#-发布流程)
- [Changelog 规范](#-changelog-规范)
- [Backport 策略](#-backport-策略)

---

## 🌿 分支策略

### Trunk-based Development

KWeaver 采用 **Trunk-based Development** 模型，核心原则：

| 原则 | 说明 |
| --- | --- |
| **main 永远可发布** | `main` 分支始终保持可发布状态，任何时候都可以从 `main` 创建 release |
| **短生命周期分支** | 功能分支和修复分支应尽量短小，通常不超过 2-3 天 |
| **小 PR** | 每个 PR 应聚焦单一职责，便于 review 和回滚 |

### 分支命名规范

| 分支类型 | 命名格式 | 说明 | 示例 |
| --- | --- | --- | --- |
| 主分支 | `main` | 永远可发布的主干 | `main` |
| 功能分支 | `feature/*` | 新功能开发 | `feature/add-oauth-support` |
| 修复分支 | `fix/*` | Bug 修复 | `fix/memory-leak-in-loader` |
| 发布分支 | `release/x.x.x` | 发布准备 | `release/1.2.0` |
| 支持分支 | `support/x.y` | 长期支持版本的维护分支 | `support/1.2` |

### 分支生命周期

```text
main ─────────────────────────────────────────────────────────►
       │                    │                    │
       │ feature/foo        │ fix/bar            │
       └──────┬─────────────┴──────┬─────────────┘
              │                    │
              ▼ (merge & delete)   ▼ (merge & delete)
```

**最佳实践：**

- ✅ 从 `main` 创建分支
- ✅ 频繁 rebase 保持与 `main` 同步
- ✅ 合并后立即删除分支
- ❌ 避免长期存在的功能分支
- ❌ 避免分支之间相互合并

---

## 🏷️ 版本号规范

### 语义化版本 (Semantic Versioning)

KWeaver 遵循 [Semantic Versioning 2.0.0](https://semver.org/lang/zh-CN/) 规范：

```
vMAJOR.MINOR.PATCH[-PRERELEASE]
```

| 版本号 | 含义 | 何时递增 |
| --- | --- | --- |
| **MAJOR** | 主版本号 | 不兼容的 API 变更 |
| **MINOR** | 次版本号 | 向后兼容的功能新增 |
| **PATCH** | 修订号 | 向后兼容的 Bug 修复 |

### 预发布版本

| 标签格式 | 说明 | 示例 |
| --- | --- | --- |
| `-alpha.N` | 内部测试版，功能不完整 | `v1.2.0-alpha.1` |
| `-beta.N` | 公开测试版，功能完整但可能有 Bug | `v1.2.0-beta.1` |
| `-rc.N` | 发布候选版，准备正式发布 | `v1.2.0-rc.1` |

### Tag 规范

| 规则 | 说明 |
| --- | --- |
| 前缀 | 必须以 `v` 开头 |
| 格式 | `vX.Y.Z` 或 `vX.Y.Z-prerelease.N` |
| 签名 | 推荐使用 GPG 签名 tag |

**正确示例：**

```bash
# 正式版本
git tag -a v1.0.0 -m "Release v1.0.0"

# 预发布版本
git tag -a v1.1.0-rc.1 -m "Release candidate 1 for v1.1.0"

# 带签名的 tag
git tag -s v1.0.0 -m "Release v1.0.0"
```

**错误示例：**

```bash
# ❌ 缺少 v 前缀
git tag 1.0.0

# ❌ 非标准预发布格式
git tag v1.0.0-RC1
git tag v1.0.0.rc.1
```

---

## 🚀 发布流程

### 自动化发布

KWeaver 采用 **Tag 触发** 的自动化发布流程：

```text
开发者 push tag  →  CI 检测到 tag  →  运行测试  →  构建制品  →  发布 Release
```

CI 将自动完成以下步骤：

1. ✅ 运行完整测试套件
2. ✅ 构建各平台二进制文件
3. ✅ 构建 Docker 镜像
4. ✅ 生成 Release Notes
5. ✅ 发布到 GitHub Releases
6. ✅ 推送镜像到 Container Registry

### 发布流程 (Release With Freeze / RC Flow)

KWeaver 采用 **冻结发布** 模式，所有版本发布都需要经过 RC（Release Candidate）验证周期。

```text
main ─────────┬─────────────────────────────────────────────────►
              │                                        ▲
              │ 创建 release 分支                       │ 合并回 main
              ▼                                        │
         release/1.2.0 ──► rc.1 ──► rc.2 ──► v1.2.0 ──┘
              │              │        │         │
              │         (修复问题) (修复问题)    │
              │              │        │         │
              └──────────────┴────────┴─────────┘
                    仅允许 bug fix 和发布相关更改
```

### 步骤

#### 1. 创建 Release 分支

```bash
# 从 main 创建 release 分支
git checkout main
git pull origin main
git checkout -b release/1.2.0

# 推送 release 分支
git push origin release/1.2.0
```

#### 2. 代码冻结 (Code Freeze)

Release 分支创建后进入**代码冻结**状态：

| 允许 ✅ | 禁止 ❌ |
| --- | --- |
| Bug 修复 | 新功能 |
| 文档更新 | 重构 |
| 版本号更新 | 性能优化（除非修复问题） |
| 配置调整 | 依赖升级（除非修复安全问题） |

#### 3. 发布 RC 版本

```bash
# 在 release 分支上发布第一个 RC
git checkout release/1.2.0
git tag -a v1.2.0-rc.1 -m "Release candidate 1 for v1.2.0"
git push origin v1.2.0-rc.1

# 如有问题修复后，发布后续 RC
git tag -a v1.2.0-rc.2 -m "Release candidate 2 for v1.2.0"
git push origin v1.2.0-rc.2
```

#### 4. RC 验证

- 在测试/预发布环境部署 RC 版本
- 执行完整的测试用例
- 进行用户验收测试（UAT）
- 收集反馈并修复发现的问题

#### 5. 发布正式版本

```bash
# RC 验证通过后，发布正式版本
git checkout release/1.2.0
git tag -a v1.2.0 -m "Release v1.2.0"
git push origin v1.2.0
```

#### 6. 合并回 main

```bash
# 将 release 分支合并回 main
git checkout main
git pull origin main
git merge release/1.2.0 --no-ff -m "Merge release/1.2.0 into main"
git push origin main

# 删除 release 分支（可选）
git branch -d release/1.2.0
git push origin --delete release/1.2.0
```

---

### 验证发布

- 检查 [GitHub Releases](https://github.com/kweaver-ai/kweaver-core/releases) 页面
- 验证制品下载和完整性
- 确认 Docker 镜像可拉取

### 发布检查清单

在创建 release tag 之前，请确认：

- [ ] 所有计划功能已合并
- [ ] 所有测试通过
- [ ] CHANGELOG.md 已更新
- [ ] 版本号符合语义化版本规范
- [ ] 文档已同步更新
- [ ] Breaking Changes 已在文档中说明
- [ ] 所有 RC 版本已验证通过
- [ ] Release 分支已合并回 main

---

## 📄 Changelog 规范

### Keep a Changelog

KWeaver 遵循 [Keep a Changelog](https://keepachangelog.com/zh-CN/) 规范。

### 文件格式

```markdown
# Changelog

本项目所有重要变更都将记录在此文件中。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/)，
本项目遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

## [Unreleased]

### Added
- 新增的功能

### Changed
- 变更的功能

### Deprecated
- 即将废弃的功能

### Removed
- 已移除的功能

### Fixed
- 修复的 Bug

### Security
- 安全相关的修复

## [1.1.0] - 2025-01-09

### Added
- 新增 OAuth 2.0 认证支持 (#123)
- 新增批量导入功能 (#456)

### Fixed
- 修复内存泄漏问题 (#789)

## [1.0.0] - 2024-12-01

### Added
- 首次发布
```

### 变更类型

| 类型 | 说明 |
| --- | --- |
| **Added** | 新增功能 |
| **Changed** | 功能变更 |
| **Deprecated** | 即将废弃（下个主版本移除） |
| **Removed** | 已移除功能 |
| **Fixed** | Bug 修复 |
| **Security** | 安全漏洞修复 |

### 编写规范

- ✅ 每条记录关联 PR 或 Issue 编号
- ✅ 按变更类型分类
- ✅ 使用用户可理解的语言描述
- ✅ Breaking Changes 放在显著位置并加粗
- ❌ 不要包含内部重构细节（除非影响用户）
- ❌ 不要包含 CI/测试相关变更

---

## 🔄 Backport 策略

### 何时需要 Backport

Backport 仅适用于 **长期支持 (LTS)** 版本的维护分支：

| 场景 | 是否 Backport |
| --- | --- |
| 安全漏洞修复 | ✅ 必须 |
| 严重 Bug 修复 | ✅ 推荐 |
| 一般 Bug 修复 | ⚠️ 视情况 |
| 新功能 | ❌ 不 Backport |
| 重构 | ❌ 不 Backport |

### 支持分支命名

```
support/x.y
```

例如：`support/1.2` 用于维护 1.2.x 系列版本。

### Backport 流程

#### 1. 在 main 上先修复

```bash
# 在 main 上创建修复分支
git checkout main
git checkout -b fix/security-issue-123

# 提交修复
git commit -m "fix(auth): patch security vulnerability CVE-2025-XXXX"

# 合并到 main
# 记录合并后的 commit hash 和 PR 编号
```

#### 2. Cherry-pick 到支持分支

```bash
# 切换到支持分支
git checkout support/1.2

# Cherry-pick 修复提交
git cherry-pick -x <commit-hash>

# -x 参数会自动添加 cherry-pick 来源信息
```

#### 3. 记录 Backport 信息

在 commit message 中添加原始 PR 链接：

```bash
git commit --amend

# 在 commit message 中添加：
# (cherry picked from commit abc1234)
# Backport of #123
```

#### 4. 创建 Patch 版本

```bash
# 在支持分支上创建补丁版本
git tag -a v1.2.1 -m "Release v1.2.1 (security patch)"
git push origin v1.2.1
```

### Backport 检查清单

- [ ] 修复已在 `main` 分支验证
- [ ] 使用 `cherry-pick -x` 保留来源信息
- [ ] commit message 包含原始 PR 链接
- [ ] 支持分支的 CHANGELOG 已更新
- [ ] Patch 版本号正确递增

---

## 📚 相关资源

- [Semantic Versioning](https://semver.org/lang/zh-CN/)
- [Conventional Commits](https://www.conventionalcommits.org/zh-hans/)
- [Keep a Changelog](https://keepachangelog.com/zh-CN/)
- [Trunk Based Development](https://trunkbaseddevelopment.com/)

---

*最后更新：2025-01-09*
