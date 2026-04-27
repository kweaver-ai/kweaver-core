# -*- coding:utf-8 -*-
from typing import Any, Dict, List, Optional, TYPE_CHECKING

from pydantic import BaseModel

if TYPE_CHECKING:
    pass


class AgentRunOptionsVo(BaseModel):
    """Agent运行选项模型(由agent-executor定义)"""

    output_vars: Optional[List[str]] = None
    incremental_output: Optional[bool] = None
    data_source: Optional[Dict[str, Any]] = None
    llm_config: Optional[Dict[str, Any]] = None
    tmp_files: Optional[List] = None

    # new add 2025年10月19日16:52:53 --start--
    agent_id: Optional[str] = None
    conversation_id: Optional[str] = None
    agent_run_id: Optional[str] = None
    # new add 2025年10月19日16:52:53 --end--

    # new add 2025年12月17日 --start--
    is_need_progress: Optional[bool] = None
    enable_dependency_cache: Optional[bool] = None
    # new add 2025年12月17日 --end--

    # new add 2026年01月25日 --start--
    resume_info: Optional[Any] = None  # 恢复执行信息，类型为 ResumeInfo
    # new add 2026年01月25日 --end--
