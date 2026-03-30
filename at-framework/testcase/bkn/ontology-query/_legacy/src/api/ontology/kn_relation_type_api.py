# coding: utf-8
"""
知识网络关系类型API封装
基于 ontology-manager-relation-type.json 文档实现
"""
import allure
from typing import Optional, Dict, List, Any
from src.api.base_api import BaseApi


class RelationTypeApi(BaseApi):
    """知识网络关系类型相关接口封装"""

    def __init__(self, user: Optional[str] = None, password: Optional[str] = None):
        super().__init__(user=user, password=password)
        self._prefix = "/api/ontology-manager/v1"

    def query_relation_paths(
        self,
        kn_id: str,
        source_object_type_id: str,
        target_object_type_id: str,
        direction: str,
        path_max_length: int = 1,
    ):
        """
        查询两个类之间的关联路径
        """
        url = f"{self._prefix}/knowledge-networks/{kn_id}/relation_paths"

        payload = {
            "concept_groups": "",
            "source_object_type": source_object_type_id,
            "target_object_type": target_object_type_id,
            "path_max_length": path_max_length,
            "direction": direction,
        }

        with allure.step(
            f"查询知识网络 {kn_id} 中从 {source_object_type_id} 到 {target_object_type_id} 的关联路径"
        ):
            return self.post(url, json=payload)

    def list_relation_types(
        self,
        kn_id: str,
        name_pattern: Optional[str] = None,
        sort: str = "update_time",
        direction: str = "desc",
        offset: int = 0,
        limit: int = 10,
        tag: Optional[str] = None,
        group_id: Optional[str] = None,
        source_object_type_id: Optional[str] = None,
        target_object_type_id: Optional[str] = None,
    ):
        """
        获取关系类列表
        """
        url = f"{self._prefix}/knowledge-networks/{kn_id}/relation-types"

        params = {
            "sort": sort,
            "direction": direction,
            "offset": offset,
            "limit": limit,
        }

        # 添加可选参数
        if name_pattern:
            params["name_pattern"] = name_pattern
        if tag:
            params["tag"] = tag
        if group_id:
            params["group_id"] = group_id
        if source_object_type_id:
            params["source_object_type_id"] = source_object_type_id
        if target_object_type_id:
            params["target_object_type_id"] = target_object_type_id

        with allure.step(f"获取知识网络 {kn_id} 的关系类型列表"):
            return self.get(url, params=params)

    def create_relation_type(
        self,
        kn_id: str,
        name: str,
        id: str,
        source_object_type_id: str,
        target_object_type_id: str,
        mapping_rules: Dict[str, Any],
        type: Optional[str] = "direct",
        tags: Optional[List[str]] = None,
        comment: Optional[str] = None,
        module_type: Optional[str] = None,
        branch: Optional[str] = None,
        exclude_fields: Optional[List[str]] = None,
    ):
        """
        创建关系类
        """
        url = f"{self._prefix}/knowledge-networks/{kn_id}/relation-types"
        payload = {
            "entries": [
                {
                    "name": name,
                    "id": id,
                    "type": type,
                    "tags": tags or [],
                    "source_object_type_id": source_object_type_id,
                    "target_object_type_id": target_object_type_id,
                    "mapping_rules": mapping_rules,
                    "comment": comment or "",
                    "module_type": module_type if module_type else "relation_type",
                    "branch": branch or "main",
                }
            ]
        }
        if exclude_fields:
            for field in exclude_fields:
                if field in payload["entries"][0]:
                    payload["entries"][0].pop(field)

        with allure.step(f"在知识网络 {kn_id} 中创建 关系类型{name}"):
            return self.post(url, json=payload)

    def get_relation_type_details(
        self, kn_id: str, rt_ids: str, include_detail: bool = False
    ):
        """
        获取关系类详情
        """
        url = f"{self._prefix}/knowledge-networks/{kn_id}/relation-types/{rt_ids}"

        params = {}
        if include_detail:
            params["include_detail"] = include_detail

        with allure.step(f"获取知识网络 {kn_id} 中关系类型 {rt_ids} 的详情"):
            return self.get(url, params=params if params else None)

    def update_relation_type(
        self,
        kn_id: str,
        name: str,
        id: str,
        source_object_type_id: str,
        target_object_type_id: str,
        mapping_rules: Dict[str, Any],
        type: Optional[str] = "direct",
        tags: Optional[List[str]] = None,
        comment: Optional[str] = None,
        module_type: Optional[str] = None,
        branch: Optional[str] = None,
        exclude_fields: Optional[List[str]] = None,
    ):
        """
        修改关系类
        """
        url = f"{self._prefix}/knowledge-networks/{kn_id}/relation-types/{id}"
        payload = {
            "name": name,
            "type": type,
            "tags": tags or [],
            "source_object_type_id": source_object_type_id,
            "target_object_type_id": target_object_type_id,
            "mapping_rules": mapping_rules,
            "comment": comment or "",
            "module_type": module_type if module_type else "relation_type",
            "branch": branch or "main",
        }

        if exclude_fields:
            for field in exclude_fields:
                if field in payload:
                    payload.pop(field)

        with allure.step(f"在知识网络 {kn_id} 中修改关系类型{name}"):
            return self.put(url, json=payload)

    def delete_relation_types(self, kn_id: str, rt_ids: str):
        """
        删除关系类
        """
        url = f"{self._prefix}/knowledge-networks/{kn_id}/relation-types/{rt_ids}"

        with allure.step(f"删除知识网络 {kn_id} 中的关系类型 {rt_ids}"):
            return self.delete(url)

    def search_relations(self, kn_id: str, concept_group: Optional[list] = None):
        url = f"{self._prefix}/knowledge-networks/{kn_id}/relation-types"

        payload = {
            "concept_groups": concept_group if concept_group else [],
            "limit": 100,
        }
        self.headers["x-http-method-override"] = "GET"

        with allure.step(f"在知识网络 {kn_id} 中检索概念分组{concept_group}下的关系类"):
            return self.post(url, json=payload, params=payload)


if __name__ == "__main__":
    """简单的使用示例"""
    # 创建API实例
    api = RelationTypeApi()

    # 示例：查询关联路径
    # response = api.query_relation_paths("test_kn_for_object_type_dlYx0QE3", "person", "organisation","forward")

    resp = api.list_relation_types(
        kn_id="auto_test_048",
    )
