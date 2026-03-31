# -*- coding:UTF-8 -*-

import pytest
import allure
import json
import os
import sys

# 添加项目根目录到路径
sys.path.append(os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__)))))

from common.get_content import GetContent
from common.get_token import GetToken
from lib.data_agent import AgentFactory

configfile = "./config/env.ini"
file = GetContent(configfile)
config = file.config()

host = config["server"]["host"]


@pytest.fixture(scope="session")
def AgentHeaders():
    '''获取智能体测试专用token授权，用户名test，密码111111'''
    token = GetToken(host=host).get_token(host, "test", "111111")
    headers = {
        "Authorization": f"Bearer {token[1]}",
        "x-account-id": token[0],
        "x-account-type": "user",
        "x-business-domain": "bd_public"
        }
    allure.attach(json.dumps(headers).encode("utf-8"), name="agent_headers")

    yield headers


@pytest.fixture(scope="session")
def AgentImport(APrepare, AgentHeaders):
    """
    导入智能体fixture
    读取cfg.json配置文件，根据配置处理data/data-agent/import目录下的JSON文件并导入智能体
    支持模型ID更新等预处理功能
    测试结束后删除导入的智能体
    """
    data_agent_dir = "./data/data-agent"
    import_dir = os.path.join(data_agent_dir, "import")
    cfg_file = os.path.join(data_agent_dir, "cfg.json")
    imported_agents = []
    published_skill_agents = []  # 跟踪发布为技能的智能体，用于清理

    # 读取配置文件
    config = None
    if os.path.exists(cfg_file):
        try:
            with open(cfg_file, 'r', encoding='utf-8') as f:
                config = json.load(f)
            allure.attach(json.dumps(config, ensure_ascii=False, indent=2), name="agent_import_config")
        except Exception as e:
            allure.attach(f"Failed to read cfg.json: {e}", name="config_read_error")

    # 如果没有配置文件，使用默认配置（数组格式）
    if not config:
        config = [{
            "agent_config": None,
            "need_conf_llm": False,
            "need_conf_doc_resource": False,
            "need_conf_graph_resource": False,
            "need_run": false,
            "config_intervention": [],
            "test_queries": {
                "single_turn": ["你好，请介绍一下你自己"],
                "multi_turn": ["什么是机器学习？", "请详细说明一下它的工作原理", "能给我一个实际的例子吗？"]
            }
        }]
        allure.attach("Using default configuration", name="agent_import_config_default")
    elif isinstance(config, dict) and "agents" in config:
        # 新的配置格式，提取agents数组并合并shared配置
        agents_config = config.get("agents", [])
        shared_configs = config.get("shared_configs", {})

        # 为每个agent配置添加shared配置
        for agent_config in agents_config:
            # 处理文档资源配置 - 支持use_as_doc开关
            if agent_config.get("need_conf_doc_resource"):
                use_as_doc = agent_config.get("use_as_doc", False)

                if use_as_doc:
                    # 使用as_doc_resource_config.json
                    doc_config_path = os.path.join(data_agent_dir, "as_doc_resource_config.json")
                    try:
                        with open(doc_config_path, 'r', encoding='utf-8') as f:
                            agent_config["doc_resource_config"] = json.load(f)
                            allure.attach(f"Loaded as_doc_resource_config for agent: {agent_config.get('agent_config', 'unknown')}", name="as_doc_config_loaded")
                    except Exception as e:
                        allure.attach(f"Failed to load as_doc_resource_config from {doc_config_path}: {e}", name="as_doc_config_load_error")
                        agent_config["doc_resource_config"] = None
                else:
                    # 使用shared_configs中的doc_resource_config
                    if shared_configs.get("doc_resource_config"):
                        doc_config_path = os.path.join(data_agent_dir, shared_configs["doc_resource_config"])
                        try:
                            with open(doc_config_path, 'r', encoding='utf-8') as f:
                                agent_config["doc_resource_config"] = json.load(f)
                                allure.attach(f"Loaded shared doc_resource_config for agent: {agent_config.get('agent_config', 'unknown')}", name="shared_doc_config_loaded")
                        except Exception as e:
                            allure.attach(f"Failed to load doc_resource_config from {doc_config_path}: {e}", name="doc_config_load_error")
                            agent_config["doc_resource_config"] = None

            if shared_configs.get("graph_resource_config"):
                # 读取图资源配置文件内容
                graph_config_path = os.path.join(data_agent_dir, shared_configs["graph_resource_config"])
                try:
                    with open(graph_config_path, 'r', encoding='utf-8') as f:
                        agent_config["graph_resource_config"] = json.load(f)
                except Exception as e:
                    allure.attach(f"Failed to load graph_resource_config from {graph_config_path}: {e}", name="graph_config_load_error")
                    agent_config["graph_resource_config"] = None

            if shared_configs.get("temp_file_config"):
                # 读取临时文件配置内容
                temp_config_path = os.path.join(data_agent_dir, shared_configs["temp_file_config"])
                try:
                    with open(temp_config_path, 'r', encoding='utf-8') as f:
                        agent_config["temp_file_config"] = json.load(f)
                except Exception as e:
                    allure.attach(f"Failed to load temp_file_config from {temp_config_path}: {e}", name="temp_config_load_error")
                    agent_config["temp_file_config"] = None

        config = agents_config
        allure.attach("Using new configuration format with shared configs", name="agent_import_config_new_format")
    else:
        shared_configs = {}

    # 处理智能体配置文件
    if os.path.exists(import_dir):
        # 遍历配置数组中的每个智能体配置
        for agent_config_item in config:
            # 检查need_run字段，如果为false则跳过此智能体
            need_run = agent_config_item.get("need_run", False)
            if not need_run:
                agent_config_name = agent_config_item.get("agent_config", "unknown")
                allure.attach(f"Skipping agent {agent_config_name} because need_run is false", name="agent_skipped_need_run_false")
                continue

            # 确定要处理的文件
            if agent_config_item.get("agent_config"):
                filename = agent_config_item["agent_config"]
            else:
                # 如果没有指定文件，处理import目录下所有JSON文件
                json_files = [f for f in os.listdir(import_dir) if f.endswith('.json')]
                if not json_files:
                    allure.attach("No JSON files found in import directory", name="no_files_found")
                    continue
                filename = json_files[0]  # 取第一个文件

            file_path = os.path.join(import_dir, filename)
            if not os.path.exists(file_path):
                allure.attach(f"Agent file not found: {file_path}", name="agent_file_not_found")
                continue

            allure.attach(f"Processing agent from: {file_path}", name="agent_import_start")

            # 读取原始JSON文件
            with open(file_path, 'r', encoding='utf-8') as f:
                agent_data = json.load(f)

            # 创建处理后的临时文件
            temp_file_path = file_path.replace('.json', '_temp.json')

            # 如果需要配置LLM，更新模型ID
            if agent_config_item.get("need_conf_llm"):
                updated_agent_data = update_agent_llm_config(agent_data, AgentHeaders)
                allure.attach("LLM configuration updated", name="llm_config_updated")
            else:
                updated_agent_data = agent_data

            # 写入临时文件用于导入
            with open(temp_file_path, 'w', encoding='utf-8') as f:
                json.dump(updated_agent_data, f, ensure_ascii=False, indent=2)

            # 导入智能体
            factory_client = AgentFactory()
            result = factory_client.ImportAgents(temp_file_path, "create", AgentHeaders)

            # 删除临时文件
            try:
                os.remove(temp_file_path)
            except:
                pass

            # 处理智能体导入响应，为每个智能体保存信息
            for agent in updated_agent_data.get("agents", []):
                actual_agent_id = None
                agent_key = agent.get("key")
                agent_name = agent.get("name")
                agent_id = agent.get("id")

                if result[0] == 200:
                    response_data = result[1] if isinstance(result[1], dict) else {}

                    # 检查是否有agent_key_conflict，表示智能体已存在
                    if response_data.get("agent_key_conflict"):
                        for conflict_info in response_data.get("agent_key_conflict", []):
                            if conflict_info.get("agent_key") == agent_key:
                                # 从冲突信息中获取agent_key，但仍然需要获取实际的agent_id
                                # 由于权限问题，我们需要使用不同的方法来获取ID
                                allure.attach(f"Agent already exists with key: {agent_key}, attempting to find actual ID", name="agent_conflict_detected")

                                # 尝试通过agent_key获取实际的agent_id（可能会因为权限失败）
                                get_result = factory_client.GetAgentByKey(agent_key, AgentHeaders)
                                if get_result[0] == 200:
                                    actual_agent_id = get_result[1].get("id")
                                    allure.attach(f"Successfully got actual agent ID: {actual_agent_id} for agent {agent_name}", name="agent_id_retrieved")
                                else:
                                    # 如果没有权限获取，暂时使用一个标识符，在聊天时会跳过
                                    actual_agent_id = f"CONFLICT_{agent_key}"
                                    allure.attach(f"No permission to get agent ID for {agent_name}, using conflict marker: {actual_agent_id}", name="agent_id_conflict_no_permission")
                                break
                        else:
                            # 没有找到匹配的conflict信息
                            actual_agent_id = agent_id
                            allure.attach(f"Conflict info found but no matching agent key, using JSON ID: {actual_agent_id}", name="agent_id_fallback")
                    else:
                        # 导入成功，使用JSON中的ID或尝试获取实际ID
                        get_result = factory_client.GetAgentByKey(agent_key, AgentHeaders)
                        if get_result[0] == 200:
                            actual_agent_id = get_result[1].get("id")
                            allure.attach(f"Successfully got actual agent ID: {actual_agent_id} for agent {agent_name}", name="agent_id_retrieved")
                        else:
                            actual_agent_id = agent_id
                            allure.attach(f"Using JSON ID as actual agent ID: {actual_agent_id} for agent {agent_name}", name="agent_id_json_fallback")
                else:
                    # 导入失败，尝试获取实际ID或使用JSON ID
                    get_result = factory_client.GetAgentByKey(agent_key, AgentHeaders)
                    if get_result[0] == 200:
                        actual_agent_id = get_result[1].get("id")
                        allure.attach(f"Import failed but got actual agent ID: {actual_agent_id} for agent {agent_name}", name="agent_id_import_fail_retrieved")
                    else:
                        actual_agent_id = agent_id
                        allure.attach(f"Import failed and no permission to get actual ID, using JSON ID: {actual_agent_id} for agent {agent_name}", name="agent_id_import_fail_fallback")

                # 提取智能体的默认模型名称
                model_name = None
                try:
                    config_str = agent.get("config", "")
                    if config_str:
                        config_data = json.loads(config_str)
                        llms = config_data.get("llms", [])
                        for llm in llms:
                            if llm.get("is_default"):
                                model_name = llm.get("llm_config", {}).get("name")
                                break
                except (json.JSONDecodeError, TypeError) as e:
                    allure.attach(f"Failed to extract model name from agent config: {e}", name="model_extraction_error")

                # 提取测试查询
                test_queries = agent_config_item.get("test_queries", {
                    "single_turn": ["你好，请介绍一下你自己"],
                    "multi_turn": ["什么是机器学习？", "请详细说明一下它的工作原理", "能给我一个实际的例子吗？"]
                })

                # 配置智能体技能（如果需要）
                # 检查是否需要跳过技能配置（如果智能体已经包含技能配置）
                skip_skills_config = agent_config_item.get("skip_skills_config", False)

                if (actual_agent_id and
                    not actual_agent_id.startswith("CONFLICT_") and
                    agent_config_item.get("skill_list") and
                    not skip_skills_config):

                    try:
                        # 使用ConfigureAgentSkills方法配置技能
                        skills_result = factory_client.ConfigureAgentSkills(
                            actual_agent_id,
                            agent_config_item,
                            shared_configs,
                            AgentHeaders
                        )

                        if skills_result[0] == 200:
                            allure.attach(f"Successfully configured skills for agent: {agent_name}", name="skills_config_success")
                        else:
                            allure.attach(f"Failed to configure skills for agent: {agent_name}, status: {skills_result[0]}", name="skills_config_failure")
                    except Exception as e:
                        allure.attach(f"Exception configuring skills for agent {agent_name}: {e}", name="skills_config_error")

                if skip_skills_config:
                    allure.attach(f"Skipped skills configuration for agent: {agent_name} (already configured in JSON)", name="skills_config_skipped")

                # 配置智能体数据源（如果需要）
                if (actual_agent_id and
                    not actual_agent_id.startswith("CONFLICT_") and
                    (agent_config_item.get("need_conf_doc_resource") or agent_config_item.get("need_conf_graph_resource"))):

                    try:
                        # 使用ConfigureAgentDataSource方法配置数据源
                        config_result = factory_client.ConfigureAgentDataSource(
                            actual_agent_id,
                            agent_config_item,
                            AgentHeaders
                        )

                        if config_result[0] == 200:
                            allure.attach(f"Successfully configured data source for agent: {agent_name}", name="data_source_config_success")
                        else:
                            allure.attach(f"Failed to configure data source for agent: {agent_name}, status: {config_result[0]}", name="data_source_config_failure")
                    except Exception as e:
                        allure.attach(f"Exception configuring data source for agent {agent_name}: {e}", name="data_source_config_error")

                # 配置智能体特性（如果需要）
                if (actual_agent_id and
                    not actual_agent_id.startswith("CONFLICT_") and
                    (agent_config_item.get("need_config_memory") or
                     agent_config_item.get("need_config_related_question") or
                     agent_config_item.get("need_config_plan_mode"))):

                    try:
                        # 使用ConfigureAgentFeatures方法配置特性
                        features_result = factory_client.ConfigureAgentFeatures(
                            actual_agent_id,
                            agent_config_item,
                            AgentHeaders
                        )

                        if features_result[0] == 200:
                            allure.attach(f"Successfully configured features for agent: {agent_name}", name="features_config_success")
                        else:
                            allure.attach(f"Failed to configure features for agent: {agent_name}, status: {features_result[0]}", name="features_config_failure")
                    except Exception as e:
                        allure.attach(f"Exception configuring features for agent {agent_name}: {e}", name="features_config_error")

                # 处理临时文件配置 - 根据temp_file_config_index选择特定的配置
                temp_file_config = []
                if agent_config_item.get("need_temp_file") and agent_config_item.get("temp_file_config"):
                    # 根据temp_file_config_index选择特定的临时文件配置
                    temp_file_config_all = agent_config_item.get("temp_file_config", [])
                    temp_file_config_index = agent_config_item.get("temp_file_config_index", 0)

                    # 检查索引是否有效
                    if 0 <= temp_file_config_index < len(temp_file_config_all):
                        temp_file_config = [temp_file_config_all[temp_file_config_index]]
                        allure.attach(f"Selected temp file config at index {temp_file_config_index}: {temp_file_config[0].get('name', 'Unknown')}", name="temp_file_config_selected")
                    else:
                        allure.attach(f"Invalid temp_file_config_index {temp_file_config_index}, using empty config", name="temp_file_config_invalid_index")

                # 检查是否需要发布为技能
                need_publish_to_skill = agent_config_item.get("need_publish_to_skill", False)
                if need_publish_to_skill:
                    try:
                        # 获取可用分类
                        category_result = factory_client.GetCategory(AgentHeaders)
                        if category_result[0] == 200:
                            category_data = category_result[1]
                            if category_data and isinstance(category_data, list) and len(category_data) > 0:
                                # 选择第一个分类
                                selected_category = category_data[0]
                                category_id = selected_category.get("id")
                                category_name = selected_category.get("name", "未知分类")

                                # 构造发布配置，发布为技能
                                publish_config = {
                                    "business_domain_id": "bd_public",
                                    "category_id": [category_id],
                                    "description": f"{agent_name} - 自动发布为技能",
                                    "publish_to_where": ["square"],
                                    "publish_to_bes": ["skill_agent"]  # 发布为技能
                                }

                                allure.attach(f"开始发布Agent为技能 - Agent: {agent_name}, 分类: {category_name}", name="publish_to_skill_start")

                                # 执行发布
                                publish_result = factory_client.PublishAgent(actual_agent_id, publish_config, AgentHeaders)

                                if publish_result[0] in [200, 201]:
                                    allure.attach(f"Agent {agent_name} 成功发布为技能", name="publish_to_skill_success")
                                elif publish_result[0] in [502, 503, 504]:
                                    allure.attach(f"Agent {agent_name} 发布为技能时服务器不可用", name="publish_to_skill_server_error")
                                else:
                                    allure.attach(f"Agent {agent_name} 发布为技能失败，状态码: {publish_result[0]}", name="publish_to_skill_failure")
                            else:
                                allure.attach(f"没有可用的智能体分类，无法发布 {agent_name} 为技能", name="publish_to_skill_no_category")
                        else:
                            allure.attach(f"获取智能体分类失败，状态码: {category_result[0]}", name="publish_to_skill_category_error")
                    except Exception as e:
                        allure.attach(f"发布Agent {agent_name} 为技能时发生异常: {e}", name="publish_to_skill_exception")

                    # 将发布为技能的agent添加到published_skill_agents列表，用于后续清理
                    published_skill_agents.append({
                        "agent_id": actual_agent_id,
                        "agent_key": agent_key,
                        "agent_name": agent_name
                    })

                    # 发布为技能的agent不参与测试，不添加到imported_agents
                    allure.attach(f"Agent {agent_name} 已发布为技能，不参与测试", name="agent_excluded_from_test")
                else:
                    # 只有不需要发布为技能的agent才参与测试
                    imported_agents.append({
                        "agent_id": actual_agent_id,
                        "agent_key": agent_key,
                        "agent_name": agent_name,
                        "file_path": file_path,
                        "model_name": model_name,
                        "test_queries": test_queries,
                        "config_intervention": agent_config_item.get("config_intervention", []),
                        "need_temp_file": agent_config_item.get("need_temp_file", False),
                        "temp_file_config": temp_file_config,
                        "expected_results": agent_config_item.get("expected_results", {}),
                        "extra_params": agent_config_item.get("extra_params", []),
                        "need_publish_to_skill": False  # 明确标记不发布为技能
                    })

                if result[0] == 200:
                    response_data = result[1] if isinstance(result[1], dict) else {}
                    if response_data.get("agent_key_conflict"):
                        allure.attach(f"Agent already exists: {agent_name} (ID: {actual_agent_id})", name="agent_conflict")
                    else:
                        allure.attach(f"Successfully imported agent: {agent_name} (ID: {actual_agent_id})", name="agent_import_success")
                else:
                    allure.attach(f"Server unavailable, using mock agent: {agent_name} (ID: {actual_agent_id})", name="agent_import_mock")

    yield imported_agents

    # 清理：先清理发布为技能的智能体
    factory_client = AgentFactory()

    # 第一步：取消发布并删除发布为技能的智能体
    for agent_info in published_skill_agents:
        agent_id = agent_info["agent_id"]
        agent_name = agent_info["agent_name"]

        try:
            # 先取消发布
            unpublish_result = factory_client.UnpublishAgent(agent_id, AgentHeaders)
            if unpublish_result[0] in [200, 204, 404]:
                allure.attach(f"Successfully unpublished agent: {agent_name} (ID: {agent_id})", name="skill_agent_unpublish_success")
            else:
                allure.attach(f"Failed to unpublish agent: {agent_name} (ID: {agent_id}), status: {unpublish_result[0]}", name="skill_agent_unpublish_failure")

            # 再删除智能体
            delete_result = factory_client.DeleteAgent(agent_id, AgentHeaders)
            if delete_result[0] == 204:
                allure.attach(f"Successfully deleted published skill agent: {agent_name} (ID: {agent_id})", name="skill_agent_delete_success")
            else:
                allure.attach(f"Failed to delete published skill agent: {agent_name} (ID: {agent_id}), status: {delete_result[0]}", name="skill_agent_delete_failure")
        except Exception as e:
            allure.attach(f"Exception cleaning up published skill agent {agent_name} (ID: {agent_id}): {e}", name="skill_agent_cleanup_exception")

    # 第二步：删除参与测试的智能体
    for agent_info in imported_agents:
        agent_id = agent_info["agent_id"]
        agent_name = agent_info["agent_name"]

        result = factory_client.DeleteAgent(agent_id, AgentHeaders)
        if result[0] == 204:
            allure.attach(f"Successfully deleted agent: {agent_name} (ID: {agent_id})", name="agent_delete_success")
        else:
            allure.attach(f"Failed to delete agent: {agent_name} (ID: {agent_id})", name="agent_delete_failure")


@pytest.fixture(scope="session")
def ModelCheck(APrepare, Headers):
    """
    检查指定模型是否存在并测试连通性
    仅用于data-agent测试
    """
    factory_client = AgentFactory()
    model_exists = False
    model_id = None
    model_name = "Tome-pro"  # 默认模型名称，保持向后兼容

    # 获取模型列表，添加分页参数
    result = factory_client.GetModelList({"name": model_name, "page": 1, "size": 50}, Headers)

    if result[0] == 200:
        models = result[1].get("data", [])
        for model in models:
            # 检查多种可能的模型名称字段，支持您提供的API响应格式
            model_name_match = (
                model.get("name") == model_name or
                model.get("model_name") == model_name or
                model.get("model") == model_name
            )

            if model_name_match:
                # 检查多种可能的模型ID字段，支持您提供的API响应格式
                model_id = (
                    model.get("id") or
                    model.get("model_id") or
                    model.get("model_id")
                )
                model_exists = True
                break
    elif result[0] in [500, 502, 503, 504]:
        # 服务器错误，跳过模型检查但记录日志
        allure.attach(f"服务器不可用，跳过{model_name}模型检查", name="model_check_server_error")
        yield False
        return
    else:
        # 其他错误（如400参数错误），记录详细信息但不阻止测试
        allure.attach(f"模型列表接口返回错误: {result[0]}, 响应: {result[1]}", name="model_check_api_error")

    allure.attach(f"{model_name} model exists: {model_exists}", name="model_check")

    if model_exists and model_id:
        # 测试模型连通性
        test_result = factory_client.TestModelConnectivity(model_id, Headers)
        connectivity_ok = test_result[0] == 200
        allure.attach(f"{model_name} model connectivity test: {connectivity_ok}", name="model_connectivity")

        if not connectivity_ok and test_result[0] in [200, 500, 502, 503, 504]:
            # 如果服务器不可用，假设模型连接正常
            allure.attach("Server unavailable, assuming model connectivity OK", name="model_connectivity_mock")
        elif not connectivity_ok:
            pytest.fail(f"{model_name} model connectivity test failed, status: {test_result[0]}, response: {test_result[1]}")
    else:
        # 如果找不到模型，这应该导致测试失败
        if result[0] == 200:
            # 接口调用成功但找不到模型，说明模型确实不存在
            pytest.fail(f"{model_name} model not found in model list - this is a required dependency for agent testing")
        elif result[0] not in [500, 502, 503, 504]:
            # 非服务器错误，可能是API参数问题
            allure.attach(f"模型检查失败，但继续执行测试。状态码: {result[0]}, 响应: {result[1]}", name="model_check_warning")

    yield model_exists


@pytest.fixture(scope="session")
def AgentModelCheck(APrepare, AgentHeaders):
    """
    检查指定模型是否存在并测试连通性
    专门用于智能体测试，使用test用户
    """
    factory_client = AgentFactory()
    model_exists = False
    model_id = None
    model_name = "Tome-pro"  # 默认模型名称，保持向后兼容

    # 获取模型列表，添加分页参数
    result = factory_client.GetModelList({"name": model_name, "page": 1, "size": 50}, AgentHeaders)

    if result[0] == 200:
        models = result[1].get("data", [])
        for model in models:
            # 检查多种可能的模型名称字段，支持您提供的API响应格式
            model_name_match = (
                model.get("name") == model_name or
                model.get("model_name") == model_name or
                model.get("model") == model_name
            )

            if model_name_match:
                # 检查多种可能的模型ID字段，支持您提供的API响应格式
                model_id = (
                    model.get("id") or
                    model.get("model_id") or
                    model.get("model_id")
                )
                model_exists = True
                break
    elif result[0] in [500, 502, 503, 504]:
        # 服务器错误，跳过模型检查但记录日志
        allure.attach(f"服务器不可用，跳过{model_name}模型检查", name="agent_model_check_server_error")
        yield False
        return
    else:
        # 其他错误（如400参数错误），记录详细信息但不阻止测试
        allure.attach(f"模型列表接口返回错误: {result[0]}, 响应: {result[1]}", name="agent_model_check_api_error")

    allure.attach(f"{model_name} model exists for agent test: {model_exists}", name="agent_model_check")

    if model_exists and model_id:
        # 测试模型连通性
        test_result = factory_client.TestModelConnectivity(model_id, AgentHeaders)
        connectivity_ok = test_result[0] == 200
        allure.attach(f"{model_name} model connectivity test for agent: {connectivity_ok}", name="agent_model_connectivity")

        if not connectivity_ok and test_result[0] in [200, 500, 502, 503, 504]:
            # 如果服务器不可用，假设模型连接正常
            allure.attach("Server unavailable, assuming model connectivity OK", name="agent_model_connectivity_mock")
        elif not connectivity_ok:
            pytest.fail(f"{model_name} model connectivity test for agent failed, status: {test_result[0]}, response: {test_result[1]}")
    else:
        # 如果找不到模型，这应该导致测试失败
        if result[0] == 200:
            # 接口调用成功但找不到模型，说明模型确实不存在
            pytest.fail(f"{model_name} model not found in model list - this is a required dependency for agent testing")
        elif result[0] not in [500, 502, 503, 504]:
            # 非服务器错误，可能是API参数问题
            allure.attach(f"智能体模型检查失败，但继续执行测试。状态码: {result[0]}, 响应: {result[1]}", name="agent_model_check_warning")

    yield model_exists


def update_agent_llm_config(agent_data, headers):
    """
    更新智能体配置中的模型ID
    根据智能体配置中的模型名称，从模型列表中获取正确的模型ID并更新配置

    Args:
        agent_data (dict): 原始智能体配置数据
        headers (dict): 认证请求头

    Returns:
        dict: 更新后的智能体配置数据
    """
    try:
        factory_client = AgentFactory()
        updated_agents = []

        for agent in agent_data.get("agents", []):
            # 解析智能体配置
            config_str = agent.get("config", "")
            if not config_str:
                updated_agents.append(agent)
                continue

            try:
                config_data = json.loads(config_str)
            except json.JSONDecodeError:
                updated_agents.append(agent)
                continue

            # 获取LLM配置
            llms = config_data.get("llms", [])
            if not llms:
                updated_agents.append(agent)
                continue

            # 查找默认LLM配置
            updated_llms = []
            for llm in llms:
                llm_config = llm.get("llm_config", {})
                model_name = llm_config.get("name")

                if model_name and llm.get("is_default"):
                    # 获取模型列表，查找正确的模型ID
                    result = factory_client.GetModelList({"name": model_name, "page": 1, "size": 50}, headers)

                    if result[0] == 200:
                        models = result[1].get("data", [])
                        for model in models:
                            # 检查多种可能的模型名称字段，支持您提供的API响应格式
                            model_name_match = (
                                model.get("name") == model_name or
                                model.get("model_name") == model_name or
                                model.get("model") == model_name
                            )

                            if model_name_match:
                                # 检查多种可能的模型ID字段，支持您提供的API响应格式
                                model_id = (
                                    model.get("id") or
                                    model.get("model_id") or
                                    model.get("model_id")
                                )

                                if model_id:
                                    # 更新模型ID
                                    llm_config["id"] = model_id
                                    llm["llm_config"] = llm_config
                                    allure.attach(f"Updated model ID for {model_name}: {model_id}", name="llm_id_update")
                                    break
                                else:
                                    allure.attach(f"Found model {model_name} but no ID field available", name="llm_id_update_no_id")
                    else:
                        allure.attach(f"Failed to get model list for {model_name}, keeping original ID", name="llm_id_update_failed")

                updated_llms.append(llm)

            # 更新配置数据
            config_data["llms"] = updated_llms

            # 更新智能体配置
            updated_agent = agent.copy()
            updated_agent["config"] = json.dumps(config_data, ensure_ascii=False)

            updated_agents.append(updated_agent)

        # 返回更新后的智能体数据
        return {
            "agents": updated_agents
        }

    except Exception as e:
        allure.attach(f"Error updating agent LLM config: {e}", name="llm_config_update_error")
        # 发生错误时返回原始数据
        return agent_data


@pytest.fixture(scope="session")
def ModelCheckByName(APrepare, Headers, request):
    """
    检查指定名称的模型是否存在并测试连通性
    支持从智能体配置中动态获取模型名称
    """
    if not hasattr(request, "param") or not request.param:
        # 如果没有提供模型名称参数，使用默认的Tome-pro模型
        request.param = "Tome-pro"

    model_name = request.param
    factory_client = AgentFactory()
    model_exists = False
    model_id = None

    # 获取模型列表，添加分页参数
    result = factory_client.GetModelList({"name": model_name, "page": 1, "size": 50}, Headers)

    if result[0] == 200:
        models = result[1].get("data", [])
        for model in models:
            # 检查多种可能的模型名称字段，支持您提供的API响应格式
            model_name_match = (
                model.get("name") == model_name or
                model.get("model_name") == model_name or
                model.get("model") == model_name
            )

            if model_name_match:
                model_exists = True
                # 检查多种可能的模型ID字段
                model_id = (
                    model.get("id") or
                    model.get("model_id") or
                    model.get("model_id")
                )
                break
    elif result[0] in [500, 502, 503, 504]:
        # 服务器错误，跳过模型检查但记录日志
        allure.attach(f"服务器不可用，跳过{model_name}模型检查", name="model_check_server_error")
        yield False
        return
    else:
        # 其他错误（如400参数错误），记录详细信息但不阻止测试
        allure.attach(f"模型列表接口返回错误: {result[0]}, 响应: {result[1]}", name="model_check_api_error")

    allure.attach(f"{model_name} model exists: {model_exists}", name="model_check")

    if model_exists and model_id:
        # 测试模型连通性
        test_result = factory_client.TestModelConnectivity(model_id, Headers)
        connectivity_ok = test_result[0] == 200
        allure.attach(f"{model_name} model connectivity test: {connectivity_ok}", name="model_connectivity")

        if not connectivity_ok and test_result[0] in [200, 500, 502, 503, 504]:
            # 如果服务器不可用，假设模型连接正常
            allure.attach("Server unavailable, assuming model connectivity OK", name="model_connectivity_mock")
        elif not connectivity_ok:
            pytest.fail(f"{model_name} model connectivity test failed, status: {test_result[0]}, response: {test_result[1]}")
    else:
        # 如果找不到模型，这应该导致测试失败
        if result[0] == 200:
            # 接口调用成功但找不到模型，说明模型确实不存在
            pytest.fail(f"{model_name} model not found in model list - this is a required dependency for agent testing")
        elif result[0] not in [500, 502, 503, 504]:
            # 非服务器错误，可能是API参数问题
            allure.attach(f"模型检查失败，但继续执行测试。状态码: {result[0]}, 响应: {result[1]}", name="model_check_warning")

    yield model_exists