"""Dolphin SDK 延迟导入管理器

提供对 Dolphin SDK 的延迟导入功能，避免在模块加载时强制依赖 Dolphin SDK。
"""

import sys
from typing import Any, Optional, Type, Dict
from functools import wraps


class LazyDolphinImporter:
    """Dolphin SDK 延迟导入管理器

    在模块级别延迟导入 Dolphin SDK，允许代码在 Dolphin SDK 不可用时正常工作。
    只有在实际需要 Dolphin SDK 功能时才会尝试导入。
    """

    _instance: Optional["LazyDolphinImporter"] = None

    def __new__(cls):
        if cls._instance is None:
            cls._instance = super().__new__(cls)
            cls._instance._initialized = False
        return cls._instance

    def __init__(self):
        if not self._initialized:
            self._import_cache: Dict[str, Any] = {}
            self._available: Optional[bool] = None
            self._initialized = True

    @property
    def available(self) -> bool:
        """检查 Dolphin SDK 是否可用"""
        if self._available is None:
            try:
                import importlib.util

                # 尝试检查是否是真实的 dolphin SDK（不是 mock）
                try:
                    dolphin_spec = importlib.util.find_spec(
                        "dolphin.core.common.exceptions"
                    )
                    # 如果找到的是真实模块（有 __file__ 属性指向实际文件）
                    if dolphin_spec and dolphin_spec.origin:
                        # 检查是否是真实的安装包，不是 mock
                        self._available = (
                            "site-packages" in dolphin_spec.origin
                            or "dist-packages" in dolphin_spec.origin
                        )
                    else:
                        self._available = False
                except (ImportError, ModuleNotFoundError):
                    self._available = False
            except Exception:
                self._available = False
        return self._available

    def get_module(self, module_path: str):
        """延迟导入 Dolphin 模块

        Args:
            module_path: 模块路径，如 'dolphin.core.common.exceptions'

        Returns:
            模块对象或 Mock 对象
        """
        if module_path in self._import_cache:
            return self._import_cache[module_path]

        try:
            module = __import__(module_path, fromlist=[""])
            self._import_cache[module_path] = module
            return module
        except ImportError:
            # SDK 不可用时返回 Mock
            if not self.available:
                # 创建 Mock 模块
                import types

                mock_module = types.ModuleType(module_path)
                sys.modules[module_path] = mock_module
                self._import_cache[module_path] = mock_module
                return mock_module
            raise

    def get_exception_class(self, exception_name: str) -> Type[Exception]:
        """获取 Dolphin 异常类

        Args:
            exception_name: 异常类名，如 'ModelException'

        Returns:
            异常类，如果 SDK 不可用则返回通用异常
        """
        cache_key = f"exception_{exception_name}"
        if cache_key in self._import_cache:
            return self._import_cache[cache_key]

        if not self.available:
            # SDK 不可用时，创建通用的异常类
            class GenericDolphinException(Exception):
                pass

            # 设置类名
            GenericDolphinException.__name__ = exception_name
            GenericDolphinException.__qualname__ = f"Dolphin{exception_name}"
            self._import_cache[cache_key] = GenericDolphinException
            return GenericDolphinException

        try:
            exceptions_module = self.get_module("dolphin.core.common.exceptions")
            exception_class = getattr(exceptions_module, exception_name)
            self._import_cache[cache_key] = exception_class
            return exception_class
        except (ImportError, AttributeError):
            # 回退到通用异常
            class GenericDolphinException(Exception):
                pass

            GenericDolphinException.__name__ = exception_name
            self._import_cache[cache_key] = GenericDolphinException
            return GenericDolphinException

    def get_var_output_class(self):
        """获取 VarOutput 类

        Returns:
            VarOutput 类或 Mock 类
        """
        cache_key = "var_output_class"
        if cache_key in self._import_cache:
            return self._import_cache[cache_key]

        if not self.available:
            # SDK 不可用时，创建 Mock VarOutput 类
            class MockVarOutput:
                def __init__(self):
                    self._storage = {}

                def set(self, key: str, value: Any) -> None:
                    self._storage[key] = value

                def get(self, key: str, default: Any = None) -> Any:
                    return self._storage.get(key, default)

                def delete(self, key: str) -> None:
                    self._storage.pop(key, None)

            self._import_cache[cache_key] = MockVarOutput
            return MockVarOutput

        try:
            context_module = self.get_module("dolphin.core.context.var_output")
            VarOutput = getattr(context_module, "VarOutput")
            self._import_cache[cache_key] = VarOutput
            return VarOutput
        except (ImportError, AttributeError):
            # 回退到 Mock
            return self.get_var_output_class()


# 全局单例
_dolphin_importer = LazyDolphinImporter()


def is_dolphin_available() -> bool:
    """检查 Dolphin SDK 是否可用"""
    return _dolphin_importer.available


def get_dolphin_exception(exception_name: str) -> Type[Exception]:
    """获取 Dolphin 异常类

    Args:
        exception_name: 异常类名

    Returns:
        异常类
    """
    return _dolphin_importer.get_exception_class(exception_name)


def get_dolphin_var_output_class():
    """获取 Dolphin VarOutput 类

    Returns:
        VarOutput 类
    """
    return _dolphin_importer.get_var_output_class()


def create_dolphin_exception(exception_name: str, message: str) -> Exception:
    """创建 Dolphin 异常实例

    Args:
        exception_name: 异常类名
        message: 异常消息

    Returns:
        异常实例
    """
    ExceptionClass = get_dolphin_exception(exception_name)
    return ExceptionClass(message)


def lazy_import_dolphin(func):
    """装饰器：延迟导入 Dolphin SDK

    用在需要使用 Dolphin SDK 的函数上，只有函数被调用时才会尝试导入。

    Example:
        @lazy_import_dolphin
        def some_function():
            from dolphin.core.context.var_output import VarOutput
            # 使用 VarOutput
    """

    @wraps(func)
    def wrapper(*args, **kwargs):
        # 在函数调用前预先检查并缓存导入
        _dolphin_importer.available
        return func(*args, **kwargs)

    return wrapper


# 预加载常用异常类到模块级别，提供向后兼容的导入路径
ModelException = get_dolphin_exception("ModelException")
SkillException = get_dolphin_exception("SkillException")
DolphinException = get_dolphin_exception("DolphinException")
