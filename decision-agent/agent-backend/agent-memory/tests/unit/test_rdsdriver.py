import sys
import os

sys.path.insert(0, os.path.join(os.path.dirname(__file__), "../.."))

import pytest


class TestRdsdriver:
    def test_connect_importable(self):
        """Test that connect is importable"""
        from src.rdsdriver import connect

        assert connect is not None

    def test_cursors_importable(self):
        """Test that cursors is importable"""
        from src.rdsdriver import cursors

        assert cursors is not None

    def test_errors_importable(self):
        """Test that error classes are importable"""
        from src.rdsdriver import (
            Error,
            InterfaceError,
            DatabaseError,
            OperationalError,
            IntegrityError,
            InternalError,
            ProgrammingError,
            NotSupportedError,
        )

        assert Error is not None
        assert InterfaceError is not None
        assert DatabaseError is not None
        assert OperationalError is not None
        assert IntegrityError is not None
        assert InternalError is not None
        assert ProgrammingError is not None
        assert NotSupportedError is not None

    def test_converters_importable(self):
        """Test that converters are importable"""
        from src.rdsdriver import converters, Binary

        assert converters is not None
        assert Binary is not None

    def test_dict_cursor_importable(self):
        """Test that DictCursor is importable"""
        from src.rdsdriver import DictCursor

        assert DictCursor is not None

    def test_escape_string_is_none(self):
        """Test that escape_string is None"""
        from src.rdsdriver import escape_string

        assert escape_string is None

    def test_all_exports_defined(self):
        """Test that all expected exports are defined"""
        from src.rdsdriver import __all__

        expected_exports = [
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

        assert set(__all__) == set(expected_exports)
