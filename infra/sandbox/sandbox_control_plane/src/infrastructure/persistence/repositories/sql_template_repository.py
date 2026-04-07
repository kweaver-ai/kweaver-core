"""
模板仓储实现

使用 SQLAlchemy 实现模板仓储接口。
按照数据表命名规范使用 f_ 前缀字段名。
"""
import re
import json
import time
from typing import List
from decimal import Decimal
from sqlalchemy import select, delete, func
from sqlalchemy.ext.asyncio import AsyncSession

from src.domain.repositories.template_repository import ITemplateRepository
from src.domain.entities.template import Template
from src.infrastructure.persistence.models.template_model import TemplateModel


class SqlTemplateRepository(ITemplateRepository):
    """
    模板仓储实现

    这是基础设施层的 Adapter，实现领域层定义的 Port。
    """

    def __init__(self, session: AsyncSession):
        self._session = session

    async def save(self, template: Template) -> None:
        """保存模板"""
        model = await self._session.get(TemplateModel, template.id)
        now_ms = int(time.time() * 1000)

        def parse_mb_value(value: str) -> int:
            """解析资源值（将 '512Mi', '1Gi' 等转换为 MB）"""
            if not value:
                return 512  # 默认值

            # 提取数字部分
            numeric_str = re.sub(r'[^0-9.]', '', value)
            if not numeric_str:
                return 512

            numeric = float(numeric_str)

            # 根据单位转换
            if 'Gi' in value or 'GB' in value or 'G' in value:
                return int(numeric * 1024)
            elif 'Mi' in value or 'MB' in value or 'M' in value:
                return int(numeric)
            elif 'Ki' in value or 'KB' in value or 'K' in value:
                return int(numeric / 1024)
            else:
                # 如果没有单位，假设是 MB
                return int(numeric)

        if model:
            # 更新现有记录
            model.f_name = template.name
            model.f_description = ""
            model.f_image_url = template.image
            model.f_base_image = template.base_image
            model.f_pre_installed_packages = json.dumps(template.pre_installed_packages, ensure_ascii=False) if template.pre_installed_packages else "[]"
            model.f_runtime_type = "python3.11"  # Default, should be from entity
            model.f_default_cpu_cores = Decimal(template.default_resources.cpu)
            model.f_default_memory_mb = parse_mb_value(template.default_resources.memory)
            model.f_default_disk_mb = parse_mb_value(template.default_resources.disk)
            model.f_default_timeout_sec = template.default_timeout_sec
            model.f_security_context = json.dumps(template.security_context, ensure_ascii=False) if template.security_context else "{}"
            model.f_updated_at = now_ms
        else:
            # 创建新记录
            model = TemplateModel.from_entity(template)
            self._session.add(model)

        await self._session.flush()

    async def find_by_id(self, template_id: str) -> Template | None:
        """根据 ID 查找模板"""
        model = await self._session.get(TemplateModel, template_id)
        return model.to_entity() if model else None

    async def find_by_name(self, name: str) -> Template | None:
        """根据名称查找模板"""
        stmt = select(TemplateModel).where(TemplateModel.f_name == name)
        result = await self._session.execute(stmt)
        model = result.scalar_one_or_none()
        return model.to_entity() if model else None

    async def find_all(self, offset: int = 0, limit: int = 100) -> List[Template]:
        """查找所有模板"""
        stmt = (
            select(TemplateModel)
            .offset(offset)
            .limit(limit)
            .order_by(TemplateModel.f_name)
        )
        result = await self._session.execute(stmt)
        return [model.to_entity() for model in result.scalars().all()]

    async def delete(self, template_id: str) -> None:
        """删除模板"""
        stmt = delete(TemplateModel).where(TemplateModel.f_id == template_id)
        await self._session.execute(stmt)
        await self._session.flush()

    async def exists(self, template_id: str) -> bool:
        """检查模板是否存在"""
        model = await self._session.get(TemplateModel, template_id)
        return model is not None

    async def exists_by_name(self, name: str) -> bool:
        """检查名称是否存在"""
        stmt = select(func.count()).select_from(TemplateModel).where(TemplateModel.f_name == name)
        result = await self._session.execute(stmt)
        return (result.scalar() or 0) > 0

    async def count(self) -> int:
        """统计模板数量"""
        stmt = select(func.count()).select_from(TemplateModel)
        result = await self._session.execute(stmt)
        return result.scalar() or 0
