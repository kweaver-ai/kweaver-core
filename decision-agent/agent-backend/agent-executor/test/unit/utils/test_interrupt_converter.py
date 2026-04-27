"""单元测试 - utils/interrupt_converter 模块"""


class TestInterruptHandleToResumeHandle:
    """测试 interrupt_handle_to_resume_handle 函数"""

    def test_convert_interrupt_handle_to_resume_handle_success(self):
        """测试成功转换 InterruptHandle 到 ResumeHandle"""
        from app.domain.vo.interrupt.interrupt_handle import InterruptHandle
        from app.utils.interrupt_converter import interrupt_handle_to_resume_handle

        # Create InterruptHandle with all fields
        interrupt_handle = InterruptHandle(
            frame_id="frame123",
            snapshot_id="snapshot456",
            resume_token="token789",
            interrupt_type="user_confirmation",
            current_block=1,
            restart_block=False,
        )

        # Execute conversion
        result = interrupt_handle_to_resume_handle(interrupt_handle)

        # Verify conversion
        assert result.frame_id == "frame123"
        assert result.snapshot_id == "snapshot456"
        assert result.resume_token == "token789"
        assert result.interrupt_type == "user_confirmation"
        assert result.current_block == 1
        assert result.restart_block is False

    def test_convert_with_restart_block_true(self):
        """测试转换 restart_block 为 True 的情况"""
        from app.domain.vo.interrupt.interrupt_handle import InterruptHandle
        from app.utils.interrupt_converter import interrupt_handle_to_resume_handle

        interrupt_handle = InterruptHandle(
            frame_id="frame123",
            snapshot_id="snapshot456",
            resume_token="token789",
            interrupt_type="tool_interrupt",
            current_block=2,
            restart_block=True,
        )

        result = interrupt_handle_to_resume_handle(interrupt_handle)

        assert result.restart_block is True
        assert result.current_block == 2
        assert result.interrupt_type == "tool_interrupt"

    def test_convert_with_large_block_index(self):
        """测试转换大索引值"""
        from app.domain.vo.interrupt.interrupt_handle import InterruptHandle
        from app.utils.interrupt_converter import interrupt_handle_to_resume_handle

        interrupt_handle = InterruptHandle(
            frame_id="frame_999",
            snapshot_id="snapshot_888",
            resume_token="token_777",
            interrupt_type="async_interrupt",
            current_block=9999,
            restart_block=False,
        )

        result = interrupt_handle_to_resume_handle(interrupt_handle)

        assert result.current_block == 9999

    def test_convert_with_various_interrupt_types(self):
        """测试不同的中断类型"""
        from app.domain.vo.interrupt.interrupt_handle import InterruptHandle
        from app.utils.interrupt_converter import interrupt_handle_to_resume_handle

        interrupt_types = [
            "tool_interrupt",
            "user_confirmation",
            "async_interrupt",
            "custom_interrupt",
            "timeout_interrupt",
        ]

        for interrupt_type in interrupt_types:
            interrupt_handle = InterruptHandle(
                frame_id="frame123",
                snapshot_id="snapshot456",
                resume_token="token789",
                interrupt_type=interrupt_type,
                current_block=0,
                restart_block=False,
            )

            result = interrupt_handle_to_resume_handle(interrupt_handle)

            assert result.interrupt_type == interrupt_type

    def test_convert_preserves_all_fields(self):
        """测试所有字段都被正确保留"""
        from app.domain.vo.interrupt.interrupt_handle import InterruptHandle
        from app.utils.interrupt_converter import interrupt_handle_to_resume_handle

        interrupt_handle = InterruptHandle(
            frame_id="test_frame",
            snapshot_id="test_snapshot",
            resume_token="test_token",
            interrupt_type="test_type",
            current_block=42,
            restart_block=True,
        )

        result = interrupt_handle_to_resume_handle(interrupt_handle)

        # Verify all fields match
        assert result.frame_id == interrupt_handle.frame_id
        assert result.snapshot_id == interrupt_handle.snapshot_id
        assert result.resume_token == interrupt_handle.resume_token
        assert result.interrupt_type == interrupt_handle.interrupt_type
        assert result.current_block == interrupt_handle.current_block
        assert result.restart_block == interrupt_handle.restart_block

    def test_convert_with_special_characters_in_ids(self):
        """测试包含特殊字符的 ID"""
        from app.domain.vo.interrupt.interrupt_handle import InterruptHandle
        from app.utils.interrupt_converter import interrupt_handle_to_resume_handle

        interrupt_handle = InterruptHandle(
            frame_id="frame-with-dashes_and_underscores",
            snapshot_id="snapshot.with.dots",
            resume_token="token/with/slashes",
            interrupt_type="special_interrupt",
            current_block=1,
            restart_block=False,
        )

        result = interrupt_handle_to_resume_handle(interrupt_handle)

        assert result.frame_id == "frame-with-dashes_and_underscores"
        assert result.snapshot_id == "snapshot.with.dots"
        assert result.resume_token == "token/with/slashes"

    def test_convert_with_unicode_values(self):
        """测试包含 Unicode 字符的值"""
        from app.domain.vo.interrupt.interrupt_handle import InterruptHandle
        from app.utils.interrupt_converter import interrupt_handle_to_resume_handle

        interrupt_handle = InterruptHandle(
            frame_id="帧123",
            snapshot_id="快照456",
            resume_token="令牌789",
            interrupt_type="中断类型",
            current_block=1,
            restart_block=False,
        )

        result = interrupt_handle_to_resume_handle(interrupt_handle)

        assert result.frame_id == "帧123"
        assert result.snapshot_id == "快照456"
        assert result.resume_token == "令牌789"

    def test_convert_returns_new_instance(self):
        """测试每次转换返回新实例"""
        from app.domain.vo.interrupt.interrupt_handle import InterruptHandle
        from app.utils.interrupt_converter import interrupt_handle_to_resume_handle

        interrupt_handle = InterruptHandle(
            frame_id="frame123",
            snapshot_id="snapshot456",
            resume_token="token789",
            interrupt_type="test",
            current_block=1,
            restart_block=False,
        )

        result1 = interrupt_handle_to_resume_handle(interrupt_handle)
        result2 = interrupt_handle_to_resume_handle(interrupt_handle)

        # Should be different instances
        assert result1 is not result2
        # But with same values
        assert result1.frame_id == result2.frame_id
