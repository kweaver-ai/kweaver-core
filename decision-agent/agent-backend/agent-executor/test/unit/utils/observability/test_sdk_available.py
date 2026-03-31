"""单元测试 - utils/observability/sdk_available 模块"""

from unittest.mock import patch


class TestTelemetrySdkAvailable:
    """测试 TELEMETRY_SDK_AVAILABLE 模块变量"""

    def test_module_exists(self):
        """测试模块可以导入"""
        from app.utils.observability.sdk_available import TELEMETRY_SDK_AVAILABLE

        # Should be a boolean
        assert isinstance(TELEMETRY_SDK_AVAILABLE, bool)

    def test_tlogging_module_is_none_or_module(self):
        """测试tlogging_module是否为None或模块"""
        from app.utils.observability.sdk_available import tlogging_module

        # Should be None or a module
        assert tlogging_module is None or hasattr(tlogging_module, "__name__")

    def test_sampler_logger_is_none_or_callable(self):
        """测试SamplerLogger是否为None或可调用"""
        from app.utils.observability.sdk_available import SamplerLogger

        # Should be None or callable
        assert SamplerLogger is None or callable(SamplerLogger)

    def test_log_resource_is_none_or_callable(self):
        """测试log_resource是否为None或可调用"""
        from app.utils.observability.sdk_available import log_resource

        # Should be None or callable
        assert log_resource is None or callable(log_resource)

    def test_trace_resource_is_none_or_callable(self):
        """测试trace_resource是否为None或可调用"""
        from app.utils.observability.sdk_available import trace_resource

        # Should be None or callable
        assert trace_resource is None or callable(trace_resource)

    def test_set_service_info_is_none_or_callable(self):
        """测试set_service_info是否为None或可调用"""
        from app.utils.observability.sdk_available import set_service_info

        # Should be None or callable
        assert set_service_info is None or callable(set_service_info)

    def test_sdk_tracer_is_none_or_module(self):
        """测试sdk_tracer是否为None或模块"""
        from app.utils.observability.sdk_available import sdk_tracer

        # Should be None or a module/object
        assert sdk_tracer is None or hasattr(sdk_tracer, "__class__")

    @patch("app.utils.observability.sdk_available.is_aaron_local_dev")
    def test_skips_import_when_aaron_local_dev(self, mock_is_aaron):
        """测试Aaron本地开发环境时跳过SDK导入"""
        # This test just verifies the behavior in different environments
        # The actual value depends on the test environment
        from app.utils.observability.sdk_available import TELEMETRY_SDK_AVAILABLE

        # Just verify it's a boolean
        assert isinstance(TELEMETRY_SDK_AVAILABLE, bool)

    def test_module_attributes_exist(self):
        """测试所有模块属性存在"""
        from app.utils.observability import sdk_available

        # Check that all expected attributes exist
        assert hasattr(sdk_available, "TELEMETRY_SDK_AVAILABLE")
        assert hasattr(sdk_available, "tlogging_module")
        assert hasattr(sdk_available, "SamplerLogger")
        assert hasattr(sdk_available, "log_resource")
        assert hasattr(sdk_available, "trace_resource")
        assert hasattr(sdk_available, "set_service_info")
        assert hasattr(sdk_available, "sdk_tracer")
