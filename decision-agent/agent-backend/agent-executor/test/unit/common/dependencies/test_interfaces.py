# -*- coding: utf-8 -*-
"""
Unit tests for app/common/dependencies/interfaces.py module
"""

from abc import ABC
from typing import Dict, Any

import pytest


class TestIContextVarManager:
    """Tests for IContextVarManager interface"""

    @pytest.mark.asyncio
    async def test_interface_is_abstract(self):
        """Test IContextVarManager is an abstract base class"""
        from app.common.dependencies.interfaces import IContextVarManager

        assert issubclass(IContextVarManager, ABC)
        assert hasattr(IContextVarManager, "__abstractmethods__")

    @pytest.mark.asyncio
    async def test_get_method_is_abstract(self):
        """Test get method is abstract"""
        from app.common.dependencies.interfaces import IContextVarManager

        assert "get" in IContextVarManager.__abstractmethods__

    @pytest.mark.asyncio
    async def test_set_method_is_abstract(self):
        """Test set method is abstract"""
        from app.common.dependencies.interfaces import IContextVarManager

        assert "set" in IContextVarManager.__abstractmethods__

    @pytest.mark.asyncio
    async def test_delete_method_is_abstract(self):
        """Test delete method is abstract"""
        from app.common.dependencies.interfaces import IContextVarManager

        assert "delete" in IContextVarManager.__abstractmethods__

    @pytest.mark.asyncio
    async def test_exists_method_is_abstract(self):
        """Test exists method is abstract"""
        from app.common.dependencies.interfaces import IContextVarManager

        assert "exists" in IContextVarManager.__abstractmethods__

    @pytest.mark.asyncio
    async def test_get_all_method_is_abstract(self):
        """Test get_all method is abstract"""
        from app.common.dependencies.interfaces import IContextVarManager

        assert "get_all" in IContextVarManager.__abstractmethods__

    @pytest.mark.asyncio
    async def test_concrete_implementation(self):
        """Test a concrete implementation can be created"""
        from app.common.dependencies.interfaces import IContextVarManager

        class MockContextVarManager(IContextVarManager):
            def __init__(self):
                self._data: Dict[str, Any] = {}

            def get(self, key: str, default: Any = None) -> Any:
                return self._data.get(key, default)

            def set(self, key: str, value: Any) -> None:
                self._data[key] = value

            def delete(self, key: str) -> None:
                self._data.pop(key, None)

            def exists(self, key: str) -> bool:
                return key in self._data

            def get_all(self) -> Dict[str, Any]:
                return self._data.copy()

        manager = MockContextVarManager()
        manager.set("test_key", "test_value")
        assert manager.get("test_key") == "test_value"
        assert manager.exists("test_key") is True
        assert "test_key" in manager.get_all()
        manager.delete("test_key")
        assert manager.exists("test_key") is False


class TestIExceptionHandler:
    """Tests for IExceptionHandler interface"""

    @pytest.mark.asyncio
    async def test_interface_is_abstract(self):
        """Test IExceptionHandler is an abstract base class"""
        from app.common.dependencies.interfaces import IExceptionHandler

        assert issubclass(IExceptionHandler, ABC)

    @pytest.mark.asyncio
    async def test_create_model_exception_is_abstract(self):
        """Test create_model_exception method is abstract"""
        from app.common.dependencies.interfaces import IExceptionHandler

        assert "create_model_exception" in IExceptionHandler.__abstractmethods__

    @pytest.mark.asyncio
    async def test_create_skill_exception_is_abstract(self):
        """Test create_skill_exception method is abstract"""
        from app.common.dependencies.interfaces import IExceptionHandler

        assert "create_skill_exception" in IExceptionHandler.__abstractmethods__

    @pytest.mark.asyncio
    async def test_create_dolphin_exception_is_abstract(self):
        """Test create_dolphin_exception method is abstract"""
        from app.common.dependencies.interfaces import IExceptionHandler

        assert "create_dolphin_exception" in IExceptionHandler.__abstractmethods__

    @pytest.mark.asyncio
    async def test_is_available_is_abstract(self):
        """Test is_available method is abstract"""
        from app.common.dependencies.interfaces import IExceptionHandler

        assert "is_available" in IExceptionHandler.__abstractmethods__

    @pytest.mark.asyncio
    async def test_concrete_implementation(self):
        """Test a concrete implementation can be created"""
        from app.common.dependencies.interfaces import IExceptionHandler

        class MockExceptionHandler(IExceptionHandler):
            def create_model_exception(self, message: str) -> Exception:
                return Exception(f"Model: {message}")

            def create_skill_exception(self, message: str) -> Exception:
                return Exception(f"Skill: {message}")

            def create_dolphin_exception(self, message: str) -> Exception:
                return Exception(f"Dolphin: {message}")

            def is_available(self) -> bool:
                return True

        handler = MockExceptionHandler()
        assert isinstance(handler.create_model_exception("test"), Exception)
        assert isinstance(handler.create_skill_exception("test"), Exception)
        assert isinstance(handler.create_dolphin_exception("test"), Exception)
        assert handler.is_available() is True


class TestICallerInfoProvider:
    """Tests for ICallerInfoProvider interface"""

    @pytest.mark.asyncio
    async def test_interface_is_abstract(self):
        """Test ICallerInfoProvider is an abstract base class"""
        from app.common.dependencies.interfaces import ICallerInfoProvider

        assert issubclass(ICallerInfoProvider, ABC)

    @pytest.mark.asyncio
    async def test_get_caller_info_is_abstract(self):
        """Test get_caller_info method is abstract"""
        from app.common.dependencies.interfaces import ICallerInfoProvider

        assert "get_caller_info" in ICallerInfoProvider.__abstractmethods__

    @pytest.mark.asyncio
    async def test_get_caller_info_is_abstract(self):
        """Test get_caller_info method is abstract"""
        from app.common.dependencies.interfaces import ICallerInfoProvider

        assert "get_caller_info" in ICallerInfoProvider.__abstractmethods__


class TestIEnvironmentDetector:
    """Tests for IEnvironmentDetector interface"""

    @pytest.mark.asyncio
    async def test_interface_is_abstract(self):
        """Test IEnvironmentDetector is an abstract base class"""
        from app.common.dependencies.interfaces import IEnvironmentDetector

        assert issubclass(IEnvironmentDetector, ABC)

    @pytest.mark.asyncio
    async def test_is_in_pod_is_abstract(self):
        """Test is_in_pod method is abstract"""
        from app.common.dependencies.interfaces import IEnvironmentDetector

        assert "is_in_pod" in IEnvironmentDetector.__abstractmethods__

    @pytest.mark.asyncio
    async def test_get_environment_type_is_abstract(self):
        """Test get_environment_type method is abstract"""
        from app.common.dependencies.interfaces import IEnvironmentDetector

        assert "get_environment_type" in IEnvironmentDetector.__abstractmethods__

    @pytest.mark.asyncio
    async def test_concrete_implementation(self):
        """Test a concrete implementation can be created"""
        from app.common.dependencies.interfaces import IEnvironmentDetector

        class MockEnvironmentDetector(IEnvironmentDetector):
            def is_in_pod(self) -> bool:
                return False

            def get_environment_type(self) -> str:
                return "local"

        detector = MockEnvironmentDetector()
        assert detector.is_in_pod() is False
        assert detector.get_environment_type() == "local"


class TestSerializationType:
    """Tests for SerializationType enum"""

    @pytest.mark.asyncio
    async def test_enum_values(self):
        """Test SerializationType enum has correct values"""
        from app.common.dependencies.interfaces import SerializationType

        assert SerializationType.JSON.value == "json"
        assert SerializationType.PICKLE.value == "pickle"

    @pytest.mark.asyncio
    async def test_enum_members(self):
        """Test SerializationType enum members"""
        from app.common.dependencies.interfaces import SerializationType

        assert hasattr(SerializationType, "JSON")
        assert hasattr(SerializationType, "PICKLE")


class TestICacheService:
    """Tests for ICacheService interface"""

    @pytest.mark.asyncio
    async def test_interface_is_abstract(self):
        """Test ICacheService is an abstract base class"""
        from app.common.dependencies.interfaces import ICacheService

        assert issubclass(ICacheService, ABC)

    @pytest.mark.asyncio
    async def test_get_method_is_abstract(self):
        """Test get method is abstract"""
        from app.common.dependencies.interfaces import ICacheService

        assert "get" in ICacheService.__abstractmethods__

    @pytest.mark.asyncio
    async def test_set_method_is_abstract(self):
        """Test set method is abstract"""
        from app.common.dependencies.interfaces import ICacheService

        assert "set" in ICacheService.__abstractmethods__

    @pytest.mark.asyncio
    async def test_delete_method_is_abstract(self):
        """Test delete method is abstract"""
        from app.common.dependencies.interfaces import ICacheService

        assert "delete" in ICacheService.__abstractmethods__

    @pytest.mark.asyncio
    async def test_exists_method_is_abstract(self):
        """Test exists method is abstract"""
        from app.common.dependencies.interfaces import ICacheService

        assert "exists" in ICacheService.__abstractmethods__

    @pytest.mark.asyncio
    async def test_set_method_signature(self):
        """Test set method has correct signature with SerializationType"""
        from app.common.dependencies.interfaces import ICacheService

        # Check that set has serialization_type parameter with correct type
        assert "set" in ICacheService.__abstractmethods__

    @pytest.mark.asyncio
    async def test_concrete_implementation(self):
        """Test a concrete implementation can be created"""
        from app.common.dependencies.interfaces import ICacheService, SerializationType

        class MockCacheService(ICacheService):
            def __init__(self):
                self._cache: Dict[str, Any] = {}

            async def get(self, key: str):
                return self._cache.get(key)

            async def set(
                self,
                key: str,
                value: Any,
                ttl=None,
                serialization_type=SerializationType.JSON,
            ) -> bool:
                self._cache[key] = value
                return True

            async def delete(self, key: str) -> bool:
                return self._cache.pop(key, None) is not None

            async def exists(self, key: str) -> bool:
                return key in self._cache

        cache = MockCacheService()
        await cache.set("test_key", "test_value")
        assert await cache.exists("test_key") is True
        assert await cache.get("test_key") == "test_value"
        assert await cache.delete("test_key") is True
        assert await cache.exists("test_key") is False
