# coding: utf-8
"""
知识网络动作类型API封装
基于 ontology-manager-action-type.json 文档实现
"""
import allure
from typing import Optional, Dict, List, Any
from src.api.base_api import BaseApi
from src.common.data_gen import random_icon, random_color


class ActionTypeApi(BaseApi):
    """知识网络动作类型相关接口封装"""

    def __init__(self, user: Optional[str] = None, password: Optional[str] = None):
        super().__init__(user=user, password=password)
        self._prefix = "/api/ontology-manager/v1"

    def list_action_types(
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
        url = f"{self._prefix}/knowledge-networks/{kn_id}/action-types"
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

    def create_action_type(
        self,
        kn_id: str,
        name: str,
        id: str,
        action_type: str,
        object_type_id: str,
        condition: Dict[str, Any],
        affect_object_type_id: Optional[str],
        parameters: List[Dict[str, Any]],
        action_source: Dict[str, Any],
        tags: List[str] = [],
        branch: str = "main",
        comment: str = "",
        concept_groups: list = [],
        exclude_fields: Optional[List[str]] = None,
    ):
        """
        创建单个动作类型的便捷方法
        """
        payload = {
            "entries": [
                {
                    "name": name,
                    "id": id,
                    "branch": branch,
                    "color": random_color(),
                    "action_type": action_type,
                    "object_type_id": object_type_id,
                    "affect": {"object_type_id": affect_object_type_id},
                    "condition": condition,
                    "parameters": parameters,
                    "action_source": action_source,
                    "comment": comment,
                    "tags": tags,
                    "concept_group": concept_groups,
                }
            ]
        }
        if exclude_fields:
            for field in exclude_fields:
                if field in payload["entries"][0]:
                    payload["entries"][0].pop(field)

        url = f"{self._prefix}/knowledge-networks/{kn_id}/action-types"
        self.headers["x-http-method-override"] = "POST"
        with allure.step(f"在知识网络 {kn_id} 中创建行动类{name}"):
            return self.post(url, json=payload)

    def delete_action_type(self, kn_id: str, action_id: str):
        url = f"{self._prefix}/knowledge-networks/{kn_id}/action-types/{action_id}"
        self.headers["x-http-method-override"] = "POST"
        with allure.step(f"在知识网络 {kn_id} 中删除行动类{action_id}"):
            return self.delete(url)

    def search_actions(self, kn_id: str, concept_group: Optional[list] = None):
        url = f"{self._prefix}/knowledge-networks/{kn_id}/action-types"

        payload = {
            "concept_groups": concept_group if concept_group else [],
            "limit": 100,
        }
        self.headers["x-http-method-override"] = "GET"

        with allure.step(f"在知识网络 {kn_id} 中检索概念分组{concept_group}下的行动类"):
            return self.post(url, json=payload, params=payload)


if __name__ == "__main__":
    """简单的使用示例"""
    # 创建API实例
    api = ActionTypeApi()
    api.search_actions("auto_test_051")
