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
    1. 清理数据流
    2. 清理组合算子（先下架再删除）
    """
    print("\n========== 开始清理 dataflow 测试数据 ==========")

    # 从test_run.py获取记录的数据ID
    dag_ids = set()
    operator_ids = set()

    try:
        from test_run import resp_values
        # 提取dag_id相关
        for key, val in resp_values.items():
            if "dag_id" in key.lower() and val:
                dag_ids.add(str(val))
                print(f"[清理] 发现数据流ID: {key} = {val}")
            if "operator_id" in key.lower() and val:
                operator_ids.add(str(val))
                print(f"[清理] 发现算子ID: {key} = {val}")
    except ImportError:
        print("[清理] 无法导入test_run模块")

    # 同时通过API查询本次创建的数据
    server_config = config.get("server", {})
    host = server_config.get("host", "")
    port = server_config.get("public_port", "443")
    base_url = f"https://{host}:{port}"

    headers = {
        "Content-Type": "application/json",
        "Accept": "application/json",
        "x-business-domain": "bd_public"
    }

    # 查询算子列表，找出本次测试创建的
    print("\n[清理] 查询算子列表...")
    try:
        url = f"{base_url}/api/agent-operator-integration/v1/operator/info/list?all=true"
        resp = requests.get(url, headers=headers, verify=False, timeout=30)
        if resp.status_code == 200:
            operators = resp.json().get("data", [])
            patterns = ["组合算子_sync_nodes_", "组合算子_combo_nodes_", "组合算子_python_nodes_", "组合算子_loop_nodes_"]
            for op in operators:
                name = op.get("name", "") or op.get("operator_info", {}).get("name", "")
                op_id = op.get("operator_id", "")
                status = op.get("status", "")
                for p in patterns:
                    if p in name:
                        operator_ids.add(op_id)
                        # 算子创建时返回的id就是dag_id，需要一并删除
                        extend_info = op.get("extend_info", {})
                        if extend_info and "dag_id" in extend_info:
                            dag_ids.add(str(extend_info["dag_id"]))
                        print(f"[清理] 发现算子: {name} (ID: {op_id})")
                        break
    except Exception as e:
        print(f"[清理] 查询算子列表失败: {e}")

    print(f"\n[清理] 待删除数据流: {len(dag_ids)} 个")
    print(f"[清理] 待删除算子: {len(operator_ids)} 个")

    if not dag_ids and not operator_ids:
        print("[清理] 无需清理的数据")
        return

    # 清理数据流
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

    # 清理组合算子（先下架再删除）
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

