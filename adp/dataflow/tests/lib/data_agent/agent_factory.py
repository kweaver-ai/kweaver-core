# -*- coding:UTF-8 -*-

import os
import json
import requests
from common.get_content import GetContent
from common.request import Request

class AgentFactory():
    def __init__(self):
        file = GetContent("./config/env.ini")
        self.config = file.config()
        self.base_url = self.config["requests"]["protocol"] + "://" + self.config["server"]["host"] + ":" + self.config["server"]["port"]+"/api"

    def ImportAgents(self, file_path, import_type="upsert", headers=None):
        """
        导入智能体
        :param file_path: JSON文件路径
        :param import_type: 导入类型 (upsert/create)
        :param headers: 请求头
        :return: (status_code, response_data)
        """
        url = self.base_url + "/agent-factory/v3/agent-inout/import"

        if not os.path.exists(file_path):
            print(f"\n❌ AGENT FACTORY IMPORT FAILED - File not found: {file_path}")
            return 400, {"error": "File not found"}

        with open(file_path, 'r', encoding='utf-8') as f:
            files = {'file': ('agent_import.json', f.read(), 'application/json')}
            data = {'import_type': import_type}

        print(f"\n{'='*80}")
        print(f"📤 AGENT FACTORY IMPORT REQUEST")
        print(f"{'='*80}")
        print(f"📤 URL: {url}")
        print(f"📁 File Path: {file_path}")
        print(f"🔧 Import Type: {import_type}")
        #print(f"📦 Data: {json.dumps(data, indent=2, ensure_ascii=False)}")
        if headers:
            print(f"🔐 Headers: {json.dumps(headers, indent=2, ensure_ascii=False)}")
        print(f"{'='*80}")

        result = Request.upload_file(self, url, files, data, headers)

        #print(f"\n📤 AGENT FACTORY IMPORT RESPONSE")
        #print(f"{'='*80}")
        #print(f"📊 Status Code: {result[0]}")
        #if isinstance(result[1], dict):
        #    print(f"📄 Response Data: {json.dumps(result[1], indent=2, ensure_ascii=False)}")
        #else:
        #    print(f"📄 Response Text: {result[1]}")
        #print(f"{'='*80}\n")

        return result

    def DeleteAgent(self, agent_id, headers=None):
        """
        删除智能体
        :param agent_id: 智能体ID
        :param headers: 请求头
        :return: (status_code, response_data)
        """
        url = self.base_url + f"/agent-factory/v3/agent/{agent_id}"

        #print(f"\n{'='*80}")
        #print(f"🗑️  AGENT FACTORY DELETE REQUEST")
        #print(f"{'='*80}")
        #print(f"📤 URL: {url}")
        #print(f"🆔 Agent ID: {agent_id}")
        #if headers:
        #    print(f"🔐 Headers: {json.dumps(headers, indent=2, ensure_ascii=False)}")
        #print(f"{'='*80}")

        result = Request.delete(self, url, None, headers)

        #print(f"\n🗑️  AGENT FACTORY DELETE RESPONSE")
        #print(f"{'='*80}")
        #print(f"📊 Status Code: {result[0]}")
        #if isinstance(result[1], dict):
        #    print(f"📄 Response Data: {json.dumps(result[1], indent=2, ensure_ascii=False)}")
        #else:
        #    print(f"📄 Response Text: {result[1]}")
        #print(f"{'='*80}\n")

        return result

    def GetAgentByKey(self, agent_key, headers=None):
        """
        根据key获取智能体详情
        :param agent_key: 智能体key
        :param headers: 请求头
        :return: (status_code, response_data)
        """
        url = self.base_url + f"/agent-factory/v3/agent/by-key/{agent_key}"

        print(f"\n{'='*80}")
        print(f"🔍 AGENT FACTORY GET BY KEY REQUEST")
        print(f"{'='*80}")
        print(f"📤 URL: {url}")
        print(f"🔑 Agent Key: {agent_key}")
        if headers:
            print(f"🔐 Headers: {json.dumps(headers, indent=2, ensure_ascii=False)}")
        print(f"{'='*80}")

        result = Request.get(self, url, headers)

        #print(f"\n🔍 AGENT FACTORY GET BY KEY RESPONSE")
        #print(f"{'='*80}")
        #print(f"📊 Status Code: {result[0]}")
        #if isinstance(result[1], dict):
        #    print(f"📄 Response Data: {json.dumps(result[1], indent=2, ensure_ascii=False)}")
        #else:
        #    print(f"📄 Response Text: {result[1]}")
        #print(f"{'='*80}\n")

        return result

    def GetAgent(self, agent_id, headers=None):
        """
        获取智能体详情
        :param agent_id: 智能体ID
        :param headers: 请求头
        :return: (status_code, response_data)
        """
        url = self.base_url + f"/agent-factory/v3/agent/{agent_id}"
        return Request.get(self, url, headers)

    def GetModelList(self, params=None, headers=None):
        """
        获取大模型列表信息接口
        :param params: 查询参数，可以使用name进行模糊搜索
        :param headers: 请求头
        :return: (status_code, response_data)
        """
        url = self.base_url + "/mf-model-manager/v1/llm/list"
        if params is None:
            params = {}

        # 添加必需的分页参数
        if "page" not in params:
            params["page"] = 1
        if "size" not in params:
            params["size"] = 50

        print(f"\n{'='*80}")
        print(f"📋 MODEL MANAGER LIST REQUEST")
        print(f"{'='*80}")
        print(f"📤 URL: {url}")
        print(f"🔍 Params: {json.dumps(params, indent=2, ensure_ascii=False)}")
        if headers:
            print(f"🔐 Headers: {json.dumps(headers, indent=2, ensure_ascii=False)}")
        print(f"{'='*80}")

        result = Request.query(self, url, params, headers)

        #print(f"\n📋 MODEL MANAGER LIST RESPONSE")
        #print(f"{'='*80}")
        #print(f"📊 Status Code: {result[0]}")
        #if isinstance(result[1], dict):
        #    print(f"📄 Response Data: {json.dumps(result[1], indent=2, ensure_ascii=False)}")
        #else:
        #    print(f"📄 Response Text: {result[1]}")
        #print(f"{'='*80}\n")

        return result

    def TestModelConnectivity(self, model_id, headers=None):
        """
        大模型测试链接接口
        :param model_id: 模型ID
        :param headers: 请求头
        :return: (status_code, response_data)
        """
        url = self.base_url + f"/mf-model-manager/v1/llm/test"
        body={
            "model_id": model_id,
            "model_series": "tome"  # 添加必需的model_series参数
        }

        print(f"\n{'='*80}")
        print(f"🔧 MODEL CONNECTIVITY TEST REQUEST")
        print(f"{'='*80}")
        print(f"📤 URL: {url}")
        print(f"🤖 Model ID: {model_id}")
        print(f"📦 Request Body: {json.dumps(body, indent=2, ensure_ascii=False)}")
        if headers:
            print(f"🔐 Headers: {json.dumps(headers, indent=2, ensure_ascii=False)}")
        print(f"{'='*80}")

        result = Request.post(self, url, body, headers)

        print(f"\n🔧 MODEL CONNECTIVITY TEST RESPONSE")
        print(f"{'='*80}")
        print(f"📊 Status Code: {result[0]}")
        if isinstance(result[1], dict):
            print(f"📄 Response Data: {json.dumps(result[1], indent=2, ensure_ascii=False)}")
        else:
            print(f"📄 Response Text: {result[1]}")
        print(f"{'='*80}\n")

        return result

    def UpdateAgent(self, agent_id, agent_config, headers=None):
        """
        更新智能体配置信息
        :param agent_id: 智能体ID
        :param agent_config: 智能体配置数据
        :param headers: 请求头
        :return: (status_code, response_data)
        """
        url = self.base_url + f"/agent-factory/v3/agent/{agent_id}"

        print(f"\n{'='*80}")
        print(f"🔄 AGENT UPDATE REQUEST")
        print(f"{'='*80}")
        print(f"📤 URL: {url}")
        print(f"🆔 Agent ID: {agent_id}")
        print(f"📦 Request Body: {json.dumps(agent_config, indent=2, ensure_ascii=False)}")
        if headers:
            print(f"🔐 Headers: {json.dumps(headers, indent=2, ensure_ascii=False)}")
        print(f"{'='*80}")

        result = Request.put(self, url, agent_config, headers)

        #print(f"\n🔄 AGENT UPDATE RESPONSE")
        #print(f"{'='*80}")
        #print(f"📊 Status Code: {result[0]}")
        #if isinstance(result[1], dict):
        #    print(f"📄 Response Data: {json.dumps(result[1], indent=2, ensure_ascii=False)}")
        #else:
        #    print(f"📄 Response Text: {result[1]}")
        #print(f"{'='*80}\n")

        return result

    def ConfigureAgentDataSource(self, agent_id, config, headers=None):
        """
        配置智能体数据源（文档或知识网络）
        :param agent_id: 智能体ID
        :param config: 配置信息，包含doc_resource_config和graph_resource_config
        :param headers: 请求头
        :return: (status_code, response_data)
        """
        print(f"\n🔧 CONFIGURE AGENT DATA SOURCE")
        print(f"{'='*80}")
        print(f"🆔 Agent ID: {agent_id}")
        print(f"📦 Config keys: {list(config.keys()) if config else 'None'}")
        print(f"📄 need_conf_doc_resource: {config.get('need_conf_doc_resource', False)}")
        print(f"📄 need_conf_graph_resource: {config.get('need_conf_graph_resource', False)}")
        print(f"📄 doc_resource_config exists: {'doc_resource_config' in config}")
        print(f"📄 graph_resource_config exists: {'graph_resource_config' in config}")
        # 强制显示doc_resource_config内容
        doc_config = config.get('doc_resource_config')
        print(f"📄 doc_resource_config value: {doc_config}")
        print(f"📄 doc_resource_config type: {type(doc_config)}")
        if doc_config:
            print(f"📄 doc_resource_config content: {json.dumps(doc_config, indent=2, ensure_ascii=False)}")
        else:
            print(f"📄 doc_resource_config: None or empty")
        print(f"{'='*80}")

        # 首先获取当前智能体配置
        get_result = self.GetAgent(agent_id, headers)
        if get_result[0] != 200:
            return get_result

        current_config = get_result[1]

        # 备份原始配置
        updated_config = current_config.copy()

        # 处理config字段是字符串的情况
        if isinstance(updated_config.get("config"), str):
            try:
                updated_config["config"] = json.loads(updated_config["config"])
            except (json.JSONDecodeError, TypeError) as e:
                print(f"Failed to parse agent config string: {e}")
                return (400, {"error": "Invalid agent config format"})

        # 确保data_source结构存在
        if "data_source" not in updated_config["config"]:
            updated_config["config"]["data_source"] = {
                "kg": [],
                "doc": [],
                "metric": [],
                "kn_entry": [],
                "knowledge_network": [],
                "advanced_config": {}
            }

        # 处理文档数据源配置
        need_doc = config.get("need_conf_doc_resource")
        doc_config = config.get("doc_resource_config")
        print(f"📄 条件检查 - need_conf_doc_resource: {need_doc}, type: {type(need_doc)}")
        print(f"📄 条件检查 - doc_resource_config: {doc_config is not None}, type: {type(doc_config)}")
        print(f"📄 条件检查 - doc_resource_config长度: {len(doc_config) if doc_config else 'N/A'}")

        if need_doc and doc_config:
            print(f"📄 开始处理文档数据源配置...")
            doc_config = config["doc_resource_config"][0]  # 取第一个配置
            print(f"📄 提取doc_config: {json.dumps(doc_config, indent=2, ensure_ascii=False)}")

            # 更新data_source.doc配置
            updated_config["config"]["data_source"]["doc"] = [{
                "ds_id": doc_config["ds_id"],
                "fields": doc_config["fields"],
                "datasets": doc_config["datasets"]
            }]
            print(f"📄 更新后的data_source.doc: {json.dumps(updated_config['config']['data_source']['doc'], indent=2, ensure_ascii=False)}")

            # 更新advanced_config.doc配置
            if "advanced_config" not in updated_config["config"]["data_source"]:
                updated_config["config"]["data_source"]["advanced_config"] = {}

            updated_config["config"]["data_source"]["advanced_config"]["doc"] = {
                "retrieval_slices_num": 150,
                "max_slice_per_cite": 16,
                "rerank_topk": 15,
                "slice_head_num": 0,
                "slice_tail_num": 2,
                "documents_num": 8,
                "document_threshold": -5.5,
                "retrieval_max_length": 12288
            }
            print(f"📄 更新后的advanced_config.doc: {json.dumps(updated_config['config']['data_source']['advanced_config']['doc'], indent=2, ensure_ascii=False)}")

            # 更新pre_dolphin配置，添加文档召回模块
            doc_retrieve_module = {
                "key": "doc_retrieve",
                "name": "文档召回模块",
                "value": "\n/judge/(tools=[\"doc_qa\"], history=True)判断【$query】是否需要到文档中召回，如果不需要召回，则直接返回\"不需要文档召回\"，否则执行工具对【$query】进行召回 -> doc_retrieval_res\n",
                "enabled": True,
                "edited": False
            }

            # 检查是否已存在doc_retrieve模块，如果不存在则添加
            has_doc_retrieve = False
            for module in updated_config["config"].get("pre_dolphin", []):
                if module.get("key") == "doc_retrieve":
                    has_doc_retrieve = True
                    break

            if not has_doc_retrieve:
                if "pre_dolphin" not in updated_config["config"]:
                    updated_config["config"]["pre_dolphin"] = []
                updated_config["config"]["pre_dolphin"].insert(0, doc_retrieve_module)
                print(f"📄 添加文档召回模块成功")
            else:
                print(f"📄 文档召回模块已存在，跳过添加")

            print(f"📄 文档数据源配置处理完成")
        else:
            print(f"📄 跳过文档数据源配置 - need_conf_doc_resource: {config.get('need_conf_doc_resource')}, doc_resource_config存在: {bool(config.get('doc_resource_config'))}")

        # 处理知识网络数据源配置
        if config.get("need_conf_graph_resource") and config.get("graph_resource_config"):
            graph_config = config["graph_resource_config"][0]  # 取第一个配置

            # 更新data_source.kg配置
            updated_config["config"]["data_source"]["kg"] = [{
                "kg_id": graph_config["kg_id"],
                "fields": graph_config["fields"],
                "field_properties": graph_config["field_properties"],
                "output_fields": graph_config["output_fields"]
            }]

            # 更新advanced_config.kg配置
            if "advanced_config" not in updated_config["config"]["data_source"]:
                updated_config["config"]["data_source"]["advanced_config"] = {}

            updated_config["config"]["data_source"]["advanced_config"]["kg"] = {
                "text_match_entity_nums": 60,
                "vector_match_entity_nums": 60,
                "graph_rag_topk": 25,
                "long_text_length": 256,
                "reranker_sim_threshold": -5.5,
                "retrieval_max_length": 12288
            }

            # 更新dolphin配置（添加规划提示）
            updated_config["config"]["dolphin"] = "\n/prompt/(output='list_str')请将原始任务拆解成多步骤任务，任务的个数应该在2~4个，前几步应是搜索资料，最后一步应该是总结内容。\n以list格式返回\n示例:\n原始任务:如何提高英语口语能力？\n拆解结果:[\"搜索资料:提高英语口语能力的最佳方法\",\"搜索资料：适合初学者的英语口语练习技巧\",\"总结内容：如何提高英语口语能力？\"]\n\n原始任务:$query\n拆解结果:->plan_list\n# eval($plan_list.answer)->plan_list\n"

            # 更新pre_dolphin配置，添加知识网络召回模块
            graph_retrieve_module = {
                "key": "graph_retrieve",
                "name": "业务知识网络召回模块",
                "value": "\n/judge/(tools=[\"graph_qa\"], history=True)判断【$query】是否需要到业务知识网络中召回，如果不需要召回，则直接返回\"不需要业务知识网络召回\"，否则执行工具对【$query】进行召回 -> graph_retrieval_res\n",
                "enabled": True,
                "edited": False
            }

            # 检查是否已存在graph_retrieve模块，如果不存在则添加
            has_graph_retrieve = False
            for module in updated_config["config"].get("pre_dolphin", []):
                if module.get("key") == "graph_retrieve":
                    has_graph_retrieve = True
                    break

            if not has_graph_retrieve:
                if "pre_dolphin" not in updated_config["config"]:
                    updated_config["config"]["pre_dolphin"] = []
                updated_config["config"]["pre_dolphin"].insert(0, graph_retrieve_module)

        # 如果需要配置文档或知识网络资源，更新context_organize模块
        if (config.get("need_conf_doc_resource") or config.get("need_conf_graph_resource")):
            # 查找并更新context_organize模块
            context_module_found = False
            for module in updated_config["config"].get("pre_dolphin", []):
                if module.get("key") == "context_organize":
                    # 根据配置的数据源动态构建context_organize的value
                    context_value_parts = ['"如果有参考文档，结合参考文档回答用户的问题。如果没有参考文档，根据用户的问题回答。\\n" -> reference']

                    # 只有配置了文档数据源时才添加文档召回检查
                    if config.get("need_conf_doc_resource"):
                        context_value_parts.append('/if/ "result" in $doc_retrieval_res[\'answer\'] and $doc_retrieval_res[\'answer\'][\'result\']:\n    $reference + "文档召回的内容：" + $doc_retrieval_res[\'answer\'][\'result\'] + "\\n" -> reference\n/end/')

                    # 只有配置了知识网络数据源时才添加知识网络召回检查
                    if config.get("need_conf_graph_resource"):
                        context_value_parts.append('/if/ "result" in $graph_retrieval_res[\'answer\'] and $graph_retrieval_res[\'answer\'][\'result\']:\n    $reference + "业务知识网络召回的内容：" + $graph_retrieval_res[\'answer\'][\'result\'] + "\\n" -> reference\n/end/')

                    context_value_parts.append('{"reference": $reference, "query": "用户的问题为："+$query} -> context')

                    # 更新context_organize的value
                    module["value"] = "\n".join(context_value_parts)
                    context_module_found = True
                    break

            # 如果没有找到context_organize模块，添加一个
            if not context_module_found:
                context_organize_module = {
                    "key": "context_organize",
                    "name": "上下文组织模块",
                    "value": "\n{\"query\": \"用户的问题为: \"+$query} -> context\n",
                    "enabled": True,
                    "edited": False
                }
                if "pre_dolphin" not in updated_config["config"]:
                    updated_config["config"]["pre_dolphin"] = []
                updated_config["config"]["pre_dolphin"].append(context_organize_module)

        # 执行更新 - 保持config为对象格式，不序列化为字符串
        final_config = updated_config.copy()
        return self.UpdateAgent(agent_id, final_config, headers)

    def GetToolBox(self, tool_box_name, headers=None):
        """
        获取工具箱信息
        :param tool_box_name: 工具箱名称
        :param headers: 请求头
        :return: (status_code, response_data)
        """
        url = self.base_url + "/agent-operator-integration/v1/tool-box/market"
        params = {
            "name": tool_box_name,
            "page": 1,
            "page_size": 20
        }

        print(f"\n{'='*80}")
        print(f"🔧 GET TOOL BOX REQUEST")
        print(f"{'='*80}")
        print(f"📤 URL: {url}")
        print(f"🔍 Tool Box Name: {tool_box_name}")
        print(f"📦 Params: {json.dumps(params, indent=2, ensure_ascii=False)}")
        if headers:
            print(f"🔐 Headers: {json.dumps(headers, indent=2, ensure_ascii=False)}")
        print(f"{'='*80}")

        result = Request.query(self, url, params, headers)

        #print(f"\n🔧 GET TOOL BOX RESPONSE")
        #print(f"{'='*80}")
        #print(f"📊 Status Code: {result[0]}")
        #if isinstance(result[1], dict):
        #    print(f"📄 Response Data: {json.dumps(result[1], indent=2, ensure_ascii=False)}")
        #else:
        #    print(f"📄 Response Text: {result[1]}")
        #print(f"{'='*80}\n")

        return result

    def GetToolsByBoxId(self, tool_box_id, headers=None):
        """
        根据工具箱ID获取工具列表
        :param tool_box_id: 工具箱ID
        :param headers: 请求头
        :return: (status_code, response_data)
        """
        url = self.base_url + f"/agent-operator-integration/v1/tool-box/{tool_box_id}/tools/list"
        params = {
            "box_id": tool_box_id,
            "page": 1,
            "page_size": 20
        }

        print(f"\n{'='*80}")
        print(f"🔧 GET TOOLS BY BOX ID REQUEST")
        print(f"{'='*80}")
        print(f"📤 URL: {url}")
        print(f"🆔 Tool Box ID: {tool_box_id}")
        print(f"📦 Params: {json.dumps(params, indent=2, ensure_ascii=False)}")
        if headers:
            print(f"🔐 Headers: {json.dumps(headers, indent=2, ensure_ascii=False)}")
        print(f"{'='*80}")

        result = Request.query(self, url, params, headers)

        print(f"\n🔧 GET TOOLS BY BOX ID RESPONSE")
        print(f"{'='*80}")
        print(f"📊 Status Code: {result[0]}")
        if isinstance(result[1], dict):
            print(f"📄 Response Data: {json.dumps(result[1], indent=2, ensure_ascii=False)}")
        else:
            print(f"📄 Response Text: {result[1]}")
        print(f"{'='*80}\n")

        return result

    def GetToolIdByName(self, tool_name, headers=None):
        """
        根据工具名称获取工具ID
        :param tool_name: 工具名称
        :param headers: 请求头
        :return: (tool_box_id, tool_id) or (None, None) if not found
        """
        try:
            # 首先获取"搜索工具"工具箱
            tool_box_result = self.GetToolBox("搜索工具", headers)
            if tool_box_result[0] != 200:
                return None, None

            tool_box_data = tool_box_result[1]
            if not tool_box_data.get("data"):
                return None, None

            # 获取工具箱ID
            tool_box_id = tool_box_data["data"][0]["box_id"]

            # 获取工具列表
            tools_result = self.GetToolsByBoxId(tool_box_id, headers)
            if tools_result[0] != 200:
                return None, None

            tools_data = tools_result[1]
            if not tools_data.get("tools"):
                return None, None

            # 查找指定工具
            for tool in tools_data["tools"]:
                if tool.get("name") == tool_name:
                    return tool_box_id, tool.get("tool_id")

            return None, None

        except Exception as e:
            print(f"Error getting tool ID by name {tool_name}: {e}")
            return None, None

    def ConfigureAgentSkills(self, agent_id, agent_config, shared_configs, headers=None):
        """
        配置智能体技能（工具和代理）
        :param agent_id: 智能体ID
        :param agent_config: 智能体配置，包含skill_list和config_intervention
        :param shared_configs: 共享配置，包含技能配置文件路径
        :param headers: 请求头
        :return: (status_code, response_data)
        """
        # 首先获取当前智能体配置
        get_result = self.GetAgent(agent_id, headers)
        if get_result[0] != 200:
            return get_result

        current_config = get_result[1]

        # 备份原始配置
        updated_config = current_config.copy()

        # 处理config字段是字符串的情况
        if isinstance(updated_config.get("config"), str):
            try:
                updated_config["config"] = json.loads(updated_config["config"])
            except (json.JSONDecodeError, TypeError) as e:
                print(f"Failed to parse agent config string: {e}")
                return (400, {"error": "Invalid agent config format"})

        # 获取技能列表和干预配置
        skill_list = agent_config.get("skill_list", [])
        config_intervention = agent_config.get("config_intervention", [])

        if not skill_list:
            # 如果skill_list为空，清空所有技能配置
            updated_config["config"]["skills"] = {
                "tools": [],
                "agents": [],
                "mcps": None
            }
        else:
            # 初始化技能配置
            if "skills" not in updated_config["config"]:
                updated_config["config"]["skills"] = {
                    "tools": [],
                    "agents": [],
                    "mcps": None
                }

            tools_list = []
            agents_list = []

            # 处理技能配置
            base_dir = "./data/data-agent"  # 技能配置文件的基础目录

            for idx, skill_name in enumerate(skill_list):
                # 获取当前技能对应的干预配置
                intervention_value = False
                if idx < len(config_intervention):
                    intervention_value = config_intervention[idx]

                if skill_name == "zhipu_search_tool":
                    # 动态获取zhipu_search_tool的工具ID
                    tool_box_id, tool_id = self.GetToolIdByName("zhipu_search_tool", headers)

                    if tool_box_id and tool_id:
                        # 加载zhipu_search_tool配置
                        zhipu_config_path = shared_configs.get("zhipu_search_tool_config", "")
                        if zhipu_config_path:
                            full_path = os.path.join(base_dir, os.path.basename(zhipu_config_path))
                            if os.path.exists(full_path):
                                try:
                                    with open(full_path, 'r', encoding='utf-8') as f:
                                        zhipu_tools = json.load(f)

                                    # 更新工具ID为实际获取的值
                                    for tool in zhipu_tools:
                                        tool["tool_box_id"] = tool_box_id
                                        tool["tool_id"] = tool_id
                                        # 设置干预配置
                                        tool["intervention"] = intervention_value
                                        if "details" in tool:
                                            tool["details"]["tool_box_id"] = tool_box_id
                                            tool["details"]["tool_id"] = tool_id
                                            # 同时更新details中的intervention配置
                                            tool["details"]["intervention"] = intervention_value

                                    tools_list.extend(zhipu_tools)
                                    print(f"Successfully updated zhipu_search_tool with real IDs: box_id={tool_box_id}, tool_id={tool_id}, intervention={intervention_value}")
                                except Exception as e:
                                    print(f"Failed to load zhipu_search_tool config from {full_path}: {e}")
                    else:
                        print(f"Failed to get real tool IDs for zhipu_search_tool, skipping tool configuration")

                elif skill_name in ["Plan_Agent", "Summary_Agent"]:
                    # 加载agent配置
                    if skill_name == "Plan_Agent":
                        agent_config_path = shared_configs.get("plan_agent_config", "")
                    elif skill_name == "Summary_Agent":
                        agent_config_path = shared_configs.get("summary_agent_config", "")

                    if agent_config_path:
                        full_path = os.path.join(base_dir, os.path.basename(agent_config_path))
                        if os.path.exists(full_path):
                            try:
                                with open(full_path, 'r', encoding='utf-8') as f:
                                    agent_skills = json.load(f)

                                # 设置agent技能的干预配置
                                for agent_skill in agent_skills:
                                    agent_skill["intervention"] = intervention_value
                                    if "details" in agent_skill:
                                        agent_skill["details"]["intervention"] = intervention_value

                                agents_list.extend(agent_skills)
                                print(f"Successfully loaded {skill_name} config with intervention={intervention_value}")
                            except Exception as e:
                                print(f"Failed to load {skill_name} config from {full_path}: {e}")
  
            # 更新技能配置
            updated_config["config"]["skills"]["tools"] = tools_list
            updated_config["config"]["skills"]["agents"] = agents_list

        # 保持config为对象格式，不序列化为字符串
        final_config = updated_config.copy()

        print(f"🔧 Final agent config before update:")
        print(f"   - Tools count: {len(tools_list)}")
        if tools_list:
            for i, tool in enumerate(tools_list):
                print(f"   - Tool {i}: tool_id={tool.get('tool_id')}, tool_box_id={tool.get('tool_box_id')}")

        # 执行更新
        return self.UpdateAgent(agent_id, final_config, headers)

    def ConfigureAgentFeatures(self, agent_id, config, headers=None):
        """
        配置智能体特性（memory、related_question、plan_mode）
        :param agent_id: 智能体ID
        :param config: 配置信息，包含need_config_memory、need_config_related_question、need_config_plan_mode
        :param headers: 请求头
        :return: (status_code, response_data)
        """
        print(f"\n🔧 CONFIGURE AGENT FEATURES")
        print(f"{'='*80}")
        print(f"🆔 Agent ID: {agent_id}")
        print(f"📄 need_config_memory: {config.get('need_config_memory', False)}")
        print(f"📄 need_config_related_question: {config.get('need_config_related_question', False)}")
        print(f"📄 need_config_plan_mode: {config.get('need_config_plan_mode', False)}")
        print(f"{'='*80}")

        # 首先获取当前智能体配置
        get_result = self.GetAgent(agent_id, headers)
        if get_result[0] != 200:
            return get_result

        current_config = get_result[1]

        # 备份原始配置
        updated_config = current_config.copy()

        # 处理config字段是字符串的情况
        if isinstance(updated_config.get("config"), str):
            try:
                updated_config["config"] = json.loads(updated_config["config"])
            except (json.JSONDecodeError, TypeError) as e:
                print(f"Failed to parse agent config string: {e}")
                return (400, {"error": "Invalid agent config format"})

        # 配置memory功能
        if config.get("need_config_memory"):
            if "memory" not in updated_config["config"]:
                updated_config["config"]["memory"] = {}
            updated_config["config"]["memory"]["is_enabled"] = True
            print(f"📄 Memory enabled")

        # 配置related_question功能
        if config.get("need_config_related_question"):
            if "related_question" not in updated_config["config"]:
                updated_config["config"]["related_question"] = {}
            updated_config["config"]["related_question"]["is_enabled"] = True
            print(f"📄 Related question enabled")

        # 配置plan_mode功能
        if config.get("need_config_plan_mode"):
            if "plan_mode" not in updated_config["config"]:
                updated_config["config"]["plan_mode"] = {}
            updated_config["config"]["plan_mode"]["is_enabled"] = True
            print(f"📄 Plan mode enabled")

        # 执行更新
        return self.UpdateAgent(agent_id, updated_config, headers)

    def CreateAgent(self, agent_config, headers=None):
        """
        创建智能体
        :param agent_config: 智能体配置数据
        :param headers: 请求头
        :return: (status_code, response_data)
        """
        url = self.base_url + "/agent-factory/v3/agent"
        return Request.post(self, url, agent_config, headers)

    def CopyAgent(self, agent_id, headers=None):
        """
        复制智能体
        :param agent_id: 智能体ID
        :param headers: 请求头
        :return: (status_code, response_data)
        """
        url = self.base_url + f"/agent-factory/v3/agent/{agent_id}/copy"
        return Request.post(self, url, None, headers)

    def CopyAgentToTemplate(self, agent_id, headers=None):
        """
        复制智能体为模板
        :param agent_id: 智能体ID
        :param headers: 请求头
        :return: (status_code, response_data)
        """
        url = self.base_url + f"/agent-factory/v3/agent/{agent_id}/copy2tpl"
        return Request.post(self, url, None, headers)

    def CopyAgentToTemplateAndPublish(self, agent_id, publish_config, headers=None):
        """
        复制智能体为模板并发布
        :param agent_id: 智能体ID
        :param publish_config: 发布配置
        :param headers: 请求头
        :return: (status_code, response_data)
        """
        url = self.base_url + f"/agent-factory/v3/agent/{agent_id}/copy2tpl-and-publish"
        return Request.post(self, url, publish_config, headers)

    def GetAgentSelfConfigFields(self, headers=None):
        """
        获取SELF_CONFIG字段结构
        :param headers: 请求头
        :return: (status_code, response_data)
        """
        url = self.base_url + "/agent-factory/v3/agent-self-config-fields"
        return Request.get(self, url, headers)

    def GetBatchAgentFields(self, agent_ids, fields, headers=None):
        """
        批量获取agent指定字段
        :param agent_ids: agent id列表
        :param fields: agent字段列表
        :param headers: 请求头
        :return: (status_code, response_data)
        """
        url = self.base_url + "/agent-factory/v3/agent-fields"
        body = {
            "agent_ids": agent_ids,
            "fields": fields
        }
        return Request.post(self, url, body, headers)

    def PublishAgent(self, agent_id, publish_config, headers=None):
        """
        发布智能体
        :param agent_id: 智能体ID
        :param publish_config: 发布配置
        :param headers: 请求头
        :return: (status_code, response_data)
        """
        url = self.base_url + f"/agent-factory/v3/agent/{agent_id}/publish"
        return Request.post(self, url, publish_config, headers)

    def UnpublishAgent(self, agent_id, headers=None):
        """
        取消发布智能体
        :param agent_id: 智能体ID
        :param headers: 请求头
        :return: (status_code, response_data)
        """
        url = self.base_url + f"/agent-factory/v3/agent/{agent_id}/unpublish"
        return Request.put(self, url, None, headers)

    def GetPublishInfo(self, agent_id, headers=None):
        """
        获取已发布智能体的发布信息
        :param agent_id: 智能体ID
        :param headers: 请求头
        :return: (status_code, response_data)
        """
        url = self.base_url + f"/agent-factory/v3/agent/{agent_id}/publish-info"
        return Request.get(self, url, headers)

    def UpdatePublishInfo(self, agent_id, publish_config, headers=None):
        """
        更新发布信息
        :param agent_id: 智能体ID
        :param publish_config: 发布配置
        :param headers: 请求头
        :return: (status_code, response_data)
        """
        url = self.base_url + f"/agent-factory/v3/agent/{agent_id}/publish-info"
        return Request.put(self, url, publish_config, headers)

    def GetCategory(self, headers=None):
        """
        获取智能体分类
        :param headers: 请求头
        :return: (status_code, response_data)
        """
        url = self.base_url + "/agent-factory/v3/category"
        return Request.get(self, url, headers)

    def GetPublishedAgentList(self, request_body, headers=None):
        """
        已发布智能体列表
        :param request_body: 请求体
        :param headers: 请求头
        :return: (status_code, response_data)
        """
        url = self.base_url + "/agent-factory/v3/published/agent"
        return Request.post(self, url, request_body, headers)

    def GetPublishedAgentTplList(self, params=None, headers=None):
        """
        已发布模板列表
        :param params: 查询参数
        :param headers: 请求头
        :return: (status_code, response_data)
        """
        url = self.base_url + "/agent-factory/v3/published/agent-tpl"
        if params is None:
            params = {}
        return Request.query(self, url, params, headers)

    def GetPublishedAgentTplDetail(self, tpl_id, headers=None):
        """
        已发布模板详情
        :param tpl_id: 模板ID
        :param headers: 请求头
        :return: (status_code, response_data)
        """
        url = self.base_url + f"/agent-factory/v3/published/agent-tpl/{tpl_id}"
        return Request.get(self, url, headers)

    def GetPublishedAgentInfoList(self, request_body, headers=None):
        """
        已发布智能体信息列表
        :param request_body: 请求体
        :param headers: 请求头
        :return: (status_code, response_data)
        """
        url = self.base_url + "/agent-factory/v3/published/agent-info-list"
        return Request.post(self, url, request_body, headers)

    def GetAgentMarketAgentInfo(self, agent_id, version, params=None, headers=None):
        """
        智能体详情（已发布或未发布）
        :param agent_id: agent ID / agent key
        :param version: agent 版本
        :param params: 查询参数
        :param headers: 请求头
        :return: (status_code, response_data)
        """
        url = self.base_url + f"/agent-factory/v3/agent-market/agent/{agent_id}/version/{version}"
        if params is None:
            params = {}
        return Request.query(self, url, params, headers)

    def GetReleaseHistory(self, agent_id, headers=None):
        """
        获取发布历史记录列表
        :param agent_id: 智能体ID
        :param headers: 请求头
        :return: (status_code, response_data)
        """
        url = self.base_url + f"/agent-factory/v3/agent/{agent_id}/release-history"
        return Request.get(self, url, headers)

    def CheckAgentPermission(self, request_body, headers=None):
        """
        检查某个agent是否有执行（使用）权限
        :param request_body: 请求体
        :param headers: 请求头
        :return: (status_code, response_data)
        """
        url = self.base_url + "/agent-factory/v3/agent-permission/execute"
        return Request.post(self, url, request_body, headers)

    def GetPermissionUserStatus(self, headers=None):
        """
        获取用户拥有的管理权限状态
        :param headers: 请求头
        :return: (status_code, response_data)
        """
        url = self.base_url + "/agent-factory/v3/agent-permission/management/user-status"
        return Request.get(self, url, headers)

    def GetPermissionListBtn(self, resource_type, from_scene, custom_space_id=None, headers=None):
        """
        获取列表中有权限的操作按钮
        :param resource_type: 资源类型 (agent, agent_tpl)
        :param from_scene: 场景 (personal_space, custom_space, square)
        :param custom_space_id: 自定义空间ID
        :param headers: 请求头
        :return: (status_code, response_data)
        """
        url = self.base_url + "/agent-factory/v3/agent-permission/list-btn"
        params = {
            "resource_type": resource_type,
            "from": from_scene
        }
        if custom_space_id:
            params["custom_space_id"] = custom_space_id
        return Request.query(self, url, params, headers)

    def GetProductList(self, params=None, headers=None):
        """
        获取产品列表
        :param params: 查询参数
        :param headers: 请求头
        :return: (status_code, response_data)
        """
        url = self.base_url + "/agent-factory/v3/product"
        if params is None:
            params = {}
        return Request.query(self, url, params, headers)

    def CreateProduct(self, product_config, headers=None):
        """
        创建产品
        :param product_config: 产品配置
        :param headers: 请求头
        :return: (status_code, response_data)
        """
        url = self.base_url + "/agent-factory/v3/product"
        return Request.post(self, url, product_config, headers)

    def GetProductDetail(self, product_id, headers=None):
        """
        获取产品详情
        :param product_id: 产品ID
        :param headers: 请求头
        :return: (status_code, response_data)
        """
        url = self.base_url + f"/agent-factory/v3/product/{product_id}"
        return Request.get(self, url, headers)

    def UpdateProduct(self, product_id, product_config, headers=None):
        """
        编辑产品
        :param product_id: 产品ID
        :param product_config: 产品配置
        :param headers: 请求头
        :return: (status_code, response_data)
        """
        url = self.base_url + f"/agent-factory/v3/product/{product_id}"
        return Request.put(self, url, product_config, headers)

    def DeleteProduct(self, product_id, headers=None):
        """
        删除产品
        :param product_id: 产品ID
        :param headers: 请求头
        :return: (status_code, response_data)
        """
        url = self.base_url + f"/agent-factory/v3/product/{product_id}"
        return Request.delete(self, url, None, headers)

    def GetPersonalSpaceAgentList(self, params=None, headers=None):
        """
        个人空间智能体列表
        :param params: 查询参数
        :param headers: 请求头
        :return: (status_code, response_data)
        """
        url = self.base_url + "/agent-factory/v3/personal-space/agent-list"
        if params is None:
            params = {}
        return Request.query(self, url, params, headers)

    def GetPersonalSpaceAgentTplList(self, params=None, headers=None):
        """
        个人空间智能体模板列表
        :param params: 查询参数
        :param headers: 请求头
        :return: (status_code, response_data)
        """
        url = self.base_url + "/agent-factory/v3/personal-space/agent-tpl-list"
        if params is None:
            params = {}
        return Request.query(self, url, params, headers)

    def GetRecentVisitAgentList(self, params=None, headers=None):
        """
        最近访问智能体列表
        :param params: 查询参数
        :param headers: 请求头
        :return: (status_code, response_data)
        """
        url = self.base_url + "/agent-factory/v3/recent-visit/agent"
        if params is None:
            params = {}
        return Request.query(self, url, params, headers)

    def ExportAgents(self, agent_ids, headers=None):
        """
        导出智能体数据
        :param agent_ids: 智能体ID列表
        :param headers: 请求头
        :return: (status_code, response_data)
        """
        url = self.base_url + "/agent-factory/v3/agent-inout/export"
        body = {
            "agent_ids": agent_ids
        }
        return Request.post(self, url, body, headers)

    def GetBuiltInAvatarList(self, headers=None):
        """
        获取内置头像列表
        :param headers: 请求头
        :return: (status_code, response_data)
        """
        url = self.base_url + "/agent-factory/v3/agent/avatar/built-in"
        return Request.get(self, url, headers)

    def GetBuiltInAvatar(self, avatar_id, headers=None):
        """
        获取内置头像
        :param avatar_id: 头像ID
        :param headers: 请求头
        :return: (status_code, response_data)
        """
        url = self.base_url + f"/agent-factory/v3/agent/avatar/built-in/{avatar_id}"
        return Request.get(self, url, headers)

    def GetFileExtMap(self, headers=None):
        """
        获取文件扩展名映射
        :param headers: 请求头
        :return: (status_code, response_data)
        """
        url = self.base_url + "/agent-factory/v3/agent/temp-zone/file-ext-map"
        return Request.get(self, url, headers)

    def BatchCheckIndexStatus(self, agent_uniq_flags, is_show_fail_infos=False, headers=None):
        """
        批量获取数据处理状态
        :param agent_uniq_flags: agent唯一标识列表
        :param is_show_fail_infos: 是否显示失败信息
        :param headers: 请求头
        :return: (status_code, response_data)
        """
        url = self.base_url + "/agent-factory/v3/agent/batch-check-index-status"
        body = {
            "agent_uniq_flags": agent_uniq_flags,
            "is_show_fail_infos": is_show_fail_infos
        }
        return Request.post(self, url, body, headers)

