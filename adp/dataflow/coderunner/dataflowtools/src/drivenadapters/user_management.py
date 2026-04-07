import json
import os
from typing import Dict, List

import httpx

from errors.errors import InternalErrException
from common.logger import logger
from common.configs import user_management_configs, is_auth_enable

class UserManagement:
    def __init__(self) -> None:
        self.addr = f"http://{user_management_configs['private_host']}:{user_management_configs['private_port']}"

    async def get_user_info(self, user_id: str) -> List[Dict]:
        if not is_auth_enable():
            return [
                {
                    "name": "mock-user",
                    "parent_deps": [
                        {
                            "id": "mock-department-id",
                            "name": "mock-department",
                        }
                    ],
                    "roles": ["admin"],
                }
            ]

        target = f"{self.addr}/api/user-management/v1/users/{user_id}/name,parent_deps,csf_level,roles,email,telephone,enabled,custom_attr"
        async with httpx.AsyncClient(timeout=900, verify=False) as client:
            resp = await client.request("GET", target)
            if resp.status_code < httpx.codes.OK or resp.status_code >= httpx.codes.MULTIPLE_CHOICES:
                logger.warn("[get_user_info] get user info error, status: {}, detail: {}".format(resp.status_code, resp.text))
                raise InternalErrException(detail=resp.text)
            return resp.json()

    async def batch_list_user_name(self, user_ids: List[str]) -> Dict:
        if not is_auth_enable():
            return {
                "user_names": [
                    {
                        "id": "mock-user-id",
                        "name": "mock-user",
                    }
                ],
                "department_names": [
                    {
                        "id": "mock-department-id",
                        "name": "mock-department",
                    }
                ],
                "contactor_names": [
                    {
                        "id": "mock-contactor-id",
                        "name": "mock-contactor",
                    }
                ],
                "app_names": [
                    {
                        "id": "mock-app-id",
                        "name": "mock-app",
                    }
                ],
            }

        user_ids = list(set(user_ids))
        target = f"{self.addr}/api/user-management/v1/names"
        payload = {"method":"GET", "user_ids": user_ids}

        async with httpx.AsyncClient(timeout=900, verify=False) as client:
            resp = await client.request("POST", target, json=payload)
            if resp.status_code < httpx.codes.OK or resp.status_code >= httpx.codes.MULTIPLE_CHOICES:
                logger.warn("[get_user_info] get user info error, status: {}, detail: {}".format(resp.status_code, resp.text))
                raise InternalErrException(detail=resp.text)
            return resp.json()

    async def batch_list_user_name_without_not_found(self, user_ids: List[str]) -> Dict:
        if not is_auth_enable():
            return {
                "user_names": [
                    {
                        "id": "mock-user-id",
                        "name": "mock-user",
                    }
                ],
                "department_names": [
                    {
                        "id": "mock-department-id",
                        "name": "mock-department",
                    }
                ],
                "contactor_names": [
                    {
                        "id": "mock-contactor-id",
                        "name": "mock-contactor",
                    }
                ],
                "app_names": [
                    {
                        "id": "mock-app-id",
                        "name": "mock-app",
                    }
                ],
            }

        try:
            user_names = await self.batch_list_user_name(user_ids)
        except Exception as e:
            detail_json = json.loads(e.detail)
            if detail_json.get("code") == 400019001:
               not_exist_ids = detail_json.get("detail").get("ids")
               new_user_ids = []
               for user_id in user_ids:
                   if user_id not in not_exist_ids:
                       new_user_ids.append(user_id)
               user_names = await self.batch_list_user_name(new_user_ids)
               return user_names
            raise e
        return user_names
