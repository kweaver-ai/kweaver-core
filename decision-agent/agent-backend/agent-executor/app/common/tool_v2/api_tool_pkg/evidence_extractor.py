"""
Evidence 提取模块

从工具响应中提取 nodes 字段，只保留 object_type_id、object_type_name、score，
构造 _evidence 结构。
"""

import os
from typing import Any, Dict, Optional


# nodes 中保留的字段白名单
_NODES_KEEP_FIELDS = ("object_type_id", "object_type_name", "score")


def is_evidence_extraction_enabled() -> bool:
    """检查环境变量是否启用 evidence 提取"""
    return os.getenv("ENABLE_EVIDENCE_EXTRACTION", "false").lower() in (
        "true",
        "1",
        "yes",
    )


def extract_evidence(result_data: Any) -> Optional[Dict]:
    """
    从工具响应数据中提取 _evidence 结构。

    从 result_data 中提取 nodes 字段，每个 node 只保留
    object_type_id、object_type_name、score 三个字段。

    输出格式:
    {
        "is_send_to_llm": false,
        "evidences": [{
            "module": "bkn",
            "content": {
                "object_instances": [
                    {"object_type_id": "...", "object_type_name": "...", "score": 0.3},
                    ...
                ]
            }
        }]
    }

    Args:
        result_data: 工具返回的响应数据（通常是 answer 字段的值）

    Returns:
        _evidence 结构字典，如果无法提取则返回 None
    """
    if not isinstance(result_data, dict):
        return None

    nodes = result_data.get("nodes")
    if not nodes or not isinstance(nodes, list):
        return None

    # 过滤每个 node，只保留指定字段
    filtered = [
        {k: n[k] for k in _NODES_KEEP_FIELDS if k in n}
        for n in nodes
        if isinstance(n, dict)
    ]

    if not filtered:
        return None

    return {
        "is_send_to_llm": False,
        "evidences": [{"module": "bkn", "content": {"object_instances": filtered}}],
    }
