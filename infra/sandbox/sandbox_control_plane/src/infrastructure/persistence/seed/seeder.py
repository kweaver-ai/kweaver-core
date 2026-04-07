"""
数据库种子数据初始化

提供统一的种子数据初始化逻辑，可以在应用启动或独立脚本中调用。
"""
from sqlalchemy import select
from structlog import get_logger

from src.infrastructure.persistence.database import db_manager
from src.infrastructure.persistence.models.runtime_node_model import RuntimeNodeModel
from src.infrastructure.persistence.models.template_model import TemplateModel
from src.infrastructure.persistence.seed.default_data import (
    get_default_runtime_nodes,
    get_default_templates,
)

logger = get_logger(__name__)


async def seed_runtime_nodes(force: bool = False) -> int:
    """
    初始化默认运行时节点

    Args:
        force: 如果为 True，即使节点已存在也会重新创建

    Returns:
        创建的节点数量
    """
    async with db_manager.get_session() as session:
        # 检查是否已存在节点
        result = await session.execute(select(RuntimeNodeModel))
        existing_nodes = result.scalars().all()

        if existing_nodes and not force:
            logger.info(
                "Runtime nodes already exist, skipping initialization",
                count=len(existing_nodes)
            )
            return 0

        # 如果 force=True，删除现有节点
        if existing_nodes and force:
            for node in existing_nodes:
                await session.delete(node)
            await session.flush()
            logger.info("Deleted existing runtime nodes", count=len(existing_nodes))

        # 创建默认节点
        default_nodes = get_default_runtime_nodes()
        for node in default_nodes:
            session.add(node)

        await session.flush()
        logger.info("Created default runtime nodes", count=len(default_nodes))
        return len(default_nodes)


async def seed_templates(force: bool = False) -> int:
    """
    初始化默认模板

    Args:
        force: 如果为 True，即使模板已存在也会重新创建

    Returns:
        创建或更新的模板数量
    """
    async with db_manager.get_session() as session:
        # 获取默认模板定义
        default_templates = get_default_templates()
        default_template_map = {t.f_id: t for t in default_templates}

        # 检查已存在的模板
        result = await session.execute(select(TemplateModel))
        existing_templates = result.scalars().all()
        existing_template_map = {t.f_id: t for t in existing_templates}

        if existing_templates and not force:
            # 更新已存在模板的镜像地址
            updated_count = 0
            for template_id, default_template in default_template_map.items():
                if template_id in existing_template_map:
                    existing_template = existing_template_map[template_id]
                    # 更新镜像地址和其他可能变化的字段
                    if existing_template.f_image_url != default_template.f_image_url:
                        existing_template.f_image_url = default_template.f_image_url
                        updated_count += 1
                        logger.info(
                            "Updated template image URL",
                            template_id=template_id,
                            old_image=existing_template.f_image_url,
                            new_image=default_template.f_image_url
                        )

            # 创建新模板（默认模板中有但数据库中没有的）
            created_count = 0
            for template_id, default_template in default_template_map.items():
                if template_id not in existing_template_map:
                    session.add(default_template)
                    created_count += 1

            await session.flush()
            logger.info(
                "Templates synced",
                updated=updated_count,
                created=created_count
            )
            return updated_count + created_count

        # 如果 force=True，删除现有模板并重新创建
        if existing_templates and force:
            for template in existing_templates:
                await session.delete(template)
            await session.flush()
            logger.info("Deleted existing templates", count=len(existing_templates))

        # 创建默认模板
        for template in default_templates:
            logger.info(
                "Creating template with image URL",
                template_id=template.f_id,
                f_image_url=template.f_image_url
            )
            session.add(template)

        await session.flush()
        logger.info("Created default templates", count=len(default_templates))
        return len(default_templates)


async def seed_default_data(force: bool = False) -> dict:
    """
    初始化所有默认数据

    Args:
        force: 如果为 True，即使数据已存在也会重新创建

    Returns:
        包含创建项数量的字典
    """
    logger.info("Starting default data seeding", force=force)

    node_count = await seed_runtime_nodes(force=force)
    template_count = await seed_templates(force=force)

    result = {
        "runtime_nodes": node_count,
        "templates": template_count,
        "total": node_count + template_count
    }

    logger.info("Completed default data seeding", **result)
    return result
