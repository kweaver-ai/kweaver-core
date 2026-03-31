# -*- coding: utf-8 -*-
"""
Unit tests for app/common/config.py module

NOTE: This test file does NOT mock app.config.config_v2 or app.config.builtin_ids_class
because mocking them would pollute the global Config object and affect other tests.
The tests here verify the structure of the config module without mocking dependencies.
"""

import sys
import pytest


class TestServerInfo:
    """Tests for server_info initialization"""

    @pytest.mark.asyncio
    async def test_server_info_initialization(self):
        """Test that server_info is properly initialized"""
        from app.common.config import server_info

        assert server_info is not None
        assert server_info.server_name == "agent-executor"
        assert server_info.server_version == "1.0.0"
        assert server_info.language == "python"
        assert server_info.python_version == sys.version


class TestObservabilityConfig:
    """Tests for observability_config initialization"""

    @pytest.mark.asyncio
    async def test_observability_config_initialization(self):
        """Test that observability_config is properly initialized"""
        from app.common.config import observability_config

        assert observability_config is not None
        assert hasattr(observability_config, "log")
        assert hasattr(observability_config, "trace")

    @pytest.mark.asyncio
    async def test_log_config_has_required_attributes(self):
        """Test log config has required attributes"""
        from app.common.config import observability_config

        log_config = observability_config.log
        assert hasattr(log_config, "log_enabled")
        assert hasattr(log_config, "log_exporter")
        assert hasattr(log_config, "log_load_interval")
        assert hasattr(log_config, "log_load_max_log")

    @pytest.mark.asyncio
    async def test_trace_config_has_required_attributes(self):
        """Test trace config has required attributes"""
        from app.common.config import observability_config

        trace_config = observability_config.trace
        assert hasattr(trace_config, "trace_enabled")
        assert hasattr(trace_config, "trace_provider")
        assert hasattr(trace_config, "trace_max_queue_size")
        assert hasattr(trace_config, "max_export_batch_size")


class TestConfigInstance:
    """Tests for Config instance"""

    @pytest.mark.asyncio
    async def test_config_instance_exists(self):
        """Test that Config instance exists"""
        from app.common.config import Config

        assert Config is not None

    @pytest.mark.asyncio
    async def test_config_has_app_attribute(self):
        """Test that Config has app attribute"""
        from app.common.config import Config

        assert hasattr(Config, "app")

    @pytest.mark.asyncio
    async def test_config_has_local_dev_attribute(self):
        """Test that Config has local_dev attribute"""
        from app.common.config import Config

        assert hasattr(Config, "local_dev")


class TestBuiltinIdsInstance:
    """Tests for BuiltinIds instance"""

    @pytest.mark.asyncio
    async def test_builtin_ids_instance_exists(self):
        """Test that BuiltinIds instance exists"""
        from app.common.config import BuiltinIds

        assert BuiltinIds is not None

    @pytest.mark.asyncio
    async def test_builtin_ids_has_agent_ids(self):
        """Test that BuiltinIds has agent_ids"""
        from app.common.config import BuiltinIds

        assert hasattr(BuiltinIds, "agent_ids")
        assert isinstance(BuiltinIds.agent_ids, dict)

    @pytest.mark.asyncio
    async def test_builtin_ids_has_tool_ids(self):
        """Test that BuiltinIds has tool_ids"""
        from app.common.config import BuiltinIds

        assert hasattr(BuiltinIds, "tool_ids")
        assert isinstance(BuiltinIds.tool_ids, dict)
