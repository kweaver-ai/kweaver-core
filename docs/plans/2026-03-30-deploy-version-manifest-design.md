# Deploy Release Manifest Design

**Date:** 2026-03-30

## Goal

Introduce a release-manifest-driven version model for `isf`, `kweaver-core`, and `kweaver-dip` so `--version=0.4.0` resolves to an exact chart-version set instead of assuming every chart publishes the same Helm version.

## Problem

The current deploy scripts pass `--version` directly to every Helm chart. This fails when a product release version such as `0.4.0` maps to chart versions like `0.4.0-20260329.1`, `0.4.0-hotfix2`, or other non-uniform tags.

## Design

Each aggregate product release publishes a manifest file. The deploy scripts keep these manifests under `deploy/release-manifests/<version>/<product>.yaml` and they are maintained manually.

Example manifest names:

- `deploy/release-manifests/0.4.0/isf.yaml`
- `deploy/release-manifests/0.4.0/kweaver-core.yaml`
- `deploy/release-manifests/0.4.0/kweaver-dip.yaml`

Manifest structure:

- Top-level product metadata: `product`, `version`
- Optional source metadata: repo name and URL
- Optional dependencies: other aggregate products and their manifest files
- Exact per-release chart versions under `releases`

## Embedded Manifest Scope

This change will implement:

- Embedded manifests for `isf`, `kweaver-core`, `kweaver-dip`
- Shell-only parser helpers in deploy scripts
- Automatic embedded manifest discovery from `--version`
- `--version_file` as an explicit override
- Exact version resolution for `download`
- Dependency-manifest propagation from `kweaver-core` to `isf`

This change will not yet implement:

- Remote manifest download by URL convention
- Full offline package publishing
- Lockfile generation automation

## Validation

Use a shell test to prove:

1. `kweaver-core-0.4.0.yaml` resolves exact chart versions.
2. `download_core` passes exact chart versions to `download_chart_to_cache`.
3. `download_core` passes the referenced ISF manifest to `download_isf`.
