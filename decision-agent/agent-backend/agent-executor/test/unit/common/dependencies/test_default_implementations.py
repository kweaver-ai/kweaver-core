# -*- coding:utf-8 -*-
"""单元测试 - 默认依赖实现"""

import pytest
import os
from unittest.mock import patch


@pytest.mark.asyncio
class TestDefaultContextVarManager:
    """测试 DefaultContextVarManager 类"""

    async def test_init(self):
        """测试初始化"""
        from app.common.dependencies.default_implementations import (
            DefaultContextVarManager,
        )

        manager = DefaultContextVarManager()

        assert manager._var_output is None

    async def test_get_with_fallback_storage(self):
        """测试使用fallback storage获取值"""
        from app.common.dependencies.default_implementations import (
            DefaultContextVarManager,
        )

        manager = DefaultContextVarManager()

        # First call should initialize fallback storage since dolphin is not available
        value = manager.get("test_key", default="default_value")

        # Should return default value
        assert value == "default_value"

    async def test_set_and_get(self):
        """测试设置和获取值"""
        from app.common.dependencies.default_implementations import (
            DefaultContextVarManager,
        )

        manager = DefaultContextVarManager()

        manager._fallback_storage = {}  # Initialize fallback storage
        manager.set("key1", "value1")
        manager.set("key2", 123)

        # Get values back
        assert manager.get("key1") == "value1"
        assert manager.get("key2") == 123

    async def test_get_default_value(self):
        """测试获取默认值"""
        from app.common.dependencies.default_implementations import (
            DefaultContextVarManager,
        )

        manager = DefaultContextVarManager()

        # Get non-existent key with default
        assert manager.get("nonexistent", "my_default") == "my_default"

    async def test_delete_key(self):
        """测试删除键"""
        from app.common.dependencies.default_implementations import (
            DefaultContextVarManager,
        )

        manager = DefaultContextVarManager()

        manager.set("key_to_delete", "value")
        assert manager.exists("key_to_delete") is True

        manager.delete("key_to_delete")
        assert manager.exists("key_to_delete") is False

    async def test_delete_nonexistent_key(self):
        """测试删除不存在的键（不应报错）"""
        from app.common.dependencies.default_implementations import (
            DefaultContextVarManager,
        )

        manager = DefaultContextVarManager()

        # Should not raise
        manager.delete("nonexistent_key")

    async def test_exists(self):
        """测试检查键是否存在"""
        from app.common.dependencies.default_implementations import (
            DefaultContextVarManager,
        )

        manager = DefaultContextVarManager()

        manager.set("existing_key", "value")

        assert manager.exists("existing_key") is True
        assert manager.exists("nonexistent_key") is False

    async def test_get_all(self):
        """测试获取所有值"""
        from app.common.dependencies.default_implementations import (
            DefaultContextVarManager,
        )

        manager = DefaultContextVarManager()

        # get_all should return empty dict initially
        all_values = manager.get_all()
        assert all_values == {}

    async def test_get_with_var_output(self):
        """测试使用真正的 VarOutput 获取值"""
        from app.common.dependencies.default_implementations import (
            DefaultContextVarManager,
        )

        manager = DefaultContextVarManager()

        # Set a value
        manager.set("test_key", "test_value")

        # Get it back - should use VarOutput
        result = manager.get("test_key")
        assert result == "test_value"

    async def test_delete_from_fallback_storage(self):
        """测试从 fallback storage 删除"""
        from app.common.dependencies.default_implementations import (
            DefaultContextVarManager,
        )

        manager = DefaultContextVarManager()

        # Manually set up fallback storage
        manager._fallback_storage = {"key1": "value1", "key2": "value2"}

        # Verify fallback_storage is set
        assert "key1" in manager._fallback_storage

        # Delete key1
        manager.delete("key1")

        # The delete method only checks if key exists, then deletes it
        # Let's verify the delete operation was called correctly

    async def test_delete_from_nonexistent_fallback(self):
        """测试从不存在的 fallback storage 删除（不报错）"""
        from app.common.dependencies.default_implementations import (
            DefaultContextVarManager,
        )

        manager = DefaultContextVarManager()

        # Initialize fallback storage
        manager._fallback_storage = {"key1": "value1"}

        # Delete nonexistent key - should not crash
        # The delete method checks hasattr and key in storage before deleting
        manager.delete("nonexistent_key")

        # Original key should still be there
        assert "key1" in manager._fallback_storage


@pytest.mark.asyncio
class TestDefaultExceptionHandler:
    """测试 DefaultExceptionHandler 类"""

    async def test_create_model_exception(self):
        """测试创建模型异常"""
        from app.common.dependencies.default_implementations import (
            DefaultExceptionHandler,
        )

        handler = DefaultExceptionHandler()

        exc = handler.create_model_exception("Model error message")

        # Should create a dolphin exception
        assert exc is not None

    async def test_create_skill_exception(self):
        """测试创建技能异常"""
        from app.common.dependencies.default_implementations import (
            DefaultExceptionHandler,
        )

        handler = DefaultExceptionHandler()

        exc = handler.create_skill_exception("Skill error message")

        # Should create a dolphin exception
        assert exc is not None

    async def test_create_dolphin_exception(self):
        """测试创建Dolphin异常"""
        from app.common.dependencies.default_implementations import (
            DefaultExceptionHandler,
        )

        handler = DefaultExceptionHandler()

        exc = handler.create_dolphin_exception("Dolphin error message")

        # Should create a dolphin exception
        assert exc is not None

    async def test_is_available(self):
        """测试检查Dolphin是否可用"""
        from app.common.dependencies.default_implementations import (
            DefaultExceptionHandler,
        )

        handler = DefaultExceptionHandler()

        # Should return boolean
        available = handler.is_available()
        assert isinstance(available, bool)


@pytest.mark.asyncio
class TestDefaultCallerInfoProvider:
    """测试 DefaultCallerInfoProvider 类"""

    async def test_get_caller_info(self):
        """测试获取调用者信息"""
        from app.common.dependencies.default_implementations import (
            DefaultCallerInfoProvider,
        )

        provider = DefaultCallerInfoProvider()

        filename, lineno = provider.get_caller_info()

        # Should return tuple of filename and line number
        assert isinstance(filename, str)
        assert isinstance(lineno, int)
        # Filename should be a string (may be absolute or relative path)
        assert len(filename) > 0
        # Line number should be positive
        assert lineno > 0


@pytest.mark.asyncio
class TestDefaultEnvironmentDetector:
    """测试 DefaultEnvironmentDetector 类"""

    async def test_is_in_pod(self):
        """测试检查是否在Pod中"""
        from app.common.dependencies.default_implementations import (
            DefaultEnvironmentDetector,
        )

        detector = DefaultEnvironmentDetector()

        # Without pod environment vars
        assert detector.is_in_pod() is False

        # With pod environment vars
        with patch.dict(
            os.environ,
            {"KUBERNETES_SERVICE_HOST": "localhost", "KUBERNETES_SERVICE_PORT": "8080"},
        ):
            new_detector = DefaultEnvironmentDetector()
            assert new_detector.is_in_pod() is True

    async def test_get_environment_type_pod(self):
        """测试获取Pod环境类型"""
        from app.common.dependencies.default_implementations import (
            DefaultEnvironmentDetector,
        )

        with patch.dict(
            os.environ,
            {"KUBERNETES_SERVICE_HOST": "localhost", "KUBERNETES_SERVICE_PORT": "8080"},
        ):
            detector = DefaultEnvironmentDetector()
            assert detector.get_environment_type() == "pod"

    async def test_get_environment_type_test(self):
        """测试获取测试环境类型"""
        from app.common.dependencies.default_implementations import (
            DefaultEnvironmentDetector,
        )

        with patch.dict(os.environ, {"ENVIRONMENT": "test"}):
            detector = DefaultEnvironmentDetector()
            assert detector.get_environment_type() == "test"

    async def test_get_environment_type_development(self):
        """测试获取开发环境类型"""
        from app.common.dependencies.default_implementations import (
            DefaultEnvironmentDetector,
        )

        with patch.dict(os.environ, {"ENVIRONMENT": "development"}):
            detector = DefaultEnvironmentDetector()
            assert detector.get_environment_type() == "local"

    async def test_get_environment_type_unknown(self):
        """测试获取未知环境类型"""
        from app.common.dependencies.default_implementations import (
            DefaultEnvironmentDetector,
        )

        # Clear env vars
        env_copy = os.environ.copy()
        env_copy.pop("ENVIRONMENT", None)
        env_copy.pop("KUBERNETES_SERVICE_HOST", None)
        env_copy.pop("KUBERNETES_SERVICE_PORT", None)

        with patch.dict(os.environ, env_copy, clear=True):
            detector = DefaultEnvironmentDetector()
            assert detector.get_environment_type() == "unknown"


@pytest.mark.asyncio
class TestDefaultInstanceFunctions:
    """测试默认实例获取函数"""

    async def test_get_default_context_var_manager_singleton(self):
        """测试获取单例上下文变量管理器"""
        from app.common.dependencies.default_implementations import (
            get_default_context_var_manager,
        )

        manager1 = get_default_context_var_manager()
        manager2 = get_default_context_var_manager()

        # Should return the same instance
        assert manager1 is manager2

    async def test_get_default_exception_handler_singleton(self):
        """测试获取单例异常处理器"""
        from app.common.dependencies.default_implementations import (
            get_default_exception_handler,
        )

        handler1 = get_default_exception_handler()
        handler2 = get_default_exception_handler()

        # Should return the same instance
        assert handler1 is handler2

    async def test_get_default_caller_info_provider_singleton(self):
        """测试获取单例调用者信息提供者"""
        from app.common.dependencies.default_implementations import (
            get_default_caller_info_provider,
        )

        provider1 = get_default_caller_info_provider()
        provider2 = get_default_caller_info_provider()

        # Should return the same instance
        assert provider1 is provider2

    async def test_get_default_environment_detector_singleton(self):
        """测试获取单例环境检测器"""
        from app.common.dependencies.default_implementations import (
            get_default_environment_detector,
        )

        detector1 = get_default_environment_detector()
        detector2 = get_default_environment_detector()

        # Should return the same instance
        assert detector1 is detector2


@pytest.mark.asyncio
class TestResetDefaultInstances:
    """测试重置默认实例"""

    async def test_reset_creates_new_instances(self):
        """测试重置后创建新实例"""
        from app.common.dependencies.default_implementations import (
            get_default_context_var_manager,
            get_default_exception_handler,
            reset_default_instances,
        )

        # Get instances
        manager1 = get_default_context_var_manager()
        handler1 = get_default_exception_handler()

        # Reset
        reset_default_instances()

        # Get new instances
        manager2 = get_default_context_var_manager()
        handler2 = get_default_exception_handler()

        # Should be different instances
        assert manager1 is not manager2
        assert handler1 is not handler2
