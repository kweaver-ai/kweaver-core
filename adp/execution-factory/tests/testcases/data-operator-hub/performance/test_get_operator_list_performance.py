# -*- coding:UTF-8 -*-

import pytest
import allure
import time
import statistics
import json
import uuid
from concurrent.futures import ThreadPoolExecutor

from common.get_content import GetContent
from lib.operator import Operator

@allure.feature("算子注册与管理性能测试：获取算子列表")
class TestGetOperatorListPerformance:
    client = Operator()
    
    def measure_latency(self, func, *args, **kwargs):
        start_time = time.time()
        result = func(*args, **kwargs)
        end_time = time.time()
        return end_time - start_time, result

    def test_setup_class(self, Headers):
        """准备测试数据：注册1000个算子，发布700个，其中200个下架"""
        filepath = "./resource/openapi/compliant/test0.json"
        api_data = GetContent(filepath).jsonfile()
        
        # 定义所有可能的category
        categories = [
            "other_category",
            "data_process",
            "data_transform",
            "data_store",
            "data_analysis",
            "data_query",
            "data_extract",
            "data_split",
            "model_train"
        ]
        
        operator_ids=[]
        versions=[]
        count=0

        # 注册1000个算子（每次10个，共100次）
        for i in range(100):
            # 修改每个路径下的summary字段避免重名
            for path in api_data["paths"]:
                for method in api_data["paths"][path]:
                    if "summary" in api_data["paths"][path][method]:
                        id = str(uuid.uuid4())
                        api_data["paths"][path][method]["summary"] = f"Test Summary {id}"
            
            # 设置category（每次注册10个算子，确保均匀分布）
            current_category = categories[i % len(categories)]          
            
            data = {
                "data": str(api_data),
                "operator_metadata_type": "openapi",
                "operator_info": {
                    "category": current_category
                }
            }
            result = self.client.RegisterOperator(data, Headers)
            assert result[0] == 200
            operators = result[1]
            
            # 处理每个算子
            for operator in operators:
                if operator["status"] == "success":
                    count = count + 1
                    operator_id = operator["operator_id"]
                    version = operator["version"]
                    operator_ids.append(operator_id)
                    versions.append(version)

        # 发布70%的算子
        update_data1 = []
        for i in range(int(count*0.7)):            
            update_data = {
                "operator_id": operator_ids[i],
                "version": versions[i],
                "status": "published"
            }
            update_data1.append(update_data)

        # 下架20%的算子
        update_data2 = []
        for i in range(int(count*0.2)):            
            update_data = {
                "operator_id": operator_ids[i],
                "version": versions[i],
                "status": "offline"
            }
            update_data2.append(update_data)

        re = self.client.UpdateOperatorStatus(update_data1, Headers)
        assert re[0] == 200

        re = self.client.UpdateOperatorStatus(update_data2, Headers)
        assert re[0] == 200    
        

    @allure.title("不同分页大小下的性能测试")
    def test_page_size_performance(self, Headers):
        page_sizes = [10, 50, 100]
        for page_size in page_sizes:
            data = {"page_size": page_size}
            latency, result = self.measure_latency(self.client.GetOperatorList, data, Headers)
            assert result[0] == 200
            print(f"分页大小{page_size}的延迟: {latency:.3f}秒")
            allure.attach(f"分页大小{page_size}的延迟: {latency:.3f}秒", "性能指标")

    @allure.title("不同状态筛选的性能测试")
    def test_status_filter_performance(self, Headers):
        statuses = ["published", "unpublish", "offline"]
        for status in statuses:
            data = {"status": status}
            latency, result = self.measure_latency(self.client.GetOperatorList, data, Headers)
            assert result[0] == 200
            print(f"状态{status}筛选的延迟: {latency:.3f}秒")
            allure.attach(f"状态{status}筛选的延迟: {latency:.3f}秒", "性能指标")

    @allure.title("不同分类筛选的性能测试")
    def test_category_filter_performance(self, Headers):
        categories = ["other_category", "data_process", "data_transform", "data_store", "data_analysis", "data_query", "data_extract", "data_split", "model_train"]
        for category in categories:
            data = {"category": category}
            latency, result = self.measure_latency(self.client.GetOperatorList, data, Headers)
            assert result[0] == 200
            print(f"分类{category}筛选的延迟: {latency:.3f}秒")
            allure.attach(f"分类{category}筛选的延迟: {latency:.3f}秒", "性能指标")

    @allure.title("组合条件筛选的性能测试")
    def test_combined_filter_performance(self, Headers):
        test_cases = [
            {"status": "published", "category": "data_process", "page_size": 50},
            {"status": "published", "sort_by": "update_time", "sort_order": "desc"},
            {"category": "data_process", "page": 2, "page_size": 20}
        ]
        
        for test_case in test_cases:
            latency, result = self.measure_latency(self.client.GetOperatorList, test_case, Headers)
            assert result[0] == 200 
            print(f"组合条件{test_case}的延迟: {latency:.3f}秒")
            allure.attach(f"组合条件{test_case}的延迟: {latency:.3f}秒", "性能指标")

    @allure.title("不同排序方式的性能测试")
    def test_sort_performance(self, Headers):
        sort_cases = [
            {"sort_by": "create_time", "sort_order": "asc"},
            {"sort_by": "create_time", "sort_order": "desc"},
            {"sort_by": "update_time", "sort_order": "asc"},
            {"sort_by": "update_time", "sort_order": "desc"},
            {"sort_by": "name", "sort_order": "asc"},
            {"sort_by": "name", "sort_order": "desc"}
        ]
        
        for sort_case in sort_cases:
            latency, result = self.measure_latency(self.client.GetOperatorList, sort_case, Headers)
            assert result[0] == 200
            print(f"排序方式{sort_case}的延迟: {latency:.3f}秒")
            allure.attach(f"排序方式{sort_case}的延迟: {latency:.3f}秒", "性能指标")

    @allure.title("大数据量分页的性能测试")
    def test_large_data_pagination_performance(self, Headers):
        pages = [1, 5, 10]
        for page in pages:
            data = {"page": page, "page_size": 10}
            latency, result = self.measure_latency(self.client.GetOperatorList, data, Headers)
            assert result[0] == 200
            print(f"页码{page}的延迟: {latency:.3f}秒")
            allure.attach(f"页码{page}的延迟: {latency:.3f}秒", "性能指标")

    @allure.title("并发获取列表的性能测试")
    def test_concurrent_performance(self, Headers):
        """测试不同并发场景下的性能"""
        # 定义不同的并发场景
        concurrent_scenarios = [
            {
                "name": "基础并发",
                "workers": 10,
                "test_cases": [{}]
            },
            {
                "name": "中等并发",
                "workers": 50,
                "test_cases": [{}, {"page_size": 50}]
            },
            {
                "name": "高并发",
                "workers": 100,
                "test_cases": [{}, {"page_size": 50}, {"status": "published"}, {"category": "data_process"}]
            },
            {
                "name": "混合并发",
                "workers": 80,
                "test_cases": [
                    {},
                    {"page_size": 50},
                    {"status": "published"},
                    {"category": "data_process"},
                    {"sort_by": "update_time", "sort_order": "desc"},
                    {"status": "published", "category": "data_process", "page_size": 50}
                ]
            }
        ]

        for scenario in concurrent_scenarios:
            print(f"\n开始{scenario['name']}测试，并发数: {scenario['workers']}")
            latencies = []
            
            with ThreadPoolExecutor(max_workers=scenario['workers']) as executor:
                # 每个测试用例执行多次
                futures = []
                for test_case in scenario['test_cases']:
                    for _ in range(3):  # 每个测试用例执行3次
                        futures.append(executor.submit(self.client.GetOperatorList, test_case, Headers))
                
                for i, future in enumerate(futures):
                    latency, result = self.measure_latency(lambda: future.result())
                    latencies.append(latency)
                    assert result[0] == 200
                    print(f"第{i+1}次请求耗时: {latency:.3f}秒")
            
            # 计算统计信息
            avg_latency = statistics.mean(latencies)
            max_latency = max(latencies)
            min_latency = min(latencies)
            std_dev = statistics.stdev(latencies) if len(latencies) > 1 else 0
            
            print(f"\n{scenario['name']}测试统计:")
            print(f"平均耗时: {avg_latency:.3f}秒")
            print(f"最大耗时: {max_latency:.3f}秒")
            print(f"最小耗时: {min_latency:.3f}秒")
            print(f"标准差: {std_dev:.3f}秒")
            print(f"总请求数: {len(latencies)}")
            
            # 记录到allure报告
            allure.attach(
                f"{scenario['name']}测试统计:\n" +
                f"并发数: {scenario['workers']}\n" +
                f"平均耗时: {avg_latency:.3f}秒\n" +
                f"最大耗时: {max_latency:.3f}秒\n" +
                f"最小耗时: {min_latency:.3f}秒\n" +
                f"标准差: {std_dev:.3f}秒\n" +
                f"总请求数: {len(latencies)}",
                "性能统计",
                allure.attachment_type.TEXT
            )

    @allure.title("长时间运行的性能测试")
    def test_long_running_performance(self, Headers):
        """测试长时间运行获取算子列表的性能（10线程并发）"""
        test_duration = 300  # 测试持续1小时
        interval = 30  # 每30秒执行一轮并发请求
        concurrent_size = 10  # 使用10个线程并发
        start_time = time.time()
        all_latencies = []
        
        # 定义不同的测试场景
        test_cases = [
            {},  # 默认参数
            {"page_size": 50},  # 较大分页
            {"status": "published"},  # 状态过滤
            {"category": "data_process"},  # 分类过滤
            {"sort_by": "update_time", "sort_order": "desc"},  # 排序
            {"status": "published", "category": "data_process", "page_size": 50},  # 组合条件
            {"page": 2, "page_size": 20},  # 分页
            {"status": "offline"},  # 下架状态
            {"category": "data_store"},  # 另一个分类
            {"sort_by": "create_time", "sort_order": "asc"}  # 另一种排序
        ]
        
        def get_operator_list(test_case):
            return self.measure_latency(self.client.GetOperatorList, test_case, Headers)
        
        while time.time() - start_time < test_duration:
            round_latencies = []
            
            # 每轮使用10个线程并发请求
            with ThreadPoolExecutor(max_workers=concurrent_size) as executor:
                futures = [executor.submit(get_operator_list, test_case) 
                          for test_case in test_cases]
                
                for future in futures:
                    latency, result = future.result()
                    assert result[0] == 200
                    round_latencies.append(latency)
            
            # 计算本轮并发的统计数据
            round_avg_latency = statistics.mean(round_latencies)
            round_max_latency = max(round_latencies)
            round_min_latency = min(round_latencies)
            
            current_time = time.time() - start_time
            print(f"时间点: {current_time:.0f}秒\n" +
                         f"本轮平均延迟: {round_avg_latency:.3f}秒\n" +
                         f"本轮最大延迟: {round_max_latency:.3f}秒\n" +
                         f"本轮最小延迟: {round_min_latency:.3f}秒")
            allure.attach(f"时间点: {current_time:.0f}秒\n" +
                         f"本轮平均延迟: {round_avg_latency:.3f}秒\n" +
                         f"本轮最大延迟: {round_max_latency:.3f}秒\n" +
                         f"本轮最小延迟: {round_min_latency:.3f}秒", 
                         "长时间运行性能")
            
            all_latencies.extend(round_latencies)
            time.sleep(interval)
        
        # 计算整体统计数据
        avg_latency = statistics.mean(all_latencies)
        max_latency = max(all_latencies)
        min_latency = min(all_latencies)
        std_dev = statistics.stdev(all_latencies) if len(all_latencies) > 1 else 0
        print(f"长时间运行平均延迟: {avg_latency:.3f}秒\n最大延迟: {max_latency:.3f}秒\n最小延迟: {min_latency:.3f}秒\n标准差: {std_dev:.3f}秒\n总执行次数: {len(all_latencies)}\n总执行轮数: {len(all_latencies) // concurrent_size}")
        allure.attach(f"长时间运行统计数据:\n" +
                     f"平均延迟: {avg_latency:.3f}秒\n" +
                     f"最大延迟: {max_latency:.3f}秒\n" +
                     f"最小延迟: {min_latency:.3f}秒\n" +
                     f"标准差: {std_dev:.3f}秒\n" +
                     f"总执行次数: {len(all_latencies)}\n" +
                     f"总执行轮数: {len(all_latencies) // concurrent_size}", 
                     "长时间运行统计") 