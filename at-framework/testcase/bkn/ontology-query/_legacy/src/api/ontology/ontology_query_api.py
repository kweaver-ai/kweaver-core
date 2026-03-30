# coding: utf-8
"""
本体查询API封装类
基于 ontology-query.json 接口文档实现
提供对象查询、子图查询、行动查询、属性查询等功能
"""
import allure
from typing import Optional, List, Dict, Any
from src.api.base_api import BaseApi


class OntologyQueryApi(BaseApi):
    """本体查询相关接口封装，基于 ontology-query.json 文档"""

    def __init__(self, user: Optional[str] = None, password: Optional[str] = None):
        super().__init__(user=user, password=password)
        self._prefix = "/api/ontology-query/v1"

    def query_objects_by_path(
        self,
        kn_id: str,
        relation_type_paths: List[str],
        ignoring_store_cache: bool = True,
        limit: Optional[int] = 10,
    ) -> Dict[str, Any]:
        """
        基于关系路径查找目标类的对象实例
        """
        url = f"{self._prefix}/knowledge-networks/{kn_id}/subgraph"
        # 设置查询参数和请求头
        params = {
            "query_type": "relation_path",
            "ignoring_store_cache": ignoring_store_cache,
        }

        payload = {
            "relation_type_paths": relation_type_paths,
            "limit": limit,
        }
        self.headers["X-HTTP-Method-Override"] = "GET"

        with allure.step(f"基于路径查询对象实例 - KN: {kn_id}"):
            return self.post(url, json=payload, params=params)

    def query_objects_by_source_target(
        self,
        kn_id: str,
        source_object_type_id: str,
        need_total: bool = False,
        ignoring_store_cache: bool = True,
        direction: str = "forward",
        path_length: Optional[int] = 1,
        limit: Optional[int] = 10,
        condition: Optional[Dict[str, Any]] = None,
    ) -> Dict[str, Any]:
        """
        基于起点查终点的对象实例
        """
        url = f"{self._prefix}/knowledge-networks/{kn_id}/subgraph"

        # 构建请求体
        payload = {
            "source_object_type_id": source_object_type_id,
            "need_total": need_total,
            "direction": direction,
            "path_length": path_length,
            "condition": condition,
            "limit": limit,
        }

        params = {"ignoring_store_cache": ignoring_store_cache}
        self.headers["X-HTTP-Method-Override"] = "GET"
        with allure.step(f"基于起点终点查询对象实例 - KN: {kn_id}"):
            return self.post(url, json=payload, params=params)

    def query_action(
        self, kn_id: str, action_id: str, unique_identity: List[Dict[str, Any]]
    ) -> Dict[str, Any]:
        """
        行动查询 - 查询行动指导(行动执行参数)

        Args:
            kn_id: 业务知识网络ID
            action_id: 行动类ID
            unique_identity: 对象唯一标识列表，是个map的数组

        Returns:
            行动查询结果，包含行动指导信息
        """
        url = f"{self._prefix}/knowledge-networks/{kn_id}/action-types/{action_id}"

        # 构建请求体
        payload = {"unique_identity": unique_identity}

        # 设置请求头
        headers = self.headers.copy()
        headers["X-HTTP-Method-Override"] = "GET"

        with allure.step(f"行动查询 - KN: {kn_id}, Action: {action_id}"):
            return self.post(url, json=payload, headers=headers)

    def query_properties(
        self,
        kn_id: str,
        ot_id: str,
        property_names: List[str],
        object_type_id: str,
        unique_identity: List[Dict[str, Any]],
        query_type: str = "value",
        dynamic_params: Optional[Dict[str, Any]] = None,
        sort_field: str = "_timestamp",
        sort_direction: str = "desc",
        offset: int = 0,
        limit: int = 10,
        search_after: Optional[List[Any]] = None,
    ) -> Dict[str, Any]:
        """
        属性查询 - 检索及获取已有属性值或计算好的属性值
        """
        # 将property_names转换为逗号分隔的字符串用于URL路径
        property_names_str = ",".join(property_names)
        url = f"{self._prefix}/knowledge-networks/{kn_id}/object-types/{ot_id}/properties/{property_names_str}"

        # 构建请求体
        payload = {
            "object_type_id": object_type_id,
            "property_name": property_names,
            "unique_identity": unique_identity,
            "query_type": query_type,
            "sort": {"field": sort_field, "direction": sort_direction},
            "offset": offset,
            "limit": limit,
        }

        if dynamic_params:
            payload["dynamic_params"] = dynamic_params

        if search_after:
            payload["search_after"] = search_after

        # 设置请求头
        headers = self.headers.copy()
        headers["X-HTTP-Method-Override"] = "GET"

        with allure.step(
            f"属性查询 - KN: {kn_id}, 对象类: {ot_id}, 属性: {property_names}"
        ):
            return self.post(url, json=payload, headers=headers)

    def query_object_type_instances(
        self,
        kn_id: str,
        ot_id: str,
        condition: Optional[Dict[str, Any]] = None,
        limit: int = 10,
        need_total: bool = False,
        include_type_info: bool = False,
        include_logic_params: bool = False,
        ignoring_store_cache: bool = False,
        search_after: Optional[List[Any]] = None,
    ) -> Dict[str, Any]:
        """
        检索指定对象类的对象的详细数据（第一次查询）
        """
        url = f"{self._prefix}/knowledge-networks/{kn_id}/object-types/{ot_id}"

        # 构建请求体
        payload = {"limit": limit, "need_total": need_total}

        if condition is not None:
            payload["condition"] = condition

        if search_after:
            payload["search_after"] = search_after
        # 设置查询参数
        params = {}
        if include_type_info:
            params["include_type_info"] = include_type_info
        if include_logic_params:
            params["include_logic_params"] = include_logic_params
        if ignoring_store_cache:
            params["ignoring_store_cache"] = ignoring_store_cache

        # 设置请求头
        self.headers["X-HTTP-Method-Override"] = "GET"

        with allure.step(f"查询对象类实例 - KN: {kn_id}, 对象类: {ot_id}"):
            return self.post(url, json=payload, params=params)


if __name__ == "__main__":
    import json
    from src.common.db_connector import MySQLConnector

    api = OntologyQueryApi()
    resp = api.query_object_type_instances(
        kn_id="test_kn_for_object_type_jcgo0p8f",
        ot_id="int",
        ignoring_store_cache=True,
    )
    print(json.loads(resp.text))
