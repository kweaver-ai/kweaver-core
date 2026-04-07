"""
Lifecycle Service

Manages container lifecycle events (startup/shutdown).
"""

import asyncio
import os
import signal
from datetime import datetime
from typing import Dict, Optional, List

import structlog

from executor.domain.ports import ILifecyclePort, ICallbackPort, IHeartbeatPort
from executor.domain.value_objects import ContainerLifecycleEvent, ExitReason


logger = structlog.get_logger(__name__)


class LifecycleService(ILifecyclePort):
    """
    Service for managing container lifecycle events.

    Handles:
    - Sending container_ready on startup
    - Sending container_exited on shutdown
    - Graceful shutdown on SIGTERM/SIGINT
    - Marking active executions as crashed
    """

    def __init__(
        self,
        callback_port: ICallbackPort,
        heartbeat_port: IHeartbeatPort,
        executor_port: int = 8080,
    ):
        """
        Initialize lifecycle service.

        Args:
            callback_port: Port for Control Plane callbacks
            heartbeat_port: Port for heartbeat management
            executor_port: HTTP API port
        """
        self._callback_port = callback_port
        self._heartbeat_port = heartbeat_port
        self._executor_port = executor_port
        self._is_shutting_down = False
        self._shutdown_complete = asyncio.Event()

        # T075: Detect container_id from environment
        self._container_id = self._get_container_id()
        self._pod_name = os.environ.get("POD_NAME", self._container_id)

    async def send_container_ready(self) -> bool:
        """
        Send container_ready event to Control Plane on startup.

        Should be called after HTTP server starts listening.
        This is now called from @app.on_event("startup") which ensures
        the HTTP server is ready before this method is invoked.

        Returns:
            True if successful, False otherwise
        """
        try:
            event = ContainerLifecycleEvent(
                event_type="ready",
                container_id=self._container_id,
                pod_name=self._pod_name,
                executor_port=self._executor_port,
                ready_at=datetime.now(),
            )

            logger.info(
                "Sending container_ready",
                container_id=self._container_id,
                executor_port=self._executor_port,
            )

            success = await self._callback_port.report_lifecycle(event)

            if success:
                logger.info("container_ready sent successfully")
            else:
                logger.warning("container_ready send failed")

            return success

        except Exception as e:
            logger.error(
                "Failed to send container_ready",
                error=str(e),
                exc_info=True,
            )
            return False

    async def send_container_exited(
        self,
        exit_code: int,
        exit_reason: str,
    ) -> bool:
        """
        Send container_exited event to Control Plane on shutdown.

        Args:
            exit_code: Process exit code
            exit_reason: Reason for exit (normal, sigterm, sigkill, oom_killed, error)

        Returns:
            True if successful, False otherwise
        """
        try:
            # Map string to ExitReason enum
            reason_enum = ExitReason(exit_reason)

            event = ContainerLifecycleEvent(
                event_type="exited",
                container_id=self._container_id,
                pod_name=self._pod_name,
                executor_port=self._executor_port,
                exit_code=exit_code,
                exit_reason=reason_enum,
                exited_at=datetime.now(),
            )

            logger.info(
                "Sending container_exited",
                container_id=self._container_id,
                exit_code=exit_code,
                exit_reason=exit_reason,
            )

            success = await self._callback_port.report_lifecycle(event)

            if success:
                logger.info("container_exited sent successfully")
            else:
                logger.warning("container_exited send failed")

            return success

        except Exception as e:
            logger.error(
                "Failed to send container_exited",
                error=str(e),
                exc_info=True,
            )
            return False

    async def shutdown(self, signum: Optional[int] = None) -> None:
        """
        Handle graceful shutdown.

        Marks active executions as crashed, sends container_exited,
        and waits for shutdown to complete.

        Args:
            signum: Signal number (SIGTERM=15, SIGKILL=9, etc.)
        """
        if self._is_shutting_down:
            logger.debug("Shutdown already in progress")
            return

        self._is_shutting_down = True

        # Determine exit code and reason from signal
        if signum == signal.SIGTERM:
            exit_code, exit_reason = 143, "sigterm"
        elif signum == signal.SIGKILL:
            exit_code, exit_reason = 137, "sigkill"
        else:
            exit_code, exit_reason = 1, "error"

        logger.info(
            "Starting shutdown",
            signal=signum,
            exit_code=exit_code,
            exit_reason=exit_reason,
        )

        # T074: Mark active executions as crashed
        await self._mark_active_executions_crashed()

        # Stop all heartbeats
        if isinstance(self._heartbeat_port, LifecycleService):
            # If heartbeat_port is also our LifecycleService, stop all
            await self._heartbeat_port.stop_all()

        # Send container_exited event
        await self.send_container_exited(exit_code, exit_reason)

        # Mark shutdown complete
        self._shutdown_complete.set()

        logger.info("Shutdown complete", exit_code=exit_code)

    async def _mark_active_executions_crashed(self) -> None:
        """
        Mark all active executions as crashed.

        T074: Report crash for each active execution via callback.

        For each active execution (from heartbeat service tracking),
        reports a crash event to the Control Plane.
        """
        if not isinstance(self._heartbeat_port, LifecycleService):
            logger.debug("Cannot mark executions crashed - heartbeat service not LifecycleService")
            return

        # Get active execution IDs from heartbeat service
        active_ids = list(self._heartbeat_port._tasks.keys())

        logger.info(
            "Marking active executions as crashed",
            active_count=len(active_ids),
        )

        for execution_id in active_ids:
            try:
                await self._report_execution_crash(execution_id)
            except Exception as e:
                logger.warning(
                    "Failed to report execution crash",
                    execution_id=execution_id,
                    error=str(e),
                )

    async def _report_execution_crash(self, execution_id: str) -> None:
        """
        Report a single execution as crashed.

        Args:
            execution_id: Unique execution identifier
        """
        try:
            from executor.domain.value_objects import HeartbeatSignal

            # Send crash heartbeat
            signal = HeartbeatSignal(
                timestamp=datetime.now(),
                progress={
                    "status": "crashed",
                    "reason": "executor_shutdown",
                },
            )

            await self._callback_port.report_heartbeat(execution_id, signal)

            logger.debug("Execution marked as crashed", execution_id=execution_id)

        except Exception as e:
            logger.warning(
                "Failed to report execution crash",
                execution_id=execution_id,
                error=str(e),
            )

    def get_container_id(self) -> str:
        """
        Get container ID from environment.

        T075: container_id detection.

        Checks:
        1. CONTAINER_ID environment variable
        2. HOSTNAME environment variable (Kubernetes pod name)
        3. Fallback to "unknown"

        Returns:
            Container identifier
        """
        return self._container_id

    def _get_container_id(self) -> str:
        """
        Detect container ID from environment variables.

        T075: container_id detection.

        Returns:
            Container identifier
        """
        # Try CONTAINER_ID first
        container_id = os.environ.get("CONTAINER_ID")
        if container_id:
            return container_id

        # Try HOSTNAME (set by Kubernetes)
        hostname = os.environ.get("HOSTNAME")
        if hostname:
            return hostname

        # Fallback
        logger.warning("Could not detect container_id from environment")
        return "unknown"

    def is_shutting_down(self) -> bool:
        """
        Check if shutdown is in progress.

        Returns:
            True if shutting down, False otherwise
        """
        return self._is_shutting_down

    async def wait_for_shutdown(self) -> None:
        """
        Wait for shutdown to complete.

        Can be called by signal handler to ensure graceful shutdown.
        """
        await self._shutdown_complete.wait()


def map_exit_code_to_reason(exit_code: int) -> str:
    """
    Map exit code to exit reason string.

    Args:
        exit_code: Process exit code

    Returns:
        Exit reason string
    """
    # SIGTERM = 143 (128 + 15)
    if exit_code == 143:
        return "sigterm"

    # SIGKILL = 137 (128 + 9)
    if exit_code == 137:
        return "sigkill"

    # OOM killed (usually 134 or other codes)
    if exit_code == 134:  # SIGABRT = 134 (128 + 6), often OOM
        return "oom_killed"

    # Normal exit
    if exit_code == 0:
        return "normal"

    # Error
    return "error"


# Global lifecycle service instance
_lifecycle_service: Optional[LifecycleService] = None


def get_lifecycle_service() -> Optional[LifecycleService]:
    """
    Get global lifecycle service instance.

    Returns:
        LifecycleService instance or None if not initialized
    """
    return _lifecycle_service


def register_lifecycle_service(service: LifecycleService) -> None:
    """
    Register global lifecycle service instance.

    Args:
        service: LifecycleService instance
    """
    global _lifecycle_service
    _lifecycle_service = service


def register_signal_handlers() -> None:
    """
    Register signal handlers for graceful shutdown.

    T073: Register SIGTERM handler using signal.signal().

    Handles SIGTERM (15) and SIGINT (2) for graceful shutdown.
    """
    def signal_handler(signum, frame):
        """Handle shutdown signal."""
        logger.info("Shutdown signal received", signal=signum)

        # Trigger async shutdown
        service = get_lifecycle_service()
        if service:
            asyncio.create_task(service.shutdown(signum))

    # Register SIGTERM
    signal.signal(signal.SIGTERM, signal_handler)
    logger.debug("SIGTERM handler registered")

    # Register SIGINT (for Ctrl+C in development)
    signal.signal(signal.SIGINT, signal_handler)
    logger.debug("SIGINT handler registered")
