import allure
import json
import os
from typing import Optional
from src.api.base_api import BaseApi
from src.common.data_gen import random_icon, random_color


class KnowledgeNetworkApi(BaseApi):
    """业务知识网络相关接口封装，基于 OpenAPI 文档 test_ontology-manager-network.json"""

    def __init__(self, user: Optional[str] = None, password: Optional[str] = None):
        super().__init__(user=user, password=password)
        self._prefix = "/api/ontology-manager/v1"

    def list_kn(
        self,
        name_pattern: Optional[str] = None,
        sort: Optional[str] = "update_time",
        direction: Optional[str] = "asc",
        offset: Optional[int] = 0,
        limit: Optional[int] = 100,
        tag: Optional[str] = "自动化",
    ):
        """
        GET /knowledge-networks
        """
        url = f"{self._prefix}/knowledge-networks"
        params = {
            "name_pattern": name_pattern,
            "sort": sort,
            "direction": direction,
            "offset": offset,
            "limit": limit,
            "tag": tag,
        }
        self.headers["x-business-domain"] = "bd_public"

        with allure.step("获取业务知识网络列表"):
            return self.get(url, params=params)

    def create_kn(
        self,
        kn_name: str,
        kn_id: str,
        tags: Optional[list] = None,
        comment: Optional[str] = None,
        exclude_fields: Optional[list] = None,
    ):
        """
        POST /knowledge-networks
        请求体参考 ReqKnowledgeNetwork：必须包含 name, branch, base_branch，可选 id/tags/comment/icon/color
        """
        payload = {
            "id": kn_id,
            "name": kn_name,
            "tags": tags if tags is not None else ["自动化", "示例"],
            "comment": comment if comment is not None else "通过自动化测试创建",
            "icon": random_icon(),
            "color": random_color(),
            "branch": "main",
            "base_branch": "",
        }
        if exclude_fields:  # 移除某些字段,参数校验使用
            for field in exclude_fields:
                payload.pop(field, None)
        url = f"{self._prefix}/knowledge-networks"
        self.headers["x-business-domain"] = "bd_public"
        with allure.step(f"创建业务知识网络，名称为{kn_name}"):
            return self.post(url, json=payload)

    def get_knowledge_network(self, kn_id: str, include_detail: Optional[bool] = False):
        """
        GET /knowledge-networks/{kn_id}
        include_detail：是否包含说明书信息
        """
        url = f"{self._prefix}/knowledge-networks/{kn_id}"
        params = (
            {"include_detail": include_detail} if include_detail is not None else None
        )
        self.headers["x-business-domain"] = "bd_public"

        return self.get(url, params=params)

    def update_kn(
        self,
        kn_name: str,
        kn_id: str,
        tags: Optional[list] = None,
        comment: Optional[str] = None,
        exclude_fields: Optional[list] = None,
    ):
        """
        PUT /knowledge-networks/{kn_id}
        请求体参考 UpdateKnowledgeNetwork：必须包含 name, branch, base_branch，可选 tags/comment/icon/color
        """
        payload = {
            "name": kn_name,
            "tags": tags if tags is not None else ["自动化", "更新"],
            "comment": comment if comment is not None else "通过自动化测试更新",
            "icon": random_icon(),
            "color": random_color(),
            "branch": "main",
            "base_branch": "",
        }
        if exclude_fields:
            for field in exclude_fields:
                payload.pop(field, None)

        url = f"{self._prefix}/knowledge-networks/{kn_id}"
        with allure.step(f"更新业务知识网络，名称为{kn_name}"):
            return self.put(url, json=payload)

    def delete_kn(self, kn_id: str):
        """
        delete /knowledge-networks/{kn_id}
        """
        url = f"{self._prefix}/knowledge-networks/{kn_id}"
        with allure.step(f"删除业务知识网络，id为{kn_id}"):
            return self.delete(url)

    def kn_job(self, kn_id: str, kn_name: str, job_type: str = "full"):
        url = f"{self._prefix}/knowledge-networks/{kn_id}/jobs"
        payload = {"name": kn_name, "job_type": job_type}
        with allure.step(f"业务知识网络执行构建任务，id为{kn_id}"):
            return self.post(url, json=payload)

    def kn_job_list(self, kn_id: str):
        url = f"{self._prefix}/knowledge-networks/{kn_id}/jobs"
        params = {"offset": 0, "limit": 50, "sort": "create_time", "direction": "asc"}
        with allure.step(f"获取业务知识网络构建任务，id为{kn_id}"):
            return self.get(url, params=params)

    def export_kn(self, kn_id: str):
        url = f"{self._prefix}/knowledge-networks/{kn_id}/"
        params = {"include_detail": False, "mode": "export"}

        with allure.step(f"导出业务知识网络，id为{kn_id}"):
            # 确保导出目录存在
            os.makedirs("./export", exist_ok=True)

            response = self.get(url, params=params, stream=True, timeout=600)
            if response.status_code == 200:
                file_name = f"{kn_id}.json"  # 默认文件名
                save_path = f"./export/{file_name}"

                # 写入所有数据块
                total_size = 0
                with open(save_path, "wb") as f:
                    for chunk in response.iter_content(chunk_size=1024):  # 分块写入
                        if chunk:
                            f.write(chunk)
                            total_size += len(chunk)
                return response

    def import_kn(self, kn_id: str):
        url = f"{self._prefix}/knowledge-networks"
        file_path = f"./export/{kn_id}.json"
        with allure.step(f"导入业务知识网络，id为{kn_id}"):
            with open(file_path, "r", encoding="utf-8") as f:
                payload = json.loads(f.read())
                payload["id"] = kn_id + "_import"
                payload["name"] = kn_id + "_import"
                self.headers["x-business-domain"] = "bd_public"
                return self.post(url, json=payload)


if __name__ == "__main__":
    kn_id = f"test_kn_for_object_type_7"
    kn_api = KnowledgeNetworkApi()
    resp = kn_api.list_kn()
    data = json.loads(resp.text)["entries"]
    for i in data:
        kn_api.delete_kn(i["id"])
