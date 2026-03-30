# SQL Versioned Layout Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Make deploy SQL initialization version-aware by resolving SQL from `deploy/scripts/sql/<version>/<product>/` without losing the current SQL content.

**Architecture:** Move SQL from the flat shared directory into product-scoped version directories, add common helpers that resolve SQL paths from `product + version`, and update `isf`, `kweaver-core`, and `kweaver-dip` installers to execute SQL only when their versioned SQL directories exist.

**Tech Stack:** Bash, git history recovery, shell tests

---

### Task 1: Add the versioned SQL tree and regression test

**Files:**
- Create: `deploy/scripts/sql/0.4.0/isf/`
- Create: `deploy/scripts/sql/0.5.0/isf/`
- Create: `deploy/scripts/sql/0.4.0/kweaver-core/`
- Create: `deploy/scripts/sql/0.5.0/kweaver-core/`
- Create: `deploy/scripts/sql/0.4.0/kweaver-dip/`
- Modify: `deploy/scripts/tests/test-version-manifest.sh`

**Step 1: Write the failing test**

Add shell tests that expect:

- SQL path resolution for `isf`, `kweaver-core`, and `kweaver-dip`
- `kweaver-core 0.4.0` to enumerate only existing module directories
- `kweaver-dip` SQL initialization to skip when the directory is absent or empty

**Step 2: Run test to verify it fails**

Run: `bash deploy/scripts/tests/test-version-manifest.sh`

Expected: failure because SQL resolution helpers and product wiring do not exist yet.

### Task 2: Add SQL resolution helpers

**Files:**
- Modify: `deploy/scripts/lib/common.sh`

**Step 1: Write minimal implementation**

Add helpers to:

- Resolve `deploy/scripts/sql/<version>/<product>`
- Check whether a directory contains SQL files
- Enumerate existing `kweaver-core` module directories for one version
- Execute one SQL directory only when it exists and has SQL files

**Step 2: Run test to verify partial progress**

Run: `bash deploy/scripts/tests/test-version-manifest.sh`

Expected: helper-level assertions pass, but service scripts still fail until they consume the new SQL layout.

### Task 3: Wire SQL version resolution into product installers

**Files:**
- Modify: `deploy/scripts/services/isf.sh`
- Modify: `deploy/scripts/services/core.sh`
- Modify: `deploy/scripts/services/dip.sh`

**Step 1: Update ISF SQL initialization**

Resolve and execute `deploy/scripts/sql/<version>/isf/` when available.

**Step 2: Update KWeaver Core SQL initialization**

Resolve `deploy/scripts/sql/<version>/kweaver-core/` and initialize only modules whose directories exist there.

**Step 3: Update KWeaver DIP SQL initialization**

Resolve and execute `deploy/scripts/sql/<version>/kweaver-dip/` only when the directory exists and contains SQL files.

**Step 4: Run test to verify it passes**

Run: `bash deploy/scripts/tests/test-version-manifest.sh`

Expected: PASS

### Task 4: Populate versioned SQL content

**Files:**
- Create: `deploy/scripts/sql/0.4.0/isf/01-init-database.sql`
- Create: `deploy/scripts/sql/0.4.0/kweaver-core/...`
- Create: `deploy/scripts/sql/0.5.0/isf/01-init-database.sql`
- Create: `deploy/scripts/sql/0.5.0/kweaver-core/...`

**Step 1: Restore 0.4.0 SQL from git history**

Copy SQL content from commit `543b6adb200e2d10d672c4db7c33d46e639460ab` into the new `0.4.0` directories.

**Step 2: Preserve the current SQL as 0.5.0**

Move the current live SQL content into the new `0.5.0` directories without changing file contents.

**Step 3: Run syntax and behavior verification**

Run:

```bash
bash -n deploy/scripts/lib/common.sh
bash -n deploy/scripts/services/isf.sh
bash -n deploy/scripts/services/core.sh
bash -n deploy/scripts/services/dip.sh
bash deploy/scripts/tests/test-version-manifest.sh
```

Expected: all commands succeed and the test prints `PASS`.
