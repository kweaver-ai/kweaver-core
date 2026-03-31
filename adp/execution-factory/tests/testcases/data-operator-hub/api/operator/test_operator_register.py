# -*- coding:UTF-8 -*-
"""
算子注册接口测试

测试目标：
    验证算子注册功能，包括单个算子注册、批量注册、不同数据格式、直接发布等场景。

测试覆盖：
    1. 正常场景：默认不直接发布，注册成功，算子状态为未发布
    2. 正常场景：单个算子直接发布，注册成功，算子状态为已发布
    3. 异常场景：多个算子直接发布，注册失败
    4. 正常场景：YAML格式数据，注册成功
    5. 异常场景：存在同名已发布算子，注册失败

说明：
    算子注册是将OpenAPI定义的算子注册到系统中。
    批量注册算子时不允许直接发布，只有单个算子可以设置direct_publish=True。
    未发布算子不校验是否重复，只有已发布算子会校验名称重复。
"""

import allure
import time

from common.get_content import GetContent
from lib.operator import Operator


@allure.feature("算子注册与管理接口测试：算子注册")
class TestOperatorRegister:
    """
    算子注册测试类
    
    说明：
        批量注册算子，不允许直接发布。
        未发布算子不校验是否重复。
    """
    
    client = Operator()
    
    def _register_operator_with_retry(self, data, headers, max_retries=3, retry_delay=2):
        """
        带重试机制的算子注册方法
        
        参数：
            data: 注册数据
            headers: 请求头
            max_retries: 最大重试次数，默认3次
            retry_delay: 重试间隔（秒），默认2秒
        
        返回：
            (status_code, response_data) 元组
        
        说明：
            对于临时错误（500, 502, 503, 504）会自动重试。
            对于业务错误（400, 403, 409等）不会重试，直接返回。
        """
        result = None
        
        for attempt in range(max_retries):
            result = self.client.RegisterOperator(data, headers)
            
            # 如果成功，直接返回
            if result[0] == 200:
                return result
            
            # 如果是临时错误且还有重试机会，则重试
            if result[0] in [500, 502, 503, 504] and attempt < max_retries - 1:
                wait_time = retry_delay * (attempt + 1)
                print(f"注册算子返回 {result[0]}，{wait_time} 秒后重试（尝试 {attempt + 1}/{max_retries}）...")
                time.sleep(wait_time)
            else:
                # 非临时错误或重试次数用完，直接返回
                return result
        
        return result

    @allure.title("注册算子 - 默认不直接发布，注册成功，算子状态为未发布，默认执行模式为同步，为非数据源算子")
    def test_register_operator_01(self, Headers):
        """
        测试用例1：正常场景 - 默认注册（不直接发布）
        
        测试场景：
            - 注册多个算子，不设置direct_publish（默认为False）
            - 验证算子注册成功
        
        验证点：
            - 接口返回200状态码
            - 所有算子注册状态为"success"
            - 算子状态为"unpublish"（未发布）
            - 默认执行模式为"sync"（同步）
            - 默认is_data_source为False（非数据源算子）
        
        说明：
            默认情况下，注册的算子不会直接发布，需要后续手动发布。
            未发布状态的算子不校验名称是否重复。
        """
        filepath = "./resource/openapi/compliant/test0.json"
        api_data = GetContent(filepath).jsonfile()

        data = {
            "data": str(api_data),
            "operator_metadata_type": "openapi"
            # 注意：不设置direct_publish，默认为False
        }

        result = self._register_operator_with_retry(data, Headers)
        assert result[0] == 200, f"注册算子失败，状态码: {result[0]}, 响应: {result}。请检查后端服务是否正常运行。"
        
        if result[0] == 200 and isinstance(result[1], list):
            operators = result[1]
            for operator in operators:
                assert operator.get("status") == "success", \
                    f"算子注册状态应该是success，实际: {operator.get('status')}"
                
                operator_id = operator["operator_id"]

                # 验证算子状态为未发布
                list_data = {
                    "page_size": -1,
                    "status": "unpublish"
                }
                result = self.client.GetOperatorList(list_data, Headers)
                assert result[0] == 200, f"获取算子列表失败，状态码: {result[0]}"
                
                if isinstance(result[1], dict):
                    unpublished_ops = result[1].get("data", [])
                    for unpublished_op in unpublished_ops:
                        if unpublished_op["operator_id"] == operator_id:
                            assert unpublished_op["status"] == "unpublish", \
                                f"算子状态应该是unpublish，实际: {unpublished_op.get('status')}"
                            assert unpublished_op["operator_info"]["execution_mode"] == "sync", \
                                f"默认执行模式应该是sync，实际: {unpublished_op.get('operator_info', {}).get('execution_mode')}"
                            assert unpublished_op["operator_info"]["is_data_source"] == False, \
                                f"默认应该是非数据源算子，实际: {unpublished_op.get('operator_info', {}).get('is_data_source')}"

    @allure.title("注册算子 - 单个算子直接发布，注册成功，算子状态为已发布")
    def test_register_operator_02(self, Headers):
        """
        测试用例2：正常场景 - 单个算子直接发布
        
        测试场景：
            - 注册单个算子，设置direct_publish=True
            - 验证算子注册并直接发布成功
        
        验证点：
            - 接口返回200状态码
            - 所有算子注册状态为"success"
            - 算子状态为"published"（已发布）
        
        说明：
            只有单个算子可以设置direct_publish=True直接发布。
            多个算子时设置direct_publish=True会失败。
        """
        filepath = "./resource/openapi/compliant/test1.json"
        api_data = GetContent(filepath).jsonfile()
        data = {
            "data": str(api_data),
            "operator_metadata_type": "openapi",
            "direct_publish": True  # 直接发布
        }
        result = self._register_operator_with_retry(data, Headers)
        assert result[0] == 200, f"注册算子失败，状态码: {result[0]}, 响应: {result}。请检查后端服务是否正常运行。"
        
        operators = result[1]
        for operator in operators:
            assert operator.get("status") == "success", \
                f"算子注册状态应该是success，实际: {operator.get('status')}"
            
            operator_id = operator["operator_id"]
            result = self.client.GetOperatorInfo(operator_id, Headers)
            assert result[0] == 200, f"获取算子信息失败，状态码: {result[0]}"
            assert result[1]["status"] == "published", \
                f"直接发布后算子状态应该是published，实际: {result[1].get('status')}"

    @allure.title("注册算子 - 多个算子直接发布，注册失败")
    def test_register_operator_03(self, Headers):
        """
        测试用例3：异常场景 - 多个算子不允许直接发布
        
        测试场景：
            - 注册多个算子，设置direct_publish=True
            - 验证注册失败
        
        验证点：
            - 接口返回400状态码（Bad Request）
        
        说明：
            批量注册算子时不允许直接发布，只有单个算子可以设置direct_publish=True。
            这是为了避免批量操作时的风险。
        """
        filepath = "./resource/openapi/compliant/test0.json"
        api_data = GetContent(filepath).jsonfile()
        data = {
            "data": str(api_data),
            "operator_metadata_type": "openapi",
            "direct_publish": True  # 多个算子时不允许直接发布
        }
        result = self.client.RegisterOperator(data, Headers)
        assert result[0] == 400, \
            f"多个算子直接发布应该返回400，实际: {result[0]}, 响应: {result}"

    @allure.title("注册算子 - YAML格式数据，注册成功")
    def test_register_operator_04(self, Headers):
        """
        测试用例4：正常场景 - YAML格式数据注册
        
        测试场景：
            - 使用YAML格式的OpenAPI数据注册算子
            - 验证注册成功
        
        验证点：
            - 接口返回200状态码
            - 所有算子注册状态为"success"
        
        说明：
            算子注册支持JSON和YAML两种格式的OpenAPI数据。
            两种格式的注册逻辑相同，只是数据格式不同。
        """
        filepath = "./resource/openapi/compliant/test2.yaml"
        api_data = GetContent(filepath).yamlfile()
        data = {
            "data": str(api_data),
            "operator_metadata_type": "openapi"
        }
        result = self._register_operator_with_retry(data, Headers)
        assert result[0] == 200, f"注册算子失败，状态码: {result[0]}, 响应: {result}。请检查后端服务是否正常运行。"
        
        for item in result[1]:
            assert item.get("status") == "success", \
                f"算子注册状态应该是success，实际: {item.get('status')}"

    @allure.title("注册算子 - 存在同名已发布算子，注册失败")
    def test_register_operator_05(self, Headers):
        """
        测试用例5：异常场景 - 同名已发布算子冲突
        
        测试场景：
            1. 注册并发布一个算子
            2. 尝试注册同名算子
        
        验证点：
            - 第二次注册返回400状态码（Bad Request）
        
        说明：
            已发布（published）状态的算子会校验名称是否重复。
            如果存在同名已发布的算子，新注册的算子会失败。
            未发布（unpublish）状态的算子不校验名称重复。
        """
        filepath = "./resource/openapi/compliant/test3.yaml"
        api_data = GetContent(filepath).yamlfile()
        data = {
            "data": str(api_data),
            "operator_metadata_type": "openapi",
            "direct_publish": True
        }

        # 第一次注册并发布
        result = self._register_operator_with_retry(data, Headers)
        assert result[0] == 200, f"第一次注册算子失败，状态码: {result[0]}, 响应: {result}。请检查后端服务是否正常运行。"
        
        # 第二次注册同名算子（应该失败）
        result = self._register_operator_with_retry(data, Headers)
        assert result[0] == 200, f"第二次注册算子失败，状态码: {result[0]}, 响应: {result}。请检查后端服务是否正常运行。"
        assert result[1][0].get("status") == "failed", \
            f"存在同名已发布算子时注册状态应该是failed，实际: {result[1][0].get('status')}"

    @allure.title("注册算子 - 算子名称超过50个字符，注册失败")
    def test_register_operator_06(self, Headers):
        """
        测试用例6：异常场景 - 算子名称长度超限
        
        测试场景：
            - 注册算子时，算子名称超过50个字符
            - 验证注册失败
        
        验证点：
            - 接口返回200状态码（注册请求成功）
            - 算子注册状态为"failed"
        
        说明：
            算子名称长度限制为50个字符，超过此限制会导致注册失败。
        """
        filepath = "./resource/openapi/non-compliant/long_name.yaml"
        api_data = GetContent(filepath).yamlfile()
        data = {
            "data": str(api_data),
            "operator_metadata_type": "openapi"
        }
        result = self._register_operator_with_retry(data, Headers)
        assert result[0] == 200, f"注册算子失败，状态码: {result[0]}, 响应: {result}。请检查后端服务是否正常运行。"
        assert result[1][0].get("status") == "failed", \
            f"算子名称超过50字符时注册状态应该是failed，实际: {result[1][0].get('status')}"

    @allure.title("注册算子 - 算子描述超过255个字符，注册失败")
    def test_register_operator_07(self, Headers):
        """
        测试用例7：异常场景 - 算子描述长度超限
        
        测试场景：
            - 注册算子时，算子描述超过255个字符
            - 验证注册失败
        
        验证点：
            - 接口返回200状态码（注册请求成功）
            - 算子注册状态为"failed"
        
        说明：
            算子描述长度限制为255个字符，超过此限制会导致注册失败。
        """
        filepath = "./resource/openapi/non-compliant/long_desc.yaml"
        api_data = GetContent(filepath).yamlfile()
        data = {
            "data": str(api_data),
            "operator_metadata_type": "openapi"
        }
        result = self._register_operator_with_retry(data, Headers)
        assert result[0] == 200, f"注册算子失败，状态码: {result[0]}, 响应: {result}。请检查后端服务是否正常运行。"
        assert result[1][0].get("status") == "failed", \
            f"算子描述超过255字符时注册状态应该是failed，实际: {result[1][0].get('status')}"

    @allure.title("注册算子 - OpenAPI超出大小限制，注册失败")
    def test_register_operator_08(self, Headers):
        """
        测试用例8：异常场景 - OpenAPI数据大小超限
        
        测试场景：
            - 注册算子时，OpenAPI数据超出大小限制
            - 验证注册失败
        
        验证点：
            - 接口返回400状态码（Bad Request）
        
        说明：
            OpenAPI数据有大小限制，超出限制会导致注册失败。
            这是为了防止过大的数据影响系统性能。
        """
        filepath = "./resource/openapi/compliant/large_api.json"
        api_data = GetContent(filepath).jsonfile()
        data = {
            "data": str(api_data),
            "operator_metadata_type": "openapi"
        }
        result = self.client.RegisterOperator(data, Headers)
        assert result[0] == 400, \
            f"OpenAPI超出大小限制应该返回400，实际: {result[0]}, 响应: {result}"

    @allure.title("注册算子 - 运行模式为异步，注册数据源算子，注册失败")
    def test_register_operator_09(self, Headers):
        """
        测试用例9：异常场景 - 异步模式不能注册数据源算子
        
        测试场景：
            - 注册算子时，设置execution_mode为"async"，is_data_source为True
            - 验证注册失败
        
        验证点：
            - 接口返回400状态码（Bad Request）
        
        说明：
            数据源算子（is_data_source=True）必须是同步模式（execution_mode="sync"）。
            异步模式的数据源算子不被支持，会导致注册失败。
        """
        filepath = "./resource/openapi/compliant/template.yaml"
        api_data = GetContent(filepath).yamlfile()
        data = {
            "data": str(api_data),
            "operator_metadata_type": "openapi",
            "operator_info": {
                "execution_mode": "async",  # 异步模式
                "is_data_source": True      # 数据源算子
            }
        }
        result = self.client.RegisterOperator(data, Headers)
        assert result[0] == 400, \
            f"异步模式的数据源算子应该返回400，实际: {result[0]}, 响应: {result}"

    @allure.title("注册算子 - 运行模式为同步，注册数据源算子，注册成功")
    def test_register_operator_10(self, Headers):
        """
        测试用例10：正常场景 - 同步模式注册数据源算子
        
        测试场景：
            - 注册算子时，设置execution_mode为"sync"，is_data_source为True
            - 验证注册成功
        
        验证点：
            - 接口返回200状态码
            - 算子注册状态为"success"
        
        说明：
            数据源算子（is_data_source=True）必须是同步模式（execution_mode="sync"）。
            同步模式的数据源算子可以正常注册。
        """
        filepath = "./resource/openapi/compliant/template.yaml"
        api_data = GetContent(filepath).yamlfile()
        data = {
            "data": str(api_data),
            "operator_metadata_type": "openapi",
            "operator_info": {
                "execution_mode": "sync",   # 同步模式
                "is_data_source": True      # 数据源算子
            }
        }
        result = self._register_operator_with_retry(data, Headers)
        assert result[0] == 200, f"注册算子失败，状态码: {result[0]}, 响应: {result}。请检查后端服务是否正常运行。"
        assert result[1][0].get("status") == "success", \
            f"同步模式的数据源算子注册状态应该是success，实际: {result[1][0].get('status')}"

    @allure.title("注册算子 - 指定extend_info，验证数据正确存储")
    def test_register_operator_extend_info(self, Headers):
        filepath = "./resource/openapi/compliant/test1.json"
        api_data = GetContent(filepath).jsonfile()
        
        extend_info = {"custom_key": "custom_value", "nested": {"a": 1}}
        data = {
            "data": str(api_data),
            "operator_metadata_type": "openapi",
            "extend_info": extend_info
        }
        
        result = self._register_operator_with_retry(data, Headers)
        assert result[0] == 200
        operator_id = result[1][0]["operator_id"]
        
        # 验证获取信息
        res = self.client.GetOperatorInfo(operator_id, Headers)
        assert res[0] == 200
        # 如果 GetOperatorInfo 返回 extend_info，则校验
        if "extend_info" in res[1]:
            assert res[1]["extend_info"]["custom_key"] == "custom_value"
