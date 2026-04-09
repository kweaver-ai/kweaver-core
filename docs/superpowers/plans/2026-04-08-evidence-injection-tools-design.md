# 证据注入功能设计文档（工具+后处理器模式）

**项目**: Decision Agent - 证据注入
**日期**: 2026-04-08
**版本**: v4.5（工具+后处理器模式，位置索引标注，EvidenceStore 管理，统一 LLM 提取，无 Prompt 注入，包含 Source 信息）
**状态**: 设计中

## 1. 需求概述

### 1.1 背景

Decision Agent 在调用工具（如 `kn_search`）后需要动态识别和标注证据引用。现有方案 `online_search_cite_tool` 存在缺点：
- 使用两次 LLM 调用，延迟高（4-6秒）
- 非流式输出，用户体验差
- 引用格式固定，不支持灵活配置

需要设计一个更优的解决方案。

**关键约束**：
- 使用**内置工具+后处理器**方式实现
- **最小化**对 agent-executor 代码的改动
- 支持**流式输出**
- 解决大多数工具没有 `_evidence` 字段的问题

### 1.2 目标

设计证据注入功能，包含：

1. **evidence_prepare（工具）**: 提取和标准化证据，由 Agent 调用
2. **EvidenceInjectProcessor（后处理器）**: 流式注入证据标注，包装 LLM 输出流

### 1.3 设计方案

| 维度 | 决策 |
|------|------|
| 证据提取 | **内置工具** - `evidence_prepare`，Agent 主动调用 |
| 证据管理 | **EvidenceStore** - 集中管理证据生命周期，不直接污染 context |
| 标注注入 | **后处理器** - `EvidenceInjectProcessor`，包装 LLM 输出流 |
| 输出格式 | **位置索引** - 使用 `start`/`end` 位置标注，非内联标签 |
| 流式支持 | 后处理器接收 LLM 流式输出，边接收边处理 |
| 状态管理 | 通过 EvidenceStore 管理，context 只存储 evidence_store_key |
| executor 改动 | **最小改动**：工具注册 + 流式输出包装 |

## 2. 架构设计

### 2.1 整体架构

```
┌─────────────────────────────────────────────────────────────────────────┐
│                  证据注入架构（工具+后处理器模式）                      │
│                                                                         │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐             │
│  │ Step 1: Agent │    │ Step 2: Agent │    │ Step 3: LLM   │             │
│  │ 调用业务工具   │───▶│ 调用         │───▶│ 生成内容     │             │
│  │ (kn_search)   │    │ evidence_    │    │ (流式)       │             │
│  └──────────────┘    │ prepare      │    └──────┬───────┘             │
│       │              └──────┬───────┘              │                     │
│       │                     │                        │                     │
│       ▼                     ▼                        ▼                     │
│  ┌──────────────────────────────────────────────────────────┐        │
│  │              工具返回 → evidence_prepare 输入                │        │
│  │  ┌────────────────────────────────────────────────────┐ │        │
│  │  │ 输入: tool_call_result                              │ │        │
│  │  │ 处理:                                              │ │        │
│  │  │ 1. 使用 LLM 提取关键实体                            │ │        │
│  │  │ 2. 如果 LLM 失败 → 规则回退提取                     │ │        │
│  │  │ 输出: 标准化 evidence 列表 + evidence_id               │        │
│  │  └────────────────────────────────────────────────────┘ │        │
│  └──────────────────────────────────────────────────────────┘        │
│                              │                                       │
│                              ▼                                       │
│  ┌──────────────────────────────────────────────────────────┐        │
│  │              EvidenceStore 管理证据                        │        │
│  │  ┌────────────────────────────────────────────────────┐  │        │
│  │  │ evidence_store.add(evidence_id, evidences)         │  │        │
│  │  │                                                    │  │        │
│  │  │ EvidenceStore 职责:                                │  │        │
│  │  │ - 集中管理证据生命周期                            │  │        │
│  │  │ - 提供 LRU 缓存机制                               │  │        │
│  │  │ - 支持证据查询和清理                              │  │        │
│  │  └────────────────────────────────────────────────────┘  │        │
│  │                                                            │        │
│  │  context.set_variable("evidence_store_key", evidence_id)  │        │
│  └──────────────────────────────────────────────────────────┘        │
│                              │                                       │
│                              ▼                                       │
│  ┌──────────────────────────────────────────────────────────┐        │
│  │              LLM 流式生成 → EvidenceInjectProcessor        │        │
│  │  ┌────────────────────────────────────────────────────┐  │        │
│  │  │ 后处理器包装:                                       │  │        │
│  │  │ 1. 从 EvidenceStore 获取 evidences                  │  │        │
│  │  │ 2. 接收 LLM 流式输出 chunk by chunk                │  │        │
│  │  │ 3. 累积文本直到句子边界                              │  │        │
│  │  │ 4. 对每个句子进行规则匹配                            │  │        │
│  │  │ 5. 返回位置索引 positions: [[start, end], ...]     │  │        │
│  │  │ 6. 输出原始文本 + _evidence 元数据                   │  │        │
│  │  └────────────────────────────────────────────────────┘  │        │
│  │                                                              │ │
│  │              输出: streaming response with _evidence   │ │
│  │                                                              │ │
│  │              最终输出格式:                                │ │
│  │              {                                          │ │
│  │                "answer": "我的名字是张三，我住上海。",   │ │
│  │                "context": {...},                         │ │
│  │                "_evidence": {                            │ │
│  │                  "e1": {                                 │ │
│  │                    "label": "张三",      # 标签显示文本  │ │
│  │                    "match_text": "张三", # 匹配的关键字  │ │
│  │                    "positions": [[5, 7]],  # 关键字位置   │ │
│  │                    "source": {       # 点击标签显示详情   │ │
│  │                      "type": "kn_search",               │ │
│  │                      "tool_name": "知识网络搜索",       │ │
│  │                      "tool_call_id": "call_xxx",       │ │
│  │                      "tool_result": {...}  # 完整结果   │ │
│  │                    }                                     │ │
│  │                  },                                      │ │
│  │                  "e2": {                                 │ │
│  │                    "label": "上海",                      │ │
│  │                    "match_text": "上海",                │ │
│  │                    "positions": [[10, 12]],              │ │
│  │                    "source": {                           │ │
│  │                      "type": "kn_search",               │ │
│  │                      "tool_name": "知识网络搜索",       │ │
│  │                      "tool_call_id": "call_xxx",       │ │
│  │                      "tool_result": {...}              │ │
│  │                    }                                     │ │
│  │                  }                                       │ │
│  │                }                                         │ │
│  │              }                                          │ │
│  └──────────────────────────────────────────────────────────────┘        │
│                                                                          │
│                              Agent 代码改动最小化:                       │
│                              1. 注册 evidence_prepare 工具             │
│                              2. 在 run_dolphin 中添加后处理器包装       │
│                              3. 在 prompt 中注入证据信息                │
│                              4. context 只存储 evidence_store_key       │
└─────────────────────────────────────────────────────────────────────────┘
```

### 2.2 数据流

```
1. Agent 调用业务工具（如 kn_search）
   └─ 返回: {"name": "张三", "address": "上海"}

2. Agent 调用 evidence_prepare（内置工具）
   ├─ 输入: tool_call_result
   ├─ 处理: LLM 提取关键实体 (失败时规则回退)
   └─ 输出: {
       "evidence_id": "ev_xxx",
       "evidences": [
           {"local_id": "e1", "object_type_name": "张三", ...},
           {"local_id": "e2", "object_type_name": "上海", ...}
       ]
   }

3. run_dolphin 将证据保存到 EvidenceStore
   ├─ evidence_store.add(evidence_id, evidences)
   └─ context.set_variable("evidence_store_key", evidence_id)
       # context 只存储 key，不存储完整的 evidences 数据

4. LLM 开始流式生成
   └─ 输出: "我的名字是张三，我住上海..."

5. EvidenceInjectProcessor 从 EvidenceStore 获取证据并包装 LLM 输出流
   ├─ evidences = evidence_store.get(evidence_id)
   ├─ 接收流式 chunk
   ├─ 累积到句子边界
   ├─ 规则匹配，返回位置索引
   └─ 输出: {
       "answer": "我的名字是张三，我住上海...",
       "_evidence": {
           "e1": {"object_type_name": "张三", "positions": [[4, 6]]},
           "e2": {"object_type_name": "上海", "positions": [[9, 11]]}
       }
   }
```

### 2.3 改动范围

**需要改动的文件**：

| 文件 | 改动类型 | 说明 |
|------|---------|------|
| `app/logic/tool/evidence_prepare.py` | **新增** | 提取和标准化证据的工具 |
| `app/logic/agent_core_logic_v2/evidence_inject_processor.py` | **新增** | 流式注入标注的后处理器 |
| `app/logic/agent_core_logic_v2/evidence_store.py` | **新增** | 证据存储管理器 |
| `app/common/utils/sentence_boundary_detector.py` | **新增** | 句子边界检测工具 |
| `app/logic/agent_core_logic_v2/run_dolphin.py` | **小幅修改** | 集成证据准备和后处理器 |
| `data_migrations/init/tools/openapi/evidence_tools.json` | **新增** | 工具的 OpenAPI 定义 |

**不需要改动**：
- ✅ 流式响应处理框架（streaming_response_handler.py）
- ✅ 中间件架构
- ✅ Agent 核心执行逻辑

## 3. 核心组件设计

### 3.1 evidence_prepare 工具

**位置**: `app/logic/tool/evidence_prepare.py`

**功能**: 提取和标准化工具调用结果中的证据

**函数签名**:
```python
async def evidence_prepare(
    tool_call_result: Dict,
    config: Optional[Dict] = None,
    context: Optional[Dict] = None,
) -> Dict:
    """
    准备证据：从工具调用结果中提取和标准化证据

    Args:
        tool_call_result: 工具调用返回的完整结果
        config: 配置参数
        context: 上下文信息

    Returns:
        {
            "evidence_id": "ev_uuid",  # 证据集 ID
            "evidences": [
                {
                    "local_id": "e1",  # 短 ID，用于 REF 标签
                    "object_type_name": "张三",
                    "aliases": ["员工张三", "张工"],
                    "object_type_id": "ot_employee",
                    "score": 0.95
                },
                ...
            ],
            "summary": "张三、上海等 2 个证据"
        }
    """
```

**实现要点**:

```python
import json
import uuid
from typing import Dict, Optional, List

async def evidence_prepare(
    tool_call_result: Dict,
    config: Optional[Dict] = None,
    context: Optional[Dict] = None,
) -> Dict:
    """准备证据：从工具调用结果中提取和标准化证据"""
    config = config or {}

    # 1. 统一使用 LLM 提取关键实体
    if config.get("enable_llm_extraction", True):
        return await _extract_by_llm(tool_call_result, config)

    # 2. 不使用 LLM 时，返回空证据
    return {
        "evidence_id": f"ev_{uuid.uuid4().hex[:8]}",
        "evidences": [],
        "summary": "无证据"
    }


async def _extract_by_llm(tool_call_result: Dict, config: Dict) -> Dict:
    """使用 LLM 提取实体"""
    from app.driven.dip.model_api_service import model_api_service

    prompt = f"""
请从以下工具返回结果中提取关键实体（人名、地名、组织名等），只提取明确的实体名称：

工具返回结果：
{json.dumps(tool_call_result, ensure_ascii=False, indent=2)}

请以 JSON 格式返回，格式如下：
{{
    "entities": [
        {{
            "name": "实体名称（必须精确）",
            "aliases": ["可能的别名1"],
            "type": "person|location|organization|other",
            "confidence": 0.9
        }}
    ]
}}

注意：
- 只提取明确的、有意义的实体名称
- 不要提取通用词汇（如"结果"、"数据"等）
- name 字段必须是工具结果中出现的精确文本
- 只返回 JSON，不要其他内容。
"""

    try:
        result = await model_api_service.call(
            model=config.get("llm_model", "default"),
            messages=[{"role": "user", "content": prompt}],
            max_tokens=2000,
            userid=config.get("userid", ""),
            timeout=config.get("timeout", 30),
        )

        # 尝试解析 JSON
        llm_result = _parse_llm_result(result)

    except Exception as e:
        # LLM 调用失败，使用规则提取
        llm_result = _fallback_extraction(tool_call_result)

    # 转换为标准格式
    evidence_id = f"ev_{uuid.uuid4().hex[:8]}"
    processed = []

    for i, entity in enumerate(llm_result.get("entities", [])):
        local_id = f"e{i+1}"
        processed.append({
            "local_id": local_id,
            "object_type_name": entity["name"],
            "aliases": entity.get("aliases", []),
            "object_type_id": f"auto_{entity.get('type', 'entity')}_{i}",
            "score": entity.get("confidence", 0.0),
        })

    return {
        "evidence_id": evidence_id,
        "evidences": processed,
        "summary": _generate_summary(processed)
    }


def _parse_llm_result(result: str) -> Dict:
    """解析 LLM 返回结果"""
    # 尝试直接解析 JSON
    try:
        return json.loads(result.strip())
    except json.JSONDecodeError:
        pass

    # 尝试提取 JSON 代码块
    import re
    json_match = re.search(r'```json\s*(.+?)\s*```', result, re.DOTALL)
    if json_match:
        try:
            return json.loads(json_match.group(1))
        except json.JSONDecodeError:
            pass

    # 尝试提取裸 JSON
    json_match = re.search(r'\{.*\}', result, re.DOTALL)
    if json_match:
        try:
            return json.loads(json_match.group(0))
        except json.JSONDecodeError:
            pass

    # 解析失败，返回空结果
    return {"entities": []}


def _fallback_extraction(tool_call_result: Dict) -> Dict:
    """回退策略：使用规则提取"""
    # 简单实现：从字符串值中提取可能的实体
    entities = []

    def extract_from_dict(obj, path=""):
        if isinstance(obj, dict):
            for key, value in obj.items():
                extract_from_dict(value, f"{path}.{key}" if path else key)
        elif isinstance(obj, list):
            for item in obj:
                extract_from_dict(item, path)
        elif isinstance(obj, str):
            # 简单的命名实体识别规则
            if len(obj) >= 2 and len(obj) <= 10:
                # 可能是人名或地名
                if any(char in obj for char in "张王李赵刘陈杨黄周吴"):
                    entities.append({
                        "name": obj,
                        "aliases": [],
                        "type": "person",
                        "confidence": 0.5
                    })

    extract_from_dict(tool_call_result)
    return {"entities": entities[:10]}  # 最多返回10个


def _generate_summary(evidences: List[Dict]) -> str:
    """生成证据摘要"""
    if not evidences:
        return "无证据"

    names = [e["object_type_name"] for e in evidences[:5]]
    if len(evidences) > 5:
        return f"{', '.join(names)} 等 {len(evidences)} 个证据"
    return f"{', '.join(names)} 等 {len(evidences)} 个证据"
```

### 3.2 EvidenceStore 证据存储管理器

**位置**: `app/logic/agent_core_logic_v2/evidence_store.py`

**功能**: 集中管理证据的生命周期，提供 LRU 缓存机制

**设计理由**:
- **解耦**: 不直接污染 context，context 只存储 `evidence_store_key`
- **可扩展**: 支持更复杂的证据管理功能（过期、LRU、持久化等）
- **线程安全**: 支持并发访问

**类设计**:
```python
import threading
import time
from typing import Dict, List, Optional
from collections import OrderedDict

class EvidenceStore:
    """证据存储管理器 - LRU 缓存模式"""

    def __init__(self, max_size: int = 1000, ttl_seconds: int = 3600):
        """
        Args:
            max_size: 最大缓存条目数，默认 1000
            ttl_seconds: 证据过期时间（秒），默认 1 小时
        """
        self.max_size = max_size
        self.ttl_seconds = ttl_seconds

        # 使用 OrderedDict 实现 LRU
        self._store: OrderedDict[str, Dict] = OrderedDict()
        self._timestamps: Dict[str, float] = {}

        # 线程锁
        self._lock = threading.RLock()

        # 统计信息
        self._stats = {
            "hits": 0,
            "misses": 0,
            "evictions": 0,
        }

    def add(
        self,
        evidence_id: str,
        evidences: List[Dict],
        tool_call_id: str = None,
        tool_name: str = None,
        tool_type: str = None,
        tool_result: Dict = None,
    ) -> None:
        """
        添加证据到存储

        Args:
            evidence_id: 证据集 ID
            evidences: 标准化的证据列表
            tool_call_id: 工具调用 ID
            tool_name: 工具名称
            tool_type: 工具类型
            tool_result: 工具返回的完整结果
        """
        with self._lock:
            # 检查是否需要淘汰
            self._ensure_capacity()

            # 添加或更新（移到末尾表示最近使用）
            if evidence_id in self._store:
                self._store.move_to_end(evidence_id)
            else:
                self._store[evidence_id] = {
                    "evidences": evidences,
                    "tool_call_id": tool_call_id,
                    "tool_name": tool_name,
                    "tool_type": tool_type,
                    "tool_result": tool_result,
                    "created_at": time.time(),
                }
                self._store.move_to_end(evidence_id)

            # 更新时间戳
            self._timestamps[evidence_id] = time.time()

    def get(self, evidence_id: str) -> Optional[Dict]:
        """
        获取证据

        Args:
            evidence_id: 证据集 ID

        Returns:
            {
                "evidences": List[Dict],  # 证据列表
                "tool_call_id": str,      # 工具调用 ID
                "tool_name": str,         # 工具名称
                "tool_type": str,         # 工具类型
                "tool_result": Dict,      # 工具返回结果
            }
            如果不存在或已过期返回 None
        """
        with self._lock:
            # 检查是否存在
            if evidence_id not in self._store:
                self._stats["misses"] += 1
                return None

            # 检查是否过期
            entry = self._store[evidence_id]
            if self._is_expired(entry):
                del self._store[evidence_id]
                del self._timestamps[evidence_id]
                self._stats["misses"] += 1
                return None

            # 更新 LRU（移到末尾）
            self._store.move_to_end(evidence_id)
            self._stats["hits"] += 1

            return {
                "evidences": entry["evidences"],
                "tool_call_id": entry.get("tool_call_id"),
                "tool_name": entry.get("tool_name"),
                "tool_type": entry.get("tool_type"),
                "tool_result": entry.get("tool_result"),
            }

    def remove(self, evidence_id: str) -> bool:
        """
        删除证据

        Args:
            evidence_id: 证据集 ID

        Returns:
            是否成功删除
        """
        with self._lock:
            if evidence_id in self._store:
                del self._store[evidence_id]
                del self._timestamps[evidence_id]
                return True
            return False

    def clear(self) -> None:
        """清空所有证据"""
        with self._lock:
            self._store.clear()
            self._timestamps.clear()

    def cleanup_expired(self) -> int:
        """
        清理过期的证据

        Returns:
            清理的条目数
        """
        with self._lock:
            expired_keys = []
            now = time.time()

            for evidence_id, entry in self._store.items():
                if self._is_expired(entry, now):
                    expired_keys.append(evidence_id)

            for key in expired_keys:
                del self._store[key]
                del self._timestamps[key]

            return len(expired_keys)

    def get_stats(self) -> Dict:
        """获取统计信息"""
        with self._lock:
            total_requests = self._stats["hits"] + self._stats["misses"]
            hit_rate = self._stats["hits"] / total_requests if total_requests > 0 else 0

            return {
                "size": len(self._store),
                "max_size": self.max_size,
                "hits": self._stats["hits"],
                "misses": self._stats["misses"],
                "evictions": self._stats["evictions"],
                "hit_rate": hit_rate,
            }

    def _ensure_capacity(self) -> None:
        """确保缓存容量，淘汰最久未使用的条目"""
        while len(self._store) >= self.max_size:
            # 弹出最旧的条目（OrderedDict 的第一个）
            oldest_key, _ = self._store.popitem(last=False)
            del self._timestamps[oldest_key]
            self._stats["evictions"] += 1

    def _is_expired(self, entry: Dict, now: float = None) -> bool:
        """检查条目是否过期"""
        if now is None:
            now = time.time()
        created_at = entry.get("created_at", now)
        return (now - created_at) > self.ttl_seconds


# 全局单例
_global_evidence_store: Optional[EvidenceStore] = None


def get_global_evidence_store() -> EvidenceStore:
    """获取全局证据存储单例"""
    global _global_evidence_store
    if _global_evidence_store is None:
        from app.common.config import Config
        max_size = getattr(Config.features, "evidence_store_max_size", 1000)
        ttl = getattr(Config.features, "evidence_store_ttl_seconds", 3600)
        _global_evidence_store = EvidenceStore(max_size=max_size, ttl_seconds=ttl)
    return _global_evidence_store
```

**使用示例**:
```python
# 在 run_dolphin.py 中
from app.logic.agent_core_logic_v2.evidence_store import get_global_evidence_store

evidence_store = get_global_evidence_store()

# 保存证据
evidence_store.add(evidence_id, evidences)

# 在 context 中只存储 key
context_variables["evidence_store_key"] = evidence_id

# 获取证据
evidences = evidence_store.get(evidence_id)
```

### 3.3 EvidenceInjectProcessor 后处理器

**位置**: `app/logic/agent_core_logic_v2/evidence_inject_processor.py`

**功能**: 流式注入证据标注

**类设计**:
```python
import re
from typing import AsyncGenerator, Dict, List, Optional, Tuple

class EvidenceInjectProcessor:
    """证据注入后处理器"""

    def __init__(
        self,
        evidences: List[Dict],
        evidence_id: str,
        tool_call_id: str = None,
        tool_name: str = None,
        tool_type: str = None,
        tool_result: Dict = None,
        config: Optional[Dict] = None,
    ):
        """
        Args:
            evidences: 标准化的证据列表
            evidence_id: 证据集 ID
            tool_call_id: 工具调用 ID
            tool_name: 工具名称
            tool_type: 工具类型
            tool_result: 工具返回的完整结果
            config: 配置参数
        """
        self.evidences = evidences
        self.evidence_id = evidence_id
        self.tool_call_id = tool_call_id
        self.tool_name = tool_name
        self.tool_type = tool_type
        self.tool_result = tool_result
        self.config = config or {}

        # 构建匹配索引
        self._build_match_index()

        # 句子边界检测器
        from app.common.utils.sentence_boundary_detector import SentenceBoundaryDetector
        self.detector = SentenceBoundaryDetector()

        # 累积缓冲区
        self.buffer = ""
        self.accumulated = ""

    def _build_match_index(self):
        """构建匹配索引以提高性能"""
        self.exact_matches = {}
        self.alias_matches = {}

        for evidence in self.evidences:
            local_id = evidence["local_id"]
            name = evidence["object_type_name"]

            # 精确匹配
            self.exact_matches[name] = {
                "local_id": local_id,
                "evidence": evidence,
                "match_type": "exact"
            }

            # 别名匹配
            for alias in evidence.get("aliases", []):
                self.alias_matches[alias] = {
                    "local_id": local_id,
                    "evidence": evidence,
                    "match_type": "alias"
                }

    async def process(
        self,
        stream: AsyncGenerator[str, None]
    ) -> AsyncGenerator[Dict, None]:
        """
        处理流式输入，注入证据标注

        Args:
            stream: LLM 流式输出的生成器

        Yields:
            {
                "answer": "带标注的文本片段",
                "context": {...},  # 原始 context
                "_evidence": {...}  # 当前 delta 的证据元数据
            }
        """
        async for chunk in stream:
            self.buffer += chunk
            self.accumulated += chunk

            # 检测句子边界
            boundaries = self.detector.detect(self.buffer)

            if boundaries:
                # 处理到最近边界的文本
                last_boundary = boundaries[-1]
                text_to_process = self.buffer[:last_boundary]

                # 匹配并获取位置标注
                text, evidence_metadata = self._annotate_text(text_to_process)

                yield {
                    "answer": text,  # 原始文本，不修改
                    "context": {},  # 由外部填充
                    "_evidence": evidence_metadata,  # 位置索引元数据
                }

                # 更新缓冲区
                self.buffer = self.buffer[last_boundary:]

        # 处理剩余内容
        if self.buffer:
            text, evidence_metadata = self._annotate_text(self.buffer)
            yield {
                "answer": text,  # 原始文本
                "context": {},
                "_evidence": evidence_metadata,
            }

    def _annotate_text(self, text: str) -> Tuple[str, Dict]:
        """
        对文本进行标注，返回位置索引

        Returns:
            (原始文本, 位置索引元数据)
        """
        if not text:
            return text, {}

        # 执行匹配
        matches = self._match_annotations(text)

        # 按位置排序
        matches.sort(key=lambda x: (x["start"], -x["end"]))
        matches = self._resolve_overlaps(matches)

        # 构建位置索引元数据
        metadata = {}
        for match in matches:
            local_id = match["local_id"]
            evidence = next(
                (e for e in self.evidences if e["local_id"] == local_id),
                None
            )
            if evidence:
                if local_id not in metadata:
                    metadata[local_id] = {
                        "label": evidence["object_type_name"],  # 标签显示
                        "match_text": match.get("matched_text", evidence["object_type_name"]),
                        "positions": [],
                        "source": {
                            "type": self.tool_type or "unknown",
                            "tool_name": self.tool_name or "未知工具",
                            "tool_call_id": self.tool_call_id,
                            "tool_result": self.tool_result,
                        }
                    }
                # 添加位置
                metadata[local_id]["positions"].append([match["start"], match["end"]])

        return text, metadata  # 返回原始文本，不修改

    def _match_annotations(self, text: str) -> List[Dict]:
        """
        匹配文本中的实体

        Returns:
            匹配结果列表，每个元素包含：
            - local_id: 证据 ID
            - start: 起始位置
            - end: 结束位置
            - match_type: 匹配类型
            - confidence: 置信度
        """
        matches = []
        used_positions = set()

        # 1. 优先精确匹配
        for name, match_info in self.exact_matches.items():
            pos = 0
            while True:
                idx = text.find(name, pos)
                if idx == -1:
                    break

                # 检查位置是否已被使用
                end = idx + len(name)
                if not self._is_position_used(used_positions, idx, end):
                    matches.append({
                        "local_id": match_info["local_id"],
                        "start": idx,
                        "end": end,
                        "match_type": "exact",
                        "confidence": 1.0,
                    })
                    self._mark_position_used(used_positions, idx, end)

                pos = idx + 1

        # 2. 别名匹配（只匹配未标注的位置）
        for alias, match_info in self.alias_matches.items():
            pos = 0
            while True:
                idx = text.find(alias, pos)
                if idx == -1:
                    break

                end = idx + len(alias)
                if not self._is_position_used(used_positions, idx, end):
                    matches.append({
                        "local_id": match_info["local_id"],
                        "start": idx,
                        "end": end,
                        "match_type": "alias",
                        "confidence": 0.9,
                    })
                    self._mark_position_used(used_positions, idx, end)

                pos = idx + 1

        return matches

    def _is_position_used(
        self,
        used: set,
        start: int,
        end: int
    ) -> bool:
        """检查位置是否已被使用"""
        for i in range(start, end):
            if i in used:
                return True
        return False

    def _mark_position_used(self, used: set, start: int, end: int):
        """标记位置为已使用"""
        for i in range(start, end):
            used.add(i)

    def _resolve_overlaps(self, matches: List[Dict]) -> List[Dict]:
        """解决重叠匹配，保留更长/高置信度的匹配"""
        if not matches:
            return matches

        # 按置信度排序
        matches.sort(key=lambda x: x["confidence"], reverse=True)

        result = []
        used_positions = set()

        for match in matches:
            start, end = match["start"], match["end"]
            if not self._is_position_used(used_positions, start, end):
                result.append(match)
                self._mark_position_used(used_positions, start, end)

        # 按位置排序返回
        result.sort(key=lambda x: x["start"])
        return result
```

### 3.4 句子边界检测器

**位置**: `app/common/utils/sentence_boundary_detector.py`

```python
import re
from typing import List

class SentenceBoundaryDetector:
    """检测流式文本中的句子边界"""

    def __init__(self):
        # 句子结束标点
        self.sentence_endings = r'。[。！！？？\n]'

        # 引号对
        self.quote_pairs = [
            ('"', '"'),
            ('"', '"'),
            ('「', '」'),
            ('『', '』'),
        ]

    def detect(self, text: str) -> List[int]:
        """
        检测句子边界位置

        Args:
            text: 待检测文本

        Returns:
            句子结束位置的索引列表
        """
        if len(text) < 10:
            return []

        boundaries = []

        for match in re.finditer(self.sentence_endings, text):
            pos = match.end()
            if not self._is_in_quotes(text, pos):
                boundaries.append(pos)

        return boundaries

    def _is_in_quotes(self, text: str, pos: int) -> bool:
        """检查位置是否在引号内"""
        before = text[:pos]

        for open_quote, close_quote in self.quote_pairs:
            open_count = before.count(open_quote)
            close_count = before.count(close_quote)
            if open_count > close_count:
                return True

        return False
```

## 4. Agent 集成

### 4.1 在 run_dolphin.py 中集成

**位置**: `app/logic/agent_core_logic_v2/run_dolphin.py`

**集成方式**: 最小化改动，在现有流程中插入证据准备和后处理器包装

```python
async def run_dolphin(
    ac: "AgentCoreV2",
    config: AgentConfigVo,
    context_variables: Dict[str, Any],
    headers: Dict[str, str],
    is_stream: bool = True,
    ...
) -> AsyncGenerator[Dict[str, Any], None]:

    # ========== 现有代码 ==========
    # ... 初始化 agent、prompt 等代码 ...

    # ========== 新增：证据准备 ==========
    evidence_store_key = None

    # 检查是否需要启用证据注入
    if Config.features.enable_evidence_injection:
        # 检查 context 中是否有工具调用结果
        tool_results = context_variables.get("_tool_call_results")
        if tool_results:
            from app.logic.tool.evidence_prepare import evidence_prepare
            from app.logic.agent_core_logic_v2.evidence_store import get_global_evidence_store

            evidence_store = get_global_evidence_store()

            # 调用 evidence_prepare 提取证据
            for tool_name, tool_result in tool_results.items():
                prepare_result = await evidence_prepare(
                    tool_call_result=tool_result,
                    config={
                        "llm_model": ac.agent_config.model,
                        "userid": get_user_account_id(headers),
                        "enable_llm_extraction": True,
                    },
                    context=context_variables
                )

                evidence_id = prepare_result["evidence_id"]
                evidences = prepare_result["evidences"]
                tool_call_id = prepare_result.get("tool_call_id")
                tool_name = prepare_result.get("tool_name")
                tool_type = prepare_result.get("tool_type")
                tool_result = prepare_result.get("tool_result")

                # 保存到 EvidenceStore（context 只存储 key）
                evidence_store.add(
                    evidence_id,
                    evidences,
                    tool_call_id=tool_call_id,
                    tool_name=tool_name,
                    tool_type=tool_type,
                    tool_result=tool_result,
                )
                evidence_store_key = evidence_id
                context_variables["evidence_store_key"] = evidence_store_key

                # 只使用第一个工具的结果（可配置）
                break

    # ========== 构造 prompt ==========
    prompt_builder = PromptBuilder(config, temp_files)
    dolphin_prompt = await prompt_builder.build()

    # ... 后续代码 ...

    # ========== LLM 执行 ==========
    agent = DolphinAgent(...)
    await agent.init(dolphin_prompt, ...)

    # ========== 流式输出处理 ==========
    if is_stream:
        # 如果有证据，使用后处理器包装
        if evidence_store_key:
            async for output in _run_with_evidence_injection(
                agent,
                evidence_store_key,
                context_variables,
                is_debug
            ):
                yield output
        else:
            # 原有流程
            async for output in process_arun_loop(agent, is_debug):
                yield output
    else:
        # 非流式处理
        ...


async def _run_with_evidence_injection(
    agent: DolphinAgent,
    evidence_store_key: str,
    context_variables: Dict[str, Any],
    is_debug: bool,
) -> AsyncGenerator[Dict, None]:
    """
    带证据注入的流式运行包装
    """
    from app.logic.agent_core_logic_v2.evidence_inject_processor import EvidenceInjectProcessor
    from app.logic.agent_core_logic_v2.evidence_store import get_global_evidence_store

    # 从 EvidenceStore 获取证据和源信息
    evidence_store = get_global_evidence_store()
    evidence_data = evidence_store.get(evidence_store_key)
    if not evidence_data:
        # 证据不存在或已过期，回退到普通处理
        StandLogger.warn(f"Evidence not found for key: {evidence_store_key}")
        async for item in process_arun_loop(agent, is_debug):
            yield item
        return

    # 创建后处理器
    processor = EvidenceInjectProcessor(
        evidences=evidence_data["evidences"],
        evidence_id=evidence_store_key,
        tool_call_id=evidence_data.get("tool_call_id"),
        tool_name=evidence_data.get("tool_name"),
        tool_type=evidence_data.get("tool_type"),
        tool_result=evidence_data.get("tool_result"),
    )

    # 创建原始流
    async def original_stream():
        async for item in process_arun_loop(agent, is_debug):
            # 只返回 answer 部分
            yield item.get("answer", "")

    # 使用后处理器包装
    async for chunk in processor.process(original_stream()):
        # 合并原始 context
        chunk["context"] = agent.executor.context.get_all_variables_values()
        yield chunk
```

### 4.2 工具调用结果的传递

**问题**: 如何在 run_dolphin 中获取工具调用结果？

**解决方案**: 在 Agent 执行工具调用时，将结果保存到 context_variables

```python
# 在 agent_core_logic_v2 的工具调用逻辑中
async def execute_tool(tool_name: str, inputs: Dict, ...):
    result = await tool_call(tool_name, inputs, ...)

    # 保存到 context 供证据准备使用
    if "_tool_call_results" not in context_variables:
        context_variables["_tool_call_results"] = {}
    context_variables["_tool_call_results"][tool_name] = result

    return result
```

## 5. 工具注册

### 5.1 OpenAPI 定义

**文件**: `data_migrations/init/tools/openapi/evidence_tools.json`

```json
{
  "openapi": "3.1.0",
  "servers": [
    {
      "url": "http://agent-executor:30778",
      "description": "服务器地址"
    }
  ],
  "info": {
    "title": "内置工具箱-证据准备工具",
    "description": "提供证据准备功能",
    "version": "1.0.0"
  },
  "paths": {
    "/api/agent-executor/v1/tools/evidence_prepare": {
      "post": {
        "summary": "evidence_prepare",
        "description": "准备证据：从工具调用结果中提取和标准化证据",
        "operationId": "evidence_prepare_post",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "$ref": "#/components/schemas/EvidencePrepareRequest"
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Successful Response",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/EvidencePrepareResponse"
                }
              }
            }
          }
        }
      }
    }
  },
  "components": {
    "schemas": {
      "EvidencePrepareRequest": {
        "type": "object",
        "properties": {
          "tool_call_result": {
            "type": "object",
            "description": "工具调用返回的完整结果"
          },
          "config": {
            "type": "object",
            "properties": {
              "enable_llm_extraction": {
                "type": "boolean",
                "description": "是否启用 LLM 提取"
              },
              "llm_model": {
                "type": "string",
                "description": "LLM 模型"
              }
            }
          }
        },
        "required": ["tool_call_result"]
      },
      "EvidencePrepareResponse": {
        "type": "object",
        "properties": {
          "evidence_id": {
            "type": "string",
            "description": "证据集 ID"
          },
          "evidences": {
            "type": "array",
            "items": {
              "type": "object",
              "properties": {
                "local_id": {
                  "type": "string",
                  "description": "短 ID，用于 REF 标签"
                },
                "object_type_name": {
                  "type": "string",
                  "description": "实体名称"
                },
                "aliases": {
                  "type": "array",
                  "items": {"type": "string"},
                  "description": "别名列表"
                },
                "object_type_id": {
                  "type": "string",
                  "description": "对象类型 ID"
                },
                "score": {
                  "type": "number",
                  "description": "置信度"
                }
              }
            }
          },
          "summary": {
            "type": "string",
            "description": "证据摘要"
          }
        },
        "required": ["evidence_id", "evidences"]
      }
    }
  }
}
```

### 5.2 工具注册

**位置**: `app/common/tool_v2/api_tool_pkg/builtin_tools.py`

```python
from app.logic.tool.evidence_prepare import evidence_prepare

# 注册内置工具
BUILTIN_TOOLS = {
    # ... 现有工具 ...
    "evidence_prepare": {
        "function": evidence_prepare,
        "description": "从工具调用结果中提取和标准化证据",
        "parameters": {
            "type": "object",
            "properties": {
                "tool_call_result": {
                    "type": "object",
                    "description": "工具调用返回的完整结果"
                },
                "config": {
                    "type": "object",
                    "description": "配置参数"
                }
            },
            "required": ["tool_call_result"]
        }
    }
}
```

## 6. 数据格式

### 6.1 evidence_prepare 输出格式

```json
{
    "evidence_id": "ev_a1b2c3d4",
    "tool_call_id": "call_20260408_001",
    "tool_name": "知识网络搜索",
    "tool_type": "kn_search",
    "tool_result": {...},  // 完整的工具返回结果
    "evidences": [
        {
            "local_id": "e1",
            "object_type_name": "张三",
            "aliases": ["员工张三", "张工"],
            "object_type_id": "ot_employee",
            "score": 0.95
        },
        {
            "local_id": "e2",
            "object_type_name": "上海闵行区",
            "aliases": ["上海市闵行区"],
            "object_type_id": "ot_location",
            "score": 0.88
        }
    ],
    "summary": "张三、上海闵行区等 2 个证据"
}
```

### 6.2 最终响应格式

```json
{
    "answer": "我的名字是张三，我住在上海闵行区。",
    "context": {...},
    "_evidence": {
        "e1": {
            "label": "张三",
            "match_text": "张三",
            "positions": [[5, 7]],
            "source": {
                "type": "kn_search",
                "tool_name": "知识网络搜索",
                "tool_call_id": "call_20260408_001",
                "tool_result": {
                    "nodes": [
                        {
                            "object_type_name": "张三",
                            "object_type_id": "ot_employee",
                            "score": 0.95,
                            "properties": {
                                "部门": "研发部",
                                "职位": "工程师"
                            }
                        }
                    ],
                    "answer": "找到员工张三的信息"
                }
            }
        },
        "e2": {
            "label": "上海闵行区",
            "match_text": "上海闵行区",
            "positions": [[11, 16]],
            "source": {
                "type": "kn_search",
                "tool_name": "知识网络搜索",
                "tool_call_id": "call_20260408_001",
                "tool_result": {
                    "nodes": [
                        {
                            "object_type_name": "上海闵行区",
                            "object_type_id": "ot_location",
                            "score": 0.88
                        }
                    ]
                }
            }
        }
    }
}
```

**说明**：
- `answer`: 原始文本，不包含任何标注标签
- `label`: 前端标签显示的文本
- `match_text`: 实际匹配的关键字（可能与 label 不同）
- `positions`: 字符位置数组 `[start, end]`，使用 UTF-8 字符索引
- `source`: 点击标签后显示的详情
  - `type`: 工具类型
  - `tool_name`: 工具名称
  - `tool_call_id`: 工具调用 ID（用于追踪）
  - `tool_result`: 工具返回的完整结果
- 支持同一实体多次出现：`positions: [[5, 7], [20, 22]]`
- 多个证据可来自同一次工具调用（相同 `tool_call_id`）

## 7. 配置

### 7.1 ConfigMap 配置

**文件**: `decision-agent/agent-backend/deploy/helm/agent-backend/templates/configmap-executor.yaml`

```yaml
features:
  # 证据注入功能
  enable_evidence_injection: {{ .Values.agentExecutor.enableEvidenceInjection | default false }}

  # EvidenceStore 配置
  evidence_store_max_size: 1000        # 最大缓存条目数
  evidence_store_ttl_seconds: 3600     # 证据过期时间（秒），默认 1 小时

  # LLM 提取配置
  llm_extraction_enabled: true
  llm_extraction_timeout: 30

  # 句子边界检测
  min_sentence_length: 10

  # 规则匹配
  enable_alias_match: true
```

### 7.2 Values 配置

```yaml
agentExecutor:
  enableEvidenceInjection: false  # 默认关闭，灰度发布控制
```

## 8. 改动对比

### 8.1 与原方案对比

| 维度 | 中间件集成方案 | 本方案（工具+后处理器） |
|------|--------------|---------------------|
| **架构复杂度** | 高（中间件层） | 低（工具+处理器） |
| **executor 改动** | 大（新增中间件、修改流程） | 小（工具调用+包装） |
| **流式处理** | 需要修改流式框架 | 使用现有流式框架 |
| **测试复杂度** | 高（需要 mock 中间件） | 低（组件可独立测试） |

### 8.2 改动文件清单

| 文件 | 改动类型 | 代码行数 |
|------|---------|---------|
| **新增文件** |
| `app/logic/tool/evidence_prepare.py` | 新增 | ~250 行 |
| `app/logic/agent_core_logic_v2/evidence_inject_processor.py` | 新增 | ~280 行 |
| `app/logic/agent_core_logic_v2/evidence_store.py` | 新增 | ~150 行 |
| `app/common/utils/sentence_boundary_detector.py` | 新增 | ~60 行 |
| `data_migrations/init/tools/openapi/evidence_tools.json` | 新增 | ~80 行 |
| **修改文件** |
| `app/logic/agent_core_logic_v2/run_dolphin.py` | 小幅修改 | ~80 行 |
| **总计** | | **~900 行** |

相比中间件方案的 1000+ 行，改动量减少约 10%。

## 9. 测试策略

### 9.1 单元测试

**evidence_prepare 测试**:
- LLM 提取功能
- LLM 调用失败的规则回退
- 边界情况（空结果、格式错误等）
- 不同工具结果格式的处理

**EvidenceStore 测试**:
- 添加和获取证据
- LRU 淘汰机制
- 过期清理
- 并发访问安全性
- 统计信息准确性

**EvidenceInjectProcessor 测试**:
- 精确匹配
- 别名匹配
- 重叠处理
- 句子边界检测
- 流式处理

### 9.2 集成测试

- 端到端流程：工具调用 → evidence_prepare → EvidenceStore → LLM 生成 → 标注注入
- EvidenceStore 缓存和淘汰机制
- 多工具调用场景
- 流式输出验证
- 证据过期场景

### 9.3 性能测试

| 指标 | 目标值 |
|------|--------|
| 证据提取延迟 | < 3s（LLM）/ < 100ms（有_evidence） |
| 标注注入延迟 | < 10ms/句 |
| 流式额外开销 | < 5% |

## 10. 实施计划

| 阶段 | 内容 | 预计工作量 |
|------|------|-----------|
| **阶段 1** | 实现 evidence_prepare 工具 | 2 人天 |
| **阶段 2** | 实现 EvidenceStore | 1 人天 |
| **阶段 3** | 实现 EvidenceInjectProcessor | 3 人天 |
| **阶段 4** | 实现句子边界检测器 | 0.5 人天 |
| **阶段 5** | 集成到 run_dolphin.py | 2 人天 |
| **阶段 6** | 单元测试和集成测试 | 2.5 人天 |
| **阶段 7** | 配置和文档 | 0.5 人天 |
| **总计** | **11.5 人天** | |

## 11. 参考资料

- `app/logic/tool/online_search_cite_tool.py` - 参考实现
- `app/logic/agent_core_logic_v2/run_dolphin.py` - 集成点
- `app/logic/agent_core_logic_v2/interrupt_utils.py` - 流式处理参考
- `data_migrations/init/tools/openapi/search_tools.json` - 工具注册参考
