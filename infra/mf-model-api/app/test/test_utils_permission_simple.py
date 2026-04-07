"""测试 utils/permission_manager.py 的基本功能"""
import pytest
from unittest.mock import AsyncMock, Mock, patch
from app.utils.permission_manager import PermissionManager, permission_manager


class TestPermissionManagerSimple:
    """测试权限管理器的基本功能"""

    def test_singleton_instance(self):
        """测试单例实例存在"""
        assert permission_manager is not None
        assert isinstance(permission_manager, PermissionManager)

    def test_instance_creation(self):
        """测试可以创建实例"""
        manager = PermissionManager()
        assert manager is not None

    def test_has_base_url(self):
        """测试实例有base_url属性"""
        manager = PermissionManager()
        assert hasattr(manager, 'base_url')
        assert isinstance(manager.base_url, str)

    def test_has_auth_url(self):
        """测试实例有auth_url属性"""
        manager = PermissionManager()
        assert hasattr(manager, 'auth_url')
        assert isinstance(manager.auth_url, str)

    def test_has_session_attribute(self):
        """测试实例有session属性"""
        manager = PermissionManager()
        assert hasattr(manager, 'session')

    @pytest.mark.asyncio
    async def test_get_session_method_exists(self):
        """测试get_session方法存在"""
        manager = PermissionManager()
        assert hasattr(manager, 'get_session')
        assert callable(manager.get_session)

    @pytest.mark.asyncio
    async def test_add_permission_method_exists(self):
        """测试add_permission方法存在"""
        manager = PermissionManager()
        assert hasattr(manager, 'add_permission')
        assert callable(manager.add_permission)

    @pytest.mark.asyncio
    async def test_check_single_permission_method_exists(self):
        """测试check_single_permission方法存在"""
        manager = PermissionManager()
        assert hasattr(manager, 'check_single_permission')
        assert callable(manager.check_single_permission)

    @pytest.mark.asyncio
    async def test_get_permission_ids_method_exists(self):
        """测试get_permission_ids方法存在"""
        manager = PermissionManager()
        assert hasattr(manager, 'get_permission_ids')
        assert callable(manager.get_permission_ids)

    @pytest.mark.asyncio
    async def test_delete_permission_method_exists(self):
        """测试delete_permission方法存在"""
        manager = PermissionManager()
        assert hasattr(manager, 'delete_permission')
        assert callable(manager.delete_permission)

    @pytest.mark.asyncio
    async def test_close_method_exists(self):
        """测试close方法存在"""
        manager = PermissionManager()
        assert hasattr(manager, 'close')
        assert callable(manager.close)

    @pytest.mark.asyncio
    async def test_add_permission_admin_user_shortcut(self):
        """测试admin用户的快捷处理"""
        manager = PermissionManager()
        result = await manager.add_permission(
            user_id="266c6a42-6131-4d62-8f39-853e7093701c",
            resource_id="test",
            resource_name="test",
            resource_type="test",
            user_name="admin",
            role="admin"
        )
        assert result is True

