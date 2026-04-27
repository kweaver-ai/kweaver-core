"""
Artifact scanner for workspace file collection.

Scans workspace directory for generated files after execution,
extracts metadata, filters hidden files, and classifies artifacts.

Implements T083-T087 [US6]: Artifact Collection
"""

import mimetypes
import hashlib
from datetime import datetime
from pathlib import Path
from typing import List

from executor.domain.value_objects import Artifact, ArtifactType
from executor.domain.ports import IArtifactScannerPort
from executor.infrastructure.logging.logging_config import get_logger

logger = get_logger()


# T084 [US6]: Filter out hidden files and directories
def _is_hidden(file_path: Path) -> bool:
    """
    Check if a file or directory is hidden.

    A file is considered hidden if:
    - The file name starts with '.'
    - Any parent directory name starts with '.'

    Args:
        file_path: Absolute or relative path to check

    Returns:
        True if file is hidden, False otherwise
    """
    # Check if any component starts with '.'
    for part in file_path.parts:
        if part.startswith("."):
            return True
    return False


# T085 [US6]: Calculate relative path from workspace root
def _get_relative_path(file_path: Path, workspace_path: Path) -> str:
    """
    Calculate relative path from workspace root.

    Args:
        file_path: Absolute path to file
        workspace_path: Absolute path to workspace root

    Returns:
        Relative path string with forward slashes
    """
    try:
        relative = file_path.relative_to(workspace_path)
        # Convert to string with forward slashes (cross-platform)
        return str(relative).replace("\\", "/")
    except ValueError:
        # File is not under workspace path
        logger.warning(
            "File not under workspace path",
            file_path=str(file_path),
            workspace_path=str(workspace_path),
        )
        return str(file_path)


# T086 [US6]: Extract file metadata
def _extract_metadata(
    file_path: Path,
    workspace_path: Path,
    include_checksum: bool = False,
) -> Artifact:
    """
    Extract metadata from a file.

    Args:
        file_path: Absolute path to file
        workspace_path: Absolute path to workspace root
        include_checksum: Whether to calculate SHA256 checksum

    Returns:
        Artifact value object
    """
    # Get relative path
    relative_path = _get_relative_path(file_path, workspace_path)

    # Get file size
    size = file_path.stat().st_size

    # Detect MIME type
    mime_type, _ = mimetypes.guess_type(str(file_path))
    if mime_type is None:
        mime_type = "application/octet-stream"

    # Get creation time
    created_at = datetime.fromtimestamp(file_path.stat().st_ctime)

    # Calculate checksum (optional)
    checksum = None
    if include_checksum:
        checksum = _calculate_checksum(file_path)

    # Classify artifact type
    artifact_type = _classify_artifact_type(relative_path)

    return Artifact(
        path=relative_path,
        size=size,
        mime_type=mime_type,
        type=artifact_type,
        created_at=created_at,
        checksum=checksum,
    )


def _calculate_checksum(file_path: Path) -> str:
    """
    Calculate SHA256 checksum of a file.

    Args:
        file_path: Path to file

    Returns:
        Hex-encoded SHA256 checksum
    """
    sha256 = hashlib.sha256()
    with open(file_path, "rb") as f:
        for chunk in iter(lambda: f.read(8192), b""):
            sha256.update(chunk)
    return sha256.hexdigest()


# T087 [US6]: Classify artifact by type
def _classify_artifact_type(relative_path: str) -> ArtifactType:
    """
    Classify artifact type based on path.

    Classification rules:
    - ArtifactType.OUTPUT: Files in output/ directory
    - ArtifactType.LOG: Files in logs/ directory or ending with .log
    - ArtifactType.ARTIFACT: All other files

    Args:
        relative_path: Relative path from workspace root

    Returns:
        ArtifactType enum value
    """
    path_lower = relative_path.lower()

    # Output files
    if path_lower.startswith("output/"):
        return ArtifactType.OUTPUT

    # Log files
    if path_lower.startswith("logs/") or path_lower.endswith(".log"):
        return ArtifactType.LOG

    # Default: artifact
    return ArtifactType.ARTIFACT


# T083 [US6]: Create artifact scanner with collect_artifacts() function
def collect_artifacts(
    workspace_path: Path,
    include_checksum: bool = False,
) -> List[Artifact]:
    """
    Scan workspace for generated files and collect metadata.

    Performs recursive scan of workspace directory, collecting all
    non-hidden files with metadata.

    Features:
    - Recursive directory scanning
    - Hidden file exclusion (files starting with '.')
    - Relative path calculation
    - File metadata extraction (size, mime_type, created_at)
    - Artifact type classification (artifact/log/output)
    - Optional checksum calculation

    Args:
        workspace_path: Path to workspace directory
        include_checksum: Whether to calculate SHA256 checksums (slower)

    Returns:
        List of Artifact value objects

    Example:
        >>> workspace = Path("/workspace")
        >>> artifacts = collect_artifacts(workspace)
        >>> for artifact in artifacts:
        ...     print(f"{artifact.path}: {artifact.size} bytes")
    """
    workspace_path = Path(workspace_path)
    workspace_str = str(workspace_path)

    # If workspace is S3 path, use local /workspace directory instead
    # S3 paths are handled by the artifact storage layer, not filesystem checks
    if workspace_str.startswith("s3:/") or workspace_str.startswith("s3://"):
        logger.info("Workspace is S3 path, using local /workspace directory for artifact scanning", s3_path=workspace_str)
        workspace_path = Path("/workspace")

    if not workspace_path.exists():
        logger.warning("Workspace path does not exist", path=str(workspace_path))
        return []

    if not workspace_path.is_dir():
        logger.error("Workspace path is not a directory", path=str(workspace_path))
        return []

    artifacts = []

    # Use rglob for recursive scanning
    for file_path in workspace_path.rglob("*"):
        # Skip directories (only process files)
        if not file_path.is_file():
            continue

        # T084: Filter out hidden files
        if _is_hidden(file_path):
            logger.debug("Skipping hidden file", path=str(file_path))
            continue

        # Extract metadata
        try:
            artifact = _extract_metadata(file_path, workspace_path, include_checksum)
            artifacts.append(artifact)
        except Exception as e:
            logger.warning(
                "Failed to extract metadata from file",
                path=str(file_path),
                error=str(e),
            )
            continue

    logger.info(
        "Artifact collection complete",
        workspace_path=str(workspace_path),
        artifact_count=len(artifacts),
    )

    return artifacts


def get_artifact_paths(workspace_path: Path) -> List[str]:
    """
    Get list of artifact paths (simple string list).

    Convenience function for when only paths are needed.
    Compatible with ExecutionResult.artifacts field.

    Args:
        workspace_path: Path to workspace directory

    Returns:
        List of relative file paths as strings
    """
    artifacts = collect_artifacts(workspace_path)
    return [artifact.path for artifact in artifacts]


class ArtifactScanner(IArtifactScannerPort):
    """
    Artifact scanner that implements the IArtifactScannerPort interface.

    This class provides an adapter between the functional artifact_scanner module
    and the port interface required by the hexagonal architecture.
    """

    def collect_artifacts(
        self,
        workspace_path: Path,
        include_hidden: bool = False,
        include_temp: bool = False,
    ) -> List[Artifact]:
        """
        Scan workspace for generated artifacts.

        Args:
            workspace_path: Path to workspace directory
            include_hidden: Whether to include hidden files
            include_temp: Whether to include temporary files

        Returns:
            List of Artifact value objects

        Raises:
            Exception: For scanning errors
        """
        return collect_artifacts(workspace_path, include_checksum=False)

    def snapshot(self, workspace_path: Path) -> set:
        """
        Create a snapshot of current workspace state.

        Args:
            workspace_path: Path to workspace directory

        Returns:
            Set of relative file paths
        """
        workspace_path = Path(workspace_path)
        workspace_str = str(workspace_path)

        # If workspace is S3 path, use local /workspace directory instead
        if workspace_str.startswith("s3:/") or workspace_str.startswith("s3://"):
            logger.info("Workspace is S3 path, using local /workspace directory for snapshot", s3_path=workspace_str)
            workspace_path = Path("/workspace")

        if not workspace_path.exists():
            return set()

        if not workspace_path.is_dir():
            return set()

        snapshot = set()

        # Use rglob for recursive scanning
        for file_path in workspace_path.rglob("*"):
            # Skip directories (only process files)
            if not file_path.is_file():
                continue

            # Skip hidden files
            if _is_hidden(file_path):
                continue

            # Get relative path
            relative_path = _get_relative_path(file_path, workspace_path)
            snapshot.add(relative_path)

        return snapshot
