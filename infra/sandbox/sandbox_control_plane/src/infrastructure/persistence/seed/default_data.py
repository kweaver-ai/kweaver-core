"""
默认数据定义

定义在不同环境中使用的默认运行时节点和模板。
所有默认数据的集中定义，方便维护和修改。
按照数据表命名规范使用 f_ 前缀字段名。
"""
import os
import time
from decimal import Decimal
from typing import List

from structlog import get_logger

from src.infrastructure.persistence.models.runtime_node_model import RuntimeNodeModel
from src.infrastructure.persistence.models.template_model import TemplateModel

logger = get_logger(__name__)


def get_default_template_image_url() -> str:
    """
    获取默认模板镜像 URL

    从环境变量 DEFAULT_TEMPLATE_IMAGE 读取，如果未设置则使用默认值。

    Returns:
        模板镜像 URL
    """
    image_url = os.getenv("DEFAULT_TEMPLATE_IMAGE", "sandbox-template-python-basic:v1.0.0")
    logger.info(
        "Getting default template image URL from environment",
        DEFAULT_TEMPLATE_IMAGE=image_url
    )
    return image_url


def get_default_runtime_nodes() -> List[RuntimeNodeModel]:
    """
    获取默认运行时节点列表

    Returns:
        默认运行时节点列表
    """
    now_ms = int(time.time() * 1000)
    return [
        RuntimeNodeModel(
            f_node_id="docker-local",
            f_hostname="sandbox-control-plane",
            f_runtime_type="docker",
            f_ip_address="127.0.0.1",
            f_api_endpoint="unix:///var/run/docker.sock",
            f_status="online",
            f_total_cpu_cores=Decimal("8.0"),
            f_total_memory_mb=16384,
            f_max_containers=50,
            f_running_containers=0,
            f_allocated_cpu_cores=Decimal("0"),
            f_allocated_memory_mb=0,
            f_cached_images="[]",
            f_labels='{"environment": "development", "type": "default"}',
            f_last_heartbeat_at=now_ms,
            # 审计字段
            f_created_at=now_ms,
            f_created_by="system",
            f_updated_at=now_ms,
            f_updated_by="system",
            f_deleted_at=0,
            f_deleted_by="",
        ),
    ]


def get_default_templates() -> List[TemplateModel]:
    """
    获取默认模板列表

    Returns:
        默认模板列表
    """
    import json
    now_ms = int(time.time() * 1000)
    return [
        TemplateModel(
            f_id="python-basic",
            f_name="Python Basic",
            f_description="基础 Python 执行环境",
            f_image_url=get_default_template_image_url(),
            f_base_image="",
            f_runtime_type="python3.11",
            f_default_cpu_cores=Decimal("1.0"),
            f_default_memory_mb=512,
            f_default_disk_mb=1024,
            f_default_timeout_sec=300,
            f_is_active=1,
            f_pre_installed_packages="[]",
            f_default_env_vars="",
            f_security_context="",
            # 审计字段
            f_created_at=now_ms,
            f_created_by="system",
            f_updated_at=now_ms,
            f_updated_by="system",
            f_deleted_at=0,
            f_deleted_by="",
        ),
    ]
