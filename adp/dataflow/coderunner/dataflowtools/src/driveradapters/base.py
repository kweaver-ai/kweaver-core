from common.response import ErrorHandler
from driveradapters.middleware import CheckToken, CheckAutomationAdmin
from tornado.web import RequestHandler
from errors.errors import InternalErrException, ErrorBuilder
from traceback import format_tb
from tornado.web import HTTPError

class BaseHandler(RequestHandler):
    '''
    此类用于处理一些公共逻辑
    '''
    def __init__(self, application, request, **kwargs):
        super().__init__(application, request, **kwargs)
        self.error_handler = ErrorHandler(self)
    
    def write_error(self, status_code, **kwargs):
        _, error, trace = kwargs.get("exc_info")
        if isinstance(error, ErrorBuilder):
            status_code = error.code / 1000000
        elif not isinstance(error, HTTPError):
            stack = format_tb(trace)
            error = InternalErrException(detail=stack)
        
        self.set_status(status_code)
        if isinstance(error, ErrorBuilder):
            self.finish(error.serialize())

class MiddlewareHandler(BaseHandler):
    def initialize(self, need_auth=True, need_automation_admin=False):
        self._middleware = []  # pylint: disable=attribute-defined-outside-init
        if need_auth:
            # 用户和权限应该统一校验
            self._reg_auth_middleware()
        if need_automation_admin:
            self._reg_automation_admin_middleware()

    def _reg_auth_middleware(self):
        instance = CheckToken(self)
        self._middleware.append(instance)

    def _reg_automation_admin_middleware(self):
        instance = CheckAutomationAdmin(self)
        self._middleware.append(instance)

    async def prepare(self):
        for middleware in self._middleware:
            if isinstance(middleware, CheckToken):
                self.user_info = await middleware.process_request()  # pylint: disable=attribute-defined-outside-init
            else:
                await middleware.process_request()