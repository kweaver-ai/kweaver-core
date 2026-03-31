import time
from typing import List
from common.logger import logger
from drivenadapters.log import BuildAuditLogParams, Log
from drivenadapters.code_runner import CodeRunner
from drivenadapters.oss_management import OSSManagement
from drivenadapters.user_management import UserManagement
from errors.errors import *
from models.models import PythonPackage
from models.driven_models import UserInfo
from models.python_pkgs import PythonPackageModel
from utils.utils import generate_random_id
from common.locale import Locale


class PackageModuleService:
    def __init__(self):
        self.oss_manager = OSSManagement()
        self.user_management = UserManagement()
        self.python_package_manager = PythonPackageModel()
        self.code_runner = CodeRunner()

    async def upload_package(self, name: str, file_body: bytes, file_size: int, user_info: UserInfo):
        name = name.lower().strip()
        try:
            info = await self.python_package_manager.get_python_package([{"field": "f_name", "value": name}])
            if info:
                raise ResourceConflictException(detail=f"package {name} already exists")
        except NotFoundException:
            pass
        
        oss_id = await self.oss_manager.get_availd_oss()
        random_id = generate_random_id()

        await self.oss_manager.upload_file(oss_id, random_id, True, file_body, file_size)
        id = generate_random_id()
        await self.python_package_manager.insert_python_package(PythonPackage(f_id=id, 
                                                                              f_name=name, f_oss_id=oss_id, 
                                                                              f_oss_key=random_id, 
                                                                              f_creator_id=user_info.user_id, 
                                                                              f_creator_name=user_info.user_name, 
                                                                              f_created_at=int(time.time())))
        try:
            detail, ext_msg = Locale.get_console_log("upload_pkg", (name), None)
            data = BuildAuditLogParams(user_info=user_info, msg=detail, ext_msg=ext_msg, out_biz_id=id, log_level=Log.NcTLogLevel_NCT_LL_INFO)
            await Log.log(data)
        except Exception as e:
            logger.warn(f"[upload_package] failed to send console log, detail: {e}")

    async def download_package(self, name: str) -> bytes:
        info = await self.python_package_manager.get_python_package([{"field": "f_name", "value": name}])
        
        try:
            buff = await self.oss_manager.download_file(info.f_oss_id, info.f_oss_key, True)
        except NotFoundException:
            raise NotFoundException(detail=f"download file {info.f_name} not found")
        return bytes(buff)

    async def list_packages(self, offset: int, limit: int) -> List[PythonPackage]:
        pkgs = await self.python_package_manager.list_python_packages(offset, limit)
        user_ids = []
        for pkg in pkgs:
            user_ids.append(pkg.get("creator_id"))
        
        names = await self.user_management.batch_list_user_name_without_not_found(user_ids)
        user_names = names.get("user_names")
        user_name_map = {}
        for name in user_names:
            user_name_map[name.get("id")] = name.get("name")
        for pkg in pkgs:
            if pkg["creator_id"] not in user_name_map:
                continue
            
            pkg["creator_name"] = user_name_map.get(pkg["creator_id"])
        
        return pkgs
    
    async def delete_package(self, id: str, user_info: UserInfo):
        try:
            info = await self.python_package_manager.get_python_package([{"field": "f_id", "value": id}])
        except NotFoundException:
            return 
        # 先删除安装包,再删除数据库记录
        await self.code_runner.delete_pkg(info.f_name)
        try:
            await self.python_package_manager.delete_python_package(id)
        except Exception as e:
            logger.warn(f"[delete_package] failed to delete package, detail: {e}")
            raise InternalErrException(detail= "db operation failed")
       
        # TODO: 删除oss中上传的源文件
        # await self.oss_manager.delete_file(id)
        
        try:
            detail, ext_msg = Locale.get_console_log("delete_pkg", (info.f_name), None)
            data = BuildAuditLogParams(user_info=user_info, msg=detail, ext_msg=ext_msg, out_biz_id=info.f_id, log_level=Log.NcTLogLevel_NCT_LL_WARN)
            await Log.log(data)
        except Exception as e:
            logger.warn(f"[delete_package] failed to send console log, detail: {e}")


