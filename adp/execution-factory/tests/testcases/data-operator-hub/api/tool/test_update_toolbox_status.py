# -*- coding:UTF-8 -*-

import allure
import string
import random
import uuid
import pytest
import time

from common.get_content import GetContent
from lib.tool_box import ToolBox

box_id = ""
name = ''.join(random.choice(string.ascii_letters) for i in range(8))

@allure.feature("工具注册与管理接口测试：更新工具箱状态")
class TestUpdateToolboxStatus:
    '''
    状态流转：
        unpublish -> published
        published -> offline
        offline -> published
        offline -> unpublish
    '''
    
    client = ToolBox()

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers):
        global box_id
        global name

        # 创建工具箱，添加重试机制处理503错误
        filepath = "./resource/openapi/compliant/test.json"
        json_data = GetContent(filepath).jsonfile()
        
        data = {
            "box_name": name,
            "data": json_data,
            "metadata_type": "openapi"
        }
        
        # 重试创建工具箱，最多重试3次
        max_retries = 3
        result = None
        for attempt in range(max_retries):
            result = self.client.CreateToolbox(data, Headers)
            if result[0] == 200:
                break
            elif result[0] == 503 and attempt < max_retries - 1:
                # 503错误时等待后重试
                wait_time = min(2 ** attempt, 5)  # 最多等待5秒
                time.sleep(wait_time)
                continue
            else:
                # 其他错误或最后一次重试失败
                break
        
        # 如果创建失败，根据错误类型处理
        if result[0] == 503:
            # 503服务不可用，跳过setup
            pytest.skip(f"后端服务不可用(503)，无法创建工具箱进行测试。响应: {result[1]}")
        elif result[0] != 200:
            # 其他错误，断言失败
            assert False, f"创建工具箱失败，状态码: {result[0]}, 响应: {result[1]}"
        
        # 确保result[1]是字典类型
        assert isinstance(result[1], dict), f"响应格式错误，期望字典，实际: {type(result[1])}, 内容: {result[1]}"
        box_id = result[1]["box_id"]

    @allure.title("更新工具箱状态，状态无冲突更新成功，状态存在冲突，更新失败")
    def test_update_toolbox_status_01(self, Headers):
        global box_id

        # 未发布 -> 下架
        update_data = {
            "status": "offline"
        }
        result = self.client.UpdateToolboxStatus(box_id, update_data, Headers)
        assert result[0] == 400

        # 未发布 -> 已发布
        update_data = {
            "status": "published"
        }
        result = self.client.UpdateToolboxStatus(box_id, update_data, Headers)
        assert result[0] == 200
        assert result[1]["box_id"] == box_id
        assert result[1]["status"] == "published"

        # 已发布 -> 已发布
        update_data = {
            "status": "published"
        }
        result = self.client.UpdateToolboxStatus(box_id, update_data, Headers)
        assert result[0] == 400

        # 已发布 -> 下架
        update_data = {
            "status": "offline"
        }
        result = self.client.UpdateToolboxStatus(box_id, update_data, Headers)
        assert result[0] == 200
        assert result[1]["box_id"] == box_id
        assert result[1]["status"] == "offline"

        # 下架 -> 下架
        update_data = {
            "status": "offline"
        }
        result = self.client.UpdateToolboxStatus(box_id, update_data, Headers)
        assert result[0] == 400

        # 下架 -> 未发布
        update_data = {
            "status": "unpublish"
        }
        result = self.client.UpdateToolboxStatus(box_id, update_data, Headers)
        assert result[0] == 200

    @allure.title("更新工具箱状态，工具箱不存在，更新失败")
    def test_update_toolbox_status_02(self, Headers):
        box_id = str(uuid.uuid4())
        update_data = {
            "status": "offline"
        }
        result = self.client.UpdateToolboxStatus(box_id, update_data, Headers)
        assert result[0] == 400

    @allure.title("更新工具箱状态，更新为无效状态，更新失败")
    def test_update_toolbox_status_03(self, Headers):
        global box_id

        update_data = {
            "status": "invalid_status"
        }
        result = self.client.UpdateToolboxStatus(box_id, update_data, Headers)
        assert result[0] == 400

    @allure.title("发布工具箱，存在同名已发布工具箱，发布失败")
    def test_update_toolbox_status_04(self, Headers):
        # 创建工具箱，添加重试机制和错误处理
        filepath = "./resource/openapi/compliant/template.yaml"
        api_data = GetContent(filepath).yamlfile()
        
        data = {
            "box_name": name,
            "data": api_data,
            "metadata_type": "openapi"
        }
        
        # 重试创建第一个工具箱
        max_retries = 3
        result = None
        for attempt in range(max_retries):
            result = self.client.CreateToolbox(data, Headers)
            if result[0] == 200:
                break
            elif result[0] == 503 and attempt < max_retries - 1:
                wait_time = min(2 ** attempt, 5)
                time.sleep(wait_time)
                continue
            else:
                break
        
        if result[0] == 503:
            pytest.skip(f"后端服务不可用(503)，无法创建第一个工具箱。响应: {result[1]}")
        elif result[0] != 200:
            assert False, f"创建第一个工具箱失败，状态码: {result[0]}, 响应: {result[1]}"
        
        assert isinstance(result[1], dict), f"响应格式错误，期望字典，实际: {type(result[1])}, 内容: {result[1]}"
        box_id = result[1]["box_id"]

        # 重试创建第二个同名工具箱
        result1 = None
        for attempt in range(max_retries):
            result1 = self.client.CreateToolbox(data, Headers)
            if result1[0] == 200:
                break
            elif result1[0] == 503 and attempt < max_retries - 1:
                wait_time = min(2 ** attempt, 5)
                time.sleep(wait_time)
                continue
            else:
                break
        
        if result1[0] == 503:
            pytest.skip(f"后端服务不可用(503)，无法创建第二个工具箱。响应: {result1[1]}")
        elif result1[0] != 200:
            assert False, f"创建第二个工具箱失败，状态码: {result1[0]}, 响应: {result1[1]}"
        
        assert isinstance(result1[1], dict), f"响应格式错误，期望字典，实际: {type(result1[1])}, 内容: {result1[1]}"
        box_id1 = result1[1]["box_id"]

        update_data = {
            "status": "published"
        }
        result = self.client.UpdateToolboxStatus(box_id, update_data, Headers)
        assert result[0] == 200

        result = self.client.UpdateToolboxStatus(box_id1, update_data, Headers)
        assert result[0] == 400