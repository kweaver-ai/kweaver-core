"""单元测试 - utils/snow_id 模块"""

import pytest
import time


class TestIdWorker:
    """测试 IdWorker 类"""

    def test_init_with_valid_ids(self):
        """测试使用有效ID初始化"""
        from app.utils.snow_id import IdWorker

        worker = IdWorker(datacenter_id=1, worker_id=1, sequence=0)

        assert worker.worker_id == 1
        assert worker.datacenter_id == 1
        assert worker.sequence == 0
        assert worker.last_timestamp == -1

    def test_init_with_zero_ids(self):
        """测试使用0 ID初始化"""
        from app.utils.snow_id import IdWorker

        worker = IdWorker(datacenter_id=0, worker_id=0, sequence=0)

        assert worker.worker_id == 0
        assert worker.datacenter_id == 0

    def test_init_with_max_ids(self):
        """测试使用最大ID初始化"""
        from app.utils.snow_id import IdWorker, MAX_WORKER_ID, MAX_DATACENTER_ID

        worker = IdWorker(
            datacenter_id=MAX_DATACENTER_ID, worker_id=MAX_WORKER_ID, sequence=0
        )

        assert worker.worker_id == MAX_WORKER_ID
        assert worker.datacenter_id == MAX_DATACENTER_ID

    def test_init_with_invalid_worker_id_raises_error(self):
        """测试使用无效worker_id抛出错误"""
        from app.utils.snow_id import IdWorker, MAX_WORKER_ID

        with pytest.raises(ValueError, match="worker_id值越界"):
            IdWorker(datacenter_id=0, worker_id=MAX_WORKER_ID + 1)

    def test_init_with_negative_worker_id_raises_error(self):
        """测试使用负worker_id抛出错误"""
        from app.utils.snow_id import IdWorker

        with pytest.raises(ValueError, match="worker_id值越界"):
            IdWorker(datacenter_id=0, worker_id=-1)

    def test_init_with_invalid_datacenter_id_raises_error(self):
        """测试使用无效datacenter_id抛出错误"""
        from app.utils.snow_id import IdWorker, MAX_DATACENTER_ID

        with pytest.raises(ValueError, match="datacenter_id值越界"):
            IdWorker(datacenter_id=MAX_DATACENTER_ID + 1, worker_id=0)

    def test_init_with_negative_datacenter_id_raises_error(self):
        """测试使用负datacenter_id抛出错误"""
        from app.utils.snow_id import IdWorker

        with pytest.raises(ValueError, match="datacenter_id值越界"):
            IdWorker(datacenter_id=-1, worker_id=0)

    def test_gen_timestamp_returns_int(self):
        """测试_gen_timestamp返回整数"""
        from app.utils.snow_id import IdWorker

        worker = IdWorker(datacenter_id=1, worker_id=1)

        timestamp = worker._gen_timestamp()

        assert isinstance(timestamp, int)
        assert timestamp > 0

    def test_get_id_returns_unique_ids(self):
        """测试get_id返回唯一ID"""
        from app.utils.snow_id import IdWorker

        worker = IdWorker(datacenter_id=1, worker_id=1)

        id1 = worker.get_id()
        id2 = worker.get_id()

        assert id1 != id2

    def test_get_id_increments_sequence(self):
        """测试get_id递增序号"""
        from app.utils.snow_id import IdWorker

        worker = IdWorker(datacenter_id=1, worker_id=1)

        # First call sets timestamp
        id1 = worker.get_id()
        # Second call should increment sequence if within same millisecond
        id2 = worker.get_id()

        # IDs should be different
        assert id1 != id2

    def test_get_id_with_clock_backward_raises_exception(self):
        """测试时钟回拨抛出异常"""
        from app.utils.snow_id import IdWorker

        worker = IdWorker(datacenter_id=1, worker_id=1)

        # Set last_timestamp to a future time
        worker.last_timestamp = int(time.time() * 1000) + 1000

        with pytest.raises(Exception):
            worker.get_id()


class TestModuleConstants:
    """测试模块常量"""

    def test_worker_id_bits(self):
        """测试WORKER_ID_BITS常量"""
        from app.utils.snow_id import WORKER_ID_BITS

        assert WORKER_ID_BITS == 5

    def test_datacenter_id_bits(self):
        """测试DATACENTER_ID_BITS常量"""
        from app.utils.snow_id import DATACENTER_ID_BITS

        assert DATACENTER_ID_BITS == 5

    def test_sequence_bits(self):
        """测试SEQUENCE_BITS常量"""
        from app.utils.snow_id import SEQUENCE_BITS

        assert SEQUENCE_BITS == 12

    def test_max_worker_id(self):
        """测试MAX_WORKER_ID常量"""
        from app.utils.snow_id import MAX_WORKER_ID

        assert MAX_WORKER_ID == 31  # 2^5 - 1

    def test_max_datacenter_id(self):
        """测试MAX_DATACENTER_ID常量"""
        from app.utils.snow_id import MAX_DATACENTER_ID

        assert MAX_DATACENTER_ID == 31  # 2^5 - 1

    def test_sequence_mask(self):
        """测试SEQUENCE_MASK常量"""
        from app.utils.snow_id import SEQUENCE_MASK

        assert SEQUENCE_MASK == 4095  # 2^12 - 1

    def test_tweepoch(self):
        """测试TWEPOCH常量"""
        from app.utils.snow_id import TWEPOCH

        assert TWEPOCH == 1288834974657


class TestSnowIdFunction:
    """测试 snow_id 函数"""

    def test_snow_id_returns_int(self):
        """测试snow_id返回整数"""
        from app.utils.snow_id import snow_id

        sid = snow_id()

        assert isinstance(sid, int)

    def test_snow_id_returns_positive(self):
        """测试snow_id返回正数"""
        from app.utils.snow_id import snow_id

        sid = snow_id()

        assert sid > 0

    def test_snow_id_returns_unique_values(self):
        """测试snow_id返回唯一值"""
        from app.utils.snow_id import snow_id
        import time

        # Add small delay to ensure different millisecond
        id1 = snow_id()
        time.sleep(0.002)  # Sleep for 2ms
        id2 = snow_id()

        assert id1 != id2

    def test_snow_id_length(self):
        """测试snow_id返回19位长度"""
        from app.utils.snow_id import snow_id

        sid = snow_id()

        # Snowflake ID should be around 19 digits
        assert len(str(sid)) >= 15
        assert len(str(sid)) <= 19


class TestModuleWorker:
    """测试模块级别的worker实例"""

    def test_worker_exists(self):
        """测试模块级别的worker实例存在"""
        from app.utils.snow_id import worker

        assert worker is not None
        assert hasattr(worker, "get_id")

    def test_worker_can_generate_id(self):
        """测试模块级别的worker可以生成ID"""
        from app.utils.snow_id import worker

        sid = worker.get_id()

        assert isinstance(sid, int)
        assert sid > 0
