# -*-coding:utf-8-*-
import asyncio
from opentelemetry import trace
from opentelemetry.trace import SpanKind, Tracer
from functools import wraps
from typing import Optional, Callable, Any, Awaitable
from exporter.ar_trace.trace_exporter import tracer
from opentelemetry.trace import SpanKind, Tracer, Status, StatusCode
from contextlib import contextmanager, asynccontextmanager
from typing import Optional, Any, AsyncGenerator, Generator



class TraceContext:
    def __init__(self) -> None:
        self.tracer: Tracer = tracer


    @contextmanager
    def start_span(
        self,
        name: str,
        kind: SpanKind = SpanKind.INTERNAL,
        attributes: Optional[dict] = None,
        set_status_on_exception: bool = True,
        status_code_on_exception: StatusCode = StatusCode.ERROR,
        success_status_code: Optional[StatusCode] = StatusCode.OK
    ) -> Generator[trace.Span, None, None]:
        """同步上下文管理器"""
        with self.tracer.start_as_current_span(
            name,
            kind=kind,
            attributes=attributes
        ) as span:
            try:
                yield span
                if success_status_code and span.is_recording():
                    span.set_status(Status(success_status_code))
            except Exception as e:
                if span.is_recording():
                    if set_status_on_exception:
                        span.set_status(Status(status_code_on_exception))
                    span.set_attribute("error", True)
                    span.record_exception(e)
                raise
    
    @asynccontextmanager
    async def start_async_span(
        self,
        name: str = None,
        kind: SpanKind = SpanKind.INTERNAL,
        attributes: Optional[dict] = None,
        set_status_on_exception: bool = True,
        status_code_on_exception: StatusCode = StatusCode.ERROR,
        success_status_code: Optional[StatusCode] = StatusCode.OK
    ) -> AsyncGenerator[trace.Span, None]:
        name = name or self.__class__.__name__
        """异步上下文管理器"""
        with self.tracer.start_as_current_span(
            name,
            kind=kind,
            attributes=attributes
        ) as span:
            try:
                yield span
                if success_status_code and span.is_recording():
                    span.set_status(Status(success_status_code))
            except Exception as e:
                if span.is_recording():
                    if set_status_on_exception:
                        span.set_status(Status(status_code_on_exception))
                    span.set_attribute("error", True)
                    span.record_exception(e)
                raise
