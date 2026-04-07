# -*- coding:UTF-8 -*-

import pytest
import allure
import json
import time
import os

from lib.data_agent import AgentFactory


@allure.feature("Agent Factory V3 API 测试")
class TestAgentFactoryV3:
    """
    Agent Factory V3 API 测试类
    根据 agent-factory/v3 接口文档编写测试用例
    """

    @pytest.fixture(scope="class")
    def factory_client(self):
        """创建 AgentFactory 客户端"""
        return AgentFactory()

    @pytest.fixture(scope="class")
    def test_data(self):
        """加载测试数据文件"""
        data_file = os.path.join(
            os.path.dirname(os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))),
            "data", "data-agent", "test_agent_factory_v3_data.json"
        )
        with open(data_file, 'r', encoding='utf-8') as f:
            data = json.load(f)
        return data

    @pytest.fixture(scope="function")
    def test_agent_config(self, test_data):
        """测试用的智能体配置"""
        config = json.loads(json.dumps(test_data["agent_config"]))  # 深拷贝
        # 替换时间戳占位符，使用更精确的时间戳确保唯一性
        config["key"] = config["key"].replace("{timestamp}", str(int(time.time() * 1000)))
        return config

    @pytest.fixture(scope="function")
    def created_agent(self, factory_client, test_agent_config, AgentHeaders):
        """创建测试用的智能体（每个测试独立实例）"""
        result = factory_client.CreateAgent(test_agent_config, AgentHeaders)
        if result[0] == 201:
            agent_id = result[1].get("id")
            yield {"id": agent_id, "key": test_agent_config.get("key")}
        else:
            yield None
        
        # 清理：删除创建的智能体
        if result[0] == 201:
            agent_id = result[1].get("id")
            factory_client.DeleteAgent(agent_id, AgentHeaders)

    # ========== Agent 相关接口测试 ==========

    @allure.story("创建智能体")
    def test_create_agent(self, factory_client, test_agent_config, AgentHeaders):
        """测试创建智能体接口"""
        result = factory_client.CreateAgent(test_agent_config, AgentHeaders)
        
        assert result[0] == 201, f"创建智能体失败，状态码: {result[0]}, 响应: {result[1]}"
        assert "id" in result[1], "响应中缺少 id 字段"
        assert "version" in result[1], "响应中缺少 version 字段"
        
        # 清理
        agent_id = result[1].get("id")
        factory_client.DeleteAgent(agent_id, AgentHeaders)

    @allure.story("获取智能体详情")
    def test_get_agent(self, factory_client, created_agent, AgentHeaders):
        """测试获取智能体详情接口"""
        if not created_agent:
            pytest.skip("智能体创建失败，跳过测试")
        
        agent_id = created_agent["id"]
        result = factory_client.GetAgent(agent_id, AgentHeaders)
        
        assert result[0] == 200, f"获取智能体详情失败，状态码: {result[0]}, 响应: {result[1]}"
        assert result[1].get("id") == agent_id, "返回的智能体ID不匹配"

    @allure.story("根据key获取智能体详情")
    def test_get_agent_by_key(self, factory_client, created_agent, AgentHeaders):
        """测试根据key获取智能体详情接口"""
        if not created_agent:
            pytest.skip("智能体创建失败，跳过测试")
        
        agent_key = created_agent["key"]
        result = factory_client.GetAgentByKey(agent_key, AgentHeaders)
        
        assert result[0] == 200, f"根据key获取智能体详情失败，状态码: {result[0]}, 响应: {result[1]}"
        assert result[1].get("key") == agent_key, "返回的智能体key不匹配"

    @allure.story("更新智能体")
    def test_update_agent(self, factory_client, created_agent, AgentHeaders):
        """测试更新智能体接口"""
        if not created_agent:
            pytest.skip("智能体创建失败，跳过测试")
        
        agent_id = created_agent["id"]
        
        # 先获取当前配置
        get_result = factory_client.GetAgent(agent_id, AgentHeaders)
        assert get_result[0] == 200, "获取智能体详情失败"
        
        current_config = get_result[1]
        current_config["name"] = "更新后的测试智能体"
        
        result = factory_client.UpdateAgent(agent_id, current_config, AgentHeaders)
        
        assert result[0] == 204, f"更新智能体失败，状态码: {result[0]}, 响应: {result[1]}"
        
        # 验证更新是否成功
        verify_result = factory_client.GetAgent(agent_id, AgentHeaders)
        assert verify_result[0] == 200, "验证更新失败"
        assert verify_result[1].get("name") == "更新后的测试智能体", "更新后的名称不匹配"

    @allure.story("删除智能体")
    def test_delete_agent(self, factory_client, test_agent_config, AgentHeaders):
        """测试删除智能体接口"""
        # 先创建一个智能体
        create_result = factory_client.CreateAgent(test_agent_config, AgentHeaders)
        assert create_result[0] == 201, "创建智能体失败"
        
        agent_id = create_result[1].get("id")
        
        # 删除智能体
        result = factory_client.DeleteAgent(agent_id, AgentHeaders)
        
        assert result[0] == 204, f"删除智能体失败，状态码: {result[0]}, 响应: {result[1]}"
        
        # 验证删除是否成功
        verify_result = factory_client.GetAgent(agent_id, AgentHeaders)
        assert verify_result[0] in [404, 403], "智能体应该已被删除"

    @allure.story("复制智能体")
    def test_copy_agent(self, factory_client, created_agent, AgentHeaders):
        """测试复制智能体接口"""
        if not created_agent:
            pytest.skip("智能体创建失败，跳过测试")
        
        agent_id = created_agent["id"]
        result = factory_client.CopyAgent(agent_id, AgentHeaders)
        
        assert result[0] == 201, f"复制智能体失败，状态码: {result[0]}, 响应: {result[1]}"
        assert "id" in result[1], "响应中缺少 id 字段"
        
        # 清理复制的智能体
        copied_agent_id = result[1].get("id")
        factory_client.DeleteAgent(copied_agent_id, AgentHeaders)

    @allure.story("复制智能体为模板")
    def test_copy_agent_to_template(self, factory_client, created_agent, AgentHeaders):
        """测试复制智能体为模板接口"""
        if not created_agent:
            pytest.skip("智能体创建失败，跳过测试")
        
        agent_id = created_agent["id"]
        result = factory_client.CopyAgentToTemplate(agent_id, AgentHeaders)
        
        assert result[0] == 201, f"复制智能体为模板失败，状态码: {result[0]}, 响应: {result[1]}"
        assert "id" in result[1], "响应中缺少 id 字段"

    @allure.story("获取SELF_CONFIG字段结构")
    def test_get_agent_self_config_fields(self, factory_client, AgentHeaders):
        """测试获取SELF_CONFIG字段结构接口"""
        result = factory_client.GetAgentSelfConfigFields(AgentHeaders)
        
        assert result[0] == 200, f"获取SELF_CONFIG字段结构失败，状态码: {result[0]}, 响应: {result[1]}"
        assert isinstance(result[1], dict), "响应应该是字典类型"

    # ========== 发布相关接口测试 ==========

    @allure.story("获取智能体分类")
    def test_get_category(self, factory_client, AgentHeaders):
        """测试获取智能体分类接口"""
        result = factory_client.GetCategory(AgentHeaders)
        
        assert result[0] == 200, f"获取智能体分类失败，状态码: {result[0]}, 响应: {result[1]}"
        assert len(result[1])>1

    @allure.story("发布智能体")
    def test_publish_agent(self, factory_client, created_agent, test_data, AgentHeaders):
        """测试发布智能体接口"""
        if not created_agent:
            pytest.skip("智能体创建失败，跳过测试")
        
        agent_id = created_agent["id"]
        publish_config = json.loads(json.dumps(test_data["publish_config"]))  # 深拷贝
        
        result = factory_client.PublishAgent(agent_id, publish_config, AgentHeaders)
        
        # 发布可能成功或失败，取决于环境配置
        assert result[0] in [201, 400, 403], f"发布智能体返回意外的状态码: {result[0]}, 响应: {result[1]}"
        
        # 如果发布成功，尝试取消发布
        if result[0] == 201:
            factory_client.UnpublishAgent(agent_id, AgentHeaders)

    @allure.story("获取发布信息")
    def test_get_publish_info(self, factory_client, created_agent, AgentHeaders):
        """测试获取发布信息接口"""
        if not created_agent:
            pytest.skip("智能体创建失败，跳过测试")
        
        agent_id = created_agent["id"]
        result = factory_client.GetPublishInfo(agent_id, AgentHeaders)
        
        # 未发布的智能体会返回404
        assert result[0] in [200, 404], f"获取发布信息返回意外的状态码: {result[0]}, 响应: {result[1]}"

    # ========== 已发布相关接口测试 ==========

    @allure.story("已发布智能体列表")
    def test_get_published_agent_list(self, factory_client, test_data, AgentHeaders):
        """测试已发布智能体列表接口"""
        request_body = json.loads(json.dumps(test_data["published_agent_list_request"]))  # 深拷贝
        result = factory_client.GetPublishedAgentList(request_body, AgentHeaders)
        
        assert result[0] == 200, f"获取已发布智能体列表失败，状态码: {result[0]}, 响应: {result[1]}"
        assert "entries" in result[1], "响应中缺少 entries 字段"

    @allure.story("已发布模板列表")
    def test_get_published_agent_tpl_list(self, factory_client, AgentHeaders):
        """测试已发布模板列表接口"""
        params = {
            "page": 1,
            "size": 10
        }
        result = factory_client.GetPublishedAgentTplList(params, AgentHeaders)
        
        assert result[0] == 200, f"获取已发布模板列表失败，状态码: {result[0]}, 响应: {result[1]}"

    @allure.story("已发布智能体信息列表")
    def test_get_published_agent_info_list(self, factory_client, test_data, AgentHeaders):
        """测试已发布智能体信息列表接口"""

        # 步骤1: 首先获取已发布智能体列表
        published_list_request = json.loads(json.dumps(test_data["published_agent_list_request"]))  # 深拷贝
        published_list_result = factory_client.GetPublishedAgentList(published_list_request, AgentHeaders)

        if published_list_result[0] != 200:
            pytest.skip(f"获取已发布智能体列表失败，跳过测试: {published_list_result[1]}")

        # 步骤2: 从已发布智能体列表中提取agent_keys
        published_agents = published_list_result[1].get("entries", [])
        if not published_agents:
            pytest.skip("没有已发布的智能体，跳过测试")

        # 获取最多5个已发布智能体的agent_key
        agent_keys = []
        for agent in published_agents[:5]:  # 限制数量避免请求过大
            agent_key = agent.get("key")
            if agent_key:
                agent_keys.append(agent_key)

        if not agent_keys:
            pytest.skip("已发布智能体中没有找到agent_key，跳过测试")

        allure.attach(f"找到 {len(agent_keys)} 个已发布智能体", name="已发布智能体数量")
        allure.attach(f"Agent Keys: {agent_keys}", name="使用的Agent Keys")

        # 步骤3: 构建获取已发布智能体信息列表的请求
        request_body = json.loads(json.dumps(test_data["published_agent_info_list_request"]))  # 深拷贝
        request_body["agent_keys"] = agent_keys

        # 步骤4: 调用GetPublishedAgentInfoList接口
        result = factory_client.GetPublishedAgentInfoList(request_body, AgentHeaders)

        # 步骤5: 验证结果
        assert result[0] == 200, f"获取已发布智能体信息列表失败，状态码: {result[0]}, 响应: {result[1]}"

        # 验证返回的数据结构和数量
        if "entries" in result[1]:
            returned_entries = result[1]["entries"]
            assert isinstance(returned_entries, list), "entries应该是列表类型"
            assert len(returned_entries) == len(agent_keys), f"返回的智能体数量({len(returned_entries)})与请求的agent_keys数量({len(agent_keys)})不匹配"

            # 验证每个条目包含必要字段
            for i, entry in enumerate(returned_entries):
                assert "key" in entry, f"第{i+1}个智能体条目缺少agent_key字段"
                assert entry["key"] in agent_keys, f"第{i+1}个智能体条目的agent_key不在请求列表中"

    @allure.story("获取发布历史记录")
    def test_get_release_history(self, factory_client, created_agent, AgentHeaders):
        """测试获取发布历史记录接口"""
        if not created_agent:
            pytest.skip("智能体创建失败，跳过测试")
        
        agent_id = created_agent["id"]
        result = factory_client.GetReleaseHistory(agent_id, AgentHeaders)
        
        assert result[0] == 200, f"获取发布历史记录失败，状态码: {result[0]}, 响应: {result[1]}"
        assert "entries" in result[1], "响应中缺少 entries 字段"

    # ========== 权限相关接口测试 ==========

    @allure.story("检查agent执行权限")
    def test_check_agent_permission(self, factory_client, created_agent, test_data, AgentHeaders):
        """测试检查agent执行权限接口"""
        if not created_agent:
            pytest.skip("智能体创建失败，跳过测试")
        
        agent_id = created_agent["id"]
        request_body = json.loads(json.dumps(test_data["permission_check_request"]))  # 深拷贝
        request_body["agent_id"] = request_body["agent_id"].replace("{agent_id}", agent_id)
        result = factory_client.CheckAgentPermission(request_body, AgentHeaders)
        
        assert result[0] == 200, f"检查agent执行权限失败，状态码: {result[0]}, 响应: {result[1]}"
        assert "is_allowed" in result[1], "响应中缺少 is_allowed 字段"

    @allure.story("获取用户管理权限状态")
    def test_get_permission_user_status(self, factory_client, AgentHeaders):
        """测试获取用户管理权限状态接口"""
        result = factory_client.GetPermissionUserStatus(AgentHeaders)
        
        assert result[0] == 200, f"获取用户管理权限状态失败，状态码: {result[0]}, 响应: {result[1]}"

    # ========== 产品相关接口测试 ==========

    @allure.story("获取产品列表")
    def test_get_product_list(self, factory_client, test_data, AgentHeaders):
        """测试获取产品列表接口"""
        params = json.loads(json.dumps(test_data["product_list_params"]))  # 深拷贝
        result = factory_client.GetProductList(params, AgentHeaders)
        
        assert result[0] == 200, f"获取产品列表失败，状态码: {result[0]}, 响应: {result[1]}"
        assert "entries" in result[1], "响应中缺少 entries 字段"

    @allure.story("创建产品")
    def test_create_product(self, factory_client, test_data, AgentHeaders):
        """测试创建产品接口"""
        product_config = json.loads(json.dumps(test_data["product_config"]))  # 深拷贝
        product_config["name"] = product_config["name"].replace("{timestamp}", str(int(time.time())))
        result = factory_client.CreateProduct(product_config, AgentHeaders)
        
        assert result[0] == 201, f"创建产品失败，状态码: {result[0]}, 响应: {result[1]}"
        assert "id" in result[1], "响应中缺少 id 字段"
        
        # 清理
        product_id = result[1].get("id")
        factory_client.DeleteProduct(product_id, AgentHeaders)

    @allure.story("获取产品详情")
    def test_get_product_detail(self, factory_client, test_data, AgentHeaders):
        """测试获取产品详情接口"""
        # 先创建一个产品
        product_config = json.loads(json.dumps(test_data["product_config"]))  # 深拷贝
        product_config["name"] = product_config["name"].replace("{timestamp}", str(int(time.time())))
        create_result = factory_client.CreateProduct(product_config, AgentHeaders)
        
        if create_result[0] == 201:
            product_id = create_result[1].get("id")
            result = factory_client.GetProductDetail(product_id, AgentHeaders)
            
            assert result[0] == 200, f"获取产品详情失败，状态码: {result[0]}, 响应: {result[1]}"
            assert result[1].get("id") == product_id, "返回的产品ID不匹配"
            
            # 清理
            factory_client.DeleteProduct(product_id, AgentHeaders)

    # ========== 个人空间相关接口测试 ==========

    @allure.story("个人空间智能体列表")
    def test_get_personal_space_agent_list(self, factory_client, test_data, AgentHeaders):
        """测试个人空间智能体列表接口"""
        params = json.loads(json.dumps(test_data["personal_space_params"]))  # 深拷贝
        result = factory_client.GetPersonalSpaceAgentList(params, AgentHeaders)
        
        assert result[0] == 200, f"获取个人空间智能体列表失败，状态码: {result[0]}, 响应: {result[1]}"

    @allure.story("个人空间智能体模板列表")
    def test_get_personal_space_agent_tpl_list(self, factory_client, test_data, AgentHeaders):
        """测试个人空间智能体模板列表接口"""
        params = json.loads(json.dumps(test_data["personal_space_params"]))  # 深拷贝
        result = factory_client.GetPersonalSpaceAgentTplList(params, AgentHeaders)
        
        assert result[0] == 200, f"获取个人空间智能体模板列表失败，状态码: {result[0]}, 响应: {result[1]}"

    @allure.story("最近访问智能体列表")
    def test_get_recent_visit_agent_list(self, factory_client, test_data, AgentHeaders):
        """测试最近访问智能体列表接口"""
        params = json.loads(json.dumps(test_data["personal_space_params"]))  # 深拷贝
        result = factory_client.GetRecentVisitAgentList(params, AgentHeaders)
        
        assert result[0] == 200, f"获取最近访问智能体列表失败，状态码: {result[0]}, 响应: {result[1]}"

    # ========== 其他接口测试 ==========

    @allure.story("获取内置头像列表")
    def test_get_built_in_avatar_list(self, factory_client, AgentHeaders):
        """测试获取内置头像列表接口"""
        result = factory_client.GetBuiltInAvatarList(AgentHeaders)
        
        assert result[0] == 200, f"获取内置头像列表失败，状态码: {result[0]}, 响应: {result[1]}"
        assert isinstance(result[1]["entries"], list), "响应应该是列表类型"

    @allure.story("获取文件扩展名映射")
    def test_get_file_ext_map(self, factory_client, AgentHeaders):
        """测试获取文件扩展名映射接口"""
        result = factory_client.GetFileExtMap(AgentHeaders)
        
        assert result[0] == 200, f"获取文件扩展名映射失败，状态码: {result[0]}, 响应: {result[1]}"
        assert isinstance(result[1], dict), "响应应该是字典类型"

    @allure.story("批量获取数据处理状态")
    def test_batch_check_index_status(self, factory_client, created_agent, test_data, AgentHeaders):
        """测试批量获取数据处理状态接口"""
        if not created_agent:
            pytest.skip("智能体创建失败，跳过测试")
        
        agent_id = created_agent["id"]
        batch_request = json.loads(json.dumps(test_data["batch_index_status_request"]))  # 深拷贝
        batch_request["agent_uniq_flags"][0]["agent_id"] = batch_request["agent_uniq_flags"][0]["agent_id"].replace("{agent_id}", agent_id)
        result = factory_client.BatchCheckIndexStatus(
            batch_request["agent_uniq_flags"], 
            batch_request["is_show_fail_infos"], 
            AgentHeaders
        )
        
        assert result[0] == 200, f"批量获取数据处理状态失败，状态码: {result[0]}, 响应: {result[1]}"
        assert "entries" in result[1], "响应中缺少 entries 字段"

