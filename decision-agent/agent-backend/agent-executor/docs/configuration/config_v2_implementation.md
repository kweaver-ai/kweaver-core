# Agent-Executor 配置管理 V2 版本实现文档

## 概述

本文档描述了agent-executor配置管理v2版本的实现，该版本使用yaml配置文件替代环境变量的方式来管理配置，参考了agent-go-common-pkg/cconf/config.go的实现方式。

## 设计目标

1. **配置文件化**：使用yaml格式的配置文件，提高配置的可读性和可维护性
2. **向后兼容**：保持与原config_class.py相同的访问方式，确保现有代码无需修改
3. **独立部署**：v2版本独立于原版本，不影响现有功能，便于逐步迁移
4. **灵活配置**：支持多种配置路径，适应不同的部署环境

## 实现架构

### 1. 目录结构

```
agent-executor/
├── app/
│   └── config/
│       ├── config_class.py              # 原版本配置（保持不变）
│       └── config_v2/                   # v2版本配置
│           ├── __init__.py
│           ├── config_class_v2.py       # v2配置类实现
│           ├── README.md                # 使用说明
│           └── test_config_v2.py        # 测试脚本
├── helm/
│   └── agent-executor/
│       └── templates/
│           ├── configmap.yaml           # 原configmap（保持不变）
│           ├── configmap-v2.yaml        # v2版本configmap（新增）
│           └── deployment.yaml          # 更新以支持v2挂载
├── .local/
│   └── conf/
│       └── agent-executor.yaml          # 本地开发示例配置
└── docs/
    └── config_v2_implementation.md      # 本文档
```

### 2. 核心组件

#### 2.1 ConfigClassV2 (config_class_v2.py)

**主要功能**：
- 从yaml配置文件加载配置
- 提供与原ConfigClass相同的类属性访问方式
- 支持配置路径的灵活指定
- 自动初始化配置

**关键方法**：
- `_get_config_path()`: 获取配置文件路径（支持环境变量、挂载路径、本地路径）
- `_load_config_file()`: 加载yaml配置文件
- `_initialize()`: 初始化所有配置项
- `get_local_dev_model_config()`: 获取模型配置
- `is_local_dev()`: 判断是否本地开发模式
- `is_o11y_log_enabled()`: 判断是否启用o11y日志
- `is_o11y_trace_enabled()`: 判断是否启用o11y追踪

#### 2.2 ConfigMap-V2 (configmap-v2.yaml)

**主要功能**：
- 将helm values渲染为yaml格式的配置文件
- 通过ConfigMap挂载到容器的/sysvol/conf/目录
- 支持所有原有的配置项

**配置结构**：
```yaml
app:           # 应用相关配置
rds:           # 数据库配置
redis:         # Redis配置
graphdb:       # 图数据库配置
opensearch:    # OpenSearch配置
services:      # 依赖服务配置
external_services:  # 外部服务配置
memory:        # 记忆相关配置
document:      # 文档相关配置
local_dev:     # 本地开发配置
outer_llm:     # 外部LLM配置
features:      # 特性开关
o11y:          # 可观测性配置
dialog_logging: # 对话日志配置
```

#### 2.3 Deployment更新

**修改内容**：
- 在volumeMounts中添加v2版本配置文件的挂载点
- 在volumes中添加v2版本configmap的volume定义

**挂载路径**：
- ConfigMap名称: `{{ .Release.Name }}-yaml-v2`
- 挂载路径: `/sysvol/conf/`
- 配置文件: `agent-executor.yaml`

## 配置路径优先级

1. **环境变量指定路径**（最高优先级）
   - 环境变量: `AGENT_EXECUTOR_CONFIG_PATH`
   - 使用场景: 需要自定义配置路径时

2. **K8s挂载路径**
   - 路径: `/sysvol/conf/`
   - 使用场景: 生产环境K8s部署

3. **本地开发路径**（最低优先级）
   - 路径: `./conf/`
   - 使用场景: 本地开发调试

## 使用指南

### 1. 本地开发

```bash
# 1. 创建配置目录
mkdir -p ./conf

# 2. 复制示例配置
cp .local/conf/agent-executor.yaml ./conf/

# 3. 根据实际环境修改配置
vim ./conf/agent-executor.yaml

# 4. 在代码中使用
from app.config.config_v2 import ConfigClassV2
Config = ConfigClassV2()
```

### 2. K8s部署

```bash
# 1. 修改helm values
vim helm/agent-executor/values.yaml

# 2. 部署应用
helm upgrade --install agent-executor ./helm/agent-executor \
  -f helm/agent-executor/values.yaml \
  -n dip

# 配置会自动通过configmap-v2.yaml挂载到容器
```

### 3. 测试验证

```bash
# 运行测试脚本
cd /Users/Zhuanz/Work/as/dip_ws/agent-executor
python app/config/config_v2/test_config_v2.py
```

## 配置项映射

### 原环境变量 → V2配置文件映射

| 原环境变量 | V2配置路径 | 说明 |
|-----------|-----------|------|
| AGENT_EXECUTOR_DEBUG | app.debug | 调试模式 |
| HOST_IP | app.host_ip | 监听IP |
| PORT | app.port | 服务端口 |
| RDSHOST | rds.host | 数据库主机 |
| RDSPORT | rds.port | 数据库端口 |
| RDSDBNAME | rds.dbname | 数据库名 |
| RDSUSER | rds.user | 数据库用户 |
| RDSPASS | rds.password | 数据库密码 |
| REDISCLUSTERMODE | redis.cluster_mode | Redis连接模式 |
| REDISHOST | redis.host | Redis主机 |
| REDISPORT | redis.port | Redis端口 |
| GRAPHDB_HOST | graphdb.host | 图数据库主机 |
| OPENSEARCH_HOST | opensearch.host | OpenSearch主机 |
| EMB_URL | external_services.emb_url | Embedding服务URL |
| MEMORY_LIMIT | memory.limit | 记忆召回条数 |
| LOCAL_DEV | local_dev.enabled | 本地开发模式 |
| IS_USE_EXPLORE_BLOCK_V2 | features.use_explore_block_v2 | 特性开关 |
| O11Y_LOG_ENABLED | o11y.log_enabled | O11Y日志开关 |

完整映射请参考配置文件示例。

## 与原版本的对比

| 特性 | 原版本 (config_class.py) | v2版本 (config_v2) |
|------|-------------------------|-------------------|
| **配置来源** | 环境变量 | yaml配置文件 |
| **配置结构** | 扁平化（所有配置在同一层级） | 层次化（按功能分组） |
| **可读性** | 一般（环境变量列表） | 好（结构化yaml） |
| **维护性** | 一般（需要在多处修改） | 好（集中管理） |
| **访问方式** | `Config.XXX` | `Config.XXX`（完全相同） |
| **部署方式** | 环境变量注入 | ConfigMap挂载 |
| **配置验证** | 运行时发现 | 可提前验证yaml格式 |
| **版本控制** | 难以追踪 | 易于追踪（配置文件） |

## 迁移策略

### 阶段1：准备阶段（当前）
- ✅ 创建v2版本配置系统
- ✅ 保持原版本不变
- ✅ 提供示例配置和文档

### 阶段2：测试阶段（下一步）
1. 在开发环境使用v2版本
2. 验证所有配置项正确加载
3. 测试各个功能模块
4. 收集反馈并优化

### 阶段3：迁移阶段（后续）
1. 在测试环境部署v2版本
2. 逐步替换生产环境
3. 监控运行状态
4. 确认稳定后废弃原版本

### 阶段4：清理阶段（最后）
1. 移除原config_class.py
2. 将config_v2重命名为config
3. 更新所有导入语句
4. 清理相关文档

## 注意事项

### 1. 兼容性
- v2版本与原版本完全独立，可以共存
- 访问方式保持一致，现有代码无需修改
- 只需修改import语句即可切换版本

### 2. 安全性
- 敏感信息（如密码）应通过Secret挂载
- 不要在配置文件中硬编码敏感信息
- 生产环境配置应严格控制访问权限

### 3. 性能
- 配置在启动时一次性加载
- 使用类属性缓存，避免重复读取
- 对性能无明显影响

### 4. 调试
- 配置加载时会打印日志
- 可通过测试脚本验证配置
- 配置文件格式错误会在启动时报错

## 问题排查

### Q1: 配置文件未找到
**现象**: 启动时提示"Config file not found"

**解决方案**:
1. 检查配置文件路径是否正确
2. 确认配置文件名为`agent-executor.yaml`
3. 查看日志中的配置路径信息

### Q2: 配置项读取错误
**现象**: 某些配置项值不正确

**解决方案**:
1. 检查yaml格式是否正确（缩进、语法）
2. 确认配置项名称与文档一致
3. 查看配置加载日志

### Q3: 配置未生效
**现象**: 修改配置后未生效

**解决方案**:
1. 确认使用的是v2版本的配置类
2. 重启服务以重新加载配置
3. 检查是否有环境变量覆盖

### Q4: K8s部署配置问题
**现象**: K8s环境配置加载失败

**解决方案**:
1. 检查ConfigMap是否正确创建
2. 确认Volume挂载配置正确
3. 查看Pod日志排查问题

## 参考资料

- agent-go-common-pkg/cconf/config.go: Go版本配置管理实现
- agent-app/helm: agent-app的helm配置参考
- agent-app/conf/config.go: agent-app的配置实现

## 附录

### A. 完整配置示例

参见: `.local/conf/agent-executor.yaml`

### B. 测试脚本

参见: `app/config/config_v2/test_config_v2.py`

### C. 使用文档

参见: `app/config/config_v2/README.md`

---

**文档版本**: v1.0
**创建日期**: 2025-01-23
**最后更新**: 2025-01-23
**维护者**: Agent-Executor Team
