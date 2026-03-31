# -*- coding:UTF-8 -*-

import pytest
import allure
import json
import time

from lib.data_agent import AgentApp, AgentFactory

@allure.feature("智能体对话功能测试")
class TestAgentChat:
    """
    智能体对话功能测试类
    测试从data/data-agent目录导入的智能体的对话功能
    """

    def validate_chat_response(self, response_data, agent_info=None, query_index=None, is_intervention_test=False, expected_validations=None):
        """
        验证对话响应数据
        :param response_data: 对话响应数据
        :param agent_info: 智能体信息，包含config_intervention配置
        :param query_index: 查询索引（用于中断测试判断奇偶次）
        :param is_intervention_test: 是否为中断测试
        :param expected_validations: 预期验证配置列表
        """
        allure.attach(json.dumps(response_data, ensure_ascii=False, indent=2), name="chat_response")

        # a. 状态码是200，否则用例失败
        assert response_data.get("error") is None, f"对话返回错误: {response_data.get('error')}"

        # b. error为空，否则用例失败
        assert response_data.get("error") is None, f"对话返回错误: {response_data.get('error')}"

        # c. message不为空，否则用例失败
        message = response_data.get("message")
        assert message is not None, "对话返回的message为空"

        # d. 尝试从不同路径获取progress数据
        content = message.get("content", {})
        middle_answer = content.get("middle_answer", {})
        progress = middle_answer.get("progress", [])

        # 如果没有progress，尝试从其他路径获取
        if not progress:
            # 某些智能体的数据结构可能不同，尝试其他可能的路径
            # 这里我们暂时允许progress为空，但会记录警告
            allure.attach("progress为空，可能是不同数据格式", name="progress_empty_warning")
            # 不再强制要求progress非空，因为有些智能体使用不同的响应格式

        # h. 验证ext.ask字段是否正确配置
        ext = message.get("ext", {})
        ask = ext.get("ask", "")

        # 获取智能体的干预配置
        has_intervention_enabled = False
        config_intervention = []
        if agent_info:
            config_intervention = agent_info.get("config_intervention", [])
            # 检查是否有任何技能开启了中断
            has_intervention_enabled = any(config_intervention) if config_intervention else False

        allure.attach(f"智能体中断配置: {config_intervention if agent_info and 'config_intervention' in agent_info else 'N/A'}", name="intervention_config")
        allure.attach(f"ask字段内容: '{ask}'", name="ask_field")

        if has_intervention_enabled:
            # 中断测试的特殊处理
            if is_intervention_test and query_index is not None:
                # 中断测试中，奇数次查询需要ask字段非空，偶数次查询ask字段应该为空
                if query_index % 2 == 1:  # 奇数次（第1、3、5...次）
                    assert ask is not None and ask != "", f"中断测试第{query_index}次查询时，ask字段不能为空，但当前值为: '{ask}'"
                    allure.attach(f"中断测试第{query_index}次查询，ask字段校验通过（非空）", name="ask_validation")
                else:  # 偶数次（第2、4、6...次）
                    assert ask is None or ask == "", f"中断测试第{query_index}次查询时，ask字段应该为空，但当前值为: '{ask}'"
                    allure.attach(f"中断测试第{query_index}次查询，ask字段校验通过（为空）", name="ask_validation")
            else:
                # 非中断测试的普通情况：有技能开启中断时，ask字段不能为空
                assert ask is not None and ask != "", f"智能体有技能开启中断时，ask字段不能为空，但当前值为: '{ask}'"
                allure.attach("智能体有技能开启中断，ask字段校验通过（非空）", name="ask_validation")
        else:
            # 如果所有技能都关闭中断，ask字段应该是空的
            assert ask is None or ask == "", f"智能体所有技能都关闭中断时，ask字段应该为空，但当前值为: '{ask}'"
            allure.attach("智能体所有技能都关闭中断，ask字段校验通过（为空）", name="ask_validation")

        # f. 从最后一个开始遍历message.content.middle_answer.progress数组
        # 如果stage是llm并且answer不为空，如果为空，再检测下一个
        # 如果遍历完answer都是空，则用例失败
        found_llm_answer = False
        for item in reversed(progress):
            if item.get("stage") == "llm":
                answer = item.get("answer")
                if answer and answer != "":
                    found_llm_answer = True
                    break

        assert found_llm_answer, "未找到有效的LLM回答"

        # g. 从最后一个开始遍历message.content.middle_answer.progress数组
        # 如果stage是llm并且answer不为空，但如果包含fail to call LLM，则用例失败
        for item in reversed(progress):
            if item.get("stage") == "llm":
                answer = item.get("answer", "")
                if answer and isinstance(answer, str) and "fail to call LLM" in answer:
                    pytest.fail(f"LLM调用失败，包含错误信息: {answer}")

        # 新增：执行详细的progress验证
        if expected_validations:
            # 如果progress为空，尝试从final_answer中构造progress数据
            if not progress:
                final_answer = content.get("final_answer", {})
                if final_answer and "answer" in final_answer:
                    # 从final_answer构造一个模拟的progress项用于验证
                    mock_progress = [{
                        "id": "mock_final_answer",
                        "stage": "llm",
                        "agent_name": "LLM",
                        "status": "completed",
                        "answer": final_answer["answer"]
                    }]
                    allure.attach("progress为空，从final_answer构造mock_progress用于验证", name="mock_progress_creation")
                    self.validate_progress_details(mock_progress, expected_validations)
                else:
                    allure.attach("无法获取到用于验证的答案数据", name="no_data_for_validation")
            else:
                self.validate_progress_details(progress, expected_validations)

    def validate_progress_details(self, progress, expected_validations):
        """
        验证progress数组的详细信息
        :param progress: progress数组
        :param expected_validations: 预期验证配置列表
        """
        allure.attach(f"开始详细progress验证，预期验证数量: {len(expected_validations)}", name="progress_validation_start")

        for validation in expected_validations:
            with allure.step(f"验证progress条件: {validation}"):
                self.validate_single_progress_item(progress, validation)

    def validate_single_progress_item(self, progress, validation):
        """
        验证单个progress项的条件
        :param progress: progress数组
        :param validation: 单个验证条件
        """
        stage = validation.get("stage")
        agent_name = validation.get("agent_name")
        answer_keywords = validation.get("answer_keywords", [])
        answer_failure_keywords = validation.get("answer_failure_keywords", [])
        not_null_fields = validation.get("not_null", [])
        must_exist = validation.get("must_exist", True)
        expected_status = validation.get("status")

        print(f"\n🔍 开始验证progress项:")
        print(f"  - 验证条件: stage={stage}, agent_name={agent_name}")
        print(f"  - 期望状态: {expected_status}")
        print(f"  - 期望关键字: {answer_keywords}")
        print(f"  - 失败关键字: {answer_failure_keywords}")
        print(f"  - 非空字段: {not_null_fields}")
        print(f"  - must_exist: {must_exist}")
        print(f"  - 总progress项数: {len(progress)}")

        found_items = []

        # 查找匹配的progress项
        for i, item in enumerate(progress):
            item_stage = item.get("stage")
            item_agent_name = item.get("agent_name")
            item_id = item.get("id", f"item_{i}")

            stage_match = stage is None or item_stage == stage
            agent_match = agent_name is None or item_agent_name == agent_name

            print(f"  - 检查progress项 {i}: ID={item_id}, stage={item_stage}, agent_name={item_agent_name}")
            print(f"    - stage_match: {stage_match} (期望: {stage})")
            print(f"    - agent_match: {agent_match} (期望: {agent_name})")

            if stage_match and agent_match:
                found_items.append(item)
                print(f"    ✅ 找到匹配项!")

        print(f"  📊 查找结果: 找到 {len(found_items)} 个匹配项")
        allure.attach(f"查找条件: stage={stage}, agent_name={agent_name}, 找到匹配项: {len(found_items)}", name="progress_search_result")

        if must_exist:
            assert found_items, f"未找到匹配的progress项: stage={stage}, agent_name={agent_name}"

        # 验证每个找到的项
        for i, item in enumerate(found_items):
            item_id = item.get('id', f'found_item_{i}')
            print(f"  🔧 开始验证第{i+1}个匹配项: {item_id}")

            with allure.step(f"验证progress项 ID: {item_id}"):
                # 验证状态
                if expected_status:
                    actual_status = item.get("status")
                    print(f"    - 期望状态: {expected_status}")
                    print(f"    - 实际状态: {actual_status}")
                    assert actual_status == expected_status, f"状态不匹配: 期望 {expected_status}, 实际 {actual_status}"
                    print(f"    ✅ 状态验证通过")
                    allure.attach(f"状态验证通过: {actual_status}", name="status_validation")
                else:
                    print(f"    - 跳过状态验证（未配置期望状态）")

                # 验证非空字段（仅对skill stage）
                if stage == "skill" and not_null_fields:
                    print(f"  🔧 开始验证非空字段（针对skill stage）")
                    self._validate_not_null_fields(item, not_null_fields)

                # 收集answer内容用于关键字验证
                answer = item.get("answer", "")
                answer_text = self._extract_answer_text(answer)

                # 验证失败关键字（OR逻辑：只要匹配到一个就失败）
                if answer_failure_keywords:
                    print(f"  🔧 开始失败关键字验证（OR逻辑：任一匹配即失败）")
                    self._validate_failure_keywords(answer_text, answer_failure_keywords, item)

        # 收集所有匹配stage的answer内容用于关键字验证
        combined_answer_text = self._collect_combined_answers(found_items, stage)

        # 验证成功关键字（OR逻辑：只要有一项存在就算成功）
        if answer_keywords:
            print(f"  🔧 开始成功关键字验证（OR逻辑：任一关键字存在即通过）")
            self._validate_success_keywords(combined_answer_text, answer_keywords, stage)

        print(f"  ✅ validate_single_progress_item 完成\n")

    def _extract_answer_text(self, answer):
        """
        从answer中提取文本内容
        :param answer: answer内容，可能是字符串、字典或列表
        :return: 提取的文本内容
        """
        if isinstance(answer, str):
            return answer
        elif isinstance(answer, dict) and "text" in answer:
            return answer["text"]
        elif isinstance(answer, list):
            return str(answer)
        else:
            return str(answer)

    def _collect_combined_answers(self, found_items, stage):
        """
        收集所有匹配stage的answer内容
        :param found_items: 匹配的progress项列表
        :param stage: stage类型
        :return: 合并的answer文本
        """
        combined_answer_text = ""
        answers = []

        for item in found_items:
            if item.get("stage") == stage:
                answer = item.get("answer", "")
                answer_text = self._extract_answer_text(answer)
                answers.append(answer_text)
                print(f"    - 收集到 {stage} answer: {str(answer_text)[:100]}...")

        print(f"  📝 收集到 {len(answers)} 个 {stage} answer")

        # 合并所有answer内容
        for i, answer_text in enumerate(answers):
            if combined_answer_text:
                combined_answer_text += "\n\n" + answer_text  # 用换行分隔不同的answer
            else:
                combined_answer_text = answer_text

        print(f"  📄 合并后的总answer长度: {len(combined_answer_text)} 字符")
        if combined_answer_text:
            print(f"  📄 合并后的总answer预览: {combined_answer_text[:300]}...")
        allure.attach(f"合并后的answer内容: {combined_answer_text[:500]}...", name="combined_answer_content")

        return combined_answer_text

    def _validate_not_null_fields(self, item, not_null_fields):
        """
        验证指定字段不为空
        :param item: progress项
        :param not_null_fields: 需要验证非空的字段列表
        """
        print(f"    - 开始验证非空字段: {not_null_fields}")

        for field_path in not_null_fields:
            try:
                # 解析字段路径，支持嵌套访问，如 "answer['choices'][0]['message']"
                value = self._get_nested_value(item, field_path)

                if value is None or (isinstance(value, str) and value.strip() == "") or (isinstance(value, (list, dict)) and len(value) == 0):
                    print(f"    ❌ 字段 '{field_path}' 为空或不存在，值: {value}")
                    allure.attach(f"字段 '{field_path}' 验证失败，值为空: {value}", name="not_null_validation_failed")
                    assert False, f"字段 '{field_path}' 不能为空，但当前值为: {value}"
                else:
                    print(f"    ✅ 字段 '{field_path}' 非空验证通过，值类型: {type(value)}")
                    allure.attach(f"字段 '{field_path}' 验证通过，值: {str(value)[:100]}...", name="not_null_validation_success")

            except Exception as e:
                print(f"    ❌ 字段 '{field_path}' 访问失败: {str(e)}")
                allure.attach(f"字段 '{field_path}' 访问失败: {str(e)}", name="not_null_validation_error")
                assert False, f"字段 '{field_path}' 访问失败: {str(e)}"

    def _get_nested_value(self, obj, field_path):
        """
        获取嵌套对象的值，支持路径如 "answer['choices'][0]['message']"
        :param obj: 源对象
        :param field_path: 字段路径
        :return: 字段值
        """
        current = obj

        # 解析字段路径
        import re
        # 匹配格式：field.subfield['key'][0]['nested_key']
        pattern = r"(\w+)|\['([^']+)'\]|\.(\w+)"
        matches = re.findall(pattern, field_path)

        for match in matches:
            if match[0]:  # 直接属性名
                key = match[0]
            elif match[1]:  # ['key'] 格式
                key = match[1]
            elif match[2]:  # .key 格式
                key = match[2]
            else:
                continue

            if isinstance(current, dict):
                current = current.get(key)
            elif isinstance(current, list) and key.isdigit():
                current = current[int(key)]
            else:
                current = getattr(current, key, None)

            if current is None:
                return None

        return current

    def _validate_failure_keywords(self, answer_text, failure_keywords, item):
        """
        验证失败关键字（OR逻辑：只要匹配到一个就失败）
        :param answer_text: answer文本内容
        :param failure_keywords: 失败关键字列表
        :param item: progress项（用于日志）
        """
        if not answer_text or not failure_keywords:
            print(f"    - 跳过失败关键字验证（answer_text为空或未配置失败关键字）")
            return

        print(f"    - 开始验证失败关键字: {failure_keywords}")
        found_failure_keywords = []

        for keyword in failure_keywords:
            if keyword in answer_text:
                print(f"    ❌ 发现失败关键字 '{keyword}'")
                found_failure_keywords.append(keyword)
                allure.attach(f"发现失败关键字: {keyword}", name="failure_keyword_found")

        if found_failure_keywords:
            print(f"    ❌ 失败关键字验证失败！发现关键字: {found_failure_keywords}")
            allure.attach(f"失败关键字验证失败，发现: {found_failure_keywords}", name="failure_keywords_validation_failed")
            assert False, f"answer中包含失败关键字: {found_failure_keywords}, 实际answer: {answer_text[:500]}..."
        else:
            print(f"    ✅ 失败关键字验证通过，未发现失败关键字")
            allure.attach("失败关键字验证通过", name="failure_keywords_validation_success")

    def _validate_success_keywords(self, combined_answer_text, success_keywords, stage):
        """
        验证成功关键字（OR逻辑：只要有一项存在就算成功）
        :param combined_answer_text: 合并的answer文本
        :param success_keywords: 成功关键字列表
        :param stage: stage类型
        """
        if not success_keywords:
            print(f"  - 跳过成功关键字验证（未配置成功关键字）")
            return

        if not combined_answer_text:
            print(f"  ⚠️ 合并后的answer_text为空，但需要验证成功关键字")
            assert not success_keywords, f"期望验证成功关键字 {success_keywords}，但合并的answer为空"
            return

        print(f"  🔍 开始成功关键字验证（OR逻辑：任一关键字存在即通过）...")
        found_keywords = []
        missing_keywords = []

        for keyword in success_keywords:
            if keyword in combined_answer_text:
                print(f"    ✅ 成功关键字 '{keyword}' 找到")
                found_keywords.append(keyword)
            else:
                print(f"    ❌ 成功关键字 '{keyword}' 缺失")
                missing_keywords.append(keyword)

        print(f"  📊 找到的成功关键字: {found_keywords}")
        print(f"  📊 缺失的成功关键字: {missing_keywords}")

        if found_keywords:
            print(f"  ✅ 成功关键字验证通过！（找到: {found_keywords}）")
            allure.attach(f"成功关键字验证通过（OR逻辑），找到: {found_keywords}", name="success_keywords_validation_success")
        else:
            print(f"  ❌ 成功关键字验证失败！所有关键字都未找到")
            allure.attach(f"成功关键字验证失败，所有关键字都未找到: {success_keywords}", name="success_keywords_validation_failed")
            assert False, f"answer中未找到任何成功关键字: {success_keywords}, 实际answer: {combined_answer_text[:500]}..."

    @pytest.fixture(scope="class")
    def test_cases_data(self, request):
        """
        收集所有需要测试的智能体和查询组合 - 用于真正的参数化测试
        """
        # 获取 AgentImport fixture 的结果
        agent_import = request.getfixturevalue("AgentImport")
        test_cases = []

        if not agent_import:
            return test_cases

        for agent_info in agent_import:
            agent_id = agent_info["agent_id"]
            agent_name = agent_info["agent_name"]
            agent_key = agent_info["agent_key"]
            model_name = agent_info.get("model_name", "Tome-pro")
            config_intervention = agent_info.get("config_intervention", [])
            need_temp_file = agent_info.get("need_temp_file", False)
            temp_file_config = agent_info.get("temp_file_config", [])
            extra_params = agent_info.get("extra_params", [])

            # 获取测试查询和预期结果配置
            test_queries_obj = agent_info.get("test_queries", {})
            single_turn_queries = test_queries_obj.get("single_turn", [])
            multi_turn_queries = test_queries_obj.get("multi_turn", [])
            expected_results = agent_info.get("expected_results", {})
            expected_single_turn = expected_results.get("single_turn", [])
            expected_multi_turn = expected_results.get("multi_turn", [])

            # 跳过无法获取实际ID的冲突智能体
            if agent_id.startswith("CONFLICT_"):
                continue

            # 检查是否为中断测试智能体
            is_intervention_agent = any(config_intervention) if config_intervention else False
            has_duplicate_query = len(multi_turn_queries) >= 2 and multi_turn_queries[0] == multi_turn_queries[1]

            # 处理单轮对话测试用例
            if single_turn_queries:
                for i, query in enumerate(single_turn_queries, 1):
                    # 查找对应的预期结果
                    expected_validations = []
                    for expected in expected_single_turn:
                        if expected.get("query") == query:
                            expected_validations = expected.get("progress_validations", [])
                            break

                    test_case = {
                        "agent_id": agent_id,
                        "agent_name": agent_name,
                        "agent_key": agent_key,
                        "model_name": model_name,
                        "config_intervention": config_intervention,
                        "need_temp_file": need_temp_file,
                        "temp_file_config": temp_file_config,
                        "extra_params": extra_params,
                        "query": query,
                        "query_index": i,
                        "query_type": "single_turn",
                        "test_id": f"{agent_name}_single_turn_query_{i}",
                        "expected_validations": expected_validations
                    }
                    test_cases.append(test_case)

            # 处理多轮对话测试用例
            if multi_turn_queries:
                # 对于中断测试智能体，只在专门的多轮对话测试中执行
                if is_intervention_agent and has_duplicate_query:
                    print(f"  ⚠️ 中断测试智能体 {agent_name} 将在专门的多轮对话测试中执行，不在数据驱动测试中重复")
                    continue

                for i, query in enumerate(multi_turn_queries, 1):
                    # 查找对应的预期结果
                    expected_validations = []
                    for expected in expected_multi_turn:
                        if expected.get("query") == query and expected.get("query_index") == i:
                            expected_validations = expected.get("progress_validations", [])
                            break

                    test_case = {
                        "agent_id": agent_id,
                        "agent_name": agent_name,
                        "agent_key": agent_key,
                        "model_name": model_name,
                        "config_intervention": config_intervention,
                        "need_temp_file": need_temp_file,
                        "temp_file_config": temp_file_config,
                        "extra_params": extra_params,
                        "query": query,
                        "query_index": i,
                        "query_type": "multi_turn",
                        "test_id": f"{agent_name}_multi_turn_query_{i}",
                        "expected_validations": expected_validations
                    }
                    test_cases.append(test_case)

            # 如果既没有单轮也没有多轮查询，使用默认查询
            if not single_turn_queries and not multi_turn_queries:
                test_case = {
                    "agent_id": agent_id,
                    "agent_name": agent_name,
                    "agent_key": agent_key,
                    "model_name": model_name,
                    "config_intervention": config_intervention,
                    "need_temp_file": need_temp_file,
                    "temp_file_config": temp_file_config,
                    "extra_params": extra_params,
                    "query": "你好，请介绍一下你自己",
                    "query_index": 1,
                    "query_type": "default",
                    "test_id": f"{agent_name}_default_query_1",
                    "expected_validations": []
                }
                test_cases.append(test_case)

        # 对测试用例进行排序，确保同一智能体的多轮对话按顺序执行
        # 排序优先级：智能体名称 -> 查询类型 -> 查询索引
        test_cases.sort(key=lambda tc: (
            tc['agent_name'],
            tc['query_type'],  # single_turn 在 multi_turn 之前
            tc['query_index']
        ))

        # 打印测试用例统计信息
        print(f"\n📊 数据驱动测试统计:")
        print(f"收集到的测试用例总数: {len(test_cases)}")
        print(f"涉及的智能体数量: {len(set(tc['agent_name'] for tc in test_cases))}")

        # 按查询类型统计
        query_type_stats = {}
        for tc in test_cases:
            query_type = tc.get('query_type', 'unknown')
            query_type_stats[query_type] = query_type_stats.get(query_type, 0) + 1

        print(f"查询类型分布:")
        for query_type, count in query_type_stats.items():
            print(f"  - {query_type}: {count} 个测试用例")

        for tc in test_cases[:5]:  # 只显示前5个
            print(f"  - {tc['test_id']}: {tc['agent_name']} - {tc['query'][:30]}...")

        if len(test_cases) > 5:
            print(f"  ... 还有 {len(test_cases) - 5} 个测试用例")

        return test_cases

    @pytest.fixture(scope="class")
    def collect_test_cases(self, request):
        """
        为了兼容性保留的 fixture
        """
        return request.getfixturevalue("test_cases_data")

    @pytest.mark.api
    def test_agent_chat_conversation(self, collect_test_cases, AgentHeaders, AgentModelCheck):
        """
        测试智能体对话功能 - 数据驱动
        测试所有从data/data-agent目录导入的智能体
        使用从cfg.json配置文件中读取的测试查询
        每个智能体的每个查询都是独立的测试用例
        """
        if not collect_test_cases:
            pytest.skip("没有测试用例")

        app_client = AgentApp()

        # 统计测试结果
        passed_count = 0
        failed_count = 0
        skipped_count = 0
        total_count = len(collect_test_cases)

        print(f"\n🧪 开始执行数据驱动测试 - 总计 {total_count} 个测试用例")

        # 为了支持多轮对话，按智能体和查询类型分组处理
        agent_conversation_context = {}  # 存储每个智能体的conversation_id

        # 遍历所有测试用例
        for test_case in collect_test_cases:
            agent_id = test_case["agent_id"]
            agent_name = test_case["agent_name"]
            agent_key = test_case["agent_key"]
            model_name = test_case["model_name"]
            config_intervention = test_case["config_intervention"]
            query = test_case["query"]
            query_index = test_case["query_index"]
            query_type = test_case["query_type"]
            test_id = test_case["test_id"]
            expected_validations = test_case.get("expected_validations", [])
            extra_params = test_case.get("extra_params", [])

            # 使用 allure 动态测试ID
            with allure.step(f"测试 {test_id}"):
                test_passed = True
                error_message = ""

                try:
                    allure.attach(f"Testing agent: {agent_name} (ID: {agent_id}, Model: {model_name})", name="agent_test_start")
                    allure.attach(f"Testing query {query_index}: {query}", name=f"test_query_{query_index}")
                    allure.attach(f"Query type: {query_type}", name="query_type")

                    # 构造智能体信息用于验证
                    agent_info = {
                        "agent_id": agent_id,
                        "agent_name": agent_name,
                        "agent_key": agent_key,
                        "model_name": model_name,
                        "config_intervention": config_intervention
                    }

                    # 添加预期验证信息到Allure
                    if expected_validations:
                        allure.attach(f"预期验证配置: {json.dumps(expected_validations, ensure_ascii=False)}", name="expected_validations")

                    # 构造对话请求数据
                    chat_request_data = {
                        "agent_id": agent_id,
                        "agent_version": "v0",
                        "query": query,
                        "chat_option":{
                            "is_need_history": True,
                            "is_need_doc_retrival_post_process": True,
                            "is_need_progress": True,
                            "enable_dependency_cache": True
                        },
                        "stream": False,
                        "inc_stream": False,
                        "executor_version": "v2"
                    }

                    # 多轮对话处理：如果是同一智能体的多轮对话，传递conversation_id
                    agent_context_key = f"{agent_id}_{agent_name}"
                    if query_type == "multi_turn" and agent_context_key in agent_conversation_context:
                        conversation_id = agent_conversation_context[agent_context_key]
                        chat_request_data["conversation_id"] = conversation_id
                        allure.attach(f"多轮对话，包含conversation_id: {conversation_id}", name="conversation_id")
                    elif query_type == "multi_turn":
                        allure.attach("多轮对话第一轮，不包含conversation_id", name="conversation_id_note")

                    # 如果开启了临时区功能，添加temp_files参数
                    if test_case.get("need_temp_file") and test_case.get("temp_file_config"):
                        chat_request_data["temp_files"] = test_case["temp_file_config"]
                        allure.attach(f"包含临时文件配置: {json.dumps(test_case['temp_file_config'], ensure_ascii=False)}", name="temp_files_config")

                    # 处理额外参数（extra_params）
                    if extra_params and len(extra_params) > 0:
                        # extra_params 是一个数组，每个元素是一个包含多个参数的对象
                        # 我们需要将第一个对象中的所有参数添加到 chat_request_data 中
                        for param_obj in extra_params:
                            if isinstance(param_obj, dict):
                                for param_name, param_config in param_obj.items():
                                    if isinstance(param_config, dict) and "value" in param_config:
                                        # 使用配置的值
                                        chat_request_data[param_name] = param_config["value"]
                                        allure.attach(f"包含额外参数 {param_name}: {json.dumps(param_config['value'], ensure_ascii=False)}", name=f"extra_param_{param_name}")
                                    else:
                                        # 如果没有 value 字段，直接使用参数配置
                                        chat_request_data[param_name] = param_config
                                        allure.attach(f"包含额外参数 {param_name}: {json.dumps(param_config, ensure_ascii=False)}", name=f"extra_param_{param_name}")
                        allure.attach(f"包含 {len(extra_params)} 个额外参数配置组", name="extra_params_summary")

                    # 发起对话请求，设置5分钟超时（避免单个测试用例太长）
                    start_time = time.time()
                    result = app_client.ChatCompletion(agent_id, chat_request_data, AgentHeaders, timeout=300)
                    end_time = time.time()
                    duration = end_time - start_time

                    allure.attach(f"对话耗时: {duration:.2f}秒", name="chat_duration")

                    # h. 测试对话接口设置超时时间，如果5分钟还没有返回值，则用例失败
                    if duration >= 300:
                        error_message = f"对话请求超时（5分钟）"
                        test_passed = False

                    # 检查服务器错误状态
                    elif result[0] in [502, 503, 504] or (isinstance(result[1], str) and "Connection refused" in result[1]):
                        # 服务器不可用错误，跳过测试
                        allure.attach("服务器不可用，跳过对话验证", name="skip_validation")
                        error_message = "服务器不可用"
                        test_passed = False
                        skipped_count += 1

                    elif result[0] == 500:
                        # 服务器内部错误，记录为测试失败
                        allure.attach(f"服务器内部错误，状态码: {result[0]}", name="server_error")
                        error_details = result[1] if isinstance(result[1], dict) else {}
                        if error_details.get("error_code"):
                            allure.attach(f"错误代码: {error_details.get('error_code')}", name="error_code")
                        if error_details.get("error_details"):
                            allure.attach(f"错误详情: {error_details.get('error_details')}", name="error_details")
                        error_message = f"服务器内部错误，状态码: {result[0]}，错误代码: {error_details.get('error_code', 'Unknown')}"
                        test_passed = False

                    else:
                        # 验证响应状态码
                        if result[0] != 200:
                            error_message = f"对话请求失败，状态码: {result[0]}"
                            test_passed = False
                        else:
                            try:
                                # 验证响应数据
                                self.validate_chat_response(result[1], agent_info, query_index=query_index, expected_validations=expected_validations)
                                allure.attach(f"智能体 {agent_name} 查询 {query_index} 测试通过", name=f"test_success_query_{query_index}")

                                # 多轮对话：存储conversation_id供后续查询使用
                                if query_type == "multi_turn" and result[0] == 200:
                                    response_conversation_id = result[1].get("conversation_id")
                                    if response_conversation_id:
                                        agent_conversation_context[agent_context_key] = response_conversation_id
                                        allure.attach(f"存储conversation_id: {response_conversation_id} for {agent_context_key}", name="store_conversation_id")
                            except AssertionError as e:
                                error_message = str(e)
                                test_passed = False

                    # 更新统计
                    if test_passed:
                        passed_count += 1
                        allure.attach("✅ 测试通过", name="test_result")
                    else:
                        failed_count += 1
                        allure.attach(f"❌ 测试失败: {error_message}", name="test_result")

                except Exception as e:
                    failed_count += 1
                    error_message = f"测试执行异常: {str(e)}"
                    test_passed = False
                    allure.attach(f"❌ 测试异常: {error_message}", name="test_result")

                print(f"  {test_id}: {'✅ PASS' if test_passed else '❌ FAIL'}{' - ' + error_message if error_message else ''}")

        # 输出最终统计结果
        print(f"\n📊 数据驱动测试最终统计:")
        print(f"总测试用例数: {total_count}")
        print(f"通过: {passed_count}")
        print(f"失败: {failed_count}")
        print(f"跳过: {skipped_count}")
        print(f"成功率: {(passed_count/total_count*100):.1f}%")

        # 如果有失败的测试用例，最终测试结果为失败
        if failed_count > 0:
            allure.attach(f"测试完成 - 成功: {passed_count}, 失败: {failed_count}, 跳过: {skipped_count}", name="final_summary")
            # 使用 pytest.fail 标记测试失败
            pytest.fail(f"有 {failed_count} 个测试用例失败，总共 {total_count} 个测试用例")

    @pytest.mark.api
    def test_agent_multi_turn_conversation(self, AgentImport, AgentHeaders, AgentModelCheck):
        """
        测试智能体多轮对话功能
        使用从cfg.json配置文件中读取的多轮测试查询
        如果multi_turn查询为空列表，则跳过测试
        支持中断对话测试：当智能体开启中断且多轮查询使用相同问题时，测试中断流程
        """
        if not AgentImport:
            pytest.skip("没有导入的智能体")

        app_client = AgentApp()

        # 优先寻找开启中断且有多轮查询的智能体，否则寻找第一个可用的智能体
        available_agent = None
        multi_turn_queries = None

        # 首先寻找中断测试候选智能体
        for agent_info in AgentImport:
            if not agent_info["agent_id"].startswith("CONFLICT_"):
                config_intervention = agent_info.get("config_intervention", [])
                multi_turn = agent_info.get("test_queries", {}).get("multi_turn", [])
                has_intervention = any(config_intervention) if config_intervention else False
                has_same_query_twice = len(multi_turn) >= 2 and multi_turn[0] == multi_turn[1]

                # 如果是中断测试场景，优先选择
                if has_intervention and has_same_query_twice and len(multi_turn) > 0:
                    available_agent = agent_info
                    multi_turn_queries = multi_turn
                    allure.attach(f"Found intervention test agent: {agent_info.get('agent_name')}", name="intervention_agent_found")
                    break

        # 如果没有找到中断测试智能体，寻找第一个有多轮查询的智能体
        if not available_agent:
            for agent_info in AgentImport:
                if not agent_info["agent_id"].startswith("CONFLICT_"):
                    multi_turn = agent_info.get("test_queries", {}).get("multi_turn", [])
                    if len(multi_turn) > 0:  # 只有当有多轮查询时才选择
                        available_agent = agent_info
                        multi_turn_queries = multi_turn
                        allure.attach(f"Found regular multi-turn agent: {agent_info.get('agent_name')}", name="regular_agent_found")
                        break

        if not available_agent:
            pytest.skip("没有可用的智能体进行多轮对话测试")

        # 检查multi_turn查询是否为空
        if not multi_turn_queries or len(multi_turn_queries) == 0:
            agent_name = available_agent["agent_name"]
            allure.attach(f"智能体 {agent_name} 没有配置多轮对话查询，跳过测试", name="skip_multi_turn_empty_queries")
            pytest.skip(f"智能体 {agent_name} 没有配置多轮对话查询")

        agent_id = available_agent["agent_id"]
        agent_name = available_agent["agent_name"]
        test_queries = multi_turn_queries

        # 检查是否开启了中断功能
        config_intervention = available_agent.get("config_intervention", [])
        has_intervention_enabled = any(config_intervention) if config_intervention else False

        # 检查是否为中断测试场景：开启了中断且有两个相同的查询
        is_intervention_test = (
            has_intervention_enabled and
            len(test_queries) >= 2 and
            test_queries[0] == test_queries[1]
        )

        allure.attach(f"Testing multi-turn conversation for agent: {agent_name} with {len(test_queries)} queries", name="multi_turn_test_start")
        allure.attach(f"智能体中断配置: {config_intervention}", name="agent_intervention_config")
        allure.attach(f"是否为中断测试: {is_intervention_test}", name="is_intervention_test")

        conversation_id = None
        first_response_assistant_message_id = None
        intervention_info = None

              # 获取预期结果配置
        expected_results = available_agent.get("expected_results", {})
        expected_multi_turn = expected_results.get("multi_turn", [])

        # 遍历所有多轮测试查询
        for i, query in enumerate(test_queries, 1):
            # 查找对应的预期结果
            expected_validations = []
            for expected in expected_multi_turn:
                if expected.get("query") == query and expected.get("query_index") == i:
                    expected_validations = expected.get("progress_validations", [])
                    break

            allure.attach(f"Multi-turn conversation {i}: {query}", name=f"multi_turn_query_{i}")

            # 添加预期验证信息到Allure
            if expected_validations:
                allure.attach(f"本轮预期验证配置: {json.dumps(expected_validations, ensure_ascii=False)}", name="expected_validations_round_{i}")

            # 构造对话请求数据
            chat_request_data = {
                "agent_id": agent_id,
                "agent_version": "v0",
                "chat_option":{
                    "is_need_history": True,
                    "is_need_doc_retrival_post_process": True,
                    "is_need_progress": True,
                    "enable_dependency_cache": True
                },                
                "query": query,
                "stream": False,
                "inc_stream": False,
                "executor_version": "v2"
            }

            # 如果开启了临时区功能，添加temp_files参数
            if available_agent.get("need_temp_file") and available_agent.get("temp_file_config"):
                chat_request_data["temp_files"] = available_agent["temp_file_config"]
                allure.attach(f"包含临时文件配置: {json.dumps(available_agent['temp_file_config'], ensure_ascii=False)}", name="temp_files_config")

            # 中断测试特殊处理：第二轮查询时添加中断相关参数
            if is_intervention_test and i == 2:
                # 添加conversation_id
                if conversation_id:
                    chat_request_data["conversation_id"] = conversation_id
                    allure.attach(f"包含conversation_id: {conversation_id}", name=f"conversation_id_round_{i}")

                # 添加中断相关的参数（从第一次响应中获取）
                if intervention_info and first_response_assistant_message_id:
                    chat_request_data["tool"] = {
                        "tool_args": intervention_info.get("tool_args", []),
                        "session_id": intervention_info.get("session_id"),
                        "tool_name": intervention_info.get("tool_name")
                    }
                    chat_request_data["interrupted_assistant_message_id"] = first_response_assistant_message_id

                    allure.attach(f"包含中断参数 - session_id: {intervention_info.get('session_id')}", name="intervention_params")
                    allure.attach(f"包含中断参数 - tool_name: {intervention_info.get('tool_name')}", name="intervention_tool_name")
                    allure.attach(f"包含interrupted_assistant_message_id: {first_response_assistant_message_id}", name="interrupted_assistant_message_id")
            else:
                # 普通多轮对话：如果不是第一轮，添加conversation_id
                if i > 1 and conversation_id:
                    chat_request_data["conversation_id"] = conversation_id
                    allure.attach(f"包含conversation_id: {conversation_id}", name=f"conversation_id_round_{i}")

            # 发起对话请求
            result = app_client.ChatCompletion(agent_id, chat_request_data, AgentHeaders, timeout=1200)

            # 检查服务器错误状态
            if result[0] in [502, 503, 504] or (isinstance(result[1], str) and "Connection refused" in result[1]):
                # 服务器不可用错误，跳过测试
                allure.attach("服务器不可用，跳过多轮对话验证", name="skip_multi_turn_validation")
                pytest.skip("服务器不可用")
            elif result[0] == 500:
                # 服务器内部错误，记录为测试失败
                allure.attach(f"第{i}轮对话服务器内部错误，状态码: {result[0]}", name="multi_turn_server_error")
                error_details = result[1] if isinstance(result[1], dict) else {}
                if error_details.get("error_code"):
                    allure.attach(f"第{i}轮对话错误代码: {error_details.get('error_code')}", name=f"error_code_round_{i}")
                if error_details.get("error_details"):
                    allure.attach(f"第{i}轮对话错误详情: {error_details.get('error_details')}", name=f"error_details_round_{i}")
                pytest.fail(f"第{i}轮对话服务器内部错误，状态码: {result[0]}，错误代码: {error_details.get('error_code', 'Unknown')}")
            else:
                # 验证响应状态码
                assert result[0] == 200, f"第{i}轮对话失败，状态码: {result[0]}"

                # 验证响应数据
                self.validate_chat_response(result[1], available_agent, query_index=i, is_intervention_test=is_intervention_test, expected_validations=expected_validations)

                # 获取conversation_id和中断信息（第一轮对话）
                if i == 1:
                    conversation_id = result[1].get("conversation_id")
                    first_response_assistant_message_id = result[1].get("assistant_message_id")

                    # 提取中断信息
                    message = result[1].get("message", {})
                    ext = message.get("ext", {})
                    ask_info = ext.get("ask")

                    if ask_info and has_intervention_enabled:
                        intervention_info = ask_info
                        allure.attach(f"第一轮对话获取到的conversation_id: {conversation_id}", name="conversation_id_extract")
                        allure.attach(f"第一轮对话获取到的assistant_message_id: {first_response_assistant_message_id}", name="assistant_message_id_extract")
                        allure.attach(f"第一轮对话获取到的中断信息: {json.dumps(intervention_info, ensure_ascii=False)}", name="intervention_info_extract")
                    else:
                        allure.attach(f"第一轮对话获取到的conversation_id: {conversation_id}", name="conversation_id_extract")
                else:
                    # 验证后续对话的conversation_id与第一轮一致
                    current_conversation_id = result[1].get("conversation_id")
                    assert current_conversation_id == conversation_id, f"第{i}轮对话的conversation_id不匹配: {current_conversation_id} != {conversation_id}"
                    allure.attach(f"第{i}轮对话conversation_id验证通过: {current_conversation_id}", name=f"conversation_id_verify_{i}")

                    # 中断测试特殊验证
                    if is_intervention_test and i == 2:
                        # 验证中断对话的assistant_message_id应该与第一轮相同
                        current_assistant_message_id = result[1].get("assistant_message_id")
                        assert current_assistant_message_id == first_response_assistant_message_id, f"中断对话的assistant_message_id不匹配: {current_assistant_message_id} != {first_response_assistant_message_id}"
                        allure.attach(f"中断对话assistant_message_id验证通过: {current_assistant_message_id}", name="intervention_assistant_message_id_verify")

        allure.attach(f"智能体 {agent_name} 多轮对话测试通过 (conversation_id: {conversation_id})", name="multi_turn_success")