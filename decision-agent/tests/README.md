# Agent AT 项目介绍

Agent AT 是一个基于 pytest 测试框架的企业级自动化测试项目，用于对 data-agent、data-operator-hub 和 data-flow 系统进行 API、功能和性能测试。项目采用中文开发规范，设计用于 Kubernetes 环境中的持续集成测试。

## 项目特点

- **测试类型全面**: 支持 API 测试、功能测试、性能测试
- **多服务覆盖**: 涵盖 data-agent、data-operator-hub、data-flow 三大核心服务
- **企业级规范**: 遵循企业命名规范和测试标准
- **容器化部署**: 基于 Docker 的测试环境，支持 Kubernetes 集成
- **自动化报告**: 集成 Allure 测试报告生成
- **数据驱动**: 配置驱动的测试数据管理

## 快速开始

```bash
# 安装依赖
pip install -r requirements/requirements.txt

# 运行所有测试
python3 -m pytest

# 运行特定服务的测试
python3 -m pytest ./testcases/data-agent/
python3 -m pytest ./testcases/data-operator-hub/
python3 -m pytest ./testcases/data-flow/

# 使用 Docker 运行
docker run -it --name agent-at --net=host \
  -v /root/agent-AT:/app \
  -v /root/.kube/config:/root/.kube/config \
  -v /usr/bin/kubectl:/usr/bin/kubectl \
  acr.aishu.cn/dip/agent-at:20250418 bash
```

# pytest 运行规范

## 1. 测试用例路径规范（2025-12-17 更新）

**新目录结构**（推荐使用）:
```bash
# 按服务组织的测试路径
./testcases/data-agent/api/                    # data-agent API 测试
./testcases/data-agent/functional/             # data-agent 功能测试
./testcases/data-agent/performance/            # data-agent 性能测试
./testcases/data-operator-hub/api/             # operator-hub API 测试
./testcases/data-operator-hub/functional/      # operator-hub 功能测试
./testcases/data-flow/api/                     # data-flow API 测试
./testcases/data-flow/functional/              # data-flow 功能测试

# 运行特定测试模块
python3 -m pytest ./testcases/data-agent/api/test_agent_factory_v3.py
python3 -m pytest ./testcases/data-agent/functional/test_agent_chat.py
python3 -m pytest ./testcases/data-operator-hub/api/category/test_create_category.py

# 运行特定测试方法
python3 -m pytest ./testcases/data-agent/api/test_agent_factory_v3.py::TestAgentFactoryV3::test_create_agent_success
```

**旧目录结构**（已废弃，但仍向后兼容）:
```bash
./testcases/api/data-agent/
./testcases/functional/data-agent/
./testcases/performance/data-agent/
```

## 2. 测试标记（Markers）

```bash
# 使用标记运行测试
python3 -m pytest -m smoke              # 冒烟测试
python3 -m pytest -m api                # API 测试
python3 -m pytest -m slow               # 耗时测试
```

## 3. 环境配置规范

- **配置文件位置**: `./config/env.ini`
- **配置格式**: INI 格式（不再使用 metadata.json）
- **配置内容**: 服务器地址、端口、数据库连接信息等

```ini
[server]
host = 192.168.232.11
port = 443
db_port = 3330
db_user = anyshare
db_pwd = eisoo.com123

[requests]
protocol = https

[hydra]
svc_name = hydra-admin
svc_port = 4445
```

## 4. 测试数据和资源管理规范

- **data 目录**: 存放结构性测试数据（JSON、YAML 等文件）
  - `data/data-agent/import/` - Agent 配置文件
  - `data/data-agent/cfg.json` - Agent 测试配置
- **resource 目录**: 存放非结构化数据（PDF、DOC、ZIP 等文档）
  - 资源文件通过 FTP 管理，不提交到仓库
  - 管道运行前自动下载到 resource 目录

## 5. Pipeline 参数规范

Pipeline YAML 中 parameter 部分需要统一，便于继承到 CI/CD 系统：

```yaml
parameters:
  - name: host
    displayName: 服务器VIP
    type: string
    default: 192.168.232.11

  - name: password
    displayName: 服务器SSH密码
    type: string
    default: eisoo.com123

  - name: testdir
    displayName: 测试用例目录
    type: string
    default: testcases/data-agent

  - name: ssh_user
    displayName: 服务器SSH用户名
    type: string
    default: root

  - name: ssh_port
    displayName: 服务器SSH端口
    type: number
    default: 22

  - name: db_port
    displayName: 数据库端口
    type: number
    default: 3330

  - name: db_user
    displayName: 数据库用户名
    type: string
    default: anyshare

  - name: db_password
    displayName: 数据库密码
    type: string
    default: eisoo.com123
```

# 分支管理

为便于 DIP 部门 AT 资产的归档，以及售后问题修复进行冒烟回归，AT 仓库需要跟随 DIP 大版本迭代。

**分支命名规范**:
- **开发分支**: `feature/[需求号]`
- **主分支**: `MISSION`
- **发布分支**: `release/DIP-[大版本号]`

**示例**:
```bash
feature/adp_5004     # 开发分支
MISSION              # 主分支
release/DIP-5.0      # 发布分支
```

# 目录结构（2025-12-17 更新）

```
agent-at/
├── CLAUDE.md                # Claude Code 项目指导文档
├── README.md                # 主项目说明文档（本文件）
├── pytest.ini               # pytest 框架的全局配置文件
├── conftest.py              # 存放可重用的测试前置和后置条件
├── main.py                  # 主入口脚本
├── cm.yaml                  # Kubernetes ConfigMap 配置
│
├── common/                  # 通用代码和工具类
│   ├── __init__.py
│   ├── get_token.py         # 认证令牌管理
│   ├── create_user.py       # 测试用户创建
│   ├── delete_user.py       # 测试用户清理
│   ├── request.py           # 通用 HTTP 请求模式
│   └── assert_tools.py      # 自定义断言工具
│
├── config/                  # 存放配置文件
│   └── env.ini              # 环境配置文件
│
├── data/                    # 存放结构性测试数据
│   └── data-agent/
│       ├── import/          # Agent 配置文件（效能分析等）
│       └── cfg.json         # Agent 测试配置
│
├── lib/                     # 存放各产品服务封装的接口
│   ├── data-agent/          # data-agent 相关的库文件
│   │   ├── agent_factory.py    # Agent Factory V3 API 客户端（40+ 方法）
│   │   └── agent_app.py        # Agent App 外部 API 客户端
│   ├── operator.py            # Operator 管理 APIs
│   ├── operator_internal.py   # 内部 Operator APIs
│   ├── mcp.py                 # MCP 协议处理
│   ├── mcp_internal.py        # 内部 MCP APIs
│   ├── impex.py               # 导入/导出功能
│   ├── permission.py          # 权限管理
│   ├── tool_box.py            # 工具管理
│   ├── relations.py           # 关系管理
│   └── dataflow_como_operator.py  # Data flow COMO 集成
│
├── logs/                    # 测试执行日志存放目录
│
├── pipeline/                # CI/CD 流水线配置
│   └── *.yaml               # 各服务的流水线配置
│
├── reports/                 # 测试报告输出目录
│   ├── xml/                 # Allure XML 报告
│   └── html/                # Allure HTML 报告
│
├── requirements/            # Python 依赖管理
│   └── requirements.txt     # 主要依赖文件
│
├── resource/                # 测试资源文件目录（FTP 管理）
│   └── data-agent/          # data-agent 的资源文件
│
├── scripts/                 # 辅助脚本目录
│   └── operator-run.sh      # Operator 自动化测试脚本
│
└── testcases/               # 测试用例目录（新结构）
    ├── data-agent/          # data-agent 测试
    │   ├── api/             # API 测试
    │   │   ├── test_agent_factory_v3.py    # Agent Factory V3 测试
    │   │   └── test_agent_app.py           # Agent App 接口测试
    │   ├── functional/      # 功能测试
    │   │   └── test_agent_chat.py          # Agent 对话功能测试
    │   └── performance/     # 性能测试
    ├── data-operator-hub/   # data-operator-hub 测试
    │   ├── api/             # API 测试
    │   │   ├── category/    # 分类管理测试
    │   │   ├── mcp/         # MCP 测试
    │   │   └── permission/  # 权限测试
    │   ├── functional/      # 功能测试
    │   └── performance/     # 性能测试
    └── data-flow/           # data-flow 测试
        ├── api/             # API 测试
        ├── functional/      # 功能测试
        └── performance/     # 性能测试
```

**目录结构变更说明**:
- **2025-12-17 前结构**: `testcases/[api|functional|performance]/[service]/test_xxx.py`（已废弃）
- **2025-12-17 后结构**: `testcases/[service]/[test-type]/test_xxx.py`（推荐）
- 旧结构仍向后兼容，但新测试用例应使用新结构

# pytest 编码规范

## 1. 命名规范

### 1.1 驼峰命名
- **类名**: 使用大驼峰命名法（PascalCase），如 `TestUserAuthentication`、`ApiClient`
- **测试方法名**: 使用小驼峰命名法（camelCase），如 `testLoginSuccess`、`testDataValidation`
- **fixture 名**: 使用小驼峰命名法，如 `dbConnection`、`testUser`

### 1.2 文件命名
- 测试文件以 `test_` 开头，如 `test_login.py`、`test_api_integration.py`
- 使用下划线连接单词，全部小写

## 2. 测试文件结构

### 2.1 导入规范
```python
# 标准库导入放在最前
import os
import json

# 第三方库导入放在中间
import pytest
import requests

# 项目内模块导入放在最后
from common.utils import logging_util
from lib.data_agent.client import ApiClient
```

### 2.2 测试类组织
- 每个测试类专注于一个功能模块
- 相关的测试方法应放在同一个类中
- 类名应反映被测试的功能，如 `TestUserRegistration`

## 3. 测试方法设计

### 3.1 测试方法命名
- 名称应明确表达测试的内容和预期结果
- 推荐格式：`test<功能><条件><预期结果>`
- 例如：`testLoginWithValidCredentialsSuccess`、`testDataUploadWhenNetworkErrorRetry`

### 3.2 测试方法结构
遵循 Arrange-Act-Assert (AAA) 模式：
```python
def testSomething(self):
    # Arrange - 准备测试数据和环境
    user = createTestUser()

    # Act - 执行被测试的操作
    result = login(user.username, user.password)

    # Assert - 验证结果符合预期
    assert result.status == "success"
```

## 4. Fixtures 使用规范

### 4.1 Fixture 设计
- 将可重用的设置代码抽取为 fixtures
- 为 fixtures 提供清晰的作用域（scope）
- 遵循单一职责原则

### 4.2 Fixture 文档
```python
@pytest.fixture
def adminUser():
    """
    创建并返回一个具有管理员权限的测试用户。

    Returns:
        User: 具有管理员权限的用户对象
    """
    user = User(role="admin")
    yield user
    user.cleanup()
```

## 5. 断言最佳实践

### 5.1 使用明确的断言
```python
# 推荐
assert user.is_active is True, "新创建的用户应该处于活跃状态"

# 不推荐
assert user.is_active
```

### 5.2 复杂断言
- 对于复杂对象，验证关键属性而非整个对象
- 使用 pytest 的 `approx()` 进行浮点数比较

## 6. 参数化测试

```python
@pytest.mark.parametrize("username,expected_result", [
    ("valid_user", True),
    ("", False),
    ("user with spaces", False),
], ids=["valid", "empty", "with_spaces"])
def testUsernameValidation(username, expected_result):
    assert validateUsername(username) == expected_result
```

## 7. Allure 测试报告

### 7.1 使用 Allure 装饰器
```python
import allure

@allure.feature("用户管理")
@allure.story("用户登录")
@allure.title("测试有效用户登录")
def testUserLoginWithValidCredentials():
    # 测试代码
    pass
```

### 7.2 生成 Allure 报告
```bash
# 运行测试并生成 XML 报告
python3 -m pytest ./testcases/data-agent --alluredir ./report/xml --clean-alluredir

# 生成 HTML 报告
allure generate ./report/xml -o ./report/html --clean

# 打开报告
allure open ./report/html
```

# 本地调试指南

## 1. 测试环境镜像

**镜像地址**: `acr.aishu.cn/dip/agent-at:20250418`

**包含内容**:
- Python 3.9
- pytest 测试框架
- allure 报告工具
- JDK 8
- thriftAPI（公司内部 thrift API）

## 2. 本地运行步骤

### 2.1 克隆代码
```bash
git clone <repository-url> /root/agent-AT
cd /root/agent-AT
```

### 2.2 启动 Docker 容器
```bash
docker run -it --name agent-at --net=host \
  -v /root/agent-AT:/app \
  -v /root/.kube/config:/root/.kube/config \
  -v /usr/bin/kubectl:/usr/bin/kubectl \
  acr.aishu.cn/dip/agent-at:20250418 bash
```

**说明**:
- `-v /root/agent-AT:/app` - 映射本地代码到容器
- `-v /root/.kube/config:/root/.kube/config` - 映射 Kubernetes 配置
- `-v /usr/bin/kubectl:/usr/bin/kubectl` - 映射 kubectl 命令
- `--net=host` - 使用主机网络

### 2.3 配置 DNS
在容器的 `/etc/resolv.conf` 中添加：
```
nameserver 10.96.0.10
search anyshare.svc.cluster.local svc.cluster.local cluster.local
options ndots:2
```

### 2.4 运行测试
```bash
# 进入容器后
cd /app

# 运行所有测试
python3 -m pytest

# 运行特定测试
python3 -m pytest ./testcases/data-agent/api/test_agent_factory_v3.py -v -s

# 运行带 Allure 报告的测试
python3 -m pytest ./testcases/data-agent --alluredir ./report/xml --clean-alluredir
allure generate ./report/xml -o ./report/html --clean
```

## 3. 常用调试命令

```bash
# 查看可用的测试用例
python3 -m pytest --collect-only ./testcases/data-agent/api/test_agent_factory_v3.py

# 运行特定测试用例
python3 -m pytest ./testcases/data-agent/api/test_agent_factory_v3.py::TestAgentFactoryV3::test_create_agent_success -v -s

# 运行标记的测试
python3 -m pytest -m smoke -v

# 显示详细输出
python3 -m pytest -v -s --tb=short

# 只运行失败的测试
python3 -m pytest --lf

# 停在第一个失败的测试
python3 -m pytest -x
```

## 4. 容器管理

```bash
# 停止容器
docker stop agent-at

# 启动已存在的容器
docker start agent-at

# 进入运行中的容器
docker exec -it agent-at bash

# 删除容器
docker rm agent-at
```

# 测试开发模式

## 1. API 测试模式

```python
@pytest.mark.api
def testOperatorRegistration(self, Headers):
    """测试操作者注册"""
    client = OperatorClient(host=host)
    result = client.RegisterOperator(data, Headers)
    assert result[0] == 201
```

## 2. 数据驱动测试模式

```python
# 配置文件: data/data-agent/cfg.json
{
    "test_queries": {
        "single_turn": ["你好，请介绍一下你自己", "你能做什么？"],
        "multi_turn": ["什么是机器学习？", "详细说明", "举个例子"]
    }
}

# 测试代码
for query in test_queries["single_turn"]:
    result = agent.ChatCompletion(agent_id, {"query": query}, headers)
    validate_response(result)
```

## 3. Library Client 模式

所有 API 客户端遵循一致的模式：
```python
class Operator():
    def __init__(self):
        file = GetContent("./config/env.ini")
        self.config = file.config()
        self.base_url = self.config["requests"]["protocol"] + "://" + \
                       self.config["server"]["host"] + ":" + \
                       self.config["server"]["port"] + "/api/agent-operator-integration/v1/operator"
```

# 关键 Session Fixtures

项目提供以下关键 session-level fixtures（在 `conftest.py` 中）：

- **APrepare**: 创建测试组织、部门和用户
- **Headers**: 外部 API 的 Bearer token 认证
- **UserHeaders**: 内部 API 的认证
- **RoleMember**: 管理 AI 管理员角色分配
- **ModifyCM**: 配置 Kubernetes ConfigMap
- **AgentImport**: 从 `data/data-agent/import/` 导入 agents
- **ModelCheck**: 验证模型存在性和连接性

# 常见问题

## 1. 导入错误
确保已安装所有依赖：
```bash
pip install -r requirements/requirements.txt
```

## 2. 认证失败
检查 `config/env.ini` 中的服务器配置和认证信息是否正确。

## 3. Kubernetes 连接问题
确保 Kubernetes 配置文件已正确映射到容器中。

## 4. 测试数据缺失
检查 `data/` 目录中是否存在所需的测试数据文件。

# 贡献指南

1. 遵循项目编码规范
2. 使用新的目录结构：`testcases/[service]/[test-type]/test_xxx.py`
3. 为测试添加适当的 Allure 装饰器和文档
4. 确保所有测试通过后再提交代码
5. 更新相关文档

# 相关文档

- [CLAUDE.md](./CLAUDE.md) - Claude Code 项目指导文档
- pytest 官方文档: https://docs.pytest.org/
- Allure 文档: https://docs.qameta.io/allure/

# 版本历史

## 2025-12-24
- 更新 README.md 文档结构
- 添加新的目录结构说明（2025-12-17 更新）
- 完善 pytest 编码规范和本地调试指南
- 添加测试开发模式和常见问题解答
- 增强快速开始指南和贡献指南

## 2025-12-17
- 重大更新：目录结构重组
- 新结构：`testcases/[service]/[test-type]/test_xxx.py`
- 新增 Agent Factory V3 测试模块
- 新增 Agent App 外部 API 测试
- 新增 Agent Chat 功能测试
- 验证所有测试套件在新结构下正常工作

---

**项目维护者**: DIP 部门测试团队
**最后更新**: 2025-12-24
**文档版本**: 2.0