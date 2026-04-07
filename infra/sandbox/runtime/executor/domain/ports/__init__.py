"""
Domain Ports

Port interfaces defining contracts between layers.
All dependencies on external systems are abstracted through ports.
"""

from .executor_port import IExecutorPort
from .callback_port import ICallbackPort
from .isolation_port import IIsolationPort
from .artifact_scanner_port import IArtifactScannerPort
from .heartbeat_port import IHeartbeatPort
from .lifecycle_port import ILifecyclePort
# Value objects are exported from value_objects module
from ..value_objects import (
    HeartbeatSignal,
    ContainerLifecycleEvent,
)

__all__ = [
    # Executor
    "IExecutorPort",
    # Callback
    "ICallbackPort",
    # Isolation
    "IIsolationPort",
    # Artifact Scanner
    "IArtifactScannerPort",
    # Heartbeat
    "IHeartbeatPort",
    # Lifecycle
    "ILifecyclePort",
    # Value objects (for convenience)
    "HeartbeatSignal",
    "ContainerLifecycleEvent",
]
