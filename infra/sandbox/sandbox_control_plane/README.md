# 沙箱平台控制中心

基于六边形架构的代码沙箱管理平台。

## 架构

本项目采用**六边形架构**（Hexagonal Architecture），确保核心业务逻辑独立于技术实现。

### 依赖方向

```
Interfaces → Application → Domain ← Infrastructure
```

- **Domain（领域层）**: 核心业务逻辑，无外部依赖
- **Application（应用层）**: 用例编排，依赖 Domain
- **Infrastructure（基础设施层）**: 技术实现，实现 Domain 定义的接口
- **Interfaces（接口层）**: 对外接口，依赖 Application

## 快速开始

### 安装

```bash
# 克隆仓库
git clone https://github.com/sandbox/sandbox-control-plane
cd sandbox-control-plane

# 使用 uv 管理项目
# 安装 uv（如果尚未安装）
# curl -LsSf https://astral.sh/uv/install.sh | sh  # Linux/macOS
# powershell -c "irm https://astral.sh/uv/install.ps1"  # Windows

# 同步依赖
uv sync

# 激活虚拟环境
source .venv/bin/activate  # Linux/macOS
.venv\Scripts\activate     # Windows
```

### 配置

```bash
# 复制环境变量模板
cp .env.example .env

# 编辑配置
vim .env
```

### 数据库迁移

```bash
# 初始化数据库
alembic upgrade head
```

### 运行

```bash
# 开发模式
uvicorn src.interfaces.rest.main:app --reload --port 8000

# 生产模式
uvicorn src.interfaces.rest.main:app --host 0.0.0.0 --port 8000 --workers 4
```

## 项目结构

```
src/
├── domain/           # 领域层（核心）
├── application/      # 应用层（用例）
├── infrastructure/   # 基础设施层（适配器）
└── interfaces/       # 接口层（API）
```

详见 [PROJECT_STRUCTURE.md](PROJECT_STRUCTURE.md)。

## API 文档

启动服务后访问：

- Swagger UI: http://localhost:8000/docs
- ReDoc: http://localhost:8000/redoc

## 测试

```bash
# 使用 uv 运行所有测试
uv run pytest

# 运行单元测试
uv run pytest tests/unit

# 运行集成测试
uv run pytest tests/integration

# 生成覆盖率报告
uv run pytest --cov=src --cov-report=html
```

## 开发

```bash
# 使用 uv 运行代码格式化
uv run black src tests

# 代码检查
uv run ruff check src tests

# 类型检查
uv run mypy src
```

## License

MIT License
