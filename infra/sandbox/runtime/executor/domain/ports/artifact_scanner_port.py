"""
Artifact Scanner Port Interface

Defines the contract for artifact collection operations.
This is an output port - implemented by infrastructure layer.
"""

from abc import ABC, abstractmethod
from pathlib import Path
from typing import List

from executor.domain.value_objects import Artifact


class IArtifactScannerPort(ABC):
    """
    Port interface for artifact scanning operations.

    Defines the contract for collecting artifacts generated
    during code execution from the workspace directory.
    """

    @abstractmethod
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
        pass

    @abstractmethod
    def snapshot(self, workspace_path: Path) -> set:
        """
        Create a snapshot of current workspace state.

        Args:
            workspace_path: Path to workspace directory

        Returns:
            Set of relative file paths
        """
        pass
