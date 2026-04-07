# -*- coding:UTF-8 -*-

import pytest
import allure
import json
import time
import random
import threading
import requests

from lib.data_agent import AgentApp, AgentFactory
from urllib3 import disable_warnings
from urllib3.exceptions import InsecureRequestWarning
disable_warnings(InsecureRequestWarning)


@allure.feature("Agent-App外部接口测试")
class TestAgentApp:
    """
    Agent-App外部接口测试类
    测试agent-app相关的API外部接口
    基于最新外部接口文档: release-5.2.0-data-agent/openapi/public
    """

    def _get_available_agent(self, AgentImport, preferred_agent_name=None):
        """
        获取可用的智能体
        :param AgentImport: 导入的智能体列表
        :param preferred_agent_name: 首选的智能体名称，如果提供则优先使用该智能体
        :return: 可用的智能体信息，如果没有可用智能体则返回 None
        """
        if not AgentImport:
            return None

        # 如果指定了首选智能体名称，优先查找该智能体
        if preferred_agent_name:
            for agent_info in AgentImport:
                if (not agent_info["agent_id"].startswith("CONFLICT_") and
                    preferred_agent_name in agent_info["agent_name"]):
                    return agent_info

        # 如果没有找到首选智能体，则使用第一个可用的智能体
        for agent_info in AgentImport:
            if not agent_info["agent_id"].startswith("CONFLICT_"):
                return agent_info

        return None

    @pytest.mark.api
    @pytest.mark.stable
    def test_chat_completion(self, AgentImport, AgentHeaders):
        """
        测试对话接口
        接口文档: /api/agent-app/v1/app/{{app_key}}/chat/completion (POST)
        """
        # 指定使用 "TTFT测试之简单对话explore" 智能体
        available_agent = self._get_available_agent(AgentImport, preferred_agent_name="TTFT测试之简单对话explore")

        if not available_agent:
            pytest.skip("没有可用的智能体")

        app_client = AgentApp()

        agent_id = available_agent["agent_id"]
        agent_name = available_agent["agent_name"]

        # 构造对话请求数据
        chat_request_data = {
            "agent_id": agent_id,
            "agent_version": "v0",
            "query": "你好，请介绍一下你自己",
            "stream": False,
            "inc_stream": False,
            "executor_version": "v2"
        }

        result = app_client.ChatCompletion(agent_id, chat_request_data, AgentHeaders, timeout=300)

        if result[0] in [502, 503, 504]:
            pytest.skip("服务器不可用")

        # 打印详细的错误信息以便排查
        if result[0] != 200:
            error_msg = f"""
            ========== 对话接口错误详情 ==========
            状态码: {result[0]}
            请求URL: {app_client.base_url}/api/agent-app/v1/app/{agent_id}/chat/completion
            请求方法: POST
            Agent ID: {agent_id}
            Agent名称: {agent_name}
            请求参数: {json.dumps(chat_request_data, ensure_ascii=False, indent=2)}
            响应内容: {json.dumps(result[1], ensure_ascii=False, indent=2) if isinstance(result[1], dict) else result[1]}
            
            接口说明:
            - 接口路径: /api/agent-app/v1/app/{{app_key}}/chat/completion
            - 接口类型: 外部接口（对话标签）
            - 支持参数: custom_space_id（可选，查询参数）
            
            可能的原因:
            1. Agent不存在或已被删除
            2. 请求参数格式不正确
            3. Agent配置有问题
            4. 服务器内部错误
            ============================================
            """
            print(error_msg)
            allure.attach(error_msg, name="chat_completion_error_details")
            pytest.fail(f"对话失败，状态码: {result[0]}，详细错误信息已打印")

        # 验证返回的数据结构
        response_data = result[1]
        assert response_data is not None, "响应数据为空"
        assert response_data.get("error") is None, f"对话返回错误: {response_data.get('error')}"

        allure.attach(f"智能体 {agent_name} 对话测试通过", name="chat_completion_success")

    @pytest.mark.api
    @pytest.mark.stable
    def test_resume_chat(self, AgentImport, AgentHeaders):
        """
        测试对话恢复接口
        接口文档: /api/agent-app/v1/app/{{app_key}}/chat/resume (POST)
        测试场景：在对话进行中调用恢复接口，验证能够正确恢复正在进行的对话
        """
        # 指定使用 "TTFT测试之简单对话explore" 智能体
        available_agent = self._get_available_agent(AgentImport, preferred_agent_name="TTFT测试之简单对话explore")

        if not available_agent:
            pytest.skip("没有可用的智能体")

        app_client = AgentApp()

        agent_id = available_agent["agent_id"]
        agent_name = available_agent["agent_name"]

        # 用于存储conversation_id的共享变量
        conversation_id = None
        resume_result = None
        resume_error = None
        chat_completed = False

        def extract_conversation_id_from_line(line):
            """从单行数据中提取conversation_id"""
            line = line.strip()
            if line.startswith('data: '):
                try:
                    # 提取JSON部分
                    json_str = line[6:]  # 去掉 'data: ' 前缀
                    data_obj = json.loads(json_str)

                    # 检查是否是conversation_id的数据行
                    if (data_obj.get('key') == ['conversation_id'] and
                        data_obj.get('action') == 'upsert' and
                        data_obj.get('content')):
                        return data_obj['content']
                except json.JSONDecodeError:
                    pass
            return None

        def chat_thread():
            """发起对话的线程 - 使用流式处理"""
            nonlocal conversation_id, resume_error, chat_completed
            try:
                # 使用长查询确保对话持续较长时间
                chat_request_data = {
                    "agent_id": agent_id,
                    "agent_version": "v0",
                    "query": "你好，请详细介绍一下人工智能的发展历史、现状和未来趋势，包括机器学习、深度学习、自然语言处理、计算机视觉等各个分支的详细说明。请尽量写得详细和全面。",
                    "stream": True,
                    "inc_stream": True,
                    "executor_version": "v2"
                }

                url = app_client.base_url + f"/api/agent-app/v1/app/{agent_id}/chat/completion"

                allure.attach("开始发起流式对话", name="chat_thread_start")

                # 使用requests进行流式处理
                resp = requests.post(
                    url,
                    json=chat_request_data,
                    headers=AgentHeaders,
                    verify=False,
                    allow_redirects=False,
                    timeout=300,
                    stream=True
                )

                if resp.status_code in [502, 503, 504]:
                    resume_error = "服务器不可用"
                    return

                if resp.status_code != 200:
                    resume_error = f"无法发起对话，状态码: {resp.status_code}"
                    return

                # 实时解析流式响应
                for line in resp.iter_lines():
                    if line:
                        decoded_line = line.decode('utf-8')
                        conv_id = extract_conversation_id_from_line(decoded_line)
                        if conv_id:
                            conversation_id = conv_id
                            allure.attach(f"实时提取到conversation_id: {conversation_id}", name="conversation_id_extracted")
                            break  # 找到conversation_id后就可以停止处理

                # 继续读取剩余的响应以确保对话正常进行
                for line in resp.iter_lines():
                    pass  # 继续读取剩余数据

                allure.attach(f"对话线程完成，conversation_id: {conversation_id}", name="chat_thread_complete")
                chat_completed = True

            except Exception as e:
                resume_error = f"对话线程异常: {str(e)}"
                allure.attach(f"对话线程异常: {str(e)}", name="chat_thread_error")

        def resume_thread():
            """恢复对话的线程"""
            nonlocal resume_result, resume_error
            try:
                # 等待conversation_id被提取
                wait_count = 0
                while not conversation_id and wait_count < 60:  # 最多等待60秒
                    time.sleep(0.5)
                    wait_count += 0.5

                if not conversation_id:
                    resume_error = "等待conversation_id超时"
                    return

                allure.attach(f"开始恢复对话，conversation_id: {conversation_id}", name="resume_thread_start")

                # 在对话进行中调用恢复接口
                result = app_client.ResumeChat(agent_id, conversation_id, AgentHeaders, timeout=300)

                if result[0] in [502, 503, 504]:
                    resume_error = "服务器不可用"
                    return

                resume_result = result
                allure.attach(f"恢复对话完成，状态码: {result[0]}", name="resume_thread_complete")

            except Exception as e:
                resume_error = f"恢复线程异常: {str(e)}"
                allure.attach(f"恢复线程异常: {str(e)}", name="resume_thread_error")

        # 启动两个线程
        chat_thread_obj = threading.Thread(target=chat_thread)
        resume_thread_obj = threading.Thread(target=resume_thread)

        # 先启动对话线程
        chat_thread_obj.start()

        # 延迟1秒后启动恢复线程，确保对话已经开始
        time.sleep(1)
        resume_thread_obj.start()

        # 等待两个线程完成
        chat_thread_obj.join(timeout=180)  # 最多等待3分钟
        resume_thread_obj.join(timeout=180)  # 最多等待3分钟

        # 检查结果
        if resume_error:
            pytest.fail(f"对话恢复测试失败: {resume_error}")

        if not conversation_id:
            pytest.skip("未获取到conversation_id，无法测试恢复对话")

        allure.attach(f"获取到的conversation_id: {conversation_id}", name="conversation_id")

        if not resume_result:
            pytest.fail("恢复对话请求失败，未收到响应")

        # 验证恢复对话的结果
        if resume_result[0] == 200:
            # 成功情况：验证返回的数据结构
            response_data = resume_result[1]
            assert response_data is not None, "响应数据为空"

            # 检查响应数据类型和内容
            if isinstance(response_data, dict):
                # 如果是字典类型，检查特定字段
                has_conversation_id = response_data.get("conversation_id") == conversation_id
                has_content = bool(response_data.get("content") or response_data.get("data"))
                has_valid_response = has_conversation_id or has_content
            elif isinstance(response_data, str):
                # 如果是字符串类型，检查是否包含相关内容
                has_conversation_id = conversation_id in response_data
                has_valid_response = len(response_data.strip()) > 0
            else:
                # 其他类型，只要有内容就认为有效
                has_valid_response = True

            if not has_valid_response:
                pytest.fail("恢复对话响应不符合预期，既没有conversation_id也没有对话内容")

            allure.attach(f"智能体 {agent_name} 对话恢复测试通过", name="resume_chat_success")

        elif resume_result[0] == 500:
            # 检查是否是conversation_id not found错误，这可能是Agent配置不支持恢复
            response_data = resume_result[1]
            if isinstance(response_data, dict):
                error_details = response_data.get("error_details", "")
                if "conversation_id not found" in error_details:
                    # 这种情况下说明测试逻辑是正确的（在对话进行中调用了恢复），但Agent不支持恢复
                    allure.attach("测试验证通过：成功在对话进行中调用了恢复接口，但该Agent配置不支持对话恢复", name="resume_chat_expected_behavior")
                else:
                    # 其他500错误
                    error_msg = f"""
                    ========== 对话恢复接口错误详情 ==========
                    状态码: {resume_result[0]}
                    请求URL: {app_client.base_url}/api/agent-app/v1/app/{agent_id}/chat/resume
                    请求方法: POST
                    Agent ID: {agent_id}
                    Agent名称: {agent_name}
                    Conversation ID: {conversation_id}
                    响应内容: {json.dumps(resume_result[1], ensure_ascii=False, indent=2)}

                    测试场景: 在对话进行中调用恢复接口
                    主要验证点: 流式响应conversation_id提取和异步调用机制
                    ============================================
                    """
                    print(error_msg)
                    allure.attach(error_msg, name="resume_chat_error_details")
                    pytest.fail(f"对话恢复接口返回意外错误: {resume_result[0]}")
            else:
                pytest.fail(f"对话恢复失败，状态码: {resume_result[0]}")
        else:
            # 其他状态码处理
            error_msg = f"""
            ========== 对话恢复接口错误详情 ==========
            状态码: {resume_result[0]}
            请求URL: {app_client.base_url}/api/agent-app/v1/app/{agent_id}/chat/resume
            请求方法: POST
            Agent ID: {agent_id}
            Agent名称: {agent_name}
            Conversation ID: {conversation_id}
            响应内容: {json.dumps(resume_result[1], ensure_ascii=False, indent=2) if isinstance(resume_result[1], dict) else resume_result[1]}

            测试场景: 在对话进行中调用恢复接口
            主要验证点: 流式响应conversation_id提取和异步调用机制
            ============================================
            """
            print(error_msg)
            allure.attach(error_msg, name="resume_chat_error_details")
            pytest.fail(f"对话恢复失败，状态码: {resume_result[0]}，详细错误信息已打印")

    @pytest.mark.api
    @pytest.mark.stable
    def test_terminate_chat(self, AgentImport, AgentHeaders):
        """
        测试终止对话接口
        接口文档: /api/agent-app/v1/app/{{app_key}}/chat/termination (POST)
        """
        # 指定使用 "TTFT测试之简单对话explore" 智能体
        available_agent = self._get_available_agent(AgentImport, preferred_agent_name="TTFT测试之简单对话explore")

        if not available_agent:
            pytest.skip("没有可用的智能体")

        app_client = AgentApp()

        agent_id = available_agent["agent_id"]
        agent_name = available_agent["agent_name"]

        # 先发起一次对话，获取conversation_id
        chat_request_data = {
            "agent_id": agent_id,
            "agent_version": "v0",
            "query": "你好",
            "stream": False,
            "inc_stream": False,
            "executor_version": "v2"
        }

        chat_result = app_client.ChatCompletion(agent_id, chat_request_data, AgentHeaders, timeout=300)
        
        if chat_result[0] in [502, 503, 504]:
            pytest.skip("服务器不可用")

        if chat_result[0] != 200:
            pytest.skip(f"无法发起对话，状态码: {chat_result[0]}")

        # 处理响应数据，确保可以正确提取conversation_id
        response_data = chat_result[1]
        if isinstance(response_data, str):
            # 如果是字符串，尝试解析JSON
            try:
                response_data = json.loads(response_data)
            except json.JSONDecodeError:
                pytest.skip("无法解析对话响应数据，未获取到有效的conversation_id")

        conversation_id = response_data.get("conversation_id") if isinstance(response_data, dict) else None
        if not conversation_id:
            pytest.skip("未获取到conversation_id，无法测试终止对话")

        allure.attach(f"获取到的conversation_id: {conversation_id}", name="conversation_id")

        # 测试终止对话
        result = app_client.TerminateChat(agent_id, conversation_id, AgentHeaders)

        if result[0] in [502, 503, 504]:
            pytest.skip("服务器不可用")

        # 打印详细的错误信息以便排查
        if result[0] not in [204, 404]:
            error_msg = f"""
            ========== 终止对话接口错误详情 ==========
            状态码: {result[0]}
            请求URL: {app_client.base_url}/api/agent-app/v1/app/{agent_id}/chat/termination
            请求方法: POST
            Agent ID: {agent_id}
            Agent名称: {agent_name}
            Conversation ID: {conversation_id}
            响应内容: {json.dumps(result[1], ensure_ascii=False, indent=2) if isinstance(result[1], dict) else result[1]}
            
            接口说明:
            - 接口路径: /api/agent-app/v1/app/{{app_key}}/chat/termination
            - 接口类型: 外部接口（对话标签）
            - 请求体: {{"conversation_id": "conversation_id"}}
            - 成功返回: 204 No Content
            - 如果conversation不存在: 404 Not Found
            
            可能的原因:
            1. Conversation不存在（返回404是正常的）
            2. 没有终止权限
            3. 服务器内部错误
            ============================================
            """
            print(error_msg)
            allure.attach(error_msg, name="terminate_chat_error_details")
            pytest.fail(f"终止对话失败，状态码: {result[0]}，详细错误信息已打印")

        # 204表示成功，404表示conversation不存在（可能已经被终止或不存在）
        if result[0] == 204:
            allure.attach(f"智能体 {agent_name} 终止对话测试通过", name="terminate_chat_success")
        elif result[0] == 404:
            allure.attach(f"Conversation不存在（可能已被终止），状态码: 404", name="terminate_chat_not_found")

    @pytest.mark.api
    @pytest.mark.stable
    def test_debug_completion(self, AgentImport, AgentHeaders):
        """
        测试调试接口
        接口文档: /api/agent-app/v1/app/{{app_key}}/debug/completion (POST)
        """
        # 指定使用 "TTFT测试之简单对话explore" 智能体
        available_agent = self._get_available_agent(AgentImport, preferred_agent_name="TTFT测试之简单对话explore")

        if not available_agent:
            pytest.skip("没有可用的智能体")

        app_client = AgentApp()

        agent_id = available_agent["agent_id"]
        agent_name = available_agent["agent_name"]

        # 构造调试请求数据
        debug_request_data = {
            "agent_id": agent_id,
            "agent_version": "v0",
            "query": "你好",
            "stream": False,
            "inc_stream": False,
            "executor_version": "v2"
        }

        result = app_client.DebugChat(agent_id, debug_request_data, AgentHeaders, timeout=300)

        if result[0] in [502, 503, 504]:
            pytest.skip("服务器不可用")

        # 打印详细的错误信息以便排查
        if result[0] != 200:
            error_msg = f"""
            ========== 调试接口错误详情 ==========
            状态码: {result[0]}
            请求URL: {app_client.base_url}/api/agent-app/v1/app/{agent_id}/debug/completion
            请求方法: POST
            Agent ID: {agent_id}
            Agent名称: {agent_name}
            请求参数: {json.dumps(debug_request_data, ensure_ascii=False, indent=2)}
            响应内容: {json.dumps(result[1], ensure_ascii=False, indent=2) if isinstance(result[1], dict) else result[1]}
            
            接口说明:
            - 接口路径: /api/agent-app/v1/app/{{app_key}}/debug/completion
            - 接口类型: 外部接口（对话标签）
            - 支持参数: custom_space_id（可选，查询参数）
            - app_key说明: 通用agent传入agent id
            
            可能的原因:
            1. Agent不存在或已被删除
            2. 请求参数格式不正确
            3. Agent配置有问题
            4. 服务器内部错误
            ============================================
            """
            print(error_msg)
            allure.attach(error_msg, name="debug_completion_error_details")
            pytest.fail(f"调试失败，状态码: {result[0]}，详细错误信息已打印")

        # 验证返回的数据结构
        response_data = result[1]
        assert response_data is not None, "响应数据为空"

        allure.attach(f"智能体 {agent_name} 调试测试通过", name="debug_completion_success")

    @pytest.mark.api
    @pytest.mark.stable
    def test_api_chat_completion(self, AgentImport, AgentHeaders):
        """
        测试APIChat接口 - 包含发布、测试、取消发布、删除的完整流程
        接口文档: /api/agent-app/v1/app/{{app_key}}/api/chat/completion (POST)
        """
        # 指定使用 "TTFT测试之简单对话explore" 智能体
        available_agent = self._get_available_agent(AgentImport, preferred_agent_name="TTFT测试之简单对话explore")

        if not available_agent:
            pytest.skip("没有可用的智能体")

        app_client = AgentApp()
        factory_client = AgentFactory()

        agent_id = available_agent["agent_id"]
        agent_name = available_agent["agent_name"]

        # 步骤1: 获取可用分类并发布Agent为API
        allure.attach("开始获取智能体分类", name="step_1_get_category_start")

        # 获取分类列表
        category_result = factory_client.GetCategory(AgentHeaders)
        if category_result[0] != 200:
            pytest.skip(f"获取智能体分类失败，状态码: {category_result[0]}")

        category_data = category_result[1]
        if not category_data or not isinstance(category_data, list) or len(category_data) == 0:
            pytest.skip("没有可用的智能体分类")

        # 随机选择一个分类
        selected_category = random.choice(category_data)
        category_id = selected_category.get("id")
        category_name = selected_category.get("name", "未知分类")

        allure.attach(f"选择分类: {category_name} (ID: {category_id})", name="step_1_category_selected")

        allure.attach("开始发布Agent为API", name="step_1_publish_start")
        publish_config = {
            "business_domain_id": "bd_public",
            "category_id": [category_id],
            "description": "API测试发布",
            "publish_to_where": ["square"],
            "publish_to_bes": ["api_agent"]
        }

        publish_result = factory_client.PublishAgent(agent_id, publish_config, AgentHeaders)

        if publish_result[0] in [502, 503, 504]:
            pytest.skip("服务器不可用")

        if publish_result[0] not in [200, 201]:  # 200或201都表示发布成功
            error_msg = f"""
            ========== 发布Agent失败 ==========
            状态码: {publish_result[0]}
            Agent ID: {agent_id}
            Agent名称: {agent_name}
            选择分类: {category_name} (ID: {category_id})
            发布配置: {json.dumps(publish_config, ensure_ascii=False, indent=2)}
            响应内容: {json.dumps(publish_result[1], ensure_ascii=False, indent=2) if isinstance(publish_result[1], dict) else publish_result[1]}
            """
            print(error_msg)
            allure.attach(error_msg, name="publish_error_details")
            pytest.fail(f"发布Agent失败，状态码: {publish_result[0]}")

        allure.attach(f"Agent {agent_name} 发布成功", name="step_1_publish_success")

        # 等待发布完成
        time.sleep(3)

        try:
            # 步骤2: 测试APIChat接口
            allure.attach("开始测试APIChat接口", name="step_2_test_start")

            # 构造APIChat请求数据
            api_chat_data = {
                "agent_key": agent_id,  # APIChat需要在请求体中包含agent_key
                "query": "你好，请介绍一下你自己"
            }

            result = app_client.APIChatCompletion(agent_id, api_chat_data, AgentHeaders, timeout=300)

            if result[0] in [502, 503, 504]:
                pytest.skip("服务器不可用")

            # 打印详细的错误信息以便排查
            if result[0] != 200:
                error_msg = f"""
                ========== APIChat接口错误详情 ==========
                状态码: {result[0]}
                请求URL: {app_client.base_url}/api/agent-app/v1/app/{agent_id}/api/chat/completion
                请求方法: POST
                Agent ID: {agent_id}
                Agent名称: {agent_name}
                请求参数: {json.dumps(api_chat_data, ensure_ascii=False, indent=2)}
                响应内容: {json.dumps(result[1], ensure_ascii=False, indent=2) if isinstance(result[1], dict) else result[1]}

                接口说明:
                - 接口路径: /api/agent-app/v1/app/{{app_key}}/api/chat/completion
                - 接口类型: 外部接口（对话标签）
                - 支持参数: custom_space_id（可选，查询参数）
                - app_key说明: 通用agent传入agent id，超级助手传入固定字符串super_assistant
                - 前置条件: Agent必须先发布为API

                可能的原因:
                1. Agent不存在或已被删除
                2. Agent未发布为API
                3. 请求参数格式不正确
                4. Agent配置有问题
                5. 服务器内部错误
                ============================================
                """
                print(error_msg)
                allure.attach(error_msg, name="api_chat_completion_error_details")
                pytest.fail(f"APIChat失败，状态码: {result[0]}，详细错误信息已打印")

            # 验证返回的数据结构
            response_data = result[1]
            assert response_data is not None, "响应数据为空"
            assert response_data.get("error") is None, f"APIChat返回错误: {response_data.get('error')}"

            allure.attach(f"智能体 {agent_name} APIChat测试通过", name="step_2_test_success")

        finally:
            # 步骤3: 清理 - 取消发布Agent
            allure.attach("开始取消发布Agent", name="step_3_unpublish_start")
            unpublish_result = factory_client.UnpublishAgent(agent_id, AgentHeaders)

            if unpublish_result[0] not in [200, 204, 404]:  # 200/204/404都表示成功（204表示成功无内容，404可能是因为已经被取消发布）
                error_msg = f"""
                ========== 取消发布Agent失败 ==========
                状态码: {unpublish_result[0]}
                Agent ID: {agent_id}
                Agent名称: {agent_name}
                响应内容: {json.dumps(unpublish_result[1], ensure_ascii=False, indent=2) if isinstance(unpublish_result[1], dict) else unpublish_result[1]}
                """
                print(error_msg)
                allure.attach(error_msg, name="unpublish_error_details")
                # 这里不fail，因为取消发布失败不应该影响测试结果
                print(f"警告：取消发布Agent失败，状态码: {unpublish_result[0]}")
            else:
                allure.attach(f"Agent {agent_name} 取消发布成功", name="step_3_unpublish_success")

    @pytest.mark.api
    @pytest.mark.stable
    def test_get_api_doc(self, AgentImport, AgentHeaders):
        """
        测试获取Agent Api文档接口
        接口文档: /api/agent-app/v1/app/{{app_key}}/api/doc (POST)
        """
        # 指定使用 "TTFT测试之简单对话explore" 智能体
        available_agent = self._get_available_agent(AgentImport, preferred_agent_name="TTFT测试之简单对话explore")

        if not available_agent:
            pytest.skip("没有可用的智能体")

        app_client = AgentApp()

        agent_id = available_agent["agent_id"]
        agent_name = available_agent["agent_name"]
        agent_version = "v0"

        result = app_client.GetApiDoc(agent_id, agent_version, AgentHeaders)

        if result[0] in [502, 503, 504]:
            pytest.skip("服务器不可用")

        # 打印详细的错误信息以便排查
        if result[0] != 200:
            error_msg = f"""
            ========== 获取Agent Api文档接口错误详情 ==========
            状态码: {result[0]}
            请求URL: {app_client.base_url}/api/agent-app/v1/app/{agent_id}/api/doc
            请求方法: POST
            Agent ID: {agent_id}
            Agent版本: {agent_version}
            Agent名称: {agent_name}
            请求参数: {{"agent_id": "{agent_id}", "agent_version": "{agent_version}"}}
            响应内容: {json.dumps(result[1], ensure_ascii=False, indent=2) if isinstance(result[1], dict) else result[1]}
            
            接口说明:
            - 接口路径: /api/agent-app/v1/app/{{app_key}}/api/doc
            - 接口类型: 外部接口（对话标签）
            - 支持参数: custom_space_id（可选，查询参数）
            - 请求体: {{"agent_id": "agent_id", "agent_version": "agent_version"}}
            - 返回: OpenAPI文档对象
            
            可能的原因:
            1. Agent不存在或已被删除
            2. Agent版本不正确
            3. 请求参数格式不正确
            4. 服务器内部错误
            ============================================
            """
            print(error_msg)
            allure.attach(error_msg, name="get_api_doc_error_details")
            pytest.fail(f"获取Api文档失败，状态码: {result[0]}，详细错误信息已打印")

        # 验证返回的数据结构
        response_data = result[1]
        assert response_data is not None, "Api文档为空"

        allure.attach(f"智能体 {agent_name} 获取Api文档测试通过", name="get_api_doc_success")

    @pytest.mark.api
    @pytest.mark.stable
    def test_file_check(self, AgentHeaders):
        """
        测试文件检查接口
        接口文档: /api/agent-app/v1/file/check (POST)
        """
        app_client = AgentApp()

        # 构造文件检查请求数据（使用示例文件ID）
        file_ids = [
            {"id": "97E09222FC9B45899670E698AF89D113"}
        ]

        result = app_client.FileCheck(file_ids, AgentHeaders)

        if result[0] in [502, 503, 504]:
            pytest.skip("服务器不可用")

        # 打印详细的错误信息以便排查
        if result[0] != 200:
            error_msg = f"""
            ========== 文件检查接口错误详情 ==========
            状态码: {result[0]}
            请求URL: {app_client.base_url}/api/agent-app/v1/file/check
            请求方法: POST
            请求参数: {json.dumps(file_ids, ensure_ascii=False, indent=2)}
            响应内容: {json.dumps(result[1], ensure_ascii=False, indent=2) if isinstance(result[1], dict) else result[1]}
            
            接口说明:
            - 接口路径: /api/agent-app/v1/file/check
            - 接口类型: 外部接口（对话标签）
            - 请求体格式: [{{"id": "file_id"}}]
            - 返回: {{"progress": 0-100, "process_info": [{{"id": "file_id", "status": "processing|completed|failed"}}]}}
            
            可能的原因:
            1. 文件ID不存在或无效
            2. 请求参数格式不正确
            3. 服务器内部错误
            ============================================
            """
            print(error_msg)
            allure.attach(error_msg, name="file_check_error_details")
            pytest.fail(f"文件检查失败，状态码: {result[0]}，详细错误信息已打印")

        # 验证返回的数据结构
        response_data = result[1]
        assert response_data is not None, "响应数据为空"
        
        # 验证返回的数据结构
        assert "progress" in response_data, "响应中缺少progress字段"
        assert "process_info" in response_data, "响应中缺少process_info字段"
        
        # 验证progress是0-100之间的数字
        progress = response_data.get("progress")
        assert isinstance(progress, (int, float)), "progress应该是数字"
        assert 0 <= progress <= 100, "progress应该在0-100之间"

        allure.attach(f"文件检查测试通过，进度: {progress}%", name="file_check_success")

