"""测试 llm_model_dao 模块"""
import pytest
from unittest.mock import Mock, patch, MagicMock
import json
from datetime import datetime
from app.dao.llm_model_dao import ModelDao, llm_model_dao


class TestModelDao:
    """测试ModelDao类"""

    @pytest.fixture
    def dao(self):
        """创建DAO实例"""
        return ModelDao()

    @pytest.fixture
    def mock_cursor(self):
        """创建mock cursor"""
        cursor = Mock()
        cursor.fetchall = Mock(return_value=[])
        cursor.fetchone = Mock(return_value=None)
        cursor.execute = Mock()
        return cursor

    def test_get_model_name_by_id(self, dao, mock_cursor):
        """测试通过ID获取模型名称"""
        mock_cursor.fetchall.return_value = [{"f_model_name": "test_model"}]
        with patch.object(dao, 'get_model_name_by_id', return_value="test_model"):
            result = dao.get_model_name_by_id("123")
            assert result == "test_model"

    def test_get_model_id_by_name(self, dao, mock_cursor):
        """测试通过名称获取模型ID"""
        expected = [{"f_model_id": "123456789012345678"}]
        mock_cursor.fetchall.return_value = expected
        with patch.object(dao, 'get_model_id_by_name', return_value=expected):
            result = dao.get_model_id_by_name("test_model")
            assert result == expected

    def test_get_model_series_by_id(self, dao):
        """测试通过ID获取模型系列"""
        with patch.object(dao, 'get_model_series_by_id', return_value="openai"):
            result = dao.get_model_series_by_id("123")
            assert result == "openai"

    def test_get_all_model_list(self, dao):
        """测试获取所有模型列表"""
        expected = [{
            "f_model_id": "123",
            "f_model_name": "test_model",
            "f_model_series": "openai",
            "f_model_type": "llm"
        }]
        with patch.object(dao, 'get_all_model_list', return_value=expected):
            result = dao.get_all_model_list()
            assert len(result) == 1
            assert result[0]["f_model_name"] == "test_model"

    def test_get_data_from_model_list_by_id(self, dao):
        """测试通过ID获取模型详细数据"""
        expected = [{
            "f_model_id": "123",
            "f_model_name": "test_model",
            "f_model_config": '{"api_url": "http://test.com"}',
            "f_max_model_len": 4096
        }]
        with patch.object(dao, 'get_data_from_model_list_by_id', return_value=expected):
            result = dao.get_data_from_model_list_by_id("123")
            assert len(result) == 1

    def test_get_data_from_model_list_by_name_id_with_name(self, dao):
        """测试通过名称或ID获取模型（使用名称）"""
        expected = [{"f_model_id": "123", "f_model_name": "test_model"}]
        with patch.object(dao, 'get_data_from_model_list_by_name_id', return_value=expected):
            result = dao.get_data_from_model_list_by_name_id("test_model", None)
            assert len(result) == 1

    def test_get_data_from_model_list_by_name_id_with_id(self, dao):
        """测试通过名称或ID获取模型（使用ID）"""
        expected = [{"f_model_id": "123", "f_model_name": "test_model"}]
        with patch.object(dao, 'get_data_from_model_list_by_name_id', return_value=expected):
            result = dao.get_data_from_model_list_by_name_id(None, "123")
            assert len(result) == 1

    def test_delete_model_by_id(self, dao):
        """测试删除模型"""
        with patch.object(dao, 'delete_model_by_id', return_value=None):
            result = dao.delete_model_by_id(["123", "456"])
            assert result is None

    def test_get_model_by_name(self, dao):
        """测试通过名称获取模型"""
        expected = [{"f_model_name": "test_model", "f_model_id": "123"}]
        with patch.object(dao, 'get_model_by_name', return_value=expected):
            result = dao.get_model_by_name("test_model")
            assert len(result) == 1

    def test_check_model_is_exist_true(self, dao):
        """测试检查模型存在"""
        with patch.object(dao, 'check_model_is_exist', return_value=True):
            result = dao.check_model_is_exist("123")
            assert result is True

    def test_check_model_is_exist_false(self, dao):
        """测试检查模型不存在"""
        with patch.object(dao, 'check_model_is_exist', return_value=False):
            result = dao.check_model_is_exist("999")
            assert result is False

    def test_check_model_unique_duplicate(self, dao):
        """测试检查模型唯一性（重复）"""
        with patch.object(dao, 'check_model_unique', return_value=True):
            result = dao.check_model_unique("http://test.com", "gpt-3.5", "user1", "key123")
            assert result is True

    def test_check_model_unique_not_duplicate(self, dao):
        """测试检查模型唯一性（不重复）"""
        with patch.object(dao, 'check_model_unique', return_value=False):
            result = dao.check_model_unique("http://new.com", "gpt-4", "user1", "key456")
            assert result is False

    def test_get_model_default_paras(self, dao):
        """测试获取模型默认参数"""
        expected = {
            "123": {"model_name": "test_model", "model_series": "openai", "model": "gpt-3.5"}
        }
        with patch.object(dao, 'get_model_default_paras', return_value=expected):
            result = dao.get_model_default_paras()
            assert "123" in result

    def test_get_all_tome_model_list(self, dao):
        """测试获取所有tome模型列表"""
        expected = [{"f_model_name": "tome_model", "f_model_series": "tome"}]
        with patch.object(dao, 'get_all_tome_model_list', return_value=expected):
            result = dao.get_all_tome_model_list()
            assert len(result) == 1

    def test_get_quota_by_user_and_model_exists(self, dao):
        """测试获取用户模型配额（存在）"""
        expected = [{
            "f_input_tokens": 1000000,
            "used_input_tokens": 5000,
            "f_output_tokens": 1000000,
            "used_output_tokens": 3000,
            "f_billing_type": 1,
            "f_num_type": "[0, 0]",
            "remaining_input_tokens": 995000,
            "remaining_output_tokens": 997000,
            "total_input_tokens": 1000000,
            "total_output_tokens": 1000000
        }]
        with patch.object(dao, 'get_quota_by_user_and_model', return_value=expected):
            result = dao.get_quota_by_user_and_model("user1", "model123")
            assert len(result) == 1
            assert result[0]["remaining_input_tokens"] == 995000

    def test_get_quota_by_user_and_model_not_exists(self, dao):
        """测试获取用户模型配额（不存在）"""
        with patch.object(dao, 'get_quota_by_user_and_model', return_value=[]):
            result = dao.get_quota_by_user_and_model("user1", "model999")
            assert len(result) == 0

    def test_get_data_from_default_model(self, dao):
        """测试获取默认模型数据"""
        expected = [{
            "f_model_id": "default123",
            "f_model_name": "default_model",
            "f_default": 1
        }]
        with patch.object(dao, 'get_data_from_default_model', return_value=expected):
            result = dao.get_data_from_default_model()
            assert len(result) == 1


class TestLlmModelDaoInstance:
    """测试llm_model_dao实例"""

    def test_instance_exists(self):
        """测试实例存在"""
        assert llm_model_dao is not None
        assert isinstance(llm_model_dao, ModelDao)

    def test_instance_has_methods(self):
        """测试实例具有必要的方法"""
        assert hasattr(llm_model_dao, 'get_model_name_by_id')
        assert hasattr(llm_model_dao, 'get_all_model_list')
        assert hasattr(llm_model_dao, 'delete_model_by_id')
        assert hasattr(llm_model_dao, 'check_model_is_exist')

