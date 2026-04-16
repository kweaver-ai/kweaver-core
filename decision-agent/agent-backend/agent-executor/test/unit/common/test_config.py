# -*- coding: utf-8 -*-
"""Unit tests for app/common/config.py module."""

import sys

import pytest


class TestServerInfo:
    @pytest.mark.asyncio
    async def test_server_info_initialization(self):
        from app.common.config import Config, server_info

        assert server_info is not None
        assert server_info.server_name == Config.o11y.service_name
        assert server_info.server_version == Config.o11y.service_version
        assert server_info.language == "python"
        assert server_info.python_version == sys.version


class TestObservabilityConfig:
    @pytest.mark.asyncio
    async def test_observability_config_initialization(self):
        from app.common.config import observability_config

        assert observability_config is not None
        assert hasattr(observability_config, "log")
        assert hasattr(observability_config, "trace")
        assert not hasattr(observability_config, "metric")

    @pytest.mark.asyncio
    async def test_log_config_only_keeps_latest_fields(self):
        from app.common.config import Config, observability_config

        log_config = observability_config.log
        assert log_config.log_enabled == Config.o11y.log_enabled
        assert log_config.log_level == Config.o11y.log_level
        assert not hasattr(log_config, "log_exporter")
        assert not hasattr(log_config, "log_load_interval")
        assert not hasattr(log_config, "log_load_max_log")

    @pytest.mark.asyncio
    async def test_trace_config_only_keeps_latest_fields(self):
        from app.common.config import Config, observability_config

        trace_config = observability_config.trace
        assert trace_config.trace_enabled == Config.o11y.trace_enabled
        assert trace_config.otlp_endpoint == Config.o11y.trace_endpoint
        assert trace_config.environment == Config.o11y.environment
        assert trace_config.sampling_rate == Config.o11y.trace_sampling_rate
        assert hasattr(trace_config, "trace_max_queue_size")
        assert not hasattr(trace_config, "trace_provider")
        assert not hasattr(trace_config, "max_export_batch_size")


class TestConfigInstance:
    @pytest.mark.asyncio
    async def test_config_instance_exists(self):
        from app.common.config import Config

        assert Config is not None

    @pytest.mark.asyncio
    async def test_config_has_app_attribute(self):
        from app.common.config import Config

        assert hasattr(Config, "app")

    @pytest.mark.asyncio
    async def test_config_has_local_dev_attribute(self):
        from app.common.config import Config

        assert hasattr(Config, "local_dev")


class TestBuiltinIdsInstance:
    @pytest.mark.asyncio
    async def test_builtin_ids_instance_exists(self):
        from app.common.config import BuiltinIds

        assert BuiltinIds is not None

    @pytest.mark.asyncio
    async def test_builtin_ids_has_agent_ids(self):
        from app.common.config import BuiltinIds

        assert hasattr(BuiltinIds, "agent_ids")
        assert isinstance(BuiltinIds.agent_ids, dict)

    @pytest.mark.asyncio
    async def test_builtin_ids_has_tool_ids(self):
        from app.common.config import BuiltinIds

        assert hasattr(BuiltinIds, "tool_ids")
        assert isinstance(BuiltinIds.tool_ids, dict)
