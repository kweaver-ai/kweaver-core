"""
Interfaces Layer

Driving adapters that initiate interactions with the system.
Contains HTTP endpoints and other interface implementations.
"""

from .http import app

__all__ = ["app"]
