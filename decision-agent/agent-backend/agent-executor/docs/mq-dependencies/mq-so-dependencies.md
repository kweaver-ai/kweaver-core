# MQ 依赖和本地库问题说明

## 概述

本文档详细说明了项目中使用的 mq-sdk 及其本地库依赖问题。

## mq-sdk 模块结构

```
mq-sdk (proton-mq-python)
├── nsq_moudle.py      # NSQ 消息队列
├── kafka_module.py    # Kafka 客户端
├── bmq_moudle.py      # BMQ 消息队列
├── tongLink_moudle.py # TongLink/TLQ
├── tlqhtp2_moudle.py  # TLQ HTTP 2.0
├── tlqhtp202_module.py # TLQ HTTP 202
└── proton_mq.py       # 统一接口（导入所有模块）
```

## 依赖情况分析

### ✅ 无需本地库的模块（随时可用）

| 模块 | 功能 | 依赖 | 使用场景 |
|------|------|------|----------|
| `nsq_moudle` | NSQ 消息队列 | 纯 Python | 轻量级消息队列 |
| `kafka_module` | Kafka 客户端 | `kafka-python` | 分布式流处理 |
| `http_client` | HTTP API 客户端 | `requests` | REST API 调用 |
| `logger` | 日志模块 | 纯 Python | 日志记录 |

### ❌ 需要本地库的模块

| 模块 | 功能 | 需要的本地库 | 报错时机 |
|------|------|-------------|----------|
| `bmq_moudle` | BMQ 消息队列 | `libbesmq-c.so` | **导入即报错** |
| `tongLink_moudle` | TongLink/TLQ | `libtlq.so` | **导入即报错** |
| `tlqhtp2_moudle` | TLQ HTTP 2.0 | `libtlqhtp2.so` | **导入即报错** |
| `tlqhtp202_module` | TLQ HTTP 202 | `libtlqhtp202.so` | **导入即报错** |

## 报错时机详解

### 1. 导入时立即报错的模块

```python
# 这些模块在 import 时就会检查本地库
import mq_sdk.bmq_moudle  # ❌ 立即报错
```

**错误信息**：
```
OSError: dlopen(libbesmq-c.so, 0x0006): tried:
  'libbesmq-c.so' (no such file),
  '/usr/lib/libbesmq-c.so' (no such file, not in dyld cache)
```

### 2. proton_mq 的问题

`proton_mq.py` 在第 7 行导入了所有模块：
```python
from mq_sdk import bmq_moudle, http_client, logger, kafka_module, nsq_moudle, tongLink_moudle, tlqhtp2_moudle, tlqhtp202_module
```

这导致：
```python
from mq_sdk.proton_mq import Connector  # ❌ 导入就失败
```

## TelemetrySDK-Python 中的使用

### 使用场景

1. **基础日志功能**（不依赖 mq-sdk）
   ```python
   from tlogging import SamplerLogger
   # ✅ 完全正常
   ```

2. **数据导出到消息队列**（依赖 mq-sdk）
   ```python
   from exporter.public.client import Client
   # ❌ 因为导入了 mq_sdk.proton_mq
   ```

### 实际影响

- ✅ 可以正常使用 `tlogging` 进行日志记录
- ❌ 无法使用需要消息队列的导出功能
- ⚠️ 即使只想用 NSQ/Kafka，也会因为 `proton_mq` 导入所有模块而失败

## 解决方案

### 方案1：条件导入（推荐）

```python
# 检查模块可用性
def get_available_mq_modules():
    available = []
    modules = {
        'nsq': 'mq_sdk.nsq_moudle',
        'kafka': 'mq_sdk.kafka_module',
        'bmq': 'mq_sdk.bmq_moudle',
    }

    for name, module in modules.items():
        try:
            exec(f'import {module}')
            available.append(name)
        except ImportError:
            pass

    return available

# 使用示例
if 'nsq' in get_available_mq_modules():
    from mq_sdk import nsq_moudle
    # 使用 NSQ
```

### 方案2：直接导入可用模块

```python
# 避免使用 proton_mq
from mq_sdk import nsq_moudle  # ✅ 可用
from mq_sdk import kafka_module  # ✅ 可用

# 不要这样做：
# from mq_sdk.proton_mq import Connector  # ❌ 会失败
```

### 方案3：安装本地库

如果确实需要 BMQ 或 TongLink：

```bash
# BMQ - 需要安装 libbesmq-c.so
# 具体安装方法取决于系统

# TongLink - 需要安装客户端库
# 联系相关团队获取安装包
```

## 开发环境配置

### 验证脚本

```python
#!/usr/bin/env python
"""检查 MQ 模块可用性"""

def check_mq_modules():
    print("检查 MQ 模块可用性...")

    modules = {
        'NSQ': 'mq_sdk.nsq_moudle',
        'Kafka': 'mq_sdk.kafka_module',
        'BMQ': 'mq_sdk.bmq_moudle',
        'TongLink': 'mq_sdk.tongLink_moudle',
        'TLQ HTTP2': 'mq_sdk.tlqhtp2_moudle',
        'TLQ HTTP202': 'mq_sdk.tlqhtp202_module',
    }

    available = []
    unavailable = []

    for name, module in modules.items():
        try:
            exec(f'import {module}')
            available.append(name)
            print(f"✅ {name} - 可用")
        except ImportError as e:
            unavailable.append((name, str(e)))
            print(f"❌ {name} - 不可用")

    print(f"\n总结：")
    print(f"可用模块 ({len(available)}): {', '.join(available)}")
    print(f"不可用模块 ({len(unavailable)}): {', '.join([n for n, _ in unavailable])}")

    if unavailable:
        print("\n不可用模块原因：")
        for name, error in unavailable:
            if 'libbesmq-c.so' in error:
                print(f"  - {name}: 缺少 libbesmq-c.so")
            elif 'libtlq' in error:
                print(f"  - {name}: 缺少 TongLink 客户端库")

if __name__ == "__main__":
    check_mq_modules()
```

## 生产环境建议

1. **优先使用纯 Python 实现**
   - NSQ：适合大多数场景
   - Kafka：适合高吞吐量场景

2. **避免使用需要本地库的模块**
   - 除非有明确的业务需求
   - 确保生产环境已安装相应库

3. **使用条件导入**
   - 提高代码的健壮性
   - 支持多种部署环境

## 常见问题

### Q: 为什么有些模块导入就报错？
A: 这些模块在顶层代码中就加载了本地库（.so 文件），如果库不存在就会立即报错。

### Q: 可以只使用 NSQ/Kafka 吗？
A: 可以。直接导入对应模块，不要通过 `proton_mq`。

### Q: 如何解决本地库问题？
A:
1. 安装对应的客户端库
2. 或使用纯 Python 的替代方案
3. 或使用条件导入

### Q: TelemetrySDK-Python 必须使用消息队列吗？
A: 不是。基础的日志功能完全不需要消息队列，只有数据导出功能才需要。

## 参考资料

- [proton-mq-python 源码](file:///Users/Zhuanz/Work/as/dip_ws/proton-mq-python)
- [TelemetrySDK-Python 源码](file:///Users/Zhuanz/Work/as/dip_ws/TelemetrySDK-Python)
- [项目 Makefile](../../Makefile.local.mk)
