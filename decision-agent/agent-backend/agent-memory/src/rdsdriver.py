# Stub module for rdsdriver - maps to pymysql for compatibility
from pymysql import (
    connect,
    cursors,
    Error,
    InterfaceError,
    DatabaseError,
    OperationalError,
    IntegrityError,
    InternalError,
    ProgrammingError,
    NotSupportedError,
    converters,
    Binary,
)
from pymysql.cursors import DictCursor

# escape_string is removed in newer pymysql versions, use pymysql.escape_string instead
escape_string = None

__all__ = [
    "connect",
    "cursors",
    "Error",
    "InterfaceError",
    "DatabaseError",
    "OperationalError",
    "IntegrityError",
    "InternalError",
    "ProgrammingError",
    "NotSupportedError",
    "escape_string",
    "converters",
    "Binary",
    "DictCursor",
]
