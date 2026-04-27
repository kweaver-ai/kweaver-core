# -*- coding:utf-8 -*-
"""单元测试 - streaming_rate_limiter 流式速率限制器"""

import pytest
import asyncio
import time


class TestStreamingRateLimiter_Initial:
    """测试 StreamingRateLimiter 初始化"""

    def test_init_with_default_rate(self):
        """测试使用默认速率初始化"""
        from app.router.middleware_pkg.streaming_rate_limiter import (
            StreamingRateLimiter,
            DEFAULT_RATE_LIMIT,
        )

        limiter = StreamingRateLimiter()

        assert limiter.rate_limit == DEFAULT_RATE_LIMIT
        assert limiter.min_interval == 1.0 / DEFAULT_RATE_LIMIT
        assert limiter.last_yield_time == 0

    def test_init_with_custom_rate(self):
        """测试使用自定义速率初始化"""
        from app.router.middleware_pkg.streaming_rate_limiter import (
            StreamingRateLimiter,
        )

        limiter = StreamingRateLimiter(rate_limit=10)

        assert limiter.rate_limit == 10
        assert limiter.min_interval == 0.1
        assert limiter.last_yield_time == 0

    def test_init_with_zero_rate(self):
        """测试速率为0时被限制为1"""
        from app.router.middleware_pkg.streaming_rate_limiter import (
            StreamingRateLimiter,
        )

        limiter = StreamingRateLimiter(rate_limit=0)

        assert limiter.rate_limit == 1
        assert limiter.min_interval == 1.0

    def test_init_with_negative_rate(self):
        """测试负速率被限制为1"""
        from app.router.middleware_pkg.streaming_rate_limiter import (
            StreamingRateLimiter,
        )

        limiter = StreamingRateLimiter(rate_limit=-10)

        assert limiter.rate_limit == 1
        assert limiter.min_interval == 1.0


@pytest.mark.asyncio
class TestStreamingRateLimiterLimitRate:
    """测试 limit_rate 函数"""

    async def test_first_call_returns_immediately(self):
        """测试第一次调用立即返回"""
        from app.router.middleware_pkg.streaming_rate_limiter import (
            StreamingRateLimiter,
        )

        limiter = StreamingRateLimiter(rate_limit=10)

        start = time.time()
        await limiter.limit_rate()
        elapsed = time.time() - start

        # First call should return immediately (< 0.01s)
        assert elapsed < 0.01

    async def test_consecutive_calls_respect_rate_limit(self):
        """测试连续调用遵守速率限制"""
        from app.router.middleware_pkg.streaming_rate_limiter import (
            StreamingRateLimiter,
        )

        limiter = StreamingRateLimiter(rate_limit=10)  # 0.1s interval

        start = time.time()
        await limiter.limit_rate()
        await limiter.limit_rate()
        elapsed = time.time() - start

        # Two calls should take at least 0.1s
        assert elapsed >= 0.09  # Allow small timing variance

    async def test_rate_limiting_with_high_rate(self):
        """测试高速率限制"""
        from app.router.middleware_pkg.streaming_rate_limiter import (
            StreamingRateLimiter,
        )

        limiter = StreamingRateLimiter(rate_limit=100)  # 0.01s interval

        start = time.time()
        for _ in range(10):
            await limiter.limit_rate()
        elapsed = time.time() - start

        # 10 calls at 100/sec should take ~0.09s
        assert elapsed >= 0.08


@pytest.mark.asyncio
class TestRateLimitedStreamingIterator:
    """测试 RateLimitedStreamingIterator 类"""

    @pytest.mark.asyncio
    def test_init_preserves_iterator(self):
        """测试初始化保留原始迭代器"""
        from app.router.middleware_pkg.streaming_rate_limiter import (
            RateLimitedStreamingIterator,
        )

        async def mock_iterator():
            yield b"data"
            yield b"chunk2"

        original_iterator = mock_iterator()

        iterator = RateLimitedStreamingIterator(original_iterator)

        assert iterator.original_iterator is original_iterator
        assert iterator.current_chunk_index == 0
        assert isinstance(iterator.rate_limiter, object)

    async def test_iter_returns_self(self):
        """测试__aiter__返回自身"""
        from app.router.middleware_pkg.streaming_rate_limiter import (
            RateLimitedStreamingIterator,
        )

        async def mock_iterator():
            yield b"data"

        iterator = RateLimitedStreamingIterator(mock_iterator())

        assert iterator.__aiter__() is iterator

    async def test_anext_without_rate_limit_first_10(self):
        """测试前10个块不限制"""
        from app.router.middleware_pkg.streaming_rate_limiter import (
            RateLimitedStreamingIterator,
        )

        chunks = [b"chunk%d" % i for i in range(10)]

        async def mock_iterator():
            for chunk in chunks:
                yield chunk

        iterator = RateLimitedStreamingIterator(mock_iterator())
        results = []
        start = time.time()

        async for chunk in iterator:
            results.append(chunk)

        elapsed = time.time() - start

        # First 10 should have minimal delay
        assert elapsed < 0.1
        # All chunks should be returned
        assert len(results) == 10
        assert results == chunks

    async def test_anext_with_rate_limit_after_10(self):
        """测试第10个块之后应用速率限制"""
        from app.router.middleware_pkg.streaming_rate_limiter import (
            RateLimitedStreamingIterator,
        )

        chunks = [b"chunk%d" % i for i in range(15)]

        async def mock_iterator():
            for chunk in chunks:
                yield chunk

        iterator = RateLimitedStreamingIterator(mock_iterator())
        results = []
        start = time.time()

        async for chunk in iterator:
            results.append(chunk)

        elapsed = time.time() - start

        # 15 chunks with rate limiting should take some time
        # First 10 are fast, last 5 are rate limited
        assert len(results) == 15
        # Should take at least some time due to rate limiting
        assert elapsed > 0.01

    async def test_empty_iterator(self):
        """测试空迭代器处理"""
        from app.router.middleware_pkg.streaming_rate_limiter import (
            RateLimitedStreamingIterator,
        )

        async def mock_iterator():
            return  # Never yields
            yield  # pragma: no cover

        iterator = RateLimitedStreamingIterator(mock_iterator())
        results = []

        try:
            async for chunk in iterator:
                results.append(chunk)
        except StopAsyncIteration:
            pass

        # Should not yield any chunks
        assert results == []

    async def test_should_rate_limit(self):
        """测试_should_rate_limit方法"""
        from app.router.middleware_pkg.streaming_rate_limiter import (
            RateLimitedStreamingIterator,
        )

        async def mock_iterator():
            yield b"data"

        iterator = RateLimitedStreamingIterator(mock_iterator())

        # First 10 should not rate limit
        for i in range(10):
            iterator.current_chunk_index = i
            assert iterator._should_rate_limit() is False

        # After 10 should rate limit
        for i in range(11, 20):
            iterator.current_chunk_index = i
            assert iterator._should_rate_limit() is True


@pytest.mark.asyncio
class TestCreateRateLimitedIterator:
    """测试 create_rate_limited_iterator 函数"""

    def test_returns_rate_limited_iterator(self):
        """测试返回速率限制迭代器"""
        from app.router.middleware_pkg.streaming_rate_limiter import (
            create_rate_limited_iterator,
            RateLimitedStreamingIterator,
        )

        async def mock_iterator():
            yield b"data"

        iterator = create_rate_limited_iterator(mock_iterator())

        assert isinstance(iterator, RateLimitedStreamingIterator)

    async def test_preserves_original_iterator(self):
        """测试保留原始迭代器"""
        from app.router.middleware_pkg.streaming_rate_limiter import (
            create_rate_limited_iterator,
        )

        async def mock_iterator():
            yield b"data"
            yield b"chunk2"

        original = mock_iterator()
        iterator = create_rate_limited_iterator(original)

        assert iterator.original_iterator is original

    async def test_full_integration(self):
        """测试完整集成流程"""
        from app.router.middleware_pkg.streaming_rate_limiter import (
            create_rate_limited_iterator,
        )

        chunks = [b"test%d" % i for i in range(5)]

        async def mock_iterator():
            for chunk in chunks:
                yield chunk

        iterator = create_rate_limited_iterator(mock_iterator())
        results = []

        async for chunk in iterator:
            results.append(chunk)

        assert results == chunks


@pytest.mark.asyncio
class TestStreamingRateLimiterEdgeCases:
    """测试边界情况"""

    async def test_very_high_rate(self):
        """测试非常高的速率限制"""
        from app.router.middleware_pkg.streaming_rate_limiter import (
            StreamingRateLimiter,
        )

        limiter = StreamingRateLimiter(rate_limit=1000)

        assert limiter.rate_limit == 1000
        assert limiter.min_interval == 0.001

    async def test_exact_timing(self):
        """测试精确的速率限制时间"""
        from app.router.middleware_pkg.streaming_rate_limiter import (
            StreamingRateLimiter,
        )

        limiter = StreamingRateLimiter(rate_limit=10)  # 0.1s per chunk

        start = time.time()
        await limiter.limit_rate()
        mid = time.time()
        await limiter.limit_rate()
        end = time.time()

        first_call_time = mid - start
        second_call_time = end - mid

        # First call should be immediate
        assert first_call_time < 0.01
        # Second call should wait ~0.1s
        assert 0.08 < second_call_time < 0.2

    async def test_concurrent_iterators_independent(self):
        """测试并发迭代器互不干扰"""
        from app.router.middleware_pkg.streaming_rate_limiter import (
            RateLimitedStreamingIterator,
        )

        async def mock_iterator(name):
            for i in range(5):
                yield f"{name}_{i}".encode()

        iterator1 = RateLimitedStreamingIterator(mock_iterator("iter1"))
        iterator2 = RateLimitedStreamingIterator(mock_iterator("iter2"))

        results1 = []
        results2 = []

        # Consume both concurrently
        async def consume1():
            async for chunk in iterator1:
                results1.append(chunk)

        async def consume2():
            async for chunk in iterator2:
                results2.append(chunk)

        await asyncio.gather(consume1(), consume2())

        assert len(results1) == 5
        assert len(results2) == 5
        assert results1[0].startswith(b"iter1")
        assert results2[0].startswith(b"iter2")
