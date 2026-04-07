from tornado.web import RequestHandler
from common.response import ErrorHandler
from logics.check_module import CheckModule
from utils.utils import *
from utils.request import *
from errors.errors import *

check_model = CheckModule()

class PyPkgHandler(RequestHandler):
    async def delete(self, name = None):
        error_handler = ErrorHandler(self)
        try:
            if name is None:
                raise BadParameterException(detail= "name is required")
            await check_model.uninstall_module(name)
            self.set_status(204)
        except Exception as e:
            error_handler.INTERNAL_ERROR(e)
