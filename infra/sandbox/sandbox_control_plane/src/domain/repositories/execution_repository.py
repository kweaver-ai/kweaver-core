"""
执行仓储接口

定义执行记录持久化的抽象接口（Port）。
"""
from abc import ABC, abstractmethod
from typing import List, Optional
from datetime import datetime

from src.domain.entities.execution import Execution


class IExecutionRepository(ABC):
    """
    执行仓储接口

    这是领域层定义的 Port，由基础设施层实现 Adapter。
    """

    @abstractmethod
    async def save(self, execution: Execution) -> None:
        """保存执行记录（创建或更新）"""
        pass

    async def commit(self) -> None:
        """Explicitly commit the transaction (optional - some repos may not implement this)"""
        pass

    @abstractmethod
    async def find_by_id(self, execution_id: str) -> Optional[Execution]:
        """根据 ID 查找执行记录"""
        pass

    @abstractmethod
    async def find_by_session_id(
        self,
        session_id: str,
        limit: int = 100
    ) -> List[Execution]:
        """根据会话 ID 查找执行记录"""
        pass

    @abstractmethod
    async def find_by_status(
        self,
        status: str,
        limit: int = 100
    ) -> List[Execution]:
        """根据状态查找执行记录"""
        pass

    @abstractmethod
    async def find_crashed_executions(
        self,
        max_retry_count: int
    ) -> List[Execution]:
        """查找可重试的崩溃执行"""
        pass

    @abstractmethod
    async def find_heartbeat_timeouts(
        self,
        timeout_threshold: datetime
    ) -> List[Execution]:
        """查找心跳超时的执行"""
        pass

    @abstractmethod
    async def delete(self, execution_id: str) -> None:
        """删除执行记录"""
        pass

    @abstractmethod
    async def delete_by_session_id(self, session_id: str) -> None:
        """删除会话的所有执行记录"""
        pass

    @abstractmethod
    async def count_by_status(self, status: str) -> int:
        """统计指定状态的执行数量"""
        pass
