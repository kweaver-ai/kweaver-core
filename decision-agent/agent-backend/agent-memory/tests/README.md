# Agent Memory Service - Tests

## 测试结构

本测试套件为 agent-memory 服务提供全面的单元测试，覆盖率达到 90% 以上。

```
tests/
├── conftest.py                 # pytest 配置和共享 fixtures
├── unit/                      # 单元测试
│   ├── adaptee/              # 适配器层测试
│   ├── application/           # 应用层测试
│   ├── domain/              # 领域层测试
│   ├── infrastructure/       # 基础设施层测试
│   ├── interfaces/          # 接口层测试
│   └── utils/               # 工具类测试
└── test_rerank_model_client.py  # 现有测试（已迁移到 unit/adaptee/）
```

## 测试覆盖的模块

### Domain 层
- `test_entities.py` - 测试 Memory 和 MemoryChunk 实体
- `test_repositories.py` - 测试仓储接口

### Application 层
- `test_build_memory.py` - 测试 BuildMemoryUseCase
- `test_retrieval_memory.py` - 测试 RetrievalMemoryUseCase
- `test_manage_memory.py` - 测试 ManageMemoryUseCase

### Interfaces 层
- `test_schemas.py` - 测试 Pydantic schemas
- `test_exceptions.py` - 测试异常类

### Utils 层
- `test_env.py` - 测试环境变量工具
- `test_logger.py` - 测试日志工具
- `test_i18n.py` - 测试国际化工具

### Adaptee 层
- `test_rerank_model_client.py` - 测试 RerankModelClient 及相关数据类

## 运行测试

### 安装依赖
```bash
make install-deps
```

### 运行所有单元测试
```bash
make test-unit
```

### 运行所有测试（包括集成测试）
```bash
make test
```

### 生成覆盖率报告
```bash
make coverage
```

### 代码质量检查
```bash
make lint
```

### 代码格式化
```bash
make format
```

## 测试覆盖率目标

- **目标覆盖率**: 90% 以上
- **覆盖率阈值**: 在 `pyproject.toml` 中配置为 `--cov-fail-under=90`
- **排除的目录**:
  - `mem0/` (第三方库)
  - `tests/` (测试代码)
  - `__pycache__/` (缓存文件)

## 测试约定

### 测试命名
- 测试文件: `test_<module_name>.py`
- 测试类: `Test<ClassName>`
- 测试函数: `test_<function_name>`

### 异步测试
所有异步测试使用 `@pytest.mark.asyncio` 装饰器。

### Fixtures
- 使用 `conftest.py` 中的共享 fixtures
- 测试特定的 fixtures 在各自的测试文件中定义

## 已修复的 Code Bad Tastes

以下代码质量问题已在本次任务中修复：

1. **重复导入**: `src/adaptee/mf_model_factory/rerank_model_client.py`
   - 移除了重复的 `import asyncio`

2. **弃用的日志方法**: `src/utils/logger.py`
   - 将 `logger.warn()` 改为 `logger.warning()`
   - 将 `logger.warnf()` 改为 `logger.warningf()`

3. **拼写错误**: `src/infrastructure/db/db_pool_wrapper.py`
   - 将 `attemp` 修正为 `attempt`

4. **无意义赋值**: `src/infrastructure/db/db_pool.py`
   - 移除了无用的 `r = w` 赋值

5. **代码缩进**: `src/utils/i18n.py`
   - 统一了 `_load_resources` 方法中的缩进

## 配置文件

### pyproject.toml
- pytest 配置
- 测试路径设置
- 覆盖率配置
- 异步测试模式

### Makefile
- 测试运行命令
- 代码质量检查命令
- 清理命令

## 注意事项

1. **排除 mem0 目录**: 所有测试排除了 `mem0/` 目录，因为这是第三方库
2. **Mock 使用**: 大量使用 `unittest.mock` 来隔离被测组件
3. **异步测试**: 所有异步方法使用 `AsyncMock` 进行 mock
4. **覆盖率报告**: 测试完成后会生成 HTML 覆盖率报告在 `htmlcov/index.html`

## 持续集成

建议在 CI/CD 流水线中运行以下命令：

```bash
make lint
make test-unit
```

确保代码质量和测试覆盖率始终保持在要求的标准之上。
