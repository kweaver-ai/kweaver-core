# Deploy SQL Versioned Layout Design

**Date:** 2026-03-30

## Goal

Add version-aware SQL initialization for `isf`, `kweaver-core`, and `kweaver-dip` so `--version=0.4.0` uses the matching SQL set without overwriting the current `0.5.0` SQL content.

## Problem

The current deploy scripts read SQL only from `deploy/scripts/sql`, which now reflects the current `0.5.0` state. This breaks older aggregate versions such as `0.4.0`, where the SQL set differs and some current modules such as `bkn` and `vega` did not exist yet.

## Design

Use a version-scoped, product-grouped SQL directory layout under `deploy/scripts/sql`:

```text
deploy/scripts/sql/
├── 0.4.0/
│   ├── isf/
│   ├── kweaver-core/
│   └── kweaver-dip/
└── 0.5.0/
    ├── isf/
    ├── kweaver-core/
    └── kweaver-dip/
```

Directory semantics:

- `deploy/scripts/sql/<version>/isf/` contains the SQL files executed by `isf install`
- `deploy/scripts/sql/<version>/kweaver-core/<module>/` contains module SQL executed by `kweaver-core install`
- `deploy/scripts/sql/<version>/kweaver-dip/` contains optional SQL executed by `kweaver-dip install`

## Execution Rules

- Install only checks the SQL directory for the product being installed
- If the versioned SQL directory exists and contains SQL files, execute it
- If the versioned SQL directory is missing or empty, skip SQL initialization for that product instead of failing
- `kweaver-core` initializes only the module directories that exist under `deploy/scripts/sql/<version>/kweaver-core/`
- `0.4.0` therefore naturally excludes newer modules such as `bkn` and `vega`
- `kweaver-dip` may execute its own SQL when `deploy/scripts/sql/<version>/kweaver-dip/` exists; otherwise it skips cleanly

## Compatibility

- Preserve current SQL content by moving it into `deploy/scripts/sql/0.5.0/isf` and `deploy/scripts/sql/0.5.0/kweaver-core`
- Restore `0.4.0` SQL from commit `543b6adb200e2d10d672c4db7c33d46e639460ab`
- When `--version` is omitted, default SQL resolution should continue to use the current aggregate version behavior, which maps to the active current version directory

## Validation

Use shell tests to prove:

1. `kweaver-core --version=0.4.0` resolves `deploy/scripts/sql/0.4.0/kweaver-core`
2. `kweaver-core --version=0.4.0` initializes only directories that exist there
3. `isf --version=0.4.0` resolves `deploy/scripts/sql/0.4.0/isf`
4. `kweaver-dip --version=0.4.0` executes product SQL only when `deploy/scripts/sql/0.4.0/kweaver-dip` exists
