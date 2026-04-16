"""Unit test configuration - mock missing external dependencies"""

import sys
from unittest.mock import MagicMock
from types import ModuleType


def create_dolphin_module(fullname):
    """Create a mock module with proper structure"""
    mock_module = ModuleType(fullname)
    mock_module.__name__ = fullname
    mock_module.__file__ = f"{fullname.replace('.', '/')}/__init__.py"
    mock_module.__loader__ = None

    if "." in fullname:
        mock_module.__package__ = fullname.rsplit(".", 1)[0]
        mock_module.__path__ = []  # Make it a package
    else:
        mock_module.__package__ = fullname
        mock_module.__path__ = []  # Make it a package

    # Add mock submodules as needed
    if fullname == "dolphin.core":
        mock_module.ModelException = type("ModelException", (Exception,), {})
        mock_module.SkillException = type("SkillException", (Exception,), {})
        mock_module.DolphinException = type("DolphinException", (Exception,), {})

    if fullname == "dolphin.core.common.exceptions":
        mock_module.ModelException = type("ModelException", (Exception,), {})
        mock_module.SkillException = type("SkillException", (Exception,), {})
        mock_module.DolphinException = type("DolphinException", (Exception,), {})

    return mock_module


class MockRedis:
    async def get(self, key):
        return None

    async def set(self, key, value, ex=None):
        return True

    async def delete(self, *keys):
        return 0

    async def exists(self, *keys):
        return 0

    async def expire(self, key, time):
        return True

    async def close(self):
        pass

    async def ping(self):
        return True

    async def setnx(self, key, value):
        return True

    async def getset(self, key, value):
        return None


class MockConnectionPool:
    def __init__(self, **kwargs):
        pass

    async def disconnect(self):
        pass


class MockSentinel:
    def __init__(self, *args, **kwargs):
        pass

    def master_for(self, *args, **kwargs):
        client = MockRedis()
        client.connection_pool = MockConnectionPool()
        return client

    def slave_for(self, *args, **kwargs):
        client = MockRedis()
        client.connection_pool = MockConnectionPool()
        return client


class DolphinModuleFinder:
    """A custom module finder that creates mock modules for dolphin.* imports"""

    def find_spec(self, fullname, path, target=None):
        if (
            fullname.startswith("dolphin.")
            or fullname == "dolphin"
            or fullname == "limiter"
            or fullname.startswith("limiter.")
            or fullname.startswith("redis.")
            or fullname == "redis"
        ):
            # Ensure parent modules are created first
            if "." in fullname:
                parent_name = fullname.rsplit(".", 1)[0]
                if parent_name not in sys.modules:
                    # Import parent to trigger its creation
                    try:
                        __import__(parent_name)
                    except ImportError:
                        pass

            # Create a mock spec for the module
            from importlib.machinery import ModuleSpec

            spec = ModuleSpec(fullname, self, origin="mock")
            spec.loader = self
            spec.submodule_search_locations = []  # Make it a package
            return spec
        return None

    def create_module(self, spec):
        mock_module = create_dolphin_module(spec.name)

        # Special handling for Tool base class
        if spec.name == "dolphin.core.utils.tools":
            # Create a real class that mimics Tool
            class MockTool:
                pass

            mock_module.Tool = MockTool

        # Special handling for Context
        if spec.name == "dolphin.core.context.context":

            class MockContext:
                pass

            mock_module.Context = MockContext

        # Special handling for VarOutput.is_serialized_dict
        if spec.name == "dolphin.core.context.var_output":
            # 创建一个真正的 VarOutput 类，包含正确的 is_serialized_dict 实现
            class MockVarOutput:
                @staticmethod
                def is_serialized_dict(var):
                    """检查是否为序列化的字典 - 与原始实现一致"""
                    return isinstance(var, dict) and var.get("__type__") == "VarOutput"

            mock_module.VarOutput = MockVarOutput

        # Special handling for ResumeHandle class
        if spec.name == "dolphin.core.coroutine.resume_handle":
            # Create a real class that mimics ResumeHandle
            class MockResumeHandle:
                def __init__(
                    self,
                    frame_id,
                    snapshot_id,
                    resume_token,
                    interrupt_type,
                    current_block,
                    restart_block,
                ):
                    self.frame_id = frame_id
                    self.snapshot_id = snapshot_id
                    self.resume_token = resume_token
                    self.interrupt_type = interrupt_type
                    self.current_block = current_block
                    self.restart_block = restart_block

            mock_module.ResumeHandle = MockResumeHandle

        # Special handling for exceptions
        if spec.name == "dolphin.core.common.exceptions":
            mock_module.ModelException = type("ModelException", (Exception,), {})
            mock_module.SkillException = type("SkillException", (Exception,), {})
            mock_module.DolphinException = type("DolphinException", (Exception,), {})

        # Special handling for limiter module
        if spec.name == "limiter":
            # Create a mock Limiter class
            class MockLimiter:
                def __init__(
                    self,
                    app=None,
                    key_func=None,
                    default_limits=None,
                    rate=None,
                    capacity=None,
                    consume=None,
                ):
                    self.app = app
                    self.key_func = key_func
                    self.default_limits = default_limits
                    self.rate = rate
                    self.capacity = capacity
                    self.consume = consume

                def limit(self, limit_string):
                    def decorator(f):
                        return f

                    return decorator

                def init_app(self, app):
                    pass

            mock_module.Limiter = MockLimiter
            mock_module.__all__ = ["Limiter"]

        # Special handling for redis module
        if spec.name == "redis.asyncio":
            mock_module.Redis = MockRedis
            mock_module.ConnectionPool = MockConnectionPool
            mock_module.from_url = lambda url, **kwargs: MockRedis()

        if spec.name == "redis.asyncio.connection":
            mock_module.ConnectionPool = MockConnectionPool

        if spec.name == "redis.asyncio.sentinel":
            mock_module.Sentinel = MockSentinel

        if spec.name == "redis":
            # Create base redis module
            mock_module.asyncio = MagicMock()

        return mock_module

    def exec_module(self, module):
        pass


# Install the custom finder
sys.meta_path.insert(0, DolphinModuleFinder())


# Setup Config.app mock for tests that need it
import logging

mock_app_config = MagicMock()
mock_app_config.get_stdlib_log_level = MagicMock(return_value=logging.INFO)
mock_app_config.enable_system_log = "false"
mock_app_config.debug = False
mock_app_config.rps_limit = 100  # Default rps_limit for limiter
mock_app_config.host_prefix = ""  # Default host prefix
mock_app_config.host_prefix_v2 = ""  # Default host prefix v2 (must not end with /)

# Import Config and set up the mock
from app.common.config import Config

Config.app = mock_app_config
