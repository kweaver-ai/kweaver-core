# Team R&D Workflow

[中文](WORKFLOW.zh.md) | English

This document defines KWeaver's team R&D collaboration standards, covering Issue management, Feature tracking, design documentation, and team notification processes. Every rule includes concrete steps and file paths to ensure practical execution.

---

## 📋 Table of Contents

- [Issue Management](#-issue-management)
- [Feature Tracking: Issue → Branch → Design Doc](#-feature-tracking-issue--branch--design-doc)
- [Design Document Specification](#-design-document-specification)
- [PR and Merge Process](#-pr-and-merge-process)
- [Email Notification Process](#-email-notification-process)

---

## 🗂 Issue Management

### Issue Types and Labels

| Type | Label | Design Doc Required | Description |
| --- | --- | --- | --- |
| Bug Report | `type: bug` | No (optional for complex bugs) | Functional defects or unexpected behavior |
| Feature Request | `type: feature` | **Required** | New features or enhancements |
| Task | `type: task` | Recommended | Engineering tasks, research, refactoring |
| Documentation | `type: docs` | No | Missing or improvable documentation |

### Priority Labels

| Priority | Label | Response SLA |
| --- | --- | --- |
| Critical | `priority: critical` | Respond within 24 hours; must complete in current Sprint |
| High | `priority: high` | Must complete in current Sprint |
| Medium | `priority: medium` | Planned for next 1–2 Sprints |
| Low | `priority: low` | Future backlog; no fixed deadline |

### Issue Lifecycle

```text
Open → Triaged → In Progress → In Review → Done
         │                                    │
         └─────────── Closed (Won't Fix) ─────┘
```

| State | Action |
| --- | --- |
| `Open` | Created; awaiting triage |
| `Triaged` | Priority, Assignee, and Milestone assigned |
| `In Progress` | Branch and design doc created; development started (**update tracking comment in Issue**, see next section) |
| `In Review` | PR submitted; awaiting Code Review |
| `Done` | PR merged; Issue closed automatically or manually |
| `Closed (Won't Fix)` | Decided not to implement; reason stated in a comment |

### Issue Triage Rules

- All Issues must be triaged (Assignee + Priority + Milestone) within **2 business days** of creation
- Cross-module Issues must `@mention` the responsible module owner with expected outcome
- Issues with no progress for 30+ days must be re-triaged or closed

---

## 🔗 Feature Tracking: Issue → Branch → Design Doc

This is the core of this specification. Every Feature (`type: feature`) must follow the complete steps below from Issue creation to code merge. The **Issue tracking comment** is the single source of truth that links all three artifacts together.

### Complete Workflow

```text
1. Create Issue
      │
      ▼
2. Triage: assign Assignee, Priority, Milestone
      │
      ▼
3. Create design doc  →  docs/design/{module}/features/{issue-id}-{desc}.md
      │
      ▼
4. Create branch  →  feature/{issue-id}-{desc}
      │
      ▼
5. Post tracking comment in Issue (branch + design doc link)
      │
      ▼
6. Develop + update design doc
      │
      ▼
7. Submit PR (linked to Issue and design doc)
      │
      ▼
8. Code Review (includes doc review)
      │
      ▼
9. Merge → update design doc status to "implemented"
```

### Step 3: Create the Design Document

**File path rule:**

```
docs/design/{module}/features/{issue-id}-{short-desc}.md
```

| Placeholder | Description | Example |
| --- | --- | --- |
| `{module}` | Module name, matching the code directory | `auth`, `knowledge-graph`, `data-agent` |
| `{issue-id}` | GitHub Issue number (without `#`) | `123` |
| `{short-desc}` | Kebab-case summary of the Issue title; 5 words max | `add-oauth-support` |

**Examples:**

```
docs/design/auth/features/123-add-oauth-support.md
docs/design/knowledge-graph/features/456-batch-import-nodes.md
docs/design/data-agent/features/789-streaming-response.md
```

**Directory structure:**

```
docs/
└── design/
    ├── auth/
    │   └── features/
    │       └── 123-add-oauth-support.md
    ├── knowledge-graph/
    │   └── features/
    │       └── 456-batch-import-nodes.md
    └── data-agent/
        ├── features/
        │   └── 789-streaming-response.md
        └── adr/                          ← Architecture Decision Records
            └── 0001-use-opensearch.md
```

### Step 4: Create the Branch

**Branch naming rule:**

```
feature/{issue-id}-{short-desc}
```

The `{issue-id}` and `{short-desc}` in the branch name **must match** the design document filename exactly:

```bash
# Examples
git checkout -b feature/123-add-oauth-support
git checkout -b feature/456-batch-import-nodes
```

### Step 5: Post Tracking Comment in the Issue

After starting development, post one tracking comment in the Issue. Update this same comment as the work progresses — do not create multiple tracking comments.

```markdown
## 📌 Development Tracking

| Item | Value |
| --- | --- |
| **Branch** | `feature/123-add-oauth-support` |
| **Design Doc** | [docs/design/auth/features/123-add-oauth-support.md](../docs/design/auth/features/123-add-oauth-support.md) |
| **Status** | In Progress |
| **Assignee** | @username |
| **ETA** | YYYY-MM-DD |
```

> **Note**: After the PR is merged, update this comment's Status to `Done` and add the PR link.

---

## 📄 Design Document Specification

### Document Frontmatter (Metadata Header)

Every design document must begin with a YAML frontmatter block recording key metadata:

```markdown
---
issue: "#123"
branch: "feature/123-add-oauth-support"
module: "auth"
status: "draft"          # draft | in-review | approved | implemented
author: "@username"
created: "2026-03-16"
pr: ""                   # fill in after PR is merged, e.g. "#456"
---
```

| Field | Required | Description |
| --- | --- | --- |
| `issue` | Yes | Linked GitHub Issue number |
| `branch` | Yes | Corresponding development branch |
| `module` | Yes | Module this feature belongs to |
| `status` | Yes | Document/feature status; see values below |
| `author` | Yes | Primary assignee |
| `created` | Yes | Document creation date |
| `pr` | No | Fill in after PR is merged |

**Status values:**

| Value | Description |
| --- | --- |
| `draft` | Design in progress; not yet reviewed |
| `in-review` | PR submitted; under review |
| `approved` | Review passed; ready to implement |
| `implemented` | Merged; feature is live |

### Design Document Template

````markdown
---
issue: "#{issue-id}"
branch: "feature/{issue-id}-{short-desc}"
module: "{module}"
status: "draft"
author: "@username"
created: "YYYY-MM-DD"
pr: ""
---

# Feature #{issue-id}: {Feature Title}

## Background and Goals

Describe the feature background, user pain points, and the goals of this development effort.

## Design

### Summary

Brief overview of the overall approach.

### API Changes (if any)

```http
POST /api/v1/{endpoint}
Content-Type: application/json

{
  "field": "value"
}
```

Response:

```json
{
  "id": "xxx",
  "status": "ok"
}
```

### Database Changes (if any)

Describe new or modified tables/columns and provide the migration script path.

### Key Flow

Describe the core logic flow in prose or pseudocode.

## Acceptance Criteria

- [ ] Functional criterion 1
- [ ] Functional criterion 2
- [ ] Test coverage meets requirements

## Test Strategy

Describe the unit tests, integration tests, or AT cases to be added.

## Impact Analysis

- **Backward compatibility**: Yes/No — explain
- **Dependency changes**: Yes/No — explain
- **Performance impact**: explain (omit if none)

## References

- Related Issue/PR links
- Related documentation links
````

### Bug Analysis Document (Optional)

For complex bugs involving multiple modules or requiring root cause analysis, an analysis document can be created at:

```
docs/design/{module}/bugs/{issue-id}-{short-desc}.md
```

Frontmatter uses the same format; the body should include: problem description, root cause analysis, fix approach, and verification steps.

### ADR (Architecture Decision Record)

For Features involving significant design decisions, the decision should be documented in the design doc and optionally archived as an ADR:

**Path:** `docs/design/{module}/adr/NNNN-{short-title}.md`

**File naming:** Sequence starts at `0001`, counted independently per module.

```markdown
---
number: "0001"
module: "{module}"
status: "accepted"       # proposed | accepted | deprecated | superseded
date: "YYYY-MM-DD"
related-issue: "#123"
---

# ADR-{module}-0001: {Decision Title}

## Context

Describe the background and constraints driving this decision.

## Decision

We will adopt ___.

## Rationale

- Reason 1
- Reason 2

## Consequences

**Positive:**
- ...

**Negative (trade-offs):**
- ...

## Alternatives Considered

Describe alternatives that were considered but not chosen, and why.
```

---

## 🔀 PR and Merge Process

### PR Description Template

The PR description must include the following to complete the three-way link: Issue → Branch → Design Doc.

```markdown
## Description

Brief summary of what changed and why.

## Links

| Item | Value |
| --- | --- |
| **Issue** | Closes #123 |
| **Design Doc** | [docs/design/auth/features/123-add-oauth-support.md](../docs/design/auth/features/123-add-oauth-support.md) |
| **Branch** | `feature/123-add-oauth-support` |

## Type of Change

- [ ] Bug fix
- [ ] New feature
- [ ] Documentation update
- [ ] Refactoring

## Testing

Describe how to verify this change (test commands, manual steps, etc.).

## Pre-Merge Checklist

- [ ] Design doc updated (status changed to `in-review`; `pr` field filled in)
- [ ] CHANGELOG.md updated (under `[Unreleased]` section)
- [ ] API documentation updated (if API changed)
- [ ] Tests added/updated; all pass locally
- [ ] No breaking changes, or breaking changes are noted in CHANGELOG
```

### Documentation Checks in Code Review

Reviewers must confirm:

- [ ] Design doc is consistent with the actual implementation
- [ ] API documentation is up to date
- [ ] CHANGELOG.md records user-facing changes
- [ ] Design doc frontmatter: `pr` field is filled in and `status` is `in-review`

### Post-Merge Actions

After the PR is merged, the merger or Assignee must:

1. Update the design doc's `status` to `implemented`
2. Update the Issue tracking comment: set Status to `Done` and add the PR link

---

## 📧 Email Notification Process

### Notification Triggers

| Event | Recipients | Triggered By |
| --- | --- | --- |
| Official Release | All team members + external subscribers | CI automation / manual |
| RC Release | Internal testing team | CI automation / manual |
| Code freeze | All engineering members | Manual (Release Manager) |
| Critical bug fix | Affected module owners + testing team | Manual |
| Sprint start | Sprint participants | Manual (project lead) |

### Email Templates

#### Release Announcement

```
Subject: [KWeaver] vX.Y.Z Official Release

KWeaver vX.Y.Z has been officially released!

## Key Changes

### ✨ Added
- Feature description (#IssueNumber, design doc link)

### 🐛 Fixed
- Fix description (#IssueNumber)

### ⚠️ Breaking Changes (if any)
- Describe the change and migration path

## Download
- GitHub Releases: <link>
- Docker: docker pull kweaver/kweaver:vX.Y.Z

## Full Changelog
<CHANGELOG link>

Thank you to all contributors!
```

#### RC Release Notification

```
Subject: [KWeaver] vX.Y.Z-rc.N Test Release — Feedback Requested

KWeaver vX.Y.Z-rc.N has been released. Please validate and provide feedback.

## Test Scope
- Change list (with design doc links)

## Known Issues
- (if any)

## How to Submit Feedback
Open a GitHub Issue and add the label: milestone: vX.Y.Z

## Feedback Deadline
YYYY-MM-DD
```

#### Code Freeze Notification

```
Subject: [KWeaver] vX.Y.Z Code Freeze Notice

The vX.Y.Z release branch has been created. Code freeze begins YYYY-MM-DD.

## Rules During Freeze
- ✅ Allowed: Bug fixes, documentation updates, version bumps
- ❌ Not allowed: New features, refactoring, dependency upgrades

## Requesting an Exception
Reply to this email with justification. Approval required from Release Manager.

## Target Release Date
YYYY-MM-DD
```

### Sending Guidelines

| Item | Requirement |
| --- | --- |
| Delivery method | Team mailing list; major Releases may use CI automation |
| Subject prefix | Always use `[KWeaver]` for easy filtering and archiving |
| Language | English; Chinese version may be added for domestic team |
| CC | **All emails must CC the QA lead**; Release announcements must also CC the project lead |
| Attachments | No attachments; use links to reference documents or artifacts |

---

## 📚 Related Resources

- [Contributing Guidelines](CONTRIBUTING.md)
- [Release Guidelines](RELEASE.md)
- [Testing Guidelines](TESTING.md)
- [Architecture Guidelines](ARCHITECTURE.md)

---

*Last updated: 2026-03-16*
