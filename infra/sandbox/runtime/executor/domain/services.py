"""
Domain Services

Business logic services that don't naturally fit within entities or value objects.
"""

import hashlib
import os
from pathlib import Path
from typing import List, Optional
from datetime import datetime
import mimetypes

from executor.domain.value_objects import Artifact, ArtifactType


class ArtifactCollector:
    """
    Service for collecting artifacts generated during execution.

    Scans the workspace directory and creates Artifact value objects
    for files created during execution.
    """

    def __init__(self, workspace_path: Path, base_snapshot: Optional[set] = None):
        """
        Initialize the artifact collector.

        Args:
            workspace_path: Path to the workspace directory
            base_snapshot: Optional snapshot of files before execution
        """
        self.workspace_path = workspace_path
        self.base_snapshot = base_snapshot or set()
        self.mime_types = mimetypes.MimeTypes()

    def collect_artifacts(
        self, include_hidden: bool = False, include_temp: bool = False
    ) -> List[Artifact]:
        """
        Collect all artifacts in the workspace.

        Args:
            include_hidden: Whether to include hidden files (starting with .)
            include_temp: Whether to include temporary files

        Returns:
            List of Artifact value objects
        """
        artifacts = []
        current_files = self._snapshot_workspace()

        # Find new files (created during execution)
        new_files = current_files - self.base_snapshot

        for file_path_str in new_files:
            file_path = self.workspace_path / file_path_str

            # Skip if not a file
            if not file_path.is_file():
                continue

            # Skip hidden files if requested
            if not include_hidden and file_path.name.startswith("."):
                continue

            # Skip temp files if requested
            if not include_temp and file_path.name.startswith(".tmp"):
                continue

            artifact = self._create_artifact(file_path)
            if artifact:
                artifacts.append(artifact)

        return artifacts

    def _snapshot_workspace(self) -> set:
        """Create a snapshot of current files in workspace."""
        files = set()
        for item in self.workspace_path.rglob("*"):
            if item.is_file():
                # Get relative path from workspace
                rel_path = item.relative_to(self.workspace_path)
                files.add(str(rel_path))
        return files

    def _create_artifact(self, file_path: Path) -> Optional[Artifact]:
        """Create an Artifact value object from a file path."""
        try:
            stat = file_path.stat()
            rel_path = str(file_path.relative_to(self.workspace_path))

            # Determine MIME type
            mime_type, _ = self.mime_types.guess_type(file_path)
            if not mime_type:
                mime_type = "application/octet-stream"

            # Determine artifact type
            artifact_type = self._determine_artifact_type(rel_path)

            # Calculate checksum
            checksum = self._calculate_checksum(file_path)

            return Artifact(
                path=rel_path,
                size=stat.st_size,
                mime_type=mime_type,
                type=artifact_type,
                created_at=datetime.fromtimestamp(stat.st_ctime),
                checksum=checksum,
            )
        except Exception:
            return None

    def _determine_artifact_type(self, path: str) -> ArtifactType:
        """Determine the type of artifact based on path."""
        if path.startswith(".tmp") or path.startswith("/tmp"):
            return ArtifactType.TEMP
        if "log" in path.lower() or path.endswith(".log"):
            return ArtifactType.LOG
        if path.startswith("output") or path.endswith(".out"):
            return ArtifactType.OUTPUT
        return ArtifactType.ARTIFACT

    def _calculate_checksum(self, file_path: Path) -> str:
        """Calculate SHA256 checksum of a file."""
        sha256_hash = hashlib.sha256()
        with open(file_path, "rb") as f:
            for byte_block in iter(lambda: f.read(4096), b""):
                sha256_hash.update(byte_block)
        return sha256_hash.hexdigest()

    def snapshot(self) -> set:
        """Create a snapshot of current workspace state."""
        return self._snapshot_workspace()
