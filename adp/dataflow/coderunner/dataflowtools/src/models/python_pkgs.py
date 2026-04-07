from typing import Dict, List

from errors.errors import NotFoundException, InternalErrException
from models.db import get_session
from models.models import PythonPackage
from common.logger import logger
from models.models import BaseModel


class PythonPackageModel(BaseModel):

    async def __serialize_python_package_obj_to_dict(self, python_packages):
        res = []
        for python_package in python_packages:
            res.append({
                "id": python_package.f_id,
                "name": python_package.f_name,
                "oss_id": python_package.f_oss_id,
                "oss_key": python_package.f_oss_key,
                "creator_id": python_package.f_creator_id,
                "creator_name": python_package.f_creator_name,
                "created_at": python_package.f_created_at
            })
        return res

    async def get_python_package(self, conditions: List[Dict] = None) -> PythonPackage:
        sqlStr = f"select f_id, f_name, f_oss_id, f_oss_key from {self.db_name}.t_python_package where 1 = 1 "
        values = {}
        for condition in conditions:
            field = condition['field']
            value = condition['value']
            sqlStr += f"and {field} = :{field} "
            values[field] = value

        with get_session(expire_on_commit=False) as session:
            try:    
                res = session.execute(sqlStr, values).fetchone()
            except Exception as e:
                logger.warn(f"[get_python_package] failed to get python package, detail: {e}")
                raise InternalErrException(detail="db operation failed")
            
            if res is None:
                raise NotFoundException(detail=f"python package {values} not found")
            row = dict(res)
            return PythonPackage(**row)
    
    async def insert_python_package(self, data: PythonPackage):
        sqlStr = f"insert into {self.db_name}.t_python_package (f_id, f_name, f_oss_id, f_oss_key, f_creator_id, f_creator_name, f_created_at) values (:id, :name, :oss_id, :oss_key, :creator_id, :creator_name, :created_at)"
        with get_session(expire_on_commit=False) as session:
            try:
                session.execute(sqlStr, {"id": data.f_id, "name": data.f_name, "oss_id": data.f_oss_id, "oss_key": data.f_oss_key, "creator_id": data.f_creator_id, "creator_name": data.f_creator_name, "created_at": data.f_created_at})
            except Exception as e:
                logger.warn(f"[insert_python_package] failed to insert python package, detail: {e}")
                raise InternalErrException(detail="db operation failed")

    async def list_python_packages(self, offset: int, limit: int) -> List[PythonPackage]:
        sqlStr = f"select f_id, f_name, f_oss_id, f_oss_key, f_creator_id, f_creator_name, f_created_at from {self.db_name}.t_python_package order by f_created_at desc limit :offset, :limit"
        with get_session(expire_on_commit=False) as session:
            try:
                res = session.execute(sqlStr, {"offset": offset, "limit": limit}).fetchall()
                if res is None:
                    return []
                return await self.__serialize_python_package_obj_to_dict(res)
            except Exception as e:
                logger.warn(f"[list_python_packages] failed to list python packages, detail: {e}")
                raise InternalErrException(detail="db operation failed")

    async def delete_python_package(self, id: str):
        sqlStr = f"delete from {self.db_name}.t_python_package where f_id = :f_id"
        with get_session(expire_on_commit=False) as session:
            try:
                session.execute(sqlStr, {"f_id": id})
            except Exception as e:
                logger.warn(f"[delete_python_package] failed to delete python package, detail: {e}")
                raise InternalErrException(detail="db operation failed")
