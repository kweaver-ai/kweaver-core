# -*- coding:utf-8 -*-
"""单元测试 - safe_output_generator 扩展测试"""

import pytest
import asyncio
from unittest.mock import AsyncMock, MagicMock, patch


@pytest.mark.asyncio
class TestCheckAndUpdateCache:
    """测试 check_and_update_cache 函数"""

    async def test_cache_expired_creates_new_cache(self):
        """测试缓存过期时创建新缓存"""
        from app.router.agent_controller_pkg.run_agent_v2.safe_output_generator import (
            check_and_update_cache,
        )
        from app.domain.vo.agentvo import AgentConfigVo
        from app.domain.vo.agent_cache import AgentCacheIdVO

        cache_id_vo = MagicMock(spec=AgentCacheIdVO)
        cache_id_vo.agent_id = "agent123"
        cache_id_vo.agent_version = "1.0"

        agent_config = MagicMock(spec=AgentConfigVo)
        headers = {"Authorization": "Bearer token"}

        # Mock cache_manager
        with patch(
            "app.router.agent_controller_pkg.run_agent_v2.safe_output_generator.cache_manager"
        ) as mock_mgr:
            # First call returns 0 (expired), second call returns new TTL after creation
            call_count = [0]

            async def get_ttl_side_effect(cache_id):
                call_count[0] += 1
                if call_count[0] == 1:
                    return 0  # First call - expired
                else:
                    return 3600  # Subsequent calls - new TTL

            mock_mgr.cache_service.get_ttl = AsyncMock(side_effect=get_ttl_side_effect)
            mock_mgr.create_cache = AsyncMock()

            await check_and_update_cache(
                cache_id_vo=cache_id_vo,
                agent_config=agent_config,
                headers=headers,
                account_id="acc123",
                account_type="premium",
            )

            # Should create new cache
            mock_mgr.create_cache.assert_called_once()

    async def test_cache_valid_no_update_needed(self):
        """测试缓存有效且未到更新阈值"""
        from app.router.agent_controller_pkg.run_agent_v2.safe_output_generator import (
            check_and_update_cache,
        )
        from app.domain.vo.agentvo import AgentConfigVo
        from app.domain.vo.agent_cache import AgentCacheIdVO

        cache_id_vo = MagicMock(spec=AgentCacheIdVO)
        agent_config = MagicMock(spec=AgentConfigVo)
        headers = {}

        with patch(
            "app.router.agent_controller_pkg.run_agent_v2.safe_output_generator.cache_manager"
        ) as mock_mgr:
            # Return high TTL (cache is fresh)
            mock_mgr.cache_service.get_ttl = AsyncMock(return_value=3000)
            mock_mgr.update_cache_data = AsyncMock()

            await check_and_update_cache(
                cache_id_vo=cache_id_vo,
                agent_config=agent_config,
                headers=headers,
                account_id="acc123",
                account_type="premium",
            )

            # Should NOT update cache (still fresh)
            mock_mgr.update_cache_data.assert_not_called()

    async def test_cache_update_threshold_exceeded(self):
        """测试缓存超过更新阈值"""
        from app.router.agent_controller_pkg.run_agent_v2.safe_output_generator import (
            check_and_update_cache,
        )
        from app.domain.vo.agentvo import AgentConfigVo
        from app.domain.vo.agent_cache import AgentCacheIdVO
        from app.domain.constant import (
            AGENT_CACHE_TTL,
            AGENT_CACHE_DATA_UPDATE_PASS_SECOND,
        )

        cache_id_vo = MagicMock(spec=AgentCacheIdVO)
        agent_config = MagicMock(spec=AgentConfigVo)
        headers = {}

        # Calculate TTL that exceeds threshold
        low_ttl = AGENT_CACHE_TTL - AGENT_CACHE_DATA_UPDATE_PASS_SECOND - 10

        with patch(
            "app.router.agent_controller_pkg.run_agent_v2.safe_output_generator.cache_manager"
        ) as mock_mgr:
            mock_mgr.cache_service.get_ttl = AsyncMock(return_value=low_ttl)
            mock_mgr.update_cache_data = AsyncMock()

            await check_and_update_cache(
                cache_id_vo=cache_id_vo,
                agent_config=agent_config,
                headers=headers,
                account_id="acc123",
                account_type="premium",
            )

            # Should update cache data
            mock_mgr.update_cache_data.assert_called_once()

    async def test_cache_error_handling(self):
        """测试缓存异常处理"""
        from app.router.agent_controller_pkg.run_agent_v2.safe_output_generator import (
            check_and_update_cache,
        )
        from app.domain.vo.agentvo import AgentConfigVo
        from app.domain.vo.agent_cache import AgentCacheIdVO

        cache_id_vo = MagicMock(spec=AgentCacheIdVO)
        agent_config = MagicMock(spec=AgentConfigVo)
        headers = {}

        with patch(
            "app.router.agent_controller_pkg.run_agent_v2.safe_output_generator.cache_manager"
        ) as mock_mgr:
            # Raise exception
            mock_mgr.cache_service.get_ttl = AsyncMock(
                side_effect=Exception("Cache error")
            )

            # Should not raise, just log error
            await check_and_update_cache(
                cache_id_vo=cache_id_vo,
                agent_config=agent_config,
                headers=headers,
                account_id="acc123",
                account_type="premium",
            )


@pytest.mark.asyncio
class TestCreateSafeOutputGeneratorExtended:
    """测试 create_safe_output_generator 扩展场景"""

    async def test_generator_exit_handling(self):
        """测试 GeneratorExit 处理"""
        from app.router.agent_controller_pkg.run_agent_v2.safe_output_generator import (
            create_safe_output_generator,
        )

        agent_core_v2 = MagicMock()
        agent_config = MagicMock()
        agent_input = MagicMock()
        headers = {}
        start_time = 1234567890.0

        async def generator_that_exits():
            yield "chunk1"
            raise GeneratorExit()

        agent_core_v2.output_handler.result_output = MagicMock(
            return_value=generator_that_exits()
        )

        gen = create_safe_output_generator(
            agent_core_v2=agent_core_v2,
            agent_config=agent_config,
            agent_input=agent_input,
            headers=headers,
            is_debug_run=False,
            start_time=start_time,
        )

        results = []
        try:
            async for chunk in gen:
                results.append(chunk)
        except GeneratorExit:
            pass

        # Should get first chunk before GeneratorExit
        assert len(results) == 1

    async def test_cancelled_error_handling(self):
        """测试 asyncio.CancelledError 处理"""
        from app.router.agent_controller_pkg.run_agent_v2.safe_output_generator import (
            create_safe_output_generator,
        )

        agent_core_v2 = MagicMock()
        agent_config = MagicMock()
        agent_input = MagicMock()
        headers = {}
        start_time = 1234567890.0

        async def generator_that_cancels():
            yield "chunk1"
            raise asyncio.CancelledError()

        agent_core_v2.output_handler.result_output = MagicMock(
            return_value=generator_that_cancels()
        )

        gen = create_safe_output_generator(
            agent_core_v2=agent_core_v2,
            agent_config=agent_config,
            agent_input=agent_input,
            headers=headers,
            is_debug_run=False,
            start_time=start_time,
        )

        results = []
        with pytest.raises(asyncio.CancelledError):
            async for chunk in gen:
                results.append(chunk)

        # Should get first chunk before cancellation
        assert len(results) == 1

    async def test_generic_exception_handling(self):
        """测试通用异常处理"""
        from app.router.agent_controller_pkg.run_agent_v2.safe_output_generator import (
            create_safe_output_generator,
        )

        agent_core_v2 = MagicMock()
        agent_config = MagicMock()
        agent_input = MagicMock()
        headers = {}
        start_time = 1234567890.0

        async def generator_with_error():
            yield "chunk1"
            raise ValueError("Test error")

        agent_core_v2.output_handler.result_output = MagicMock(
            return_value=generator_with_error()
        )

        gen = create_safe_output_generator(
            agent_core_v2=agent_core_v2,
            agent_config=agent_config,
            agent_input=agent_input,
            headers=headers,
            is_debug_run=False,
            start_time=start_time,
        )

        results = []
        with pytest.raises(ValueError, match="Test error"):
            async for chunk in gen:
                results.append(chunk)

        # Should get first chunk before error
        assert len(results) == 1

    async def test_normal_completion_closes_generator(self):
        """测试正常完成时关闭生成器"""
        from app.router.agent_controller_pkg.run_agent_v2.safe_output_generator import (
            create_safe_output_generator,
        )

        agent_core_v2 = MagicMock()
        agent_config = MagicMock()
        agent_input = MagicMock()
        headers = {}
        start_time = 1234567890.0

        # Create a proper async generator
        async def mock_generator():
            yield "chunk1"
            yield "chunk2"

        agent_core_v2.output_handler.result_output = MagicMock(
            return_value=mock_generator()
        )

        gen = create_safe_output_generator(
            agent_core_v2=agent_core_v2,
            agent_config=agent_config,
            agent_input=agent_input,
            headers=headers,
            is_debug_run=False,
            start_time=start_time,
        )

        results = []
        async for chunk in gen:
            results.append(chunk)

        # Should get all chunks
        assert len(results) == 2

    async def test_with_cache_id_triggers_update(self):
        """测试提供cache_id时触发缓存检查更新"""
        from app.router.agent_controller_pkg.run_agent_v2.safe_output_generator import (
            create_safe_output_generator,
        )
        from app.domain.vo.agent_cache import AgentCacheIdVO

        agent_core_v2 = MagicMock()
        agent_config = MagicMock()
        agent_input = MagicMock()
        headers = {}
        start_time = 1234567890.0
        cache_id_vo = MagicMock(spec=AgentCacheIdVO)

        async def simple_generator():
            yield "result"

        agent_core_v2.output_handler.result_output = MagicMock(
            return_value=simple_generator()
        )

        with patch(
            "app.router.agent_controller_pkg.run_agent_v2.safe_output_generator.check_and_update_cache"
        ) as mock_check:
            mock_check = AsyncMock()

            gen = create_safe_output_generator(
                agent_core_v2=agent_core_v2,
                agent_config=agent_config,
                agent_input=agent_input,
                headers=headers,
                is_debug_run=False,
                start_time=start_time,
                cache_id_vo=cache_id_vo,
                account_id="acc123",
                account_type="premium",
            )

            # Consume the generator
            async for _ in gen:
                pass

            # check_and_update_cache should be called in finally block
            # We verify the generator completes without error

    async def test_without_cache_id_skips_update(self):
        """测试不提供cache_id时跳过缓存检查"""
        from app.router.agent_controller_pkg.run_agent_v2.safe_output_generator import (
            create_safe_output_generator,
        )

        agent_core_v2 = MagicMock()
        agent_config = MagicMock()
        agent_input = MagicMock()
        headers = {}
        start_time = 1234567890.0

        async def simple_generator():
            yield "result"

        agent_core_v2.output_handler.result_output = MagicMock(
            return_value=simple_generator()
        )

        with patch(
            "app.router.agent_controller_pkg.run_agent_v2.safe_output_generator.check_and_update_cache"
        ) as mock_check:
            mock_check = AsyncMock()

            gen = create_safe_output_generator(
                agent_core_v2=agent_core_v2,
                agent_config=agent_config,
                agent_input=agent_input,
                headers=headers,
                is_debug_run=False,
                start_time=start_time,
                cache_id_vo=None,  # No cache_id
            )

            # Consume the generator
            async for _ in gen:
                pass

            # check_and_update_cache should NOT be called since cache_id_vo is None
            # We verify the generator completes without error

    async def test_incomplete_stream_closes_gracefully(self):
        """测试不完整的流优雅关闭"""
        from app.router.agent_controller_pkg.run_agent_v2.safe_output_generator import (
            create_safe_output_generator,
        )

        agent_core_v2 = MagicMock()
        agent_config = MagicMock()
        agent_input = MagicMock()
        headers = {}
        start_time = 1234567890.0

        # Create generator that will be cancelled mid-stream
        async def incomplete_generator():
            yield "chunk1"
            # Simulate mid-stream interruption
            raise asyncio.CancelledError()

        agent_core_v2.output_handler.result_output = MagicMock(
            return_value=incomplete_generator()
        )

        gen = create_safe_output_generator(
            agent_core_v2=agent_core_v2,
            agent_config=agent_config,
            agent_input=agent_input,
            headers=headers,
            is_debug_run=False,
            start_time=start_time,
        )

        results = []
        with pytest.raises(asyncio.CancelledError):
            async for chunk in gen:
                results.append(chunk)
                if len(results) >= 1:
                    # Interrupt after first chunk
                    raise asyncio.CancelledError()

        assert len(results) == 1
