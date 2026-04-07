# -*- coding:UTF-8 -*-

import pytest
import allure
import uuid
import json
import time

from lib.dataflow_como_operator import AutomationClient


@allure.feature("automation服务接口测试：工作流节点类型")
class TestWorkflowNodeTypes:
    client = AutomationClient()



    @allure.title("创建数据流--小模型reranker")
    def test_create_dataflow_reranker(self, Headers):
        """测试创建数据流，使用小模型reranker"""
        unique_title = f"reranker_{str(uuid.uuid4())[:8]}"
        data = {
            "title": unique_title,
            "steps": [
                {
                    "id": "0",
                    "operator": "@trigger/form",
                    "parameters": {
                        "fields": [
                            {
                                "key": "mtUYYHCWfFZrdznQ",
                                "type": "string",
                                "name": "文本一"
                            },
                            {
                                "key": "lursMrTbvvOtVLej",
                                "type": "string",
                                "name": "文本二"
                            }
                        ]
                    },
                    "title": ""
                },
                {
                    "id": "1",
                    "title": "",
                    "operator": "@llm/reranker",
                    "parameters": {
                        "documents": [
                            "人工智能是模拟人类智能的技术系统",
                            "机器学习是让计算机从数据中学习规律",
                            "深度学习是机器学习的一个分支，使用神经网络",
                            "TensorFlow和PyTorch都是流行的深度学习框架"
                        ],
                        "model": "reranker",
                        "query": "人工智能是什么"
                    }
                },
                {
                    "id": "1002",
                    "operator": "@internal/tool/py3",
                    "parameters": {
                        "input_params": [
                            {
                                "id": "rf5gz",
                                "key": "a",
                                "type": "array",
                                "value": " [     \"Python基础语法教程\",     \"Java高级编程技巧\",      \"Python数据分析实战\",     \"C++内存管理详解\"   ]"
                            }
                        ],
                        "output_params": [
                            {
                                "id": "vftxz",
                                "key": "b",
                                "type": "array"
                            }
                        ],
                        "code": "def main(a):\n    b=a\n    return b"
                    }
                },
                {
                    "id": "1001",
                    "operator": "@llm/reranker",
                    "parameters": {
                        "model": "reranker",
                        "query": "如何学习Python编程",
                        "documents": "{{__1002.b}}"
                    },
                    "title": "Reranker(1)"
                }
            ]
        }

        result = self.client.CreateDag(data, Headers)
        
        assert result[0] == 201, f"创建数据流应返回状态码201，实际为: {result[0]}"
        assert "id" in result[1], "响应应包含id字段"
        
        dag_id = result[1]["id"]
        allure.attach(str(dag_id), name="创建的DAG ID")
        return dag_id

    @allure.title("创建数据流--小模型embedding")
    def test_create_dataflow_embedding(self, Headers):
        """测试创建数据流，使用小模型embedding"""
        unique_title = f"embedding_{str(uuid.uuid4())[:8]}"
        data = {
            "title": unique_title,
            "steps": [
                {
                    "id": "0",
                    "title": "",
                    "operator": "@trigger/form",
                    "parameters": {
                        "fields": [
                            {
                                "key": "RvDPGCjXWTEtrVad",
                                "name": "aa",
                                "type": "string"
                            }
                        ]
                    }
                },
                {
                    "id": "1",
                    "title": "",
                    "operator": "@llm/embedding",
                    "parameters": {
                        "input": [
                            "今天天气很好",
                            "今天天气不错",
                            "今天天气晴朗",
                            "我喜欢吃苹果",
                            "苹果公司发布了新手机"
                        ],
                        "model": "embedding"
                    }
                },
                {
                    "id": "1003",
                    "operator": "@internal/tool/py3",
                    "parameters": {
                        "input_params": [
                            {
                                "id": "fjo54",
                                "key": "a",
                                "type": "array",
                                "value": "['人工智能技术','人工智能是模拟人类智能的科学','人工智能是计算机科学的一个分支，旨在创造能够执行智能任务的机器']"
                            }
                        ],
                        "output_params": [
                            {
                                "id": "8q6b5",
                                "key": "b",
                                "type": "array"
                            }
                        ],
                        "code": "def main(a):\n    b=a\n    return b"
                    }
                },
                {
                    "id": "1001",
                    "operator": "@llm/embedding",
                    "parameters": {
                        "model": "embedding",
                        "input": "{{__1003.b}}"
                    },
                    "title": "Embedding(1)"
                }
            ]
        }

        result = self.client.CreateDag(data, Headers)
        
        assert result[0] == 201, f"创建数据流应返回状态码201，实际为: {result[0]}"
        assert "id" in result[1], "响应应包含id字段"
        
        dag_id = result[1]["id"]
        allure.attach(str(dag_id), name="创建的DAG ID")
        return dag_id

    @allure.title("运行数据流--python节点")
    def test_run_dataflow_python(self, Headers):
        """测试运行包含python节点的数据流"""
        dag_id = self.test_create_dataflow_python_node(Headers)
        
        run_data = {}
        result = self.client.RunInstance(dag_id, run_data, Headers)
        
        assert result[0] in [200, 201, 202], f"运行数据流应返回状态码200/201/202，实际为: {result[0]}"
        
        time.sleep(3)
        return dag_id

    @allure.title("运行数据流--reranker")
    def test_run_dataflow_reranker(self, Headers):
        """测试运行包含reranker的数据流"""
        dag_id = self.test_create_dataflow_reranker(Headers)
        
        run_data = {
            "data": {}
        }
        result = self.client.RunInstanceForm(dag_id, run_data, Headers)
        
        assert result[0] in [200, 201, 202], f"运行数据流应返回状态码200/201/202，实际为: {result[0]}"
        
        time.sleep(3)
        return dag_id

    @allure.title("运行数据流--embedding")
    def test_run_dataflow_embedding(self, Headers):
        """测试运行包含embedding的数据流"""
        dag_id = self.test_create_dataflow_embedding(Headers)
        
        run_data = {
            "data": {
                "RvDPGCjXWTEtrVad": "a"
            }
        }
        result = self.client.RunInstanceForm(dag_id, run_data, Headers)
        
        assert result[0] in [200, 201, 202], f"运行数据流应返回状态码200/201/202，实际为: {result[0]}"
        
        time.sleep(3)
        return dag_id

    @allure.title("完整流程测试--变量节点")
    def test_workflow_variable_lifecycle(self, Headers):
        """测试变量节点工作流的完整生命周期"""
        # 1. 创建工作流
        dag_id = self.test_create_workflow_variable_node(Headers)
        
        # 2. 运行工作流
        run_result = self.client.RunInstance(dag_id, {}, Headers)
        assert run_result[0] in [200, 201, 202], "运行工作流失败"
        
        # 3. 等待执行完成
        time.sleep(3)
        
        # 4. 查看运行日志
        results_result = self.client.GetDagResultsV2(dag_id, {}, Headers)
        assert results_result[0] == 200, "查看运行日志失败"
        assert "results" in results_result[1], "响应应包含results字段"
        
        # 5. 查看日志详情
        if len(results_result[1]["results"]) > 0:
            run_id = results_result[1]["results"][0]["id"]
            detail_result = self.client.GetDagResultDetail(dag_id, run_id, Headers)
            assert detail_result[0] == 200, "查看日志详情失败"

    @allure.title("完整流程测试--文本节点")
    def test_workflow_text_lifecycle(self, Headers):
        """测试文本节点工作流的完整生命周期"""
        # 1. 创建工作流
        dag_id = self.test_create_workflow_text_node(Headers)
        
        # 2. 运行工作流（表单触发）
        run_data = {
            "data": {
                "ZHmeLdTgSBbsQHuS": "test_value"
            }
        }
        run_result = self.client.RunInstanceForm(dag_id, run_data, Headers)
        assert run_result[0] in [200, 201, 202], "运行工作流失败"
        
        # 3. 等待执行完成
        time.sleep(3)
        
        # 4. 查看运行日志
        results_result = self.client.GetDagResultsV2(dag_id, {}, Headers)
        assert results_result[0] == 200, "查看运行日志失败"
        assert "results" in results_result[1], "响应应包含results字段"
        
        # 5. 查看日志详情
        if len(results_result[1]["results"]) > 0:
            run_id = results_result[1]["results"][0]["id"]
            detail_result = self.client.GetDagResultDetail(dag_id, run_id, Headers)
            assert detail_result[0] == 200, "查看日志详情失败"

    @allure.title("完整流程测试--python节点")
    def test_workflow_python_lifecycle(self, Headers):
        """测试python节点数据流的完整生命周期"""
        # 1. 创建数据流
        dag_id = self.test_create_dataflow_python_node(Headers)
        
        # 2. 运行数据流
        run_result = self.client.RunInstance(dag_id, {}, Headers)
        assert run_result[0] in [200, 201, 202], "运行数据流失败"
        
        # 3. 等待执行完成
        time.sleep(5)
        
        # 4. 查看运行日志
        results_result = self.client.GetDagResultsV2(dag_id, {}, Headers)
        assert results_result[0] == 200, "查看运行日志失败"
        assert "results" in results_result[1], "响应应包含results字段"
        
        # 5. 查看日志详情
        if len(results_result[1]["results"]) > 0:
            run_id = results_result[1]["results"][0]["id"]
            detail_result = self.client.GetDagResultDetail(dag_id, run_id, Headers)
            assert detail_result[0] == 200, "查看日志详情失败"

