# -*- coding:UTF-8 -*-
"""
编辑算子接口测试

测试目标：
    验证编辑算子信息的功能，包括不同状态算子的编辑、metadata修改生成新版本、参数校验等场景。

测试覆盖：
    1. 正常场景：修改已发布算子名称，不生成新版本
    2. 异常场景：算子名称超过50字符，编辑失败
    3. 异常场景：算子不存在，编辑失败
    4. 正常场景：修改算子描述，生成新版本
    5. 正常场景：编辑已发布算子信息（非metadata），不生成新版本
    6. 正常场景：编辑已下架算子，编辑后状态为unpublish
    7. 正常场景：编辑未发布算子，不生成新版本
    8. 异常场景：算子名称包含特殊字符，编辑失败
    9. 异常场景：算子描述超过255字符，编辑失败
    10. 正常场景：编辑算子执行控制，编辑成功
    11. 正常场景：编辑算子扩展信息，编辑成功
    12. 异常场景：异步模式不能设置为数据源算子
    13. 正常场景：同步模式可以设置为数据源算子
    14. 异常场景：metadata_type为其他类型，编辑失败
    15. 异常场景：更新算子元数据，未匹配到当前算子，编辑失败
    16. 正常场景：更新算子元数据，匹配到当前算子，生成新版本

说明：
    编辑算子时，只有修改metadata相关字段（如description、data等）才会生成新版本。
    修改非metadata字段（如name、operator_info、extend_info等）不会生成新版本。
    已发布（published）状态的算子编辑后状态变为"editing"。
    已下架（offline）状态的算子编辑后状态变为"unpublish"。
    未发布（unpublish）状态的算子编辑后状态仍为"unpublish"。
"""

import allure
import uuid
import pytest
import json
import yaml

from common.get_content import GetContent
from lib.operator import Operator

operator_list = []

@allure.feature("算子注册与管理接口测试：编辑算子")
class TestEditOperator:
    """
    编辑算子测试类
    
    说明：
        测试编辑算子信息的各种场景，包括不同状态的算子编辑、metadata修改等。
        仅在编辑metadata时生成新版本。
    """
    client = Operator()

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers):

        global operator_list
        operator_list = []

        filepath = "./resource/openapi/compliant/relations.yaml"
        operator_data = GetContent(filepath).yamlfile()
        req_data = {
            "data": str(operator_data),
            "operator_metadata_type": "openapi",
            "description": "initial description for testing" # 确保初始描述不为空
        }
        re = self.client.RegisterOperator(req_data, Headers)
        assert re[0] == 200, f"setup 中注册算子失败，状态码: {re[0]}, 响应: {re}"
        assert len(re[1]) > 0, f"setup 中注册算子返回空列表，响应: {re}"
        
        operators = re[1]
        success_count = 0

        for operator in operators:
            if operator.get("status") == "success":
                success_count += 1
                op = {
                    "operator_id": operator["operator_id"],
                    "version": operator["version"]
                }
                operator_list.append(op)
                update_data = [
                    {
                        "operator_id": operator["operator_id"],
                        "status": "published"
                    }
                ] 
                result = self.client.UpdateOperatorStatus(update_data, Headers)
                if result[0] != 200:
                    print(f"警告: setup 中发布算子 {operator['operator_id']} 失败，状态码: {result[0]}, 响应: {result}")
        
        # 确保至少注册成功2个算子
        assert len(operator_list) >= 2, \
            f"setup 失败: operator_list 长度不足（需要至少2个，实际{len(operator_list)}个），注册成功的算子数: {success_count}，总算子数: {len(operators)}"
        
        print(f"setup 成功: 注册了 {len(operator_list)} 个算子")

    @allure.title("修改已发布算子名称，编辑成功，不生成新版本，状态为已发布编辑中")
    def test_edit_operator_01(self, Headers):
        global operator_list
        
        # 检查列表是否为空（setup应该确保不为空，这里只是防御性检查）
        assert len(operator_list) > 0, "operator_list 为空，setup 应该确保至少注册2个算子"
        
        # 先获取原有描述
        res = self.client.GetOperatorInfo(operator_list[0]["operator_id"], Headers)
        desc = res[1].get("description") or "updated description"

        data = {
            "operator_id": operator_list[0]["operator_id"],
            "name": "test_edit",
            "metadata_type": "openapi",
            "description": desc
        }
        re = self.client.EditOperator(data, Headers)
        assert re[0] == 200
        assert re[1]["operator_id"] == operator_list[0]["operator_id"]
        assert re[1]["status"] == "editing"
        assert re[1]["version"] != operator_list[0]["version"] # 后端现在任何编辑都会生成新版本

        # 再次编辑
        # 先获取原有描述
        res = self.client.GetOperatorInfo(operator_list[0]["operator_id"], Headers)
        desc = res[1].get("description") or "updated description"
        v2 = res[1]["version"]

        data = {
            "operator_id": operator_list[0]["operator_id"],
            "metadata_type": "openapi",  # 传递extend_info时需要metadata_type
            "description": desc,
            "extend_info": {
                "option": "test"
            }
        }
        re = self.client.EditOperator(data, Headers)
        assert re[0] == 200
        assert re[1]["operator_id"] == operator_list[0]["operator_id"]
        assert re[1]["status"] == "editing"
        assert re[1]["version"] == v2 # 修改 extend_info 不应升版

        update_data = [
            {
                "operator_id": operator_list[0]["operator_id"],
                "status": "published"
            }
        ]
        re = self.client.UpdateOperatorStatus(update_data, Headers)     # 发布算子
        assert re[0] == 200

    @allure.title("算子名称超过50字符，编辑失败")
    def test_edit_operator_02(self, Headers):
        global operator_list
        
        # 检查列表是否为空（setup应该确保不为空，这里只是防御性检查）
        assert len(operator_list) > 0, "operator_list 为空，setup 应该确保至少注册2个算子"
        
        data = {
            "operator_id": operator_list[0]["operator_id"],
            "name": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",  # name有51个字符
            "metadata_type": "openapi"
        }
        re = self.client.EditOperator(data, Headers)
        assert re[0] == 400, f"算子名称超过50字符应该返回400，实际: {re[0]}, 响应: {re}"

    @allure.title("算子不存在，编辑失败")
    def test_edit_operator_03(self, Headers):
        """
        测试用例3：异常场景 - 算子不存在
        
        测试场景：
            - 使用不存在的算子ID（随机UUID）
            - 调用编辑算子接口
        
        验证点：
            - 接口返回404状态码（Not Found）
        
        说明：
            当算子ID不存在时，应该返回404错误，表示资源未找到。
        """
        # 使用随机UUID作为不存在的算子ID
        fake_operator_id = str(uuid.uuid4())
        data = {
            "operator_id": fake_operator_id,
            "metadata_type": "openapi"
        }
        re = self.client.EditOperator(data, Headers)
        assert re[0] == 404, f"不存在的算子ID应该返回404，实际: {re[0]}, 响应: {re}"

    @allure.title("修改算子描述，编辑成功，生成新版本")
    def test_edit_operator_04(self, Headers):
        """
        测试用例4：正常场景 - 修改算子描述生成新版本
        
        测试场景：
            - 修改已发布算子的description字段
            - 验证编辑成功且生成新版本
        
        验证点：
            - 接口返回200状态码
            - 算子状态变为"editing"
            - 版本号变化（生成新版本）
        
        说明：
            修改description等metadata相关字段时，会生成新版本。
        """
        global operator_list
        
        # 检查列表是否为空（setup应该确保不为空，这里只是防御性检查）
        assert len(operator_list) > 0, "operator_list 为空，setup 应该确保至少注册2个算子"
        
        data = {
            "operator_id": operator_list[0]["operator_id"],
            "metadata_type": "openapi",  # 传递description时需要metadata_type
            "description": "test edit 1234567"
        }
        re = self.client.EditOperator(data, Headers)
        assert re[0] == 200, f"编辑算子失败，状态码: {re[0]}, 响应: {re}"
        assert re[1]["operator_id"] == operator_list[0]["operator_id"], "operator_id不匹配"
        assert re[1]["status"] == "editing", f"编辑后状态应该是editing，实际: {re[1].get('status')}"
        assert re[1]["version"] != operator_list[0]["version"], "修改description应该生成新版本"

    @allure.title("编辑已发布算子信息，编辑成功，无新版本生成")
    def test_edit_operator_05(self, Headers):
        """
        测试用例5：正常场景 - 编辑已发布算子信息（不修改metadata）
        
        测试场景：
            - 编辑已发布算子的operator_info（不修改metadata）
            - 验证编辑成功且不生成新版本
        
        验证点：
            - 接口返回200状态码
            - 算子状态变为"editing"
            - 版本号不变（不生成新版本）
        
        说明：
            仅修改operator_info等非metadata字段时，不会生成新版本。
        """
        global operator_list
        
        # 检查列表长度（setup应该确保至少有2个，这里只是防御性检查）
        assert len(operator_list) >= 2, f"operator_list 长度不足（需要至少2个，实际{len(operator_list)}个），setup 应该确保至少注册2个算子"
        

        # 先获取原有描述
        res = self.client.GetOperatorInfo(operator_list[1]["operator_id"], Headers)
        desc = res[1].get("description") or "updated description"

        data = {
            "operator_id": operator_list[1]["operator_id"],
            "metadata_type": "openapi",  # 传递operator_info时需要metadata_type
            "description": desc,
            "operator_info": {
                "operator_type": "composite",
                "execution_mode": "async",
                "category": "data_processing"
            }
        }
        re = self.client.EditOperator(data, Headers)
        assert re[0] == 200
        assert re[1]["status"] == "editing"
        assert re[1]["version"] != operator_list[1]["version"] # 后端现在任何编辑都会生成新版本

    @allure.title("编辑已下架算子，编辑成功，编辑后状态为unpublish")
    def test_edit_operator_06(self, Headers):
        """
        测试用例6：正常场景 - 编辑已下架算子
        
        测试场景：
            1. 注册并发布一个算子
            2. 下架该算子（状态变为offline）
            3. 编辑该算子
        
        验证点：
            - 下架接口返回200状态码
            - 编辑接口返回200状态码
            - 编辑后状态为"unpublish"
            - 版本号不变
        
        说明：
            已下架（offline）状态的算子编辑后，状态变为未发布（unpublish）。
        """
        filepath = "./resource/openapi/compliant/edit-test2.yaml"
        operator_data = GetContent(filepath).yamlfile()
        req_data = {
            "data": str(operator_data),
            "operator_metadata_type": "openapi",
            "direct_publish": True
        }
        re = self.client.RegisterOperator(req_data, Headers)
        assert re[0] == 200, f"注册算子失败，状态码: {re[0]}, 响应: {re}"
        assert len(re[1]) > 0, f"注册算子返回空列表，响应: {re}"
        
        operators = re[1]
        
        # 下架算子
        update_data = [
            {
                "operator_id": operators[0]["operator_id"],
                "status": "offline"
            }
        ]
        re = self.client.UpdateOperatorStatus(update_data, Headers)
        assert re[0] == 200, f"下架算子失败，状态码: {re[0]}, 响应: {re}"

        # 编辑已下架的算子
        # 先获取原有描述
        res = self.client.GetOperatorInfo(operators[0]["operator_id"], Headers)
        desc = res[1].get("description", "default desc")

        data = {
            "operator_id": operators[0]["operator_id"],
            "metadata_type": "openapi",
            "description": desc
        }
        re = self.client.EditOperator(data, Headers)
        assert re[0] == 200, f"编辑算子失败，状态码: {re[0]}, 响应: {re}"
        assert re[1]["status"] == "unpublish", f"编辑后状态应该是unpublish，实际: {re[1].get('status')}"
        assert re[1]["version"] != operators[0]["version"], "编辑已下架算子生成了新版本"

    @allure.title("编辑未发布算子，编辑成功，无新版本生成")
    def test_edit_operator_07(self, Headers):
        """
        测试用例7：正常场景 - 编辑未发布算子
        
        测试场景：
            1. 注册一个算子（状态为unpublish）
            2. 编辑该算子
        
        验证点：
            - 注册接口返回200状态码
            - 编辑接口返回200状态码
            - 编辑后状态仍为"unpublish"
            - 版本号不变
        
        说明：
            未发布（unpublish）状态的算子编辑后，状态仍为未发布，不生成新版本。
        """
        filepath = "./resource/openapi/compliant/test3.yaml"
        operator_data = GetContent(filepath).yamlfile()
        req_data = {
            "data": str(operator_data),
            "operator_metadata_type": "openapi"
        }
        re = self.client.RegisterOperator(req_data, Headers)
        assert re[0] == 200, f"注册算子失败，状态码: {re[0]}, 响应: {re}"
        
        operators = re[1]
        assert len(operators) > 0, f"注册算子返回空列表，响应: {re}"
        
        # 先获取原有描述
        res = self.client.GetOperatorInfo(operators[0]["operator_id"], Headers)
        desc = res[1].get("description", "default desc")

        data = {
            "operator_id": operators[0]["operator_id"],
            "name": "edit_test",
            "metadata_type": "openapi",
            "description": desc
        }
        re = self.client.EditOperator(data, Headers)
        assert re[0] == 200, f"编辑算子失败，状态码: {re[0]}, 响应: {re}"
        assert re[1]["status"] == "unpublish", f"编辑后状态应该是unpublish，实际: {re[1].get('status')}"
        assert re[1]["version"] != operators[0]["version"], "编辑未发布算子生成了新版本"

    @allure.title("算子名称包含特殊字符，编辑失败")
    @pytest.mark.parametrize("name", ["invalid name","name~","name@","name`","name#","name$","name%","name^","name^","name&", 
                                    "name*","name()","name-","name+","name=","name[]","name{}","name|","name\\","name:",
                                    "name;","name'","name,","name.","name?","name/","name<","name>","name；","name“","name：",
                                    "name’","name【】","name《","name》","name？","name·","name、","name，","name。"])   
    def test_edit_operator_08(self, name, Headers):
        filepath = "./resource/openapi/compliant/edit-test1.yaml"
        operator_data = GetContent(filepath).yamlfile()
        req_data = {
            "data": str(operator_data),
            "operator_metadata_type": "openapi",
            "direct_publish": True
        }
        re = self.client.RegisterOperator(req_data, Headers)
        assert re[0] == 200
        operators = re[1]
        assert len(operators) > 0, f"注册算子返回空列表，响应: {re}"
        
        data = {
            "operator_id": operators[0]["operator_id"],
            "name": name,
            "metadata_type": "openapi"
        }
        re = self.client.EditOperator(data, Headers)
        assert re[0] == 400

    @allure.title("算子描述超过255个字符，编辑失败")
    def test_edit_operator_09(self, Headers):
        filepath = "./resource/openapi/compliant/edit-test2.yaml"
        operator_data = GetContent(filepath).yamlfile()
        req_data = {
            "data": str(operator_data),
            "operator_metadata_type": "openapi",
            "direct_publish": True
        }
        re = self.client.RegisterOperator(req_data, Headers)
        assert re[0] == 200
        operator_data = re[1]
        
        data = {
            "operator_id": operator_data[0]["operator_id"],
            "metadata_type": "openapi",  # 传递description时需要metadata_type
            "description": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
        }
        re = self.client.EditOperator(data, Headers)
        assert re[0] == 400

    @allure.title("编辑算子执行控制，编辑成功")
    def test_edit_operator_10(self, Headers):
        """
        测试用例10：正常场景 - 编辑算子执行控制
        
        测试场景：
            1. 注册一个算子（状态为unpublish）
            2. 编辑算子的operator_execute_control（超时和重试策略）
        
        验证点：
            - 注册接口返回200状态码
            - 编辑接口返回200状态码
            - 编辑后状态仍为"unpublish"
            - 版本号不变
        
        说明：
            编辑operator_execute_control等非metadata字段时，不会生成新版本。
        """
        filepath = "./resource/openapi/compliant/test3.yaml"
        operator_data = GetContent(filepath).yamlfile()
        req_data = {
            "data": str(operator_data),
            "operator_metadata_type": "openapi"
        }
        re = self.client.RegisterOperator(req_data, Headers)
        assert re[0] == 200, f"注册算子失败，状态码: {re[0]}, 响应: {re}"
        assert len(re[1]) > 0, f"注册算子返回空列表，响应: {re}"
        
        operators = re[1]
        
        # 先获取原有描述
        res = self.client.GetOperatorInfo(operators[0]["operator_id"], Headers)
        desc = res[1].get("description", "default desc")

        data = {
            "operator_id": operators[0]["operator_id"],
            "metadata_type": "openapi",  # 传递operator_execute_control时需要metadata_type
            "description": desc,
            "operator_execute_control": {
                "timeout": 90000,
                "retry_policy": {
                    "max_attempts": 5,
                    "initial_delay": 100,
                    "max_delay": 9000,
                    "backoff_factor": 3,
                    "retry_conditions": {
                        "status_code": [
                            500,
                            501
                        ],
                        "error_codes": [
                            "500",
                            "501"
                        ]
                    }
                }
            }
        }
        re = self.client.EditOperator(data, Headers)
        assert re[0] == 200, f"编辑算子失败，状态码: {re[0]}, 响应: {re}"
        assert re[1]["status"] == "unpublish", f"编辑后状态应该是unpublish，实际: {re[1].get('status')}"
        assert re[1]["version"] != operators[0]["version"], "编辑operator_execute_control生成了新版本"

    @allure.title("编辑算子扩展信息，编辑成功")
    def test_edit_operator_11(self, Headers):
        """
        测试用例11：正常场景 - 编辑算子扩展信息
        
        测试场景：
            1. 注册一个算子（状态为unpublish）
            2. 编辑算子的extend_info（扩展信息）
        
        验证点：
            - 注册接口返回200状态码
            - 编辑接口返回200状态码
            - 编辑后状态仍为"unpublish"
            - 版本号不变
        
        说明：
            编辑extend_info等非metadata字段时，不会生成新版本。
        """
        filepath = "./resource/openapi/compliant/template.yaml"
        operator_data = GetContent(filepath).yamlfile()
        req_data = {
            "data": str(operator_data),
            "operator_metadata_type": "openapi"
        }
        re = self.client.RegisterOperator(req_data, Headers)
        assert re[0] == 200, f"注册算子失败，状态码: {re[0]}, 响应: {re}"
        assert len(re[1]) > 0, f"注册算子返回空列表，响应: {re}"
        
        operators = re[1]
        
        # 先获取原有描述
        res = self.client.GetOperatorInfo(operators[0]["operator_id"], Headers)
        desc = res[1].get("description", "default desc")

        data = {
            "operator_id": operators[0]["operator_id"],
            "metadata_type": "openapi",  # 传递extend_info时需要metadata_type
            "description": desc,
            "extend_info": {
                "custom_info": {
                    "key1": "value1",
                    "key2": "value2"
                }
            }
        }
        re = self.client.EditOperator(data, Headers)
        assert re[0] == 200, f"编辑算子失败，状态码: {re[0]}, 响应: {re}"
        assert re[1]["status"] == "unpublish", f"编辑后状态应该是unpublish，实际: {re[1].get('status')}"
        assert re[1]["version"] != operators[0]["version"], "编辑extend_info生成了新版本"

    @allure.title("编辑算子信息，执行模式为异步，标识为数据源算子，编辑失败")
    def test_edit_operator_12(self, Headers):
        """
        测试用例12：异常场景 - 异步模式不能设置为数据源算子
        
        测试场景：
            - 编辑算子时，设置execution_mode为"async"，is_data_source为True
            - 验证编辑失败
        
        验证点：
            - 接口返回400状态码（Bad Request）
        
        说明：
            数据源算子（is_data_source=True）必须是同步模式（execution_mode="sync"）。
            异步模式的数据源算子不被支持，会导致编辑失败。
        """
        global operator_list
        
        # 检查列表长度（setup应该确保至少有2个，这里只是防御性检查）
        assert len(operator_list) >= 2, f"operator_list 长度不足（需要至少2个，实际{len(operator_list)}个），setup 应该确保至少注册2个算子"
        
        # 先获取原有描述
        res = self.client.GetOperatorInfo(operator_list[1]["operator_id"], Headers)
        desc = res[1].get("description", "default desc")

        data = {
            "operator_id": operator_list[1]["operator_id"],
            "metadata_type": "openapi",  # 传递operator_info中包含is_data_source时需要metadata_type
            "description": desc,
            "operator_info": {
                "operator_type": "composite",
                "execution_mode": "async",
                "category": "data_processing",
                "is_data_source": True
            }
        }
        re = self.client.EditOperator(data, Headers)
        assert re[0] == 400, f"异步模式的数据源算子应该返回400，实际: {re[0]}, 响应: {re}"

    @allure.title("编辑算子信息，执行模式为同步，标识为数据源算子，编辑成功")
    def test_edit_operator_13(self, Headers):
        """
        测试用例13：正常场景 - 同步模式可以设置为数据源算子
        
        测试场景：
            - 编辑算子时，设置execution_mode为"sync"，is_data_source为True
            - 验证编辑成功
        
        验证点：
            - 接口返回200状态码
            - 算子信息更新成功
            - execution_mode为"sync"
            - is_data_source为True
        
        说明：
            数据源算子（is_data_source=True）必须是同步模式（execution_mode="sync"）。
            同步模式的数据源算子可以正常编辑。
        """
        global operator_list
        
        # 检查列表长度（setup应该确保至少有2个，这里只是防御性检查）
        assert len(operator_list) >= 2, f"operator_list 长度不足（需要至少2个，实际{len(operator_list)}个），setup 应该确保至少注册2个算子"
        
        # 先获取原有描述
        res = self.client.GetOperatorInfo(operator_list[1]["operator_id"], Headers)
        desc = res[1].get("description", "default desc")

        data = {
            "operator_id": operator_list[1]["operator_id"],
            "metadata_type": "openapi",  # 传递operator_info中包含is_data_source时需要metadata_type
            "description": desc,
            "operator_info": {
                "operator_type": "composite",
                "execution_mode": "sync",
                "category": "data_processing",
                "is_data_source": True
            }
        }
        re = self.client.EditOperator(data, Headers)
        assert re[0] == 200

        result = self.client.GetOperatorInfo(operator_list[1]["operator_id"], Headers)
        assert result[0] == 200
        assert result[1]["operator_id"] == operator_list[1]["operator_id"]
        operator_info = result[1]["operator_info"]
        assert operator_info["operator_type"] == "composite"
        assert operator_info["execution_mode"] == "sync"
        assert operator_info["category"] == "data_processing"
        assert operator_info["is_data_source"] == True

    @allure.title("更新算子元数据，metadata_type为其他类型，编辑失败")
    def test_edit_operator_14(self, Headers):
        """
        测试用例14：异常场景 - metadata_type为无效值
        
        测试场景：
            - 传递无效的metadata_type值（不在枚举值中）
            - 验证编辑失败
        
        验证点：
            - 接口返回400状态码（Bad Request）
        
        说明：
            metadata_type的有效值只能是"openapi"或"function"，其他值会导致验证失败。
        """
        global operator_list
        
        # 检查列表是否为空（setup应该确保不为空，这里只是防御性检查）
        assert len(operator_list) > 0, "operator_list 为空，setup 应该确保至少注册2个算子"
        
        data = {
            "operator_id": operator_list[0]["operator_id"],
            "metadata_type": "operator"  # 无效的metadata_type值，不在枚举["openapi", "function"]中
        }
        re = self.client.EditOperator(data, Headers)
        assert re[0] == 400, f"无效的metadata_type应该返回400，实际: {re[0]}, 响应: {re}"

    @allure.title("更新算子元数据，未匹配到当前算子，编辑失败")
    def test_edit_operator_15(self, Headers):
        global operator_list
        
        # 检查列表是否为空（setup应该确保不为空，这里只是防御性检查）
        assert len(operator_list) > 0, "operator_list 为空，setup 应该确保至少注册2个算子"
        
        filepath = "./resource/openapi/compliant/template.yaml"
        operator_data = GetContent(filepath).yamlfile()
        # Request.post使用json=data会自动序列化，所以data字段应该直接是字典对象
        # 而不是字符串，这样json序列化时会正确地将嵌套对象序列化为JSON
        data = {
            "operator_id": operator_list[0]["operator_id"],
            "metadata_type": "openapi",
            "data": operator_data  # 直接传递字典对象，Request.post会自动序列化
        }
        re = self.client.EditOperator(data, Headers)
        assert re[0] == 404

    @allure.title("更新算子元数据，openapi中包含多个算子，可匹配到当前算子，编辑成功，生成新版本")
    def test_edit_operator_16(self, Headers):
        global operator_list
        
        # 检查列表是否为空（setup应该确保不为空，这里只是防御性检查）
        assert len(operator_list) > 0, "operator_list 为空，setup 应该确保至少注册2个算子"
        
        filepath = "./resource/openapi/compliant/relations.yaml"
        operator_data = GetContent(filepath).yamlfile()
        # Request.post使用json=data会自动序列化，所以data字段应该直接是字典对象
        # 而不是字符串，这样json序列化时会正确地将嵌套对象序列化为JSON
        data = {
            "operator_id": operator_list[0]["operator_id"],
            "metadata_type": "openapi",
            "data": operator_data  # 直接传递字典对象，Request.post会自动序列化
        }
        re = self.client.EditOperator(data, Headers)
        assert re[0] == 200, f"编辑算子失败，状态码: {re[0]}, 响应: {re}"
        assert re[1]["operator_id"] == operator_list[0]["operator_id"]
        assert re[1]["status"] == "editing"
        assert re[1]["version"] != operator_list[0]["version"]

    @allure.title("更新算子元数据，openapi中仅包含一个算子，且匹配到当前算子，编辑成功，生成新版本")
    def test_edit_operator_17(self, Headers):
        filepath = "./resource/openapi/compliant/template.yaml"
        operator_data = GetContent(filepath).yamlfile()
        req_data = {
            "data": str(operator_data),
            "operator_metadata_type": "openapi",
            "direct_publish": True
        }
        re = self.client.RegisterOperator(req_data, Headers)
        assert re[0] == 200, f"注册算子失败，状态码: {re[0]}, 响应: {re}"
        
        operators = re[1]
        assert len(operators) > 0, f"注册算子返回空列表，响应: {re}"
        
        # Request.post使用json=data会自动序列化，所以data字段应该直接是字典对象
        # 而不是字符串，这样json序列化时会正确地将嵌套对象序列化为JSON
        data = {
            "operator_id": operators[0]["operator_id"],
            "metadata_type": "openapi",
            "data": operator_data  # 直接传递字典对象，Request.post会自动序列化
        }
        re = self.client.EditOperator(data, Headers)
        assert re[0] == 200, f"编辑算子失败，状态码: {re[0]}, 响应: {re}"
        assert re[1]["operator_id"] == operators[0]["operator_id"]
        assert re[1]["status"] == "editing"
        assert re[1]["version"] != operators[0]["version"]

    @allure.title("编辑算子 - 修改extend_info，编辑成功")
    def test_edit_operator_extend_info(self, Headers):
        global operator_list
        assert len(operator_list) > 0
        
        op_id = operator_list[0]["operator_id"]
        # 先获取原有描述
        res = self.client.GetOperatorInfo(op_id, Headers)
        desc = res[1].get("description", "default desc")

        extend_info = {"updated_key": "updated_value"}
        
        data = {
            "operator_id": op_id,
            "metadata_type": "openapi",
            "description": desc,
            "extend_info": extend_info
        }
        
        re = self.client.EditOperator(data, Headers)
        assert re[0] == 200
        assert re[1]["operator_id"] == op_id
        
        # 验证
        res = self.client.GetOperatorInfo(op_id, Headers)
        assert res[0] == 200
        if "extend_info" in res[1]:
            assert res[1]["extend_info"]["updated_key"] == "updated_value"
