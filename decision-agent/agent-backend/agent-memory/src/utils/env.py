import os
from typing import Optional


def getenv_int(
    name: str, default: Optional[int] = None, raise_error: bool = False
) -> Optional[int]:
    """
    获取环境变量并转换为整数

    Args:
        name: 环境变量名
        default: 当环境变量不存在或转换失败时返回的默认值 (默认为None)
        raise_error: 是否在转换失败时抛出异常 (默认为False)

    Returns:
        转换后的整数值，如果失败则返回default或None

    Raises:
        ValueError: 当raise_error=True且转换失败时
    """
    value = os.getenv(name)
    if value is None:
        if raise_error:
            raise ValueError(f"Environment variable '{name}' not set")
        return default

    try:
        return int(value)
    except ValueError:
        if raise_error:
            raise ValueError(
                f"Environment variable '{name}' is not a valid integer: '{value}'"
            )
        return default
