# -*- coding:UTF-8 -*-

import allure
import uuid
import pytest

from common.get_content import GetContent
from lib.operator import Operator

published_operator_infos = []
unpublish_operator_infos = []
offline_operator_infos = []

@allure.feature("算子注册与管理接口测试：更新算子信息")
class TestOperatorUpdateInfo:
  
    client = Operator()

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers):
        global published_operator_infos, unpublish_operator_infos, offline_operator_infos

        filepath = "./resource/openapi/compliant/setup.json"
        operator_data = GetContent(filepath).jsonfile()
        req_data = {
            "data": str(operator_data),
            "operator_metadata_type": "openapi"
        }
        re = self.client.RegisterOperator(req_data, Headers)
        assert re[0] == 200
        count = int(len(re[1]))
        
        # 计算需要发布的算子数量（三分之二）
        publish_count = int(count * 2 / 3)
        
        # 获取所有算子ID和version
        operator_ids = []
        versions = []
        for operator in re[1]:
            operator_ids.append(operator["operator_id"])
            versions.append(operator["version"])
        
        # 发布三分之二的算子，然后下架其中的一半
        publish_datas = []
        offline_datas = []
        for i in range(publish_count):
            publish_data = {
                "operator_id": operator_ids[i],
                "version": versions[i],
                "status": "published"
            }
            publish_datas.append(publish_data)

            if i%2 == 0:
                offline_data = {
                    "operator_id": operator_ids[i],
                    "version": versions[i],
                    "status": "offline"
                }
                offline_datas.append(offline_data)

        result = self.client.UpdateOperatorStatus(publish_datas, Headers)
        assert result[0] == 200
        result = self.client.UpdateOperatorStatus(offline_datas, Headers)
        assert result[0] == 200

        data = {"status": "published"}
        re = self.client.GetOperatorList(data, Headers)
        
        for operator in re[1]["data"]:
            operator_info = {
                "operator_id": operator["operator_id"],
                "version": operator["version"]
            }
            published_operator_infos.append(operator_info)

        data = {"status": "unpublish"}
        re = self.client.GetOperatorList(data, Headers)
        for operator in re[1]["data"]:
            operator_info = {
                "operator_id": operator["operator_id"],
                "version": operator["version"]
            }
            unpublish_operator_infos.append(operator_info)

        data = {"status": "offline"}
        re = self.client.GetOperatorList(data, Headers)
        for operator in re[1]["data"]:
            operator_info = {
                "operator_id": operator["operator_id"],
                "version": operator["version"]
            }
            offline_operator_infos.append(operator_info)
            

    @allure.title("算子不存在，更新算子信息失败")
    def test_update_operator_info_01(self, Headers):
        filepath = "./resource/openapi/compliant/update-test1.yaml"
        api_data = GetContent(filepath).yamlfile()

        data = {
            "operator_id": str(uuid.uuid4()),
            "data": str(api_data),
            "operator_metadata_type": "openapi"
        }

        result = self.client.UpdateOperatorInfo(data, Headers)
        assert result[0] == 404

    @allure.title("data中包含多个算子，更新算子信息失败")
    def test_update_operator_info_02(self, Headers):
        global published_operator_infos

        filepath = "./resource/openapi/compliant/update-test2.yaml"
        api_data = GetContent(filepath).yamlfile()
        data = {
            "operator_id": published_operator_infos[0]["operator_id"],
            "data": str(api_data),
            "operator_metadata_type": "openapi"
        }
        result = self.client.UpdateOperatorInfo(data, Headers)
        assert result[0] == 400

    @allure.title("更新已下架算子信息，更新成功, 生成新版本，状态为未发布")
    def test_update_operator_info_03(self, Headers):
        global offline_operator_infos

        filepath = "./resource/openapi/compliant/update-test1.yaml"
        api_data = GetContent(filepath).yamlfile()
        data = {
            "operator_id": offline_operator_infos[0]["operator_id"],
            "data": str(api_data),
            "operator_metadata_type": "openapi"
        }
        result = self.client.UpdateOperatorInfo(data, Headers)
        assert result[0] == 200
        assert result[1][0]["status"] == "success"
        assert result[1][0]["version"] != offline_operator_infos[0]["version"]
        
        re = self.client.GetOperatorInfo(result[1][0]["operator_id"], Headers)  
        assert re[0] == 200
        assert re[1]["status"] == "unpublish"
        
    @allure.title("更新已发布算子信息，更新成功，默认更新后的版本为已发布编辑中状态")
    def test_update_operator_info_04(self, Headers):
        global published_operator_infos

        filepath = "./resource/openapi/compliant/update-test1.yaml"
        api_data = GetContent(filepath).yamlfile()
        data = {
            "operator_id": published_operator_infos[0]["operator_id"],
            "data": str(api_data),
            "operator_metadata_type": "openapi"
        }
        result = self.client.UpdateOperatorInfo(data, Headers)
        assert result[0] == 200
        assert result[1][0]["status"] == "success"
        assert result[1][0]["operator_id"] == published_operator_infos[0]["operator_id"]
        assert result[1][0]["version"] != published_operator_infos[0]["version"]
        
        re = self.client.GetOperatorInfo(result[1][0]["operator_id"], Headers)  
        assert re[0] == 200
        assert re[1]["status"] == "editing"

    @allure.title("更新未发布算子信息，更新后直接发布，更新成功，无新版本生成")
    def test_update_operator_info_05(self, Headers):
        global unpublish_operator_infos

        filepath = "./resource/openapi/compliant/update-test1.yaml"
        api_data = GetContent(filepath).yamlfile()
        data = {
            "operator_id": unpublish_operator_infos[0]["operator_id"],
            "data": str(api_data),
            "operator_metadata_type": "openapi",
            "direct_publish": True
        }
        result = self.client.UpdateOperatorInfo(data, Headers)
        assert result[0] == 200
        assert result[1][0]["status"] == "success"
        assert result[1][0]["operator_id"] == unpublish_operator_infos[0]["operator_id"]
        assert result[1][0]["version"] == unpublish_operator_infos[0]["version"]
        
        re = self.client.GetOperatorInfo(result[1][0]["operator_id"], Headers)  
        assert re[0] == 200
        assert re[1]["status"] == "published"

    @allure.title("更新算子信息，执行模式为异步，标识为数据源算子，更新失败")
    def test_update_operator_info_06(self, Headers):
        global unpublish_operator_infos

        filepath = "./resource/openapi/compliant/update-test1.yaml"
        api_data = GetContent(filepath).yamlfile()
        data = {
            "operator_id": unpublish_operator_infos[0]["operator_id"],
            "data": str(api_data),
            "operator_metadata_type": "openapi",
            "operator_info": {
                "execution_mode": "async",
                "is_data_source": True
            }
        }
        result = self.client.UpdateOperatorInfo(data, Headers)
        assert result[0] == 400

    @allure.title("更新算子信息，执行模式为异步，标识为数据源算子，更新失败")
    def test_update_operator_info_07(self, Headers):
        global unpublish_operator_infos

        filepath = "./resource/openapi/compliant/update-test1.yaml"
        api_data = GetContent(filepath).yamlfile()
        data = {
            "operator_id": unpublish_operator_infos[0]["operator_id"],
            "data": str(api_data),
            "operator_metadata_type": "openapi",
            "operator_info": {
                "execution_mode": "sync",
                "is_data_source": True
            }
        }
        result = self.client.UpdateOperatorInfo(data, Headers)
        assert result[0] == 200
        operator_id = result[1][0]["operator_id"]
        result = self.client.GetOperatorInfo(operator_id, Headers)
        assert result[0] == 200
        assert result[1]["operator_id"] == operator_id
        operator_info = result[1]["operator_info"]
        assert operator_info["execution_mode"] == "sync"
        assert operator_info["is_data_source"] == True
        
        
        
        
        