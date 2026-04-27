"""
模板 DTO 单元测试

测试 TemplateDTO 的功能。
"""
import pytest
from datetime import datetime

from src.application.dtos.template_dto import TemplateDTO
from src.domain.entities.template import Template
from src.domain.value_objects.resource_limit import ResourceLimit


class TestTemplateDTO:
    """模板 DTO 测试"""

    def test_create_with_required_fields(self):
        """测试使用必填字段创建"""
        dto = TemplateDTO(
            id="python-test",
            name="Python Test",
            image_url="python:3.11",
            runtime_type="python3.11",
            default_cpu_cores=1.0,
            default_memory_mb=512,
            default_disk_mb=1024,
            default_timeout_sec=300
        )

        assert dto.id == "python-test"
        assert dto.name == "Python Test"
        assert dto.image_url == "python:3.11"
        assert dto.runtime_type == "python3.11"
        assert dto.default_cpu_cores == 1.0
        assert dto.default_memory_mb == 512
        assert dto.default_disk_mb == 1024
        assert dto.default_timeout_sec == 300
        assert dto.default_env_vars is None
        assert dto.is_active is True

    def test_create_with_all_fields(self):
        """测试使用所有字段创建"""
        now = datetime.now()
        dto = TemplateDTO(
            id="python-test",
            name="Python Test",
            image_url="python:3.11",
            runtime_type="python3.11",
            default_cpu_cores=1.0,
            default_memory_mb=512,
            default_disk_mb=1024,
            default_timeout_sec=300,
            default_env_vars={"DEBUG": "true"},
            is_active=False,
            created_at=now,
            updated_at=now
        )

        assert dto.default_env_vars == {"DEBUG": "true"}
        assert dto.is_active is False
        assert dto.created_at == now
        assert dto.updated_at == now

    def test_from_entity(self):
        """测试从领域实体创建 DTO"""
        template = Template(
            id="python-test",
            name="Python Test",
            image="python:3.11",
            base_image="python:3.11-slim",
            default_resources=ResourceLimit(
                cpu="1",
                memory="512Mi",
                disk="1Gi"
            ),
            default_timeout_sec=300
        )

        dto = TemplateDTO.from_entity(template)

        assert dto.id == "python-test"
        assert dto.name == "Python Test"
        assert dto.image_url == "python:3.11"
        assert dto.default_cpu_cores == 1.0
        assert dto.default_memory_mb == 512
        assert dto.default_disk_mb == 1024
        assert dto.default_timeout_sec == 300

    def test_from_entity_with_gi_memory(self):
        """测试从领域实体创建 DTO（Gi 内存单位）"""
        template = Template(
            id="python-test",
            name="Python Test",
            image="python:3.11",
            base_image="python:3.11-slim",
            default_resources=ResourceLimit(
                cpu="2",
                memory="2Gi",
                disk="10Gi"
            )
        )

        dto = TemplateDTO.from_entity(template)

        assert dto.default_memory_mb == 2048  # 2Gi = 2048MB
        assert dto.default_disk_mb == 10240  # 10Gi = 10240MB

    def test_from_entity_with_large_resources(self):
        """测试从领域实体创建 DTO（大资源）"""
        template = Template(
            id="python-test",
            name="Python Test",
            image="python:3.11",
            base_image="python:3.11-slim",
            default_resources=ResourceLimit(
                cpu="4",
                memory="8Gi",
                disk="50Gi"
            )
        )

        dto = TemplateDTO.from_entity(template)

        assert dto.default_cpu_cores == 4.0
        assert dto.default_memory_mb == 8192  # 8Gi = 8192MB
        assert dto.default_disk_mb == 51200  # 50Gi = 51200MB

    def test_from_entity_with_default_resources(self):
        """测试从领域实体创建 DTO（使用默认资源）"""
        template = Template(
            id="python-test",
            name="Python Test",
            image="python:3.11",
            base_image="python:3.11-slim",
            default_resources=ResourceLimit.default()
        )

        dto = TemplateDTO.from_entity(template)

        assert dto.default_cpu_cores == 1.0
        assert dto.default_memory_mb == 512
        assert dto.default_disk_mb == 1024

    def test_to_dict(self):
        """测试转换为字典"""
        now = datetime.now()
        dto = TemplateDTO(
            id="python-test",
            name="Python Test",
            image_url="python:3.11",
            runtime_type="python3.11",
            default_cpu_cores=1.0,
            default_memory_mb=512,
            default_disk_mb=1024,
            default_timeout_sec=300,
            default_env_vars={"DEBUG": "true"},
            is_active=True,
            created_at=now,
            updated_at=now
        )

        result = dto.to_dict()

        assert result["id"] == "python-test"
        assert result["name"] == "Python Test"
        assert result["image_url"] == "python:3.11"
        assert result["runtime_type"] == "python3.11"
        assert result["default_cpu_cores"] == 1.0
        assert result["default_memory_mb"] == 512
        assert result["default_disk_mb"] == 1024
        assert result["default_timeout_sec"] == 300
        assert result["default_env_vars"] == {"DEBUG": "true"}
        assert result["is_active"] is True
        assert result["created_at"] == now.isoformat()
        assert result["updated_at"] == now.isoformat()

    def test_to_dict_without_dates(self):
        """测试转换为字典（无日期）"""
        dto = TemplateDTO(
            id="python-test",
            name="Python Test",
            image_url="python:3.11",
            runtime_type="python3.11",
            default_cpu_cores=1.0,
            default_memory_mb=512,
            default_disk_mb=1024,
            default_timeout_sec=300
        )

        result = dto.to_dict()

        assert result["created_at"] is None
        assert result["updated_at"] is None

    def test_is_dataclass(self):
        """测试是数据类"""
        from dataclasses import is_dataclass

        assert is_dataclass(TemplateDTO)
