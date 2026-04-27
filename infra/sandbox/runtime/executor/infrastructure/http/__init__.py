"""
HTTP Infrastructure

HTTP client adapters for Control Plane communication.
"""

from .callback_client import CallbackClient, get_callback_client

__all__ = ["CallbackClient", "get_callback_client"]
