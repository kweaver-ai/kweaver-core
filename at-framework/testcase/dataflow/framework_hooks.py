"""Framework Hooks: dataflow"""

import requests
import urllib3
import time

urllib3.disable_warnings(urllib3.connectionpool.InsecureRequestWarning)


def session_setup(session_id, config):
    """会话开始前的设置"""
    pass


def session_clean_up(config, allure):
    """
    会话结束后清理本次测试创建的数据
    1. 使用 /dags?type=data-flow API 获取所有数据流列表
    2. 根据名称模式匹配删除数据流
    3. 从算子列表获取组合算子并删除
    """
    print("\n========== 开始清理 dataflow 测试数据 ==========")

    server_config = config.get("server", {})
    host = server_config.get("host", "")
    port = server_config.get("public_port", "443")
    base_url = f"https://{host}:{port}"

    headers = {
        "Content-Type": "application/json",
        "Accept": "application/json",
        "x-business-domain": "bd_public"
    }

    # 收集需要删除的ID
    dag_ids = set()
    operator_ids = set()

    # 数据流名称模式（测试创建的）
    dag_patterns = ["upload_", "版本管理测试", "版本创建测试", "更新测试", "完整流程测试", "完整流程更新"]
    # 组合算子名称模式
    operator_patterns = ["组合算子_sync_nodes_", "组合算子_combo_nodes_", "组合算子_python_nodes_", "组合算子_loop_nodes_"]

    # 1. 使用正确的API获取数据流列表
    print("[清理] 查询数据流列表 (type=data-flow)...")
    try:
        page = 0
        limit = 100
        while True:
            url = f"{base_url}/api/automation/v1/dags?type=data-flow&limit={limit}&page={page}"
            resp = requests.get(url, headers=headers, verify=False, timeout=30)
            if resp.status_code != 200:
                print(f"[清理] 查询数据流失败: {resp.status_code}")
                break
            data = resp.json()
            dags = data.get("dags", [])
            if not dags:
                break
            for dag in dags:
                dag_id = dag.get("id", "")
                title = dag.get("title", "")
                for p in dag_patterns:
                    if p in title:
                        dag_ids.add(str(dag_id))
                        print(f"[清理] 发现数据流: {title} (ID: {dag_id})")
                        break
            page += 1
            total = data.get("total", 0)
            if len(dag_ids) >= total or page * limit >= total:
                break
    except Exception as e:
        print(f"[清理] 查询数据流列表异常: {e}")

    # 2. 查询算子列表，找出组合算子
    print("\n[清理] 查询算子列表...")
    try:
        url = f"{base_url}/api/agent-operator-integration/v1/operator/info/list?all=true"
        resp = requests.get(url, headers=headers, verify=False, timeout=30)
        if resp.status_code == 200:
            operators = resp.json().get("data", [])
            for op in operators:
                name = op.get("name", "") or op.get("operator_info", {}).get("name", "")
                op_id = op.get("operator_id", "")
                for p in operator_patterns:
                    if p in name:
                        operator_ids.add(op_id)
                        print(f"[清理] 发现组合算子: {name} (ID: {op_id})")
                        break
    except Exception as e:
        print(f"[清理] 查询算子列表失败: {e}")

    # 3. 从 resp_values 获取额外记录
    try:
        from test_run import resp_values
        for key, val in resp_values.items():
            if "dag_id" in key.lower() and val:
                dag_ids.add(str(val))
                print(f"[清理] 从resp_values添加数据流: {key} = {val}")
            if "operator_id" in key.lower() and val:
                operator_ids.add(str(val))
                print(f"[清理] 从resp_values添加算子: {key} = {val}")
    except ImportError:
        print("[清理] 无法导入test_run模块")

    print(f"\n[清理] 待删除数据流: {len(dag_ids)} 个")
    print(f"[清理] 待删除算子: {len(operator_ids)} 个")

    if not dag_ids and not operator_ids:
        print("[清理] 无需清理的数据")
        return

    # 4. 清理数据流
    print("\n[清理] 删除数据流...")
    for dag_id in dag_ids:
        url = f"{base_url}/api/automation/v1/data-flow/flow/{dag_id}"
        try:
            resp = requests.delete(url, headers=headers, verify=False, timeout=30)
            if resp.status_code in [200, 204]:
                print(f"[清理] ✓ 删除数据流: {dag_id}")
            else:
                print(f"[清理] ✗ 删除数据流失败: {dag_id} - {resp.status_code}")
        except Exception as e:
            print(f"[清理] ✗ 删除数据流异常: {dag_id} - {e}")
        time.sleep(0.2)

    # 5. 清理组合算子（先下架再删除）
    print("\n[清理] 删除算子...")
    for op_id in operator_ids:
        # 下架
        offline_url = f"{base_url}/api/agent-operator-integration/v1/operator/status"
        try:
            requests.post(offline_url, json=[{"operator_id": op_id, "status": "offline"}],
                         headers=headers, verify=False, timeout=30)
        except:
            pass
        time.sleep(0.1)

        # 删除
        del_url = f"{base_url}/api/agent-operator-integration/v1/operator/delete"
        try:
            resp = requests.delete(del_url, json=[{"operator_id": op_id}],
                                  headers=headers, verify=False, timeout=30)
            if resp.status_code == 200:
                print(f"[清理] ✓ 删除算子: {op_id}")
            else:
                print(f"[清理] ✗ 删除算子失败: {op_id} - {resp.status_code}")
        except Exception as e:
            print(f"[清理] ✗ 删除算子异常: {op_id} - {e}")
        time.sleep(0.2)

    print("\n========== dataflow 测试数据清理完成 ==========\n")


def test_setup(test_id, config):
    pass


def test_teardown(test_id, config):
    pass

