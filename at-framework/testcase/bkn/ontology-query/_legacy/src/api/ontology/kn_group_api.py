import allure
from typing import Optional
from src.api.base_api import BaseApi
from src.common.data_gen import random_icon, random_color


class ConceptGroupApi(BaseApi):
    """业务知识网络相关接口封装，基于 OpenAPI 文档 test_ontology-manager-network.json"""

    def __init__(self, user: Optional[str] = None, password: Optional[str] = None):
        super().__init__(user=user, password=password)
        self._prefix = "/api/ontology-manager/v1"

    def list_concept_group(self, kn_id: str = None,
                           name_pattern: Optional[str] = None,
                           sort: Optional[str] = "update_time",
                           direction: Optional[str] = "asc",
                           offset: Optional[int] = 0,
                           limit: Optional[int] = 100,
                           tag: Optional[str] = "自动化"):
        url = f"{self._prefix}/knowledge-networks/{kn_id}/concept-groups"
        params = {
            "name_pattern": name_pattern,
            "sort": sort,
            "direction": direction,
            "offset": offset,
            "limit": limit,
            "tag": tag,
        }
        self.headers['x-business-domain'] = 'bd_public'
        with allure.step(f"获取业务知识网络{kn_id}下的概念分组"):
            return self.get(url, params=params)

    def create_concept_groups(self, kn_id: str, groups_name: str, groups_id: str):
        payload = {
            "kn_id": kn_id,
            "id": groups_id,
            "name": groups_name,
            "tags": ["自动化", "示例"],
            "comment": "通过自动化测试创建",
            "icon": "icon-dip-group",
            "color": random_color(),
            "branch": "main"
        }
        url = f"{self._prefix}/knowledge-networks/{kn_id}/concept-groups"

        # self.headers['x-business-domain'] = 'bd_public'
        with allure.step(f"创建概念分组，名称为{groups_name}"):
            return self.post(url, json=payload)

    def update_concept_groups(self, kn_id: str, groups_name: str, groups_id: str):
        payload = {
            "kn_id": kn_id,
            "name": groups_name,
            "tags": ["自动化", "示例"],
            "comment": "通过自动化测试创建",
            "icon": "icon-dip-group",
            "color": random_color(),
            "branch": "main"
        }
        url = f"{self._prefix}/knowledge-networks/{kn_id}/concept-groups/{groups_id}"

        # self.headers['x-business-domain'] = 'bd_public'
        with allure.step(f"修改概念分组，名称为{groups_name}"):
            return self.put(url, json=payload)

    def delete_concept_groups(self, kn_id: str, groups_id: str):
        url = f"{self._prefix}/knowledge-networks/{kn_id}/concept-groups/{groups_id}"
        with allure.step(f"删除概念分组，id为{groups_id}"):
            return self.delete(url)

    def concept_groups_add_object(self, kn_id: str, groups_id: str, objects: list):
        url = f"{self._prefix}/knowledge-networks/{kn_id}/concept-groups/{groups_id}/object-types"

        payload = {
            "entries": [{"id": item} for item in objects]
        }
        with allure.step(f"概念分组{groups_id}添加对象类，{str(objects)}"):
            return self.post(url, json=payload)

    def concept_groups_remove_object(self, kn_id: str, groups_id: str, objects: list):
        objects_id = ",".join(objects)

        url = f"{self._prefix}/knowledge-networks/{kn_id}/concept-groups/{groups_id}/object-types/{objects_id}"

        with allure.step(f"概念分组{groups_id}移除对象类，{objects_id}"):
            return self.delete(url)

    def get_groups_views(self, kn_id: str, groups_id: str):
        url = f"{self._prefix}/knowledge-networks/{kn_id}/concept-groups/{groups_id}"

        with allure.step(f"查看概念分组{groups_id}详情"):
            return self.get(url)


if __name__ == '__main__':
    kn_api = ConceptGroupApi()
    kn_api.get_groups_views("auto_test_048", 'test_add')
