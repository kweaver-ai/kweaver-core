# Agent Memory

基于六边形架构的Agent记忆管理服务，用于构建、召回和管理Agent的记忆数据。

## 项目概述

项目背景: 基于六边形架构的Agent记忆管理服务，用于构建、召回和管理Agent的记忆数据。
目标用户: 需要记忆管理能力的AI Agent系统开发者。
核心问题: 提供结构化记忆存储、高效召回和灵活管理机制，解决Agent状态持久化和上下文恢复问题。

## 项目结构

```
src/
├── application/          # 应用层
│   ├── memory            # 记忆构建相关应用逻辑
│   │   ├── build_memory.py
│   │   ├── manage_memory.py
│   │   └── retrieval_memory.py
│   └── __init__.py
├── config/               # 配置文件和初始化逻辑
│   ├── config.yaml       # 主配置文件
│   ├── config.py         # 配置加载逻辑
│   └── __init__.py
├── domain/               # 领域层
│   └── memory            # 记忆管理核心业务逻辑
│       ├── entities.py   # 实体定义
│       ├── mem0_adapter.py # Mem0 适配器
│       └── repositories.py # 存储库接口
├── infrastructure/       # 基础设施层
│   └── db                # 数据库访问
│       ├── db_pool.py    # 数据库连接池
│       └── db_pool_wrapper.py # 连接池封装
├── interfaces/           # 接口层
│   └── api               # API 接口定义
│       ├── exceptions.py # 异常处理
│       ├── middleware.py # 请求中间件
│       ├── routes.py     # 路由定义
│       └── schemas.py    # 数据模型定义
├── locale/               # 国际化资源
│   ├── en_US             # 英文错误信息
│   │   └── errors.toml
│   └── zh_CN             # 中文错误信息
│       └── errors.toml
├── utils/                # 通用工具类
│   ├── i18n.py           # 国际化支持
│   └── logger.py         # 日志记录工具
├── __init__.py
└── main.py               # 项目启动入口文件

```
## 技术选型

- Python: 3.10.18
- Web框架: FastAPI
- LLM支持: Anthropic、OpenAI、Gemini、Ollama 等
- 向量数据库: Opensearch
- 数据库连接: PyMySQL + SQLAlchemy
- 配置管理: PyYAML
- 日志: 自定义Logger工具
- 测试: pytest

## 快速开始

### 安装依赖

```bash
pip install -r requirements.txt
```

3. 修改主配置文件
/src/config/config.yaml
将 LLM、 Embedding Model、Vector Database 的配置修改为实际使用的模型

4. 运行服务
```bash
python src/main.py
```

5. 本地调式
```
# 依赖服务端口映射
kubectl port-forward svc/opensearch-master -n resource --address=0.0.0.0 30920:9200 > /dev/null &
kubectl port-forward svc/mf-model-api -n anyshare --address=0.0.0.0 9898:9898 > /dev/null &



# 设置依赖服务调用地址
export RERANK_URL=http://192.168.124.90:9898/api/private/mf-model-api/v1/small-model/reranker
export OPENSEARCH_HOST=192.168.124.90
export EMBEDDING_MODEL_BASE_URL=http://192.168.124.90:9898/api/private/mf-model-api/v1/small-model
export LLM_BASE_URL=http://192.168.124.90:9898/api/private/mf-model-api/v1
export LLM_MODEL=DeepseekV3

```
## API文档

启动服务后访问：http://localhost:8000/docs

## 配置

修改 `src/config/config.yaml` 中的 LLM、Embedding Model、Vector Database 配置。

## 测试

安装测试依赖：

```bash
pip install -r requirements-dev.txt
```

运行测试：

```bash
pytest tests/
```

或者运行特定测试文件：

```bash
pytest tests/test_rerank_model_client.py
```

## 代码质量

运行代码检查（包括 ruff lint 和 format）：

```bash
make lint
```

或者使用 uv 直接运行：

```bash
uv run pre-commit run -a --config .pre-commit-config.yaml
```
