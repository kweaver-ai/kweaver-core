class ParamValidationErrors(object):
    """参数校验类错误码"""
    ParamMissing = "ParamMissing"
    ParamTypeError = "ParamTypeError"


class PermissionErrors(object):
    """权限类错误码"""
    Unauthorized = "Unauthorized"
    Forbidden = "Forbidden"


class BusinessLogicErrors(object):
    """业务逻辑类错误码"""
    InvalidOperation = "InvalidOperation"
    ResourceNotFound = "ResourceNotFound"
