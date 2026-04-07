"""
Dolphin SDK 异常
"""

from typing import Optional, Dict, Type, Any

from app.common.exceptions.base_exception import BaseException
from app.common.errors.custom_errors_pkg import (
    DolphinSDKModelError,
    DolphinSDKSkillError,
    DolphinSDKBaseError,
)

# 使用延迟导入替代直接导入 Dolphin SDK
# from dolphin.core.common.exceptions import (
#     ModelException,
#     SkillException,
#     DolphinException,
# )  # ← 移除直接导入


def _get_dolphin_exception_classes() -> Dict[Type[Exception], Any]:
    """延迟获取 Dolphin 异常类映射

    Returns:
        异常类到错误类的映射字典
    """
    try:
        from app.common.dependencies import get_dolphin_exception

        return {
            get_dolphin_exception("ModelException"): DolphinSDKModelError,
            get_dolphin_exception("SkillException"): DolphinSDKSkillError,
            get_dolphin_exception("DolphinException"): DolphinSDKBaseError,
        }
    except ImportError:
        # 如果依赖不可用，返回空映射
        return {}


class DolphinSDKException(BaseException):
    """
    Dolphin SDK 异常

    将 Dolphin SDK 的原生异常包装为项目的异常格式
    """

    def __init__(
        self,
        raw_exception: Exception,
        agent_id: Optional[str],
        session_id: Optional[str],
        user_id: Optional[str] = None,
    ):
        """
        初始化 Dolphin SDK 异常

        Args:
            raw_exception: 原始异常对象
            agent_id: Agent ID
            session_id: 会话 ID
            user_id: 用户 ID
        """
        # 延迟获取异常映射
        exception_to_error_map = _get_dolphin_exception_classes()

        # 根据原始异常类型选择错误定义
        error_func = DolphinSDKBaseError  # 默认值
        for exception_class, error_func_mapped in exception_to_error_map.items():
            try:
                if isinstance(raw_exception, exception_class):
                    error_func = error_func_mapped
                    break
            except TypeError:
                # 异常类型检查失败，跳过
                continue

        # error_func() 会自动从 Config 获取 debug 模式
        super().__init__(
            error=error_func(),
            error_details=f"dolphinsdk exception, details: [{str(raw_exception)}], context info: [agent_id: {agent_id}, session_id: {session_id}, user_id: {user_id}]",
            error_link="",
        )
