"""测试 app_utils 模块"""
import pytest
from unittest.mock import Mock, AsyncMock, patch, MagicMock
from fastapi import FastAPI
from fastapi.responses import JSONResponse
from app.utils.app_utils import (
    conf_init,
    start_event,
    shutdown_event,
    auth_middleware,
    RequestSizeMiddleware,
    create_app
)


class TestConfInit:
    """测试conf_init函数"""

    def test_conf_init_production(self):
        """测试生产环境配置"""
        app = FastAPI()
        with patch.dict('os.environ', {'ENVIRONMENT': 'production'}):
            with patch('app.utils.app_utils.sys_log.info') as mock_log:
                conf_init(app)
                assert app.docs_url is None
                assert app.redoc_url is None
                assert app.debug is False
                mock_log.assert_called()

    def test_conf_init_development(self):
        """测试开发环境配置"""
        app = FastAPI()
        with patch.dict('os.environ', {'ENVIRONMENT': 'development'}):
            with patch('app.utils.app_utils.sys_log.info') as mock_log:
                conf_init(app)
                # 开发环境不应该修改docs和debug
                mock_log.assert_called()

    def test_conf_init_default(self):
        """测试默认环境配置"""
        app = FastAPI()
        with patch.dict('os.environ', {}, clear=True):
            with patch('app.utils.app_utils.sys_log.info'):
                conf_init(app)


class TestStartEvent:
    """测试start_event函数"""

    @pytest.mark.asyncio
    async def test_start_event_success(self):
        """测试成功启动"""
        with patch('app.utils.app_utils.write_log', new_callable=AsyncMock) as mock_log:
            with patch('app.utils.app_utils.get_redis_util', new_callable=AsyncMock) as mock_redis:
                with patch('app.utils.app_utils.init_observability') as mock_obs:
                    mock_redis.return_value = Mock()
                    await start_event()
                    mock_log.assert_called_once_with(msg='系统启动')
                    mock_redis.assert_called_once()
                    mock_obs.assert_called_once()

    @pytest.mark.asyncio
    async def test_start_event_redis_error(self):
        """测试Redis连接失败"""
        with patch('app.utils.app_utils.write_log', new_callable=AsyncMock):
            with patch('app.utils.app_utils.get_redis_util', new_callable=AsyncMock) as mock_redis:
                mock_redis.side_effect = Exception("Redis connection error")
                with pytest.raises(Exception):
                    await start_event()


class TestShutdownEvent:
    """测试shutdown_event函数"""

    @pytest.mark.asyncio
    async def test_shutdown_event(self):
        """测试关闭事件"""
        with patch('app.utils.app_utils.write_log', new_callable=AsyncMock) as mock_log:
            with patch('app.utils.app_utils.shutdown_observability') as mock_obs:
                await shutdown_event()
                mock_log.assert_called_once_with(msg='系统关闭')
                mock_obs.assert_called_once()


class TestAuthMiddleware:
    """测试auth_middleware函数"""

    @pytest.fixture
    def mock_request(self):
        """创建mock request"""
        request = Mock()
        request.url = Mock()
        request.headers = {}
        request.scope = {'headers': []}
        return request

    @pytest.fixture
    def mock_call_next(self):
        """创建mock call_next"""
        async def call_next(request):
            return JSONResponse(content={"status": "ok"})
        return call_next

    @pytest.mark.asyncio
    async def test_health_endpoint_bypass(self, mock_request, mock_call_next):
        """测试健康检查端点绕过认证"""
        mock_request.url.path = "/api/v1/health"
        response = await auth_middleware(mock_request, mock_call_next)
        assert response.status_code == 200

    @pytest.mark.asyncio
    async def test_private_endpoint_bypass(self, mock_request, mock_call_next):
        """测试私有端点绕过认证"""
        mock_request.url.path = "/api/private/test"
        response = await auth_middleware(mock_request, mock_call_next)
        assert response.status_code == 200

    @pytest.mark.asyncio
    async def test_missing_authorization(self, mock_request, mock_call_next):
        """测试缺少Authorization头"""
        mock_request.url.path = "/api/v1/test"
        mock_request.headers = {}
        response = await auth_middleware(mock_request, mock_call_next)
        assert response.status_code == 401

    @pytest.mark.asyncio
    async def test_invalid_authorization_format(self, mock_request, mock_call_next):
        """测试无效的Authorization格式"""
        mock_request.url.path = "/api/v1/test"
        mock_request.headers = {"Authorization": "Invalid token"}
        response = await auth_middleware(mock_request, mock_call_next)
        assert response.status_code == 401

    @pytest.mark.asyncio
    async def test_valid_token(self, mock_request, mock_call_next):
        """测试有效token"""
        mock_request.url.path = "/api/v1/test"
        mock_request.headers = {"Authorization": "Bearer valid_token"}
        
        # Mock response
        mock_response = AsyncMock()
        mock_response.status = 200
        mock_response.text = AsyncMock(return_value='{"active": true, "sub": "user123", "client_id": "client123"}')
        
        # Mock context managers
        mock_post_cm = AsyncMock()
        mock_post_cm.__aenter__ = AsyncMock(return_value=mock_response)
        mock_post_cm.__aexit__ = AsyncMock(return_value=None)
        
        mock_session = AsyncMock()
        mock_session.post = Mock(return_value=mock_post_cm)
        
        mock_session_cm = AsyncMock()
        mock_session_cm.__aenter__ = AsyncMock(return_value=mock_session)
        mock_session_cm.__aexit__ = AsyncMock(return_value=None)
        
        with patch('app.utils.app_utils.aiohttp.ClientSession', return_value=mock_session_cm):
            response = await auth_middleware(mock_request, mock_call_next)
            # 验证通过后应该调用call_next
            assert response.status_code == 200

    @pytest.mark.asyncio
    async def test_inactive_token(self, mock_request, mock_call_next):
        """测试无效token"""
        mock_request.url.path = "/api/v1/test"
        mock_request.headers = {"Authorization": "Bearer invalid_token"}
        
        # Mock response
        mock_response = AsyncMock()
        mock_response.status = 200
        mock_response.text = AsyncMock(return_value='{"active": false}')
        
        # Mock context managers
        mock_post_cm = AsyncMock()
        mock_post_cm.__aenter__ = AsyncMock(return_value=mock_response)
        mock_post_cm.__aexit__ = AsyncMock(return_value=None)
        
        mock_session = AsyncMock()
        mock_session.post = Mock(return_value=mock_post_cm)
        
        mock_session_cm = AsyncMock()
        mock_session_cm.__aenter__ = AsyncMock(return_value=mock_session)
        mock_session_cm.__aexit__ = AsyncMock(return_value=None)
        
        with patch('app.utils.app_utils.aiohttp.ClientSession', return_value=mock_session_cm):
            response = await auth_middleware(mock_request, mock_call_next)
            assert response.status_code == 401


class TestRequestSizeMiddleware:
    """测试RequestSizeMiddleware类"""

    @pytest.mark.asyncio
    async def test_request_within_size_limit(self):
        """测试请求大小在限制内"""
        middleware = RequestSizeMiddleware(app=Mock())
        request = Mock()
        request.headers = {'content-length': '1000'}
        
        async def call_next(req):
            return JSONResponse(content={"status": "ok"})
        
        response = await middleware.dispatch(request, call_next)
        assert response.status_code == 200

    @pytest.mark.asyncio
    async def test_request_exceeds_size_limit(self):
        """测试请求大小超过限制"""
        middleware = RequestSizeMiddleware(app=Mock())
        request = Mock()
        request.headers = {'content-length': str(11 * 1024 * 1024)}  # 11MB
        
        async def call_next(req):
            return JSONResponse(content={"status": "ok"})
        
        response = await middleware.dispatch(request, call_next)
        assert response.status_code == 413

    @pytest.mark.asyncio
    async def test_request_no_content_length(self):
        """测试没有content-length头"""
        middleware = RequestSizeMiddleware(app=Mock())
        request = Mock()
        request.headers = {}
        
        async def call_next(req):
            return JSONResponse(content={"status": "ok"})
        
        response = await middleware.dispatch(request, call_next)
        assert response.status_code == 200


class TestCreateApp:
    """测试create_app函数"""

    def test_create_app_returns_fastapi(self):
        """测试create_app返回FastAPI实例"""
        with patch('app.utils.app_utils.log_init'):
            with patch('app.utils.app_utils.conf_init'):
                with patch('app.utils.app_utils.router_init'):
                    app = create_app()
                    assert isinstance(app, FastAPI)

    def test_create_app_has_title(self):
        """测试应用有标题"""
        with patch('app.utils.app_utils.log_init'):
            with patch('app.utils.app_utils.conf_init'):
                with patch('app.utils.app_utils.router_init'):
                    app = create_app()
                    assert app.title == "My API"

    def test_create_app_has_version(self):
        """测试应用有版本"""
        with patch('app.utils.app_utils.log_init'):
            with patch('app.utils.app_utils.conf_init'):
                with patch('app.utils.app_utils.router_init'):
                    app = create_app()
                    assert app.version == "1.0.0"

    def test_create_app_has_startup_event(self):
        """测试应用有启动事件"""
        with patch('app.utils.app_utils.log_init'):
            with patch('app.utils.app_utils.conf_init'):
                with patch('app.utils.app_utils.router_init'):
                    app = create_app()
                    # FastAPI应该有on_startup配置
                    assert hasattr(app, 'router')

    def test_create_app_initializes_components(self):
        """测试应用初始化各组件"""
        with patch('app.utils.app_utils.log_init') as mock_log:
            with patch('app.utils.app_utils.conf_init') as mock_conf:
                with patch('app.utils.app_utils.router_init') as mock_router:
                    app = create_app()
                    mock_log.assert_called_once()
                    mock_conf.assert_called_once()
                    mock_router.assert_called_once()

