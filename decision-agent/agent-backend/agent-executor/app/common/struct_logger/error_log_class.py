import sys
import time
import traceback
import json
from typing import Optional, Dict, Any


class ErrorLog:
    """错误日志类，用于生成结构化的错误日志信息"""

    def __init__(
        self,
        message: str,
        caller_frame: Optional[Any] = None,
        caller_traceback: Optional[str] = None,
    ):
        """
        初始化错误日志

        @param message: 实际内容(字符串类型)
        @param caller_frame: 调用者上下文（请使用sys._getframe()）
        @param caller_traceback: 调用者当前堆栈信息（请使用traceback.format_exc()，调用位置不在except Exception：下，请不要传参）
        """
        self.message = message
        self.caller_frame = caller_frame
        self.caller_traceback = caller_traceback
        self._log_info = self._generate_log_info()

    def _generate_log_info(self) -> Dict[str, Any]:
        """
        生成日志信息字典
        """
        log_info = {}
        log_info["message"] = self.message

        if self.caller_frame:
            log_info["caller"] = (
                self.caller_frame.f_code.co_filename
                + ":"
                + str(self.caller_frame.f_lineno)
            )
        else:
            # 如果没有提供caller_frame，使用当前调用栈
            frame = sys._getframe(2)  # 跳过当前方法和__init__方法
            log_info["caller"] = frame.f_code.co_filename + ":" + str(frame.f_lineno)

        log_info["stack"] = self.caller_traceback if self.caller_traceback else ""
        log_info["time"] = time.strftime(
            "%Y-%m-%d %H:%M:%S", time.localtime(time.time())
        )

        return log_info

    def to_dict(self) -> Dict[str, Any]:
        """
        返回日志信息的字典格式
        """
        return self._log_info.copy()

    def to_json(self) -> str:
        return json.dumps(self._log_info, ensure_ascii=False)

    def get_message(self) -> str:
        """
        获取错误消息
        """
        return self._log_info["message"]

    def get_caller(self) -> str:
        """
        获取调用者信息
        """
        return self._log_info["caller"]

    def get_stack(self) -> str:
        """
        获取堆栈信息
        """
        return self._log_info["stack"]

    def get_time(self) -> str:
        """
        获取时间戳
        """
        return self._log_info["time"]

    def __str__(self) -> str:
        """
        返回日志信息的字符串表示
        """
        return f"ErrorLog(message={self.message}, caller={self.get_caller()}, time={self.get_time()})"

    def __repr__(self) -> str:
        """
        返回日志信息的详细字符串表示
        """
        return self.__str__()


def create_error_log(message: str, include_traceback: bool = False) -> ErrorLog:
    """
    便捷函数：创建错误日志对象

    @param message: 错误消息
    @param include_traceback: 是否包含堆栈跟踪信息
    @return: ErrorLog对象
    """
    caller_frame = sys._getframe(1)  # 获取调用者的frame
    caller_traceback = traceback.format_exc() if include_traceback else ""

    return ErrorLog(message, caller_frame, caller_traceback)


def get_error_log_dict(
    message: str, caller_frame, caller_traceback: str = ""
) -> Dict[str, Any]:
    """
    兼容性函数：保持与原有get_error_log函数的兼容性

    @param message: 实际内容(字符串类型)
    @param caller_frame: 调用者上下文（请使用sys._getframe()）
    @param caller_traceback: 调用者当前堆栈信息（请使用traceback.format_exc()，调用位置不在except Exception：下，请不要传参）
    @return: 日志信息字典
    """
    error_log = ErrorLog(message, caller_frame, caller_traceback)
    return error_log.to_dict()


def get_error_log_json(message: str, caller_frame, caller_traceback: str = "") -> str:
    """
    兼容性函数：保持与原有get_error_log函数的兼容性

    @param message: 实际内容(字符串类型)
    @param caller_frame: 调用者上下文（请使用sys._getframe()）
    @param caller_traceback: 调用者当前堆栈信息（请使用traceback.format_exc()，调用位置不在except Exception：下，请不要传参）
    @return: 日志信息json
    """
    error_log = ErrorLog(message, caller_frame, caller_traceback)
    return error_log.to_json()
