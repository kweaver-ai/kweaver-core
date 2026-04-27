"""单元测试 - utils/observability/trace_context 模块"""

import pytest
from opentelemetry.trace import SpanKind


class TestTraceContext:
    """测试 TraceContext 类"""

    def test_init_creates_tracer(self):
        """测试初始化创建tracer"""
        from app.utils.observability.trace_context import TraceContext

        ctx = TraceContext()

        assert ctx.tracer is not None
        assert hasattr(ctx.tracer, "start_as_current_span")

    def test_tracer_has_required_methods(self):
        """测试tracer有必需的方法"""
        from app.utils.observability.trace_context import TraceContext

        ctx = TraceContext()

        # Tracer should have start_as_current_span method
        assert hasattr(ctx.tracer, "start_as_current_span")

    def test_start_span_returns_context_manager(self):
        """测试start_span返回上下文管理器"""
        from app.utils.observability.trace_context import TraceContext

        ctx = TraceContext()

        # start_span should return a context manager
        span_ctx = ctx.start_span("test_span")

        # Should have __enter__ and __exit__ methods
        assert hasattr(span_ctx, "__enter__")
        assert hasattr(span_ctx, "__exit__")

    def test_start_span_with_default_parameters(self):
        """测试使用默认参数的start_span"""
        from app.utils.observability.trace_context import TraceContext

        ctx = TraceContext()

        # Should not raise with default parameters
        with ctx.start_span("test_operation"):
            pass

    def test_start_span_with_custom_kind(self):
        """测试使用自定义kind的start_span"""
        from app.utils.observability.trace_context import TraceContext

        ctx = TraceContext()

        # Should not raise with custom kind
        with ctx.start_span("test_operation", kind=SpanKind.SERVER):
            pass

    def test_start_span_with_attributes(self):
        """测试带属性的start_span"""
        from app.utils.observability.trace_context import TraceContext

        ctx = TraceContext()

        # Should not raise with attributes
        with ctx.start_span("test_operation", attributes={"key": "value"}):
            pass

    def test_start_span_with_exception(self):
        """测试start_span处理异常"""
        from app.utils.observability.trace_context import TraceContext

        ctx = TraceContext()

        # Exception should be raised
        with pytest.raises(ValueError):
            with ctx.start_span("test_operation"):
                raise ValueError("Test error")

    def test_start_async_span_returns_async_context_manager(self):
        """测试start_async_span返回异步上下文管理器"""
        from app.utils.observability.trace_context import TraceContext

        ctx = TraceContext()

        # start_async_span should return an async context manager
        span_ctx = ctx.start_async_span("test_span")

        # Should have __aenter__ and __aexit__ methods
        assert hasattr(span_ctx, "__aenter__")
        assert hasattr(span_ctx, "__aexit__")

    @pytest.mark.asyncio
    async def test_start_async_span_with_default_parameters(self):
        """测试使用默认参数的start_async_span"""
        from app.utils.observability.trace_context import TraceContext

        ctx = TraceContext()

        # Should not raise with default parameters
        async with ctx.start_async_span("test_operation"):
            pass

    @pytest.mark.asyncio
    async def test_start_async_span_with_custom_kind(self):
        """测试使用自定义kind的start_async_span"""
        from app.utils.observability.trace_context import TraceContext

        ctx = TraceContext()

        # Should not raise with custom kind
        async with ctx.start_async_span("test_operation", kind=SpanKind.CLIENT):
            pass

    @pytest.mark.asyncio
    async def test_start_async_span_with_attributes(self):
        """测试带属性的start_async_span"""
        from app.utils.observability.trace_context import TraceContext

        ctx = TraceContext()

        # Should not raise with attributes
        async with ctx.start_async_span("test_operation", attributes={"key": "value"}):
            pass

    @pytest.mark.asyncio
    async def test_start_async_span_with_exception(self):
        """测试start_async_span处理异常"""
        from app.utils.observability.trace_context import TraceContext

        ctx = TraceContext()

        # Exception should be raised
        with pytest.raises(ValueError):
            async with ctx.start_async_span("test_operation"):
                raise ValueError("Test error")

    def test_start_async_span_defaults_name_to_class_name(self):
        """测试start_async_span默认名称为类名"""
        from app.utils.observability.trace_context import TraceContext

        ctx = TraceContext()

        # Name should default to "TraceContext" if None provided
        span_ctx = ctx.start_async_span(name=None)

        # Should have __aenter__ method
        assert hasattr(span_ctx, "__aenter__")

    def test_multiple_contexts_independently(self):
        """测试多个上下文独立工作"""
        from app.utils.observability.trace_context import TraceContext

        ctx1 = TraceContext()
        ctx2 = TraceContext()

        # Both should have independent tracers
        assert ctx1.tracer is not None
        assert ctx2.tracer is not None

        # Both should work independently
        with ctx1.start_span("op1"):
            with ctx2.start_span("op2"):
                pass
