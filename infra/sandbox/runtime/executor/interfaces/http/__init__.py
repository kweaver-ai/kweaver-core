"""
HTTP Interface

FastAPI endpoints and HTTP-related adapters.
"""

from .rest import app, create_app, main

__all__ = ["app", "create_app", "main"]
