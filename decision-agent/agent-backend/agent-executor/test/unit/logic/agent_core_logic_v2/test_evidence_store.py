# -*- coding: utf-8 -*-
"""单元测试 - evidence_store 模块"""

import time
import threading
import pytest


class TestEvidenceStoreBasic:
    """测试 EvidenceStore 基本操作"""

    def test_add_and_get(self):
        """添加和获取证据"""
        from app.logic.agent_core_logic_v2.evidence_store import (
            EvidenceStore,
        )

        store = EvidenceStore(max_size=100, ttl_seconds=3600)
        evidences = [
            {
                "local_id": "e1",
                "object_type_name": "张三",
                "aliases": [],
                "object_type_id": "ot_employee",
                "score": 0.95,
            }
        ]

        store.add("ev_test_1", evidences)
        result = store.get("ev_test_1")

        assert result is not None
        assert len(result) == 1
        assert result[0]["object_type_name"] == "张三"

    def test_get_nonexistent(self):
        """获取不存在的证据返回 None"""
        from app.logic.agent_core_logic_v2.evidence_store import (
            EvidenceStore,
        )

        store = EvidenceStore()
        assert store.get("nonexistent") is None

    def test_get_empty_key(self):
        """空 key 返回 None"""
        from app.logic.agent_core_logic_v2.evidence_store import (
            EvidenceStore,
        )

        store = EvidenceStore()
        assert store.get("") is None
        assert store.get(None) is None

    def test_remove(self):
        """删除证据"""
        from app.logic.agent_core_logic_v2.evidence_store import (
            EvidenceStore,
        )

        store = EvidenceStore()
        store.add("ev_1", [{"object_type_name": "A"}])

        assert store.remove("ev_1") is True
        assert store.get("ev_1") is None
        assert store.remove("ev_1") is False

    def test_clear(self):
        """清空所有证据"""
        from app.logic.agent_core_logic_v2.evidence_store import (
            EvidenceStore,
        )

        store = EvidenceStore()
        store.add("ev_1", [{"object_type_name": "A"}])
        store.add("ev_2", [{"object_type_name": "B"}])

        store.clear()
        assert len(store) == 0

    def test_len_and_contains(self):
        """__len__ 和 __contains__ 方法"""
        from app.logic.agent_core_logic_v2.evidence_store import (
            EvidenceStore,
        )

        store = EvidenceStore()
        assert len(store) == 0
        assert "ev_1" not in store

        store.add("ev_1", [{"object_type_name": "A"}])
        assert len(store) == 1
        assert "ev_1" in store


class TestEvidenceStoreLRU:
    """测试 LRU 淘汰机制"""

    def test_lru_eviction(self):
        """超过 max_size 时淘汰最久未访问的条目"""
        from app.logic.agent_core_logic_v2.evidence_store import (
            EvidenceStore,
        )

        store = EvidenceStore(max_size=3, ttl_seconds=3600)

        for i in range(4):
            store.add(
                f"ev_{i}",
                [{"object_type_name": f"实体{i}"}],
            )

        assert len(store) <= 3
        assert store.get("ev_0") is None
        assert store.get("ev_3") is not None

    def test_access_moves_to_end(self):
        """访问条目将其移到最近位置（不被淘汰）"""
        from app.logic.agent_core_logic_v2.evidence_store import (
            EvidenceStore,
        )

        store = EvidenceStore(max_size=3, ttl_seconds=3600)

        store.add("ev_1", [{"object_type_name": "A"}])
        store.add("ev_2", [{"object_type_name": "B"}])
        store.add("ev_3", [{"object_type_name": "C"}])

        store.get("ev_1")

        store.add("ev_4", [{"object_type_name": "D"}])

        assert store.get("ev_1") is not None
        assert store.get("ev_2") is None

    def test_update_existing(self):
        """更新已存在的条目并移到末尾"""
        from app.logic.agent_core_logic_v2.evidence_store import (
            EvidenceStore,
        )

        store = EvidenceStore(max_size=3, ttl_seconds=3600)

        store.add("ev_1", [{"object_type_name": "A"}])
        store.add("ev_2", [{"object_type_name": "B"}])

        store.add("ev_1", [{"object_type_name": "A_updated"}])

        result = store.get("ev_1")
        assert result[0]["object_type_name"] == "A_updated"


class TestEvidenceStoreTTL:
    """测试 TTL 过期机制"""

    def test_expired_entry_returns_none(self):
        """过期条目返回 None"""
        from app.logic.agent_core_logic_v2.evidence_store import (
            EvidenceStore,
        )

        store = EvidenceStore(max_size=100, ttl_seconds=1)

        store.add("ev_1", [{"object_type_name": "A"}])
        result = store.get("ev_1")
        assert result is not None

        time.sleep(1.1)
        result = store.get("ev_1")
        assert result is None

    def test_cleanup_expired(self):
        """清理过期条目"""
        from app.logic.agent_core_logic_v2.evidence_store import (
            EvidenceStore,
        )

        store = EvidenceStore(max_size=100, ttl_seconds=1)

        store.add("ev_expired", [{"object_type_name": "过期"}])
        store.add(
            "ev_fresh",
            [{"object_type_name": "新鲜", "score": 1.0}],
        )

        time.sleep(1.1)

        removed = store.cleanup_expired()
        assert removed >= 1
        assert store.get("ev_expired") is None


class TestEvidenceStoreThreadSafety:
    """测试线程安全性"""

    def test_concurrent_add(self):
        """并发添加不报错"""
        from app.logic.agent_core_logic_v2.evidence_store import (
            EvidenceStore,
        )

        store = EvidenceStore(max_size=1000, ttl_seconds=3600)

        def add_items(start_idx):
            for i in range(100):
                store.add(
                    f"ev_{start_idx}_{i}",
                    [{"object_type_name": f"实体{i}"}],
                )

        threads = [
            threading.Thread(target=add_items, args=(t * 100,))
            for t in range(5)
        ]

        for t in threads:
            t.start()
        for t in threads:
            t.join()

        assert len(store) > 0

    def test_concurrent_read_write(self):
        """并发读写不报错"""
        from app.logic.agent_core_logic_v2.evidence_store import (
            EvidenceStore,
        )

        store = EvidenceStore(max_size=500, ttl_seconds=3600)
        store.add("ev_shared", [{"object_type_name": "共享实体"}])

        errors = []

        def reader():
            try:
                for _ in range(50):
                    store.get("ev_shared")
            except Exception as e:
                errors.append(e)

        def writer():
            try:
                for i in range(50):
                    store.add(
                        f"ev_{i}",
                        [{"object_type_name": f"实体{i}"}],
                    )
            except Exception as e:
                errors.append(e)

        threads = [threading.Thread(target=reader) for _ in range(3)]
        threads += [threading.Thread(target=writer) for _ in range(2)]

        for t in threads:
            t.start()
        for t in threads:
            t.join()

        assert len(errors) == 0, f"Concurrent errors: {errors}"


class TestEvidenceStoreStats:
    """测试统计信息"""

    def test_stats_structure(self):
        """统计信息包含所有必需字段"""
        from app.logic.agent_core_logic_v2.evidence_store import (
            EvidenceStore,
        )

        store = EvidenceStore(max_size=10, ttl_seconds=3600)

        stats = store.get_stats()

        expected_keys = {
            "size",
            "max_size",
            "ttl_seconds",
            "hits",
            "misses",
            "hit_rate",
            "evictions",
            "expirations",
            "total_adds",
        }
        assert set(stats.keys()) == expected_keys

    def test_hit_rate_calculation(self):
        """命中率计算正确"""
        from app.logic.agent_core_logic_v2.evidence_store import (
            EvidenceStore,
        )

        store = EvidenceStore(ttl_seconds=3600)

        store.add("ev_1", [{"object_type_name": "A"}])
        store.get("ev_1")
        store.get("ev_nonexistent")

        stats = store.get_stats()
        assert stats["hits"] == 1
        assert stats["misses"] == 1
        assert stats["hit_rate"] == 0.5

    def test_eviction_counted(self):
        """淘汰次数被记录"""
        from app.logic.agent_core_logic_v2.evidence_store import (
            EvidenceStore,
        )

        store = EvidenceStore(max_size=2, ttl_seconds=3600)

        store.add("ev_1", [{}])
        store.add("ev_2", [{}])
        store.add("ev_3", [{}])

        stats = store.get_stats()
        assert stats["evictions"] >= 1


class TestGlobalEvidenceStore:
    """测试全局单例"""

    def test_singleton_pattern(self):
        """多次调用返回同一实例"""
        from app.logic.agent_core_logic_v2.evidence_store import (
            EvidenceStore,
            reset_global_evidence_store,
            get_global_evidence_store,
        )
        from unittest import mock

        reset_global_evidence_store()

        with mock.patch(
            "app.common.config.Config"
        ) as mock_config:
            mock_config.features.evidence_store_max_size = 100
            mock_config.features.evidence_store_ttl_seconds = 3600

            s1 = get_global_evidence_store()
            s2 = get_global_evidence_store()

            assert s1 is s2

        reset_global_evidence_store()

    def test_reset_clears_all(self):
        """重置清空所有数据"""
        from app.logic.agent_core_logic_v2.evidence_store import (
            reset_global_evidence_store,
            get_global_evidence_store,
        )
        from unittest import mock

        with mock.patch(
            "app.common.config.Config"
        ) as mock_config:
            mock_config.features.evidence_store_max_size = 100
            mock_config.features.evidence_store_ttl_seconds = 3600

            store = get_global_evidence_store()
            store.add("test", [{}])
            assert len(store) > 0

            reset_global_evidence_store()

            new_store = get_global_evidence_store()
            assert len(new_store) == 0

        reset_global_evidence_store()


class TestEdgeCases:
    """边界情况测试"""

    def test_add_empty_evidences(self):
        """添加空证据列表不做任何操作"""
        from app.logic.agent_core_logic_v2.evidence_store import (
            EvidenceStore,
        )

        store = EvidenceStore()
        store.add("ev_empty", [])
        assert len(store) == 0

    def test_add_empty_key(self):
        """空 key 不添加"""
        from app.logic.agent_core_logic_v2.evidence_store import (
            EvidenceStore,
        )

        store = EvidenceStore()
        store.add("", [{"object_type_name": "A"}])
        store.add(None, [{"object_type_name": "B"}])
        assert len(store) == 0

    def test_contains_with_expired(self):
        """__contains__ 对过期条目返回 False"""
        from app.logic.agent_core_logic_v2.evidence_store import (
            EvidenceStore,
        )

        store = EvidenceStore(max_size=100, ttl_seconds=1)
        store.add("ev_1", [{"object_type_name": "A"}])

        assert "ev_1" in store

        time.sleep(1.1)
        assert "ev_1" not in store
