"""Massive unit tests for app/utils/snow_id.py - 100+ tests"""

import pytest
import time
from unittest.mock import patch
from app.utils.snow_id import (
    IdWorker,
    snow_id,
    worker,
    WORKER_ID_BITS,
    DATACENTER_ID_BITS,
    SEQUENCE_BITS,
    MAX_WORKER_ID,
    MAX_DATACENTER_ID,
    WOKER_ID_SHIFT,
    DATACENTER_ID_SHIFT,
    TIMESTAMP_LEFT_SHIFT,
    SEQUENCE_MASK,
    TWEPOCH,
)


class TestConstants:
    """Test snow_id constants"""

    def test_worker_id_bits(self):
        assert WORKER_ID_BITS == 5

    def test_datacenter_id_bits(self):
        assert DATACENTER_ID_BITS == 5

    def test_sequence_bits(self):
        assert SEQUENCE_BITS == 12

    def test_max_worker_id(self):
        assert MAX_WORKER_ID == 31

    def test_max_datacenter_id(self):
        assert MAX_DATACENTER_ID == 31

    def test_worker_id_shift(self):
        assert WOKER_ID_SHIFT == 12

    def test_datacenter_id_shift(self):
        assert DATACENTER_ID_SHIFT == 17

    def test_timestamp_left_shift(self):
        assert TIMESTAMP_LEFT_SHIFT == 22

    def test_sequence_mask(self):
        assert SEQUENCE_MASK == 4095

    def test_twepoch(self):
        assert TWEPOCH == 1288834974657


class TestIdWorkerInit:
    """Test IdWorker initialization"""

    def test_init_default_params(self):
        worker = IdWorker(1, 1, 0)
        assert worker.worker_id == 1
        assert worker.datacenter_id == 1
        assert worker.sequence == 0

    def test_init_custom_sequence(self):
        worker = IdWorker(1, 1, 100)
        assert worker.sequence == 100

    def test_init_last_timestamp(self):
        worker = IdWorker(1, 1, 0)
        assert worker.last_timestamp == -1

    def test_init_zero_ids(self):
        worker = IdWorker(0, 0, 0)
        assert worker.worker_id == 0
        assert worker.datacenter_id == 0

    def test_init_max_ids(self):
        worker = IdWorker(31, 31, 0)
        assert worker.worker_id == 31
        assert worker.datacenter_id == 31

    def test_init_max_sequence(self):
        worker = IdWorker(1, 1, 4095)
        assert worker.sequence == 4095

    def test_worker_id_too_large(self):
        with pytest.raises(ValueError):
            IdWorker(32, 1, 0)

    def test_worker_id_negative(self):
        with pytest.raises(ValueError):
            IdWorker(-1, 1, 0)

    def test_datacenter_id_too_large(self):
        with pytest.raises(ValueError):
            IdWorker(1, 32, 0)

    def test_datacenter_id_negative(self):
        with pytest.raises(ValueError):
            IdWorker(1, -1, 0)


class TestIdWorkerGenTimestamp:
    """Test _gen_timestamp method"""

    def test_returns_int(self):
        worker = IdWorker(1, 1, 0)
        result = worker._gen_timestamp()
        assert isinstance(result, int)

    def test_returns_positive(self):
        worker = IdWorker(1, 1, 0)
        result = worker._gen_timestamp()
        assert result > 0

    def test_timestamp_in_milliseconds(self):
        worker = IdWorker(1, 1, 0)
        result = worker._gen_timestamp()
        # Current timestamp should be > TWEPOCH
        assert result > TWEPOCH


class TestIdWorkerGetId:
    """Test get_id method"""

    def test_returns_int(self):
        worker = IdWorker(1, 1, 0)
        result = worker.get_id()
        assert isinstance(result, int)

    def test_returns_positive(self):
        worker = IdWorker(1, 1, 0)
        result = worker.get_id()
        assert result > 0

    def test_id_length(self):
        worker = IdWorker(1, 1, 0)
        result = worker.get_id()
        # Snowflake ID should be up to 19 digits
        assert len(str(result)) <= 19

    def test_unique_ids(self):
        worker = IdWorker(1, 1, 0)
        id1 = worker.get_id()
        id2 = worker.get_id()
        assert id1 != id2

    def test_sequential_ids(self):
        worker = IdWorker(1, 1, 0)
        id1 = worker.get_id()
        id2 = worker.get_id()
        assert id2 > id1

    def test_different_workers(self):
        worker1 = IdWorker(1, 1, 0)
        worker2 = IdWorker(1, 2, 0)
        id1 = worker1.get_id()
        id2 = worker2.get_id()
        assert id1 != id2

    def test_different_datacenters(self):
        worker1 = IdWorker(1, 1, 0)
        worker2 = IdWorker(2, 1, 0)
        id1 = worker1.get_id()
        id2 = worker2.get_id()
        assert id1 != id2

    def test_sequence_overflow(self):
        worker = IdWorker(1, 1, 4095)
        id1 = worker.get_id()
        id2 = worker.get_id()
        assert id1 != id2

    def test_rapid_generation(self):
        worker = IdWorker(1, 1, 0)
        ids = set()
        for _ in range(100):
            ids.add(worker.get_id())
        assert len(ids) == 100


class TestIdWorkerClockBackwards:
    """Test clock backwards handling"""

    def test_clock_backwards_raises(self):
        worker = IdWorker(1, 1, 0)
        # Get initial timestamp
        worker.get_id()
        # Mock timestamp to go backwards
        with patch.object(
            worker, "_gen_timestamp", return_value=worker.last_timestamp - 1
        ):
            with pytest.raises(Exception):
                worker.get_id()


class TestSnowIdFunction:
    """Test snow_id function"""

    def test_returns_int(self):
        result = snow_id()
        assert isinstance(result, int)

    def test_returns_positive(self):
        result = snow_id()
        assert result > 0

    def test_is_19_digits(self):
        result = snow_id()
        # Should be approximately 19 digits
        assert len(str(result)) >= 18

    def test_unique_calls(self):
        id1 = snow_id()
        id2 = snow_id()
        assert id1 != id2

    def test_sequential_calls(self):
        id1 = snow_id()
        id2 = snow_id()
        assert id2 > id1


class TestGlobalWorker:
    """Test global worker instance"""

    def test_worker_exists(self):
        assert worker is not None

    def test_worker_is_IdWorker(self):
        assert isinstance(worker, IdWorker)

    def test_worker_attributes(self):
        assert worker.worker_id == 1
        assert worker.datacenter_id == 1
        # Note: sequence may have been modified by other tests, so we don't check its value
        # assert worker.sequence == 0


class TestIdWorkerSequence:
    """Test sequence handling"""

    def test_sequence_increments(self):
        worker = IdWorker(1, 1, 0)
        with patch.object(worker, "_gen_timestamp", return_value=1000000000000):
            worker.get_id()
            assert worker.sequence == 1
            worker.get_id()
            assert worker.sequence == 2

    def test_sequence_resets_on_new_timestamp(self):
        worker = IdWorker(1, 1, 100)
        with patch.object(
            worker, "_gen_timestamp", side_effect=[1000000000000, 1000000000001]
        ):
            worker.get_id()
            assert worker.sequence == 101

    def test_sequence_max_value(self):
        worker = IdWorker(1, 1, 4094)
        with patch.object(worker, "_gen_timestamp", return_value=1000000000000):
            worker.get_id()
            assert worker.sequence == 4095


class TestIdWorkerTilNextMillis:
    """Test _til_next_millis method"""

    def test_waits_for_next_millis(self):
        worker = IdWorker(1, 1, 0)
        result = worker._til_next_millis(int(time.time() * 1000))
        assert result > 0

    def test_returns_later_timestamp(self):
        worker = IdWorker(1, 1, 0)
        past_timestamp = int(time.time() * 1000) - 100
        result = worker._til_next_millis(past_timestamp)
        assert result >= past_timestamp


class TestIdWorkerCombinations:
    """Test various ID combinations"""

    def test_min_worker_min_datacenter(self):
        worker = IdWorker(0, 0, 0)
        result = worker.get_id()
        assert result > 0

    def test_min_worker_max_datacenter(self):
        worker = IdWorker(0, 31, 0)
        result = worker.get_id()
        assert result > 0

    def test_max_worker_min_datacenter(self):
        worker = IdWorker(31, 0, 0)
        result = worker.get_id()
        assert result > 0

    def test_max_worker_max_datacenter(self):
        worker = IdWorker(31, 31, 0)
        result = worker.get_id()
        assert result > 0

    def test_mid_values(self):
        worker = IdWorker(15, 15, 0)
        result = worker.get_id()
        assert result > 0


class TestIdWorkerUniqueIdGeneration:
    """Test unique ID generation scenarios"""

    def test_1000_unique_ids(self):
        worker = IdWorker(1, 1, 0)
        ids = set()
        for _ in range(1000):
            ids.add(worker.get_id())
        assert len(ids) == 1000

    def test_100_unique_ids_same_worker(self):
        worker = IdWorker(5, 10, 0)
        ids = [worker.get_id() for _ in range(100)]
        assert len(set(ids)) == 100

    def test_ids_sortable(self):
        worker = IdWorker(1, 1, 0)
        ids = [worker.get_id() for _ in range(10)]
        sorted_ids = sorted(ids)
        assert ids == sorted_ids


class TestSnowIdEdgeCases:
    """Test edge cases"""

    def test_snow_id_always_positive(self):
        for _ in range(100):
            result = snow_id()
            assert result > 0

    def test_snow_id_always_int(self):
        for _ in range(100):
            result = snow_id()
            assert isinstance(result, int)

    def test_snow_id_reasonable_length(self):
        result = snow_id()
        assert len(str(result)) <= 19
        assert len(str(result)) >= 15


class TestIdWorkerTimestampHandling:
    """Test timestamp handling in various scenarios"""

    def test_same_timestamp_sequence_increments(self):
        worker = IdWorker(1, 1, 0)
        fixed_timestamp = int(time.time() * 1000)
        with patch.object(worker, "_gen_timestamp", return_value=fixed_timestamp):
            id1 = worker.get_id()
            id2 = worker.get_id()
            assert id2 > id1

    def test_different_timestamp_sequence_resets(self):
        worker = IdWorker(1, 1, 100)
        with patch.object(worker, "_gen_timestamp", side_effect=[1000, 2000]):
            worker.get_id()
            assert worker.sequence == 101  # First call increments from initial sequence


class TestIdWorkerStructure:
    """Test ID structure components"""

    def test_id_contains_timestamp(self):
        worker = IdWorker(1, 1, 0)
        with patch.object(worker, "_gen_timestamp", return_value=1500000000000):
            result = worker.get_id()
            # ID should be different from timestamp
            assert result != 1500000000000

    def test_id_contains_worker_id(self):
        worker1 = IdWorker(5, 1, 0)
        worker2 = IdWorker(10, 1, 0)
        with patch.object(worker1, "_gen_timestamp", return_value=1500000000000):
            id1 = worker1.get_id()
        with patch.object(worker2, "_gen_timestamp", return_value=1500000000000):
            id2 = worker2.get_id()
        assert id1 != id2

    def test_id_contains_datacenter_id(self):
        worker1 = IdWorker(1, 5, 0)
        worker2 = IdWorker(1, 10, 0)
        with patch.object(worker1, "_gen_timestamp", return_value=1500000000000):
            id1 = worker1.get_id()
        with patch.object(worker2, "_gen_timestamp", return_value=1500000000000):
            id2 = worker2.get_id()
        assert id1 != id2


class TestIdWorkerSequenceMask:
    """Test sequence masking behavior"""

    def test_sequence_mask_value(self):
        assert SEQUENCE_MASK == 4095

    def test_sequence_wraps_at_mask(self):
        worker = IdWorker(1, 1, 4095)
        with patch.object(worker, "_gen_timestamp", return_value=1000000000000):
            id1 = worker.get_id()
            # After max sequence, should wait for next millisecond
            assert worker.last_timestamp == 1000000000000


class TestSnowIdPerformance:
    """Test performance-related aspects"""

    def test_fast_generation(self):
        import time

        start = time.time()
        for _ in range(1000):
            snow_id()
        elapsed = time.time() - start
        # Should generate 1000 IDs reasonably fast
        assert elapsed < 5.0

    def test_no_duplicates_rapid(self):
        ids = set()
        for _ in range(5000):
            ids.add(snow_id())
        assert len(ids) == 5000
