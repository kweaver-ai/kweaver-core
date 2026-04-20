"""
Framework Hooks: agent-backend 模块
会话级别的清理和初始化逻辑

清理策略：
1. 在会话结束时清理所有测试过程中创建的资源
2. 清理顺序：已发布模板 -> 普通模板 -> 已发布Agent -> 普通Agent -> 产品
3. 已发布的资源必须先取消发布再删除
"""

import logging
import re
import requests
from typing import Dict, Any, List
from urllib3 import disable_warnings
from urllib3.exceptions import InsecureRequestWarning

disable_warnings(InsecureRequestWarning)
logger = logging.getLogger(__name__)

# 全局变量，用于记录创建的资源ID
_created_resources = {
    "agents": [],              # 未发布的Agent
    "published_agents": [],    # 已发布的Agent（需要先取消发布）
    "products": [],            # 产品
    "templates": [],           # 未发布的模板
    "published_templates": []  # 已发布的模板（需要先取消发布）
}


def _get_base_url(config: Dict[str, Any]) -> str:
    """获取基础URL"""
    host = config.get("server", {}).get("host", "localhost")
    port = config.get("server", {}).get("port", "443")
    protocol = config.get("requests", {}).get("protocol", "https")
    return f"{protocol}://{host}:{port}"


def _get_auth_header(config: Dict[str, Any]) -> Dict[str, str]:
    """获取请求头（不注入鉴权信息）"""
    return {"Content-Type": "application/json"}
    return {"Content-Type": "application/json"}


def _unpublish_agent(base_url: str, headers: Dict[str, str], agent_id: str) -> bool:
    """取消发布Agent"""
    try:
        unpublish_url = f"{base_url}/api/agent-factory/v3/agent/{agent_id}/unpublish"
        resp = requests.put(unpublish_url, headers=headers, verify=False, timeout=10)
        if resp.status_code in [200, 204, 404]:
            logger.info(f"[agent-backend] 已取消发布Agent: {agent_id}")
            return True
        else:
            logger.warning(f"[agent-backend] 取消发布Agent失败: {agent_id}, status: {resp.status_code}")
            return False
    except Exception as e:
        logger.error(f"[agent-backend] 取消发布Agent异常: {agent_id}, error: {e}")
        return False


def _delete_agent(base_url: str, headers: Dict[str, str], agent_id: str) -> bool:
    """删除单个Agent（仅删除，不处理取消发布）"""
    try:
        delete_url = f"{base_url}/api/agent-factory/v3/agent/{agent_id}"
        resp = requests.delete(delete_url, headers=headers, verify=False, timeout=10)
        if resp.status_code in [200, 204, 404]:
            logger.info(f"[agent-backend] 已删除Agent: {agent_id}")
            return True
        else:
            logger.warning(f"[agent-backend] 删除Agent失败: {agent_id}, status: {resp.status_code}")
            return False
    except Exception as e:
        logger.error(f"[agent-backend] 删除Agent异常: {agent_id}, error: {e}")
        return False


def _unpublish_and_delete_agent(base_url: str, headers: Dict[str, str], agent_id: str) -> bool:
    """取消发布并删除Agent（用于已发布的Agent）"""
    logger.info(f"[agent-backend] 清理已发布Agent: {agent_id}")
    # 1. 先取消发布
    _unpublish_agent(base_url, headers, agent_id)
    # 2. 再删除
    return _delete_agent(base_url, headers, agent_id)


def _unpublish_template(base_url: str, headers: Dict[str, str], tpl_id: str) -> bool:
    """取消发布模板"""
    try:
        unpublish_url = f"{base_url}/api/agent-factory/v3/agent-tpl/{tpl_id}/unpublish"
        resp = requests.put(unpublish_url, headers=headers, verify=False, timeout=10)
        if resp.status_code in [200, 204, 404]:
            logger.info(f"[agent-backend] 已取消发布模板: {tpl_id}")
            return True
        else:
            logger.warning(f"[agent-backend] 取消发布模板失败: {tpl_id}, status: {resp.status_code}")
            return False
    except Exception as e:
        logger.error(f"[agent-backend] 取消发布模板异常: {tpl_id}, error: {e}")
        return False


def _delete_template(base_url: str, headers: Dict[str, str], tpl_id: str) -> bool:
    """删除单个模板（仅删除，不处理取消发布）"""
    try:
        delete_url = f"{base_url}/api/agent-factory/v3/agent-tpl/{tpl_id}"
        resp = requests.delete(delete_url, headers=headers, verify=False, timeout=10)
        if resp.status_code in [200, 204, 404]:
            logger.info(f"[agent-backend] 已删除模板: {tpl_id}")
            return True
        else:
            logger.warning(f"[agent-backend] 删除模板失败: {tpl_id}, status: {resp.status_code}")
            return False
    except Exception as e:
        logger.error(f"[agent-backend] 删除模板异常: {tpl_id}, error: {e}")
        return False


def _unpublish_and_delete_template(base_url: str, headers: Dict[str, str], tpl_id: str) -> bool:
    """取消发布并删除模板（用于已发布的模板）"""
    logger.info(f"[agent-backend] 清理已发布模板: {tpl_id}")
    # 1. 先取消发布
    _unpublish_template(base_url, headers, tpl_id)
    # 2. 再删除
    return _delete_template(base_url, headers, tpl_id)


def _delete_product(base_url: str, headers: Dict[str, str], product_id: str) -> bool:
    """删除单个产品"""
    try:
        delete_url = f"{base_url}/api/agent-factory/v3/product/{product_id}"
        resp = requests.delete(delete_url, headers=headers, verify=False, timeout=10)
        if resp.status_code in [200, 204, 404]:
            logger.info(f"[agent-backend] 已删除产品: {product_id}")
            return True
        else:
            logger.warning(f"[agent-backend] 删除产品失败: {product_id}, status: {resp.status_code}")
            return False
    except Exception as e:
        logger.error(f"[agent-backend] 删除产品异常: {product_id}, error: {e}")
        return False


def session_setup(session_id: str, config: Dict[str, Any]) -> None:
    """
    会话开始前执行
    :param session_id: 会话 ID
    :param config: 测试配置
    """
    logger.info(f"[agent-backend] Session setup: {session_id}")
    # 清空资源记录
    global _created_resources
    _created_resources = {
        "agents": [],
        "published_agents": [],
        "products": [],
        "templates": [],
        "published_templates": []
    }


def _should_cleanup_agent(agent: Dict[str, Any]) -> bool:
    """判断Agent是否需要清理（名称匹配"智能体_"后跟六位数字）"""
    name = agent.get("name", "")
    # 匹配"智能体_"后跟六位数字（如：测试智能体_abc123, 智能体_123456）
    return bool(re.search(r'智能体_\d{6}', name))


def _should_cleanup_product(product: Dict[str, Any]) -> bool:
    """判断产品是否需要清理（名称包含"产品_"）"""
    name = product.get("name", "")
    return "产品_" in name


def session_clean_up(config: Dict[str, Any], allure_module) -> None:
    """
    会话结束后执行清理

    清理策略：
    1. 获取所有Agent，过滤名称包含"测试智能体_"的进行清理
    2. 已发布的Agent先取消发布再删除
    3. 获取所有模板，过滤名称包含"测试智能体_"或"模板测试"的进行清理
    4. 获取所有产品，过滤名称包含"产品"的进行清理

    :param config: 测试配置
    :param allure_module: allure模块
    """
    logger.info("[agent-backend] Session cleanup started")

    base_url = _get_base_url(config)
    headers = _get_auth_header(config)
    headers["x-business-domain"] = "bd_public"

    # ========== 1. 清理Agent ==========
    agent_deleted = 0
    list_url = f"{base_url}/api/agent-factory/v3/personal-space/agent-list"

    # 1.1 先清理已发布的Agent（需要先取消发布再删除）
    try:
        max_iterations = 10
        for iteration in range(max_iterations):
            resp = requests.get(list_url, headers=headers,
                              params={"name": "智能体_", "size": 100, "publish_status": "published"},
                              verify=False, timeout=30)
            if resp.status_code != 200:
                logger.warning(f"[agent-backend] Failed to get published agents: {resp.status_code}")
                break

            agents = resp.json().get("entries", [])
            # 过滤需要清理的Agent（名称匹配"智能体_六位数字"）
            agents_to_cleanup = [a for a in agents if _should_cleanup_agent(a)]
            if not agents_to_cleanup:
                break

            for agent in agents_to_cleanup:
                agent_id = agent.get("id", "")
                _unpublish_agent(base_url, headers, agent_id)
                if _delete_agent(base_url, headers, agent_id):
                    agent_deleted += 1
    except Exception as e:
        logger.error(f"[agent-backend] 清理已发布Agent异常: {e}")

    # 1.2 清理未发布的Agent
    try:
        max_iterations = 20
        for iteration in range(max_iterations):
            resp = requests.get(list_url, headers=headers,
                              params={"name": "智能体_", "size": 100, "publish_status": "unpublished"},
                              verify=False, timeout=30)
            if resp.status_code != 200:
                logger.warning(f"[agent-backend] Failed to get unpublished agents: {resp.status_code}")
                break

            agents = resp.json().get("entries", [])
            # 过滤需要清理的Agent（名称匹配"智能体_六位数字"）
            agents_to_cleanup = [a for a in agents if _should_cleanup_agent(a)]
            if not agents_to_cleanup:
                break

            for agent in agents_to_cleanup:
                agent_id = agent.get("id", "")
                if _delete_agent(base_url, headers, agent_id):
                    agent_deleted += 1
    except Exception as e:
        logger.error(f"[agent-backend] 清理未发布Agent异常: {e}")

    logger.info(f"[agent-backend] 共删除 {agent_deleted} 个测试Agent")

    # ========== 2. 清理模板 ==========
    tpl_deleted = 0
    tpl_list_url = f"{base_url}/api/agent-factory/v3/personal-space/agent-tpl-list"

    # 2.1 清理已发布的模板
    try:
        max_iterations = 5
        for iteration in range(max_iterations):
            resp = requests.get(tpl_list_url, headers=headers,
                              params={"name": "测试智能体", "size": 100, "publish_status": "published"},
                              verify=False, timeout=30)
            if resp.status_code != 200:
                logger.warning(f"[agent-backend] Failed to get published templates: {resp.status_code}")
                break

            templates = resp.json().get("entries", [])
            if not templates:
                break

            for tpl in templates:
                tpl_id = tpl.get("id", "")
                _unpublish_template(base_url, headers, str(tpl_id))
                if _delete_template(base_url, headers, str(tpl_id)):
                    tpl_deleted += 1
    except Exception as e:
        logger.error(f"[agent-backend] 清理已发布模板异常: {e}")

    # 2.2 清理未发布的模板
    try:
        max_iterations = 5
        for iteration in range(max_iterations):
            resp = requests.get(tpl_list_url, headers=headers,
                              params={"name": "测试智能体", "size": 100, "publish_status": "unpublished"},
                              verify=False, timeout=30)
            if resp.status_code != 200:
                logger.warning(f"[agent-backend] Failed to get unpublished templates: {resp.status_code}")
                break

            templates = resp.json().get("entries", [])
            if not templates:
                break

            for tpl in templates:
                tpl_id = tpl.get("id", "")
                if _delete_template(base_url, headers, str(tpl_id)):
                    tpl_deleted += 1
    except Exception as e:
        logger.error(f"[agent-backend] 清理未发布模板异常: {e}")

    logger.info(f"[agent-backend] 共删除 {tpl_deleted} 个测试模板")

    # ========== 3. 清理产品 ==========
    product_deleted = 0
    product_list_url = f"{base_url}/api/agent-factory/v3/product"

    try:
        max_iterations = 10
        for iteration in range(max_iterations):
            resp = requests.get(product_list_url, headers=headers,
                              params={"page_size": 100},
                              verify=False, timeout=30)
            if resp.status_code != 200:
                logger.warning(f"[agent-backend] Failed to get products: {resp.status_code}")
                break

            products = resp.json().get("entries", [])
            # 过滤需要清理的产品（名称包含"产品_"）
            products_to_cleanup = [p for p in products if _should_cleanup_product(p)]
            if not products_to_cleanup:
                break

            for product in products_to_cleanup:
                product_id = product.get("id", "")
                if _delete_product(base_url, headers, str(product_id)):
                    product_deleted += 1
    except Exception as e:
        logger.error(f"[agent-backend] 清理产品异常: {e}")

    logger.info(f"[agent-backend] 共删除 {product_deleted} 个测试产品")
    logger.info(f"[agent-backend] Cleanup completed - Agent: {agent_deleted}, Template: {tpl_deleted}, Product: {product_deleted}")


def test_setup(test_id: str, config: Dict[str, Any]) -> None:
    """
    单个测试用例开始前执行
    :param test_id: 测试用例 ID
    :param config: 测试配置
    """
    logger.debug(f"[agent-backend] Test setup: {test_id}")


def test_teardown(test_id: str, config: Dict[str, Any]) -> None:
    """
    单个测试用例结束后执行
    :param test_id: 测试用例 ID
    :param config: 测试配置
    """
    logger.debug(f"[agent-backend] Test teardown: {test_id}")


# ==================== 资源注册函数 ====================
# 这些函数可以在用例执行过程中调用来记录创建的资源

def register_agent(agent_id: str, published: bool = False) -> None:
    """
    注册创建的Agent

    :param agent_id: Agent ID
    :param published: 是否已发布（已发布的Agent需要先取消发布再删除）
    """
    global _created_resources
    if published:
        if agent_id not in _created_resources["published_agents"]:
            _created_resources["published_agents"].append(agent_id)
            logger.info(f"[agent-backend] Registered published agent: {agent_id}")
    else:
        if agent_id not in _created_resources["agents"]:
            _created_resources["agents"].append(agent_id)
            logger.info(f"[agent-backend] Registered agent: {agent_id}")


def register_product(product_id: str) -> None:
    """注册创建的产品"""
    global _created_resources
    if product_id not in _created_resources["products"]:
        _created_resources["products"].append(product_id)
        logger.info(f"[agent-backend] Registered product: {product_id}")


def register_template(tpl_id: str, published: bool = False) -> None:
    """
    注册创建的模板

    :param tpl_id: 模板 ID
    :param published: 是否已发布（已发布的模板需要先取消发布再删除）
    """
    global _created_resources
    if published:
        if tpl_id not in _created_resources["published_templates"]:
            _created_resources["published_templates"].append(tpl_id)
            logger.info(f"[agent-backend] Registered published template: {tpl_id}")
    else:
        if tpl_id not in _created_resources["templates"]:
            _created_resources["templates"].append(tpl_id)
            logger.info(f"[agent-backend] Registered template: {tpl_id}")


def mark_agent_published(agent_id: str) -> None:
    """
    将Agent标记为已发布状态（用于Agent从未发布变为已发布的情况）

    :param agent_id: Agent ID
    """
    global _created_resources
    # 从普通agent列表移除
    if agent_id in _created_resources["agents"]:
        _created_resources["agents"].remove(agent_id)
    # 添加到已发布agent列表
    if agent_id not in _created_resources["published_agents"]:
        _created_resources["published_agents"].append(agent_id)
        logger.info(f"[agent-backend] Marked agent as published: {agent_id}")


def mark_template_published(tpl_id: str) -> None:
    """
    将模板标记为已发布状态

    :param tpl_id: 模板 ID
    """
    global _created_resources
    # 从普通模板列表移除
    if tpl_id in _created_resources["templates"]:
        _created_resources["templates"].remove(tpl_id)
    # 添加到已发布模板列表
    if tpl_id not in _created_resources["published_templates"]:
        _created_resources["published_templates"].append(tpl_id)
        logger.info(f"[agent-backend] Marked template as published: {tpl_id}")