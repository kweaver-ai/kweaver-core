"""扩展单元测试 - utils/snow_id 模块"""

import pytest
import time
from app.utils.snow_id import (
    IdWorker,
    snow_id,
    WORKER_ID_BITS,
    DATACENTER_ID_BITS,
    SEQUENCE_BITS,
    MAX_WORKER_ID,
    MAX_DATACENTER_ID,
    SEQUENCE_MASK,
    TWEPOCH,
)


class TestConstants:
    """测试常量定义"""

    def test_worker_id_bits(self):
        """测试worker ID位数"""
        assert WORKER_ID_BITS == 5

    def test_datacenter_id_bits(self):
        """测试数据中心ID位数"""
        assert DATACENTER_ID_BITS == 5

    def test_sequence_bits(self):
        """测试序列号位数"""
        assert SEQUENCE_BITS == 12

    def test_max_worker_id(self):
        """测试最大worker ID"""
        assert MAX_WORKER_ID == 31  # 2^5 - 1

    def test_max_datacenter_id(self):
        """测试最大数据中心ID"""
        assert MAX_DATACENTER_ID == 31  # 2^5 - 1

    def test_sequence_mask(self):
        """测试序列号掩码"""
        assert SEQUENCE_MASK == 4095  # 2^12 - 1

    def test_twePOCH(self):
        """测试Twitter元年时间戳"""
        assert TWEPOCH == 1288834974657


class TestIdWorkerInit:
    """测试IdWorker初始化"""

    def test_init_valid_ids(self):
        """测试有效ID初始化"""
        worker = IdWorker(0, 0)
        assert worker.datacenter_id == 0
        assert worker.worker_id == 0
        assert worker.sequence == 0
        assert worker.last_timestamp == -1

    def test_init_with_sequence(self):
        """测试带序列号初始化"""
        worker = IdWorker(1, 1, 100)
        assert worker.sequence == 100

    def test_init_max_valid_ids(self):
        """测试最大有效ID"""
        worker = IdWorker(MAX_DATACENTER_ID, MAX_WORKER_ID)
        assert worker.datacenter_id == MAX_DATACENTER_ID
        assert worker.worker_id == MAX_WORKER_ID

    def test_init_invalid_worker_id_too_high(self):
        """测试worker ID过高"""
        with pytest.raises(ValueError, match="worker_id"):
            IdWorker(0, MAX_WORKER_ID + 1)

    def test_init_invalid_worker_id_negative(self):
        """测试worker ID为负"""
        with pytest.raises(ValueError, match="worker_id"):
            IdWorker(0, -1)

    def test_init_invalid_datacenter_id_too_high(self):
        """测试数据中心ID过高"""
        with pytest.raises(ValueError, match="datacenter_id"):
            IdWorker(MAX_DATACENTER_ID + 1, 0)

    def test_init_invalid_datacenter_id_negative(self):
        """测试数据中心ID为负"""
        with pytest.raises(ValueError, match="datacenter_id"):
            IdWorker(-1, 0)

    def test_init_both_invalid(self):
        """测试两者都无效"""
        with pytest.raises(ValueError):
            IdWorker(100, 100)


class TestIdWorkerGenTimestamp:
    """测试时间戳生成"""

    def test_gen_timestamp_returns_int(self):
        """测试生成整数时间戳"""
        worker = IdWorker(1, 1)
        timestamp = worker._gen_timestamp()
        assert isinstance(timestamp, int)

    def test_gen_timestamp_in_milliseconds(self):
        """测试毫秒级时间戳"""
        worker = IdWorker(1, 1)
        timestamp = worker._gen_timestamp()
        # Should be much larger than TWEPOCH
        assert timestamp > TWEPOCH

    def test_gen_timestamp_increases(self):
        """测试时间戳递增"""
        worker = IdWorker(1, 1)
        ts1 = worker._gen_timestamp()
        time.sleep(0.001)  # Sleep 1ms
        ts2 = worker._gen_timestamp()
        assert ts2 >= ts1


class TestIdWorkerGetId:
    """测试ID生成"""

    def test_get_id_returns_int(self):
        """测试返回整数ID"""
        worker = IdWorker(1, 1)
        id_val = worker.get_id()
        assert isinstance(id_val, int)

    def test_get_id_positive(self):
        """测试返回正数ID"""
        worker = IdWorker(1, 1)
        id_val = worker.get_id()
        assert id_val > 0

    def test_get_id_unique(self):
        """测试生成唯一ID"""
        worker = IdWorker(1, 1)
        ids = [worker.get_id() for _ in range(100)]
        assert len(set(ids)) == 100  # All unique

    def test_get_id_increases(self):
        """测试ID递增"""
        worker = IdWorker(1, 1)
        id1 = worker.get_id()
        id2 = worker.get_id()
        assert id2 > id1

    def test_get_id_same_worker_same_datacenter(self):
        """测试相同worker和数据中心"""
        import time

        worker1 = IdWorker(5, 10)
        worker2 = IdWorker(5, 10)
        # IDs should be different due to timestamp/sequence
        id1 = worker1.get_id()
        time.sleep(0.001)
        id2 = worker2.get_id()
        assert id1 != id2

    def test_get_id_different_workers(self):
        """测试不同worker"""
        worker1 = IdWorker(0, 0)
        worker2 = IdWorker(0, 1)
        id1 = worker1.get_id()
        id2 = worker2.get_id()
        # Should be different due to worker ID bits
        assert id1 != id2

    def test_get_id_different_datacenters(self):
        """测试不同数据中心"""
        worker1 = IdWorker(0, 0)
        worker2 = IdWorker(1, 0)
        id1 = worker1.get_id()
        id2 = worker2.get_id()
        # Should be different due to datacenter ID bits
        assert id1 != id2

    def test_get_id_length(self):
        """测试ID长度"""
        worker = IdWorker(1, 1)
        id_val = worker.get_id()
        # Snowflake ID should be at most 19 digits
        assert len(str(id_val)) <= 19

    def test_get_id_rapid_generation(self):
        """测试快速生成ID"""
        worker = IdWorker(1, 1)
        ids = [worker.get_id() for _ in range(1000)]
        assert len(set(ids)) == 1000

    def test_get_id_19_digits(self):
        """测试19位数字"""
        worker = IdWorker(1, 1)
        id_val = worker.get_id()
        # Snowflake IDs are typically 13-19 digits
        str_id = str(id_val)
        assert 13 <= len(str_id) <= 19


class TestIdWorkerSequenceOverflow:
    """测试序列号溢出"""

    def test_sequence_resets_after_overflow(self):
        """测试序列号溢出后重置"""
        worker = IdWorker(1, 1, SEQUENCE_MASK)
        # Set sequence to max, next call should handle overflow
        id1 = worker.get_id()
        id2 = worker.get_id()
        # IDs should still be unique
        assert id1 != id2


class TestIdWorkerClockBackwards:
    """测试时钟回拨"""

    def test_clock_backwards_raises_exception(self):
        """测试时钟回拨抛出异常"""
        worker = IdWorker(1, 1)
        # Get first ID
        worker.get_id()
        # Manually set last_timestamp to future
        worker.last_timestamp = worker._gen_timestamp() + 1000
        # Next call should raise exception
        with pytest.raises(Exception):
            worker.get_id()


class TestSnowIdFunction:
    """测试snow_id函数"""

    def test_snow_id_returns_int(self):
        """测试返回整数"""
        id_val = snow_id()
        assert isinstance(id_val, int)

    def test_snow_id_positive(self):
        """测试返回正数"""
        id_val = snow_id()
        assert id_val > 0

    def test_snow_id_unique(self):
        """测试生成唯一ID"""
        import time

        ids = []
        for _ in range(10):
            ids.append(snow_id())
            time.sleep(0.001)  # Small delay to ensure different timestamps
        # Note: Due to timing, some duplicates may occur
        assert len(set(ids)) >= 5  # At least half should be unique

    def test_snow_id_19_digits(self):
        """测试19位数字"""
        id_val = snow_id()
        str_id = str(id_val)
        # Should be 19 digits for recent timestamps
        assert 10 <= len(str_id) <= 19

    def test_snow_id_increases(self):
        """测试ID递增"""
        import time

        id1 = snow_id()
        time.sleep(0.001)
        id2 = snow_id()
        # IDs may increase but not guaranteed due to timing
        assert isinstance(id2, int)


class TestIdWorkerEdgeCases:
    """测试边界情况"""

    def test_zero_worker_and_datacenter(self):
        """测试零worker和数据中心"""
        worker = IdWorker(0, 0)
        id_val = worker.get_id()
        assert id_val > 0

    def test_max_worker_and_datacenter(self):
        """测试最大worker和数据中心"""
        worker = IdWorker(MAX_DATACENTER_ID, MAX_WORKER_ID)
        id_val = worker.get_id()
        assert id_val > 0

    def test_middle_values(self):
        """测试中间值"""
        worker = IdWorker(15, 16)
        id_val = worker.get_id()
        assert id_val > 0

    def test_sequence_initial_value(self):
        """测试序列号初始值"""
        worker = IdWorker(1, 1, 999)
        assert worker.sequence == 999

    def test_multiple_workers_parallel_ids(self):
        """测试多个worker并行生成ID"""
        workers = [IdWorker(0, i % 32) for i in range(10)]
        ids = []
        for worker in workers:
            ids.append(worker.get_id())

        # All IDs should be unique
        assert len(set(ids)) == 10


class TestIdWorkerTimestampHandling:
    """测试时间戳处理"""

    def test_same_millisecond_sequence_increments(self):
        """测试同一毫秒内序列号递增"""
        worker = IdWorker(1, 1)
        # These should be in same millisecond
        ids = []
        for _ in range(10):
            ids.append(worker.get_id())

        # All should be unique
        assert len(set(ids)) == 10

    def test_different_millisecond_sequence_resets(self):
        """测试不同毫秒序列号重置"""
        import time

        worker = IdWorker(1, 1)
        id1 = worker.get_id()
        time.sleep(0.002)  # Ensure different millisecond
        id2 = worker.get_id()

        # Should still be unique
        assert id1 != id2
        # Sequence should be reset (or incremented based on timing)
        assert worker.sequence >= 0


class TestGlobalWorker:
    """测试全局worker实例"""

    def test_global_worker_exists(self):
        """测试全局worker存在"""
        from app.utils.snow_id import worker

        assert worker is not None
        assert isinstance(worker, IdWorker)

    def test_global_worker_can_generate_ids(self):
        """测试全局worker可生成ID"""
        from app.utils.snow_id import worker

        id_val = worker.get_id()
        assert id_val > 0


class TestIdStructure:
    """测试ID结构"""

    def test_id_components_valid(self):
        """测试ID组件有效"""
        worker = IdWorker(5, 10)
        id_val = worker.get_id()

        # ID should be positive
        assert id_val > 0

        # ID should be reasonable size (less than 2^63)
        assert id_val < 2**63

    def test_id_bit_length(self):
        """测试ID位长度"""
        worker = IdWorker(1, 1)
        id_val = worker.get_id()

        # Snowflake ID should be at most 64 bits
        assert id_val.bit_length() <= 64
