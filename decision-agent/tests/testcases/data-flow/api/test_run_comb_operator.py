# # -*- coding:UTF-8 -*-
# import pytest
# import allure
# import uuid
# import json
# from lib.dataflow_como_operator import AutomationClient

# @allure.feature("automation服务接口测试：运行组合算子")
# class TestRunCombinationOperator:
#     client = AutomationClient()

#     @allure.title("创建并运行组合算子（开始+sync_node+结束）")
#     def test_create_and_run_comb_operator(self, Headers):
#         # 读取节点
#         with open('./data/data-flow/start_node.json', encoding='utf-8') as f:
#             start_nodes = json.load(f)
#         with open('./data/data-flow/end_node.json', encoding='utf-8') as f:
#             end_nodes = json.load(f)
#         with open('./data/data-flow/sync_nodes.json', encoding='utf-8') as f:
#             sync_nodes = json.load(f)

#         # 只用第一个节点
#         steps = [start_nodes[0], sync_nodes[0], end_nodes[0]]

#         # 组合算子数据
#         unique_title = f"运行组合算子_{str(uuid.uuid4())[:8]}"
#         data = {
#             "title": unique_title,
#             "description": "自动化测试-运行组合算子",
#             "category": "other_category",
#             "steps": steps,
#             "outputs": [
#                 {
#                     "description": {"type": "text"},
#                     "type": "string",
#                     "key": "success"
#                 }
#             ]
#         }

#         # 创建组合算子
#         result = self.client.CreateCombinationOperator(data, Headers)
#         assert result[0] == 201
#         dag_id = result[1].get("operator_id") or result[1].get("id")
#         assert dag_id, "未获取到dag_id"

#         # 运行组合算子
#         run_body = {"str": "test"}
#         run_result = self.client.RuneOperator(dag_id, run_body, Headers)
#         assert run_result[0] == 200, f"运行组合算子失败，返回：{run_result}"


#     @allure.title("创建并运行异步组合算子（开始+python+结束）")
#     def test_create_and_run_async_comb_operator(self, Headers):
#         # 读取节点
#         with open('./data/data-flow/start_node.json', encoding='utf-8') as f:
#             start_nodes = json.load(f)
#         with open('./data/data-flow/end_node.json', encoding='utf-8') as f:
#             end_nodes = json.load(f)
#         with open('./data/data-flow/python_nodes.json', encoding='utf-8') as f:
#             python_nodes = json.load(f)

#         # 只用第一个节点
#         steps = [start_nodes[0], python_nodes[0], end_nodes[0]]

#         # 组合算子数据
#         unique_title = f"async组合算子_{str(uuid.uuid4())[:8]}"
#         data = {
#             "title": unique_title,
#             "description": "自动化测试-异步组合算子",
#             "category": "other_category",
#             "steps": steps,
#             "outputs": [
#                 {
#                     "description": {"type": "text"},
#                     "type": "string",
#                     "key": "success"
#                 }
#             ]
#         }

#         # 创建组合算子
#         result = self.client.CreateCombinationOperator(data, Headers)
#         assert result[0] == 201
#         dag_id = result[1].get("operator_id") or result[1].get("id")
#         assert dag_id, "未获取到dag_id"

#         # 运行组合算子
#         run_body = {"str": "test"}
#         run_result = self.client.RuneOperator(dag_id, run_body, Headers)
#         assert run_result[0] == 202, f"运行异步组合算子失败，返回：{run_result}"

#     @allure.title("创建并运行python组合算子（开始+combo_node+结束）")
#     def test_create_and_run_comb_operator2(self, Headers):
#         # 读取节点
#         with open('./data/data-flow/start_node.json', encoding='utf-8') as f:
#             start_nodes = json.load(f)
#         with open('./data/data-flow/end_node.json', encoding='utf-8') as f:
#             end_nodes = json.load(f)
#         with open('./data/data-flow/combo_nodes.json', encoding='utf-8') as f:
#             combo_nodes = json.load(f)

#         # 只用第一个节点
#         steps = [start_nodes[0], combo_nodes[0], end_nodes[0]]
#         # 组合算子数据
#         unique_title = f"运行组合算子_{str(uuid.uuid4())[:8]}"
#         data = {
#             "title": unique_title,
#             "description": "自动化测试-运行组合算子",
#             "category": "other_category",
#             "steps": steps,
#             "outputs": [
#                 {
#                     "description": {"type": "text"},
#                     "type": "string",
#                     "key": "success"
#                 }
#             ]
#         }

#         # 创建组合算子
#         result = self.client.CreateCombinationOperator(data, Headers)
#         assert result[0] == 201
#         dag_id = result[1].get("operator_id") or result[1].get("id")
#         assert dag_id, "未获取到dag_id"

#         # 运行组合算子
#         run_body = {"str": "test"}
#         run_result = self.client.RuneOperator(dag_id, run_body, Headers)
#         assert run_result[0] == 200, f"运行组合算子失败，返回：{run_result}"


#     @allure.title("创建并运行多个组合算子（开始+python+sync_nodes+结束）")
#     def test_create_and_run_comb_operator3(self, Headers):
#         # 读取节点
#         with open('./data/data-flow/start_node.json', encoding='utf-8') as f:
#             start_nodes = json.load(f)
#         with open('./data/data-flow/end_node.json', encoding='utf-8') as f:
#             end_nodes = json.load(f)
#         with open('./data/data-flow/sync_nodes.json', encoding='utf-8') as f:
#             sync_nodes = json.load(f)
#         with open('./data/data-flow/python_nodes.json', encoding='utf-8') as f:
#             python_nodes = json.load(f)
#         # 只用第一个节点
#         steps = [start_nodes[0], python_nodes[0], sync_nodes[0], end_nodes[0]]
#         # 组合算子数据
#         unique_title = f"运行组合算子_{str(uuid.uuid4())[:8]}"
#         data = {
#             "title": unique_title,
#             "description": "自动化测试-运行组合算子",
#             "category": "other_category",
#             "steps": steps,
#             "outputs": [
#                 {
#                     "description": {"type": "text"},
#                     "type": "string",
#                     "key": "success"
#                 }
#             ]
#         }

#         # 创建组合算子
#         result = self.client.CreateCombinationOperator(data, Headers)
#         assert result[0] == 201
#         dag_id = result[1].get("operator_id") or result[1].get("id")
#         assert dag_id, "未获取到dag_id"

#         # 运行组合算子
#         run_body = {"str": "test"}
#         run_result = self.client.RuneOperator(dag_id, run_body, Headers)
#         assert run_result[0] == 202, f"运行组合算子失败，返回：{run_result}"


#     @allure.title("创建并运行组合算子（开始[1]+sync_node+结束）- 所有必填参数")
#     def test_create_and_run_comb_operator_start1_all_required(self, Headers):
#         #读取节点
#         with open('./data/data-flow/start_node.json', encoding='utf-8') as f:
#             start_nodes = json.load(f)
#         with open('./data/data-flow/end_node.json', encoding='utf-8') as f:
#             end_nodes = json.load(f)
#         with open('./data/data-flow/sync_nodes.json', encoding='utf-8') as f:
#             sync_nodes = json.load(f)

#         # 只用第二个开始节点
#         steps = [start_nodes[1], sync_nodes[0], end_nodes[0]]

#         # 组合算子数据
#         unique_title = f"组合算子_start1_{str(uuid.uuid4())[:8]}"
#         data = {
#             "title": unique_title,
#             "description": "自动化测试-开始节点1所有必填参数",
#             "category": "other_category",
#             "steps": steps,
#             "outputs": [
#                 {
#                     "description": {"type": "text"},
#                     "type": "string",
#                     "key": "success"
#                 }
#             ]
#         }

#         # 创建组合算子
#         result = self.client.CreateCombinationOperator(data, Headers)
#         assert result[0] == 201
#         dag_id = result[1].get("operator_id") or result[1].get("id")
#         assert dag_id, "未获取到dag_id"

#         # 自动提取所有必填参数 key 并构造 run_body
#         fields = start_nodes[1].get("parameters", {}).get("fields", [])
#         run_body = {}
#         for field in fields:
#             key = field.get("key")
#             typ = field.get("type")
#             # 根据类型填充测试值
#             if typ == "string":
#                 run_body[key] = "test"
#             elif typ == "number":
#                 run_body[key] = 123
#             elif typ == "array":
#                 run_body[key] = [1, 2, 3]
#             elif typ == "object":
#                 run_body[key] = {"a": 1}
#             else:
#                 run_body[key] = None

#         # 运行组合算子（方法名为 RuneOperator）
#         run_result = self.client.RuneOperator(dag_id, run_body, Headers)
#         assert run_result[0] in (200, 202), f"运行组合算子失败，返回：{run_result}"

#     @allure.title("获取DAG运行结果（v2，支持分页）")
#     def test_get_dag_result_v2(self, Headers):
#         # 读取节点
#         with open('./data/data-flow/loop_nodes.json', encoding='utf-8') as f:
#             sync_nodes = json.load(f)

#         steps = loop_nodes
#         unique_title = f"获取DAG结果v2_{str(uuid.uuid4())[:8]}"
#         data = {
#             "title": unique_title,
#             "description": "自动化测试-获取DAG结果v2",
#             "category": "other_category",
#             "steps": steps,
#             "outputs": [
#                 {
#                     "description": {"type": "text"},
#                     "type": "string",
#                     "key": "success"
#                 }
#             ]
#         }
#         # 创建组合算子
#         result = self.client.CreateCombinationOperator(data, Headers)
#         assert result[0] == 201
#         dag_id = result[1].get("operator_id") or result[1].get("id")
#         assert dag_id, "未获取到dag_id"
#         # 运行组合算子
#         run_body = {"n": 20}
#         run_result = self.client.RuneOperator(dag_id, run_body, Headers)
#         assert run_result[0] == 200
#         result_id = run_result[1].get("result_id") or run_result[1].get("id")
#         assert result_id, "未获取到result_id"
#         # 获取运行结果（v2，分页）
#         params = {"page": 0, "limit": 10}
#         dag_result = self.client.GetDagResultV2(dag_id, result_id, params, Headers)
#         assert dag_result[0] == 200
#         assert "limit" in dag_result[1]
#         assert "page" in dag_result[1]
#         assert "results" in dag_result[1]
#         assert "total" in dag_result[1]

#     @allure.title("获取运行记录（v2）")
#     def test_get_dag_results_v2(self, Headers):
#         # 创建并运行组合算子，获取 dag_id
#         with open('./data/data-flow/loop_nodes.json', encoding='utf-8') as f:
#             loop_nodes = json.load(f)
#         steps = loop_nodes
#         unique_title = f"运行记录v2_{str(uuid.uuid4())[:8]}"
#         data = {
#             "title": unique_title,
#             "description": "自动化测试-运行记录v2",
#             "category": "other_category",
#             "steps": steps,
#             "outputs": [
#                 {
#                     "description": {"type": "text"},
#                     "type": "string",
#                     "key": "success"
#                 }
#             ]
#         }
#         result = self.client.CreateCombinationOperator(data, Headers)
#         assert result[0] == 201
#         dag_id = result[1].get("operator_id") or result[1].get("id")
#         assert dag_id, "未获取到dag_id"
#         # 运行组合算子
#         run_body = {"n": 20}
#         run_result = self.client.RuneOperator(dag_id, run_body, Headers)
#         assert run_result[0] == 200
#         # 获取运行记录
#         params = {"page": 0, "limit": 20, "sortBy": "started_at", "order": "desc"}
#         results = self.client.GetDagResultsV2(dag_id, params, Headers)
#         assert results[0] == 200
#         assert "results" in results[1]
#         assert "total" in results[1]

#     @allure.title("获取执行日志（v2）")
#     def test_get_dag_result_log_v2(self, Headers):
#         # 先获取运行记录，拿到 result_id
#         with open('./data/data-flow/loop_nodes.json', encoding='utf-8') as f:
#             loop_nodes = json.load(f)
#         steps = loop_nodes
#         unique_title = f"执行日志v2_{str(uuid.uuid4())[:8]}"
#         data = {
#             "title": unique_title,
#             "description": "自动化测试-执行日志v2",
#             "category": "other_category",
#             "steps": steps,
#             "outputs": [
#                 {
#                     "description": {"type": "text"},
#                     "type": "string",
#                     "key": "success"
#                 }
#             ]
#         }
#         result = self.client.CreateCombinationOperator(data, Headers)
#         assert result[0] == 201
#         dag_id = result[1].get("operator_id") or result[1].get("id")
#         assert dag_id, "未获取到dag_id"
#         run_body = {"n": 100}
#         run_result = self.client.RuneOperator(dag_id, run_body, Headers)
#         assert run_result[0] == 200
#         # 获取运行记录
#         params = {"page": 0, "limit": 1, "sortBy": "started_at", "order": "desc"}
#         results = self.client.GetDagResultsV2(dag_id, params, Headers)
#         assert results[0] == 200
#         result_list = results[1].get("results", [])
#         assert result_list, "未获取到运行记录"
#         result_id = result_list[0].get("id")
#         assert result_id, "未获取到result_id"
#         # 获取执行日志
#         log_params = {"page": 0, "limit": 1}
#         log = self.client.GetDagResultLogV2(dag_id, result_id, log_params, Headers)
#         assert log[0] == 200
#         assert "results" in log[1]


import datetime
 
timestamp = 1749781537
date_object = datetime.datetime.fromtimestamp(timestamp)
formatted_date = date_object.strftime('%Y-%m-%d')
print(formatted_date)