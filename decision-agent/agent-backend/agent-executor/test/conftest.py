"""测试配置

设置测试环境和全局 fixtures，解决测试隔离问题。
"""

import os
import sys
import pytest
from unittest.mock import MagicMock

# ============================================================
# 1. 在导入任何 app 模块之前，预先 mock dolphin SDK
# ============================================================


def setup_dolphin_mocks():
    """设置 Dolphin SDK 的 Mock

    这个函数必须在任何 app 模块导入之前调用，以防止 ImportError。
    """

    # 如果 dolphin 模块已经被真实模块加载，跳过
    if "dolphin" in sys.modules and hasattr(sys.modules["dolphin"], "__file__"):
        return

    # 创建 Mock 异常类
    def create_exception_class(name: str):
        """创建一个模拟的异常类"""

        class MockException(Exception):
            def __init__(self, message: str = ""):
                super().__init__(message)

            def __str__(self):
                return str(self.args[0]) if self.args else ""

        MockException.__name__ = name
        MockException.__qualname__ = name
        return MockException

    ModelException = create_exception_class("ModelException")
    SkillException = create_exception_class("SkillException")
    DolphinException = create_exception_class("DolphinException")

    # 设置 VarOutput Mock 类
    class MockVarOutput:
        """模拟的 VarOutput 类"""

        _instance = None
        _storage = {}

        def __new__(cls):
            if cls._instance is None:
                cls._instance = super().__new__(cls)
            return cls._instance

        def __init__(self):
            # 每次测试时重置存储
            pass

        @classmethod
        def reset(cls):
            """重置存储（测试用）"""
            cls._storage = {}

        def set(self, key: str, value):
            """设置变量"""
            MockVarOutput._storage[key] = value

        def get(self, key: str, default=None):
            """获取变量"""
            return MockVarOutput._storage.get(key, default)

        def delete(self, key: str):
            """删除变量"""
            MockVarOutput._storage.pop(key, None)

        @classmethod
        def is_serialized_dict(cls, var):
            """检查是否为序列化的字典

            与原始 VarOutput.is_serialized_dict 实现保持一致：
            检查 __type__ 字段是否等于 "VarOutput"
            """
            return isinstance(var, dict) and var.get("__type__") == "VarOutput"

    # 设置 ResumeHandle Mock 类
    class MockResumeHandle:
        """模拟的 ResumeHandle 类"""

        def __init__(
            self,
            frame_id="",
            snapshot_id="",
            resume_token="",
            interrupt_type="",
            current_block="",
            restart_block="",
        ):
            self.frame_id = frame_id
            self.snapshot_id = snapshot_id
            self.resume_token = resume_token
            self.interrupt_type = interrupt_type
            self.current_block = current_block
            self.restart_block = restart_block

    # 设置 Tool Mock 类
    class MockTool:
        """模拟的 Tool 类"""

        pass

    # 创建所有需要的模块
    modules_to_create = {
        "dolphin": None,
        "dolphin.core": None,
        "dolphin.core.common": {
            "constants": {
                "KEY_SESSION_ID": "session_id",
                "KEY_USER_ID": "user_id",
                "KEY_AGENT_ID": "agent_id",
            },
            "exceptions": {
                "ModelException": ModelException,
                "SkillException": SkillException,
                "DolphinException": DolphinException,
            },
        },
        "dolphin.core.context": None,
        "dolphin.core.context.context": None,
        "dolphin.core.context.var_output": {
            "VarOutput": MockVarOutput,
        },
        "dolphin.core.coroutine": None,
        "dolphin.core.coroutine.resume_handle": {
            "ResumeHandle": MockResumeHandle,
        },
        "dolphin.core.utils": None,
        "dolphin.core.utils.tools": {
            "Tool": MockTool,
        },
    }

    # 递归创建模块
    def create_module(fullname, attributes=None):
        """创建一个 mock 模块"""
        from importlib.machinery import ModuleSpec

        mock_module = MagicMock()
        mock_module.__name__ = fullname
        mock_module.__file__ = f"{fullname.replace('.', '/')}/__init__.py"
        mock_module.__path__ = []  # 使其成为包

        # 添加 __spec__ 属性，避免 AttributeError
        mock_module.__spec__ = ModuleSpec(fullname, None, origin="mock")

        if attributes:
            for attr_name, attr_value in attributes.items():
                setattr(mock_module, attr_name, attr_value)

        sys.modules[fullname] = mock_module
        return mock_module

    # 按顺序创建所有模块
    for module_name in sorted(modules_to_create.keys()):
        if module_name not in sys.modules:
            create_module(module_name, modules_to_create[module_name])


# 在导入任何 app 模块之前设置 Mock
setup_dolphin_mocks()


# ============================================================
# 2. pytest fixtures
# ============================================================


@pytest.fixture(scope="session", autouse=True)
def setup_test_environment():
    """设置测试环境 - 在所有测试开始前执行一次"""
    # 设置环境变量
    os.environ["ENVIRONMENT"] = "test"

    yield

    # 清理
    os.environ.pop("ENVIRONMENT", None)


@pytest.fixture(autouse=True)
def reset_global_state():
    """每个测试前重置全局状态

    这解决了测试隔离问题，确保每个测试都有干净的状态。
    """
    # 重置 MockVarOutput 的存储

    try:
        if "dolphin.core.context.var_output" in sys.modules:
            var_output_module = sys.modules["dolphin.core.context.var_output"]
            if hasattr(var_output_module, "VarOutput"):
                var_output_module.VarOutput.reset()
    except Exception:
        pass

    # 重置依赖管理的默认实例
    try:
        from app.common.dependencies import reset_default_instances

        reset_default_instances()
    except ImportError:
        pass

    yield

    # 测试后清理
    try:
        if "dolphin.core.context.var_output" in sys.modules:
            var_output_module = sys.modules["dolphin.core.context.var_output"]
            if hasattr(var_output_module, "VarOutput"):
                var_output_module.VarOutput.reset()
    except Exception:
        pass


@pytest.fixture
def mock_context_var_manager():
    """提供 Mock 的上下文变量管理器"""

    class SimpleMockContextVarManager:
        def __init__(self):
            self._storage = {}

        def get(self, key, default=None):
            return self._storage.get(key, default)

        def set(self, key, value):
            self._storage[key] = value

        def delete(self, key):
            self._storage.pop(key, None)

        def exists(self, key):
            return key in self._storage

        def get_all(self):
            return self._storage.copy()

    manager = SimpleMockContextVarManager()
    return manager


@pytest.fixture
def mock_exception_handler():
    """提供 Mock 的异常处理器"""

    class SimpleMockExceptionHandler:
        def create_model_exception(self, message: str) -> Exception:
            return Exception(f"[Model] {message}")

        def create_skill_exception(self, message: str) -> Exception:
            return Exception(f"[Skill] {message}")

        def create_dolphin_exception(self, message: str) -> Exception:
            return Exception(f"[Dolphin] {message}")

        def is_available(self) -> bool:
            return False

    return SimpleMockExceptionHandler()


@pytest.fixture
def mock_caller_info_provider():
    """提供 Mock 的调用者信息提供者"""

    class SimpleMockCallerInfoProvider:
        def get_caller_info(self):
            return ("test_file.py", 42)

    return SimpleMockCallerInfoProvider()


@pytest.fixture
def mock_environment_detector():
    """提供 Mock 的环境检测器"""

    class SimpleMockEnvironmentDetector:
        def is_in_pod(self) -> bool:
            return False

        def get_environment_type(self) -> str:
            return "test"

    return SimpleMockEnvironmentDetector()


# ============================================================
# 3. 测试配置
# ============================================================


def pytest_configure(config):
    """Pytest 配置钩子"""
    # 设置标记
    config.addinivalue_line(
        "markers", "slow: marks tests as slow (deselect with '-m \"not slow\"')"
    )
    config.addinivalue_line("markers", "integration: marks tests as integration tests")
    config.addinivalue_line("markers", "unit: marks tests as unit tests")

    # 过滤 Pydantic V2 迁移警告
    config.addinivalue_line(
        "filterwarnings", "ignore::pydantic.warnings.PydanticDeprecatedSince20"
    )
    # 过滤 AsyncMock 协程未等待的警告（来自未被使用的 mock）
    config.addinivalue_line(
        "filterwarnings",
        "ignore:coroutine 'AsyncMockMixin._execute_mock_call' was never awaited:RuntimeWarning",
    )


# ============================================================
# 4. 导入钩子 - 确保 dolphin 始终被 mock
# ============================================================


# 注意：在实际测试中，我们不需要替换 __import__，
# 因为 setup_dolphin_mocks() 已经在模块导入前被调用了
# 并且 test/unit/conftest.py 中已经使用了 meta_path 的方式来处理导入
