# -*- coding: utf-8 -*-
"""单元测试 - driven/dip/model_api_service 模块"""

import pytest
from unittest.mock import MagicMock, patch

from aioresponses import aioresponses
from app.driven.dip.model_api_service import ModelApiService, model_api_service
from app.common.errors import CodeException


@pytest.fixture
def mock_config():
    with patch("app.driven.dip.model_api_service.Config") as mock:
        config = MagicMock()
        config.services.mf_model_api.host = "localhost"
        config.services.mf_model_api.port = 8080
        yield config


@pytest.fixture
def model_service():
    return ModelApiService()


class TestModelApiServiceInit:
    def test_init(self, mock_config):
        with patch("app.driven.dip.model_api_service.Config", mock_config):
            service = ModelApiService()
            assert service._host == "localhost"
            assert service._port == 8080
            assert service._basic_url == "http://localhost:8080"
            assert isinstance(service.headers, dict)


class TestCallStreamObj:
    """测试 call_stream_obj 方法"""

    @pytest.mark.asyncio
    async def test_call_stream_obj_success(self, model_service, mock_config):
        """测试流式调用成功"""
        with patch("app.driven.dip.model_api_service.Config", mock_config):
            service = ModelApiService()

            # 准备 SSE (Server-Sent Events) 响应
            sse_data = (
                b'data: {"choices": [{"delta": {"content": "hello"}}]}\n'
                b'data: {"choices": [{"delta": {"content": "world"}}]}\n'
                b'data: {"choices": [{"delta": {"content": "!"}}]}\n'
            )

            with aioresponses() as m:
                m.post(
                    "http://localhost:8080/api/private/mf-model-api/v1/chat/completions",
                    status=200,
                    body=sse_data,
                    headers={"Content-Type": "text/event-stream"},
                )

                results = []
                async for item in service.call_stream_obj(
                    model="test-model", messages=[{"role": "user", "content": "hello"}]
                ):
                    results.append(item)

                assert len(results) == 3
                assert results[0] == {"content": "hello"}
                assert results[1] == {"content": "world"}
                assert results[2] == {"content": "!"}

    @pytest.mark.asyncio
    async def test_call_stream_obj_with_all_params(self, model_service, mock_config):
        """测试带所有参数的流式调用"""
        with patch("app.driven.dip.model_api_service.Config", mock_config):
            service = ModelApiService()

            with aioresponses() as m:
                m.post(
                    "http://localhost:8080/api/private/mf-model-api/v1/chat/completions",
                    status=200,
                    body=b'data: {"choices": [{"delta": {"content": "response"}}]}\n',
                )

                results = []
                async for item in service.call_stream_obj(
                    model="gpt-4",
                    messages=[{"role": "user", "content": "test"}],
                    top_p=0.9,
                    temperature=0.7,
                    presence_penalty=0.5,
                    frequency_penalty=0.3,
                    max_tokens=2000,
                    top_k=50,
                    userid="test-user",
                ):
                    results.append(item)

                assert len(results) == 1

    @pytest.mark.asyncio
    async def test_call_stream_obj_error_response(self, model_service, mock_config):
        """测试错误响应"""
        with patch("app.driven.dip.model_api_service.Config", mock_config):
            service = ModelApiService()

            with aioresponses() as m:
                m.post(
                    "http://localhost:8080/api/private/mf-model-api/v1/chat/completions",
                    status=500,
                    body=b"Internal Server Error",
                )

                with pytest.raises(CodeException):
                    results = []
                    async for item in service.call_stream_obj(
                        model="test-model",
                        messages=[{"role": "user", "content": "hello"}],
                    ):
                        results.append(item)

    @pytest.mark.asyncio
    async def test_call_stream_obj_top_k_default(self, model_service, mock_config):
        """测试 top_k 默认值为 100"""
        with patch("app.driven.dip.model_api_service.Config", mock_config):
            service = ModelApiService()

            with aioresponses() as m:
                m.post(
                    "http://localhost:8080/api/private/mf-model-api/v1/chat/completions",
                    status=200,
                    body=b"",
                )

                results = []
                async for item in service.call_stream_obj(
                    model="test-model",
                    messages=[{"role": "user", "content": "hello"}],
                    top_k=None,  # Should default to 100
                ):
                    results.append(item)

                assert len(results) == 0

    @pytest.mark.asyncio
    async def test_call_stream_obj_empty_stream(self, model_service, mock_config):
        """测试空流"""
        with patch("app.driven.dip.model_api_service.Config", mock_config):
            service = ModelApiService()

            with aioresponses() as m:
                m.post(
                    "http://localhost:8080/api/private/mf-model-api/v1/chat/completions",
                    status=200,
                    body=b"",
                )

                results = []
                async for item in service.call_stream_obj(
                    model="test-model", messages=[{"role": "user", "content": "hello"}]
                ):
                    results.append(item)

                assert len(results) == 0

    @pytest.mark.asyncio
    async def test_call_stream_obj_malformed_json_skipped(
        self, model_service, mock_config
    ):
        """测试格式错误的 JSON 被跳过"""
        with patch("app.driven.dip.model_api_service.Config", mock_config):
            service = ModelApiService()

            # Mix of valid and invalid JSON
            sse_data = (
                b'data: {"choices": [{"delta": {"content": "valid"}}]}\n'
                b"data: invalid json\n"
                b'data: {"choices": [{"delta": {"content": "also valid"}}]}\n'
                b"data: [DONE]\n"
            )

            with aioresponses() as m:
                m.post(
                    "http://localhost:8080/api/private/mf-model-api/v1/chat/completions",
                    status=200,
                    body=sse_data,
                )

                results = []
                async for item in service.call_stream_obj(
                    model="test-model", messages=[{"role": "user", "content": "hello"}]
                ):
                    results.append(item)

                # Invalid JSON should be skipped
                assert len(results) == 2
                assert results[0] == {"content": "valid"}
                assert results[1] == {"content": "also valid"}

    @pytest.mark.asyncio
    async def test_call_stream_obj_with_reasoning_content(
        self, model_service, mock_config
    ):
        """测试包含 reasoning_content 的响应"""
        with patch("app.driven.dip.model_api_service.Config", mock_config):
            service = ModelApiService()

            sse_data = b'data: {"choices": [{"delta": {"content": "response", "reasoning_content": "thinking"}}]}\n'

            with aioresponses() as m:
                m.post(
                    "http://localhost:8080/api/private/mf-model-api/v1/chat/completions",
                    status=200,
                    body=sse_data,
                )

                results = []
                async for item in service.call_stream_obj(
                    model="test-model", messages=[{"role": "user", "content": "hello"}]
                ):
                    results.append(item)

                assert len(results) == 1
                assert results[0] == {
                    "content": "response",
                    "reasoning_content": "thinking",
                }


class TestCallStream:
    """测试 call_stream 方法"""

    @pytest.mark.asyncio
    async def test_call_stream_content_only(self, model_service):
        """测试只返回 content 字段"""

        async def mock_stream_obj(*args, **kwargs):
            yield {"content": "hello", "other": "data"}
            yield {"content": "world"}
            yield {"other": "data"}
            yield {"content": None}

        with patch.object(model_service, "call_stream_obj", mock_stream_obj):
            results = []
            async for item in model_service.call_stream(
                model="test-model", messages=[{"role": "user", "content": "hello"}]
            ):
                results.append(item)

            assert results == ["hello", "world", "", ""]

    @pytest.mark.asyncio
    async def test_call_stream_with_parameters(self, model_service):
        """测试带参数调用"""
        called_params = {}

        async def mock_stream_obj(*args, **kwargs):
            called_params.update(kwargs)
            yield {"content": "test"}

        with patch.object(model_service, "call_stream_obj", mock_stream_obj):
            results = []
            async for item in model_service.call_stream(
                model="test-model",
                messages=[{"role": "user", "content": "hello"}],
                top_p=0.9,
                temperature=0.7,
            ):
                results.append(item)

            assert called_params["top_p"] == 0.9
            assert called_params["temperature"] == 0.7


class TestCallObj:
    """测试 call_obj 方法"""

    @pytest.mark.asyncio
    async def test_call_obj_basic_success(self, model_service, mock_config):
        """测试非流式调用成功"""
        with patch("app.driven.dip.model_api_service.Config", mock_config):
            service = ModelApiService()

            response_data = {"choices": [{"message": {"content": "test response"}}]}

            with aioresponses() as m:
                m.post(
                    "http://localhost:8080/api/private/mf-model-api/v1/chat/completions",
                    status=200,
                    payload=response_data,
                )

                result = await service.call_obj(
                    model="test-model", messages=[{"role": "user", "content": "hello"}]
                )

                assert result == {"content": "test response"}

    @pytest.mark.asyncio
    async def test_call_obj_with_all_params(self, model_service, mock_config):
        """测试带所有参数的非流式调用"""
        with patch("app.driven.dip.model_api_service.Config", mock_config):
            service = ModelApiService()

            response_data = {"choices": [{"message": {"content": "response"}}]}

            with aioresponses() as m:
                m.post(
                    "http://localhost:8080/api/private/mf-model-api/v1/chat/completions",
                    status=200,
                    payload=response_data,
                )

                result = await service.call_obj(
                    model="gpt-4",
                    messages=[{"role": "user", "content": "test"}],
                    top_p=0.9,
                    temperature=0.7,
                    presence_penalty=0.5,
                    frequency_penalty=0.3,
                    max_tokens=2000,
                    top_k=50,
                    userid="test-user",
                )

                assert result == {"content": "response"}

    @pytest.mark.asyncio
    async def test_call_obj_with_reasoning_content(self, model_service, mock_config):
        """测试返回包含 reasoning_content 的响应"""
        with patch("app.driven.dip.model_api_service.Config", mock_config):
            service = ModelApiService()

            response_data = {
                "choices": [
                    {
                        "message": {
                            "content": "response",
                            "reasoning_content": "thinking",
                        }
                    }
                ]
            }

            with aioresponses() as m:
                m.post(
                    "http://localhost:8080/api/private/mf-model-api/v1/chat/completions",
                    status=200,
                    payload=response_data,
                )

                result = await service.call_obj(
                    model="test-model", messages=[{"role": "user", "content": "hello"}]
                )

                assert result == {
                    "content": "response",
                    "reasoning_content": "thinking",
                }

    @pytest.mark.asyncio
    async def test_call_obj_error_response(self, model_service, mock_config):
        """测试错误响应"""
        with patch("app.driven.dip.model_api_service.Config", mock_config):
            service = ModelApiService()

            with aioresponses() as m:
                m.post(
                    "http://localhost:8080/api/private/mf-model-api/v1/chat/completions",
                    status=500,
                    body=b"Internal Server Error",
                )

                with pytest.raises(CodeException):
                    await service.call_obj(
                        model="test-model",
                        messages=[{"role": "user", "content": "hello"}],
                    )

    @pytest.mark.asyncio
    async def test_call_obj_top_k_default(self, model_service, mock_config):
        """测试 top_k 默认值为 100"""
        with patch("app.driven.dip.model_api_service.Config", mock_config):
            service = ModelApiService()

            response_data = {"choices": [{"message": {"content": "response"}}]}

            with aioresponses() as m:
                m.post(
                    "http://localhost:8080/api/private/mf-model-api/v1/chat/completions",
                    status=200,
                    payload=response_data,
                )

                result = await service.call_obj(
                    model="test-model",
                    messages=[{"role": "user", "content": "hello"}],
                    top_k=None,
                )

                assert result == {"content": "response"}


class TestCall:
    """测试 call 方法"""

    @pytest.mark.asyncio
    async def test_call_returns_content(self, model_service):
        """测试返回 content 字段"""
        with patch.object(model_service, "call_obj") as mock_call:
            mock_call.return_value = {"content": "test content"}

            result = await model_service.call(
                model="test-model", messages=[{"role": "user", "content": "hello"}]
            )

            assert result == "test content"

    @pytest.mark.asyncio
    async def test_call_with_no_content(self, model_service):
        """测试没有 content 字段时返回空字符串"""
        with patch.object(model_service, "call_obj") as mock_call:
            mock_call.return_value = {"other": "data"}

            result = await model_service.call(
                model="test-model", messages=[{"role": "user", "content": "hello"}]
            )

            assert result == ""

    @pytest.mark.asyncio
    async def test_call_with_none_content(self, model_service):
        """测试 content 为 None 时返回空字符串"""
        with patch.object(model_service, "call_obj") as mock_call:
            mock_call.return_value = {"content": None}

            result = await model_service.call(
                model="test-model", messages=[{"role": "user", "content": "hello"}]
            )

            assert result == ""

    @pytest.mark.asyncio
    async def test_call_with_parameters(self, model_service):
        """测试带参数调用"""
        with patch.object(model_service, "call_obj") as mock_call:
            mock_call.return_value = {"content": "test"}

            result = await model_service.call(
                model="test-model",
                messages=[{"role": "user", "content": "hello"}],
                top_p=0.9,
                temperature=0.7,
            )

            mock_call.assert_called_once()
            call_kwargs = mock_call.call_args.kwargs
            assert call_kwargs["top_p"] == 0.9
            assert call_kwargs["temperature"] == 0.7


class TestModelApiServiceSingleton:
    """测试单例"""

    def test_singleton_instance(self):
        """测试全局单例"""

        assert model_api_service is not None
        assert isinstance(model_api_service, ModelApiService)

    def test_singleton_same_instance(self):
        """测试多次获取是同一个实例"""
        from app.driven.dip.model_api_service import model_api_service as svc1
        from app.driven.dip.model_api_service import model_api_service as svc2

        assert svc1 is svc2
