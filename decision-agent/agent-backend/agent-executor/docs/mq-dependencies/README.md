# MQ 依赖文档

## 文档说明

本目录包含项目中 MQ（消息队列）相关的依赖说明，特别是关于本地库（.so 文件）依赖的问题。

## 文档列表

- **[MQ 依赖和本地库问题说明](./mq-so-dependencies.md)** - 详细说明了 mq-sdk 各模块的依赖情况和解决方案

## 快速了解

### 可用的 MQ 模块（无需额外安装）
- ✅ NSQ - 纯 Python 实现
- ✅ Kafka - 依赖 kafka-python
- ✅ HTTP 客户端 - 依赖 requests

### 不可用的 MQ 模块（需要本地库）
- ❌ BMQ - 需要 `libbesmq-c.so`
- ❌ TongLink/TLQ - 需要客户端库

### 使用建议
1. 优先使用 NSQ 或 Kafka
2. 避免导入 `proton_mq`（会导入所有模块）
3. 使用条件导入处理不同环境
