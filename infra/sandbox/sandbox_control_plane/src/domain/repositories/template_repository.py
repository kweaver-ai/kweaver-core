"""
模板仓储接口

定义模板持久化的抽象接口（Port）。
"""
from abc import ABC, abstractmethod
from typing import List, Optional

from src.domain.entities.template import Template


class ITemplateRepository(ABC):
    """
    模板仓储接口

    这是领域层定义的 Port，由基础设施层实现 Adapter。
    """

    @abstractmethod
    async def save(self, template: Template) -> None:
        """保存模板（创建或更新）"""
        pass

    @abstractmethod
    async def find_by_id(self, template_id: str) -> Optional[Template]:
        """根据 ID 查找模板"""
        pass

    @abstractmethod
    async def find_by_name(self, name: str) -> Optional[Template]:
        """根据名称查找模板"""
        pass

    @abstractmethod
    async def find_all(
        self,
        offset: int = 0,
        limit: int = 100
    ) -> List[Template]:
        """查找所有模板"""
        pass

    @abstractmethod
    async def delete(self, template_id: str) -> None:
        """删除模板"""
        pass

    @abstractmethod
    async def exists(self, template_id: str) -> bool:
        """检查模板是否存在"""
        pass

    @abstractmethod
    async def exists_by_name(self, name: str) -> bool:
        """检查模板名称是否存在"""
        pass

    @abstractmethod
    async def count(self) -> int:
        """统计模板总数"""
        pass
