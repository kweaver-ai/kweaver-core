# -*-coding:utf-8-*-
from opentelemetry.trace import SpanKind
from functools import wraps
from typing import Optional, Callable, AsyncGenerator, Any, Awaitable

from opentelemetry.trace import Status, StatusCode

from app.utils.observability.sdk_available import TELEMETRY_SDK_AVAILABLE
from app.utils.common import func_judgment


def internal_span(
    name: Optional[str] = None,
    attributes: Optional[dict] = None,
) -> Callable:
    """
    创建一个用于自动生成 OpenTelemetry SpanKind.INTERNAL 类型 span 的注解

    参数:
        name: span 的名称，如果未提供则使用被注解函数的名称
        attributes: 要添加到 span 的属性字典

    返回:
        包装后的函数
    """

    def decorator(func: Callable) -> Callable:
        # 延迟导入 Config 避免循环依赖
        from app.common.config import Config

        # 如果 SDK 不可用或追踪未启用，直接返回原函数
        if not TELEMETRY_SDK_AVAILABLE or not Config.is_o11y_trace_enabled():
            return func

        from exporter.ar_trace.trace_exporter import tracer

        # 设置 span 名称（如果未提供则使用函数名）
        span_name = name or func.__name__
        is_async, is_stream = func_judgment(func)
        # if asyncio.iscoroutinefunction(func):
        if is_async and is_stream:
            # 异步生成器函数处理
            @wraps(func)
            async def async_generator_wrapper(
                *args, **kwargs
            ) -> AsyncGenerator[Any, Any]:
                from opentelemetry.trace import use_span
                from app.common.stand_log import StandLogger

                span = tracer.start_span(
                    span_name, kind=SpanKind.INTERNAL, attributes=attributes
                )

                span_context = span.get_span_context()
                StandLogger.info_log(
                    f"[trace_wrapper] Created span '{span_name}', trace_id={format(span_context.trace_id, '032x')}, span_id={format(span_context.span_id, '016x')}"
                )

                # 将span设置为当前上下文，使得子span可以自动关联
                with use_span(span, end_on_exit=False):
                    try:
                        span.set_status(Status(StatusCode.OK))
                        span.set_attribute("error", False)

                        kwargs["span"] = span

                        result = func(*args, **kwargs)
                        async for item in result:
                            yield item

                    except Exception as e:
                        if span.is_recording():
                            span.set_status(Status(StatusCode.ERROR))
                            span.set_attribute("error", True)
                            span.record_exception(e)
                        raise
                    finally:
                        StandLogger.info_log(
                            f"[trace_wrapper] Ending span '{span_name}'"
                        )
                        span.end()

            return async_generator_wrapper
        elif is_async:
            # 异步函数处理
            @wraps(func)
            async def async_wrapper(*args, **kwargs) -> Awaitable[Any]:
                from opentelemetry.trace import use_span
                from app.common.stand_log import StandLogger

                span = tracer.start_span(
                    span_name, kind=SpanKind.INTERNAL, attributes=attributes
                )

                span_context = span.get_span_context()
                StandLogger.info_log(
                    f"[trace_wrapper] Created span '{span_name}', trace_id={format(span_context.trace_id, '032x')}, span_id={format(span_context.span_id, '016x')}"
                )

                # 将span设置为当前上下文，使得子span可以自动关联
                with use_span(span, end_on_exit=False):
                    try:
                        span.set_status(Status(StatusCode.OK))
                        span.set_attribute("error", False)

                        kwargs["span"] = span

                        result = await func(*args, **kwargs)
                        return result

                    except Exception as e:
                        if span.is_recording():
                            span.set_status(Status(StatusCode.ERROR))
                            span.set_attribute("error", True)
                            span.record_exception(e)
                        raise
                    finally:
                        StandLogger.info_log(
                            f"[trace_wrapper] Ending span '{span_name}'"
                        )
                        span.end()

            return async_wrapper
        else:

            @wraps(func)
            def sync_wrapper(*args, **kwargs) -> Any:
                from opentelemetry.trace import use_span
                from app.common.stand_log import StandLogger

                # 创建 INTERNAL 类型的 span
                span = tracer.start_span(
                    span_name, kind=SpanKind.INTERNAL, attributes=attributes
                )

                span_context = span.get_span_context()
                StandLogger.info_log(
                    f"[trace_wrapper] Created span '{span_name}', trace_id={format(span_context.trace_id, '032x')}, span_id={format(span_context.span_id, '016x')}"
                )

                # 将span设置为当前上下文，使得子span可以自动关联
                with use_span(span, end_on_exit=False):
                    try:
                        kwargs["span"] = span
                        # 执行被注解的函数
                        result = func(*args, **kwargs)
                        span.set_status(Status(StatusCode.OK))
                        return result
                    except Exception as e:
                        # 记录异常信息
                        if span.is_recording():
                            span.set_status(Status(StatusCode.ERROR))
                            span.set_attribute("error", True)
                            span.record_exception(e)
                        raise  # 重新抛出异常，不影响原有逻辑
                    finally:
                        StandLogger.info_log(
                            f"[trace_wrapper] Ending span '{span_name}'"
                        )
                        span.end()

            return sync_wrapper

    return decorator
