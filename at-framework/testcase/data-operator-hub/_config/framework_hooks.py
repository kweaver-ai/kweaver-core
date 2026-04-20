"""Framework Hooks: data-operator-hub"""

import requests
import urllib3
import time

urllib3.disable_warnings(urllib3.connectionpool.InsecureRequestWarning)


def session_setup(session_id, config):
    """会话开始前的设置"""
    pass


def session_clean_up(session_id, config):
    """
    会话结束后的清理
    1. 获取所有工具箱列表
    2. 对每个工具箱：先下架（如果已发布），再删除
    3. 同样处理MCP服务
    """
    print("\n========== 开始清理测试数据 ==========\n")

    # 获取配置信息
    server_config = config.get("server", {})
    host = server_config.get("host", "")
    port = server_config.get("public_port", "443")
    protocol = "https"
    base_url = f"{protocol}://{host}:{port}"

    headers = {
        "Content-Type": "application/json",
        "Accept": "application/json"
    }

    # 清理工具箱
    _cleanup_toolboxes(base_url, headers)

    # 清理MCP服务
    _cleanup_mcp_servers(base_url, headers)

    # 清理算子（可选）
    _cleanup_operators(base_url, headers)

    print("\n========== 测试数据清理完成 ==========\n")


def test_setup(test_id, config):
    pass


def test_teardown(test_id, config):
    pass


def _cleanup_toolboxes(base_url, headers):
    """清理工具箱数据"""
    print("--- 清理工具箱 ---")

    # 获取所有工具箱
    list_url = f"{base_url}/api/agent-operator-integration/v1/tool-box/list?all=true"
    try:
        resp = requests.get(list_url, headers=headers, verify=False, timeout=30)
        if resp.status_code != 200:
            print(f"获取工具箱列表失败: {resp.status_code}")
            return

        data = resp.json()
        toolboxes = data.get("data", [])
        print(f"找到 {len(toolboxes)} 个工具箱")

        for box in toolboxes:
            box_id = box.get("box_id")
            box_name = box.get("box_name", "")
            status = box.get("status", "")

            if not box_id:
                continue

            print(f"处理工具箱: {box_name} (ID: {box_id}, 状态: {status})")

            # 如果已发布，先下架
            if status == "published":
                _offline_toolbox(base_url, headers, box_id, box_name)
                time.sleep(0.2)

            # 删除工具箱
            _delete_toolbox(base_url, headers, box_id, box_name)
            time.sleep(0.2)

    except Exception as e:
        print(f"清理工具箱出错: {e}")


def _offline_toolbox(base_url, headers, box_id, box_name):
    """下架工具箱"""
    url = f"{base_url}/api/agent-operator-integration/v1/tool-box/{box_id}/status"
    data = {"status": "offline"}
    try:
        resp = requests.post(url, json=data, headers=headers, verify=False, timeout=30)
        if resp.status_code == 200:
            print(f"  ✓ 下架成功: {box_name}")
        else:
            print(f"  ✗ 下架失败: {box_name} - {resp.status_code}")
    except Exception as e:
        print(f"  ✗ 下架异常: {box_name} - {e}")


def _delete_toolbox(base_url, headers, box_id, box_name):
    """删除工具箱"""
    url = f"{base_url}/api/agent-operator-integration/v1/tool-box/{box_id}"
    try:
        resp = requests.delete(url, headers=headers, verify=False, timeout=30)
        if resp.status_code == 200:
            print(f"  ✓ 删除成功: {box_name}")
        else:
            print(f"  ✗ 删除失败: {box_name} - {resp.status_code}")
    except Exception as e:
        print(f"  ✗ 删除异常: {box_name} - {e}")


def _cleanup_mcp_servers(base_url, headers):
    """清理MCP服务数据"""
    print("\n--- 清理MCP服务 ---")

    # 获取所有MCP服务
    list_url = f"{base_url}/api/agent-operator-integration/v1/mcp/list?all=true"
    try:
        resp = requests.get(list_url, headers=headers, verify=False, timeout=30)
        if resp.status_code != 200:
            print(f"获取MCP列表失败: {resp.status_code}")
            return

        data = resp.json()
        mcps = data.get("data", [])
        print(f"找到 {len(mcps)} 个MCP服务")

        for mcp in mcps:
            mcp_id = mcp.get("id") or mcp.get("mcp_id")
            mcp_name = mcp.get("name", "")
            status = mcp.get("status", "")

            if not mcp_id:
                continue

            print(f"处理MCP: {mcp_name} (ID: {mcp_id}, 状态: {status})")

            # 如果已发布，先下架
            if status == "published":
                _offline_mcp(base_url, headers, mcp_id, mcp_name)
                time.sleep(0.2)

            # 删除MCP
            _delete_mcp(base_url, headers, mcp_id, mcp_name)
            time.sleep(0.2)

    except Exception as e:
        print(f"清理MCP服务出错: {e}")


def _offline_mcp(base_url, headers, mcp_id, mcp_name):
    """下架MCP服务"""
    url = f"{base_url}/api/agent-operator-integration/v1/mcp/{mcp_id}/status"
    data = {"status": "offline"}
    try:
        resp = requests.post(url, json=data, headers=headers, verify=False, timeout=30)
        if resp.status_code == 200:
            print(f"  ✓ 下架成功: {mcp_name}")
        else:
            print(f"  ✗ 下架失败: {mcp_name} - {resp.status_code}")
    except Exception as e:
        print(f"  ✗ 下架异常: {mcp_name} - {e}")


def _delete_mcp(base_url, headers, mcp_id, mcp_name):
    """删除MCP服务"""
    url = f"{base_url}/api/agent-operator-integration/v1/mcp/{mcp_id}"
    try:
        resp = requests.delete(url, headers=headers, verify=False, timeout=30)
        if resp.status_code == 200:
            print(f"  ✓ 删除成功: {mcp_name}")
        else:
            print(f"  ✗ 删除失败: {mcp_name} - {resp.status_code}")
    except Exception as e:
        print(f"  ✗ 删除异常: {mcp_name} - {e}")


def _cleanup_operators(base_url, headers):
    """清理算子数据"""
    print("\n--- 清理算子 ---")

    # 获取所有算子
    list_url = f"{base_url}/api/agent-operator-integration/v1/operator/info/list?all=true"
    try:
        resp = requests.get(list_url, headers=headers, verify=False, timeout=30)
        if resp.status_code != 200:
            print(f"获取算子列表失败: {resp.status_code}")
            return

        data = resp.json()
        operators = data.get("data", [])
        print(f"找到 {len(operators)} 个算子")

        for op in operators:
            op_id = op.get("operator_id")
            op_name = op.get("name", "") or op.get("operator_info", {}).get("name", "")
            status = op.get("status", "")

            if not op_id:
                continue

            print(f"处理算子: {op_name} (ID: {op_id}, 状态: {status})")

            # 如果已发布，先下架
            if status == "published":
                _offline_operator(base_url, headers, op_id, op_name)
                time.sleep(0.2)

            # 删除算子
            _delete_operator(base_url, headers, op_id, op_name)
            time.sleep(0.2)

    except Exception as e:
        print(f"清理算子出错: {e}")


def _offline_operator(base_url, headers, op_id, op_name):
    """下架算子"""
    url = f"{base_url}/api/agent-operator-integration/v1/operator/status"
    data = {"operator_id": op_id, "status": "offline"}
    try:
        resp = requests.post(url, json=data, headers=headers, verify=False, timeout=30)
        if resp.status_code == 200:
            print(f"  ✓ 下架成功: {op_name}")
        else:
            print(f"  ✗ 下架失败: {op_name} - {resp.status_code}")
    except Exception as e:
        print(f"  ✗ 下架异常: {op_name} - {e}")


def _delete_operator(base_url, headers, op_id, op_name):
    """删除算子"""
    url = f"{base_url}/api/agent-operator-integration/v1/operator/delete"
    data = {"operator_ids": [op_id]}
    try:
        resp = requests.delete(url, json=data, headers=headers, verify=False, timeout=30)
        if resp.status_code == 200:
            print(f"  ✓ 删除成功: {op_name}")
        else:
            print(f"  ✗ 删除失败: {op_name} - {resp.status_code}")
    except Exception as e:
        print(f"  ✗ 删除异常: {op_name} - {e}")