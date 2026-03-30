# Deploy Release Manifest Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add local release-manifest-driven version resolution for aggregate deploy products so `--version=0.4.0` maps to exact Helm chart versions.

**Architecture:** Keep manifests external to charts and treat them as aggregate release metadata. Add manifest parsing helpers in `deploy/scripts/lib/common.sh`, then let `core`, `isf`, and `dip` resolve per-release chart versions from a local `--version_file` before download or install. Store manual manifests under `deploy/release-manifests/<version>/<product>.yaml`.

**Tech Stack:** Bash, shell tests

---

### Task 1: Add embedded manifests and failing test

**Files:**
- Create: `deploy/release-manifests/0.4.0/isf.yaml`
- Create: `deploy/release-manifests/0.4.0/kweaver-core.yaml`
- Create: `deploy/release-manifests/0.4.0/kweaver-dip.yaml`
- Create: `deploy/scripts/tests/test-version-manifest.sh`

**Step 1: Write the failing test**

Add a shell test that expects helper functions to:

- Read an exact chart version from a manifest
- Resolve a dependency manifest path from `kweaver-core-0.4.0.yaml`
- Make `download_core` call `download_chart_to_cache` with manifest-defined versions

**Step 2: Run test to verify it fails**

Run: `bash deploy/scripts/tests/test-version-manifest.sh`

Expected: failure because manifest helper functions and wiring do not exist yet.

### Task 2: Add shell-only manifest parser helpers

**Files:**
- Modify: `deploy/scripts/lib/common.sh`

**Step 1: Write minimal implementation**

Add helpers to:

- Normalize a manifest file path
- Validate `product` and `version`
- Read release versions
- Read dependency versions and local manifest paths

Use shell and `awk` for parsing to avoid introducing any runtime dependency beyond standard tooling.

**Step 2: Run test to verify partial progress**

Run: `bash deploy/scripts/tests/test-version-manifest.sh`

Expected: parser checks advance further, but product scripts still fail until they consume manifest data.

### Task 3: Wire manifest support into aggregate deploy scripts

**Files:**
- Modify: `deploy/scripts/services/isf.sh`
- Modify: `deploy/scripts/services/core.sh`
- Modify: `deploy/scripts/services/dip.sh`
- Modify: `deploy/deploy.sh`
- Modify: `deploy/README.md`

**Step 1: Add CLI plumbing**

Add `--version_file` to aggregate module parsers and help text.

**Step 2: Resolve per-release versions**

Before each `download` and repo/local install, resolve the concrete chart version from the manifest when `--version_file` is provided.

**Step 3: Propagate dependency manifests**

When `kweaver-core` uses a manifest and installs/downloads ISF, pass the dependency manifest file and version to ISF.

**Step 4: Run test to verify it passes**

Run: `bash deploy/scripts/tests/test-version-manifest.sh`

Expected: PASS

### Task 4: Verify and document

**Files:**
- Modify: `deploy/README.md`

**Step 1: Add usage example**

Document embedded manifest usage and the optional `--version_file` override.

**Step 2: Run final verification**

Run:

```bash
bash deploy/scripts/tests/test-version-manifest.sh
```

Expected: PASS
