import sys
import os

sys.path.insert(0, os.path.join(os.path.dirname(__file__), "../../.."))

import pytest
from unittest.mock import patch, AsyncMock, MagicMock
from src.utils.env import getenv_int


class TestGetenvInt:
    def test_getenv_int_with_valid_value(self):
        """Test with valid integer environment variable"""
        with patch.dict(os.environ, {"TEST_VAR": "123"}):
            result = getenv_int("TEST_VAR")
            assert result == 123

    def test_getenv_int_with_missing_var(self):
        """Test with missing environment variable"""
        result = getenv_int("NONEXISTENT_VAR")
        assert result is None

    def test_getenv_int_with_missing_var_and_default(self):
        """Test with missing environment variable and default value"""
        result = getenv_int("NONEXISTENT_VAR", default=42)
        assert result == 42

    def test_getenv_int_with_invalid_value(self):
        """Test with non-integer environment variable"""
        with patch.dict(os.environ, {"TEST_VAR": "not_a_number"}):
            result = getenv_int("TEST_VAR")
            assert result is None

    def test_getenv_int_with_invalid_value_and_default(self):
        """Test with non-integer environment variable and default value"""
        with patch.dict(os.environ, {"TEST_VAR": "not_a_number"}):
            result = getenv_int("TEST_VAR", default=99)
            assert result == 99

    def test_getenv_int_with_zero_value(self):
        """Test with zero value"""
        with patch.dict(os.environ, {"TEST_VAR": "0"}):
            result = getenv_int("TEST_VAR")
            assert result == 0

    def test_getenv_int_with_negative_value(self):
        """Test with negative value"""
        with patch.dict(os.environ, {"TEST_VAR": "-100"}):
            result = getenv_int("TEST_VAR")
            assert result == -100

    def test_getenv_int_raise_error_on_missing(self):
        """Test raise_error=True with missing environment variable"""
        with pytest.raises(ValueError) as exc:
            getenv_int("NONEXISTENT_VAR", raise_error=True)
        assert "not set" in str(exc.value)

    def test_getenv_int_raise_error_on_invalid(self):
        """Test raise_error=True with invalid integer value"""
        with patch.dict(os.environ, {"TEST_VAR": "invalid"}):
            with pytest.raises(ValueError) as exc:
                getenv_int("TEST_VAR", raise_error=True)
            assert "not a valid integer" in str(exc.value)

    def test_getenv_int_with_whitespace(self):
        """Test with whitespace around value"""
        with patch.dict(os.environ, {"TEST_VAR": "  456  "}):
            result = getenv_int("TEST_VAR")
            assert result == 456

    def test_getenv_int_with_float_string(self):
        """Test with float string value"""
        with patch.dict(os.environ, {"TEST_VAR": "3.14"}):
            result = getenv_int("TEST_VAR", default=0)
            assert result == 0  # Should fail conversion and return default
