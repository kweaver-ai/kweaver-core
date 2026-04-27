"""测试 controller/model_audit_controller.py 模块"""
import pytest
from unittest.mock import AsyncMock, Mock, patch
from fastapi.responses import JSONResponse
from app.controller.model_audit_controller import add_llm_model_call_log
from app.interfaces import logics


class TestModelAuditController:
    """测试模型审计控制器"""

    @pytest.fixture
    def mock_audit_request(self):
        """创建审计请求的mock"""
        request = Mock(spec=logics.AddModelUsedAudit)
        request.model_id = "123456789012345678"
        request.user_id = "user123"
        request.input_tokens = 100
        request.output_tokens = 50
        request.total_time = 1.5
        request.first_time = 0.5
        request.status = "success"
        return request

    @pytest.mark.asyncio
    async def test_add_llm_model_call_log_success(self, mock_audit_request):
        """测试成功添加LLM模型调用日志"""
        with patch('app.controller.model_audit_controller.kafka_client') as mock_kafka:
            mock_kafka.produce_async.return_value = True
            
            # 不应该抛出异常
            await add_llm_model_call_log(mock_audit_request)
            
            # 验证kafka_client.produce_async被调用
            mock_kafka.produce_async.assert_called_once()

    @pytest.mark.asyncio
    async def test_add_llm_model_call_log_kafka_full(self, mock_audit_request):
        """测试Kafka队列满时的处理"""
        with patch('app.controller.model_audit_controller.kafka_client') as mock_kafka:
            mock_kafka.produce_async.return_value = False
            
            # 不应该抛出异常，只是记录警告
            await add_llm_model_call_log(mock_audit_request)
            
            mock_kafka.produce_async.assert_called_once()

    @pytest.mark.asyncio
    async def test_add_llm_model_call_log_exception(self, mock_audit_request):
        """测试添加日志时发生异常"""
        with patch('app.controller.model_audit_controller.kafka_client') as mock_kafka:
            mock_kafka.produce_async.side_effect = Exception("Kafka connection error")
            
            # 不应该抛出异常，只是记录错误
            await add_llm_model_call_log(mock_audit_request)

    @pytest.mark.asyncio
    async def test_add_llm_model_call_log_message_format(self, mock_audit_request):
        """测试消息格式是否正确"""
        with patch('app.controller.model_audit_controller.kafka_client') as mock_kafka:
            mock_kafka.produce_async.return_value = True
            
            await add_llm_model_call_log(mock_audit_request)
            
            # 获取调用参数
            call_args = mock_kafka.produce_async.call_args
            
            # 验证key和value都被正确编码为bytes
            assert 'value' in call_args.kwargs
            assert 'key' in call_args.kwargs
            assert isinstance(call_args.kwargs['value'], bytes)
            assert isinstance(call_args.kwargs['key'], bytes)

    @pytest.mark.asyncio
    async def test_add_llm_model_call_log_with_different_status(self):
        """测试不同状态的日志记录"""
        for status in ['success', 'failed', 'timeout']:
            request = Mock(spec=logics.AddModelUsedAudit)
            request.model_id = "123456789012345678"
            request.user_id = "user123"
            request.input_tokens = 100
            request.output_tokens = 50
            request.total_time = 1.5
            request.first_time = 0.5
            request.status = status
            
            with patch('app.controller.model_audit_controller.kafka_client') as mock_kafka:
                mock_kafka.produce_async.return_value = True
                
                await add_llm_model_call_log(request)
                
                mock_kafka.produce_async.assert_called_once()

    @pytest.mark.asyncio
    async def test_add_llm_model_call_log_zero_tokens(self):
        """测试零token的日志记录"""
        request = Mock(spec=logics.AddModelUsedAudit)
        request.model_id = "123456789012345678"
        request.user_id = "user123"
        request.input_tokens = 0
        request.output_tokens = 0
        request.total_time = 0.1
        request.first_time = 0.1
        request.status = "success"
        
        with patch('app.controller.model_audit_controller.kafka_client') as mock_kafka:
            mock_kafka.produce_async.return_value = True
            
            await add_llm_model_call_log(request)
            
            mock_kafka.produce_async.assert_called_once()

    @pytest.mark.asyncio
    async def test_add_llm_model_call_log_large_tokens(self):
        """测试大量token的日志记录"""
        request = Mock(spec=logics.AddModelUsedAudit)
        request.model_id = "123456789012345678"
        request.user_id = "user123"
        request.input_tokens = 100000
        request.output_tokens = 50000
        request.total_time = 30.0
        request.first_time = 5.0
        request.status = "success"
        
        with patch('app.controller.model_audit_controller.kafka_client') as mock_kafka:
            mock_kafka.produce_async.return_value = True
            
            await add_llm_model_call_log(request)
            
            mock_kafka.produce_async.assert_called_once()

    @pytest.mark.asyncio
    async def test_add_llm_model_call_log_timing(self):
        """测试日志记录的时间消耗"""
        request = Mock(spec=logics.AddModelUsedAudit)
        request.model_id = "123456789012345678"
        request.user_id = "user123"
        request.input_tokens = 100
        request.output_tokens = 50
        request.total_time = 1.5
        request.first_time = 0.5
        request.status = "success"
        
        with patch('app.controller.model_audit_controller.kafka_client') as mock_kafka:
            mock_kafka.produce_async.return_value = True
            
            import time
            start = time.time()
            await add_llm_model_call_log(request)
            elapsed = time.time() - start
            
            # 异步发送应该很快完成（小于1秒）
            assert elapsed < 1.0

