"""依赖管理和延迟导入模块

提供：
1. Dolphin SDK 延迟导入机制
2. 依赖注入接口定义
3. 默认实现
"""

from app.common.dependencies.dolphin_lazy_import import (
    LazyDolphinImporter,
    is_dolphin_available,
    get_dolphin_exception,
    get_dolphin_var_output_class,
    create_dolphin_exception,
    lazy_import_dolphin,
    ModelException,
    SkillException,
    DolphinException,
)

from app.common.dependencies.interfaces import (
    IContextVarManager,
    IExceptionHandler,
    ICallerInfoProvider,
    IEnvironmentDetector,
    ICacheService,
    SerializationType,
)

from app.common.dependencies.default_implementations import (
    DefaultContextVarManager,
    DefaultExceptionHandler,
    DefaultCallerInfoProvider,
    DefaultEnvironmentDetector,
    get_default_context_var_manager,
    get_default_exception_handler,
    get_default_caller_info_provider,
    get_default_environment_detector,
    reset_default_instances,
)

__all__ = [
    # 延迟导入
    "LazyDolphinImporter",
    "is_dolphin_available",
    "get_dolphin_exception",
    "get_dolphin_var_output_class",
    "create_dolphin_exception",
    "lazy_import_dolphin",
    "ModelException",
    "SkillException",
    "DolphinException",
    # 接口
    "IContextVarManager",
    "IExceptionHandler",
    "ICallerInfoProvider",
    "IEnvironmentDetector",
    "ICacheService",
    "SerializationType",
    # 默认实现
    "DefaultContextVarManager",
    "DefaultExceptionHandler",
    "DefaultCallerInfoProvider",
    "DefaultEnvironmentDetector",
    "get_default_context_var_manager",
    "get_default_exception_handler",
    "get_default_caller_info_provider",
    "get_default_environment_detector",
    "reset_default_instances",
]
