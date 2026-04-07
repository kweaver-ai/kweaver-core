import sys
import os

sys.path.insert(0, os.path.join(os.path.dirname(__file__), "../../.."))

import pytest
from unittest.mock import patch, MagicMock
from src.utils.logger import Logger


@pytest.fixture(autouse=True)
def reset_singleton():
    """Reset logger singleton before each test"""
    Logger._instance = None
    Logger._initialized = False
    yield


class TestLogger:
    @pytest.fixture
    def logger(self):
        """Create a logger instance for testing"""
        return Logger()

    def test_singleton_pattern(self):
        """Test that Logger implements singleton pattern correctly"""
        logger1 = Logger()
        logger2 = Logger()
        assert logger1 is logger2

    def test_logger_initialization(self, logger):
        """Test logger initialization"""
        assert logger.logger is not None
        assert logger.logger.name == "agent_memory"
        assert logger.logger.level <= 30  # Should be DEBUG (10) or INFO (20)

    @patch("src.utils.logger.Path")
    def test_log_directory_creation(self, mock_path):
        """Test that log directory is created"""
        mock_log_dir = MagicMock()
        mock_path.return_value = mock_log_dir
        mock_path.__truediv__ = MagicMock(return_value=mock_log_dir)

        mock_logger = MagicMock()
        with patch("src.utils.logger.logging.getLogger", return_value=mock_logger):
            with patch.object(mock_log_dir, "mkdir"):
                with patch("src.utils.logger.RotatingFileHandler"):
                    logger = Logger()

                    mock_log_dir.mkdir.assert_called_once_with(exist_ok=True)

    @patch("src.utils.logger.logging.getLogger")
    def test_info_method(self, mock_get_logger):
        """Test info logging method"""
        mock_logger = MagicMock()
        mock_get_logger.return_value = mock_logger

        logger = Logger()
        logger.info("Test message %s", "value", extra={"key": "value"})

        mock_logger.info.assert_called_once()

    @patch("src.utils.logger.logging.getLogger")
    def test_error_method(self, mock_get_logger):
        """Test error logging method"""
        mock_logger = MagicMock()
        mock_get_logger.return_value = mock_logger

        logger = Logger()
        logger.error("Error message %s", "value")

        mock_logger.error.assert_called_once()

    @patch("src.utils.logger.logging.getLogger")
    def test_debug_method(self, mock_get_logger):
        """Test debug logging method"""
        mock_logger = MagicMock()
        mock_get_logger.return_value = mock_logger

        logger = Logger()
        logger.debug("Debug message %s", "value")

        mock_logger.debug.assert_called_once()

    @patch("src.utils.logger.logging.getLogger")
    def test_infof_method(self, mock_get_logger):
        """Test formatted info logging method"""
        mock_logger = MagicMock()
        mock_get_logger.return_value = mock_logger

        logger = Logger()
        logger.infof("Test %s %d", "message", 42, extra={"key": "value"})

        mock_logger.info.assert_called_once_with(
            "Test %s %d" % ("message", 42), extra={"key": "value"}
        )

    @patch("src.utils.logger.logging.getLogger")
    def test_errorf_method(self, mock_get_logger):
        """Test formatted error logging method"""
        mock_logger = MagicMock()
        mock_get_logger.return_value = mock_logger

        logger = Logger()
        logger.errorf("Error %s", "test", exc_info=True)

        mock_logger.error.assert_called_once()
        call_args = mock_logger.error.call_args
        assert call_args[0][0] == "Error test"
        assert call_args[1]["exc_info"] is True

    @patch("src.utils.logger.logging.getLogger")
    def test_debugf_method(self, mock_get_logger):
        """Test formatted debug logging method"""
        mock_logger = MagicMock()
        mock_get_logger.return_value = mock_logger

        logger = Logger()
        logger.debugf("Debug %s", "info")

        mock_logger.debug.assert_called_once()

    @patch("src.utils.logger.logging.getLogger")
    def test_warn_method(self, mock_get_logger):
        """Test warning logging method"""
        mock_logger = MagicMock()
        mock_get_logger.return_value = mock_logger

        logger = Logger()
        logger.warning("Warning message")

        mock_logger.warning.assert_called_once()

    @patch("src.utils.logger.logging.getLogger")
    def test_warnf_method(self, mock_get_logger):
        """Test formatted warn logging method"""
        mock_logger = MagicMock()
        mock_get_logger.return_value = mock_logger

        logger = Logger()
        logger.warningf("Warning %s", "test")

        mock_logger.warning.assert_called_once()

    @patch("src.utils.logger.logging.getLogger")
    def test_initialized_flag(self, mock_get_logger):
        """Test that _initialized flag prevents reinitialization"""
        mock_logger = MagicMock()
        mock_get_logger.return_value = mock_logger

        logger = Logger()
        logger._initialized = False

        initial_setup_calls = mock_get_logger.call_count

        logger._setup_logger()

        # Should be called once more when _setup_logger is called after resetting flag
        assert mock_get_logger.call_count == initial_setup_calls + 1
