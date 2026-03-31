# -*- coding:utf-8 -*-
"""Agent 实例管理器 - 管理 agent_run_id 与 DolphinAgent 实例的映射"""

from typing import Dict, Optional, Tuple, TYPE_CHECKING
from threading import Lock
import time

if TYPE_CHECKING:
    from dolphin.sdk.agent.dolphin_agent import DolphinAgent
    from app.logic.agent_core_logic_v2.agent_core_v2 import AgentCoreV2


class AgentInstanceManager:
    """管理 agent_run_id 到 Agent 实例的映射（单例）"""

    _instance: Optional["AgentInstanceManager"] = None
    _lock: Lock = Lock()

    def __new__(cls):
        if cls._instance is None:
            with cls._lock:
                if cls._instance is None:
                    cls._instance = super().__new__(cls)
                    cls._instance._init()
        return cls._instance

    def _init(self):
        # agent_run_id -> (DolphinAgent, AgentCoreV2, timestamp)
        self._instances: Dict[str, Tuple] = {}
        self._instance_lock = Lock()
        # 实例过期时间（秒）
        self._expire_seconds = 60 * 60  # 1小时
        # 启动后台清理线程
        self._start_cleanup_thread()

    def _start_cleanup_thread(self):
        """启动后台清理线程"""
        import threading

        def cleanup_loop():
            while True:
                time.sleep(60)  # 每分钟清理一次
                try:
                    self.cleanup_expired()
                except Exception:
                    pass  # 忽略清理异常

        thread = threading.Thread(
            target=cleanup_loop, daemon=True, name="agent-instance-cleanup"
        )
        thread.start()

    def register(
        self,
        agent_run_id: str,
        agent: "DolphinAgent",
        agent_core: "AgentCoreV2",
    ) -> None:
        """注册 Agent 实例

        Args:
            agent_run_id: Agent 运行 ID
            agent: DolphinAgent 实例
            agent_core: AgentCoreV2 实例
        """
        with self._instance_lock:
            self._instances[agent_run_id] = (agent, agent_core, time.time())

    def get(self, agent_run_id: str) -> Optional[Tuple["DolphinAgent", "AgentCoreV2"]]:
        """获取 Agent 实例

        Args:
            agent_run_id: Agent 运行 ID

        Returns:
            (DolphinAgent, AgentCoreV2) 或 None
        """
        with self._instance_lock:
            data = self._instances.get(agent_run_id)
            if data is None:
                return None
            agent, agent_core, timestamp = data
            # 检查是否过期
            if time.time() - timestamp > self._expire_seconds:
                del self._instances[agent_run_id]
                return None
            return (agent, agent_core)

    def remove(self, agent_run_id: str) -> None:
        """移除 Agent 实例

        Args:
            agent_run_id: Agent 运行 ID
        """
        with self._instance_lock:
            self._instances.pop(agent_run_id, None)

    def cleanup_expired(self) -> None:
        """清理过期实例"""
        current_time = time.time()
        with self._instance_lock:
            expired_ids = [
                run_id
                for run_id, (_, _, ts) in self._instances.items()
                if current_time - ts > self._expire_seconds
            ]
            for run_id in expired_ids:
                del self._instances[run_id]


# 全局单例
agent_instance_manager = AgentInstanceManager()
