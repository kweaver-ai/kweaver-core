"""扩展的 Dolphin Mock 配置

为 test/unit/conftest.py 提供额外的 dolphin 子模块 mock。
"""

import sys
from unittest.mock import MagicMock


def setup_extended_dolphin_mocks():
    """设置扩展的 Dolphin SDK mock

    添加更多 dolphin 子模块的 mock，以支持更全面的测试。
    """

    # 需要额外 mock 的模块列表
    extra_modules = [
        "dolphin.core.common.constants",
        "dolphin.core.utils",
        "dolphin.core.utils.tools",
    ]

    for module_name in extra_modules:
        if module_name not in sys.modules:
            mock_module = MagicMock()
            mock_module.__name__ = module_name
            mock_module.__file__ = f"{module_name.replace('.', '/')}/__init__.py"

            # 为特定模块添加属性
            if "constants" in module_name:
                mock_module.KEY_SESSION_ID = "session_id"
                mock_module.KEY_USER_ID = "user_id"
                mock_module.KEY_AGENT_ID = "agent_id"

            sys.modules[module_name] = mock_module


# 自动调用
if "dolphin.core.common.constants" not in sys.modules:
    setup_extended_dolphin_mocks()
