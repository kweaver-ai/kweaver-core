# -*- coding:UTF-8 -*-

import time
import json
from common.get_content import GetContent
from common.request import Request

class AgentApp():
    def __init__(self):
        file = GetContent("./config/env.ini")
        self.config = file.config()
        self.base_url = self.config["requests"]["protocol"] + "://" + self.config["server"]["host"] + ":" + self.config["server"]["port"]

    def ChatCompletion(self, agent_id, query_data, headers=None, timeout=1200):
        """
        智能体对话接口
        :param agent_id: 智能体ID (作为app_key传入)
        :param query_data: 对话请求数据
        :param headers: 请求头
        :param timeout: 超时时间，默认20分钟(1200秒)
        :return: (status_code, response_data)
        """
        url = self.base_url + f"/api/agent-app/v1/app/{agent_id}/chat/completion"

        # 设置默认参数
        default_data = {
            "stream": False,
            "inc_stream": False,
            "executor_version": "v2"
        }

        # 合并用户数据
        chat_data = {**default_data, **query_data}

        print(f"\n{'='*80}")
        print(f"🤖 AGENT APP CHAT COMPLETION REQUEST")
        print(f"{'='*80}")
        print(f"📤 URL: {url}")
        print(f"⏱️  Timeout: {timeout}s")
        print(f"📋 Agent ID: {agent_id}")
        print(f"🔤 Query: {query_data.get('input', {}).get('query', 'N/A')}")
        print(f"📦 Request Data: {json.dumps(chat_data, indent=2, ensure_ascii=False)}")
        if headers:
            print(f"🔐 Headers: {json.dumps(headers, indent=2, ensure_ascii=False)}")
        print(f"{'='*80}")

        result = Request.post_with_timeout(self, url, chat_data, headers, timeout)

        #print(f"\n📡 AGENT APP CHAT COMPLETION RESPONSE")
        #print(f"{'='*80}")
        #print(f"📊 Status Code: {result[0]}")
        #if isinstance(result[1], dict):
        #    print(f"📄 Response Data: {json.dumps(result[1], indent=2, ensure_ascii=False)}")
        #else:
        #    print(f"📄 Response Text: {result[1]}")
        #print(f"{'='*80}\n")

        return result

    def DebugChat(self, agent_id, debug_data, headers=None, timeout=1200):
        """
        智能体调试接口
        :param agent_id: 智能体ID
        :param debug_data: 调试数据
        :param headers: 请求头
        :param timeout: 超时时间，默认20分钟
        :return: (status_code, response_data)
        """
        url = self.base_url + f"/api/agent-app/v1/app/{agent_id}/debug/completion"
        # 设置默认参数
        default_data = {
            "stream": False,
            "inc_stream": False,
            "executor_version": "v2"
        }

        # 合并用户数据
        chat_data = {**default_data, **debug_data}

        print(f"\n{'='*80}")
        print(f"🤖 AGENT APP CHAT COMPLETION REQUEST")
        print(f"{'='*80}")
        print(f"📤 URL: {url}")
        print(f"⏱️  Timeout: {timeout}s")
        print(f"📋 Agent ID: {agent_id}")
        print(f"🔤 Query: {debug_data.get('input', {}).get('query', 'N/A')}")
        print(f"📦 Request Data: {json.dumps(chat_data, indent=2, ensure_ascii=False)}")
        if headers:
            print(f"🔐 Headers: {json.dumps(headers, indent=2, ensure_ascii=False)}")
        print(f"{'='*80}")
        result = Request.post_with_timeout(self, url, debug_data, headers, timeout)
        return result

    def GetAgentAppInfo(self, agent_id, headers=None):
        """
        获取智能体应用详情
        :param agent_id: 智能体ID
        :param headers: 请求头
        :return: (status_code, response_data)
        """
        url = self.base_url + f"/api/agent-app/v1/app/{agent_id}"

        print(f"\n{'='*80}")
        print(f"📋 AGENT APP INFO REQUEST")
        print(f"{'='*80}")
        print(f"📤 URL: {url}")
        print(f"🆔 Agent ID: {agent_id}")
        if headers:
            print(f"🔐 Headers: {json.dumps(headers, indent=2, ensure_ascii=False)}")
        print(f"{'='*80}")

        result = Request.get(self, url, headers)

        #print(f"\n📋 AGENT APP INFO RESPONSE")
        #print(f"{'='*80}")
        #print(f"📊 Status Code: {result[0]}")
        #if isinstance(result[1], dict):
        #    print(f"📄 Response Data: {json.dumps(result[1], indent=2, ensure_ascii=False)}")
        #else:
        #    print(f"📄 Response Text: {result[1]}")
        #print(f"{'='*80}\n")

        return result

    def TerminateChat(self, agent_id, conversation_id, headers=None):
        """
        终止对话
        :param agent_id: 智能体ID
        :param conversation_id: 对话ID
        :param headers: 请求头
        :return: (status_code, response_data)
        """
        url = self.base_url + f"/api/agent-app/v1/app/{agent_id}/chat/termination"
        data = {"conversation_id": conversation_id}

        print(f"\n{'='*80}")
        print(f"🛑 AGENT APP CHAT TERMINATION REQUEST")
        print(f"{'='*80}")
        print(f"📤 URL: {url}")
        print(f"🆔 Agent ID: {agent_id}")
        print(f"💬 Conversation ID: {conversation_id}")
        print(f"📦 Request Data: {json.dumps(data, indent=2, ensure_ascii=False)}")
        if headers:
            print(f"🔐 Headers: {json.dumps(headers, indent=2, ensure_ascii=False)}")
        print(f"{'='*80}")

        result = Request.post(self, url, data, headers)

        print(f"\n🛑 AGENT APP CHAT TERMINATION RESPONSE")
        print(f"{'='*80}")
        print(f"📊 Status Code: {result[0]}")
        if isinstance(result[1], dict):
            print(f"📄 Response Data: {json.dumps(result[1], indent=2, ensure_ascii=False)}")
        else:
            print(f"📄 Response Text: {result[1]}")
        print(f"{'='*80}\n")

        return result

    def ResumeChat(self, agent_id, conversation_id, headers=None, timeout=1200):
        """
        对话恢复接口
        :param agent_id: 智能体ID (作为app_key传入)
        :param conversation_id: 对话ID
        :param headers: 请求头
        :param timeout: 超时时间，默认20分钟(1200秒)
        :return: (status_code, response_data)
        """
        url = self.base_url + f"/api/agent-app/v1/app/{agent_id}/chat/resume"
        data = {"conversation_id": conversation_id}

        print(f"\n{'='*80}")
        print(f"🔄 AGENT APP CHAT RESUME REQUEST")
        print(f"{'='*80}")
        print(f"📤 URL: {url}")
        print(f"⏱️  Timeout: {timeout}s")
        print(f"📋 Agent ID: {agent_id}")
        print(f"💬 Conversation ID: {conversation_id}")
        print(f"📦 Request Data: {json.dumps(data, indent=2, ensure_ascii=False)}")
        if headers:
            print(f"🔐 Headers: {json.dumps(headers, indent=2, ensure_ascii=False)}")
        print(f"{'='*80}")

        result = Request.post_with_timeout(self, url, data, headers, timeout)

        print(f"\n🔄 AGENT APP CHAT RESUME RESPONSE")
        print(f"{'='*80}")
        print(f"📊 Status Code: {result[0]}")
        if isinstance(result[1], dict):
            print(f"📄 Response Data: {json.dumps(result[1], indent=2, ensure_ascii=False)}")
        else:
            print(f"📄 Response Text: {result[1]}")
        print(f"{'='*80}\n")

        return result

    def APIChatCompletion(self, agent_id, query_data, headers=None, timeout=1200, custom_space_id=None):
        """
        APIChat接口
        :param agent_id: 智能体ID (作为app_key传入)
        :param query_data: 对话请求数据
        :param headers: 请求头
        :param timeout: 超时时间，默认20分钟(1200秒)
        :param custom_space_id: 自定义空间ID（可选）
        :return: (status_code, response_data)
        """
        url = self.base_url + f"/api/agent-app/v1/app/{agent_id}/api/chat/completion"
        if custom_space_id:
            url += f"?custom_space_id={custom_space_id}"

        print(f"\n{'='*80}")
        print(f"🤖 AGENT APP API CHAT COMPLETION REQUEST")
        print(f"{'='*80}")
        print(f"📤 URL: {url}")
        print(f"⏱️  Timeout: {timeout}s")
        print(f"📋 Agent ID: {agent_id}")
        if custom_space_id:
            print(f"📋 Custom Space ID: {custom_space_id}")
        print(f"📦 Request Data: {json.dumps(query_data, indent=2, ensure_ascii=False)}")
        if headers:
            print(f"🔐 Headers: {json.dumps(headers, indent=2, ensure_ascii=False)}")
        print(f"{'='*80}")

        result = Request.post_with_timeout(self, url, query_data, headers, timeout)

        print(f"\n🤖 AGENT APP API CHAT COMPLETION RESPONSE")
        print(f"{'='*80}")
        print(f"📊 Status Code: {result[0]}")
        if isinstance(result[1], dict):
            print(f"📄 Response Data: {json.dumps(result[1], indent=2, ensure_ascii=False)}")
        else:
            print(f"📄 Response Text: {result[1]}")
        print(f"{'='*80}\n")

        return result

    def GetApiDoc(self, agent_id, agent_version, headers=None, custom_space_id=None):
        """
        获取Agent Api文档
        :param agent_id: 智能体ID
        :param agent_version: 智能体版本
        :param headers: 请求头
        :param custom_space_id: 自定义空间ID（可选）
        :return: (status_code, response_data)
        """
        url = self.base_url + f"/api/agent-app/v1/app/{agent_id}/api/doc"
        data = {
            "agent_id": agent_id,
            "agent_version": agent_version
        }
        params = {}
        if custom_space_id:
            params["custom_space_id"] = custom_space_id

        print(f"\n{'='*80}")
        print(f"📚 AGENT APP API DOC REQUEST")
        print(f"{'='*80}")
        print(f"📤 URL: {url}")
        print(f"📋 Agent ID: {agent_id}")
        print(f"📋 Agent Version: {agent_version}")
        if custom_space_id:
            print(f"📋 Custom Space ID: {custom_space_id}")
        print(f"📦 Request Data: {json.dumps(data, indent=2, ensure_ascii=False)}")
        if headers:
            print(f"🔐 Headers: {json.dumps(headers, indent=2, ensure_ascii=False)}")
        print(f"{'='*80}")

        # 如果有查询参数，需要修改URL
        if params:
            url += "?" + "&".join([f"{k}={v}" for k, v in params.items()])

        result = Request.post(self, url, data, headers)

        print(f"\n📚 AGENT APP API DOC RESPONSE")
        print(f"{'='*80}")
        print(f"📊 Status Code: {result[0]}")
        if isinstance(result[1], dict):
            print(f"📄 Response Data: {json.dumps(result[1], indent=2, ensure_ascii=False)}")
        else:
            print(f"📄 Response Text: {result[1]}")
        print(f"{'='*80}\n")

        return result

    def FileCheck(self, file_ids, headers=None):
        """
        文件检查接口
        :param file_ids: 文件ID列表，格式: [{"id": "file_id"}]
        :param headers: 请求头
        :return: (status_code, response_data)
        """
        url = self.base_url + "/api/agent-app/v1/file/check"

        print(f"\n{'='*80}")
        print(f"📁 AGENT APP FILE CHECK REQUEST")
        print(f"{'='*80}")
        print(f"📤 URL: {url}")
        print(f"📦 Request Data: {json.dumps(file_ids, indent=2, ensure_ascii=False)}")
        if headers:
            print(f"🔐 Headers: {json.dumps(headers, indent=2, ensure_ascii=False)}")
        print(f"{'='*80}")

        result = Request.post(self, url, file_ids, headers)

        print(f"\n📁 AGENT APP FILE CHECK RESPONSE")
        print(f"{'='*80}")
        print(f"📊 Status Code: {result[0]}")
        if isinstance(result[1], dict):
            print(f"📄 Response Data: {json.dumps(result[1], indent=2, ensure_ascii=False)}")
        else:
            print(f"📄 Response Text: {result[1]}")
        print(f"{'='*80}\n")

        return result