# -*- coding:UTF-8 -*-

import allure
import uuid
import pytest
import time

from common.get_content import GetContent
from lib.operator import Operator

published_operator_infos = []
unpublish_operator_infos = []
offline_operator_infos = []
# 保存每个算子注册时使用的原始OpenAPI数据
operator_openapi_data = {}

@allure.feature("算子注册与管理接口测试：更新算子信息")
class TestOperatorUpdateInfo:
  
    client = Operator()
    
    def _call_with_retry(self, func, *args, max_retries=3, retry_delay=2, **kwargs):
        """
        带重试机制的通用调用方法
        
        参数：
            func: 要调用的函数
            *args: 位置参数
            max_retries: 最大重试次数，默认3次
            retry_delay: 重试间隔（秒），默认2秒
            **kwargs: 关键字参数
        
        返回：
            函数调用的结果
        
        说明：
            对于临时错误（500, 502, 503, 504）会自动重试。
            对于业务错误（400, 403, 404, 409等）不会重试，直接返回。
        """
        result = None
        
        for attempt in range(max_retries):
            result = func(*args, **kwargs)
            
            # 如果成功，直接返回
            if result[0] == 200:
                return result
            
            # 如果是临时错误且还有重试机会，则重试
            if result[0] in [500, 502, 503, 504] and attempt < max_retries - 1:
                wait_time = retry_delay * (attempt + 1)
                func_name = func.__name__ if hasattr(func, '__name__') else str(func)
                print(f"{func_name} 返回 {result[0]}，{wait_time} 秒后重试（尝试 {attempt + 1}/{max_retries}）...")
                time.sleep(wait_time)
            else:
                # 非临时错误或重试次数用完，直接返回
                return result
        
        return result
    
    def _extract_single_api_from_openapi(self, openapi_data, target_path, target_method):
        """
        从OpenAPI定义中提取单个API路径的定义
        
        参数：
            openapi_data: OpenAPI定义（字典格式）
            target_path: 目标API路径
            target_method: 目标HTTP方法（大写，如GET、POST）
        
        返回：
            只包含目标API路径的OpenAPI定义（字典格式）
        """
        import copy
        
        try:
            # 创建OpenAPI定义的副本
            single_api_openapi = copy.deepcopy(openapi_data)
            
            # 提取目标路径的定义
            paths = single_api_openapi.get("paths", {})
            if target_path not in paths:
                return None
            
            target_path_def = paths[target_path]
            
            # 如果指定了方法，只保留该方法
            if target_method and target_method.lower() in target_path_def:
                # 只保留指定的方法
                method_def = target_path_def[target_method.lower()]
                single_api_openapi["paths"] = {
                    target_path: {
                        target_method.lower(): method_def
                    }
                }
            else:
                # 如果没有指定方法或方法不存在，保留整个路径的所有方法
                single_api_openapi["paths"] = {
                    target_path: target_path_def
                }
            
            return single_api_openapi
        except Exception as e:
            print(f"提取单个API定义时出错: {e}")
            return None

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers):
        global published_operator_infos, unpublish_operator_infos, offline_operator_infos, operator_openapi_data

        filepath = "./resource/openapi/compliant/setup.json"
        operator_data = GetContent(filepath).jsonfile()
        # 保存原始OpenAPI数据字符串，用于后续更新算子信息
        original_openapi_str = str(operator_data)
        
        req_data = {
            "data": original_openapi_str,
            "operator_metadata_type": "openapi"
        }
        re = self._call_with_retry(self.client.RegisterOperator, req_data, Headers)
        assert re[0] == 200, f"setup 中注册算子失败，状态码: {re[0]}, 响应: {re}。请检查后端服务是否正常运行。"
        count = int(len(re[1]))
        
        # 计算需要发布的算子数量（三分之二）
        publish_count = int(count * 2 / 3)
        
        # 获取所有算子ID和version，并为每个算子构建只包含其API路径的OpenAPI定义
        operator_ids = []
        versions = []
        import json
        import copy
        
        # 预先获取所有API路径列表，用于备用匹配
        paths_list = list(operator_data.get("paths", {}).keys())
        
        for index, operator in enumerate(re[1]):
            operator_id = operator["operator_id"]
            operator_ids.append(operator_id)
            versions.append(operator["version"])
            
            # 获取算子详细信息，提取path和method
            operator_info_result = self._call_with_retry(self.client.GetOperatorInfo, operator_id, Headers)
            if operator_info_result[0] == 200:
                operator_info = operator_info_result[1]
                
                # 尝试多种方式获取path和method
                api_path = None
                api_method = None
                
                # 方式1: 从metadata.openapi中获取
                metadata = operator_info.get("metadata", {})
                if metadata:
                    openapi_metadata = metadata.get("openapi", {})
                    if openapi_metadata:
                        api_path = openapi_metadata.get("path")
                        api_method = openapi_metadata.get("method", "")
                
                # 方式2: 如果方式1失败，尝试从metadata直接获取（某些版本可能结构不同）
                if not api_path and metadata:
                    api_path = metadata.get("path")
                    api_method = metadata.get("method", "")
                
                # 方式3: 如果还是失败，使用备用方案：按注册顺序匹配API路径
                # 由于注册时系统会为每个API创建算子，我们可以按顺序匹配
                if not api_path and index < len(paths_list):
                    api_path = paths_list[index]
                    # 获取该路径的第一个方法
                    path_def = operator_data["paths"].get(api_path, {})
                    if path_def and isinstance(path_def, dict):
                        # 获取第一个HTTP方法（get, post, put, delete等）
                        api_method = list(path_def.keys())[0] if path_def else None
                
                if api_path:
                    api_method = api_method.upper() if api_method else None
                    # 从原始OpenAPI数据中提取单个API路径的定义
                    single_api_openapi = self._extract_single_api_from_openapi(operator_data, api_path, api_method)
                    if single_api_openapi:
                        # 转换为JSON字符串格式保存（与注册时使用的格式一致）
                        operator_openapi_data[operator_id] = json.dumps(single_api_openapi, ensure_ascii=False)
                    else:
                        # 如果提取失败，记录警告但不保存（后续测试会跳过）
                        print(f"警告: 无法为算子 {operator_id} 提取单个API定义（path: {api_path}, method: {api_method}）")
                else:
                    # 如果无法获取path，记录警告并打印调试信息
                    print(f"警告: 算子 {operator_id} 的path信息不可用（索引: {index}, 总路径数: {len(paths_list)}）")
                    # 只在调试模式下打印详细metadata结构，避免输出过多
                    if metadata:
                        metadata_keys = list(metadata.keys())
                        print(f"调试信息: metadata包含的键 = {metadata_keys}")
            else:
                # 如果获取算子信息失败，使用备用方案：按索引匹配
                if index < len(paths_list):
                    api_path = paths_list[index]
                    path_def = operator_data["paths"].get(api_path, {})
                    if path_def and isinstance(path_def, dict):
                        api_method = list(path_def.keys())[0] if path_def else None
                        api_method = api_method.upper() if api_method else None
                        single_api_openapi = self._extract_single_api_from_openapi(operator_data, api_path, api_method)
                        if single_api_openapi:
                            operator_openapi_data[operator_id] = json.dumps(single_api_openapi, ensure_ascii=False)
                        else:
                            print(f"警告: 无法获取算子 {operator_id} 的详细信息，且备用方案也失败，状态码: {operator_info_result[0]}")
                else:
                    print(f"警告: 无法获取算子 {operator_id} 的详细信息，状态码: {operator_info_result[0]}")
        
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

        result = self._call_with_retry(self.client.UpdateOperatorStatus, publish_datas, Headers)
        assert result[0] == 200, f"setup 中发布算子失败，状态码: {result[0]}, 响应: {result}。请检查后端服务是否正常运行。"
        result = self._call_with_retry(self.client.UpdateOperatorStatus, offline_datas, Headers)
        assert result[0] == 200, f"setup 中下架算子失败，状态码: {result[0]}, 响应: {result}。请检查后端服务是否正常运行。"

        data = {"status": "published"}
        re = self._call_with_retry(self.client.GetOperatorList, data, Headers)
        assert re[0] == 200, f"获取已发布算子列表失败，状态码: {re[0]}, 响应: {re}。请检查后端服务是否正常运行。"
        
        # 检查响应格式
        if isinstance(re[1], dict) and "data" in re[1]:
            for operator in re[1]["data"]:
                operator_info = {
                    "operator_id": operator["operator_id"],
                    "version": operator["version"]
                }
                published_operator_infos.append(operator_info)
        else:
            print(f"警告: 获取已发布算子列表响应格式异常: {re[1]}")

        data = {"status": "unpublish"}
        re = self._call_with_retry(self.client.GetOperatorList, data, Headers)
        assert re[0] == 200, f"获取未发布算子列表失败，状态码: {re[0]}, 响应: {re}。请检查后端服务是否正常运行。"
        
        # 检查响应格式
        if isinstance(re[1], dict) and "data" in re[1]:
            for operator in re[1]["data"]:
                operator_info = {
                    "operator_id": operator["operator_id"],
                    "version": operator["version"]
                }
                unpublish_operator_infos.append(operator_info)
        else:
            print(f"警告: 获取未发布算子列表响应格式异常: {re[1]}")

        data = {"status": "offline"}
        re = self._call_with_retry(self.client.GetOperatorList, data, Headers)
        assert re[0] == 200, f"获取已下架算子列表失败，状态码: {re[0]}, 响应: {re}。请检查后端服务是否正常运行。"
        
        # 检查响应格式
        if isinstance(re[1], dict) and "data" in re[1]:
            for operator in re[1]["data"]:
                operator_info = {
                    "operator_id": operator["operator_id"],
                    "version": operator["version"]
                }
                offline_operator_infos.append(operator_info)
        else:
            print(f"警告: 获取已下架算子列表响应格式异常: {re[1]}")
        
        # 验证列表不为空
        if len(published_operator_infos) == 0:
            print(f"警告: published_operator_infos 为空，某些测试可能会失败")
        if len(unpublish_operator_infos) == 0:
            print(f"警告: unpublish_operator_infos 为空，某些测试可能会失败")
        if len(offline_operator_infos) == 0:
            print(f"警告: offline_operator_infos 为空，某些测试可能会失败")
            

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

        # 检查列表是否为空
        if len(published_operator_infos) == 0:
            pytest.skip("published_operator_infos 为空，跳过此测试")

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

        # 检查列表是否为空
        if len(offline_operator_infos) == 0:
            pytest.skip("offline_operator_infos 为空，跳过此测试")

        # 验证算子是否存在
        operator_id = offline_operator_infos[0]["operator_id"]
        check_result = self.client.GetOperatorInfo(operator_id, Headers)
        if check_result[0] != 200:
            pytest.skip(f"算子 {operator_id} 不存在或无法访问，状态码: {check_result[0]}")

        # 使用算子注册时的原始OpenAPI数据，确保路径和方法匹配
        global operator_openapi_data
        if operator_id not in operator_openapi_data:
            pytest.skip(f"算子 {operator_id} 的OpenAPI数据未找到")
        
        data = {
            "operator_id": operator_id,
            "data": operator_openapi_data[operator_id],
            "operator_metadata_type": "openapi"
        }
        result = self._call_with_retry(self.client.UpdateOperatorInfo, data, Headers)
        assert result[0] == 200, f"更新算子信息失败，状态码: {result[0]}, 响应: {result}。算子ID: {operator_id}"
        assert result[1][0]["status"] == "success"
        assert result[1][0]["version"] != offline_operator_infos[0]["version"]
        
        re = self._call_with_retry(self.client.GetOperatorInfo, result[1][0]["operator_id"], Headers)  
        assert re[0] == 200, f"获取算子信息失败，状态码: {re[0]}, 响应: {re}"
        assert re[1]["status"] == "unpublish"
        
    @allure.title("更新已发布算子信息，更新成功，默认更新后的版本为已发布编辑中状态")
    def test_update_operator_info_04(self, Headers):
        global published_operator_infos

        # 检查列表是否为空
        if len(published_operator_infos) == 0:
            pytest.skip("published_operator_infos 为空，跳过此测试")

        # 验证算子是否存在
        operator_id = published_operator_infos[0]["operator_id"]
        check_result = self.client.GetOperatorInfo(operator_id, Headers)
        if check_result[0] != 200:
            pytest.skip(f"算子 {operator_id} 不存在或无法访问，状态码: {check_result[0]}")

        # 使用算子注册时的原始OpenAPI数据，确保路径和方法匹配
        global operator_openapi_data
        if operator_id not in operator_openapi_data:
            pytest.skip(f"算子 {operator_id} 的OpenAPI数据未找到")
        
        data = {
            "operator_id": operator_id,
            "data": operator_openapi_data[operator_id],
            "operator_metadata_type": "openapi"
        }
        result = self._call_with_retry(self.client.UpdateOperatorInfo, data, Headers)
        assert result[0] == 200, f"更新算子信息失败，状态码: {result[0]}, 响应: {result}。算子ID: {operator_id}"
        assert result[1][0]["status"] == "success"
        assert result[1][0]["operator_id"] == published_operator_infos[0]["operator_id"]
        assert result[1][0]["version"] != published_operator_infos[0]["version"]
        
        re = self._call_with_retry(self.client.GetOperatorInfo, result[1][0]["operator_id"], Headers)  
        assert re[0] == 200, f"获取算子信息失败，状态码: {re[0]}, 响应: {re}"
        assert re[1]["status"] == "editing"

    @allure.title("更新未发布算子信息，更新后直接发布，更新成功，无新版本生成")
    def test_update_operator_info_05(self, Headers):
        global unpublish_operator_infos

        # 检查列表是否为空
        if len(unpublish_operator_infos) == 0:
            pytest.skip("unpublish_operator_infos 为空，跳过此测试")

        # 验证算子是否存在
        operator_id = unpublish_operator_infos[0]["operator_id"]
        check_result = self.client.GetOperatorInfo(operator_id, Headers)
        if check_result[0] != 200:
            pytest.skip(f"算子 {operator_id} 不存在或无法访问，状态码: {check_result[0]}")

        # 使用算子注册时的原始OpenAPI数据，确保路径和方法匹配
        global operator_openapi_data
        if operator_id not in operator_openapi_data:
            pytest.skip(f"算子 {operator_id} 的OpenAPI数据未找到")
        
        data = {
            "operator_id": operator_id,
            "data": operator_openapi_data[operator_id],
            "operator_metadata_type": "openapi",
            "direct_publish": True
        }
        result = self._call_with_retry(self.client.UpdateOperatorInfo, data, Headers)
        assert result[0] == 200, f"更新算子信息失败，状态码: {result[0]}, 响应: {result}。算子ID: {operator_id}"
        assert result[1][0]["status"] == "success"
        assert result[1][0]["operator_id"] == unpublish_operator_infos[0]["operator_id"]
        assert result[1][0]["version"] == unpublish_operator_infos[0]["version"]
        
        re = self._call_with_retry(self.client.GetOperatorInfo, result[1][0]["operator_id"], Headers)  
        assert re[0] == 200, f"获取算子信息失败，状态码: {re[0]}, 响应: {re}"
        assert re[1]["status"] == "published"

    @allure.title("更新算子信息，执行模式为异步，标识为数据源算子，更新失败")
    def test_update_operator_info_06(self, Headers):
        global unpublish_operator_infos

        # 检查列表是否为空
        if len(unpublish_operator_infos) == 0:
            pytest.skip("unpublish_operator_infos 为空，跳过此测试")

        # 验证算子是否存在
        operator_id = unpublish_operator_infos[0]["operator_id"]
        check_result = self.client.GetOperatorInfo(operator_id, Headers)
        if check_result[0] != 200:
            pytest.skip(f"算子 {operator_id} 不存在或无法访问，状态码: {check_result[0]}")

        # 使用算子注册时的原始OpenAPI数据，确保路径和方法匹配
        global operator_openapi_data
        if operator_id not in operator_openapi_data:
            pytest.skip(f"算子 {operator_id} 的OpenAPI数据未找到")
        
        data = {
            "operator_id": operator_id,
            "data": operator_openapi_data[operator_id],
            "operator_metadata_type": "openapi",
            "operator_info": {
                "execution_mode": "async",
                "is_data_source": True
            }
        }
        result = self._call_with_retry(self.client.UpdateOperatorInfo, data, Headers)
        assert result[0] == 400, f"更新算子信息应该返回400，实际: {result[0]}, 响应: {result}。算子ID: {operator_id}"

    @allure.title("更新算子信息，执行模式为异步，标识为数据源算子，更新失败")
    def test_update_operator_info_07(self, Headers):
        global unpublish_operator_infos

        # 检查列表是否为空
        if len(unpublish_operator_infos) == 0:
            pytest.skip("unpublish_operator_infos 为空，跳过此测试")

        # 验证算子是否存在（尝试找到第一个有效的算子）
        global operator_openapi_data
        operator_id = None
        for op_info in unpublish_operator_infos:
            op_id = op_info["operator_id"]
            check_result = self.client.GetOperatorInfo(op_id, Headers)
            if check_result[0] == 200 and op_id in operator_openapi_data:
                operator_id = op_id
                break
        
        if operator_id is None:
            pytest.skip(f"未找到有效的未发布算子，所有算子都无法访问或缺少OpenAPI数据")

        # 使用算子注册时的原始OpenAPI数据，确保路径和方法匹配
        data = {
            "operator_id": operator_id,
            "data": operator_openapi_data[operator_id],
            "operator_metadata_type": "openapi",
            "operator_info": {
                "execution_mode": "sync",
                "is_data_source": True
            }
        }
        result = self._call_with_retry(self.client.UpdateOperatorInfo, data, Headers)
        assert result[0] == 200, f"更新算子信息失败，状态码: {result[0]}, 响应: {result}。算子ID: {operator_id}"
        operator_id = result[1][0]["operator_id"]
        result = self._call_with_retry(self.client.GetOperatorInfo, operator_id, Headers)
        assert result[0] == 200, f"获取算子信息失败，状态码: {result[0]}, 响应: {result}"
        assert result[1]["operator_id"] == operator_id
        operator_info = result[1]["operator_info"]
        assert operator_info["execution_mode"] == "sync"
        assert operator_info["is_data_source"] == True
        
        
        
        
        