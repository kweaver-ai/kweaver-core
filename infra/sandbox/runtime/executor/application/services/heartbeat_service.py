"""
Heartbeat Service

Manages periodic heartbeat signals during code execution.
"""

import asyncio
from datetime import datetime
from typing import Dict, Optional

import structlog

from executor.domain.ports import IHeartbeatPort, ICallbackPort
from executor.domain.value_objects import HeartbeatSignal


logger = structlog.get_logger(__name__)


class HeartbeatService(IHeartbeatPort):
    """
    Service for managing heartbeat during execution.

    Sends heartbeat signals every 5 seconds to indicate liveness.
    """

    def __init__(self, callback_port: ICallbackPort, interval: float = 5.0):
        """
        Initialize heartbeat service.

        Args:
            callback_port: Port for sending callbacks to Control Plane
            interval: Heartbeat interval in seconds (default: 5.0)
        """
        self._callback_port = callback_port
        self._interval = interval
        self._tasks: Dict[str, asyncio.Task] = {}
        self._stop_events: Dict[str, asyncio.Event] = {}
        self._lock = asyncio.Lock()

    async def start_heartbeat(self, execution_id: str) -> None:
        """
        Start heartbeat for an execution.

        Begins sending heartbeat signals every 5 seconds
        until stopped.

        Args:
            execution_id: Unique execution identifier
        """
        async with self._lock:
            if execution_id in self._tasks:
                logger.warning("Heartbeat already running", execution_id=execution_id)
                return

            stop_event = asyncio.Event()
            self._stop_events[execution_id] = stop_event

            # Create background task for heartbeat loop
            task = asyncio.create_task(
                self._heartbeat_loop(execution_id, stop_event)
            )
            self._tasks[execution_id] = task

            logger.debug("Heartbeat started", execution_id=execution_id)

    async def stop_heartbeat(self, execution_id: str, hold_lock: bool = True) -> None:
        """
        Stop heartbeat for an execution.

        Args:
            execution_id: Unique execution identifier
            hold_lock: Whether to acquire the lock (internal use)
        """
        async def _stop():
            if execution_id not in self._tasks:
                logger.debug("Heartbeat not running", execution_id=execution_id)
                return

            # Signal stop
            stop_event = self._stop_events.get(execution_id)
            if stop_event:
                stop_event.set()

            # Wait for task to complete
            task = self._tasks.pop(execution_id, None)
            if task:
                try:
                    await asyncio.wait_for(task, timeout=2.0)
                except asyncio.TimeoutError:
                    task.cancel()
                    logger.warning("Heartbeat task cancelled", execution_id=execution_id)

            # Cleanup
            self._stop_events.pop(execution_id, None)

            logger.debug("Heartbeat stopped", execution_id=execution_id)

        if hold_lock:
            async with self._lock:
                await _stop()
        else:
            await _stop()

    async def send_heartbeat(
        self,
        execution_id: str,
        signal: HeartbeatSignal,
    ) -> bool:
        """
        Send a single heartbeat signal.

        Args:
            execution_id: Unique execution identifier
            signal: Heartbeat signal to send

        Returns:
            True if successful, False otherwise
        """
        try:
            success = await self._callback_port.report_heartbeat(execution_id, signal)
            return success
        except Exception as e:
            logger.warning(
                "Failed to send heartbeat",
                execution_id=execution_id,
                error=str(e),
            )
            return False

    def is_running(self, execution_id: str) -> bool:
        """
        Check if heartbeat is running for an execution.

        Args:
            execution_id: Unique execution identifier

        Returns:
            True if heartbeat is running, False otherwise
        """
        return execution_id in self._tasks

    async def _heartbeat_loop(
        self,
        execution_id: str,
        stop_event: asyncio.Event,
    ) -> None:
        """
        Main heartbeat loop.

        Sends heartbeat every interval seconds until stop event is set.

        Args:
            execution_id: Unique execution identifier
            stop_event: Event to signal stop
        """
        try:
            while not stop_event.is_set():
                # Send heartbeat
                signal = HeartbeatSignal(
                    timestamp=datetime.now(),
                    progress={"status": "running"},
                )
                await self.send_heartbeat(execution_id, signal)

                # Wait for interval or stop event
                try:
                    await asyncio.wait_for(stop_event.wait(), timeout=self._interval)
                except asyncio.TimeoutError:
                    # Timeout is expected - continue loop
                    pass

        except asyncio.CancelledError:
            logger.debug("Heartbeat loop cancelled", execution_id=execution_id)
        except Exception as e:
            logger.error(
                "Heartbeat loop error",
                execution_id=execution_id,
                error=str(e),
            )

    async def stop_all(self) -> None:
        """
        Stop all active heartbeats.

        Called during shutdown.
        """
        async with self._lock:
            execution_ids = list(self._tasks.keys())

            for execution_id in execution_ids:
                await self.stop_heartbeat(execution_id, hold_lock=False)

            logger.info("All heartbeats stopped", count=len(execution_ids))


# Global heartbeat service instance
_heartbeat_service: Optional[HeartbeatService] = None


def get_heartbeat_service() -> Optional[HeartbeatService]:
    """
    Get global heartbeat service instance.

    Returns:
        HeartbeatService instance or None if not initialized
    """
    return _heartbeat_service


def register_heartbeat_service(service: HeartbeatService) -> None:
    """
    Register global heartbeat service instance.

    Args:
        service: HeartbeatService instance
    """
    global _heartbeat_service
    _heartbeat_service = service
