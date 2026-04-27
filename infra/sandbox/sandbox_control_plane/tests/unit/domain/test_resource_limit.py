"""
资源限制值对象单元测试

测试 ResourceLimit 值对象的行为。
"""
import pytest

from src.domain.value_objects.resource_limit import ResourceLimit


class TestResourceLimit:
    """资源限制值对象测试"""

    def test_create_default(self):
        """测试创建默认资源限制"""
        limit = ResourceLimit.default()

        assert limit.cpu == "1"
        assert limit.memory == "512Mi"
        assert limit.disk == "1Gi"
        assert limit.max_processes == 128

    def test_create_custom(self):
        """测试创建自定义资源限制"""
        limit = ResourceLimit(
            cpu="2",
            memory="1Gi",
            disk="10Gi",
            max_processes=256
        )

        assert limit.cpu == "2"
        assert limit.memory == "1Gi"
        assert limit.disk == "10Gi"
        assert limit.max_processes == 256

    def test_invalid_cpu(self):
        """测试无效的 CPU 值"""
        with pytest.raises(ValueError, match="Invalid cpu format"):
            ResourceLimit(
                cpu="invalid",
                memory="512Mi",
                disk="1Gi"
            )

    def test_negative_cpu(self):
        """测试负数 CPU"""
        with pytest.raises(ValueError, match="cpu must be positive"):
            ResourceLimit(
                cpu="-1",
                memory="512Mi",
                disk="1Gi"
            )

    def test_invalid_memory_format(self):
        """测试无效的内存格式"""
        with pytest.raises(ValueError, match="Invalid memory format"):
            ResourceLimit(
                cpu="1",
                memory="invalid",
                disk="1Gi"
            )

    def test_invalid_disk_format(self):
        """测试无效的磁盘格式"""
        with pytest.raises(ValueError, match="Invalid disk format"):
            ResourceLimit(
                cpu="1",
                memory="512Mi",
                disk="invalid"
            )

    def test_negative_max_processes(self):
        """测试负数最大进程数"""
        with pytest.raises(ValueError, match="max_processes must be positive"):
            ResourceLimit(
                cpu="1",
                memory="512Mi",
                disk="1Gi",
                max_processes=-1
            )

    def test_with_cpu(self):
        """测试创建新的 CPU 限制"""
        limit = ResourceLimit.default()
        new_limit = limit.with_cpu("2")

        # 原对象不变
        assert limit.cpu == "1"

        # 新对象有新值
        assert new_limit.cpu == "2"
        assert new_limit.memory == limit.memory

    def test_with_memory(self):
        """测试创建新的内存限制"""
        limit = ResourceLimit.default()
        new_limit = limit.with_memory("1Gi")

        # 原对象不变
        assert limit.memory == "512Mi"

        # 新对象有新值
        assert new_limit.memory == "1Gi"
        assert new_limit.cpu == limit.cpu

    def test_frozen(self):
        """测试值对象不可变"""
        limit = ResourceLimit.default()

        with pytest.raises(Exception):  # frozen.dataclass 会抛出异常
            limit.cpu = "2"

    def test_valid_size_formats(self):
        """测试有效的大小格式"""
        valid_formats = [
            ("512Mi", True),
            ("1Gi", True),
            ("10Gi", True),
            ("100Mi", True),
            ("invalid", False),
            ("100", False),
            ("", False),
        ]

        for size, should_be_valid in valid_formats:
            if should_be_valid:
                ResourceLimit(
                    cpu="1",
                    memory=size,
                    disk="1Gi"
                )
            else:
                with pytest.raises(ValueError):
                    ResourceLimit(
                        cpu="1",
                        memory=size,
                        disk="1Gi"
                    )
