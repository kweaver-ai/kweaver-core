"""
EvidenceStore 证据存储器

集中管理证据生命周期，支持 LRU 缓存、TTL 过期、线程安全。
"""

import logging
import threading
import time
from collections import OrderedDict
from typing import Any, Dict, List, Optional

logger = logging.getLogger(__name__)


class _EvidenceEntry:
    """证据条目，包含数据和元信息"""

    __slots__ = ("evidences", "created_at")

    def __init__(self, evidences: List[Dict[str, Any]]):
        self.evidences = evidences
        self.created_at = time.time()


class EvidenceStore:
    """
    证据存储器，基于 LRU 策略和 TTL 过期机制管理证据。

    特性：
    - LRU 淘汰：当存储数量超过 max_size 时，淘汰最久未访问的条目
    - TTL 过期：超过 ttl_seconds 未被访问的条目自动失效
    - 线程安全：使用 RLock 保护所有操作
    - 统计信息：记录命中率、缓存大小等指标
    """

    def __init__(self, max_size: int = 1000, ttl_seconds: int = 3600):
        """
        初始化存储器。

        Args:
            max_size: 最大缓存条目数（LRU 淘汰阈值）
            ttl_seconds: 条目过期时间（秒）
        """
        self._max_size = max_size
        self._ttl_seconds = ttl_seconds
        self._cache: OrderedDict[str, _EvidenceEntry] = OrderedDict()
        self._lock = threading.RLock()

        self._stats = {
            "hits": 0,
            "misses": 0,
            "evictions": 0,
            "expirations": 0,
            "adds": 0,
        }

    def add(self, evidence_id: str, evidences: List[Dict[str, Any]]) -> None:
        """
        添加或更新证据。

        如果 evidence_id 已存在，则更新数据并移到最近访问位置。
        如果超过最大容量，则淘汰最久未访问的条目。

        Args:
            evidence_id: 证据唯一标识符
            evidences: 证据列表
        """
        if not evidence_id or not evidences:
            return

        with self._lock:
            is_update = evidence_id in self._cache
            if is_update:
                del self._cache[evidence_id]

            evicted = None
            while len(self._cache) >= self._max_size:
                evicted = self._cache.popitem(last=False)
                self._stats["evictions"] += 1

            self._cache[evidence_id] = _EvidenceEntry(evidences)
            self._stats["adds"] += 1

            logger.debug(
                f"[EvidenceStore] add: id={evidence_id}, "
                f"count={len(evidences)}, size={len(self._cache)}"
            )

    def get(self, evidence_id: str) -> Optional[List[Dict[str, Any]]]:
        """
        获取证据。

        返回证据列表，如果不存在或已过期则返回 None。
        命中时将条目移到最近访问位置。

        Args:
            evidence_id: 证据唯一标识符

        Returns:
            证据列表，不存在或已过期返回 None
        """
        if not evidence_id:
            return None

        with self._lock:
            entry = self._cache.get(evidence_id)
            if entry is None:
                self._stats["misses"] += 1
                return None

            if self._is_expired(entry):
                del self._cache[evidence_id]
                self._stats["expirations"] += 1
                self._stats["misses"] += 1
                logger.debug(f"[EvidenceStore] expired: id={evidence_id}")
                return None

            self._cache.move_to_end(evidence_id)
            self._stats["hits"] += 1

            return entry.evidences

    def remove(self, evidence_id: str) -> bool:
        """
        删除证据。

        Args:
            evidence_id: 证据唯一标识符

        Returns:
            是否成功删除
        """
        with self._lock:
            if evidence_id in self._cache:
                del self._cache[evidence_id]
                logger.debug(f"[EvidenceStore] remove: id={evidence_id}")
                return True
            return False

    def cleanup_expired(self) -> int:
        """
        清理所有过期证据。

        Returns:
            清理的数量
        """
        with self._lock:
            now = time.time()
            expired_keys = [
                key
                for key, entry in self._cache.items()
                if now - entry.created_at > self._ttl_seconds
            ]

            for key in expired_keys:
                del self._cache[key]

            count = len(expired_keys)
            if count:
                self._stats["expirations"] += count
                logger.info(
                    f"[EvidenceStore] cleanup_expired: removed={count}, "
                    f"remaining={len(self._cache)}"
                )

            return count

    def get_stats(self) -> Dict[str, Any]:
        """
        获取统计信息。

        Returns:
            包含命中率、缓存大小等指标的字典
        """
        with self._lock:
            total_requests = self._stats["hits"] + self._stats["misses"]
            hit_rate = (
                self._stats["hits"] / total_requests
                if total_requests > 0
                else 0.0
            )

            return {
                "size": len(self._cache),
                "max_size": self._max_size,
                "ttl_seconds": self._ttl_seconds,
                "hits": self._stats["hits"],
                "misses": self._stats["misses"],
                "hit_rate": round(hit_rate, 4),
                "evictions": self._stats["evictions"],
                "expirations": self._stats["expirations"],
                "total_adds": self._stats["adds"],
            }

    def clear(self) -> None:
        """清空所有证据"""
        with self._lock:
            self._cache.clear()
            logger.debug("[EvidenceStore] cleared all entries")

    def __len__(self) -> int:
        with self._lock:
            return len(self._cache)

    def __contains__(self, evidence_id: str) -> bool:
        with self._lock:
            if evidence_id not in self._cache:
                return False
            entry = self._cache[evidence_id]
            if self._is_expired(entry):
                del self._cache[evidence_id]
                return False
            return True

    def _is_expired(self, entry: _EvidenceEntry) -> bool:
        """检查条目是否已过期"""
        return time.time() - entry.created_at > self._ttl_seconds


_global_store: Optional[EvidenceStore] = None
_global_lock = threading.Lock()


def get_global_evidence_store() -> EvidenceStore:
    """
    获取全局单例 EvidenceStore。

    Returns:
        全局 EvidenceStore 实例
    """
    global _global_store

    if _global_store is None:
        with _global_lock:
            if _global_store is None:
                from app.common.config import Config

                max_size = getattr(
                    Config.evidence, "store_max_size", 1000
                )
                ttl = getattr(
                    Config.evidence, "store_ttl_seconds", 3600
                )
                _global_store = EvidenceStore(max_size=max_size, ttl_seconds=ttl)
                logger.info(
                    f"[EvidenceStore] Global store initialized: "
                    f"max_size={max_size}, ttl={ttl}"
                )

    return _global_store


def reset_global_evidence_store() -> None:
    """重置全局单例（主要用于测试）"""
    global _global_store
    with _global_lock:
        if _global_store is not None:
            _global_store.clear()
        _global_store = None
