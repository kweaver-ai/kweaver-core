# coding: utf-8
"""
知识网络对象类型API封装
基于 ontology-manager-object-type.json 文档实现
"""
import allure
from typing import Optional, Dict, List, Any, Union
from src.api.base_api import BaseApi
from src.common.data_gen import random_icon, random_color


class ObjectTypeApi(BaseApi):
    """知识网络对象类型相关接口封装"""

    def __init__(self, user: Optional[str] = None, password: Optional[str] = None):
        super().__init__(user=user, password=password)
        self._prefix = "/api/ontology-manager/v1"

    def list_object_types(
        self,
        kn_id: str,
        name_pattern: str = "",
        sort: str = "update_time",
        direction: str = "desc",
        offset: int = 0,
        limit: int = 10,
        tag: str = "",
        group_id: str = "",
    ):
        """
        获取对象类列表
        """
        url = f"{self._prefix}/knowledge-networks/{kn_id}/object-types"
        params = {
            "name_pattern": name_pattern,
            "sort": sort,
            "direction": direction,
            "offset": offset,
            "limit": limit,
            "tag": tag,
            "group_id": group_id,
        }

        with allure.step(f"获取知识网络 {kn_id} 的对象类列表"):
            return self.get(url, params=params)

    def create_object_types(
        self,
        kn_id: str,
        object_id: str,
        name: str,
        data_source: Optional[str],
        data_properties: Optional[List[Dict[str, Any]]] = None,
        tags: Optional[List[str]] = None,
        comment: Optional[str] = None,
        icon: Optional[str] = None,
        color: Optional[str] = None,
        branch: Optional[str] = None,
        concept_groups: Optional[List[Dict[str, Any]]] = None,
        primary_keys: str = None,
        incremental_key: Optional[str] = None,
        display_key: str = None,
        exclude_fields: Optional[List[str]] = None,
    ):
        """
        创建对象类
        """
        url = f"{self._prefix}/knowledge-networks/{kn_id}/object-types"
        params = {"kn_id": kn_id}
        payload = {
            "entries": [
                {
                    "id": object_id,
                    "name": name,
                    "tags": tags or [],
                    "comment": comment or "",
                    "icon": icon or random_icon(),
                    "color": color or random_color(),
                    "branch": branch or "main",
                    "concept_groups": concept_groups or [],
                    "data_source": {"type": "data_view", "id": data_source},
                    "data_properties": data_properties or [],
                    "logic_properties": [],
                    "primary_keys": (
                        [primary_keys]
                        if primary_keys is not None
                        else [data_properties[0]["name"]]
                    ),
                    "incremental_key": incremental_key,
                    "display_key": (
                        display_key
                        if display_key is not None
                        else data_properties[0]["name"]
                    ),
                    "module_type": "object_type",
                }
            ]
        }

        if exclude_fields:
            for field in exclude_fields:
                if field in payload["entries"][0]:
                    payload["entries"][0].pop(field)

        self.headers["x-http-method-override"] = "POST"

        with allure.step(f"在知识网络 {kn_id} 中创建对象类"):
            return self.post(url, json=payload, params=params)

    def update_object_types(
        self,
        kn_id: str,
        object_id: str,
        name: str,
        data_source: Optional[str],
        data_properties: Optional[List[Dict[str, Any]]] = None,
        tags: Optional[List[str]] = None,
        comment: Optional[str] = None,
        icon: Optional[str] = None,
        color: Optional[str] = None,
        branch: Optional[str] = None,
        concept_groups: Optional[List[Dict[str, Any]]] = None,
        primary_keys: str = None,
        incremental_key: Optional[str] = None,
        display_key: str = None,
        exclude_fields: Optional[List[str]] = None,
    ):
        """
        更新对象类
        """
        url = f"{self._prefix}/knowledge-networks/{kn_id}/object-types/{object_id}"
        params = {"kn_id": kn_id}
        payload = {
            "id": object_id,
            "name": name,
            "tags": tags or ["自动化测试"],
            "comment": comment or "自动化测试",
            "icon": icon or random_icon(),
            "color": color or random_color(),
            "branch": branch or "main",
            "concept_groups": concept_groups or [],
            "data_source": {"type": "data_view", "id": data_source},
            "data_properties": data_properties or [],
            "logic_properties": [],
            "primary_keys": (
                [primary_keys]
                if primary_keys is not None
                else [data_properties[0]["name"]]
            ),
            "incremental_key": incremental_key,
            "display_key": (
                display_key if display_key is not None else data_properties[0]["name"]
            ),
        }

        if exclude_fields:
            for field in exclude_fields:
                if field in payload:
                    payload.pop(field)

        self.headers["x-http-method-override"] = "POST"

        with allure.step(f"在知识网络 {kn_id} 中更新对象类{name}"):
            return self.put(url, json=payload, params=params)

    def search_objects(self, kn_id: str, concept_group: Optional[list] = None):
        url = f"{self._prefix}/knowledge-networks/{kn_id}/object-types"

        payload = {
            "concept_groups": concept_group if concept_group else [],
        }
        self.headers["x-http-method-override"] = "GET"

        with allure.step(f"在知识网络 {kn_id} 中检索概念分组{concept_group}下的对象类"):
            return self.post(url, json=payload, params=payload)

    def delete_objects(self, kn_id: str, object_id: str):
        url = f"{self._prefix}/knowledge-networks/{kn_id}/object-types/{object_id}"

        with allure.step(f"在知识网络 {kn_id} 中删除对象类{object_id}"):
            return self.delete(url)


if __name__ == "__main__":
    from src.common.business_util import random_data_properties

    api = ObjectTypeApi()
    api.search_objects("auto_test_048", ["information"])
