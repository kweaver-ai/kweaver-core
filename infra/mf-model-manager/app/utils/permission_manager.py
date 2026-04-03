import aiohttp
from typing import List, Dict, Optional

from app.core.config import base_config
from app.dao.small_model_dao import small_model_dao
from app.logs.stand_log import StandLogger


class PermissionManager:
    def __init__(self):
        self.base_url = f"http://{base_config.AUTHORIZATIONPRIVATEHOST}:{base_config.AUTHORIZATIONPRIVATEPORT}"
        self.auth_url = f"{self.base_url}/api/authorization/v1/policy"
        self.check_single_auth_url = f"{self.base_url}/api/authorization/v1/operation-check"
        # self.check_resource_list_url = f"{self.base_url}/api/authorization/v1/resource-list"
        self.resource_filter_url = f"{self.base_url}/api/authorization/v1/resource-filter"
        self.delete_resource_url = f"{self.base_url}/api/authorization/v1/policy-delete"
        self.session: Optional[aiohttp.ClientSession] = None

    async def get_session(self) -> aiohttp.ClientSession:
        if self.session is None or self.session.closed:
            self.session = aiohttp.ClientSession(connector=aiohttp.TCPConnector(ssl=False))
        return self.session

    async def add_permission(self, user_id: str, resource_id: str, resource_name: str, resource_type: str,
                             user_name: str, role: str) -> bool:
        # admin用户无需授权
        if user_id == "266c6a42-6131-4d62-8f39-853e7093701c":
            return True
        """添加权限"""
        payload = [{
            "accessor": {
                "id": user_id,
                "type": role,
                "name": user_name
            },
            "resource": {
                "id": resource_id,
                "type": resource_type,
                "name": resource_name
            },
            "operation": {
                "allow": [
                    {"id": "display"},
                    {"id": "modify"},
                    {"id": "delete"},
                    {"id": "execute"}
                ],
                "deny": []
            },
            "condition": "{}",
            "expires_at": "1970-01-01T08:00:00+08:00"
        }]
        # 使用filter接口过滤权限不再需要手动添加
        # admin_user_id = "266c6a42-6131-4d62-8f39-853e7093701c"
        # if user_id != admin_user_id:
        #     payload.append({
        #         "accessor": {
        #             "id": admin_user_id,
        #             "type": "user",
        #             "name": "admin"
        #         },
        #         "resource": {
        #             "id": resource_id,
        #             "type": resource_type,
        #             "name": resource_name
        #         },
        #         "operation": {
        #             "allow": [
        #                 {"id": "display"},
        #                 {"id": "modify"},
        #                 {"id": "delete"},
        #                 {"id": "execute"},
        #                 {"id": "authorize"}
        #             ],
        #             "deny": []
        #         },
        #         "condition": "{}",
        #         "expires_at": "1970-01-01T08:00:00+08:00"
        #     })
        try:
            session = await self.get_session()
            async with session.post(
                    self.auth_url,
                    json=payload,
                    headers={'Content-Type': 'application/json'}
            ) as response:
                if response.status == 204:
                    return True
        except Exception as e:
            StandLogger.error(e.args)
        return False

    async def close(self):
        if self.session and not self.session.closed:
            await self.session.close()

    async def check_single_permission(self, user_id: str, resource_id: str, operations: str,
                                      resource_type: str, role: str) -> bool:
        """校验用户对资源的权限"""
        payload = {
            "method": "GET",
            "accessor": {
                "id": user_id,
                "type": role
            },
            "resource": {
                "id": resource_id,
                "type": resource_type
            },
            "operation": [operations]
        }

        try:
            session = await self.get_session()
            async with session.post(
                    self.check_single_auth_url,
                    json=payload,
                    headers={'Content-Type': 'application/json'}
            ) as response:
                if response.status == 200:
                    result = await response.json()
                    return result.get('result', False)
                return False
        except Exception as e:
            StandLogger.error(e.args)
            return False

    async def get_permission_ids(self, user_id: str, operation: str,
                                 resource_type: str, resource_name: str, role: str) -> list:
        """获取资源列表"""
        payload = {
            "method": "GET",
            "accessor": {
                "id": user_id,
                "type": role
            },
            "resource":
                {
                    "type": resource_type,
                    "name": resource_name
                }
            ,
            "operation": [
                operation
            ]
        }
        model_ids = small_model_dao.get_all_ids()
        resources = []
        for model_id in model_ids:
            resources.append({
                "id": model_id['f_model_id'],
                "type": "small_model",
                "name": "小模型"
            })
        payload = {
            "method": "GET",
            "accessor": {
                "id": user_id,
                "type": "user"
            },
            "resources": resources,
            "operation": [
                operation
            ]
        }
        operation_ids = []
        try:
            session = await self.get_session()
            async with session.post(
                    self.resource_filter_url,
                    json=payload,
                    headers={'Content-Type': 'application/json'}
            ) as response:
                if response.status == 200:
                    result = await response.json()
                    operation_ids = [item['id'] for item in result if operation != "*"]
                    return operation_ids
        except Exception as e:
            StandLogger.error(e.args)
        return operation_ids

    async def delete_permission(self, resource_type: str, resource_ids: list) -> bool:
        """删除权限"""
        session = await self.get_session()
        resources = [{"id": resource_id, "type": resource_type} for resource_id in resource_ids]
        payload = {"resources": resources,
                   "method": "DELETE"}
        try:
            async with session.post(
                    self.delete_resource_url,
                    json=payload,
                    headers={"Content-Type": "application/json"}
            ) as response:
                if response.status == 204:
                    return True
        except Exception as e:
            StandLogger.error(e.args)
        return False


permission_manager = PermissionManager()
