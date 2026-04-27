# KWeaver-Core Testing Specification

> Version: 0.2.0
> Scope: All modules of KWeaver-Core and related products (Decision Agent, ISF, TraceAI, etc.)

---

## 1. Background

KWeaver-Core consists of multiple modules across business domains (ontology, vega, execution-factory, dataflow, etc.), developed in parallel by multiple teams. This specification applies to Decision Agent, ISF, TraceAI, and other related product modules.

Tests are primarily authored, executed, and maintained by AI Agents; humans review and approve.

This specification defines the **target state** of the testing system: **what modules should look like**. It does not cover current-state analysis, migration paths, or coding style constraints.

> Java modules (`vega-gateway`, `data-connection`) are in maintenance mode and planned for retirement; they are out of scope.

### 1.1 Design Principles

1. **Agent-First**: Tests are generated and maintained by Agents; the spec defines machine interfaces (Makefile targets, artifact paths, isolation mechanisms) only, not human constraints (naming, formatting, checklists).
2. **Blueprint-Driven**: Agents learn style by reading existing tests in blueprint modules, not by memorizing rules from documentation.
3. **Verify by Running**: Compliance is validated by scripts (pass = compliant), not by manual review.
4. **Unified Contract, Unified Implementation**: All in-scope modules **must** expose a unified interface via Makefile; target names, artifact paths, and layering are consistent across the project; test frameworks (goconvey, pytest, etc.) may vary by language.

---

## 2. Test Layers

| Layer | Abbr | Definition | External Deps | Time Budget | Isolation |
|-------|------|------------|---------------|-------------|-----------|
| Unit Test | UT | Validates single function/method logic, fully mocked | None | Per module < 60s | Go: no build tag |
| Integration Test | IT | Validates module interaction with real middleware | Local middleware | Per module < 5min | Go: `//go:build integration` |
| Acceptance Test | AT | Validates deployed service API behavior, HTTP black-box | Running service | Full suite < 30min | Go: `//go:build at`; Python: `pytest -m api` |
| Performance Test | PT | Validates latency and throughput under load | Running service | Per scenario | Python: `pytest -m performance` |

**Key constraints**:

- **`test` means UT only**: `make test` must pass on any machine with no external dependencies; must be fully mocked.
- **`test` does not run IT/AT/PT**: IT/AT/PT must be isolated via build tags or markers and run via `test-integration`, `test-at`, `test-performance` only.
- IT connection info must be injected via environment variables (`TEST_DB_URL`, etc.), not hardcoded.

---

## 3. Module Standard Interface

### 3.1 Makefile Targets

**Mandatory**: All in-scope modules must provide a Makefile as the sole standard interface. Unified entry scripts and compliance checks operate on Makefile targets.

Each module's Makefile **must** include these targets:

| Target | Purpose | Required |
|--------|---------|----------|
| `test` | Run UT (no external deps) | Yes |
| `test-cover` | Run UT + generate coverage | Yes (Go required; Python AT may omit if no UT) |
| `lint` | Static checks | Yes |
| `ci` | CI entry: Go = `lint + test-cover`; Python AT = `lint + test-at` | Yes |

Reserved targets (implement as needed):

| Target | Purpose |
|--------|---------|
| `test-race` | UT + race detection (Go) |
| `test-integration` | Integration tests |
| `test-at` | Acceptance tests |
| `test-performance` | Performance tests |
| `generate-mock` | Regenerate mock code (Go) |

### 3.2 Artifact Output

All modules output to `<module>/test-result/` (add to `.gitignore`). Outputs by module type:

| Artifact | Go Module | Python AT Module |
|----------|-----------|------------------|
| coverage.out | Required | — |
| coverage.xml | Required (requires gocover-cobertura) | — |
| coverage.html | Required | — |
| junit.xml | — | Required |
| allure/ | — | Required |

```
test-result/
├── coverage.out          # Go raw coverage
├── coverage.xml          # Cobertura XML (CI consumption; Go uses gocover-cobertura)
├── coverage.html         # HTML coverage report (local view)
├── junit.xml             # JUnit XML test report (Python AT)
└── allure/               # Allure raw data (Python AT)
```

### 3.3 Go Module Makefile

```makefile
.PHONY: generate-mock lint test test-cover test-race test-integration ci

generate-mock:
	go generate ./...

lint:
	golangci-lint run ./... --exclude-dirs=server/tests,server/mocks

UT_PACKAGES = $(shell go list ./... | grep -v /server/tests/ | grep -v /server/mocks)

test:
	go test $(UT_PACKAGES) -gcflags=all=-l -v -count=1

test-cover:
	@mkdir -p test-result
	go test $(UT_PACKAGES) -gcflags=all=-l \
		-coverprofile=test-result/coverage.out \
		-covermode=atomic
	go tool cover -func=test-result/coverage.out
	go tool cover -html=test-result/coverage.out -o test-result/coverage.html
	@command -v gocover-cobertura >/dev/null 2>&1 && gocover-cobertura < test-result/coverage.out > test-result/coverage.xml || \
		(echo "WARN: gocover-cobertura not found, coverage.xml not generated; install for CI compliance")

test-race:
	go test $(UT_PACKAGES) -gcflags=all=-l -race -count=1

test-integration:
	go test -tags=integration -v ./server/tests/integration/... -timeout 5m

ci: lint test-cover
```

### 3.4 Python AT Makefile

Python AT modules are primarily acceptance tests. `test` must have no external deps: run `--collect-only` to validate test collection, not actual AT execution. `test-at` is the AT entry point.

```makefile
.PHONY: test test-at test-smoke lint ci

# test: UT semantics only; no external deps; validates test suite loads
test:
	python3 -m pytest testcases/ --collect-only -q

test-at:
	@mkdir -p test-result
	python3 -m pytest testcases/ -v -s --tb=short -m api \
		--junitxml=test-result/junit.xml \
		--alluredir=test-result/allure

test-smoke:
	python3 -m pytest testcases/ -v -s --tb=short -m smoke

lint:
	python3 -m flake8 testcases/ common/ lib/ --max-line-length=120

ci: lint test-at
```

---

## 4. Dependency Versions

All modules in the same repo must use the same test dependency versions.

### 4.1 Go

| Purpose | Choice |
|---------|--------|
| Assertions | `github.com/smartystreets/goconvey` v1.8+ |
| Mock generation | `go.uber.org/mock` v0.5+ |
| SQL Mock | `github.com/DATA-DOG/go-sqlmock` v1.5+ |
| Monkey Patch | `github.com/agiledragon/gomonkey/v2` v2.14+ (only when interface mocking is not feasible) |
| Coverage to Cobertura | `github.com/boumenot/gocover-cobertura` (install required; for coverage.xml) |

### 4.2 Python

| Purpose | Choice |
|---------|--------|
| Test framework | pytest |
| Reporting | allure-pytest |

---

## 5. Unified Entry

### 5.1 `scripts/ci-run.sh`

A unified entry script at repo root, shared by local and CI. Provides semantic commands; users do not need to know Makefile target names.

**Usage**:

```bash
./scripts/ci-run.sh ut                            # All modules UT
./scripts/ci-run.sh it                            # All modules IT
./scripts/ci-run.sh at                            # All modules AT
./scripts/ci-run.sh all                           # UT + IT
./scripts/ci-run.sh ci                            # lint + UT + coverage
./scripts/ci-run.sh lint                          # lint only
./scripts/ci-run.sh cover                         # coverage only

# Filter by module
./scripts/ci-run.sh ut ontology/ontology-manager  # Single module
./scripts/ci-run.sh ut ontology/                  # By domain
./scripts/ci-run.sh it vega/                      # vega domain IT
```

**Command → Makefile target mapping**:

| Command | Makefile target | Description |
|---------|-----------------|-------------|
| `ut` | `test` | Unit tests (no external deps) |
| `it` | `test-integration` | Integration tests |
| `at` | `test-at` | Acceptance tests |
| `all` | `test` + `test-integration` | UT + IT |
| `ci` | `ci` | Full CI (Go: lint + test-cover; Python AT: lint + test-at) |
| `lint` | `lint` | Static checks |
| `cover` | `test-cover` | UT + coverage report |

**Script**:

```bash
#!/usr/bin/env bash
set -euo pipefail

COMMAND="${1:?Usage: $0 <ut|it|at|all|ci|lint|cover> [module-filter]}"
FILTER="${2:-}"

case "$COMMAND" in
    ut)    TARGETS=("test") ;;
    it)    TARGETS=("test-integration") ;;
    at)    TARGETS=("test-at") ;;
    all)   TARGETS=("test" "test-integration") ;;
    ci)    TARGETS=("ci") ;;
    lint)  TARGETS=("lint") ;;
    cover) TARGETS=("test-cover") ;;
    *)     echo "Unknown command: $COMMAND"; exit 1 ;;
esac

GO_MODULES=(
    context-loader/agent-retrieval
    ontology/ontology-manager
    ontology/ontology-query
    vega/mdl-data-model
    vega/mdl-uniquery
    vega/mdl-data-model-job
    vega/vega-backend
    vega/vega-gateway-pro
    execution-factory/operator-integration
    dataflow/flow-automation
)

PYTHON_MODULES=(
    execution-factory/tests
    dataflow/tests
)

PASSED=()
FAILED=()
SKIPPED=()

run_module() {
    local mod="$1" target="$2"

    if [[ -n "$FILTER" && "$mod" != "$FILTER"* ]]; then
        return
    fi

    if [[ ! -f "$mod/Makefile" ]]; then
        FAILED+=("$mod:Makefile missing")
        return
    fi

    if ! grep -q "^${target}:" "$mod/Makefile" 2>/dev/null; then
        SKIPPED+=("$mod:$target")
        return
    fi

    echo ""
    echo "━━━ $target: $mod ━━━"
    if make -C "$mod" "$target"; then
        PASSED+=("$mod:$target")
    else
        FAILED+=("$mod:$target")
    fi
}

ALL_MODULES=("${GO_MODULES[@]}" "${PYTHON_MODULES[@]}")

for target in "${TARGETS[@]}"; do
    for mod in "${ALL_MODULES[@]}"; do
        run_module "$mod" "$target"
    done
done

echo ""
echo "════════════════════════════════"
echo "  PASSED:  ${#PASSED[@]}"
echo "  FAILED:  ${#FAILED[@]}"
echo "  SKIPPED: ${#SKIPPED[@]}"
if [[ ${#FAILED[@]} -gt 0 ]]; then
    for m in "${FAILED[@]}"; do echo "    FAIL $m"; done
    echo "════════════════════════════════"
    exit 1
fi
echo "════════════════════════════════"
```

### 5.2 `scripts/check-test-compliance.sh`

Automatically checks whether modules comply with this specification. Usable by Agents or CI.

```bash
#!/usr/bin/env bash
set -euo pipefail

ERRORS=0

check() {
    local desc="$1" cmd="$2"
    if eval "$cmd" > /dev/null 2>&1; then
        echo "  PASS  $desc"
    else
        echo "  FAIL  $desc"
        ERRORS=$((ERRORS + 1))
    fi
}

GO_MODULES=(
    context-loader/agent-retrieval
    ontology/ontology-manager
    ontology/ontology-query
    vega/mdl-data-model
    vega/mdl-uniquery
    vega/mdl-data-model-job
    vega/vega-backend
    vega/vega-gateway-pro
    execution-factory/operator-integration
    dataflow/flow-automation
)

PYTHON_MODULES=(
    execution-factory/tests
    dataflow/tests
)

ALL_MODULES=("${GO_MODULES[@]}" "${PYTHON_MODULES[@]}")

for mod in "${ALL_MODULES[@]}"; do
    echo ""
    echo "=== $mod ==="
    check "Makefile exists" "test -f $mod/Makefile"
    check "make test target" "grep -q '^test:' $mod/Makefile 2>/dev/null"
    check "make lint target" "grep -q '^lint:' $mod/Makefile 2>/dev/null"
    check "make ci target" "grep -q '^ci:' $mod/Makefile 2>/dev/null"
    check "test-result/ in .gitignore" "grep -rq 'test-result' .gitignore 2>/dev/null"

    if [[ -f "$mod/go.mod" ]] || [[ -f "$mod/server/go.mod" ]]; then
        check "make test-cover target (Go)" "grep -q '^test-cover:' $mod/Makefile 2>/dev/null"
        check "no golang/mock (deprecated)" \
            "! grep -rq 'github.com/golang/mock' $mod/"
    fi
done

echo ""
echo "════════════════════════════════"
if [[ $ERRORS -gt 0 ]]; then
    echo "  $ERRORS issues found"
    exit 1
else
    echo "  All checks passed"
fi
echo "════════════════════════════════"
```

### 5.3 Acceptance Criteria

No manual checklist. Two commands:

```bash
./scripts/check-test-compliance.sh   # → All checks passed
./scripts/ci-run.sh ut               # → PASSED: 12, FAILED: 0
```

---

## 6. CI Integration

### 6.1 Trigger Strategy

| Event | Run | Time Budget |
|-------|-----|-------------|
| PR submit/update | Path-filter changed modules, run `make ci` | < 5 min |
| Merge to main | Full `./scripts/ci-run.sh ci` | < 15 min |
| Nightly | Full `./scripts/ci-run.sh at` | < 30 min |

### 6.2 Coverage Display

No gate for now; coverage numbers shown on PR only.

- Go: `go tool cover -func` summary to PR comment
- Unified CI format: Cobertura XML (Go via `gocover-cobertura`)

### 6.3 Path Filter Example

```yaml
# PR change detection (pseudo)
paths:
  "ontology/**"           → ./scripts/ci-run.sh ci ontology/
  "vega/**"               → ./scripts/ci-run.sh ci vega/
  "dataflow/**"           → ./scripts/ci-run.sh ci dataflow/
  "context-loader/**"     → ./scripts/ci-run.sh ci context-loader/
  "execution-factory/**"  → ./scripts/ci-run.sh ci execution-factory/
```

---

## 7. Tradeoffs

Recorded decisions and rationale.

| # | Decision | Choice | Rationale | Alternatives |
|---|---|---|---|---|
| 1 | Performance Testing | Not in scope; Makefile target reserved | UT/IT foundation first; PT ROI low | Include PT |
| 2 | Assertion library | goconvey (6 modules already use) | Agent learns from blueprints; no strict rule | testify |
| 3 | Mock library | `go.uber.org/mock` | `golang/mock` archived 2023 | Keep golang/mock |
| 4 | IT infrastructure | Local middleware, env vars | Lowest barrier | Docker Compose |
| 5 | Coverage gate | None; display only | Visibility first | Trend / hard threshold |
| 6 | Unified entry | `scripts/ci-run.sh` semantic commands | More flexible than root Makefile | Root Makefile / Taskfile |
| 7 | Java modules | Out of scope (maintenance, retiring) | Not worth investing | Include |
| 8 | Coding style | Not specified; Agent learns from blueprints | High maintenance; Agent does not need rules | Detailed style doc |

---

## 8. Verifiable Rules Checklist

Check items for validation tools, grouped as MUST / SHOULD / OPTIONAL.

### MUST (required; non-compliance fails build)

| # | Rule | Verification |
|---|------|---------------|
| M1 | Module has Makefile | `test -f $mod/Makefile` |
| M2 | Makefile has `test` target | `grep -q '^test:' $mod/Makefile` |
| M3 | Makefile has `ci` target | `grep -q '^ci:' $mod/Makefile` |
| M4 | Makefile has `lint` target | `grep -q '^lint:' $mod/Makefile` |
| M5 | `make test` passes with no external deps | Run `make -C $mod test`, exit 0 |
| M6 | `test-result/` in `.gitignore` | `grep -rq 'test-result' .gitignore` |
| M7 | Go: Makefile has `test-cover` | `grep -q '^test-cover:' $mod/Makefile` (when go.mod exists) |
| M8 | Go: no `github.com/golang/mock` | `! grep -rq 'github.com/golang/mock' $mod/` (when go.mod exists) |

### SHOULD (recommended; may warn)

| # | Rule | Verification |
|---|------|---------------|
| S1 | Go: `test-cover` produces coverage.xml | After `make test-cover`, check `test-result/coverage.xml` exists |
| S2 | Python AT: `test-at` produces junit.xml and allure/ | After `make test-at`, check artifacts exist |
| S3 | UT per module < 60s | Time `make test` |

### OPTIONAL

| # | Rule | Verification |
|---|------|---------------|
| O1 | Module implements `test-integration` | `grep -q '^test-integration:' $mod/Makefile` |
| O2 | Module implements `test-at` | `grep -q '^test-at:' $mod/Makefile` |
| O3 | Module implements `test-performance` | `grep -q '^test-performance:' $mod/Makefile` |

---

## Appendix A: Agent Work Reference

**Reference practices** for Agents generating/maintaining tests; not formal spec. Agents may use for style; compliance is defined by Section 8.

### Blueprint Modules

| Language | Blueprint | Path | Learn |
|----------|-----------|------|-------|
| Go UT | `ontology/ontology-manager` | `server/logics/*/`, `server/drivenadapters/*/` | Assertion style, mock usage, coverage |
| Go Makefile | `context-loader/agent-retrieval` | `Makefile` | Target design, package exclusion |
| Go Mock | `ontology/ontology-manager` | `server/interfaces/` | `//go:generate` directives |
| Go AT | `vega/vega-backend` | `server/tests/at/` | Config-driven, HTTP black-box |
| Python AT | `execution-factory/tests` | `conftest.py`, `lib/` | Fixture layering, API client |

### Agent Workflow Suggestion

```
1. Read target module source (functions/methods under test)
2. Read blueprint module tests for style
3. Read target module existing tests (if any), keep style consistent
4. Generate test code
5. make test → verify pass
6. make test-cover → confirm coverage change (Go modules)
```

---

> **Maintenance**: This document defines the target state. Propose changes via PR.
>
> 中文版：见 [TESTING.zh.md](TESTING.zh.md)
