"""测试 llm_controller 模块"""
import pytest
from unittest.mock import Mock, AsyncMock, patch, MagicMock
import json
from fastapi.responses import JSONResponse
from app.controller.llm_controller import used_model_openai


class TestUsedModelOpenai:
    """测试used_model_openai函数"""

    @pytest.fixture
    def valid_request(self):
        """有效请求数据"""
        return {
            "model": "test_model",
            "model_id": "",
            "messages": [
                {"role": "user", "content": "你好", "tool_calls": None, "tool_call_id": None}
            ],
            "temperature": 0.7,
            "top_p": 0.9,
            "frequency_penalty": 0,
            "presence_penalty": 0,
            "max_tokens": 1000,
            "top_k": 1,
            "response_format": None,
            "stop": None,
            "stream": False,
            "cache": True
        }

    @pytest.fixture
    def mock_model_data(self):
        """Mock模型数据"""
        return [{
            "f_model_id": "123456789012345678",
            "f_model_name": "test_model",
            "f_model_series": "openai",
            "f_model": "gpt-3.5-turbo",
            "f_model_config": '{"api_url": "http://test.com", "api_model": "gpt-3.5-turbo", "api_key": "test_key"}',
            "f_max_model_len": 4096,
            "f_quota": False
        }]

    @pytest.mark.asyncio
    async def test_stream_parameter_type_error(self, valid_request):
        """测试stream参数类型错误"""
        request = valid_request.copy()
        request["stream"] = "invalid"  # 应该是bool
        
        with patch('app.controller.llm_controller.get_redis_util', new_callable=AsyncMock):
            result = await used_model_openai(request, "user1", "zh", "test")
            assert isinstance(result, JSONResponse)
            assert result.status_code == 400

    @pytest.mark.asyncio
    async def test_missing_model_and_id(self, valid_request):
        """测试缺少model和model_id"""
        request = valid_request.copy()
        request["model"] = ""
        request["model_id"] = ""
        
        mock_redis = AsyncMock()
        mock_redis.get_str = AsyncMock(return_value=None)
        
        with patch('app.controller.llm_controller.get_redis_util', return_value=mock_redis):
            with patch('app.controller.llm_controller.llm_model_dao.get_data_from_default_model', return_value=[]):
                result = await used_model_openai(request, "user1", "zh", "test")
                assert isinstance(result, JSONResponse)
                assert result.status_code == 400

    @pytest.mark.asyncio
    async def test_max_tokens_exceeds_limit(self, valid_request, mock_model_data):
        """测试max_tokens超过限制"""
        request = valid_request.copy()
        request["max_tokens"] = 10000000  # 超过限制
        
        mock_redis = AsyncMock()
        mock_redis.get_str = AsyncMock(return_value=json.dumps(mock_model_data))
        mock_redis.set_str = AsyncMock()
        
        with patch('app.controller.llm_controller.get_redis_util', return_value=mock_redis):
            with patch('app.controller.llm_controller.llm_model_dao.get_data_from_model_list_by_name_id', return_value=mock_model_data):
                result = await used_model_openai(request, "user1", "zh", "test")
                assert isinstance(result, JSONResponse)
                assert result.status_code == 400

    @pytest.mark.asyncio
    async def test_quota_check_no_space(self, valid_request, mock_model_data):
        """测试配额不足"""
        request = valid_request.copy()
        quota_model_data = mock_model_data.copy()
        quota_model_data[0]["f_quota"] = True
        
        mock_redis = AsyncMock()
        mock_redis.get_str = AsyncMock(side_effect=[
            json.dumps(quota_model_data),
            json.dumps([{"remaining_input_tokens": 0, "remaining_output_tokens": 0}])
        ])
        
        with patch('app.controller.llm_controller.get_redis_util', return_value=mock_redis):
            with patch('app.controller.llm_controller.llm_model_dao.get_quota_by_user_and_model', return_value=[{
                "remaining_input_tokens": 0,
                "remaining_output_tokens": 0
            }]):
                result = await used_model_openai(request, "user1", "zh", "test")
                assert isinstance(result, JSONResponse)
                assert result.status_code == 400

    @pytest.mark.asyncio
    async def test_openai_model_success(self, valid_request, mock_model_data):
        """测试OpenAI模型成功调用"""
        request = valid_request.copy()
        
        mock_redis = AsyncMock()
        mock_redis.get_str = AsyncMock(return_value=json.dumps(mock_model_data))
        mock_redis.set_str = AsyncMock()
        
        mock_response = {
            "choices": [{"message": {"content": "你好！"}}],
            "usage": {"total_tokens": 10}
        }
        
        with patch('app.controller.llm_controller.get_redis_util', return_value=mock_redis):
            with patch('app.controller.llm_controller.llm_model_dao.get_data_from_model_list_by_name_id', return_value=mock_model_data):
                with patch('app.controller.llm_controller.OpenAIClientRequest') as mock_client:
                    mock_client_instance = AsyncMock()
                    mock_client_instance.chat_completion = AsyncMock(return_value=mock_response)
                    mock_client.return_value = mock_client_instance
                    
                    result = await used_model_openai(request, "user1", "zh", "test")
                    assert isinstance(result, JSONResponse)
                    assert result.status_code == 200

    @pytest.mark.asyncio
    @pytest.mark.skip(reason="Claude客户端在llm_utils模块中，需要更复杂的mock")
    async def test_claude_model(self, valid_request):
        """测试Claude模型"""
        pass

    @pytest.mark.asyncio
    @pytest.mark.skip(reason="Baidu客户端在llm_utils模块中，需要更复杂的mock")
    async def test_baidu_model(self, valid_request):
        """测试Baidu模型"""
        pass

    @pytest.mark.asyncio
    async def test_stream_mode(self, valid_request, mock_model_data):
        """测试流式返回模式"""
        request = valid_request.copy()
        request["stream"] = True
        
        mock_redis = AsyncMock()
        mock_redis.get_str = AsyncMock(return_value=json.dumps(mock_model_data))
        
        with patch('app.controller.llm_controller.get_redis_util', return_value=mock_redis):
            with patch('app.controller.llm_controller.llm_model_dao.get_data_from_model_list_by_name_id', return_value=mock_model_data):
                with patch('app.controller.llm_controller.OpenAIClientRequest') as mock_client:
                    mock_client_instance = AsyncMock()
                    async def mock_stream():
                        yield "data: test"
                    mock_client_instance.chat_completion_stream_openai = mock_stream
                    mock_client.return_value = mock_client_instance
                    
                    result = await used_model_openai(request, "user1", "zh", "test")
                    # 流式返回应该是EventSourceResponse
                    assert result is not None

