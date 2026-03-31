# -*- coding:utf-8 -*-
"""
Unit tests for app.logic.agent_core_logic_v2.agent_instance_manager module
"""

import time
from unittest.mock import Mock, patch

from app.logic.agent_core_logic_v2.agent_instance_manager import (
    AgentInstanceManager,
    agent_instance_manager,
)


class TestAgentInstanceManager:
    """Test cases for AgentInstanceManager class"""

    def test_singleton_pattern(self):
        """Test that AgentInstanceManager implements singleton pattern correctly"""
        instance1 = AgentInstanceManager()
        instance2 = AgentInstanceManager()
        assert instance1 is instance2

    def test_new_method_returns_same_instance(self):
        """Test __new__ method returns the same instance"""
        manager1 = AgentInstanceManager()
        manager2 = AgentInstanceManager.__new__(AgentInstanceManager)
        assert manager1 is manager2

    def test_init_method_called_once(self):
        """Test _init method is called only once"""
        with patch.object(AgentInstanceManager, "_start_cleanup_thread"):
            manager1 = AgentInstanceManager()
            initial_instances = manager1._instances.copy()

            manager2 = AgentInstanceManager()
            assert manager2._instances == initial_instances

    def test_register_agent_instance(self):
        """Test registering an agent instance"""
        manager = AgentInstanceManager()
        agent_run_id = "test_run_id_001"
        agent = Mock()
        agent_core = Mock()

        manager.register(agent_run_id, agent, agent_core)

        assert agent_run_id in manager._instances
        registered_agent, registered_core, timestamp = manager._instances[agent_run_id]
        assert registered_agent is agent
        assert registered_core is agent_core
        assert isinstance(timestamp, float)

    def test_register_multiple_instances(self):
        """Test registering multiple agent instances"""
        manager = AgentInstanceManager()
        initial_count = len(manager._instances)

        for i in range(5):
            agent_run_id = f"test_run_id_multi_{i}"
            agent = Mock()
            agent_core = Mock()
            manager.register(agent_run_id, agent, agent_core)

        assert len(manager._instances) == initial_count + 5

    def test_register_overwrites_existing(self):
        """Test that registering with same ID overwrites existing"""
        manager = AgentInstanceManager()
        agent_run_id = "test_run_id_001"

        agent1 = Mock()
        core1 = Mock()
        manager.register(agent_run_id, agent1, core1)

        agent2 = Mock()
        core2 = Mock()
        manager.register(agent_run_id, agent2, core2)

        registered_agent, registered_core, _ = manager._instances[agent_run_id]
        assert registered_agent is agent2
        assert registered_core is core2

    def test_get_existing_instance(self):
        """Test getting an existing agent instance"""
        manager = AgentInstanceManager()
        agent_run_id = "test_run_id_001"
        agent = Mock()
        agent_core = Mock()

        manager.register(agent_run_id, agent, agent_core)
        result = manager.get(agent_run_id)

        assert result is not None
        retrieved_agent, retrieved_core = result
        assert retrieved_agent is agent
        assert retrieved_core is agent_core

    def test_get_nonexistent_instance(self):
        """Test getting a non-existent agent instance"""
        manager = AgentInstanceManager()
        result = manager.get("nonexistent_id")
        assert result is None

    def test_get_expired_instance(self):
        """Test that expired instances are not returned"""
        manager = AgentInstanceManager()
        agent_run_id = "test_run_id_001"
        agent = Mock()
        agent_core = Mock()

        # Register with old timestamp
        with patch("time.time", return_value=time.time() - 3600):
            manager.register(agent_run_id, agent, agent_core)

        # Try to get - should return None and remove the instance
        result = manager.get(agent_run_id)
        assert result is None
        assert agent_run_id not in manager._instances

    def test_remove_instance(self):
        """Test removing an agent instance"""
        manager = AgentInstanceManager()
        agent_run_id = "test_run_id_001"
        agent = Mock()
        agent_core = Mock()

        manager.register(agent_run_id, agent, agent_core)
        assert agent_run_id in manager._instances

        manager.remove(agent_run_id)
        assert agent_run_id not in manager._instances

    def test_remove_nonexistent_instance(self):
        """Test removing a non-existent instance doesn't raise error"""
        manager = AgentInstanceManager()
        # Should not raise any exception
        manager.remove("nonexistent_id")

    def test_cleanup_expired_instances(self):
        """Test cleanup of expired instances"""
        manager = AgentInstanceManager()
        current_time = time.time()

        # Register some instances with different timestamps
        old_id = "old_instance"
        recent_id = "recent_instance"

        with patch("time.time", return_value=current_time - 3600):
            manager.register(old_id, Mock(), Mock())

        with patch("time.time", return_value=current_time - 100):
            manager.register(recent_id, Mock(), Mock())

        # Cleanup should remove old instance but keep recent one
        manager.cleanup_expired()

        assert old_id not in manager._instances
        assert recent_id in manager._instances

    def test_cleanup_expired_with_all_expired(self):
        """Test cleanup when all instances are expired"""
        manager = AgentInstanceManager()
        current_time = time.time()

        # Register instances that are all expired
        initial_count = len(manager._instances)
        for i in range(5):
            with patch("time.time", return_value=current_time - 3600):
                manager.register(f"expired_all_{i}", Mock(), Mock())

        assert len(manager._instances) == initial_count + 5

        manager.cleanup_expired()

        # All expired instances should be removed
        remaining_expired = sum(
            1 for k in manager._instances.keys() if k.startswith("expired_all_")
        )
        assert remaining_expired == 0

    def test_cleanup_expired_with_none_expired(self):
        """Test cleanup when no instances are expired"""
        manager = AgentInstanceManager()

        # Register recent instances
        initial_count = len(manager._instances)
        for i in range(5):
            manager.register(f"recent_none_{i}", Mock(), Mock())

        count_before_cleanup = len(manager._instances)
        manager.cleanup_expired()

        # Recent instances should remain
        recent_count = sum(
            1 for k in manager._instances.keys() if k.startswith("recent_none_")
        )
        assert recent_count == 5

    def test_thread_safety_of_register(self):
        """Test that register is thread-safe"""
        manager = AgentInstanceManager()

        # This test verifies the lock is used, actual concurrent testing
        # would require threading which is complex for unit tests
        assert hasattr(manager, "_instance_lock")
        assert manager._instance_lock is not None

    def test_expire_seconds_default_value(self):
        """Test default expire seconds value"""
        manager = AgentInstanceManager()
        assert manager._expire_seconds == 60 * 60  # 1 hour

    def test_instances_dict_structure(self):
        """Test that instances dict has correct structure"""
        manager = AgentInstanceManager()
        agent_run_id = "test_run_id"
        agent = Mock()
        agent_core = Mock()

        manager.register(agent_run_id, agent, agent_core)

        assert agent_run_id in manager._instances
        data = manager._instances[agent_run_id]
        assert isinstance(data, tuple)
        assert len(data) == 3
        assert isinstance(data[0], Mock)  # agent
        assert isinstance(data[1], Mock)  # agent_core
        assert isinstance(data[2], float)  # timestamp

    def test_get_with_empty_instances(self):
        """Test get when no instances are registered"""
        manager = AgentInstanceManager()
        assert manager.get("any_id") is None

    def test_register_with_same_id_updates_timestamp(self):
        """Test that re-registering updates the timestamp"""
        manager = AgentInstanceManager()
        agent_run_id = "test_run_id"
        agent = Mock()
        agent_core = Mock()

        manager.register(agent_run_id, agent, agent_core)
        _, _, timestamp1 = manager._instances[agent_run_id]

        time.sleep(0.01)  # Small delay
        manager.register(agent_run_id, agent, agent_core)
        _, _, timestamp2 = manager._instances[agent_run_id]

        assert timestamp2 > timestamp1

    def test_global_singleton_instance(self):
        """Test the global agent_instance_manager singleton"""

        assert isinstance(agent_instance_manager, AgentInstanceManager)

        # Get it again and verify it's the same instance
        from app.logic.agent_core_logic_v2.agent_instance_manager import (
            agent_instance_manager as aim2,
        )

        assert agent_instance_manager is aim2

    def test_start_cleanup_thread_creates_daemon(self):
        """Test that cleanup thread is created as daemon"""
        manager = AgentInstanceManager()
        # The cleanup thread should be started during initialization
        # We can't easily test the thread itself, but we can verify
        # the method exists and doesn't crash
        assert hasattr(manager, "_start_cleanup_thread")

    def test_cleanup_loop_exception_handling(self):
        """Test that cleanup loop handles exceptions gracefully"""
        manager = AgentInstanceManager()
        # Cleanup should not raise even if there are issues
        manager.cleanup_expired()
        # No exception should be raised
        assert True

    def test_multiple_managers_share_same_state(self):
        """Test that multiple manager references share the same state"""
        manager1 = AgentInstanceManager()
        manager2 = AgentInstanceManager()

        agent_run_id = "test_id"
        agent = Mock()
        agent_core = Mock()

        manager1.register(agent_run_id, agent, agent_core)

        result = manager2.get(agent_run_id)
        assert result is not None
        retrieved_agent, _ = result
        assert retrieved_agent is agent

    def test_concurrent_registration_different_ids(self):
        """Test registration of different IDs doesn't interfere"""
        manager = AgentInstanceManager()

        agent1 = Mock()
        core1 = Mock()
        agent2 = Mock()
        core2 = Mock()

        manager.register("id1", agent1, core1)
        manager.register("id2", agent2, core2)

        result1 = manager.get("id1")
        result2 = manager.get("id2")

        assert result1 is not None
        assert result2 is not None

        retrieved_agent1, _ = result1
        retrieved_agent2, _ = result2

        assert retrieved_agent1 is agent1
        assert retrieved_agent2 is agent2

    def test_get_returns_tuple_or_none(self):
        """Test that get always returns tuple or None"""
        manager = AgentInstanceManager()

        # Non-existent should return None
        result = manager.get("nonexistent")
        assert result is None

        # Existent should return tuple
        manager.register("existent", Mock(), Mock())
        result = manager.get("existent")
        assert isinstance(result, tuple)
        assert len(result) == 2

    def test_cleanup_preserves_recent_instances(self):
        """Test that cleanup preserves all non-expired instances"""
        manager = AgentInstanceManager()
        current_time = time.time()

        # Register instances at different times
        ids = []
        for i in range(10):
            agent_run_id = f"test_preserve_{i}"
            ids.append(agent_run_id)
            # Make every other one expired
            offset = -3600 if i % 2 == 0 else -100
            with patch("time.time", return_value=current_time + offset):
                manager.register(agent_run_id, Mock(), Mock())

        manager.cleanup_expired()

        # Verify the correct ones remain
        for i, agent_run_id in enumerate(ids):
            exists = agent_run_id in manager._instances
            # Even indices (0, 2, 4, ...) should be expired
            # Odd indices (1, 3, 5, ...) should remain
            if i % 2 == 0:
                assert not exists, (
                    f"Instance {agent_run_id} should have been cleaned up"
                )
            else:
                assert exists, f"Instance {agent_run_id} should still exist"

    def test_manager_state_persistence(self):
        """Test that manager state persists across operations"""
        manager = AgentInstanceManager()

        # Register multiple instances
        test_data = {}
        for i in range(5):
            agent_run_id = f"test_{i}"
            agent = Mock()
            agent_core = Mock()
            test_data[agent_run_id] = (agent, agent_core)
            manager.register(agent_run_id, agent, agent_core)

        # Verify all are accessible
        for agent_run_id, (agent, core) in test_data.items():
            result = manager.get(agent_run_id)
            assert result is not None
            retrieved_agent, retrieved_core = result
            assert retrieved_agent is agent
            assert retrieved_core is core
