# 全局变量，用于存储初始存储列表和 session_setup 创建的存储 ID
_initial_storage_ids = None
_session_setup_storage_ids = []

import requests
import json
import os


def _get_storage_api_base_url(config):
    """获取存储 API 基础 URL"""
    oss = config.get("oss_info") or {}
    base = (oss.get("storage_api_base_url") or "").strip()
    if not base:
        server_host = config.get('server', {}).get('host', '')
        if server_host:
            base = f"http://{server_host}:8081"
    return base.rstrip("/") if base else None


def session_setup(session_id, config):
    """
    会话开始前执行，记录初始存储列表
    :param session_id: 会话 ID
    :param config: 测试配置
    """
    global _initial_storage_ids
    
    print(f"[oss] Session setup: {session_id}")
    
    if os.environ.get('AT_CLEAN_UP') != '1':
        print('清理功能未启用，跳过记录初始存储')
        return
    
    base_url = _get_storage_api_base_url(config)
    if not base_url:
        print('未配置 storage_api_base_url')
        return
    
    print(f'Base URL: {base_url}')
    print('Recording initial storages...')
    
    headers = {
        "Content-Type": "application/json", 
        "Accept": "application/json"
    }
    
    # 记录存储列表
    try:
        storage_list_url = f"{base_url}/api/v1/storages"
        response = requests.get(storage_list_url, params={"page": 1, "size": 1000}, headers=headers, verify=False)
        if response.status_code == 200:
            data = response.json()
            storage_ids = [item['storage_id'] for item in data.get('data', [])]
            _initial_storage_ids = storage_ids
            os.environ['AT_OSS_INITIAL_STORAGE_IDS'] = json.dumps(storage_ids)
            print(f'Recorded {len(storage_ids)} initial storages')
        else:
            print(f'Failed to get storage list: {response.status_code}, {response.text}')
    except Exception as e:
        print(f'Exception getting storage list: {e}')


def session_clean_up(config, allure):
    """
    会话清理函数，在测试会话结束时执行
    :param config: 配置对象
    :param allure: allure 对象，用于添加报告信息
    """
    global _initial_storage_ids, _session_setup_storage_ids
    
    print(f"[oss] Session clean up called, AT_CLEAN_UP={os.environ.get('AT_CLEAN_UP')}")
    
    if os.environ.get('AT_CLEAN_UP') != '1':
        print('清理功能未启用')
        allure.attach('清理功能未启用', '请设置环境变量 AT_CLEAN_UP=1 来启用清理功能', allure.attachment_type.TEXT)
        return
    
    base_url = _get_storage_api_base_url(config)
    if not base_url:
        print('未配置 storage_api_base_url')
        allure.attach('清理失败', '未配置 storage_api_base_url', allure.attachment_type.TEXT)
        return
    
    print(f"Base URL: {base_url}")
    
    # 从环境变量读取初始存储 ID
    if _initial_storage_ids is None:
        env_storage_ids = os.environ.get('AT_OSS_INITIAL_STORAGE_IDS')
        if env_storage_ids:
            _initial_storage_ids = json.loads(env_storage_ids)
            print(f'从环境变量读取初始存储 ID: {_initial_storage_ids}')
        else:
            print('环境变量中无初始存储 ID')
            _initial_storage_ids = None
    
    # 从环境变量读取 session_setup 创建的存储 ID
    env_session_setup_storage_ids = os.environ.get('AT_OSS_SESSION_SETUP_STORAGE_IDS')
    if env_session_setup_storage_ids:
        _session_setup_storage_ids = json.loads(env_session_setup_storage_ids)
        print(f'从环境变量读取 session_setup 创建的存储 ID: {_session_setup_storage_ids}')
    else:
        print('环境变量中无 session_setup 创建的存储 ID')
    
    # 删除新增的存储
    print('Deleting new storages')
    _delete_new_storages(base_url, allure)
    
    # 清理环境变量
    os.environ.pop('AT_OSS_INITIAL_STORAGE_IDS', None)
    os.environ.pop('AT_OSS_SESSION_SETUP_STORAGE_IDS', None)
    print('已清理相关环境变量')


def _get_current_storage_list(base_url):
    """获取当前存储列表"""
    try:
        url = f"{base_url}/api/v1/storages"
        headers = {
            "Content-Type": "application/json", 
            "Accept": "application/json"
        }
        response = requests.get(url, params={"page": 1, "size": 1000}, headers=headers, verify=False)
        if response.status_code == 200:
            data = response.json()
            return data.get('data', [])
        else:
            print(f'获取存储列表失败：{response.status_code}')
            return []
    except Exception as e:
        print(f'获取存储列表异常：{e}')
        return []


def _delete_storage(base_url, allure, storage_id):
    """删除单个存储"""
    try:
        url = f"{base_url}/api/v1/storages/{storage_id}"
        headers = {
            "Content-Type": "application/json", 
            "Accept": "application/json"
        }
        response = requests.delete(url, headers=headers, verify=False)
        if 200 <= response.status_code < 300:
            print(f'Successfully deleted storage: {storage_id}')
            allure.attach(f'删除新增存储成功', f'存储 ID: {storage_id}', allure.attachment_type.TEXT)
        else:
            print(f'Failed to delete storage {storage_id}: {response.status_code}, {response.text}')
            allure.attach(f'删除新增存储失败', f'存储 ID: {storage_id}, 状态码：{response.status_code}, 响应：{response.text}', allure.attachment_type.TEXT)
    except Exception as e:
        print(f'Exception deleting storage {storage_id}: {e}')
        allure.attach(f'删除新增存储异常', f'存储 ID: {storage_id}, 异常：{str(e)}', allure.attachment_type.TEXT)


def _delete_new_storages(base_url, allure):
    """
    删除新增的存储
    :param base_url: 基础 URL
    :param allure: allure 对象
    """
    global _initial_storage_ids, _session_setup_storage_ids
    
    print('Deleting new storages...')
    
    # 获取当前存储列表
    current_storage_data = _get_current_storage_list(base_url)
    current_storage_ids = [item['storage_id'] for item in current_storage_data]
    print(f'Current storages: {current_storage_ids}')
    
    # 计算需要删除的存储
    if _initial_storage_ids is not None:
        new_storage_ids = [storage_id for storage_id in current_storage_ids if storage_id not in _initial_storage_ids]
        all_storage_ids_to_delete = list(set(str(storage_id) for storage_id in (new_storage_ids + _session_setup_storage_ids)))
    else:
        print('没有初始存储列表，只删除 session_setup 创建的存储')
        all_storage_ids_to_delete = list(set(str(storage_id) for storage_id in _session_setup_storage_ids))
    
    print(f'Storages to delete: {all_storage_ids_to_delete}')
    
    # 执行删除
    for storage_id in all_storage_ids_to_delete:
        _delete_storage(base_url, allure, storage_id)
