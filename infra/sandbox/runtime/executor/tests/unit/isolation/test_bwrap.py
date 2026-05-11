"""
Unit tests for Bubblewrap isolation module.

Tests functions that can be unit tested without requiring bwrap to be installed.
"""

import pytest
from unittest.mock import patch, MagicMock
import subprocess

from executor.infrastructure.isolation.bwrap import (
    check_bwrap_available,
    get_bwrap_version,
)


class TestCheckBwrapAvailable:
    """Tests for check_bwrap_available function."""

    def test_bwrap_available(self):
        """Test when bwrap is available."""
        with patch('shutil.which') as mock_which:
            mock_which.return_value = "/usr/bin/bwrap"

            result = check_bwrap_available()

            assert result is True
            mock_which.assert_called_once_with("bwrap")

    def test_bwrap_not_available(self):
        """Test when bwrap is not available."""
        with patch('shutil.which') as mock_which:
            mock_which.return_value = None

            with pytest.raises(RuntimeError, match="Bubblewrap.*is not installed"):
                check_bwrap_available()


class TestGetBwrapVersion:
    """Tests for get_bwrap_version function."""

    def test_get_version_success(self):
        """Test getting bwrap version successfully."""
        with patch('subprocess.run') as mock_run:
            mock_result = MagicMock()
            mock_result.returncode = 0
            mock_result.stdout = "bwrap 1.7.0\n"
            mock_run.return_value = mock_result

            version = get_bwrap_version()

            assert version == "1.7.0"
            mock_run.assert_called_once()

    def test_get_version_failure(self):
        """Test when bwrap version command fails."""
        with patch('subprocess.run') as mock_run:
            mock_result = MagicMock()
            mock_result.returncode = 1
            mock_result.stdout = ""
            mock_run.return_value = mock_result

            with pytest.raises(RuntimeError, match="Failed to get bwrap version"):
                get_bwrap_version()

    def test_get_version_timeout(self):
        """Test timeout when getting bwrap version."""
        with patch('subprocess.run') as mock_run:
            mock_run.side_effect = subprocess.TimeoutExpired("bwrap", 5)

            with pytest.raises(RuntimeError, match="Timeout"):
                get_bwrap_version()

    def test_get_version_not_found(self):
        """Test when bwrap is not found."""
        with patch('subprocess.run') as mock_run:
            mock_run.side_effect = FileNotFoundError()

            with pytest.raises(RuntimeError, match="bwrap not found"):
                get_bwrap_version()


class TestBubblewrapRunnerInit:
    """Tests for BubblewrapRunner initialization."""

    def test_init(self):
        """Test BubblewrapRunner initialization."""
        from pathlib import Path
        from executor.infrastructure.isolation.bwrap import BubblewrapRunner

        workspace = Path("/tmp/workspace")
        runner = BubblewrapRunner(workspace)

        assert runner.workspace_path == workspace
        assert runner._base_args is not None
        assert "bwrap" in runner._base_args

    def test_base_args_include_isolation(self):
        """Test that base args include isolation flags."""
        from pathlib import Path
        from executor.infrastructure.isolation.bwrap import BubblewrapRunner

        workspace = Path("/tmp/workspace")
        runner = BubblewrapRunner(workspace)

        args = runner._base_args

        # Check for common isolation flags
        assert "--unshare-pid" in args or "--unshare-all" in args

    def test_base_args_include_go_path(self):
        """Test that Go installed under /usr/local/go/bin is available inside bwrap."""
        from pathlib import Path
        from executor.infrastructure.isolation.bwrap import BubblewrapRunner

        runner = BubblewrapRunner(Path("/tmp/workspace"))
        args = runner._base_args

        path_index = args.index("PATH")
        assert "/usr/local/go/bin" in args[path_index + 1]

    def test_base_args_bind_dependency_install_path(self):
        """Test that session dependency install path is available inside bwrap."""
        from pathlib import Path
        from executor.infrastructure.config import settings
        from executor.infrastructure.isolation.bwrap import BubblewrapRunner

        runner = BubblewrapRunner(Path("/tmp/workspace"))
        args = runner._base_args

        ro_bind_indices = [index for index, value in enumerate(args) if value == "--ro-bind"]
        ro_bind_pairs = [(args[index + 1], args[index + 2]) for index in ro_bind_indices]

        assert (settings.dependency_install_path, settings.dependency_install_path) in ro_bind_pairs

    def test_inject_env_args_adds_setenv_before_separator(self):
        """Test environment variables are injected into bwrap command."""
        from pathlib import Path
        from executor.infrastructure.isolation.bwrap import BubblewrapRunner

        runner = BubblewrapRunner(Path("/tmp/workspace"))
        cmd = ["bwrap", "--clearenv", "--", "python3", "-c", "print('hi')"]

        updated = runner._inject_env_args(
            cmd,
            {"EVENT_JSON": "{}", "PYTHONPATH": "/opt/sandbox-venv"},
        )

        separator_index = updated.index("--")
        prefix = updated[:separator_index]
        assert "--setenv" in prefix
        assert "EVENT_JSON" in prefix
        assert "PYTHONPATH" in prefix

    def test_build_base_args_uses_custom_working_directory(self):
        """Test custom working directory is propagated to bwrap chdir."""
        from pathlib import Path
        from executor.infrastructure.isolation.bwrap import BubblewrapRunner

        runner = BubblewrapRunner(Path("/tmp/workspace"))
        args = runner._build_base_args("/workspace/skill/mini-wiki")

        chdir_index = args.index("--chdir")
        assert args[chdir_index + 1] == "/workspace/skill/mini-wiki"
