from typing import Dict, Any, Optional, Type


class ErrorInfo:
    def __init__(self, code: int, message: str, cause: str):
        self.code = code
        self.message = message
        self.cause = cause


class ErrorBuilder(Exception):
    info: Optional[ErrorInfo] = None

    def __init__(self, code: Optional[int] = None, message: Optional[str] = None,
                 cause: Optional[str] = None, detail: Optional[str] = None):
        super().__init__()
        self.code = code or (self.info.code if self.info else 500000000)
        self.message = message or (self.info.message if self.info else 'Server unavailable.')
        self.cause = cause or (self.info.cause if self.info else '')
        self.detail = detail

    def __repr__(self):
        return f'{self.__class__.__name__}(code={self.code!r}, message={self.message!r}, ' \
               f'cause={self.cause!r}, detail={self.detail!r})'

    def serialize(self) -> Dict[str, Any]:
        result = {
            "code": self.code,
            "message": self.message,
            "cause": self.cause
        }
        if self.detail is not None:
            result['detail'] = self.detail
        return result


def create_exception_class(name: str, error_info: ErrorInfo) -> Type[ErrorBuilder]:
    """创建异常类"""
    return type(name, (ErrorBuilder,), {'info': error_info})

INTERNAL_ERROR = ErrorInfo(500000000, "Server unavailable", "The request is abnormal due to an internal error of the server")
BAD_REQUEST_ERROR = ErrorInfo(400000000, "Bad Request", "The passed in parameter is incorrect")
UNAUTHORIZED_ERROR = ErrorInfo(401000000, "Unauthorized", "Token is invalid")
NOT_FOUND_ERROR = ErrorInfo(404000000, "Not Found", "Resource not found")
REQUEST_PROCESSING_ERROR = ErrorInfo(202000000, "Request Accepted", "The server is processing the task")
NO_PERMISSION_ERROR = ErrorInfo(403000000, "Forbidden", "No permission to access the resource")
RESOURCE_CONFLICT_ERROR = ErrorInfo(409000000, "Conflict", "Resource already exists")

InternalErrException = create_exception_class('InternalErrException', INTERNAL_ERROR)

BadParameterException = create_exception_class('BadParameterException', BAD_REQUEST_ERROR)

UnauthorizedException = create_exception_class('UnauthorizedException', UNAUTHORIZED_ERROR)

NotFoundException = create_exception_class('NotFoundException', NOT_FOUND_ERROR)

RequestProcessingException = create_exception_class('RequestProcessingException', REQUEST_PROCESSING_ERROR)

NoPermissionException = create_exception_class('NoPermissionException', NO_PERMISSION_ERROR)

ResourceConflictException = create_exception_class('ResourceConflictException', RESOURCE_CONFLICT_ERROR)