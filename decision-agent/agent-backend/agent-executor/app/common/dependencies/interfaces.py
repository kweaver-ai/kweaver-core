"""依赖注入接口定义

定义系统中需要的外部依赖接口，实现依赖倒置原则。
"""

from abc import ABC, abstractmethod
from typing import Any, Optional, Dict, Tuple
from enum import Enum


class IContextVarManager(ABC):
    """上下文变量管理器接口

    用于管理运行时上下文变量，替代直接依赖 Dolphin SDK。
    """

    @abstractmethod
    def get(self, key: str, default: Any = None) -> Any:
        """获取变量值

        Args:
            key: 变量键
            default: 默认值

        Returns:
            变量值
        """
        pass

    @abstractmethod
    def set(self, key: str, value: Any) -> None:
        """设置变量值

        Args:
            key: 变量键
            value: 变量值
        """
        pass

    @abstractmethod
    def delete(self, key: str) -> None:
        """删除变量

        Args:
            key: 变量键
        """
        pass

    @abstractmethod
    def exists(self, key: str) -> bool:
        """检查变量是否存在

        Args:
            key: 变量键

        Returns:
            是否存在
        """
        pass

    @abstractmethod
    def get_all(self) -> Dict[str, Any]:
        """获取所有变量

        Returns:
            所有变量的字典
        """
        pass


class IExceptionHandler(ABC):
    """异常处理器接口

    提供统一的异常创建和处理接口。
    """

    @abstractmethod
    def create_model_exception(self, message: str) -> Exception:
        """创建模型异常

        Args:
            message: 异常消息

        Returns:
            异常实例
        """
        pass

    @abstractmethod
    def create_skill_exception(self, message: str) -> Exception:
        """创建技能异常

        Args:
            message: 异常消息

        Returns:
            异常实例
        """
        pass

    @abstractmethod
    def create_dolphin_exception(self, message: str) -> Exception:
        """创建 Dolphin 异常

        Args:
            message: 异常消息

        Returns:
            异常实例
        """
        pass

    @abstractmethod
    def is_available(self) -> bool:
        """检查 Dolphin SDK 是否可用"""
        pass


class ICallerInfoProvider(ABC):
    """调用者信息提供者接口"""

    @abstractmethod
    def get_caller_info(self) -> Tuple[str, int]:
        """获取调用者信息

        Returns:
            (文件名, 行号) 元组
        """
        pass


class IEnvironmentDetector(ABC):
    """环境检测器接口"""

    @abstractmethod
    def is_in_pod(self) -> bool:
        """检查是否在 Kubernetes Pod 中运行"""
        pass

    @abstractmethod
    def get_environment_type(self) -> str:
        """获取环境类型

        Returns:
            环境类型：'pod', 'local', 'unknown'
        """
        pass


class SerializationType(Enum):
    """序列化类型"""

    JSON = "json"
    PICKLE = "pickle"


class ICacheService(ABC):
    """缓存服务接口"""

    @abstractmethod
    async def get(self, key: str) -> Optional[Any]:
        """获取缓存

        Args:
            key: 缓存键

        Returns:
            缓存值，不存在返回 None
        """
        pass

    @abstractmethod
    async def set(
        self,
        key: str,
        value: Any,
        ttl: Optional[int] = None,
        serialization_type: SerializationType = SerializationType.JSON,
    ) -> bool:
        """设置缓存

        Args:
            key: 缓存键
            value: 缓存值
            ttl: 过期时间（秒）
            serialization_type: 序列化类型

        Returns:
            是否成功
        """
        pass

    @abstractmethod
    async def delete(self, key: str) -> bool:
        """删除缓存

        Args:
            key: 缓存键

        Returns:
            是否成功
        """
        pass

    @abstractmethod
    async def exists(self, key: str) -> bool:
        """检查缓存是否存在

        Args:
            key: 缓存键

        Returns:
            是否存在
        """
        pass
