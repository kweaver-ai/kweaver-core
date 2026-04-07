"""
文件制品值对象单元测试

测试 Artifact 值对象的行为。
"""
import pytest
from datetime import datetime

from src.domain.value_objects.artifact import Artifact, ArtifactType


class TestArtifactType:
    """制品类型枚举测试"""

    def test_artifact_values(self):
        """测试制品类型枚举值"""
        assert ArtifactType.ARTIFACT == "artifact"
        assert ArtifactType.LOG == "log"
        assert ArtifactType.OUTPUT == "output"


class TestArtifact:
    """制品值对象测试"""

    def test_create_artifact_success(self):
        """测试成功创建制品"""
        artifact = Artifact(
            path="output/result.csv",
            size=1024,
            mime_type="text/csv",
            type=ArtifactType.ARTIFACT,
            created_at=datetime.now()
        )

        assert artifact.path == "output/result.csv"
        assert artifact.size == 1024
        assert artifact.mime_type == "text/csv"
        assert artifact.type == ArtifactType.ARTIFACT

    def test_create_artifact_factory_method(self):
        """测试使用工厂方法创建制品"""
        artifact = Artifact.create(
            path="logs/execution.log",
            size=2048,
            mime_type="text/plain",
            type="log"
        )

        assert artifact.path == "logs/execution.log"
        assert artifact.size == 2048
        assert artifact.mime_type == "text/plain"
        assert artifact.type == ArtifactType.LOG

    def test_create_artifact_with_checksum(self):
        """测试创建带校验和的制品"""
        artifact = Artifact.create(
            path="output/data.json",
            size=4096,
            mime_type="application/json",
            type="artifact",
            checksum="a1b2c3d4e5f6"
        )

        assert artifact.checksum == "a1b2c3d4e5f6"

    def test_create_artifact_invalid_size(self):
        """测试无效的文件大小"""
        with pytest.raises(ValueError, match="size cannot be negative"):
            Artifact.create(
                path="output/file.txt",
                size=-1,
                mime_type="text/plain"
            )

    def test_create_artifact_empty_path(self):
        """测试空路径"""
        with pytest.raises(ValueError, match="path cannot be empty"):
            Artifact.create(
                path="",
                size=100,
                mime_type="text/plain"
            )

    def test_is_log(self):
        """测试是否为日志文件"""
        log_artifact = Artifact.create(
            path="logs/output.log",
            size=1024,
            mime_type="text/plain",
            type="log"
        )
        assert log_artifact.is_log() is True

        data_artifact = Artifact.create(
            path="output/data.csv",
            size=1024,
            mime_type="text/csv",
            type="artifact"
        )
        assert data_artifact.is_log() is False

    def test_is_output(self):
        """测试是否为输出文件"""
        output_artifact = Artifact.create(
            path="output/stdout.txt",
            size=2048,
            mime_type="text/plain",
            type="output"
        )
        assert output_artifact.is_output() is True

        log_artifact = Artifact.create(
            path="logs/output.log",
            size=1024,
            mime_type="text/plain",
            type="log"
        )
        assert log_artifact.is_output() is False

    def test_immutability(self):
        """测试不可变性"""
        artifact = Artifact.create(
            path="output/file.txt",
            size=1024,
            mime_type="text/plain"
        )

        # frozen=True 意味着不可修改
        with pytest.raises(Exception):  # FrozenInstanceError from dataclasses
            artifact.size = 2048

    def test_different_mime_types(self):
        """测试不同的 MIME 类型"""
        csv_artifact = Artifact.create(
            path="data.csv",
            size=1024,
            mime_type="text/csv",
            type="artifact"
        )
        assert csv_artifact.mime_type == "text/csv"

        json_artifact = Artifact.create(
            path="data.json",
            size=2048,
            mime_type="application/json",
            type="artifact"
        )
        assert json_artifact.mime_type == "application/json"

        png_artifact = Artifact.create(
            path="plot.png",
            size=4096,
            mime_type="image/png",
            type="artifact"
        )
        assert png_artifact.mime_type == "image/png"

    def test_zero_size_artifact(self):
        """测试零大小制品"""
        artifact = Artifact.create(
            path="output/empty.txt",
            size=0,
            mime_type="text/plain"
        )
        assert artifact.size == 0

    def test_artifact_equality(self):
        """测试制品相等性"""
        dt = datetime.now()
        artifact1 = Artifact(
            path="output/file.txt",
            size=1024,
            mime_type="text/plain",
            type=ArtifactType.ARTIFACT,
            created_at=dt
        )
        artifact2 = Artifact(
            path="output/file.txt",
            size=1024,
            mime_type="text/plain",
            type=ArtifactType.ARTIFACT,
            created_at=dt
        )
        assert artifact1 == artifact2

        artifact3 = Artifact(
            path="output/other.txt",
            size=1024,
            mime_type="text/plain",
            type=ArtifactType.ARTIFACT,
            created_at=dt
        )
        assert artifact1 != artifact3
