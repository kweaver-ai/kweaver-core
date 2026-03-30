# Embedded Release Manifests Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Embed aggregate release manifests under `deploy/release-manifests/<version>/<product>.yaml` and make deploy commands auto-discover them from `--version`.

**Architecture:** Keep the manifest schema unchanged, but move from ad hoc mock files to a stable on-disk layout owned by `deploy`. The manifests are edited manually and discovered by product/version. `core`, `isf`, and `dip` auto-resolve the manifest path when `--version_file` is absent.

**Tech Stack:** Bash, shell tests

---

### Task 1: Add failing test for embedded manifest auto-discovery

**Files:**
- Modify: `deploy/scripts/tests/test-version-manifest.sh`

**Step 1: Write the failing test**

Add a test that expects:

- `resolve_embedded_release_manifest` to return `deploy/release-manifests/0.4.0/kweaver-core.yaml`
- `download_core` to auto-use that manifest when only `HELM_CHART_VERSION=0.4.0` is set

**Step 2: Run test to verify it fails**

Run: `bash deploy/scripts/tests/test-version-manifest.sh`

Expected: FAIL because the embedded manifest path and auto-discovery do not exist yet.

### Task 2: Create embedded manifest directory structure

**Files:**
- Create: `deploy/release-manifests/0.4.0/isf.yaml`
- Create: `deploy/release-manifests/0.4.0/kweaver-core.yaml`
- Create: `deploy/release-manifests/0.4.0/kweaver-dip.yaml`
- Create: `deploy/release-manifests/0.5.0/isf.yaml`
- Create: `deploy/release-manifests/0.5.0/kweaver-core.yaml`
- Create: `deploy/release-manifests/0.5.0/kweaver-dip.yaml`

**Step 1: Add the manual manifests**

Create and maintain the embedded manifests directly under `deploy/release-manifests/<version>/<product>.yaml`.

**Step 2: Verify the files exist**

Run: `find deploy/release-manifests -maxdepth 2 -type f | sort`

Expected: the embedded manifest tree is present.

### Task 3: Wire automatic embedded manifest resolution into deploy scripts

**Files:**
- Modify: `deploy/scripts/lib/common.sh`
- Modify: `deploy/scripts/services/core.sh`
- Modify: `deploy/scripts/services/isf.sh`
- Modify: `deploy/scripts/services/dip.sh`
- Modify: `deploy/deploy.sh`
- Modify: `deploy/README.md`

**Step 1: Add common helper**

Implement a helper that resolves:

`deploy/release-manifests/<version>/<product>.yaml`

when the file exists.

**Step 2: Auto-bind per-product manifest**

At the start of `install` and `download`, if `--version_file` is empty and `--version` is set, bind the embedded manifest path automatically.

**Step 3: Re-run tests**

Run: `bash deploy/scripts/tests/test-version-manifest.sh`

Expected: PASS

### Task 4: Verification

**Files:**
- Modify: `deploy/README.md`

**Step 1: Add usage examples**

Document:

- automatic embedded manifest resolution
- optional explicit `--version_file`
- generator command

**Step 2: Run syntax and behavior checks**

Run:

```bash
bash -n deploy/deploy.sh
bash -n deploy/scripts/lib/common.sh
bash -n deploy/scripts/services/core.sh
bash -n deploy/scripts/services/isf.sh
bash -n deploy/scripts/services/dip.sh
bash deploy/scripts/tests/test-version-manifest.sh
```

Expected: all pass
