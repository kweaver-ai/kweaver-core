"""
Environment configuration for sandbox-executor.

Loads configuration from environment variables using pydantic-settings.
"""

from pydantic import BaseModel, Field
from typing import Literal


class Settings(BaseModel):
    """Application settings loaded from environment variables."""

    # Control Plane Configuration
    control_plane_url: str = Field(
        default="http://localhost:8000",
        description="Control Plane base URL for callbacks",
    )
    internal_api_token: str = Field(
        default="dev_token_only_for_testing", description="API token for internal callbacks"
    )

    # Executor Configuration
    executor_port: int = Field(default=8080, ge=1024, le=65535, description="HTTP API port")
    log_level: Literal["DEBUG", "INFO", "WARNING", "ERROR"] = Field(
        default="INFO", description="Logging level"
    )
    log_format: Literal["text", "json"] = Field(
        default="text", description="Logging format - text for human-readable, json for structured logs"
    )
    workspace_path: str = Field(
        default="/workspace", description="Workspace directory for file operations"
    )
    dependency_install_path: str = Field(
        default="/opt/sandbox-venv",
        description="Directory for dynamically installed session dependencies",
    )
    pip_cache_path: str = Field(
        default="/tmp/pip-cache",
        description="Cache directory for pip install operations",
    )

    # Execution Configuration
    default_timeout: int = Field(default=30, ge=1, le=3600, description="Default timeout in seconds")
    max_timeout: int = Field(default=3600, ge=1, le=3600, description="Maximum timeout in seconds")

    # Heartbeat Configuration
    heartbeat_interval: int = Field(default=5, ge=1, le=60, description="Heartbeat interval in seconds")

    # Retry Configuration
    max_retries: int = Field(default=5, ge=1, le=10, description="Maximum callback retry attempts")
    base_retry_delay: float = Field(default=1.0, ge=0.1, le=10.0, description="Base retry delay in seconds")
    max_retry_delay: float = Field(default=10.0, ge=1.0, le=60.0, description="Maximum retry delay in seconds")


# Global settings instance
# For production, use pydantic-settings to load from environment
# For MVP, use default values
settings = Settings()
