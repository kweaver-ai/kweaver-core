# -*- coding:utf-8 -*-
from pydantic import BaseModel


class ReactConfigVo(BaseModel):
    """ReAct 模式相关开关"""

    disable_history_in_a_conversation: bool = False
    disable_llm_cache: bool = False
