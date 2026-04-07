# -*- coding:UTF-8 -*-

import allure

from common.get_content import GetContent
from lib.operator import Operator


@allure.feature("算子注册与管理接口测试：更新算子状态")
class TestUpdateOperatorStatus:
    '''
    算子状态转换：
        unpublish -> published
        published -> offline
        offline -> published
        offline -> unpublish
        published -> editing
        editing -> published
    '''
    client = Operator()
    
    @allure.title("单个算子状态更新，算子存在，更新算子状态，更新成功")
    def test_update_operator_status_01(self, Headers):
        filepath = "./resource/openapi/compliant/update-test1.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        data = {
            "data": str(yaml_data),
            "operator_metadata_type": "openapi",
            "direct_publish": False
        }
        result = self.client.RegisterOperator(data, Headers)  
        assert result[0] == 200
        assert len(result[1]) == 1
        operator_id = result[1][0]["operator_id"]
        
        update_data = [
            {
                "operator_id": operator_id,
                "status": "published"
            }
        ] 
        result = self.client.UpdateOperatorStatus(update_data, Headers)    # 未发布 -> 已发布
        assert result[0] == 200

        result = self.client.GetOperatorInfo(operator_id, Headers)
        assert result[0] == 200
        assert result[1]["status"] == "published"

        update_data = [
            {
                "operator_id": operator_id,
                "status": "offline"
            }
        ] 
        result = self.client.UpdateOperatorStatus(update_data, Headers)    # 已发布 -> 下架
        assert result[0] == 200

        data = {
            "page_size": -1,
            "status": "offline"
        }
        result = self.client.GetOperatorList(data, Headers)
        assert result[0] == 200
        operator_ids = []
        for operator in result[1]["data"]:
            operator_ids.append(operator["operator_id"])

        assert operator_id in operator_ids

        update_data = [
            {
                "operator_id": operator_id,
                "status": "published"
            }
        ] 
        result = self.client.UpdateOperatorStatus(update_data, Headers)    # 下架 -> 已发布
        assert result[0] == 200

        result = self.client.GetOperatorInfo(operator_id, Headers)
        assert result[0] == 200
        assert result[1]["status"] == "published"

    @allure.title("单个算子状态更新，算子不存在，更新算子状态，更新失败")
    def test_update_operator_status_02(self, Headers):
        update_data = [
            {
                "operator_id": "85a4000e-bcbf-45a2-a933-ae8569009649",
                "status": "published"
            }
        ] 
        result = self.client.UpdateOperatorStatus(update_data, Headers)
        assert result[0] == 404

    @allure.title("批量算子状态更新，算子均存在，更新算子状态，更新成功")
    def test_update_operator_status_03(self, Headers):
        filepath = "./resource/openapi/compliant/update-test2.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        data = {
            "data": str(yaml_data),
            "operator_metadata_type": "openapi",
            "direct_publish": False
        }
        result = self.client.RegisterOperator(data, Headers)
        assert result[0] == 200
        update_datas = []
        for operator in result[1]:
            update_data = {
                "operator_id": operator["operator_id"],
                "status": "published"
            }
            update_datas.append(update_data)
        
        result = self.client.UpdateOperatorStatus(update_datas, Headers)
        assert result[0] == 200

        for operator in update_datas:
            id = operator["operator_id"]
            result = self.client.GetOperatorInfo(id, Headers)
            assert result[0] == 200
            assert result[1]["status"] == "published"

    @allure.title("批量算子状态更新，部分算子不存在，更新算子状态，更新失败")
    def test_update_operator_status_04(self, Headers):
        filepath = "./resource/openapi/compliant/update-test3.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        data = {
            "data": str(yaml_data),
            "operator_metadata_type": "openapi",
            "direct_publish": False
        }
        result = self.client.RegisterOperator(data, Headers)
        assert result[0] == 200
        update_datas = []
        for operator in result[1]:
            update_data = {
                "operator_id": operator["operator_id"],
                "status": "published"
            }
            update_datas.append(update_data)

        not_exits_operator = {
            "operator_id": "85a4000e-bcbf-45a2-a933-ae8569009649",
            "status": "published"
        }
        update_datas.append(not_exits_operator)

        result = self.client.UpdateOperatorStatus(update_datas, Headers)
        assert result[0] == 404

    @allure.title("算子状态更新为未知状态，更新失败")
    def test_update_operator_status_05(self, Headers):
        filepath = "./resource/openapi/compliant/update-test4.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        data = {
            "data": str(yaml_data),
            "operator_metadata_type": "openapi",
            "direct_publish": False
        }
        result = self.client.RegisterOperator(data, Headers)
        assert result[0] == 200
        assert len(result[1]) == 1
        operator_id = result[1][0]["operator_id"]
        version = result[1][0]["version"]
        
        update_data = [
            {
                "operator_id": operator_id,
                "status": "abandon"
            }
        ] 
        result = self.client.UpdateOperatorStatus(update_data, Headers)
        assert result[0] == 400

    @allure.title("算子状态转换存在冲突时更新失败，无冲突时更新成功")
    def test_update_operator_status_06(self, Headers):
        filepath = "./resource/openapi/compliant/update-test4.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        data = {
            "data": str(yaml_data),
            "operator_metadata_type": "openapi"
        }
        result = self.client.RegisterOperator(data, Headers)
        assert result[0] == 200
        assert len(result[1]) == 1
        operator_id = result[1][0]["operator_id"]

        update_data = [
            {
                "operator_id": operator_id,
                "status": "editing"
            }
        ] 
        result = self.client.UpdateOperatorStatus(update_data, Headers)    # 未发布 -> 已发布编辑中
        assert result[0] == 400

        update_data = [
            {
                "operator_id": operator_id,
                "status": "offline"
            }
        ] 
        result = self.client.UpdateOperatorStatus(update_data, Headers)    # 未发布 -> 已下架
        assert result[0] == 400

        update_data = [
            {
                "operator_id": operator_id,
                "status": "published"
            }
        ] 
        result = self.client.UpdateOperatorStatus(update_data, Headers)    # 未发布 -> 已发布
        assert result[0] == 200
        
        update_data = [
            {
                "operator_id": operator_id,
                "status": "unpublish"
            }
        ] 
        result = self.client.UpdateOperatorStatus(update_data, Headers)    # 已发布 -> 未发布
        assert result[0] == 400

        update_data = [
            {
                "operator_id": operator_id,
                "status": "published"
            }
        ] 
        result = self.client.UpdateOperatorStatus(update_data, Headers)    # 已发布 -> 已发布
        assert result[0] == 400

        update_data = [
            {
                "operator_id": operator_id,
                "status": "offline"
            }
        ] 
        result = self.client.UpdateOperatorStatus(update_data, Headers)    # 已发布 -> 下架
        assert result[0] == 200

        update_data = [
            {
                "operator_id": operator_id,
                "status": "unpublish"
            }
        ] 
        result = self.client.UpdateOperatorStatus(update_data, Headers)    # 下架 -> 未发布
        assert result[0] == 200

        update_data = [
            {
                "operator_id": operator_id,
                "status": "published"
            }
        ] 
        result = self.client.UpdateOperatorStatus(update_data, Headers)    # 下架 -> 已发布
        assert result[0] == 200

        update_data = [
            {
                "operator_id": operator_id,
                "status": "editing"
            }
        ] 
        result = self.client.UpdateOperatorStatus(update_data, Headers)    # 已发布 -> 已发布编辑中
        assert result[0] == 200

        update_data = [
            {
                "operator_id": operator_id,
                "status": "offline"
            }
        ] 
        result = self.client.UpdateOperatorStatus(update_data, Headers)    # 已发布编辑中 -> 下架
        assert result[0] == 400

        update_data = [
            {
                "operator_id": operator_id,
                "status": "unpublish"
            }
        ] 
        result = self.client.UpdateOperatorStatus(update_data, Headers)    # 已发布编辑中 -> 未发布
        assert result[0] == 400

        update_data = [
            {
                "operator_id": operator_id,
                "status": "editing"
            }
        ] 
        result = self.client.UpdateOperatorStatus(update_data, Headers)    # 已发布编辑中 -> 已发布
        assert result[0] == 200

    @allure.title("发布算子，已存在同名已发布算子，发布失败")
    def test_update_operator_status_07(self, Headers):
        filepath = "./resource/openapi/compliant/template.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        data = {
            "data": str(yaml_data),
            "operator_metadata_type": "openapi"
        }
        result = self.client.RegisterOperator(data, Headers)
        assert result[0] == 200
        operator_id = result[1][0]["operator_id"]

        result1 = self.client.RegisterOperator(data, Headers)
        assert result1[0] == 200
        operator_id1 = result1[1][0]["operator_id"]

        update_data = [
            {
                "operator_id": operator_id,
                "status": "published"
            }
        ] 
        result = self.client.UpdateOperatorStatus(update_data, Headers)
        assert result[0] == 200

        update_data1 = [
            {
                "operator_id": operator_id1,
                "status": "published"
            }
        ] 
        result = self.client.UpdateOperatorStatus(update_data1, Headers)
        assert result[0] == 409