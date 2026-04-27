"""
流式响应速率限制器
"""

import asyncio
import time
from typing import AsyncIterator
from collections.abc import AsyncIterator as ABCAsyncIterator

# 默认速率限制：每秒20个chunk
DEFAULT_RATE_LIMIT = 20


class StreamingRateLimiter:
    """流式响应速率限制器"""

    def __init__(self, rate_limit: int = DEFAULT_RATE_LIMIT):
        self.rate_limit = max(rate_limit, 1)  # 至少每秒1个
        self.min_interval = 1.0 / self.rate_limit
        self.last_yield_time = 0

    async def limit_rate(self):
        """执行速率限制"""
        current_time = time.time()
        elapsed = current_time - self.last_yield_time

        if elapsed < self.min_interval:
            await asyncio.sleep(self.min_interval - elapsed)

        self.last_yield_time = time.time()


class RateLimitedStreamingIterator(ABCAsyncIterator[bytes]):
    """带速率限制的流式迭代器"""

    def __init__(self, original_iterator: AsyncIterator[bytes]):
        self.original_iterator = original_iterator
        self.current_chunk_index = 0
        self.rate_limiter = StreamingRateLimiter(rate_limit=DEFAULT_RATE_LIMIT)

    def __aiter__(self):
        return self

    async def __anext__(self):
        chunk = await self.original_iterator.__anext__()
        self.current_chunk_index += 1

        if self._should_rate_limit():
            await self.rate_limiter.limit_rate()

        return chunk

    def _should_rate_limit(self) -> bool:
        """判断是否应该进行速率限制：前10个不限制，后面的限制"""
        return self.current_chunk_index > 10


def create_rate_limited_iterator(
    original_iterator: AsyncIterator[bytes],
) -> AsyncIterator[bytes]:
    """创建带速率限制的迭代器"""
    return RateLimitedStreamingIterator(original_iterator)
