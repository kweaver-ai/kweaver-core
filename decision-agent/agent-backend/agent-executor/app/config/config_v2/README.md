# 配置管理 V2 版本

## 概述

配置管理v2版本使用yaml配置文件替代环境变量的方式来管理配置，参考了agent-go-common-pkg/cconf/config.go的实现方式。使用dataclass模型按功能模块组织配置项，提高代码的可维护性和可读性。

## 主要特性

1. **配置文件化**：使用yaml格式的配置文件，结构清晰，易于维护
2. **模块化设计**：使用dataclass将配置按功能模块拆分，便于管理
3. **单例模式**：确保全局只有一个配置实例，配置只加载一次
4. **路径灵活**：支持通过环境变量指定配置文件路径
5. **兼容性**：保持与原config_class.py相同的访问方式（如`Config.DEBUG`, `Config.HOST_IP`等）
6. **独立性**：v2版本独立于原版本，不影响现有功能

## 目录结构

```
app/config/
├── config_class.py          # 原版本配置（保持不变）
├── config_v2/               # v2版本配置
│   ├── __init__.py
│   ├── config_class_v2.py   # v2配置类
│   ├── models/              # 配置模型
│   │   ├── __init__.py
│   │   ├── app_config.py           # 应用配置
│   │   ├── database_config.py      # 数据库配置
│   │   ├── service_config.py       # 服务配置
│   │   ├── feature_config.py       # 功能特性配置
│   │   └── observability_config.py # 可观测性配置
│   └── README.md            # 本文档
```

## 配置模型说明

### 1. AppConfig (app_config.py)
应用相关配置，包括：
- 调试模式
- 服务监听配置
- API路径前缀
- 限流配置
- 日志配置

### 2. 数据库配置 (database_config.py)
- **RdsConfig**: 关系型数据库配置
- **RedisConfig**: Redis数据库配置（支持standalone/sentinel/master-slave模式）
- **GraphDBConfig**: 图数据库配置
- **OpenSearchConfig**: OpenSearch配置

### 3. 服务配置 (service_config.py)
- **ServicesConfig**: 依赖服务配置（模型服务、Agent服务、知识服务等）
- **ExternalServicesConfig**: 外部服务配置（Embedding、Rerank等）

### 4. 功能特性配置 (feature_config.py)
- **MemoryConfig**: 记忆相关配置
- **DocumentConfig**: 文档问答配置
- **LocalDevConfig**: 本地开发模式配置
- **OuterLLMConfig**: 外部LLM配置
- **FeaturesConfig**: 特性开关配置

### 5. 可观测性配置 (observability_config.py)
- **O11yConfig**: 可观测性配置
- **DialogLoggingConfig**: 对话日志配置

## 配置文件路径

配置文件查找优先级：

1. **环境变量指定路径**：`AGENT_EXECUTOR_CONFIG_PATH`
2. **K8s挂载路径**：`/sysvol/conf/`（生产环境）
3. **本地开发路径**：`./conf/`（本地开发）

配置文件名：`agent-executor.yaml`

## 使用方法

### 1. 直接访问配置模型（推荐 - 新方式）

`ConfigClassV2`继承自`ConfigState`，使用单例模式，可以直接访问配置模型：

```python
from app.config.config_v2 import ConfigClassV2

# 初始化配置（单例模式，多次调用返回同一实例）
Config = ConfigClassV2()

# 直接访问配置模型（推荐）
print(Config.app.debug)           # 应用配置
print(Config.app.host_ip)
print(Config.app.app_root)

print(Config.rds.host)             # 数据库配置
print(Config.rds.port)

print(Config.redis.cluster_mode)  # Redis配置
print(Config.redis.host)

print(Config.services.agent_factory.port) # 服务配置

print(Config.memory.limit)         # 记忆配置
print(Config.features.use_explore_block_v2)  # 特性开关
```

### 2. 向后兼容方式（原方式）

为了保持向后兼容，仍然支持原有的大写属性访问方式：

```python
from app.config.config_v2 import ConfigClassV2

Config = ConfigClassV2()

# 原方式访问（向后兼容）
print(Config.DEBUG)
print(Config.HOST_IP)
print(Config.RDSHOST)
print(Config.HOST_AGENT_APP)
```

### 3. 导入特定配置模型

```python
from app.config.config_v2 import AppConfig, RdsConfig, ServicesConfig

# 如果需要单独使用某个配置模型
app_config = AppConfig.from_dict({
    'debug': True,
    'port': 30778,
    'log_level': 'debug'
})

print(app_config.debug)
print(app_config.port)
```

### 访问方式对比

| 访问方式 | 示例 | 推荐度 | 说明 |
|---------|------|--------|------|
| **直接访问模型** | `Config.app.debug` | ⭐⭐⭐⭐⭐ | 推荐新代码使用，结构清晰 |
| **向后兼容** | `Config.DEBUG` | ⭐⭐⭐ | 保持旧代码兼容 |
| **独立模型** | `AppConfig.from_dict({...})` | ⭐⭐⭐⭐ | 测试或独立使用 |

### 单例模式说明

`ConfigClassV2` 使用单例模式实现，确保全局只有一个配置实例：

```python
from app.config.config_v2 import ConfigClassV2

# 多次调用返回同一个实例
config1 = ConfigClassV2()
config2 = ConfigClassV2()
config3 = ConfigClassV2()

print(config1 is config2)  # True
print(config2 is config3)  # True

# 配置对象也是共享的
print(config1.app is config2.app)  # True
```

**单例模式的优势**：
- ✅ 配置文件只加载一次，提高性能
- ✅ 内存占用更少
- ✅ 配置状态全局一致
- ✅ 避免重复初始化

### 2. 本地开发

将示例配置文件复制到本地：

```bash
# 复制示例配置
cp .local/conf/agent-executor.yaml ./conf/agent-executor.yaml

# 根据实际环境修改配置
vim ./conf/agent-executor.yaml
```

### 3. K8s部署

配置通过helm的configmap-v2.yaml自动生成并挂载到容器的`/sysvol/conf/`目录。

修改配置：
1. 编辑 `helm/agent-executor/values.yaml`
2. 配置会自动渲染到 `configmap-v2.yaml`
3. 部署时自动挂载到容器

### 4. 指定自定义配置路径

```bash
# 通过环境变量指定配置文件所在目录
export AGENT_EXECUTOR_CONFIG_PATH=/path/to/your/config/dir
```

## 配置项说明

### 1. app相关
- `debug`: 是否开启调试模式
- `host_ip`: 服务监听IP
- `port`: 服务端口
- `host_prefix`: API路径前缀v1
- `host_prefix_v2`: API路径前缀v2
- `rps_limit`: 每秒最大请求数
- `enable_system_log`: 是否启用系统日志
- `log_level`: 日志级别

### 2. rds相关
- `host`: 数据库主机
- `port`: 数据库端口
- `dbname`: 数据库名
- `user`: 数据库用户
- `password`: 数据库密码
- `db_type`: 数据库类型

### 3. redis相关
- `cluster_mode`: 连接模式（standalone/sentinel/master-slave）
- 根据不同模式配置相应的连接参数

### 4. graphdb相关
- `host`: 图数据库主机
- `port`: 图数据库端口
- `type`: 图数据库类型
- `read_only_user`: 只读用户
- `read_only_password`: 只读密码

### 5. opensearch相关
- `host`: opensearch主机
- `port`: opensearch端口
- `user`: opensearch用户
- `password`: opensearch密码

### 6. services相关
配置各个依赖服务的地址和端口

### 7. external_services相关
- `emb_url`: embedding服务地址
- `embedding_dimension`: embedding维度
- `rerank_url`: rerank服务地址

### 8. memory相关
- `limit`: 召回记忆条数
- `threshold`: 召回记忆向量评分阈值
- `rerank_threshold`: 召回记忆重排序评分阈值

### 9. document相关
- `enable_sensitive_word_detection`: 是否启用敏感词检测

### 10. local_dev相关
- `enabled`: 是否本地开发模式

### 11. outer_llm相关
- `api`: 外部LLM API地址
- `api_key`: 外部LLM API密钥
- `model`: 外部LLM模型名
- `model_list`: 模型列表配置

### 12. features相关
- `use_explore_block_v2`: 是否使用explore_block v2版本

### 13. o11y相关
- `log_enabled`: 是否启用o11y日志
- `trace_enabled`: 是否启用o11y追踪

### 14. dialog_logging相关
- `enable_dialog_logging`: 是否启用对话日志
- `use_single_log_file`: 是否使用单一日志文件
- `single_profile_file_path`: profile日志文件路径
- `single_trajectory_file_path`: trajectory日志文件路径

## 与原版本的区别

| 特性 | 原版本 (config_class.py) | v2版本 (config_v2) |
|------|-------------------------|-------------------|
| 配置来源 | 环境变量 | yaml配置文件 |
| 配置结构 | 扁平化 | 层次化 |
| 可读性 | 一般 | 好 |
| 维护性 | 一般 | 好 |
| 访问方式 | `Config.XXX` | `Config.XXX`（相同） |

## 迁移指南

从原版本迁移到v2版本：

1. **准备配置文件**：根据示例配置文件创建`agent-executor.yaml`
2. **修改导入**：将`from app.config.config_class import ConfigClass`改为`from app.config.config_v2 import ConfigClassV2`
3. **测试验证**：确保所有配置项都能正确读取
4. **逐步替换**：建议先在开发环境测试，确认无误后再部署到生产环境

## 注意事项

1. **向后兼容**：v2版本不影响原版本，两者可以共存
2. **配置优先级**：环境变量 > 挂载路径 > 本地路径
3. **配置文件格式**：必须是有效的yaml格式
4. **敏感信息**：生产环境的敏感配置应通过secret挂载

## 示例

完整的配置文件示例请参考：`.local/conf/agent-executor.yaml`

## 问题排查

### 配置文件未找到
- 检查配置文件路径是否正确
- 检查配置文件名是否为`agent-executor.yaml`
- 查看日志中的配置文件加载信息

### 配置项读取错误
- 检查yaml格式是否正确
- 检查配置项名称是否匹配
- 查看日志中的错误信息

### 配置未生效
- 确认使用的是v2版本的配置类
- 检查配置文件是否被正确加载
- 重启服务确保配置重新加载
