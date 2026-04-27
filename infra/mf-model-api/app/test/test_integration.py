"""集成测试"""
import pytest
from unittest.mock import Mock, AsyncMock, patch
from fastapi import FastAPI
from fastapi.testclient import TestClient


class TestAppIntegration:
    """应用集成测试"""

    @pytest.fixture
    def app(self):
        """创建测试应用"""
        with patch('app.utils.app_utils.log_init'):
            with patch('app.utils.app_utils.router_init'):
                from app.utils.app_utils import create_app
                return create_app()

    @pytest.fixture
    def client(self, app):
        """创建测试客户端"""
        return TestClient(app)

    def test_app_creation(self, app):
        """测试应用创建"""
        assert app is not None
        assert isinstance(app, FastAPI)

    def test_app_has_title(self, app):
        """测试应用标题"""
        assert app.title == "My API"

    def test_app_has_version(self, app):
        """测试应用版本"""
        assert app.version == "1.0.0"


class TestHealthCheckIntegration:
    """健康检查集成测试"""

    @pytest.fixture
    def app(self):
        """创建带健康检查的应用"""
        app = FastAPI()
        from app.routers.llm_router import health_route
        app.include_router(health_route)
        return app

    @pytest.fixture
    def client(self, app):
        """创建测试客户端"""
        return TestClient(app)

    def test_health_endpoints_work(self, client):
        """测试健康检查端点工作"""
        # 测试ready端点
        response = client.get("/health/ready")
        assert response.status_code == 200
        assert response.json() == {"res": 0}
        
        # 测试alive端点
        response = client.get("/health/alive")
        assert response.status_code == 200
        assert response.json() == {"res": 0}


class TestModuleImports:
    """模块导入测试"""

    def test_import_commons(self):
        """测试导入commons模块"""
        from app.commons import response, snow_id
        assert response is not None
        assert snow_id is not None

    def test_import_utils(self):
        """测试导入utils模块"""
        from app.utils import str_util, comment_utils
        assert str_util is not None
        assert comment_utils is not None

    def test_import_core(self):
        """测试导入core模块"""
        from app.core import config
        assert config is not None

    def test_import_dao(self):
        """测试导入dao模块"""
        from app.dao import llm_model_dao
        assert llm_model_dao is not None

    def test_import_controller(self):
        """测试导入controller模块"""
        from app.controller import llm_controller
        assert llm_controller is not None

    def test_import_routers(self):
        """测试导入routers模块"""
        from app.routers import llm_router
        assert llm_router is not None


class TestConfigurationIntegration:
    """配置集成测试"""

    def test_base_config_accessible(self):
        """测试base_config可访问"""
        from app.core.config import base_config
        assert base_config is not None
        assert hasattr(base_config, 'PORTDEFAULT')

    def test_server_info_accessible(self):
        """测试server_info可访问"""
        from app.core.config import server_info
        assert server_info is not None
        assert server_info.server_name == "agent-executor"

    def test_observability_config_accessible(self):
        """测试observability_config可访问"""
        from app.core.config import observability_config
        assert observability_config is not None
        assert hasattr(observability_config, 'log')


class TestDAOIntegration:
    """DAO层集成测试"""

    def test_llm_model_dao_instance(self):
        """测试llm_model_dao实例"""
        from app.dao.llm_model_dao import llm_model_dao
        assert llm_model_dao is not None
        assert hasattr(llm_model_dao, 'get_model_name_by_id')
        assert hasattr(llm_model_dao, 'get_all_model_list')


class TestUtilsIntegration:
    """工具函数集成测试"""

    def test_snow_id_generation(self):
        """测试雪花ID生成"""
        import time
        from app.commons.snow_id import snow_id
        id1 = snow_id()
        time.sleep(0.001)  # 确保时间戳不同
        id2 = snow_id()
        assert id1 != id2
        assert isinstance(id1, int)
        assert isinstance(id2, int)

    def test_md5_generation(self):
        """测试MD5生成"""
        from app.utils.str_util import get_md5_key
        result1 = get_md5_key("test")
        result2 = get_md5_key("test")
        assert result1 == result2
        assert len(result1) == 32

    def test_random_string_generation(self):
        """测试随机字符串生成"""
        from app.utils.str_util import generate_random_string
        str1 = generate_random_string(32)
        str2 = generate_random_string(32)
        assert str1 != str2
        assert len(str1) == 32
        assert len(str2) == 32

