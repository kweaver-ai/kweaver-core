# 证据注入功能实现计划

**日期**: 2026-04-08
**来源**: `kweaver/docs/superpowers/specs/2026-04-08-evidence-injection-tools-design.md`
**版本**: v4.5 (EvidenceStore 管理模式，统一 LLM 提取，无 Prompt 注入，包含 Source 信息)
**状态**: 待实施

## 1. 概述

本计划实现基于**内置工具+后处理器+EvidenceStore**模式的证据注入功能，支持流式输出，最小化对 `agent-executor` 核心代码的改动。

### 1.1 核心目标

1. **evidence_prepare 工具**: 从工具调用结果中提取和标准化证据
2. **EvidenceStore**: 集中管理证据生命周期，LRU 缓存模式
3. **EvidenceInjectProcessor 后处理器**: 流式注入证据标注，使用位置索引格式
4. **最小化改动**: context 只存储 `evidence_store_key`，不存储完整 evidences

### 1.2 架构概览

```
工具调用 → evidence_prepare 提取证据 → EvidenceStore.add(evidence_id, evidences)
                                        ↓
context["evidence_store_key"] = evidence_id
                                        ↓
LLM 生成 (流式) → EvidenceStore.get(evidence_id) → EvidenceInjectProcessor 包装 → 输出 + _evidence 元数据
```

### 1.3 设计优势

| 方面 | 原设计 (直接保存到 context) | 新设计 (EvidenceStore) |
|------|---------------------------|----------------------|
| **Context 污染** | evidences 完整数据污染 context | 只存储 key |
| **可扩展性** | 难以添加过期、LRU 等功能 | 支持复杂管理功能 |
| **内存管理** | 无限制增长 | LRU 淘汰 + TTL 过期 |
| **线程安全** | 不保证 | RLock 保护 |

## 2. 实现任务分解

### 阶段 1: 核心工具实现 (2 人天)

#### 任务 1.1: 实现 evidence_prepare 工具
**文件**: `app/logic/tool/evidence_prepare.py` (新增)

**功能**:
- 统一使用 LLM 提取实体（不区分是否有 _evidence）
- LLM 提取失败时使用规则回退
- 标准化输出格式

**关键接口**:
```python
async def evidence_prepare(
    tool_call_result: Dict,
    config: Optional[Dict] = None,
    context: Optional[Dict] = None,
) -> Dict:
    """
    Returns:
        {
            "evidence_id": "ev_uuid",
            "evidences": [
                {
                    "local_id": "e1",
                    "object_type_name": "张三",
                    "aliases": ["员工张三"],
                    "object_type_id": "ot_employee",
                    "score": 0.95
                }
            ],
            "summary": "张三等 1 个证据"
        }
    """
```

**子任务**:
1. 实现 `_process_with_evidence()` - 处理已有 _evidence
2. 实现 `_extract_by_llm()` - 使用 LLM 提取实体
3. 实现 `_parse_llm_result()` - 解析 LLM 返回的 JSON
4. 实现 `_fallback_extraction()` - 规则提取回退策略
5. 实现 `_generate_summary()` - 生成证据摘要

#### 任务 1.2: 创建工具 OpenAPI 定义
**文件**: `data_migrations/init/tools/openapi/evidence_tools.json` (新增)

**内容**:
- 定义 `evidence_prepare` 工具的 OpenAPI 3.1.0 规范
- 请求参数: `tool_call_result`, `config`
- 响应格式: 标准 evidence 结构

#### 任务 1.3: 注册内置工具
**文件**: `app/logic/tool/__init__.py` (修改)

**改动**: 导出 `evidence_prepare` 函数供工具注册使用

---

### 阶段 2: EvidenceStore 实现 (1 人天)

#### 任务 2.1: 实现 EvidenceStore 核心逻辑
**文件**: `app/logic/agent_core_logic_v2/evidence_store.py` (新增)

**功能**:
- LRU 缓存管理（基于 OrderedDict）
- TTL 过期机制
- 线程安全保护
- 统计信息收集

**关键接口**:
```python
class EvidenceStore:
    def __init__(self, max_size: int = 1000, ttl_seconds: int = 3600):
        """初始化存储器"""

    def add(self, evidence_id: str, evidences: List[Dict]) -> None:
        """添加证据"""

    def get(self, evidence_id: str) -> Optional[List[Dict]]:
        """获取证据，返回 None 表示不存在或已过期"""

    def remove(self, evidence_id: str) -> bool:
        """删除证据"""

    def cleanup_expired(self) -> int:
        """清理过期证据，返回清理数量"""

    def get_stats(self) -> Dict:
        """获取统计信息（命中率、缓存大小等）"""


def get_global_evidence_store() -> EvidenceStore:
    """获取全局单例"""
```

**子任务**:
1. 实现 LRU 淘汰逻辑
2. 实现 TTL 过期检查
3. 实现线程安全保护
4. 实现统计信息收集
5. 编写单元测试

#### 任务 2.2: 配置支持
**文件**: `app/config/config_v2/models/app_config.py` (修改)

**新增配置**:
```python
@dataclass
class FeatureConfig:
    # EvidenceStore 配置
    evidence_store_max_size: int = 1000
    evidence_store_ttl_seconds: int = 3600
```

---

### 阶段 3: 后处理器实现 (3 人天)

#### 任务 2.1: 实现句子边界检测器
**文件**: `app/common/utils/sentence_boundary_detector.py` (新增)

**功能**:
- 检测句子边界位置
- 处理引号内的句子结束符
- 支持中英文标点

**关键接口**:
```python
class SentenceBoundaryDetector:
    def detect(self, text: str) -> List[int]:
        """返回句子结束位置的索引列表"""
```

#### 任务 2.2: 实现 EvidenceInjectProcessor
**文件**: `app/logic/agent_core_logic_v2/evidence_inject_processor.py` (新增)

**功能**:
- 包装 LLM 流式输出
- 累积文本到句子边界
- 规则匹配实体名称
- 生成位置索引元数据

**关键接口**:
```python
class EvidenceInjectProcessor:
    async def process(
        self,
        stream: AsyncGenerator[str, None]
    ) -> AsyncGenerator[Dict, None]:
        """
        Yields:
            {
                "answer": "原始文本片段",
                "context": {...},
                "_evidence": {
                    "e1": {
                        "object_type_name": "张三",
                        "positions": [[4, 6]]
                    }
                }
            }
        """

    def _annotate_text(self, text: str) -> Tuple[str, Dict]:
        """返回 (原始文本, 位置索引元数据)"""
```

**子任务**:
1. `_build_match_index()` - 构建精确匹配和别名匹配索引
2. `_match_annotations()` - 匹配文本中的实体
3. `_resolve_overlaps()` - 解决重叠匹配
4. `_is_position_used()` / `_mark_position_used()` - 位置标记管理

---

### 阶段 4: agent-executor 集成 (2 人天)

#### 任务 4.1: 修改 run_dolphin.py
**文件**: `app/logic/agent_core_logic_v2/run_dolphin.py` (修改)

**集成点**:

1. **证据准备阶段** (在 LLM 执行前):
```python
# 在 prompt 构建之后，LLM 执行之前
evidence_store_key = None

if Config.features.enable_evidence_injection:
    evidence_store_key = await _prepare_evidence(context_variables, headers)
    # 无需注入到 prompt，EvidenceInjectProcessor 会通过规则匹配标注
```

2. **流式输出包装** (在 `process_arun_loop` 调用处):
```python
if is_stream and evidence_store_key:
    async for output in _run_with_evidence_injection(
        agent, evidence_store_key, context_variables, is_debug
    ):
        yield output
else:
    async for output in process_arun_loop(agent, is_debug):
        yield output
```

**关键变化**:
- context 只存储 `evidence_store_key`，不存储完整 evidences
- `_run_with_evidence_injection` 接收 `evidence_store_key`，内部从 EvidenceStore 获取 evidences
- 需要处理证据不存在/已过期的情况

**新增函数**:
- `_prepare_evidence()` - 检查并调用 evidence_prepare，返回 evidence_store_key
- `_run_with_evidence_injection()` - 包装流式输出（接收 evidence_store_key）

#### 任务 4.2: 工具调用结果传递机制
**文件**: `app/logic/agent_core_logic_v2/run_dolphin.py` (修改)

**方案**: 在 Agent 执行工具调用时，将结果保存到 context_variables

```python
# 在工具调用逻辑中
context_variables["_tool_call_results"] = context_variables.get("_tool_call_results", {})
context_variables["_tool_call_results"][tool_name] = result
```

**替代方案**: 从 `agent.executor.context` 中获取最近的工具调用结果

---

### 阶段 5: 配置和测试 (3 人天)

#### 任务 4.1: 配置支持
**文件**: `app/config/config_v2/models/app_config.py` (修改)

**新增配置**:
```python
@dataclass
class FeatureConfig:
    # 证据注入功能开关
    enable_evidence_injection: bool = False

    # LLM 提取配置
    llm_extraction_timeout: int = 30
    llm_extraction_model: str = ""

    # 规则匹配配置
    enable_alias_match: bool = True
    min_sentence_length: int = 10
```

#### 任务 5.2: 单元测试
**文件**: `tests/logic/tool/test_evidence_prepare.py` (新增)

**测试用例**:
1. LLM 提取成功场景
2. LLM 调用失败的规则回退
3. 空结果和格式错误处理
4. 不同工具结果格式的处理

**文件**: `tests/logic/agent_core_logic_v2/test_evidence_store.py` (新增)

**测试用例**:
1. 添加和获取证据
2. LRU 淘汰机制
3. TTL 过期清理
4. 并发访问安全性
5. 统计信息准确性

**文件**: `tests/logic/agent_core_logic_v2/test_evidence_inject_processor.py` (新增)

**测试用例**:
1. 精确匹配
2. 别名匹配
3. 重叠处理
4. 句子边界检测
5. 流式处理模拟

#### 任务 5.3: 集成测试
**文件**: `tests/integration/test_evidence_injection_e2e.py` (新增)

**测试场景**:
1. 完整流程: kn_search → evidence_prepare → EvidenceStore → LLM → 标注注入
2. EvidenceStore 缓存和淘汰机制
3. 多工具调用场景
4. 流式输出验证
5. 证据过期场景

---

## 3. 数据流设计

### 3.1 evidence_prepare 调用流程

```
┌─────────────────────────────────────────────────────────────────┐
│ 1. Agent 调用 kn_search (或其他业务工具)                          │
│    └─ 返回: {"nodes": [...], "answer": "..."}                    │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 2. run_dolphin 检测到工具调用结果                                  │
│    └─ 调用 evidence_prepare(tool_call_result)                   │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 3. evidence_prepare 处理                                         │
│    ├─ 统一使用 LLM 提取关键实体                                  │
│    └─ LLM 失败 → 规则回退                                       │
│    └─ 返回: {evidence_id, evidences, summary}                   │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 4. 保存到 EvidenceStore                                          │
│    evidence_store.add(evidence_id, evidences)                   │
│    └─ context_variables["evidence_store_key"] = evidence_id     │
│       (只存储 key，不存储完整 evidences)                         │
└─────────────────────────────────────────────────────────────────┘
```

### 3.2 LLM 生成和后处理器流程

```
┌─────────────────────────────────────────────────────────────────┐
│ 5. LLM 开始流式生成                                              │
│    agent.arun() → "我的名字是张三，我住上海..."                  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 6. EvidenceInjectProcessor 从 EvidenceStore 获取证据            │
│    evidences = evidence_store.get(evidence_id)                  │
│    ├─ 接收流式 chunk                                           │
│    ├─ 累积到句子边界                                            │
│    ├─ 规则匹配 "张三" → e1, "上海" → e2                         │
│    └─ 返回位置索引: {e1: [[4,6]], e2: [[9,11]]}                │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 7. 最终输出                                                      │
│    {                                                           │
│      "answer": "我的名字是张三，我住上海。",                      │
│      "_evidence": {                                             │
│        "e1": {"object_type_name": "张三", "positions": [[4,6]]},│
│        "e2": {"object_type_name": "上海", "positions": [[9,11]]}│
│      }                                                          │
│    }                                                           │
└─────────────────────────────────────────────────────────────────┘
```

## 4. 关键设计决策

### 4.1 为什么使用位置索引而非内联标签？

| 方案 | 优点 | 缺点 |
|------|------|------|
| **位置索引** (选择) | • 原始文本纯净<br>• 易于前端渲染<br>• 支持多语言 | • 需要额外元数据 |
| 内联标签 `<i index="">` | • 自包含 | • 污染原始文本<br>• 不利于复制<br>• 与现有冲突 |

### 4.2 为什么选择工具+后处理器模式？

| 方案 | 代码改动 | 复杂度 | 流式支持 |
|------|---------|--------|---------|
| **工具+后处理器** (选择) | ~900 行 | 低 | 原生 |
| 中间件集成 | ~1200 行 | 高 | 需修改 |

### 4.3 为什么使用 EvidenceStore？

| 方案 | 优点 | 缺点 |
|------|------|------|
| **EvidenceStore** (选择) | • Context 不被污染<br>• LRU 淘汰机制<br>• TTL 过期支持<br>• 线程安全 | • 额外的抽象层 |
| 直接保存到 context | • 简单直接 | • Context 膨胀<br>• 无内存管理<br>• 无过期机制 |

### 4.4 如何获取工具调用结果？

**方案 A** (推荐): 在工具调用时保存到 context
- 优点: 可靠，不需要修改 Agent 核心逻辑
- 缺点: 需要在工具调用处添加代码

**方案 B**: 从 `agent.executor.context` 中读取
- 优点: 无需额外保存
- 缺点: 需要了解 SDK 内部实现

**方案 C**: 通过 `process_arun_loop` 的输出检测
- 优点: 完全解耦
- 缺点: 需要解析输出，不够可靠

## 5. 改动文件清单

| 文件 | 类型 | 代码行数 |
|------|------|---------|
| **新增文件** |
| `app/logic/tool/evidence_prepare.py` | 新增 | ~250 行 |
| `app/logic/agent_core_logic_v2/evidence_inject_processor.py` | 新增 | ~280 行 |
| `app/logic/agent_core_logic_v2/evidence_store.py` | 新增 | ~150 行 |
| `app/common/utils/sentence_boundary_detector.py` | 新增 | ~60 行 |
| `data_migrations/init/tools/openapi/evidence_tools.json` | 新增 | ~80 行 |
| `tests/logic/tool/test_evidence_prepare.py` | 新增 | ~150 行 |
| `tests/logic/agent_core_logic_v2/test_evidence_store.py` | 新增 | ~120 行 |
| `tests/logic/agent_core_logic_v2/test_evidence_inject_processor.py` | 新增 | ~200 行 |
| `tests/integration/test_evidence_injection_e2e.py` | 新增 | ~100 行 |
| **修改文件** |
| `app/logic/agent_core_logic_v2/run_dolphin.py` | 修改 | ~80 行 |
| `app/logic/tool/__init__.py` | 修改 | ~5 行 |
| `app/config/config_v2/models/app_config.py` | 修改 | ~20 行 |
| **总计** | | **~1495 行** |

## 6. 风险和缓解措施

### 风险 1: LLM 提取实体不准确
**缓解**: 
- 提供规则回退策略
- 支持手动配置别名
- 置信度阈值过滤

### 风险 2: 流式处理延迟增加
**缓解**:
- 异步处理，不阻塞主流程
- 句子边界缓存优化
- 可配置的开关

### 风险 3: 与现有代码冲突
**缓解**:
- 使用 feature flag 控制
- 充分的单元测试
- 灰度发布

## 7. 验收标准

### 功能验收
1. `evidence_prepare` 能正确提取工具结果中的证据
2. LLM 生成的内容能正确匹配证据名称
3. 位置索引准确标注证据在文本中的位置
4. 流式输出不受影响

### 性能验收
1. 证据提取延迟 < 3s (LLM) / < 100ms (已有 _evidence)
2. 标注注入延迟 < 10ms/句
3. 流式额外开销 < 5%

### 质量验收
1. 单元测试覆盖率 >= 80%
2. 集成测试通过
3. 代码审查通过

## 8. 实施顺序

1. **第 1-2 天**: 实现 evidence_prepare 工具 (阶段 1)
2. **第 3 天**: 实现 EvidenceStore (阶段 2)
3. **第 4-6 天**: 实现 EvidenceInjectProcessor (阶段 3)
4. **第 7-8 天**: agent-executor 集成 (阶段 4)
5. **第 9-11 天**: 配置和测试 (阶段 5)

## 9. 参考文件

- Spec: `kweaver/docs/superpowers/specs/2026-04-08-evidence-injection-tools-design.md`
- 现有证据提取: `app/common/tool_v2/api_tool_pkg/evidence_extractor.py`
- 流式处理: `app/logic/agent_core_logic_v2/interrupt_utils.py`
- LLM 调用: `app/driven/dip/model_api_service.py`
- 参考实现: `app/logic/tool/online_search_cite_tool.py`
