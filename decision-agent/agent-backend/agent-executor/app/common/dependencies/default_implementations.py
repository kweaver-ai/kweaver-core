"""默认实现

使用延迟导入的 Dolphin SDK 实现，提供与真实 SDK 兼容的接口。
"""

import os
import inspect
from typing import Any, Dict, Tuple, Optional

from app.common.dependencies.interfaces import (
    IContextVarManager,
    IExceptionHandler,
    ICallerInfoProvider,
    IEnvironmentDetector,
)
from app.common.dependencies.dolphin_lazy_import import (
    get_dolphin_var_output_class,
    get_dolphin_exception,
    is_dolphin_available,
)


class DefaultContextVarManager(IContextVarManager):
    """默认上下文变量管理器"""

    def __init__(self):
        self._var_output = None

    def _get_var_output(self):
        """延迟获取 VarOutput 实例"""
        if self._var_output is None:
            VarOutputClass = get_dolphin_var_output_class()
            self._var_output = VarOutputClass()
        return self._var_output

    def get(self, key: str, default: Any = None) -> Any:
        try:
            var_output = self._get_var_output()
            return var_output.get(key, default)
        except AttributeError:
            # Mock VarOutput 可能没有 get 方法，使用备用实现
            if not hasattr(self, "_fallback_storage"):
                self._fallback_storage = {}
            return self._fallback_storage.get(key, default)

    def set(self, key: str, value: Any) -> None:
        try:
            var_output = self._get_var_output()
            var_output.set(key, value)
        except AttributeError:
            if not hasattr(self, "_fallback_storage"):
                self._fallback_storage = {}
            self._fallback_storage[key] = value

    def delete(self, key: str) -> None:
        try:
            var_output = self._get_var_output()
            var_output.delete(key)
        except AttributeError:
            if hasattr(self, "_fallback_storage") and key in self._fallback_storage:
                del self._fallback_storage[key]

    def exists(self, key: str) -> bool:
        try:
            var_output = self._get_var_output()
            # VarOutput 可能没有 exists 方法
            value = var_output.get(key)
            return value is not None
        except AttributeError:
            if hasattr(self, "_fallback_storage"):
                return key in self._fallback_storage
            return False

    def get_all(self) -> Dict[str, Any]:
        if hasattr(self, "_fallback_storage"):
            return self._fallback_storage.copy()
        return {}


class DefaultExceptionHandler(IExceptionHandler):
    """默认异常处理器"""

    def create_model_exception(self, message: str) -> Exception:
        return get_dolphin_exception("ModelException")(message)

    def create_skill_exception(self, message: str) -> Exception:
        return get_dolphin_exception("SkillException")(message)

    def create_dolphin_exception(self, message: str) -> Exception:
        return get_dolphin_exception("DolphinException")(message)

    def is_available(self) -> bool:
        return is_dolphin_available()


class DefaultCallerInfoProvider(ICallerInfoProvider):
    """默认调用者信息提供者"""

    def get_caller_info(self) -> Tuple[str, int]:
        """获取调用者信息"""
        cur_pwd = os.getcwd()
        caller_frame = inspect.stack()[2]  # 跳过两层栈（当前函数和调用者）
        caller_filename = caller_frame.filename.split(cur_pwd)[-1][1:]
        caller_lineno = caller_frame.lineno
        return caller_filename, caller_lineno


class DefaultEnvironmentDetector(IEnvironmentDetector):
    """默认环境检测器"""

    def is_in_pod(self) -> bool:
        """检查是否在 Kubernetes Pod 中运行"""
        return (
            "KUBERNETES_SERVICE_HOST" in os.environ
            and "KUBERNETES_SERVICE_PORT" in os.environ
        )

    def get_environment_type(self) -> str:
        """获取环境类型"""
        if self.is_in_pod():
            return "pod"
        elif os.getenv("ENVIRONMENT") == "test":
            return "test"
        elif os.getenv("ENVIRONMENT") == "development":
            return "local"
        else:
            return "unknown"


# 全局默认实例
_default_context_var_manager: Optional[DefaultContextVarManager] = None
_default_exception_handler: Optional[DefaultExceptionHandler] = None
_default_caller_info_provider: Optional[DefaultCallerInfoProvider] = None
_default_environment_detector: Optional[DefaultEnvironmentDetector] = None


def get_default_context_var_manager() -> DefaultContextVarManager:
    """获取默认上下文变量管理器"""
    global _default_context_var_manager
    if _default_context_var_manager is None:
        _default_context_var_manager = DefaultContextVarManager()
    return _default_context_var_manager


def get_default_exception_handler() -> DefaultExceptionHandler:
    """获取默认异常处理器"""
    global _default_exception_handler
    if _default_exception_handler is None:
        _default_exception_handler = DefaultExceptionHandler()
    return _default_exception_handler


def get_default_caller_info_provider() -> DefaultCallerInfoProvider:
    """获取默认调用者信息提供者"""
    global _default_caller_info_provider
    if _default_caller_info_provider is None:
        _default_caller_info_provider = DefaultCallerInfoProvider()
    return _default_caller_info_provider


def get_default_environment_detector() -> DefaultEnvironmentDetector:
    """获取默认环境检测器"""
    global _default_environment_detector
    if _default_environment_detector is None:
        _default_environment_detector = DefaultEnvironmentDetector()
    return _default_environment_detector


def reset_default_instances():
    """重置所有默认实例（用于测试）"""
    global _default_context_var_manager
    global _default_exception_handler
    global _default_caller_info_provider
    global _default_environment_detector

    _default_context_var_manager = None
    _default_exception_handler = None
    _default_caller_info_provider = None
    _default_environment_detector = None
