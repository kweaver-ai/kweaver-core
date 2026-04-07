import io
from driveradapters.base import MiddlewareHandler
from logics.pkg_module import PackageModuleService
from utils.utils import *
from utils.request import *
from errors.errors import *


class PyPkgHandler(MiddlewareHandler):

    async def put(self):
        request_body = self.request.files.get('file', [])
        if len(request_body) == 0:
            raise BadParameterException(detail= "request file body empty")
        file_info = request_body[0]
        file_body = file_info['body']
        file_size = len(file_body)
        file_name = os.path.splitext(os.path.basename(file_info['filename']))[0]

        file_name_ext = file_info['filename'].rsplit('.', 1)[-1]
        if file_name_ext != "tar":
            raise BadParameterException(detail= "file extension must be tar")
        if file_size > 1<<31:
            raise BadParameterException(detail= "file size must be less than 2GB")

        with io.BytesIO(file_body) as byte_stream:
            with io.BufferedReader(byte_stream) as buffered_reader:
                await PackageModuleService().upload_package(file_name, buffered_reader, file_size, self.user_info)
        self.set_status(201)

    async def get(self, param = None):
        if param == None:
            await self.get_package_list()
        else:
            await self.get_package(param)

    async def get_package(self, param = None):
        buff = await PackageModuleService().download_package(param)
        self.set_status(200)
        self.set_header("Content-Type", "application/octet-stream")
        self.write(buff)
    
    async def get_package_list(self):
        page = self.get_argument("page", 0)
        limit = self.get_argument("limit", 20)
        offset = int(page) * int(limit)
        pkgs = await PackageModuleService().list_packages(offset, int(limit))
        self.set_status(200)
        self.set_header("Content-Type", "application/json")
        response_content = {"pkgs": pkgs}
        self.write(response_content)

    async def delete(self, id = None):
        await PackageModuleService().delete_package(id, self.user_info)
        self.set_status(204)
