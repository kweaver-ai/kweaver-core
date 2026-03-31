# -*- coding:UTF-8 -*-

import pytest
import allure
import time
import statistics
import json
import yaml
from concurrent.futures import ThreadPoolExecutor

from common.get_content import GetContent
from lib.operator import Operator

@allure.feature("算子注册与管理性能测试：注册算子")
class TestRegisterOperatorPerformance:
    client = Operator()
    
    def measure_latency(self, func, *args, **kwargs):
        """测量接口调用延迟"""
        start_time = time.time()
        result = func(*args, **kwargs)
        end_time = time.time()
        return end_time - start_time, result

    def prepare_single_operator_data(self, category):
        """准备单个算子数据"""
        filepath = "./resource/openapi/compliant/test3.yaml"
        api_data = GetContent(filepath).yamlfile()
        
        return {
            "data": str(api_data),
            "operator_metadata_type": "openapi",
            "operator_info": {
                "category": category
            }
        }

    def prepare_batch_operator_data(self, category):
        """准备批量算子数据"""
        filepath = "./resource/openapi/compliant/test0.json"
        api_data = GetContent(filepath).jsonfile()
        
        return {
            "data": str(api_data),
            "operator_metadata_type": "openapi",
            "operator_info": {
                "category": category
            }
        }

    @allure.title("单个算子注册性能测试")
    def test_single_operator_registration_performance(self, Headers):
        """测试单个算子注册的性能"""
        categories = [
            "other_category", "data_process", "data_transform", 
            "data_store", "data_analysis", "data_query", 
            "data_extract", "data_split", "model_train"
        ]
        
        latencies = []
        for i in range(100):  # 测试100次单个注册
            category = categories[i % len(categories)]
            data = self.prepare_single_operator_data(category)
            
            latency, result = self.measure_latency(self.client.RegisterOperator, data, Headers)
            assert result[0] == 200
            latencies.append(latency)

            print(f"第{i+1}次注册延迟: {latency:.3f}秒")
            
            allure.attach(f"第{i+1}次注册延迟: {latency:.3f}秒", "单次注册性能")
        
        avg_latency = statistics.mean(latencies)
        max_latency = max(latencies)
        min_latency = min(latencies)
        print(f"平均延迟: {avg_latency:.3f}秒\n最大延迟: {max_latency:.3f}秒\n最小延迟: {min_latency:.3f}秒")

        allure.attach(f"平均延迟: {avg_latency:.3f}秒\n最大延迟: {max_latency:.3f}秒\n最小延迟: {min_latency:.3f}秒", 
                     "单个注册统计数据")

    @allure.title("批量算子注册性能测试")
    def test_batch_operator_registration_performance(self, Headers):
        """测试批量算子注册的性能（每次最多10个）"""
        categories = [
            "other_category", "data_process", "data_transform", 
            "data_store", "data_analysis", "data_query", 
            "data_extract", "data_split", "model_train"
        ]
        
        latencies = []
        for i in range(100):  # 测试100次批量注册
            category = categories[i % len(categories)]
            data = self.prepare_batch_operator_data(category)
            
            latency, result = self.measure_latency(self.client.RegisterOperator, data, Headers)
            assert result[0] == 200
            latencies.append(latency)
            
            print(f"第{i+1}次注册延迟: {latency:.3f}秒")

            allure.attach(f"第{i+1}次注册延迟: {latency:.3f}秒", "单次注册性能")
        
        avg_latency = statistics.mean(latencies)
        max_latency = max(latencies)
        min_latency = min(latencies)

        print(f"平均延迟: {avg_latency:.3f}秒\n最大延迟: {max_latency:.3f}秒\n最小延迟: {min_latency:.3f}秒")

        allure.attach(f"平均延迟: {avg_latency:.3f}秒\n最大延迟: {max_latency:.3f}秒\n最小延迟: {min_latency:.3f}秒", 
                     "单个注册统计数据")

    @allure.title("并发算子注册性能测试")
    def test_concurrent_operator_registration_performance(self, Headers):
        """测试并发注册算子的性能"""
        categories = [
            "other_category", "data_process", "data_transform", 
            "data_store", "data_analysis", "data_query", 
            "data_extract", "data_split", "model_train"
        ]
        
        concurrent_sizes = [5, 10, 20, 50, 100]  # 测试不同的并发数
        for concurrent_size in concurrent_sizes:
            latencies = []
            
            def register_operator(index):
                category = categories[index % len(categories)]
                # 使用批量算子注册文件
                data = self.prepare_batch_operator_data(category)
                return self.measure_latency(self.client.RegisterOperator, data, Headers)
            
            with ThreadPoolExecutor(max_workers=concurrent_size) as executor:
                futures = [executor.submit(register_operator, i) 
                          for i in range(concurrent_size)]
                
                for future in futures:
                    latency, result = future.result()
                    assert result[0] == 200
                    latencies.append(latency)
            
            avg_latency = statistics.mean(latencies)
            max_latency = max(latencies)
            min_latency = min(latencies)

            print(f"并发数{concurrent_size}平均延迟: {avg_latency:.3f}秒\n最大延迟: {max_latency:.3f}秒\n最小延迟: {min_latency:.3f}秒")
            allure.attach(f"并发数{concurrent_size}统计数据:\n" +
                         f"平均延迟: {avg_latency:.3f}秒\n" +
                         f"最大延迟: {max_latency:.3f}秒\n" +
                         f"最小延迟: {min_latency:.3f}秒", 
                         "并发注册统计")

    @allure.title("长时间运行注册性能测试")
    def test_long_running_registration_performance(self, Headers):
        """测试长时间运行注册算子的性能稳定性（10线程并发）"""
        categories = [
            "other_category", "data_process", "data_transform", 
            "data_store", "data_analysis", "data_query", 
            "data_extract", "data_split", "model_train"
        ]
        
        test_duration = 300  # 测试持续1小时
        interval = 30  # 每30秒执行一轮并发注册
        concurrent_size = 10  # 使用10个线程并发
        start_time = time.time()
        all_latencies = []
        index = 0
        
        def register_operator(idx):
            category = categories[idx % len(categories)]
            data = self.prepare_batch_operator_data(category)
            return self.measure_latency(self.client.RegisterOperator, data, Headers)
        
        while time.time() - start_time < test_duration:
            round_latencies = []
            
            # 每轮使用10个线程并发注册
            with ThreadPoolExecutor(max_workers=concurrent_size) as executor:
                futures = [executor.submit(register_operator, index + i) 
                          for i in range(concurrent_size)]
                
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
            index += concurrent_size
            time.sleep(interval)
        
        # 计算整体统计数据
        avg_latency = statistics.mean(all_latencies)
        max_latency = max(all_latencies)
        min_latency = min(all_latencies)
        std_dev = statistics.stdev(all_latencies) if len(all_latencies) > 1 else 0
        
        print(f"长时间运行平均延迟: {avg_latency:.3f}秒\n最大延迟: {max_latency:.3f}秒\n最小延迟: {min_latency:.3f}秒")

        allure.attach(f"长时间运行统计数据:\n" +
                     f"平均延迟: {avg_latency:.3f}秒\n" +
                     f"最大延迟: {max_latency:.3f}秒\n" +
                     f"最小延迟: {min_latency:.3f}秒\n" +
                     f"标准差: {std_dev:.3f}秒\n" +
                     f"总执行次数: {len(all_latencies)}\n" +
                     f"总执行轮数: {len(all_latencies) // concurrent_size}", 
                     "长时间运行统计") 