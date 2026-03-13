# KWeaver-Core 研发测试规范

> 版本：0.2.0
> 适用范围：KWeaver-Core 及关联产品（Decision Agent、ISF、TraceAI 等）的所有模块

---

## 1. 背景

KWeaver-Core 由多个模块组成，分属多个业务域（ontology、vega、execution-factory、dataflow 等），由多个小组并行开发。本规范同样适用于 Decision Agent、ISF、TraceAI 等关联产品的模块。

测试的编写、运行、修复主要由 AI Agent 完成，人工负责审核和决策。

本规范定义测试体系的目标态：**模块应该长什么样**。不涉及现状分析、迁移路径或编码风格约束。

> Java 模块（`vega-gateway`、`data-connection`）处于维护模式，计划退场，不纳入本规范。

### 1.1 设计原则

1. **Agent-First**：测试由 Agent 生成和维护，规范只定义机器接口（Makefile target、产物路径、隔离机制），不定义人工约束（命名规范、格式要求、打勾清单）
2. **蓝本驱动**：Agent 通过读取蓝本模块的已有测试学习风格，不通过文档记忆规则
3. **验证靠跑**：合规性通过脚本自动验证（能跑通 = 合规），不通过人工 review 检查
4. **统一契约，统一实现**：所有纳入规范的模块**必须**通过 Makefile 暴露统一 target；target 名称、产物路径、分层机制全项目一致；具体测试框架（goconvey、pytest 等）按语言各取所长

---

## 2. 测试分层

| 层级 | 缩写 | 定义 | 外部依赖 | 时间预算 | 隔离机制 |
|------|------|------|----------|----------|----------|
| 单元测试 | UT | 验证单个函数/方法逻辑，全 mock | 无 | 单模块 < 60s | Go: 无 build tag |
| 集成测试 | IT | 验证模块与真实中间件的交互 | 本地中间件 | 单模块 < 5min | Go: `//go:build integration` |
| 验收测试 | AT | 验证完整部署服务的 API 行为，HTTP 黑盒 | 运行中的服务 | 全套 < 30min | Go: `//go:build at`；Python: `pytest -m api` |
| 性能测试 | PT | 验证并发/压力下的延迟和吞吐量 | 运行中的服务 | 按场景定义 | Python: `pytest -m performance` |

**关键约束**：

- **`test` 仅代表 UT**：`make test` 在任何机器上无外部依赖即可通过，必须全 mock
- **`test` 不触发 IT/AT/PT**：IT/AT/PT 必须通过 build tag 或 marker 隔离，由 `test-integration`、`test-at`、`test-performance` 单独入口执行
- IT 连接信息通过环境变量注入（`TEST_DB_URL` 等），不硬编码

---

## 3. 模块标准接口

### 3.1 Makefile Target

**强制要求**：所有纳入规范的模块必须提供 Makefile，作为唯一标准接口。统一入口脚本与合规检查均基于 Makefile 执行。

每个模块的 Makefile **必须**包含以下 target：

| 目标 | 用途 | 必选 |
|------|------|------|
| `test` | 运行 UT（无外部依赖） | 是 |
| `test-cover` | 运行 UT + 生成覆盖率 | 是（Go 必选；Python AT 模块无 UT 时可不实现） |
| `lint` | 静态检查 | 是 |
| `ci` | CI 入口：Go 为 `lint + test-cover`；Python AT 为 `lint + test-at` | 是 |

预留 target（按需实现）：

| 目标 | 用途 |
|------|------|
| `test-race` | UT + 竞态检测（Go） |
| `test-integration` | 集成测试 |
| `test-at` | 验收测试 |
| `test-performance` | 性能测试 |
| `generate-mock` | 重新生成 mock 代码（Go） |

### 3.2 产物输出

所有模块统一输出到 `<module>/test-result/`（加入 `.gitignore`）。各类型模块产出如下：

| 产物 | Go 模块 | Python AT 模块 |
|------|---------|----------------|
| coverage.out | 必须 | — |
| coverage.xml | 必须（需安装 gocover-cobertura） | — |
| coverage.html | 必须 | — |
| junit.xml | — | 必须 |
| allure/ | — | 必须 |

```
test-result/
├── coverage.out          # Go 原始覆盖率
├── coverage.xml          # Cobertura XML（CI 消费，Go 用 gocover-cobertura 生成）
├── coverage.html         # HTML 覆盖率报告（本地查看）
├── junit.xml             # JUnit XML 测试报告（Python AT）
└── allure/               # Allure 原始数据（Python AT）
```

### 3.3 Go 模块 Makefile

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

Python AT 模块以验收测试为主。`test` 必须无外部依赖：仅做用例收集校验（`--collect-only`），不实际执行 AT。`test-at` 为 AT 入口。

```makefile
.PHONY: test test-at test-smoke lint ci

# test 仅做 UT 语义：无外部依赖，验证用例可加载
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

## 4. 依赖版本

同一仓库内各模块必须使用统一的测试依赖版本。

### 4.1 Go

| 用途 | 选型 |
|------|------|
| 断言 | `github.com/smartystreets/goconvey` v1.8+ |
| Mock 生成 | `go.uber.org/mock` v0.5+ |
| SQL Mock | `github.com/DATA-DOG/go-sqlmock` v1.5+ |
| Monkey Patch | `github.com/agiledragon/gomonkey/v2` v2.14+（仅用于无法接口 mock 的场景） |
| 覆盖率转 Cobertura | `github.com/boumenot/gocover-cobertura`（需安装，用于生成 coverage.xml） |

### 4.2 Python

| 用途 | 选型 |
|------|------|
| 测试框架 | pytest |
| 报告 | allure-pytest |

---

## 5. 统一入口

### 5.1 `scripts/ci-run.sh`

根目录提供统一入口脚本，本地和 CI 共用。提供语义化命令，不要求使用者知道底层 Makefile target 叫什么。

**用法**：

```bash
./scripts/ci-run.sh ut                            # 所有模块 UT
./scripts/ci-run.sh it                            # 所有模块 IT
./scripts/ci-run.sh at                            # 所有模块 AT
./scripts/ci-run.sh all                           # UT + IT
./scripts/ci-run.sh ci                            # lint + UT + 覆盖率
./scripts/ci-run.sh lint                          # 只跑 lint
./scripts/ci-run.sh cover                         # 只跑覆盖率

# 指定模块 / 业务域
./scripts/ci-run.sh ut ontology/ontology-manager  # 单个模块
./scripts/ci-run.sh ut ontology/                  # 按业务域过滤
./scripts/ci-run.sh it vega/                      # vega 域的 IT
```

**命令 → Makefile target 映射**：

| 命令 | Makefile target | 说明 |
|------|-----------------|------|
| `ut` | `test` | 单元测试（无外部依赖） |
| `it` | `test-integration` | 集成测试 |
| `at` | `test-at` | 验收测试 |
| `all` | `test` + `test-integration` | UT + IT |
| `ci` | `ci` | CI 全流程（Go: lint + test-cover；Python AT: lint + test-at） |
| `lint` | `lint` | 静态检查 |
| `cover` | `test-cover` | UT + 覆盖率报告 |

**脚本**：

```bash
#!/usr/bin/env bash
set -euo pipefail

COMMAND="${1:?Usage: $0 <ut|it|at|all|ci|lint|cover> [module-filter]}"
FILTER="${2:-}"

# 命令 → Makefile target 映射
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

# Go 模块
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

# Python AT 套件
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

    # 跳过模块没有实现的可选 target
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

# 汇总报告
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

自动检测各模块是否符合本规范。Agent 或 CI 均可执行。

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

### 5.3 验收标准

不需要人工逐项检查。两条命令：

```bash
./scripts/check-test-compliance.sh   # → All checks passed
./scripts/ci-run.sh ut               # → PASSED: 12, FAILED: 0
```

---

## 6. CI 集成

### 6.1 触发策略

| 事件 | 运行内容 | 时间预算 |
|------|----------|----------|
| PR 提交/更新 | 按 path filter 检测变更模块，执行 `make ci` | < 5 min |
| merge 到 main | `./scripts/ci-run.sh ci` 全量 | < 15 min |
| nightly | `./scripts/ci-run.sh at` 全量 | < 30 min |

### 6.2 覆盖率展示

本期暂不设门禁，仅在 PR 上展示覆盖率数字。

- Go: `go tool cover -func` 输出摘要到 PR comment
- 统一 CI 消费格式: Cobertura XML（Go 用 `gocover-cobertura` 转换）

### 6.3 Path Filter 示例

```yaml
# PR 变更检测（伪代码）
paths:
  "ontology/**"           → ./scripts/ci-run.sh ci ontology/
  "vega/**"               → ./scripts/ci-run.sh ci vega/
  "dataflow/**"           → ./scripts/ci-run.sh ci dataflow/
  "context-loader/**"     → ./scripts/ci-run.sh ci context-loader/
  "execution-factory/**"  → ./scripts/ci-run.sh ci execution-factory/
```

---

## 7. Tradeoffs

记录已做出的关键决策及理由。

| # | 决策点 | 决策 | 理由 | 备选方案 |
|---|---|---|---|---|
| 1 | Performance Testing | 本期不纳入，Makefile 预留 target | UT/IT 基础未就绪，PT 的 ROI 低 | 纳入 PT |
| 2 | 断言库统一 | goconvey（项目已有 6 模块在用） | Agent 从蓝本学习即可统一，无需强制规则 | testify |
| 3 | Mock 库 | `go.uber.org/mock` | `golang/mock` 2023 年已 archived | 继续用 golang/mock |
| 4 | IT 基础设施 | 本地直连中间件，环境变量配置 | 最低门槛 | Docker Compose |
| 5 | 覆盖率门禁 | 暂不设，仅展示 | 先建立可见性 | 趋势制 / 硬阈值 |
| 6 | 统一入口 | `scripts/ci-run.sh` 语义化命令 | 比根 Makefile 灵活；`ut`/`it`/`at`/`all` 比 Makefile target 名直观 | 根 Makefile / Taskfile |
| 7 | Java 模块 | 不纳入规范（维护模式，计划退场） | 不值得投入测试建设 | 纳入 |
| 8 | 编码规范 | 不写，Agent 从蓝本学习 | 人工约束维护成本高，Agent 不需要背规矩 | 详细编码规范文档 |

---

## 8. 可验证规则清单

以下为验证工具可执行的检查项，按 MUST / SHOULD / OPTIONAL 分类。

### MUST（必须满足，不合规则构建失败）

| # | 规则 | 验证方式 |
|---|------|----------|
| M1 | 模块存在 Makefile | `test -f $mod/Makefile` |
| M2 | Makefile 包含 `test` target | `grep -q '^test:' $mod/Makefile` |
| M3 | Makefile 包含 `ci` target | `grep -q '^ci:' $mod/Makefile` |
| M4 | Makefile 包含 `lint` target | `grep -q '^lint:' $mod/Makefile` |
| M5 | `make test` 无外部依赖即可通过 | 执行 `make -C $mod test`，退出码 0 |
| M6 | `test-result/` 在 `.gitignore` 中 | `grep -rq 'test-result' .gitignore` |
| M7 | Go 模块：Makefile 包含 `test-cover` | `grep -q '^test-cover:' $mod/Makefile`（当存在 go.mod 时） |
| M8 | Go 模块：不使用 `github.com/golang/mock` | `! grep -rq 'github.com/golang/mock' $mod/`（当存在 go.mod 时） |

### SHOULD（建议满足，可告警）

| # | 规则 | 验证方式 |
|---|------|----------|
| S1 | Go 模块：`test-cover` 产出 coverage.xml | 执行 `make test-cover` 后检查 `test-result/coverage.xml` 存在 |
| S2 | Python AT 模块：`test-at` 产出 junit.xml 与 allure/ | 执行 `make test-at` 后检查产物存在 |
| S3 | UT 单模块执行时间 < 60s | 计时执行 `make test` |

### OPTIONAL（可选）

| # | 规则 | 验证方式 |
|---|------|----------|
| O1 | 模块实现 `test-integration` | `grep -q '^test-integration:' $mod/Makefile` |
| O2 | 模块实现 `test-at` | `grep -q '^test-at:' $mod/Makefile` |
| O3 | 模块实现 `test-performance` | `grep -q '^test-performance:' $mod/Makefile` |

---

## 附录 A：Agent 工作参考

以下为 Agent 生成/维护测试时的**参考实践**，非正式规范条款。Agent 可据此学习风格，但合规性以第 8 节可验证规则为准。

### 蓝本模块

| 语言 | 蓝本模块 | 参考路径 | 学什么 |
|------|----------|----------|--------|
| Go UT | `ontology/ontology-manager` | `server/logics/*/`, `server/drivenadapters/*/` | 断言风格、mock 用法、场景覆盖 |
| Go Makefile | `context-loader/agent-retrieval` | `Makefile` | target 设计、包排除 |
| Go Mock | `ontology/ontology-manager` | `server/interfaces/` | `//go:generate` 指令 |
| Go AT | `vega/vega-backend` | `server/tests/at/` | config-driven、HTTP 黑盒 |
| Python AT | `execution-factory/tests` | `conftest.py`, `lib/` | fixture 分层、API client |

### Agent 工作流建议

```
1. 读取目标模块源码（被测函数/方法）
2. 读取蓝本模块的已有测试，学习风格
3. 读取目标模块已有测试（如有），保持风格一致
4. 生成测试代码
5. make test → 验证通过
6. make test-cover → 确认覆盖率变化（Go 模块）
```

---

> **维护说明**：本文档定义目标态。如需变更决策，提 PR 讨论。
>
> 英文版（默认）：见 [TESTING.md](TESTING.md)
