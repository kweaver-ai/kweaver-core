"""测试 snow_id 模块"""
import pytest
from app.commons.snow_id import IdWorker, snow_id, worker, SEQUENCE_MASK


class TestIdWorker:
    """测试IdWorker类"""

    def test_init_valid_params(self):
        """测试有效参数初始化"""
        id_worker = IdWorker(datacenter_id=1, worker_id=1, sequence=0)
        assert id_worker.datacenter_id == 1
        assert id_worker.worker_id == 1
        assert id_worker.sequence == 0

    def test_init_invalid_worker_id(self):
        """测试无效的worker_id"""
        with pytest.raises(ValueError, match="worker_id值越界"):
            IdWorker(datacenter_id=1, worker_id=32, sequence=0)
        
        with pytest.raises(ValueError, match="worker_id值越界"):
            IdWorker(datacenter_id=1, worker_id=-1, sequence=0)

    def test_init_invalid_datacenter_id(self):
        """测试无效的datacenter_id"""
        with pytest.raises(ValueError, match="datacenter_id值越界"):
            IdWorker(datacenter_id=32, worker_id=1, sequence=0)
        
        with pytest.raises(ValueError, match="datacenter_id值越界"):
            IdWorker(datacenter_id=-1, worker_id=1, sequence=0)

    def test_gen_timestamp(self):
        """测试生成时间戳"""
        id_worker = IdWorker(datacenter_id=1, worker_id=1)
        timestamp = id_worker._gen_timestamp()
        assert isinstance(timestamp, int)
        assert timestamp > 0

    def test_get_id_single(self):
        """测试生成单个ID"""
        id_worker = IdWorker(datacenter_id=1, worker_id=1)
        id1 = id_worker.get_id()
        assert isinstance(id1, int)
        assert id1 > 0

    def test_get_id_unique(self):
        """测试生成的ID唯一性"""
        id_worker = IdWorker(datacenter_id=1, worker_id=1)
        ids = set()
        for _ in range(1000):
            ids.add(id_worker.get_id())
        assert len(ids) == 1000

    def test_get_id_sequence_increment(self):
        """测试同一毫秒内序列号递增"""
        id_worker = IdWorker(datacenter_id=1, worker_id=1)
        # 在同一时间戳内生成多个ID
        id1 = id_worker.get_id()
        id2 = id_worker.get_id()
        # ID应该递增
        assert id2 > id1

    def test_get_id_sequence_reset(self):
        """测试序列号重置"""
        id_worker = IdWorker(datacenter_id=1, worker_id=1)
        # 生成足够多的ID以触发序列号重置
        for _ in range(SEQUENCE_MASK + 1):
            id_worker.get_id()
        # 序列号应该被重置
        assert id_worker.sequence >= 0

    def test_til_next_millis(self):
        """测试等待下一毫秒"""
        id_worker = IdWorker(datacenter_id=1, worker_id=1)
        last_timestamp = id_worker._gen_timestamp()
        next_timestamp = id_worker._til_next_millis(last_timestamp)
        assert next_timestamp > last_timestamp

    def test_clock_backwards(self):
        """测试时钟回拨异常"""
        id_worker = IdWorker(datacenter_id=1, worker_id=1)
        # 先生成一个ID
        id_worker.get_id()
        # 模拟时钟回拨
        id_worker.last_timestamp = id_worker._gen_timestamp() + 10000
        with pytest.raises(Exception):
            id_worker.get_id()


class TestSnowIdFunction:
    """测试snow_id函数"""

    def test_snow_id_returns_int(self):
        """测试snow_id返回整数"""
        result = snow_id()
        assert isinstance(result, int)
        assert result > 0

    def test_snow_id_unique(self):
        """测试snow_id生成唯一ID"""
        import time
        ids = set()
        for _ in range(100):
            ids.add(snow_id())
            # 添加微小延迟确保序列号递增
            time.sleep(0.0001)
        assert len(ids) == 100


class TestWorkerGlobal:
    """测试全局worker对象"""

    def test_worker_exists(self):
        """测试全局worker对象存在"""
        assert worker is not None
        assert isinstance(worker, IdWorker)

    def test_worker_generates_id(self):
        """测试全局worker可以生成ID"""
        id1 = worker.get_id()
        assert isinstance(id1, int)
        assert id1 > 0

