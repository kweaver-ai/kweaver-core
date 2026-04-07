"""
Performance metrics collector for sandbox code execution.

Collects wall-clock time, CPU time, memory usage, and I/O statistics
during code execution using Python's time, psutil, and resource modules.
"""

import time
import os
import gc
from typing import Optional

from executor.domain.value_objects import ExecutionMetrics
from executor.infrastructure.logging.logging_config import get_logger


logger = get_logger()


class MetricsCollector:
    """
    Context manager for collecting execution metrics.

    Tracks wall-clock time, CPU time, memory usage, and I/O statistics
    during code execution.

    Examples:
        >>> with MetricsCollector() as metrics:
        ...     execute_code()
        ... print(metrics.duration_ms)
        123.45
    """

    def __init__(self, collect_memory: bool = True, collect_io: bool = True):
        """
        Initialize metrics collector.

        Args:
            collect_memory: Whether to collect memory metrics (requires psutil)
            collect_io: Whether to collect I/O metrics (requires psutil)
        """
        self.collect_memory = collect_memory
        self.collect_io = collect_io

        # Timing metrics
        self._start_time: Optional[float] = None
        self._end_time: Optional[float] = None
        self._start_cpu: Optional[float] = None
        self._end_cpu: Optional[float] = None

        # I/O metrics
        self._start_io_read: Optional[int] = None
        self._start_io_write: Optional[int] = None
        self._end_io_read: Optional[int] = None
        self._end_io_write: Optional[int] = None

        # Process tracking
        self._process = None

        # Final metrics
        self.metrics: Optional[ExecutionMetrics] = None

    def __enter__(self) -> "MetricsCollector":
        """Start metrics collection."""
        self._start_collection()
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        """Finalize metrics collection."""
        self._end_collection()
        return False  # Don't suppress exceptions

    def _start_collection(self):
        """Record start time and resource usage."""
        try:
            # Record wall-clock time
            self._start_time = time.perf_counter()

            # Record CPU time
            self._start_cpu = time.process_time()

            # Try to get process object for advanced metrics
            try:
                import psutil

                self._process = psutil.Process()

                # Record I/O counters
                if self.collect_io:
                    try:
                        io_counters = self._process.io_counters()
                        self._start_io_read = io_counters.read_bytes
                        self._start_io_write = io_counters.write_bytes
                    except (psutil.AccessDenied, AttributeError):
                        logger.debug("I/O counters not available")
                        self.collect_io = False

                # Force garbage collection before measuring memory
                if self.collect_memory:
                    gc.collect()

            except ImportError:
                logger.debug("psutil not available, skipping advanced metrics")
                self.collect_memory = False
                self.collect_io = False

            logger.debug("Started metrics collection")

        except Exception as e:
            logger.error("Failed to start metrics collection", error=str(e))
            # Set defaults so collection doesn't fail completely
            self._start_time = time.perf_counter()
            self._start_cpu = time.process_time()

    def _end_collection(self):
        """Calculate final metrics from start/end values."""
        try:
            # Record end times
            self._end_time = time.perf_counter()
            self._end_cpu = time.process_time()

            # Calculate duration metrics (in milliseconds)
            duration_s = self._end_time - self._start_time
            cpu_s = self._end_cpu - self._start_cpu

            duration_ms = duration_s * 1000
            cpu_time_ms = cpu_s * 1000

            # Collect optional metrics
            peak_memory_mb = None
            io_read_bytes = None
            io_write_bytes = None

            if self._process is not None:
                try:
                    # Memory metrics
                    if self.collect_memory:
                        memory_info = self._process.memory_info()
                        peak_memory_mb = memory_info.rss / (1024 * 1024)  # Convert to MB

                    # I/O metrics
                    if self.collect_io and self._start_io_read is not None:
                        io_counters = self._process.io_counters()
                        io_read_bytes = io_counters.read_bytes - self._start_io_read
                        io_write_bytes = io_counters.write_bytes - self._start_io_write

                except Exception as e:
                    logger.debug("Failed to collect advanced metrics", error=str(e))

            # Create final metrics object
            self.metrics = ExecutionMetrics(
                duration_ms=duration_ms,
                cpu_time_ms=cpu_time_ms,
                peak_memory_mb=peak_memory_mb,
                io_read_bytes=io_read_bytes,
                io_write_bytes=io_write_bytes,
            )

            logger.debug(
                "Collected execution metrics",
                duration_ms=f"{duration_ms:.2f}",
                cpu_time_ms=f"{cpu_time_ms:.2f}",
                peak_memory_mb=f"{peak_memory_mb:.2f}" if peak_memory_mb else None,
            )

        except Exception as e:
            logger.error("Failed to finalize metrics", error=str(e))
            # Create minimal metrics on error
            if self._start_time and self._end_time:
                duration_ms = (self._end_time - self._start_time) * 1000
                cpu_time_ms = 0.0
                self.metrics = ExecutionMetrics(
                    duration_ms=duration_ms,
                    cpu_time_ms=cpu_time_ms,
                )
            else:
                # Fallback to current metrics
                self.metrics = ExecutionMetrics(duration_ms=0.0, cpu_time_ms=0.0)


def collect_current_memory_mb() -> Optional[float]:
    """
    Get current process memory usage in MB.

    Args:
        None

    Returns:
        Current memory usage in MB, or None if psutil is not available
    """
    try:
        import psutil

        process = psutil.Process()
        memory_info = process.memory_info()
        return memory_info.rss / (1024 * 1024)
    except Exception:
        return None


def collect_current_io() -> Optional[tuple[int, int]]:
    """
    Get current process I/O counters.

    Args:
        None

    Returns:
        Tuple of (read_bytes, write_bytes), or None if not available
    """
    try:
        import psutil

        process = psutil.Process()
        io_counters = process.io_counters()
        return (io_counters.read_bytes, io_counters.write_bytes)
    except Exception:
        return None
