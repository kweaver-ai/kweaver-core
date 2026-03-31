"""Pickle序列化器 - 仅在JSON不适用时使用

用于序列化无法JSON化的对象，如：
- DolphinAgent实例
- TriditionalToolkit实例
- 包含复杂引用的对象
"""

import pickle
from typing import Any


class PickleSerializer:
    """Pickle序列化器 - 仅在JSON不适用时使用

    警告：
    - Pickle序列化的数据不可跨语言使用
    - 存在安全风险，只应用于可信数据
    - 序列化后的数据体积可能较大

    适用场景：
    - DolphinAgent等复杂SDK实例
    - TriditionalToolkit等工具包对象
    - 包含循环引用的对象
    """

    @staticmethod
    def serialize(data: Any) -> bytes:
        """序列化为bytes

        Args:
            data: 要序列化的数据，可以是任何Python对象

        Returns:
            序列化后的字节数据

        Raises:
            pickle.PicklingError: 当对象无法被pickle序列化时

        Example:
            >>> serializer = PickleSerializer()
            >>> agent = DolphinAgent(...)
            >>> bytes_data = serializer.serialize(agent)
        """
        return pickle.dumps(data, protocol=pickle.HIGHEST_PROTOCOL)

    @staticmethod
    def deserialize(data: bytes) -> Any:
        """从bytes反序列化

        Args:
            data: Pickle序列化的字节数据

        Returns:
            反序列化后的Python对象

        Raises:
            pickle.UnpicklingError: 当数据无法被反序列化时

        Example:
            >>> serializer = PickleSerializer()
            >>> bytes_data = b'\\x80\\x05...'
            >>> agent = serializer.deserialize(bytes_data)
        """
        return pickle.loads(data)
