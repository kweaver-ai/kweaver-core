"""单元测试 - domain/vo/interrupt 模块"""

import pytest
from pydantic import ValidationError

from app.domain.vo.interrupt.interrupt_handle import InterruptHandle
from app.domain.vo.interrupt.interrupt_data import (
    ToolArg,
    InterruptConfig,
    InterruptData,
)
from app.domain.vo.interrupt.tool_interrupt_info import ToolInterruptInfo


class TestInterruptHandle:
    """测试 InterruptHandle VO"""

    def test_interrupt_handle_creation(self):
        """测试创建 InterruptHandle"""
        handle = InterruptHandle(
            frame_id="frame_123",
            snapshot_id="snapshot_456",
            resume_token="token_789",
            interrupt_type="tool_interrupt",
            current_block=5,
            restart_block=True,
        )

        assert handle.frame_id == "frame_123"
        assert handle.snapshot_id == "snapshot_456"
        assert handle.resume_token == "token_789"
        assert handle.interrupt_type == "tool_interrupt"
        assert handle.current_block == 5
        assert handle.restart_block is True

    def test_interrupt_handle_required_fields(self):
        """测试必填字段"""
        with pytest.raises(ValidationError):
            InterruptHandle(
                frame_id="frame_123",
                # Missing required fields
            )

    def test_interrupt_handle_model_dump(self):
        """测试 model_dump 方法"""
        handle = InterruptHandle(
            frame_id="frame_123",
            snapshot_id="snapshot_456",
            resume_token="token_789",
            interrupt_type="tool_interrupt",
            current_block=5,
            restart_block=False,
        )

        data = handle.model_dump()

        assert data["frame_id"] == "frame_123"
        assert data["restart_block"] is False

    def test_interrupt_handle_json_schema(self):
        """测试 JSON schema 生成"""
        schema = InterruptHandle.model_json_schema()
        assert "properties" in schema
        assert "frame_id" in schema["properties"]
        assert "snapshot_id" in schema["properties"]


class TestToolArg:
    """测试 ToolArg VO"""

    def test_tool_arg_creation(self):
        """测试创建 ToolArg"""
        arg = ToolArg(key="param1", value="value1", type="string")

        assert arg.key == "param1"
        assert arg.value == "value1"
        assert arg.type == "string"

    def test_tool_arg_with_different_types(self):
        """测试不同类型的参数"""
        arg_int = ToolArg(key="count", value=10, type="integer")
        arg_bool = ToolArg(key="enabled", value=True, type="boolean")
        arg_dict = ToolArg(key="config", value={"key": "value"}, type="object")

        assert arg_int.value == 10
        assert arg_bool.value is True
        assert arg_dict.value == {"key": "value"}


class TestInterruptConfig:
    """测试 InterruptConfig VO"""

    def test_interrupt_config_creation(self):
        """测试创建 InterruptConfig"""
        config = InterruptConfig(
            requires_confirmation=True,
            confirmation_message="Please confirm this action",
        )

        assert config.requires_confirmation is True
        assert config.confirmation_message == "Please confirm this action"

    def test_interrupt_config_false_confirmation(self):
        """测试不需要确认的情况"""
        config = InterruptConfig(
            requires_confirmation=False, confirmation_message="No confirmation needed"
        )

        assert config.requires_confirmation is False


class TestInterruptData:
    """测试 InterruptData VO"""

    def test_interrupt_data_minimal(self):
        """测试最小 InterruptData"""
        data = InterruptData(tool_name="search_tool")

        assert data.tool_name == "search_tool"
        assert data.tool_description is None
        assert data.tool_args == []
        assert data.interrupt_config is None

    def test_interrupt_data_full(self):
        """测试完整 InterruptData"""
        args = [
            ToolArg(key="query", value="test", type="string"),
            ToolArg(key="limit", value=10, type="integer"),
        ]
        config = InterruptConfig(
            requires_confirmation=True, confirmation_message="Run search?"
        )

        data = InterruptData(
            tool_name="search_tool",
            tool_description="Search the web",
            tool_args=args,
            interrupt_config=config,
        )

        assert data.tool_name == "search_tool"
        assert data.tool_description == "Search the web"
        assert len(data.tool_args) == 2
        assert data.interrupt_config.requires_confirmation is True

    def test_interrupt_data_with_args(self):
        """测试带参数的 InterruptData"""
        args = [
            ToolArg(key="param1", value="value1", type="string"),
            ToolArg(key="param2", value=123, type="integer"),
            ToolArg(key="param3", value=True, type="boolean"),
        ]

        data = InterruptData(tool_name="test_tool", tool_args=args)

        assert len(data.tool_args) == 3
        assert data.tool_args[0].key == "param1"
        assert data.tool_args[1].value == 123


class TestToolInterruptInfo:
    """测试 ToolInterruptInfo VO"""

    def test_tool_interrupt_info_empty(self):
        """测试空的 ToolInterruptInfo"""
        info = ToolInterruptInfo()

        assert info.handle is None
        assert info.data is None

    def test_tool_interrupt_info_with_handle(self):
        """测试带 handle 的 ToolInterruptInfo"""
        handle = InterruptHandle(
            frame_id="frame_123",
            snapshot_id="snapshot_456",
            resume_token="token_789",
            interrupt_type="tool_interrupt",
            current_block=5,
            restart_block=True,
        )

        info = ToolInterruptInfo(handle=handle)

        assert info.handle is not None
        assert info.handle.frame_id == "frame_123"
        assert info.data is None

    def test_tool_interrupt_info_with_data(self):
        """测试带 data 的 ToolInterruptInfo"""
        data = InterruptData(tool_name="test_tool", tool_description="Test tool")

        info = ToolInterruptInfo(data=data)

        assert info.data is not None
        assert info.data.tool_name == "test_tool"
        assert info.handle is None

    def test_tool_interrupt_info_full(self):
        """测试完整的 ToolInterruptInfo"""
        handle = InterruptHandle(
            frame_id="frame_123",
            snapshot_id="snapshot_456",
            resume_token="token_789",
            interrupt_type="tool_interrupt",
            current_block=5,
            restart_block=True,
        )

        data = InterruptData(
            tool_name="test_tool",
            tool_args=[ToolArg(key="key1", value="value1", type="string")],
        )

        info = ToolInterruptInfo(handle=handle, data=data)

        assert info.handle is not None
        assert info.data is not None
        assert info.handle.frame_id == "frame_123"
        assert info.data.tool_name == "test_tool"

    def test_tool_interrupt_info_model_dump(self):
        """测试 model_dump"""
        handle = InterruptHandle(
            frame_id="frame_123",
            snapshot_id="snapshot_456",
            resume_token="token_789",
            interrupt_type="tool",
            current_block=1,
            restart_block=False,
        )

        info = ToolInterruptInfo(handle=handle)
        dumped = info.model_dump()

        assert dumped["handle"]["frame_id"] == "frame_123"
        assert dumped["data"] is None
