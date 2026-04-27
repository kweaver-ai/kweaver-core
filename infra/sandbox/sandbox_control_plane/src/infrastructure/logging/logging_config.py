"""
Logging configuration for Sandbox Control Plane.

Configures structlog for human-readable text logging (default) with optional JSON format.
Supports request context tracing and colored console output.
"""

import logging
import sys
from typing import Any, Optional

import structlog
from structlog.types import EventDict, Processor

# Color codes for terminal output
COLORS = {
    "debug": "\033[36m",     # Cyan
    "info": "\033[32m",      # Green
    "warning": "\033[33m",   # Yellow
    "error": "\033[31m",     # Red
    "critical": "\033[35m",  # Magenta
    "reset": "\033[0m",      # Reset
}


def add_color(logger: Any, method_name: str, event_dict: EventDict) -> EventDict:
    """
    Add ANSI color codes to log level for better console readability.
    """
    if method_name == "info":
        level_color = COLORS["info"]
    elif method_name == "debug":
        level_color = COLORS["debug"]
    elif method_name == "warning":
        level_color = COLORS["warning"]
    elif method_name == "error":
        level_color = COLORS["error"]
    elif method_name == "critical":
        level_color = COLORS["critical"]
    else:
        level_color = COLORS["reset"]

    # Add color to the level name
    if "level" in event_dict:
        event_dict["level"] = f"{level_color}{event_dict['level'].upper()}{COLORS['reset']}"

    return event_dict


def human_readable_renderer(
    logger: Any,
    method_name: str,
    event_dict: EventDict
) -> str:
    """
    Human-readable log format renderer.

    Format: [timestamp] [level] [logger] message key=value key2=value2
    Example: [2025-01-14 10:30:45] [INFO] [session_service] Session created session_id=abc123
    """
    # Extract basic fields
    timestamp = event_dict.pop("timestamp", "")
    level = event_dict.pop("level", "INFO").upper()
    logger_name = event_dict.pop("logger_name", event_dict.pop("logger", "unknown"))
    message = event_dict.pop("event", "")

    # Build the log line
    parts = []

    # Timestamp
    if timestamp:
        parts.append(f"[{timestamp}]")

    # Level with color (will be added by add_color processor)
    parts.append(f"[{level}]")

    # Logger name
    if logger_name != "root":
        parts.append(f"[{logger_name}]")

    # Message
    parts.append(message)

    # Add key-value pairs (sorted for consistency)
    for key, value in sorted(event_dict.items()):
        if key not in ("exc_info", "stack_info"):  # Handle separately
            # Format value based on type
            if isinstance(value, str):
                parts.append(f"{key}={value}")
            elif isinstance(value, (int, float, bool)):
                parts.append(f"{key}={value}")
            else:
                parts.append(f"{key}={repr(value)}")

    log_line = " ".join(parts)

    # Add exception info if present
    exc_info = event_dict.get("exc_info")
    if exc_info:
        log_line += "\n" + "".join(exc_info)

    return log_line + "\n"


def configure_logging(
    log_level: str = "INFO",
    log_format: str = "text",
    json_format: bool = False
) -> None:
    """
    Configure structlog for the application.

    Args:
        log_level: Logging level (DEBUG, INFO, WARNING, ERROR, CRITICAL)
        log_format: Format type - "text" (default) or "json"
        json_format: Legacy parameter, use log_format instead
    """
    # Determine if JSON format should be used
    use_json = json_format or log_format.lower() == "json"

    # Configure standard logging
    logging.basicConfig(
        format="%(message)s",
        stream=sys.stdout,
        level=getattr(logging, log_level.upper(), logging.INFO),
        force=True,  # Force reconfiguration
    )

    # Build processors list
    processors = [
        # Add log level
        structlog.stdlib.add_log_level,
        # Add logger name
        structlog.stdlib.add_logger_name,
        # Add timestamp
        structlog.processors.TimeStamper(fmt="%Y-%m-%d %H:%M:%S"),
        # Handle exceptions
        structlog.processors.StackInfoRenderer(),
        structlog.processors.format_exc_info,
        # Decode unicode
        structlog.processors.UnicodeDecoder(),
    ]

    # Add format-specific processors
    if use_json:
        # JSON format for production/log aggregation
        processors.append(structlog.processors.JSONRenderer())
    else:
        # Human-readable format with colors
        processors.append(add_color)
        processors.append(human_readable_renderer)

    # Configure structlog
    structlog.configure(
        processors=processors,
        wrapper_class=structlog.stdlib.BoundLogger,
        context_class=dict,
        logger_factory=structlog.stdlib.LoggerFactory(),
        cache_logger_on_first_use=True,
    )


def get_logger(name: str = None, **context) -> structlog.stdlib.BoundLogger:
    """
    Get a configured logger instance.

    Args:
        name: Logger name (usually __name__)
        **context: Additional context to bind to the logger

    Returns:
        Configured logger instance

    Example:
        logger = get_logger(__name__)
        logger.info("Processing request", request_id="123", user_id="abc")
    """
    if name:
        return structlog.get_logger(name, **context)
    return structlog.get_logger(**context)


def bind_context(**context) -> None:
    """
    Bind global context to all loggers.

    Example:
        bind_context(request_id="123", session_id="abc")
    """
    structlog.contextvars.bind_contextvars(**context)


def clear_context() -> None:
    """Clear all bound context variables."""
    structlog.contextvars.clear_contextvars()


class RequestLogger:
    """
    Helper class for request-scoped logging with automatic context management.

    Example:
        request_logger = RequestLogger(request_id="123", session_id="abc")
        request_logger.info("Processing request")
        # Output: [2025-01-14 10:30:45] [INFO] [api] Processing request request_id=123 session_id=abc
    """

    def __init__(self, **context):
        """Initialize with request context."""
        self._context = context
        self._logger = get_logger(**context)

    def info(self, event: str, **kwargs):
        """Log info message with context."""
        self._logger.info(event, **{**self._context, **kwargs})

    def debug(self, event: str, **kwargs):
        """Log debug message with context."""
        self._logger.debug(event, **{**self._context, **kwargs})

    def warning(self, event: str, **kwargs):
        """Log warning message with context."""
        self._logger.warning(event, **{**self._context, **kwargs})

    def error(self, event: str, **kwargs):
        """Log error message with context."""
        self._logger.error(event, **{**self._context, **kwargs})

    def exception(self, event: str, **kwargs):
        """Log exception message with context."""
        self._logger.exception(event, **{**self._context, **kwargs})
