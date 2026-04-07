"""测试 utils/common.py 模块"""
import pytest
import os
from unittest.mock import Mock, patch
from app.utils.common import (
    GetCallerInfo,
    IsInPod,
    GetFailureThreshold,
    SetFailureThreshold,
    GetRecoveryTimeout,
    SetRecoveryTimeout,
    get_user_info,
    validate_required_params
)


class TestCommonFunctions:
    """测试common模块的公共函数"""

    def test_get_caller_info(self):
        """测试获取调用者信息"""
        filename, lineno = GetCallerInfo()
        # 验证返回值类型
        assert isinstance(filename, str)
        assert isinstance(lineno, int)
        assert lineno > 0

    def test_is_in_pod_true(self):
        """测试在Pod环境中"""
        with patch.dict(os.environ, {
            'KUBERNETES_SERVICE_HOST': 'localhost',
            'KUBERNETES_SERVICE_PORT': '8080'
        }):
            assert IsInPod() is True

    def test_is_in_pod_false(self):
        """测试不在Pod环境中"""
        with patch.dict(os.environ, {}, clear=True):
            assert IsInPod() is False

    def test_is_in_pod_partial_env(self):
        """测试只有部分环境变量时"""
        with patch.dict(os.environ, {'KUBERNETES_SERVICE_HOST': 'localhost'}, clear=True):
            assert IsInPod() is False

        with patch.dict(os.environ, {'KUBERNETES_SERVICE_PORT': '8080'}, clear=True):
            assert IsInPod() is False

    def test_failure_threshold_get_default(self):
        """测试获取默认失败阈值"""
        threshold = GetFailureThreshold()
        assert isinstance(threshold, int)
        assert threshold >= 0

    def test_failure_threshold_set_and_get(self):
        """测试设置和获取失败阈值"""
        original = GetFailureThreshold()
        try:
            SetFailureThreshold(20)
            assert GetFailureThreshold() == 20

            SetFailureThreshold(5)
            assert GetFailureThreshold() == 5
        finally:
            # 恢复原始值
            SetFailureThreshold(original)

    def test_recovery_timeout_get_default(self):
        """测试获取默认恢复超时"""
        timeout = GetRecoveryTimeout()
        assert isinstance(timeout, int)
        assert timeout >= 0

    def test_recovery_timeout_set_and_get(self):
        """测试设置和获取恢复超时"""
        original = GetRecoveryTimeout()
        try:
            SetRecoveryTimeout(10)
            assert GetRecoveryTimeout() == 10

            SetRecoveryTimeout(30)
            assert GetRecoveryTimeout() == 30
        finally:
            # 恢复原始值
            SetRecoveryTimeout(original)

    @pytest.mark.asyncio
    async def test_get_user_info_all_headers(self):
        """测试从请求中获取用户信息 - 所有header都存在"""
        request = Mock()
        request.headers = {
            'x-account-id': 'user123',
            'x-account-type': 'admin',
            'accept-language': 'en-US'
        }

        userId, language, role = await get_user_info(request)

        assert userId == 'user123'
        assert language == 'en-US'
        assert role == 'admin'

    @pytest.mark.asyncio
    async def test_get_user_info_missing_headers(self):
        """测试从请求中获取用户信息 - 缺少header"""
        request = Mock()
        request.headers = {}

        userId, language, role = await get_user_info(request)

        assert userId == ""
        assert language == "zh-CN"  # 默认值
        assert role == ""

    @pytest.mark.asyncio
    async def test_get_user_info_partial_headers(self):
        """测试从请求中获取用户信息 - 部分header"""
        request = Mock()
        request.headers = {
            'x-account-id': 'user456'
        }

        userId, language, role = await get_user_info(request)

        assert userId == 'user456'
        assert language == "zh-CN"  # 默认值
        assert role == ""

    @pytest.mark.asyncio
    async def test_validate_required_params_all_present(self):
        """测试参数验证 - 所有必需参数都存在"""
        params_dict = {
            'name': 'test',
            'age': 25,
            'email': 'test@example.com'
        }
        required_params = ['name', 'age']

        missing = await validate_required_params(params_dict, required_params)

        assert missing == []

    @pytest.mark.asyncio
    async def test_validate_required_params_some_missing(self):
        """测试参数验证 - 部分参数缺失"""
        params_dict = {
            'name': 'test'
        }
        required_params = ['name', 'age', 'email']

        missing = await validate_required_params(params_dict, required_params)

        assert set(missing) == {'age', 'email'}

    @pytest.mark.asyncio
    async def test_validate_required_params_all_missing(self):
        """测试参数验证 - 所有参数都缺失"""
        params_dict = {}
        required_params = ['name', 'age', 'email']

        missing = await validate_required_params(params_dict, required_params)

        assert set(missing) == {'name', 'age', 'email'}

    @pytest.mark.asyncio
    async def test_validate_required_params_empty_required(self):
        """测试参数验证 - 没有必需参数"""
        params_dict = {
            'name': 'test',
            'age': 25
        }
        required_params = []

        missing = await validate_required_params(params_dict, required_params)

        assert missing == []

    @pytest.mark.asyncio
    async def test_get_user_info_with_kwargs(self):
        """测试get_user_info接受额外参数"""
        request = Mock()
        request.headers = {
            'x-account-id': 'user789',
            'x-account-type': 'user',
            'accept-language': 'zh-CN'
        }

        # 测试传入额外的kwargs不会导致错误
        userId, language, role = await get_user_info(request, extra_param="test")

        assert userId == 'user789'
        assert language == 'zh-CN'
        assert role == 'user'

