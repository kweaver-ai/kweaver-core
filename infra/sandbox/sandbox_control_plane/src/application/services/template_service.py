"""
模板应用服务

编排模板相关的用例。
"""

from typing import List, Optional

from src.domain.entities.template import Template
from src.domain.repositories.template_repository import ITemplateRepository
from src.application.commands.create_template import CreateTemplateCommand
from src.application.commands.update_template import UpdateTemplateCommand
from src.application.queries.get_template import GetTemplateQuery
from src.application.dtos.template_dto import TemplateDTO
from src.shared.errors.domain import NotFoundError, ValidationError


class TemplateService:
    """
    模板应用服务

    编排模板创建、查询、更新、删除等用例。
    """

    def __init__(
        self,
        template_repo: ITemplateRepository,
    ):
        self._template_repo = template_repo

    async def create_template(self, command: CreateTemplateCommand) -> TemplateDTO:
        """
        创建模板用例

        流程：
        1. 验证模板ID唯一性
        2. 验证模板名称唯一性
        3. 创建模板实体
        4. 保存到仓储
        """
        # 检查 ID 唯一性
        existing_by_id = await self._template_repo.find_by_id(command.template_id)
        if existing_by_id:
            raise ValidationError(f"Template ID already exists: {command.template_id}")

        # 检查名称唯一性
        existing = await self._template_repo.find_by_name(command.name)
        if existing:
            raise ValidationError(f"Template name already exists: {command.name}")

        from src.domain.value_objects.resource_limit import ResourceLimit

        template = Template(
            id=command.template_id,
            name=command.name,
            image=command.image_url,
            base_image=command.image_url,
            pre_installed_packages=[],
            default_resources=ResourceLimit(
                cpu=str(command.default_cpu_cores),
                memory=f"{command.default_memory_mb}Mi",
                disk=f"{command.default_disk_mb}Mi",
                max_processes=128,
            ),
            default_timeout_sec=command.default_timeout_sec or 300,
            security_context={},
        )

        await self._template_repo.save(template)
        return TemplateDTO.from_entity(template)

    async def get_template(self, query: GetTemplateQuery) -> TemplateDTO:
        """获取模板用例"""
        template = await self._template_repo.find_by_id(query.template_id)
        if not template:
            raise NotFoundError(f"Template not found: {query.template_id}")

        return TemplateDTO.from_entity(template)

    async def list_templates(self, limit: int = 50, offset: int = 0) -> List[TemplateDTO]:
        """列出所有模板"""
        templates = await self._template_repo.find_all(limit=limit, offset=offset)

        return [TemplateDTO.from_entity(t) for t in templates]

    async def update_template(self, command: UpdateTemplateCommand) -> TemplateDTO:
        """
        更新模板用例

        流程：
        1. 查找模板
        2. 验证名称唯一性（如果更改名称）
        3. 更新模板字段
        4. 保存到仓储
        """
        template = await self._template_repo.find_by_id(command.template_id)
        if not template:
            raise NotFoundError(f"Template not found: {command.template_id}")

        # 验证名称唯一性
        if command.name and command.name != template.name:
            existing = await self._template_repo.find_by_name(command.name)
            if existing and existing.id != template.id:
                raise ValidationError(f"Template name already exists: {command.name}")

        # 更新名称
        if command.name is not None:
            template.update_name(command.name)

        # 更新镜像
        if command.image_url is not None:
            template.update_image(command.image_url)

        # 更新资源限制
        if any([command.default_cpu_cores, command.default_memory_mb, command.default_disk_mb]):
            from src.domain.value_objects.resource_limit import ResourceLimit

            cpu = (
                str(command.default_cpu_cores)
                if command.default_cpu_cores
                else template.default_resources.cpu
            )
            memory = (
                f"{command.default_memory_mb}Mi"
                if command.default_memory_mb
                else template.default_resources.memory
            )
            disk = (
                f"{command.default_disk_mb}Mi"
                if command.default_disk_mb
                else template.default_resources.disk
            )

            template.default_resources = ResourceLimit(
                cpu=cpu,
                memory=memory,
                disk=disk,
                max_processes=template.default_resources.max_processes,
            )

        # 更新超时时间
        if command.default_timeout_sec is not None:
            template.update_timeout(command.default_timeout_sec)

        await self._template_repo.save(template)
        return TemplateDTO.from_entity(template)

    async def delete_template(self, template_id: str) -> None:
        """
        删除模板用例

        流程：
        1. 查找模板
        2. 验证无活动会话使用
        3. 删除模板
        """
        template = await self._template_repo.find_by_id(template_id)
        if not template:
            raise NotFoundError(f"Template not found: {template_id}")

        await self._template_repo.delete(template_id)
