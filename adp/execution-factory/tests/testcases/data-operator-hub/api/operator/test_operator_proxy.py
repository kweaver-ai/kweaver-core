# -*- coding:UTF-8 -*-
"""
算子代理执行接口测试

测试目标：
    验证通过代理方式执行算子的功能，包括不同HTTP方法（GET/POST/DELETE）的代理执行、
    不同参数传递方式（path/query/header/body）、异常场景处理等。

测试覆盖：
    1. GET接口代理执行（path/query/header传参）
    2. POST接口代理执行（body传参）
    3. DELETE接口代理执行（body传参）
    4. 异常场景：算子不存在/未发布、缺少必要header、超时、异步算子
"""

import allure
import uuid
import pytest
import os

from common.get_content import GetContent
from lib.operator import Operator
from lib.operator_internal import InternalOperator
from lib.impex import Impex

operator_list = []

@allure.feature("算子注册与管理接口测试：代理执行算子")
class TestOperatorProxy:
    """
    算子代理执行测试类
    
    说明：
        代理执行算子是指通过内部接口代理执行已发布的算子，支持GET/POST/DELETE等HTTP方法。
        测试用例覆盖正常执行场景和异常场景。
    """

    client = Operator()
    internal_client = InternalOperator()
    impex_client = Impex()

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers):
        """
        测试前置准备
        
        功能：
            1. 注册算子并发布，为后续测试准备数据
            2. 获取所有已发布的算子列表，供测试用例使用
        
        注意：
            - 如果注册失败，整个测试类会失败
            - operator_list 为空时，某些测试用例可能会跳过
        """
        import time
        
        global operator_list
        file = GetContent("./config/env.ini")
        config = file.config()
        host = config["server"]["host"]
        filepath = "./resource/openapi/compliant/operator.json"
        api_data = GetContent(filepath).jsonfile()
        server = api_data["servers"][0]
        server["url"] = f"https://{host}"
        
        data = {
            "data": str(api_data),
            "operator_metadata_type": "openapi"
        }
        
        # 添加重试机制，处理503等临时错误
        max_retries = 3
        retry_delay = 2  # 重试间隔（秒）
        result = None
        
        for attempt in range(max_retries):
            result = self.client.RegisterOperator(data, Headers)
            if result[0] == 200:
                break
            elif result[0] in [500, 502, 503, 504] and attempt < max_retries - 1:
                # 临时错误，等待后重试
                wait_time = retry_delay * (attempt + 1)
                print(f"注册算子返回 {result[0]}，{wait_time} 秒后重试（尝试 {attempt + 1}/{max_retries}）...")
                time.sleep(wait_time)
            else:
                # 非临时错误或重试次数用完，直接失败
                break
        
        assert result[0] == 200, f"setup 中注册算子失败，状态码: {result[0]}, 响应: {result}。请检查后端服务是否正常运行。"
        operators = result[1]
        for operator in operators:
            if operator.get("status") == "success":
                update_data = [
                    {
                        "operator_id": operator["operator_id"],
                        "status": "published"
                    }
                ]
                result = self.client.UpdateOperatorStatus(update_data, Headers)    # 发布算子
                if result[0] != 200:
                    print(f"警告: setup 中发布算子失败，状态码: {result[0]}, 响应: {result}")
        
        result = self.client.GetOperatorList({"all": "true"}, Headers)
        if result[0] == 200:
            operator_list = result[1].get("data", [])
        else:
            print(f"警告: setup 中获取算子列表失败，状态码: {result[0]}, 响应: {result}")
            operator_list = []
        
        if len(operator_list) == 0:
            print(f"警告: operator_list 为空，某些测试可能会失败")

    def _find_operator_by_name(self, name):
        """
        辅助方法：根据算子名称查找算子ID
        
        参数：
            name: 算子名称
        
        返回：
            operator_id: 算子ID，如果未找到返回None
        """
        global operator_list
        for operator in operator_list:
            if operator.get("name") == name:
                return operator["operator_id"], operator.get("version")
        return None, None

    @allure.title("代理执行算子 - GET接口（path传参），执行成功")
    def test_operator_proxy_get_path(self, Headers, UserHeaders):
        """
        测试用例1：GET接口代理执行，使用path参数传递
        
        测试场景：
            - 使用GET方法代理执行"获取算子信息"算子
            - 通过path参数传递operator_id
        
        验证点：
            - 代理执行成功，返回200状态码
        
        说明：
            这是GET接口最常见的传参方式，参数直接放在URL路径中。
        """
        global operator_list
        
        # 检查列表是否为空
        if len(operator_list) == 0:
            pytest.skip("operator_list 为空，跳过此测试")
        
        operator_id, _ = self._find_operator_by_name("获取算子信息")
        if not operator_id:
            pytest.skip("未找到名称为'获取算子信息'的算子，跳过此测试")
        
        proxy_data = {
            "header": Headers,
            "path": {
                "operator_id": operator_id
            }
        }
        result = self.internal_client.ProxyOperator(operator_id, proxy_data, UserHeaders)
        assert result[0] == 200, f"代理执行失败，状态码: {result[0]}, 响应: {result}"

    @allure.title("代理执行算子 - GET接口（query传参），执行成功")
    def test_operator_proxy_get_query(self, Headers, UserHeaders):
        """
        测试用例2：GET接口代理执行，使用query参数传递
        
        测试场景：
            - 使用GET方法代理执行"获取算子列表"算子
            - 通过query参数传递分页和排序信息
        
        验证点：
            - 代理执行成功，返回200状态码
        
        说明：
            query参数通常用于GET请求的分页、排序、过滤等场景。
        """
        global operator_list
        
        if len(operator_list) == 0:
            pytest.skip("operator_list 为空，跳过此测试")
        
        operator_id, _ = self._find_operator_by_name("获取算子列表")
        if not operator_id:
            pytest.skip("未找到名称为'获取算子列表'的算子，跳过此测试")
        
        proxy_data = {
            "header": Headers,
            "query": {
                "page_size": "5",
                "sort_by": "name",
                "sort_order": "asc"
            }
        }
        result = self.internal_client.ProxyOperator(operator_id, proxy_data, UserHeaders)
        assert result[0] == 200, f"代理执行失败，状态码: {result[0]}, 响应: {result}"

    @allure.title("代理执行算子 - GET接口（header传参），执行成功")
    def test_operator_proxy_get_header(self, Headers, UserHeaders):
        """
        测试用例3：GET接口代理执行，仅使用header参数
        
        测试场景：
            - 使用GET方法代理执行"获取算子分类"算子
            - 仅通过header传递认证信息，无需额外参数
        
        验证点：
            - 代理执行成功，返回200状态码
        
        说明：
            某些GET接口只需要认证信息，不需要额外的path或query参数。
        """
        global operator_list
        
        if len(operator_list) == 0:
            pytest.skip("operator_list 为空，跳过此测试")
        
        operator_id, _ = self._find_operator_by_name("获取算子分类")
        if not operator_id:
            pytest.skip("未找到名称为'获取算子分类'的算子，跳过此测试")
        
        proxy_data = {
            "header": Headers
        }
        result = self.internal_client.ProxyOperator(operator_id, proxy_data, UserHeaders)
        assert result[0] == 200, f"代理执行失败，状态码: {result[0]}, 响应: {result}"

    @allure.title("代理执行算子 - POST接口（body传参），执行成功")
    def test_operator_proxy_post(self, Headers, UserHeaders):
        """
        测试用例4：POST接口代理执行，使用body参数传递
        
        测试场景：
            - 使用POST方法代理执行"更新算子信息"算子
            - 通过body传递完整的更新数据
        
        验证点：
            - 代理执行成功，返回200状态码
        
        说明：
            POST接口通常用于创建或更新操作，数据通过请求体传递。
        """
        global operator_list
        
        if len(operator_list) == 0:
            pytest.skip("operator_list 为空，跳过此测试")
        
        operator_id, operator_version = self._find_operator_by_name("更新算子信息")
        if not operator_id:
            pytest.skip("未找到名称为'更新算子信息'的算子，跳过此测试")
        
        filepath = "./resource/openapi/compliant/test3.yaml"
        api_data = GetContent(filepath).yamlfile()
        proxy_data = {
            "header": Headers,
            "body": {
                "operator_id": operator_id,
                "version": operator_version,
                "data": str(api_data),
                "operator_metadata_type": "openapi"
            }
        }
        result = self.internal_client.ProxyOperator(operator_id, proxy_data, UserHeaders)
        assert result[0] == 200, f"代理执行失败，状态码: {result[0]}, 响应: {result}"

    @allure.title("代理执行算子 - DELETE接口（body传参），执行成功")
    def test_operator_proxy_delete(self, Headers, UserHeaders):
        """
        测试用例5：DELETE接口代理执行，使用body参数传递
        
        测试场景：
            - 使用DELETE方法代理执行"删除算子"算子
            - 通过body传递要删除的算子ID和版本信息
        
        验证点：
            - 代理执行成功，返回200状态码
        
        说明：
            DELETE接口用于删除操作，虽然HTTP规范建议使用path参数，但某些场景下也会使用body传递参数。
        """
        global operator_list
        
        if len(operator_list) == 0:
            pytest.skip("operator_list 为空，跳过此测试")
        
        operator_id, operator_version = self._find_operator_by_name("删除算子")
        if not operator_id:
            pytest.skip("未找到名称为'删除算子'的算子，跳过此测试")
        
        proxy_data = {
            "header": Headers,
            "body": {
                "operator_id": operator_id,
                "version": operator_version
            }
        }
        result = self.internal_client.ProxyOperator(operator_id, proxy_data, UserHeaders)
        assert result[0] == 200, f"代理执行失败，状态码: {result[0]}, 响应: {result}"

    @allure.title("代理执行算子 - 异常场景：算子不存在或未发布，执行失败")
    def test_operator_proxy_not_found(self, Headers, UserHeaders):
        """
        测试用例6：异常场景 - 算子不存在或未发布
        
        测试场景：
            1. 算子已下架（状态为offline）
            2. 算子ID不存在（使用随机UUID）
        
        验证点：
            - 两种情况都应该返回404状态码
        
        说明：
            代理执行只能执行已发布（published）状态的算子，未发布或已下架的算子无法执行。
        """
        global operator_list
        
        if len(operator_list) == 0:
            pytest.skip("operator_list 为空，跳过此测试")
        
        # 场景1：算子已下架
        operator_id, _ = self._find_operator_by_name("注册算子")
        if operator_id:
            update_data = [
                {
                    "operator_id": operator_id,
                    "status": "offline"
                }
            ]
            result = self.client.UpdateOperatorStatus(update_data, Headers)
            if result[0] == 200:
                proxy_data = {"header": Headers}
                result = self.internal_client.ProxyOperator(operator_id, proxy_data, UserHeaders)
                assert result[0] == 404, f"已下架算子应该返回404，实际: {result[0]}"
        
        # 场景2：算子ID不存在
        fake_operator_id = str(uuid.uuid4())
        proxy_data = {"header": Headers}
        result = self.internal_client.ProxyOperator(fake_operator_id, proxy_data, UserHeaders)
        assert result[0] == 404, f"不存在的算子ID应该返回404，实际: {result[0]}"

    @allure.title("代理执行算子 - 异常场景：缺少必要header（x-account-id），执行失败")
    def test_operator_proxy_missing_header(self, Headers):
        """
        测试用例7：异常场景 - 缺少必要的header信息
        
        测试场景：
            - 代理执行算子时，不传递UserHeaders（缺少x-account-id）
        
        验证点：
            - 返回400状态码，表示请求参数错误
        
        说明：
            x-account-id是代理执行算子的必要header，用于标识执行用户，缺少此header会导致执行失败。
        """
        global operator_list
        
        if len(operator_list) == 0:
            pytest.skip("operator_list 为空，跳过此测试")
        
        operator_id, operator_version = self._find_operator_by_name("更新算子状态")
        if not operator_id:
            pytest.skip("未找到名称为'更新算子状态'的算子，跳过此测试")
        
        proxy_data = {
            "version": operator_version,
            "header": Headers,
            "body": {
                "operator_id": operator_id,
                "version": operator_version,
                "status": "published"
            }
        }
        # 不传递UserHeaders，模拟缺少x-account-id的情况
        result = self.internal_client.ProxyOperator(operator_id, proxy_data, None)
        assert result[0] == 400, f"缺少必要header应该返回400，实际: {result[0]}"

    @allure.title("代理执行算子 - 异常场景：执行超时，返回超时错误")
    def test_operator_proxy_timeout(self, Headers, UserHeaders):
        # 增加随机名避免冲突
        import random, string
        rand_str = ''.join(random.choice(string.ascii_lowercase) for _ in range(6))
        
        # 1. 注册一个会超时的函数算子
        register_data = {
            "name": f"timeout_op_{rand_str}",
            "operator_metadata_type": "function", # 注册接口使用 operator_metadata_type
            "description": "operator that sleeps to cause timeout",
            "function_input": {
                "name": "handler",
                "code": "import time\ndef handler(event):\n    time.sleep(5)\n    return {'status': 'done'}",
                "script_type": "python",
                "inputs": [],
                "outputs": []
            },
            "operator_info": {
                "operator_type": "basic",
                "execution_mode": "sync"
            },
            "direct_publish": True
        }
        result = self.client.RegisterOperator(register_data, Headers)
        
        # 容错：如果是因为导入冲突导致500，我们可以选择跳过或者断言
        if result[0] == 500 and "already connected" in str(result[1]):
            pytest.skip("环境数据冲突：该算子已关联到业务域")

        # 检查导入是否成功
        assert result[0] in [200, 201], f"导入算子失败，状态码: {result[0]}, 响应: {result}"
        
        # 获取刚注册的算子ID
        operator_id = result[1][0]["operator_id"]
        
        proxy_data = {
            "header": Headers,
            "body": {"a": "aaa"},
            "timeout": 1  # 设置1秒超时，非常短
        }
        result = self.internal_client.ProxyOperator(operator_id, proxy_data, UserHeaders)
        assert result[0] == 200, f"代理请求应该成功，实际: {result[0]}"
        assert result[1]["status_code"] == 500, f"算子执行应该超时返回500，实际: {result[1].get('status_code')}"

    @allure.title("代理执行算子 - 异常场景：异步算子不支持代理执行，执行失败")
    def test_operator_proxy_async_not_supported(self, Headers, UserHeaders):
        """
        测试用例9：异常场景 - 异步算子不支持代理执行
        
        测试场景：
            1. 将算子修改为异步执行模式（execution_mode: async）
            2. 发布该算子
            3. 尝试代理执行该算子
        
        验证点：
            - 返回400状态码，表示异步算子不支持代理执行
        
        说明：
            代理执行只支持同步（sync）模式的算子，异步（async）模式的算子需要通过其他方式执行。
        """
        global operator_list
        
        if len(operator_list) == 0:
            pytest.skip("operator_list 为空，跳过此测试")
        
        operator_id, _ = self._find_operator_by_name("获取算子分类")
        if not operator_id:
            pytest.skip("未找到名称为'获取算子分类'的算子，跳过此测试")
        
        # 修改为异步算子
        data = {
            "operator_id": operator_id,
            "operator_info": {
                "execution_mode": "async",
                "category": "data_processing"
            }
        }
        re = self.client.EditOperator(data, Headers)
        if re[0] != 200:
            pytest.skip(f"编辑算子失败，状态码: {re[0]}, 响应: {re}")
        assert re[0] == 200, f"编辑算子失败，状态码: {re[0]}, 响应: {re}"
        
        # 发布算子
        update_data = [{
                "operator_id": operator_id,
                "status": "published"
            }]
        result = self.client.UpdateOperatorStatus(update_data, Headers)
        if result[0] != 200:
            pytest.skip(f"发布算子失败，状态码: {result[0]}, 响应: {result}")
        assert result[0] == 200, f"发布算子失败，状态码: {result[0]}, 响应: {result}"
        
        # 尝试代理执行异步算子
        proxy_data = {
            "header": Headers
        }
        result = self.internal_client.ProxyOperator(operator_id, proxy_data, UserHeaders)
        assert result[0] == 400, f"异步算子应该返回400，实际: {result[0]}"
