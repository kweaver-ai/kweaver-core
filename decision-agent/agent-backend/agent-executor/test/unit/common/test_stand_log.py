"""单元测试 - common/stand_log 模块"""

import logging
import stat
from unittest.mock import MagicMock, patch
import sys


class TestModuleConstants:
    """测试模块常量"""

    def test_module_constants_exist(self):
        """测试模块常量存在"""
        from app.common.stand_log import (
            SYSTEM_LOG,
            BUSINESS_LOG,
            CREATE,
            DELETE,
            DOWNLOAD,
            UPDATE,
            UPLOAD,
            LOGIN,
            LOG_FILE,
            LOG_DIR,
            LOG_FORMATTER,
        )

        assert SYSTEM_LOG == "SystemLog"
        assert BUSINESS_LOG == "BusinessLog"
        assert CREATE == "create"
        assert DELETE == "delete"
        assert DOWNLOAD == "download"
        assert UPDATE == "update"
        assert UPLOAD == "upload"
        assert LOGIN == "login"
        assert LOG_FILE == "/dev/fd/1"
        assert LOG_DIR == "log"
        assert isinstance(LOG_FORMATTER, logging.Formatter)


class TestStandLogLogging:
    """测试 StandLog_logging 类"""

    @patch("app.common.stand_log.Config")
    @patch("app.common.stand_log.os.access")
    @patch("app.common.stand_log.os.stat")
    @patch("app.common.stand_log.os.chmod")
    @patch("app.common.stand_log.os.makedirs")
    def test_init_creates_log_directory_if_not_exists(
        self, m_makedirs, m_chmod, m_stat, m_access, m_config
    ):
        """测试日志目录不存在时创建"""
        m_stat.return_value.st_mode = stat.S_IRUSR | stat.S_IXUSR
        m_access.return_value = True
        m_config.app.get_stdlib_log_level.return_value = logging.INFO

        from app.common.stand_log import StandLog_logging

        logger = StandLog_logging()

        m_makedirs.assert_called_once_with("log", mode=0o755, exist_ok=True)
        m_chmod.assert_called_once_with(
            "log", stat.S_IRUSR | stat.S_IWUSR | stat.S_IXUSR
        )
        assert logger._info_logger is not None

    @patch("app.common.stand_log.Config")
    @patch("app.common.stand_log.os.access")
    @patch("app.common.stand_log.os.stat")
    @patch("app.common.stand_log.os.chmod")
    @patch("app.common.stand_log.os.makedirs")
    def test_init_with_existing_directory(
        self, m_makedirs, m_chmod, m_stat, m_access, m_config
    ):
        """测试目录已存在时正常初始化"""
        m_stat.return_value.st_mode = 0o755
        m_access.return_value = True
        m_config.app.get_stdlib_log_level.return_value = logging.INFO

        from app.common.stand_log import StandLog_logging

        logger = StandLog_logging()

        m_makedirs.assert_called_once_with("log", mode=0o755, exist_ok=True)
        m_chmod.assert_not_called()
        assert logger._info_logger is not None
        assert logger._fastapi_logger is not None

    @patch("app.common.stand_log.Config")
    @patch("app.common.stand_log.os.access")
    @patch("app.common.stand_log.os.stat")
    @patch("app.common.stand_log.os.chmod")
    @patch("app.common.stand_log.os.makedirs")
    def test_get_request_logger(self, m_makedirs, m_chmod, m_stat, m_access, m_config):
        """测试获取请求日志记录器"""
        m_stat.return_value.st_mode = 0o755
        m_access.return_value = True
        m_config.app.get_stdlib_log_level.return_value = logging.INFO

        from app.common.stand_log import StandLog_logging

        logger = StandLog_logging()
        request_logger = logger.get_request_logger()

        m_makedirs.assert_called_once_with("log", mode=0o755, exist_ok=True)
        m_chmod.assert_not_called()
        assert request_logger is not None
        assert request_logger == logger._fastapi_logger

    @patch("builtins.print")
    @patch("app.common.stand_log.Config")
    @patch("app.common.stand_log.os.access")
    @patch("app.common.stand_log.os.stat")
    @patch("app.common.stand_log.os.chmod")
    @patch("app.common.stand_log.os.makedirs")
    def test_init_prints_error_when_log_dir_is_not_writable(
        self, m_makedirs, m_chmod, m_stat, m_access, m_config, m_print
    ):
        """测试日志目录不可写时打印初始化错误"""
        m_stat.return_value.st_mode = 0o755
        m_access.return_value = False
        m_config.app.get_stdlib_log_level.return_value = logging.INFO

        from app.common.stand_log import StandLog_logging

        logger = StandLog_logging()

        m_makedirs.assert_called_once_with("log", mode=0o755, exist_ok=True)
        m_chmod.assert_not_called()
        m_print.assert_called_once()
        assert logger._info_logger is None

    @patch("app.common.stand_log.Config")
    @patch("app.common.stand_log.os.path.exists")
    @patch("app.common.stand_log.IsInPod")
    def test_need_print_true_when_not_in_pod(self, m_is_in_pod, m_exists, m_config):
        """测试不在Pod中时需要打印"""
        m_is_in_pod.return_value = False
        m_exists.return_value = True
        m_config.app.get_stdlib_log_level.return_value = logging.INFO

        from app.common.stand_log import StandLog_logging, SYSTEM_LOG

        logger = StandLog_logging()
        # 通过检查日志方法是否正常工作来验证
        logger._info_logger.info = MagicMock()

        logger.info("test message", SYSTEM_LOG)

        logger._info_logger.info.assert_called()

    @patch("app.common.stand_log.Config")
    @patch("app.common.stand_log.os.path.exists")
    @patch("app.common.stand_log.IsInPod")
    def test_debug_method(self, m_is_in_pod, m_exists, m_config):
        """测试debug方法"""
        m_is_in_pod.return_value = False
        m_exists.return_value = True
        m_config.app.get_stdlib_log_level.return_value = logging.INFO

        from app.common.stand_log import StandLog_logging, SYSTEM_LOG

        logger = StandLog_logging()
        logger._info_logger.debug = MagicMock()

        logger.debug("test debug", SYSTEM_LOG)

        logger._info_logger.debug.assert_called_once()

    @patch("app.common.stand_log.Config")
    @patch("app.common.stand_log.os.path.exists")
    @patch("app.common.stand_log.IsInPod")
    def test_info_method(self, m_is_in_pod, m_exists, m_config):
        """测试info方法"""
        m_is_in_pod.return_value = False
        m_exists.return_value = True
        m_config.app.get_stdlib_log_level.return_value = logging.INFO

        from app.common.stand_log import StandLog_logging, SYSTEM_LOG

        logger = StandLog_logging()
        logger._info_logger.info = MagicMock()

        logger.info("test info", SYSTEM_LOG)

        logger._info_logger.info.assert_called_once()

    @patch("app.common.stand_log.Config")
    @patch("app.common.stand_log.os.path.exists")
    @patch("app.common.stand_log.IsInPod")
    def test_warn_method(self, m_is_in_pod, m_exists, m_config):
        """测试warn方法"""
        m_is_in_pod.return_value = False
        m_exists.return_value = True
        m_config.app.get_stdlib_log_level.return_value = logging.INFO

        from app.common.stand_log import StandLog_logging, SYSTEM_LOG

        logger = StandLog_logging()
        logger._info_logger.warning = MagicMock()

        logger.warn("test warning", SYSTEM_LOG)

        logger._info_logger.warning.assert_called_once()

    @patch("app.common.stand_log.Config")
    @patch("app.common.stand_log.os.path.exists")
    @patch("app.common.stand_log.IsInPod")
    def test_error_method(self, m_is_in_pod, m_exists, m_config):
        """测试error方法"""
        m_is_in_pod.return_value = False
        m_exists.return_value = True
        m_config.app.get_stdlib_log_level.return_value = logging.INFO

        from app.common.stand_log import StandLog_logging, SYSTEM_LOG

        logger = StandLog_logging()
        logger._info_logger.error = MagicMock()

        logger.error("test error", SYSTEM_LOG)

        logger._info_logger.error.assert_called_once()

    @patch("app.common.stand_log.Config")
    @patch("app.common.stand_log.os.path.exists")
    @patch("app.common.stand_log.IsInPod")
    def test_info_log_method(self, m_is_in_pod, m_exists, m_config):
        """测试info_log方法"""
        m_is_in_pod.return_value = False
        m_exists.return_value = True
        m_config.app.get_stdlib_log_level.return_value = logging.INFO

        from app.common.stand_log import StandLog_logging

        logger = StandLog_logging()
        logger._info_logger.info = MagicMock()

        logger.info_log("test info log")

        logger._info_logger.info.assert_called_once()

    @patch("app.common.stand_log.Config")
    @patch("app.common.stand_log.os.path.exists")
    @patch("app.common.stand_log.IsInPod")
    def test_debug_log_method(self, m_is_in_pod, m_exists, m_config):
        """测试debug_log方法"""
        m_is_in_pod.return_value = False
        m_exists.return_value = True
        m_config.app.get_stdlib_log_level.return_value = logging.INFO

        from app.common.stand_log import StandLog_logging

        logger = StandLog_logging()
        logger._info_logger.debug = MagicMock()

        logger.debug_log("test debug log")

        logger._info_logger.debug.assert_called_once()

    @patch("app.common.stand_log.Config")
    @patch("app.common.stand_log.os.path.exists")
    @patch("app.common.stand_log.IsInPod")
    def test_console_request_log(self, m_is_in_pod, m_exists, m_config):
        """测试控制台请求日志"""
        m_is_in_pod.return_value = False
        m_exists.return_value = True
        m_config.app.get_stdlib_log_level.return_value = logging.INFO

        from app.common.stand_log import StandLog_logging

        logger = StandLog_logging()
        logger._fastapi_logger.info = MagicMock()

        req_info = {
            "request_id": "req123",
            "method": "POST",
            "path": "/api/test",
        }

        logger.console_request_log(req_info)

        logger._fastapi_logger.info.assert_called_once()

    @patch("app.common.stand_log.Config")
    @patch("app.common.stand_log.os.path.exists")
    def test_generate_curl_command_get_request(self, m_exists, m_config):
        """测试生成GET请求的curl命令"""
        m_exists.return_value = True
        m_config.app.get_stdlib_log_level.return_value = logging.INFO

        from app.common.stand_log import StandLog_logging

        logger = StandLog_logging()

        req_info = {
            "method": "GET",
            "path": "/api/test",
            "headers": {"host": "example.com"},
            "query_params": {"param1": "value1"},
        }

        curl = logger._generate_curl_command(req_info)

        assert "curl" in curl
        assert "example.com" in curl
        assert "/api/test" in curl
        assert "param1=value1" in curl

    @patch("app.common.stand_log.Config")
    @patch("app.common.stand_log.os.path.exists")
    def test_generate_curl_command_post_request(self, m_exists, m_config):
        """测试生成POST请求的curl命令"""
        m_exists.return_value = True
        m_config.app.get_stdlib_log_level.return_value = logging.INFO

        from app.common.stand_log import StandLog_logging

        logger = StandLog_logging()

        req_info = {
            "method": "POST",
            "path": "/api/test",
            "headers": {"host": "example.com", "Content-Type": "application/json"},
            "body": {"key": "value"},
        }

        curl = logger._generate_curl_command(req_info)

        assert "curl" in curl
        assert "-X POST" in curl
        assert "example.com" in curl
        assert "-d" in curl

    @patch("app.common.stand_log.Config")
    @patch("app.common.stand_log.os.path.exists")
    def test_generate_curl_command_with_json_body(self, m_exists, m_config):
        """测试生成带JSON体的curl命令"""
        m_exists.return_value = True
        m_config.app.get_stdlib_log_level.return_value = logging.INFO

        from app.common.stand_log import StandLog_logging

        logger = StandLog_logging()

        req_info = {
            "method": "POST",
            "path": "/api/test",
            "headers": {"host": "example.com"},
            "body": '{"key": "value"}',
        }

        curl = logger._generate_curl_command(req_info)

        assert "-d" in curl


class TestGetErrorLog:
    """测试 get_error_log 函数"""

    def test_get_error_log_basic(self):
        """测试基本错误日志获取"""
        from app.common.stand_log import get_error_log

        frame = sys._getframe()

        result = get_error_log("Test error", frame)

        assert isinstance(result, dict)
        assert "message" in result
        assert "caller" in result
        assert "stack" in result
        assert "time" in result
        assert result["message"] == "Test error"

    def test_get_error_log_with_traceback(self):
        """测试带追踪的错误日志获取"""
        from app.common.stand_log import get_error_log

        frame = sys._getframe()
        traceback_str = "Error traceback\nline 1\nline 2"

        result = get_error_log("Test error", frame, traceback_str)

        assert result["stack"] == traceback_str


class TestStandLogger:
    """测试 StandLogger 实例"""

    @patch("app.common.stand_log.Config")
    @patch("app.common.stand_log.os.path.exists")
    def test_stand_logger_instance(self, m_exists, m_config):
        """测试StandLogger是StandLog_logging实例"""
        m_exists.return_value = True
        m_config.app.get_stdlib_log_level.return_value = logging.INFO

        from app.common.stand_log import StandLogger, StandLog_logging

        assert isinstance(StandLogger, StandLog_logging)
