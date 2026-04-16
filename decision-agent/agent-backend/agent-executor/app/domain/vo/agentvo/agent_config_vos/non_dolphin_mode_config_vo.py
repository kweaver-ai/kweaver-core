# -*- coding:utf-8 -*-
from pydantic import BaseModel


class NonDolphinModeConfigVo(BaseModel):
    """非 Dolphin 模式相关开关"""

    disable_history_in_a_conversation: bool = False
    disable_llm_cache: bool = False
