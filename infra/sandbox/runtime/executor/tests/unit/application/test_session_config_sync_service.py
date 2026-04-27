"""
Tests for session config sync service.
"""

from datetime import datetime, timezone
from pathlib import Path
from tempfile import TemporaryDirectory
from unittest.mock import patch

import pytest

from executor.application.services.session_config_sync_service import (
    InstalledDependency,
    SessionConfigSyncRequest,
    SessionConfigSyncService,
    SessionConfigValidationError,
)


class TestSessionConfigSyncService:
    """Tests for SessionConfigSyncService."""

    @pytest.fixture
    def service(self):
        with TemporaryDirectory() as tmpdir:
            install_path = Path(tmpdir) / "deps"
            cache_path = Path(tmpdir) / "pip-cache"
            yield SessionConfigSyncService(
                install_path=install_path,
                pip_cache_path=cache_path,
            )

    @pytest.mark.asyncio
    async def test_sync_empty_dependencies_resets_install_dir(self, service):
        request = SessionConfigSyncRequest(
            session_id="sess_1",
            language_runtime="python3.11",
            python_package_index_url="https://pypi.org/simple/",
            dependencies=[],
            sync_mode="replace",
        )

        result = await service.sync(request)

        assert result.status == "completed"
        assert result.installed_dependencies == []
        assert service._install_path.exists()

    def test_reset_install_directory_preserves_root_directory(self, service):
        service._install_path.mkdir(parents=True, exist_ok=True)
        service._pip_cache_path.mkdir(parents=True, exist_ok=True)
        (service._install_path / "obsolete.txt").write_text("old")
        nested_dir = service._install_path / "nested"
        nested_dir.mkdir()
        (nested_dir / "package.py").write_text("print('stale')")
        (service._pip_cache_path / "cache.bin").write_text("cache")

        service._reset_install_directory()

        assert service._install_path.exists()
        assert service._pip_cache_path.exists()
        assert list(service._install_path.iterdir()) == []
        assert list(service._pip_cache_path.iterdir()) == []

    @pytest.mark.asyncio
    async def test_sync_invokes_pip_and_scans_installed_distributions(self, service):
        request = SessionConfigSyncRequest(
            session_id="sess_1",
            language_runtime="python3.11",
            python_package_index_url="https://mirror.example/simple",
            dependencies=["requests==2.31.0"],
            sync_mode="replace",
        )

        class FakeProcess:
            returncode = 0

            async def communicate(self):
                return (b"ok", b"")

        fake_distribution = type(
            "FakeDistribution",
            (),
            {
                "metadata": {"Name": "requests"},
                "version": "2.31.0",
            },
        )()

        with patch(
            "executor.application.services.session_config_sync_service.asyncio.create_subprocess_exec",
            return_value=FakeProcess(),
        ) as subprocess_mock, patch(
            "executor.application.services.session_config_sync_service.importlib.metadata.distributions",
            return_value=[fake_distribution],
        ):
            result = await service.sync(request)

        assert subprocess_mock.await_count == 1
        assert len(result.installed_dependencies) == 1
        assert result.installed_dependencies[0].name == "requests"
        assert result.installed_dependencies[0].version == "2.31.0"

    def test_get_pip_python_executable_falls_back_to_base_python(self, service):
        with patch(
            "executor.application.services.session_config_sync_service.importlib.util.find_spec",
            return_value=None,
        ), patch(
            "executor.application.services.session_config_sync_service.sys.base_prefix",
            "/usr/local",
        ):
            assert service._get_pip_python_executable() == "/usr/local/bin/python3"

    @pytest.mark.asyncio
    async def test_sync_rejects_non_python_runtime(self, service):
        request = SessionConfigSyncRequest(
            session_id="sess_1",
            language_runtime="nodejs20",
            python_package_index_url="https://pypi.org/simple/",
            dependencies=["requests==2.31.0"],
            sync_mode="replace",
        )

        with pytest.raises(SessionConfigValidationError):
            await service.sync(request)
