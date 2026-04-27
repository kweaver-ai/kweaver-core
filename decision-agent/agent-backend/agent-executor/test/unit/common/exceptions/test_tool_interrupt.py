"""Tests for app.common.exceptions.tool_interrupt module."""

import pytest
from unittest.mock import MagicMock
import sys
from pathlib import Path
from types import ModuleType

# Import the module directly by file path to avoid circular imports through __init__.py
# Get the project root (go up from test/unit/common/exceptions to project root)
project_root = Path(__file__).parent.parent.parent.parent.parent
tool_interrupt_file = (
    project_root / "app" / "common" / "exceptions" / "tool_interrupt.py"
)

# Use importlib to import the module directly
import importlib.util

spec = importlib.util.spec_from_file_location(
    "tool_interrupt_module", str(tool_interrupt_file)
)
tool_interrupt_module = importlib.util.module_from_spec(spec)


# Mock the dolphin dependency before loading
# Use proper module structure to avoid breaking other tests
class MockResumeHandle:
    pass


# Create proper mock modules that are packages (with __path__)
def create_mock_package(name):
    """Create a mock module that is a package"""
    module = ModuleType(name)
    module.__path__ = []  # Mark as package
    return module


# Only set up mocks if they don't already exist
if "dolphin" not in sys.modules:
    sys.modules["dolphin"] = create_mock_package("dolphin")
if "dolphin.core" not in sys.modules:
    sys.modules["dolphin.core"] = create_mock_package("dolphin.core")
if "dolphin.core.coroutine" not in sys.modules:
    sys.modules["dolphin.core.coroutine"] = create_mock_package(
        "dolphin.core.coroutine"
    )
if "dolphin.core.coroutine.resume_handle" not in sys.modules:
    resume_handle_module = create_mock_package("dolphin.core.coroutine.resume_handle")
    resume_handle_module.ResumeHandle = MockResumeHandle
    sys.modules["dolphin.core.coroutine.resume_handle"] = resume_handle_module

# Now load the module
spec.loader.exec_module(tool_interrupt_module)

ToolInterruptInfo = tool_interrupt_module.ToolInterruptInfo
ToolInterruptException = tool_interrupt_module.ToolInterruptException


class TestToolInterruptInfo:
    """Tests for ToolInterruptInfo dataclass."""

    def test_tool_interrupt_info_creation(self):
        """Test creating ToolInterruptInfo with handle and data."""
        mock_handle = MagicMock()
        data = {"tool_name": "test_tool", "tool_args": []}

        info = ToolInterruptInfo(handle=mock_handle, data=data)

        assert info.handle == mock_handle
        assert info.data == data
        assert info.data["tool_name"] == "test_tool"

    def test_tool_interrupt_info_with_empty_data(self):
        """Test ToolInterruptInfo with empty data dict."""
        mock_handle = MagicMock()
        info = ToolInterruptInfo(handle=mock_handle, data={})

        assert info.handle == mock_handle
        assert info.data == {}

    def test_tool_interrupt_info_with_complex_data(self):
        """Test ToolInterruptInfo with complex nested data."""
        mock_handle = MagicMock()
        data = {
            "tool_name": "complex_tool",
            "tool_description": "A complex tool",
            "tool_args": [
                {"key": "param1", "value": "value1", "type": "string"},
                {"key": "param2", "value": 42, "type": "integer"},
            ],
            "interrupt_config": {
                "requires_confirmation": True,
                "confirmation_message": "Please confirm",
            },
        }

        info = ToolInterruptInfo(handle=mock_handle, data=data)

        assert info.data["tool_name"] == "complex_tool"
        assert len(info.data["tool_args"]) == 2
        assert info.data["interrupt_config"]["requires_confirmation"] is True


class TestGetResumeHandleClass:
    """Tests for _get_resume_handle_class function."""

    def test_get_resume_handle_class_returns_class(self):
        """Test that _get_resume_handle_class returns a class."""
        resume_handle_class = tool_interrupt_module._get_resume_handle_class()

        assert resume_handle_class is not None
        assert callable(resume_handle_class)

    def test_get_resume_handle_class_mock_attributes(self):
        """Test that mock ResumeHandle has expected attributes."""
        resume_handle_class = tool_interrupt_module._get_resume_handle_class()

        # Create instance
        handle = resume_handle_class(
            frame_id="test_frame",
            snapshot_id="test_snapshot",
            resume_token="test_token",
            interrupt_type="test_type",
            current_block="test_current",
            restart_block="test_restart",
        )

        assert hasattr(handle, "frame_id")
        assert hasattr(handle, "snapshot_id")
        assert hasattr(handle, "resume_token")
        assert hasattr(handle, "interrupt_type")
        assert hasattr(handle, "current_block")
        assert hasattr(handle, "restart_block")

    def test_get_resume_handle_class_mock_values(self):
        """Test that mock ResumeHandle stores values correctly."""
        resume_handle_class = tool_interrupt_module._get_resume_handle_class()

        handle = resume_handle_class(
            frame_id="frame123",
            snapshot_id="snap456",
            resume_token="token789",
            interrupt_type="user_input",
            current_block="block1",
            restart_block="block2",
        )

        assert handle.frame_id == "frame123"
        assert handle.snapshot_id == "snap456"
        assert handle.resume_token == "token789"
        assert handle.interrupt_type == "user_input"
        assert handle.current_block == "block1"
        assert handle.restart_block == "block2"

    def test_get_resume_handle_class_default_values(self):
        """Test that mock ResumeHandle can be created with defaults."""
        resume_handle_class = tool_interrupt_module._get_resume_handle_class()

        handle = resume_handle_class()

        assert handle.frame_id == ""
        assert handle.snapshot_id == ""
        assert handle.resume_token == ""
        assert handle.interrupt_type == ""
        assert handle.current_block == ""
        assert handle.restart_block == ""

    def test_get_resume_handle_class_partial_values(self):
        """Test that mock ResumeHandle can be created with partial values."""
        resume_handle_class = tool_interrupt_module._get_resume_handle_class()

        handle = resume_handle_class(frame_id="test_frame", interrupt_type="test")

        assert handle.frame_id == "test_frame"
        assert handle.interrupt_type == "test"
        assert handle.snapshot_id == ""
        assert handle.current_block == ""


class TestResumeHandleClassModuleLevel:
    """Tests for module-level ResumeHandleClass variable."""

    def test_resume_handle_class_exists(self):
        """Test that ResumeHandleClass is defined at module level."""
        assert hasattr(tool_interrupt_module, "ResumeHandleClass")
        assert tool_interrupt_module.ResumeHandleClass is not None

    def test_resume_handle_class_is_callable(self):
        """Test that ResumeHandleClass is callable (a class)."""
        assert callable(tool_interrupt_module.ResumeHandleClass)

    def test_resume_handle_class_creates_instance(self):
        """Test that ResumeHandleClass can create instances."""
        handle = tool_interrupt_module.ResumeHandleClass(
            frame_id="test", snapshot_id="test"
        )

        assert handle is not None
        assert hasattr(handle, "frame_id")


class TestToolInterruptException:
    """Tests for ToolInterruptException class."""

    def test_tool_interrupt_exception_creation(self):
        """Test creating ToolInterruptException."""
        mock_handle = MagicMock()
        data = {"tool_name": "test_tool"}
        info = ToolInterruptInfo(handle=mock_handle, data=data)

        exception = ToolInterruptException(info)

        assert exception.interrupt_info == info
        assert "test_tool" in str(exception)

    def test_tool_interrupt_exception_message(self):
        """Test ToolInterruptException message format."""
        mock_handle = MagicMock()
        data = {"tool_name": "search_tool"}
        info = ToolInterruptInfo(handle=mock_handle, data=data)

        exception = ToolInterruptException(info)

        assert str(exception) == "Tool interrupt: search_tool"

    def test_tool_interrupt_exception_with_unknown_tool(self):
        """Test ToolInterruptException when tool_name is missing."""
        mock_handle = MagicMock()
        data = {}
        info = ToolInterruptInfo(handle=mock_handle, data=data)

        exception = ToolInterruptException(info)

        assert "unknown" in str(exception)

    def test_tool_interrupt_exception_with_none_data(self):
        """Test ToolInterruptException when data is None."""
        mock_handle = MagicMock()
        info = ToolInterruptInfo(handle=mock_handle, data=None)

        exception = ToolInterruptException(info)

        assert "unknown" in str(exception)

    def test_tool_interrupt_exception_is_exception(self):
        """Test that ToolInterruptException is an Exception subclass."""
        assert issubclass(ToolInterruptException, Exception)

        mock_handle = MagicMock()
        info = MagicMock()
        info.data = {"tool_name": "test"}

        exception = ToolInterruptException(info)
        assert isinstance(exception, Exception)

    def test_tool_interrupt_exception_can_be_raised(self):
        """Test that ToolInterruptException can be raised and caught."""
        mock_handle = MagicMock()
        data = {"tool_name": "interrupting_tool"}
        info = ToolInterruptInfo(handle=mock_handle, data=data)

        with pytest.raises(ToolInterruptException) as exc_info:
            raise ToolInterruptException(info)

        assert exc_info.value.interrupt_info == info
        assert "interrupting_tool" in str(exc_info.value)

    def test_tool_interrupt_exception_attributes(self):
        """Test that ToolInterruptException has expected attributes."""
        mock_handle = MagicMock()
        data = {
            "tool_name": "attribute_test",
            "tool_args": [{"key": "test", "value": "value"}],
        }
        info = ToolInterruptInfo(handle=mock_handle, data=data)

        exception = ToolInterruptException(info)

        assert hasattr(exception, "interrupt_info")
        assert exception.interrupt_info.data == data
