"""单元测试 - utils/observability/trace_wrapper 模块"""

import pytest
from unittest.mock import MagicMock, patch
import sys


class TestInternalSpan:
    """测试 internal_span 装饰器"""

    def test_returns_original_function_when_sdk_unavailable(self):
        """测试SDK不可用时返回原函数"""
        with patch(
            "app.utils.observability.trace_wrapper.TELEMETRY_SDK_AVAILABLE", False
        ):
            from app.utils.observability.trace_wrapper import internal_span

            @internal_span()
            def test_func():
                return "result"

            result = test_func()
            assert result == "result"

    @patch("app.utils.observability.trace_wrapper.TELEMETRY_SDK_AVAILABLE", True)
    @patch("app.common.config.Config")
    def test_returns_original_function_when_trace_disabled(self, m_config):
        """测试追踪未启用时返回原函数"""
        m_config.is_o11y_trace_enabled.return_value = False
        from app.utils.observability.trace_wrapper import internal_span

        @internal_span()
        def test_func():
            return "result"

        result = test_func()
        assert result == "result"

    @patch("app.utils.observability.trace_wrapper.TELEMETRY_SDK_AVAILABLE", True)
    @patch("app.common.config.Config")
    @patch("app.utils.observability.trace_wrapper.func_judgment")
    def test_sync_function_wrapper(self, m_func_judgment, m_config):
        """测试同步函数包装"""
        m_config.is_o11y_trace_enabled.return_value = True
        m_func_judgment.return_value = (False, False)  # not async, not stream

        # Mock the tracer module
        mock_tracer = MagicMock()
        mock_span = MagicMock()
        mock_tracer.start_span.return_value = mock_span
        mock_span.is_recording.return_value = True

        mock_exporter_module = MagicMock()
        mock_exporter_module.tracer = mock_tracer
        sys.modules["exporter.ar_trace.trace_exporter"] = mock_exporter_module

        try:
            from app.utils.observability.trace_wrapper import internal_span

            @internal_span(name="test_span")
            def test_func(x, y, span=None):
                return x + y

            result = test_func(1, 2)

            assert result == 3
            mock_tracer.start_span.assert_called_once()
            mock_span.set_status.assert_called()
            mock_span.end.assert_called_once()
        finally:
            del sys.modules["exporter.ar_trace.trace_exporter"]

    @patch("app.utils.observability.trace_wrapper.TELEMETRY_SDK_AVAILABLE", True)
    @patch("app.common.config.Config")
    @patch("app.utils.observability.trace_wrapper.func_judgment")
    def test_sync_function_with_attributes(self, m_func_judgment, m_config):
        """测试带属性的同步函数"""
        m_config.is_o11y_trace_enabled.return_value = True
        m_func_judgment.return_value = (False, False)

        mock_tracer = MagicMock()
        mock_span = MagicMock()
        mock_tracer.start_span.return_value = mock_span
        mock_span.is_recording.return_value = True

        mock_exporter_module = MagicMock()
        mock_exporter_module.tracer = mock_tracer
        sys.modules["exporter.ar_trace.trace_exporter"] = mock_exporter_module

        try:
            from app.utils.observability.trace_wrapper import internal_span

            attributes = {"key": "value"}

            @internal_span(name="test_span", attributes=attributes)
            def test_func(span=None):
                return "result"

            result = test_func()

            assert result == "result"
        finally:
            del sys.modules["exporter.ar_trace.trace_exporter"]

    @patch("app.utils.observability.trace_wrapper.TELEMETRY_SDK_AVAILABLE", True)
    @patch("app.common.config.Config")
    @patch("app.utils.observability.trace_wrapper.func_judgment")
    def test_sync_function_exception_handling(self, m_func_judgment, m_config):
        """测试同步函数异常处理"""
        m_config.is_o11y_trace_enabled.return_value = True
        m_func_judgment.return_value = (False, False)

        mock_tracer = MagicMock()
        mock_span = MagicMock()
        mock_tracer.start_span.return_value = mock_span
        mock_span.is_recording.return_value = True

        mock_exporter_module = MagicMock()
        mock_exporter_module.tracer = mock_tracer
        sys.modules["exporter.ar_trace.trace_exporter"] = mock_exporter_module

        try:
            from app.utils.observability.trace_wrapper import internal_span

            @internal_span()
            def test_func(span=None):
                raise ValueError("test error")

            with pytest.raises(ValueError, match="test error"):
                test_func()

            mock_span.set_status.assert_called()
            mock_span.record_exception.assert_called_once()
            mock_span.end.assert_called_once()
        finally:
            del sys.modules["exporter.ar_trace.trace_exporter"]

    @pytest.mark.asyncio
    @patch("app.utils.observability.trace_wrapper.TELEMETRY_SDK_AVAILABLE", True)
    @patch("app.common.config.Config")
    @patch("app.utils.observability.trace_wrapper.func_judgment")
    async def test_async_function_wrapper(self, m_func_judgment, m_config):
        """测试异步函数包装"""
        m_config.is_o11y_trace_enabled.return_value = True
        m_func_judgment.return_value = (True, False)  # async, not stream

        mock_tracer = MagicMock()
        mock_span = MagicMock()
        mock_tracer.start_span.return_value = mock_span
        mock_span.is_recording.return_value = True

        mock_exporter_module = MagicMock()
        mock_exporter_module.tracer = mock_tracer
        sys.modules["exporter.ar_trace.trace_exporter"] = mock_exporter_module

        try:
            from app.utils.observability.trace_wrapper import internal_span

            @internal_span(name="async_span")
            async def test_func(x, span=None):
                return x * 2

            result = await test_func(5)

            assert result == 10
            mock_tracer.start_span.assert_called_once()
            mock_span.end.assert_called_once()
        finally:
            del sys.modules["exporter.ar_trace.trace_exporter"]

    @pytest.mark.asyncio
    @patch("app.utils.observability.trace_wrapper.TELEMETRY_SDK_AVAILABLE", True)
    @patch("app.common.config.Config")
    @patch("app.utils.observability.trace_wrapper.func_judgment")
    async def test_async_function_exception(self, m_func_judgment, m_config):
        """测试异步函数异常处理"""
        m_config.is_o11y_trace_enabled.return_value = True
        m_func_judgment.return_value = (True, False)

        mock_tracer = MagicMock()
        mock_span = MagicMock()
        mock_tracer.start_span.return_value = mock_span
        mock_span.is_recording.return_value = True

        mock_exporter_module = MagicMock()
        mock_exporter_module.tracer = mock_tracer
        sys.modules["exporter.ar_trace.trace_exporter"] = mock_exporter_module

        try:
            from app.utils.observability.trace_wrapper import internal_span

            @internal_span()
            async def test_func(span=None):
                raise RuntimeError("async error")

            with pytest.raises(RuntimeError, match="async error"):
                await test_func()

            mock_span.set_status.assert_called()
            mock_span.record_exception.assert_called_once()
        finally:
            del sys.modules["exporter.ar_trace.trace_exporter"]

    @pytest.mark.asyncio
    @patch("app.utils.observability.trace_wrapper.TELEMETRY_SDK_AVAILABLE", True)
    @patch("app.common.config.Config")
    @patch("app.utils.observability.trace_wrapper.func_judgment")
    async def test_async_generator_wrapper(self, m_func_judgment, m_config):
        """测试异步生成器函数包装"""
        m_config.is_o11y_trace_enabled.return_value = True
        m_func_judgment.return_value = (True, True)  # async, stream

        mock_tracer = MagicMock()
        mock_span = MagicMock()
        mock_tracer.start_span.return_value = mock_span
        mock_span.is_recording.return_value = True

        mock_exporter_module = MagicMock()
        mock_exporter_module.tracer = mock_tracer
        sys.modules["exporter.ar_trace.trace_exporter"] = mock_exporter_module

        try:
            from app.utils.observability.trace_wrapper import internal_span

            @internal_span(name="generator_span")
            async def test_func(n, span=None):
                for i in range(n):
                    yield i * 2

            results = []
            async for item in test_func(3):
                results.append(item)

            assert results == [0, 2, 4]
            mock_tracer.start_span.assert_called_once()
            mock_span.end.assert_called_once()
        finally:
            del sys.modules["exporter.ar_trace.trace_exporter"]

    @pytest.mark.asyncio
    @patch("app.utils.observability.trace_wrapper.TELEMETRY_SDK_AVAILABLE", True)
    @patch("app.common.config.Config")
    @patch("app.utils.observability.trace_wrapper.func_judgment")
    async def test_async_generator_exception(self, m_func_judgment, m_config):
        """测试异步生成器异常处理"""
        m_config.is_o11y_trace_enabled.return_value = True
        m_func_judgment.return_value = (True, True)

        mock_tracer = MagicMock()
        mock_span = MagicMock()
        mock_tracer.start_span.return_value = mock_span
        mock_span.is_recording.return_value = True

        mock_exporter_module = MagicMock()
        mock_exporter_module.tracer = mock_tracer
        sys.modules["exporter.ar_trace.trace_exporter"] = mock_exporter_module

        try:
            from app.utils.observability.trace_wrapper import internal_span

            @internal_span()
            async def test_func(span=None):
                yield 1
                raise ValueError("generator error")

            with pytest.raises(ValueError, match="generator error"):
                async for _ in test_func():
                    pass

            mock_span.set_status.assert_called()
            mock_span.record_exception.assert_called_once()
        finally:
            del sys.modules["exporter.ar_trace.trace_exporter"]

    @patch("app.utils.observability.trace_wrapper.TELEMETRY_SDK_AVAILABLE", True)
    @patch("app.common.config.Config")
    @patch("app.utils.observability.trace_wrapper.func_judgment")
    def test_uses_function_name_when_name_not_provided(self, m_func_judgment, m_config):
        """测试未提供名称时使用函数名"""
        m_config.is_o11y_trace_enabled.return_value = True
        m_func_judgment.return_value = (False, False)

        mock_tracer = MagicMock()
        mock_span = MagicMock()
        mock_tracer.start_span.return_value = mock_span
        mock_span.is_recording.return_value = True

        mock_exporter_module = MagicMock()
        mock_exporter_module.tracer = mock_tracer
        sys.modules["exporter.ar_trace.trace_exporter"] = mock_exporter_module

        try:
            from app.utils.observability.trace_wrapper import internal_span

            @internal_span()
            def my_custom_function(span=None):
                return "result"

            my_custom_function()

            # 验证使用了函数名作为 span 名称
            mock_tracer.start_span.assert_called_once()
            call_args = mock_tracer.start_span.call_args
            assert call_args[0][0] == "my_custom_function"
        finally:
            del sys.modules["exporter.ar_trace.trace_exporter"]

    @patch("app.utils.observability.trace_wrapper.TELEMETRY_SDK_AVAILABLE", True)
    @patch("app.common.config.Config")
    @patch("app.utils.observability.trace_wrapper.func_judgment")
    def test_sync_function_with_empty_attributes(self, m_func_judgment, m_config):
        """测试空属性字典的同步函数"""
        m_config.is_o11y_trace_enabled.return_value = True
        m_func_judgment.return_value = (False, False)

        mock_tracer = MagicMock()
        mock_span = MagicMock()
        mock_tracer.start_span.return_value = mock_span
        mock_span.is_recording.return_value = True

        mock_exporter_module = MagicMock()
        mock_exporter_module.tracer = mock_tracer
        sys.modules["exporter.ar_trace.trace_exporter"] = mock_exporter_module

        try:
            from app.utils.observability.trace_wrapper import internal_span

            @internal_span(name="test_span", attributes={})
            def test_func(span=None):
                return "result"

            result = test_func()
            assert result == "result"
            mock_tracer.start_span.assert_called_once()
        finally:
            del sys.modules["exporter.ar_trace.trace_exporter"]

    @patch("app.utils.observability.trace_wrapper.TELEMETRY_SDK_AVAILABLE", True)
    @patch("app.common.config.Config")
    @patch("app.utils.observability.trace_wrapper.func_judgment")
    def test_sync_function_when_span_not_recording(self, m_func_judgment, m_config):
        """测试span不记录时同步函数异常处理"""
        m_config.is_o11y_trace_enabled.return_value = True
        m_func_judgment.return_value = (False, False)

        mock_tracer = MagicMock()
        mock_span = MagicMock()
        mock_tracer.start_span.return_value = mock_span
        mock_span.is_recording.return_value = False

        mock_exporter_module = MagicMock()
        mock_exporter_module.tracer = mock_tracer
        sys.modules["exporter.ar_trace.trace_exporter"] = mock_exporter_module

        try:
            from app.utils.observability.trace_wrapper import internal_span

            @internal_span()
            def test_func(span=None):
                raise ValueError("test error")

            with pytest.raises(ValueError, match="test error"):
                test_func()

            # 当 span 不记录时，不应该调用这些方法
            mock_span.record_exception.assert_not_called()
        finally:
            del sys.modules["exporter.ar_trace.trace_exporter"]

    @pytest.mark.asyncio
    @patch("app.utils.observability.trace_wrapper.TELEMETRY_SDK_AVAILABLE", True)
    @patch("app.common.config.Config")
    @patch("app.utils.observability.trace_wrapper.func_judgment")
    async def test_async_function_with_empty_attributes(
        self, m_func_judgment, m_config
    ):
        """测试空属性字典的异步函数"""
        m_config.is_o11y_trace_enabled.return_value = True
        m_func_judgment.return_value = (True, False)

        mock_tracer = MagicMock()
        mock_span = MagicMock()
        mock_tracer.start_span.return_value = mock_span
        mock_span.is_recording.return_value = True

        mock_exporter_module = MagicMock()
        mock_exporter_module.tracer = mock_tracer
        sys.modules["exporter.ar_trace.trace_exporter"] = mock_exporter_module

        try:
            from app.utils.observability.trace_wrapper import internal_span

            @internal_span(name="async_span", attributes={})
            async def test_func(span=None):
                return "result"

            result = await test_func()
            assert result == "result"
        finally:
            del sys.modules["exporter.ar_trace.trace_exporter"]

    @pytest.mark.asyncio
    @patch("app.utils.observability.trace_wrapper.TELEMETRY_SDK_AVAILABLE", True)
    @patch("app.common.config.Config")
    @patch("app.utils.observability.trace_wrapper.func_judgment")
    async def test_async_function_when_span_not_recording(
        self, m_func_judgment, m_config
    ):
        """测试span不记录时异步函数异常处理"""
        m_config.is_o11y_trace_enabled.return_value = True
        m_func_judgment.return_value = (True, False)

        mock_tracer = MagicMock()
        mock_span = MagicMock()
        mock_tracer.start_span.return_value = mock_span
        mock_span.is_recording.return_value = False

        mock_exporter_module = MagicMock()
        mock_exporter_module.tracer = mock_tracer
        sys.modules["exporter.ar_trace.trace_exporter"] = mock_exporter_module

        try:
            from app.utils.observability.trace_wrapper import internal_span

            @internal_span()
            async def test_func(span=None):
                raise RuntimeError("async error")

            with pytest.raises(RuntimeError, match="async error"):
                await test_func()

            mock_span.record_exception.assert_not_called()
        finally:
            del sys.modules["exporter.ar_trace.trace_exporter"]
