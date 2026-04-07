# -*- coding:UTF-8 -*-

import pytest
import allure
import uuid
import json
from jsonschema import Draft7Validator
import os

from lib.dataflow_como_operator import AutomationClient


@allure.feature("automation服务接口测试：创建组合算子")
class TestCreateCombinationOperator:
    client = AutomationClient()

    # @pytest.mark.parametrize("node_type", [
    #     "sync_nodes",    # sync场景1
    #     "combo_nodes",   # sync场景2
    #     "python_nodes",  # async场景
    # ])
    # @allure.title("通过json节点组合创建组合算子-分类型场景")
    # def test_create_combination_operator_by_type(self, Headers, node_type):
    #     # 读取所有节点
    #     with open('./data/data-flow/start_node.json', encoding='utf-8') as f:
    #         start_node = json.load(f)
    #     with open('./data/data-flow/end_node.json', encoding='utf-8') as f:
    #         end_node = json.load(f)
    #     with open('./data/data-flow/python_nodes.json', encoding='utf-8') as f:
    #         python_nodes = json.load(f)
    #     with open('./data/data-flow/sync_nodes.json', encoding='utf-8') as f:
    #         sync_nodes = json.load(f)
    #     with open('./data/data-flow/combo_nodes.json', encoding='utf-8') as f:
    #         combo_nodes = json.load(f)

    #     if node_type == "sync_nodes":
    #         middle_nodes = sync_nodes
    #     elif node_type == "combo_nodes":
    #         middle_nodes = combo_nodes
    #     else:
    #         middle_nodes = python_nodes

    #     steps = [start_node[0]] + middle_nodes + [end_node[0]]

    #     unique_title = f"组合算子_{node_type}_{str(uuid.uuid4())[:8]}"
    #     data = {
    #         "title": unique_title,
    #         "description": f"{node_type}节点组合创建",
    #         "category": "other_category",
    #         "steps": steps,
    #         "outputs": [
    #             {
    #                 "description": {"type": "text"},
    #                 "type": "string",
    #                 "key": "shuchu"
    #             }
    #         ]
    #     }

    #     result = self.client.CreateCombinationOperator(data, Headers)
    #     assert result[0] == 201

   

    # @allure.title("创建组合算子失败 - 缺少必要参数title")
    # def test_create_combination_operator_missing_title(self, Headers):
    #     with open('./data/data-flow/start_node.json', encoding='utf-8') as f:
    #         start_node = json.load(f)
    #     with open('./data/data-flow/end_node.json', encoding='utf-8') as f:
    #         end_node = json.load(f)
    #     with open('./data/data-flow/sync_nodes.json', encoding='utf-8') as f:
    #         sync_nodes = json.load(f)
    #     steps = [start_node[0], sync_nodes[0], end_node[0]]
    #     data = {
    #         "description": "算子实现-测试数据",
    #         "steps": steps,
    #         "category": "data_split"
    #     }
    #     result = self.client.CreateCombinationOperator(data, Headers)
    #     assert result[0] == 400

    # @allure.title("创建组合算子失败 - title已存在（仅开始和结束节点）")
    # def test_create_combination_operator_duplicate_title_simple(self, Headers):
    #     # 只取第一个开始节点和第一个结束节点
    #     with open('./data/data-flow/start_node.json', encoding='utf-8') as f:
    #         start_nodes = json.load(f)
    #     with open('./data/data-flow/end_node.json', encoding='utf-8') as f:
    #         end_nodes = json.load(f)
    #     start_node = start_nodes[0]
    #     end_node = end_nodes[0]
    #     steps = [start_node, end_node]
    #     unique_title = f"重复标题测试_{str(uuid.uuid4())[:8]}"
    #     data = {
    #         "title": unique_title,
    #         "description": "测试数据",
    #         "category": "other_category",
    #         "steps": steps,
    #         "outputs": [
    #             {
    #                 "description": {"type": "text"},
    #                 "type": "string",
    #                 "key": "success"
    #             }
    #         ]
    #     }
    #     # 第一次创建
    #     first_result = self.client.CreateCombinationOperator(data, Headers)
    #     assert first_result[0] == 201
    #     # 第二次用相同title创建
    #     second_result = self.client.CreateCombinationOperator(data, Headers)
    #     assert second_result[0] == 409  # 冲突状态码

    # @allure.title("创建组合算子失败 - 无效的category")
    # def test_create_combination_operator_invalid_category(self, Headers):
    #     with open('./data/data-flow/start_node.json', encoding='utf-8') as f:
    #         start_node = json.load(f)
    #     with open('./data/data-flow/end_node.json', encoding='utf-8') as f:
    #         end_node = json.load(f)
    #     with open('./data/data-flow/sync_nodes.json', encoding='utf-8') as f:
    #         sync_nodes = json.load(f)
    #     steps = [start_node, sync_nodes[0], end_node]
    #     data = {
    #         "title": f"无效分类测试_{str(uuid.uuid4())[:8]}",
    #         "description": "测试无效分类",
    #         "steps": steps,
    #         "category": "invalid_category"  # 无效分类
    #     }
    #     result = self.client.CreateCombinationOperator(data, Headers)
    #     assert result[0] == 400

    # @allure.title("创建组合算子失败 - 无效的steps结构")
    # def test_create_combination_operator_invalid_steps(self, Headers):
    #     # 构造一个无效的开始节点（缺少 operator 字段）
    #     invalid_start_node = {
    #         "id": "0",
    #         "title": "",
    #         "parameters": {
    #             "fields": [
    #                 {
    #                     "key": "text",
    #                     "name": "text",
    #                     "required": True,
    #                     "type": "string"
    #                 }
    #             ]
    #         }
    #     }
    #     with open('./data/data-flow/end_node.json', encoding='utf-8') as f:
    #         end_node = json.load(f)
    #     with open('./data/data-flow/sync_nodes.json', encoding='utf-8') as f:
    #         sync_nodes = json.load(f)
    #     steps = [invalid_start_node, sync_nodes[0], end_node]
    #     data = {
    #         "title": f"无效步骤测试_{str(uuid.uuid4())[:8]}",
    #         "description": "测试无效步骤结构",
    #         "steps": steps,
    #         "category": "data_split"
    #     }
    #     result = self.client.CreateCombinationOperator(data, Headers)
    #     assert result[0] == 400

    # @allure.title("所有开始/结束节点组合+sync节点")
    # def test_create_operator_all_start_end_with_sync(self, Headers):
    #     with open('./data/data-flow/start_node.json', encoding='utf-8') as f:
    #         start_nodes = json.load(f)
    #     with open('./data/data-flow/end_node.json', encoding='utf-8') as f:
    #         end_nodes = json.load(f)
    #     with open('./data/data-flow/sync_nodes.json', encoding='utf-8') as f:
    #         sync_nodes = json.load(f)

    #     outputs = [
    #                 {
    #                     "name": "str-end名称",
    #                     "description": {
    #                         "text": "str-end说明",
    #                         "type": "text"
    #                     },
    #                     "type": "string",
    #                     "key": "str"
    #                 },
    #                 {
    #                     "name": "num-end名称",
    #                     "description": {
    #                         "text": "num-end说明",
    #                         "type": "text"
    #                     },
    #                     "type": "number",
    #                     "key": "numend"
    #                 },
    #                 {
    #                     "name": "obj-end",
    #                     "description": {
    #                         "text": "obj-end",
    #                         "type": "text"
    #                     },
    #                     "type": "object",
    #                     "key": "obj"
    #                 },
    #                 {
    #                     "name": "arrend",
    #                     "description": {
    #                         "text": "arrend",
    #                         "type": "text"
    #                     },
    #                     "type": "array",
    #                     "key": "arr"
    #                 }
    # ]  # 你的最大outputs结构 # 你的最大outputs结构

    #     for i, start_node in enumerate(start_nodes):
    #         for j, end_node in enumerate(end_nodes):
    #             for k, sync_node in enumerate(sync_nodes):
    #                 steps = [start_node, sync_node, end_node]
    #                 data = {
    #                     "title": f"sync场景_{i}_{j}_{k}_{str(uuid.uuid4())[:8]}",
    #                     "description": f"sync_nodes: 开始{i} + sync{k} + 结束{j}",
    #                     "category": "other_category",
    #                     "steps": steps,
    #                     "outputs": outputs
    #                 }
    #                 result = self.client.CreateCombinationOperator(data, Headers)
    #                 assert result[0] == 201, f"sync场景{i}-{j}-{k}创建失败，返回：{result}"

    # @allure.title("所有开始/结束节点组合+combo节点")
    # def test_create_operator_all_start_end_with_combo(self, Headers):
    #     with open('./data/data-flow/start_node.json', encoding='utf-8') as f:
    #         start_nodes = json.load(f)
    #     with open('./data/data-flow/end_node.json', encoding='utf-8') as f:
    #         end_nodes = json.load(f)
    #     with open('./data/data-flow/combo_nodes.json', encoding='utf-8') as f:
    #         combo_nodes = json.load(f)

    #     outputs = [
    #                 {
    #                     "name": "str-end名称",
    #                     "description": {
    #                         "text": "str-end说明",
    #                         "type": "text"
    #                     },
    #                     "type": "string",
    #                     "key": "str"
    #                 },
    #                 {
    #                     "name": "num-end名称",
    #                     "description": {
    #                         "text": "num-end说明",
    #                         "type": "text"
    #                     },
    #                     "type": "number",
    #                     "key": "numend"
    #                 },
    #                 {
    #                     "name": "obj-end",
    #                     "description": {
    #                         "text": "obj-end",
    #                         "type": "text"
    #                     },
    #                     "type": "object",
    #                     "key": "obj"
    #                 },
    #                 {
    #                     "name": "arrend",
    #                     "description": {
    #                         "text": "arrend",
    #                         "type": "text"
    #                     },
    #                     "type": "array",
    #                     "key": "arr"
    #                 }
    # ]  # 你的最大outputs结构  # 你的最大outputs结构

    #     for i, start_node in enumerate(start_nodes):
    #         for j, end_node in enumerate(end_nodes):
    #             for k, combo_node in enumerate(combo_nodes):
    #                 steps = [start_node, combo_node, end_node]
    #                 data = {
    #                     "title": f"combo场景_{i}_{j}_{k}_{str(uuid.uuid4())[:8]}",
    #                     "description": f"combo_nodes: 开始{i} + combo{k} + 结束{j}",
    #                     "category": "other_category",
    #                     "steps": steps,
    #                     "outputs": outputs
    #                 }
    #                 result = self.client.CreateCombinationOperator(data, Headers)
    #                 assert result[0] == 201, f"combo场景{i}-{j}-{k}创建失败，返回：{result}"

    # @allure.title("所有开始/结束节点组合+python节点")
    # def test_create_operator_all_start_end_with_python(self, Headers):
    #     with open('./data/data-flow/start_node.json', encoding='utf-8') as f:
    #         start_nodes = json.load(f)
    #     with open('./data/data-flow/end_node.json', encoding='utf-8') as f:
    #         end_nodes = json.load(f)
    #     with open('./data/data-flow/python_nodes.json', encoding='utf-8') as f:
    #         python_nodes = json.load(f)

    #     outputs = [
    #                 {
    #                     "name": "str-end名称",
    #                     "description": {
    #                         "text": "str-end说明",
    #                         "type": "text"
    #                     },
    #                     "type": "string",
    #                     "key": "str"
    #                 },
    #                 {
    #                     "name": "num-end名称",
    #                     "description": {
    #                         "text": "num-end说明",
    #                         "type": "text"
    #                     },
    #                     "type": "number",
    #                     "key": "numend"
    #                 },
    #                 {
    #                     "name": "obj-end",
    #                     "description": {
    #                         "text": "obj-end",
    #                         "type": "text"
    #                     },
    #                     "type": "object",
    #                     "key": "obj"
    #                 },
    #                 {
    #                     "name": "arrend",
    #                     "description": {
    #                         "text": "arrend",
    #                         "type": "text"
    #                     },
    #                     "type": "array",
    #                     "key": "arr"
    #                 }
    # ]  # 你的最大outputs结构

    #     for i, start_node in enumerate(start_nodes):
    #         for j, end_node in enumerate(end_nodes):
    #             for k, python_node in enumerate(python_nodes):
    #                 steps = [start_node, python_node, end_node]
    #                 data = {
    #                     "title": f"python场景_{i}_{j}_{k}_{str(uuid.uuid4())[:8]}",
    #                     "description": f"python_nodes: 开始{i} + python{k} + 结束{j}",
    #                     "category": "other_category",
    #                     "steps": steps,
    #                     "outputs": outputs
    #                 }
    #                 result = self.client.CreateCombinationOperator(data, Headers)
    #                 assert result[0] == 201, f"python场景{i}-{j}-{k}创建失败，返回：{result}"

   


    @allure.title("循环节点组合算子场景")
    def test_create_combination_operator_by_type(self, Headers):
        # 读取所有节点
        with open('./data/data-flow/loop_nodes.json', encoding='utf-8') as f:
            loop_nodes = json.load(f)

        steps = loop_nodes
        unique_title = f"组合算子_loop_nodes_{str(uuid.uuid4())[:8]}"
        data = {
            "title": unique_title,
            "description": "loop_nodes节点组合创建",
            "category": "other_category",
            "steps": steps
          
        }

        result = self.client.CreateCombinationOperator(data, Headers)
        print(result)
        assert result[0] == 201