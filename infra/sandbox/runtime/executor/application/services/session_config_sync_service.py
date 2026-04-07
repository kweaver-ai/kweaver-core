"""
Session config sync service.

Synchronizes session-level Python dependency configuration into the executor
runtime by performing a full install into the target dependency directory.
"""

from __future__ import annotations

import asyncio
import importlib.util
import importlib.metadata
import os
import shutil
import sys
from dataclasses import dataclass
from datetime import datetime, timezone
from pathlib import Path

import structlog


logger = structlog.get_logger(__name__)


@dataclass(frozen=True)
class SessionConfigSyncRequest:
    """Request for synchronizing session dependency configuration."""

    session_id: str
    language_runtime: str
    python_package_index_url: str
    dependencies: list[str]
    sync_mode: str


@dataclass(frozen=True)
class InstalledDependency:
    """Installed dependency metadata."""

    name: str
    version: str
    install_location: str
    install_time: datetime
    is_from_template: bool = False


@dataclass(frozen=True)
class SessionConfigSyncResult:
    """Result for session dependency synchronization."""

    status: str
    installed_dependencies: list[InstalledDependency]
    error: str
    started_at: datetime
    completed_at: datetime


class SessionConfigSyncError(Exception):
    """Base error for dependency sync failures."""


class SessionConfigValidationError(SessionConfigSyncError):
    """Invalid sync request."""


class SessionDependencyInstallError(SessionConfigSyncError):
    """Dependency installation failed."""


class SessionConfigSyncService:
    """Application service for session dependency synchronization."""

    def __init__(self, install_path: Path, pip_cache_path: Path):
        self._install_path = install_path
        self._pip_cache_path = pip_cache_path

    async def sync(self, request: SessionConfigSyncRequest) -> SessionConfigSyncResult:
        """Synchronize final dependency state for a session."""
        started_at = datetime.now(timezone.utc)
        self._validate_request(request)

        logger.info(
            "Starting session dependency sync",
            session_id=request.session_id,
            language_runtime=request.language_runtime,
            dependency_count=len(request.dependencies),
            sync_mode=request.sync_mode,
            install_path=str(self._install_path),
        )

        await asyncio.to_thread(self._reset_install_directory)

        if request.dependencies:
            await self._install_dependencies(request)

        installed_dependencies = await asyncio.to_thread(self._scan_installed_dependencies)
        completed_at = datetime.now(timezone.utc)

        logger.info(
            "Session dependency sync completed",
            session_id=request.session_id,
            dependency_count=len(installed_dependencies),
            completed_at=completed_at.isoformat(),
        )

        return SessionConfigSyncResult(
            status="completed",
            installed_dependencies=installed_dependencies,
            error="",
            started_at=started_at,
            completed_at=completed_at,
        )

    def _validate_request(self, request: SessionConfigSyncRequest) -> None:
        if not request.session_id:
            raise SessionConfigValidationError("session_id is required")
        if not request.language_runtime.startswith("python"):
            raise SessionConfigValidationError("only Python runtime supports dependency sync")
        if request.sync_mode not in {"replace", "merge"}:
            raise SessionConfigValidationError("sync_mode must be replace or merge")
        if not request.python_package_index_url:
            raise SessionConfigValidationError("python_package_index_url is required")

    def _reset_install_directory(self) -> None:
        self._install_path.mkdir(parents=True, exist_ok=True)
        self._clear_directory(self._install_path)
        self._pip_cache_path.mkdir(parents=True, exist_ok=True)
        self._clear_directory(self._pip_cache_path)

    def _clear_directory(self, path: Path) -> None:
        """Clear directory contents without removing the directory itself."""
        for child in path.iterdir():
            if child.is_dir() and not child.is_symlink():
                shutil.rmtree(child)
            else:
                child.unlink()

    async def _install_dependencies(self, request: SessionConfigSyncRequest) -> None:
        command = [
            self._get_pip_python_executable(),
            "-m",
            "pip",
            "install",
            "--target",
            str(self._install_path),
            "--cache-dir",
            str(self._pip_cache_path),
            "--disable-pip-version-check",
            "--no-warn-script-location",
            "--index-url",
            request.python_package_index_url,
            *request.dependencies,
        ]

        env = os.environ.copy()
        env["PYTHONPATH"] = self._build_pythonpath(env.get("PYTHONPATH"))

        process = await asyncio.create_subprocess_exec(
            *command,
            stdout=asyncio.subprocess.PIPE,
            stderr=asyncio.subprocess.PIPE,
            env=env,
        )
        stdout, stderr = await process.communicate()

        if process.returncode != 0:
            error = stderr.decode("utf-8", errors="replace").strip() or stdout.decode(
                "utf-8", errors="replace"
            ).strip()
            logger.error(
                "Dependency installation failed",
                session_id=request.session_id,
                returncode=process.returncode,
                error=error,
            )
            raise SessionDependencyInstallError(error)

    def _get_pip_python_executable(self) -> str:
        """
        Return a Python executable that can run `-m pip`.

        The executor process runs inside a uv-managed virtualenv where `pip`
        may be intentionally absent. In that case fall back to the base
        interpreter used to create the virtualenv.
        """
        if importlib.util.find_spec("pip") is not None:
            return sys.executable

        return os.path.join(sys.base_prefix, "bin", "python3")

    def _scan_installed_dependencies(self) -> list[InstalledDependency]:
        installed_at = datetime.now(timezone.utc)
        dependencies: list[InstalledDependency] = []

        for distribution in importlib.metadata.distributions(path=[str(self._install_path)]):
            name = distribution.metadata.get("Name") or distribution.metadata.get("Summary")
            if not name:
                continue
            dependencies.append(
                InstalledDependency(
                    name=name,
                    version=distribution.version,
                    install_location=str(self._install_path),
                    install_time=installed_at,
                    is_from_template=False,
                )
            )

        dependencies.sort(key=lambda dep: dep.name.lower())
        return dependencies

    def _build_pythonpath(self, existing_pythonpath: str | None) -> str:
        parts = [str(self._install_path)]
        if existing_pythonpath:
            parts.append(existing_pythonpath)
        return ":".join(parts)
