# -*- coding:utf-8 -*-
"""
异常链提取
"""

from typing import List, Tuple, Optional
from types import TracebackType


def extract_exception_chain(
    exc: BaseException,
) -> List[Tuple[BaseException, Optional[TracebackType]]]:
    """
    提取异常链（包括 __cause__ 和 __context__）

    Args:
        exc: 异常对象

    Returns:
        List[Tuple[BaseException, TracebackType]]: 异常链列表
    """
    chain = []
    seen = set()
    current = exc

    while current is not None and id(current) not in seen:
        seen.add(id(current))
        chain.append((current, current.__traceback__))

        # 优先使用 __cause__（显式链接），否则使用 __context__（隐式链接）
        if current.__cause__ is not None:
            current = current.__cause__
        elif current.__context__ is not None and not current.__suppress_context__:
            current = current.__context__
        else:
            current = None

    return chain
