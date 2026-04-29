# 全局变量，用于存储初始模型列表和session_setup创建的模型ID
_initial_llm_ids = None
_initial_small_model_ids = None
_session_setup_llm_ids = []
_session_setup_small_model_ids = []

import requests
import json
import os


def _get_auth_token(cfg):
    """获取认证 token，参考 test_run.py 的 _default_bearer_auth 逻辑"""
    try:
        from common import at_env
    except Exception:
        return None
    
    source = at_env.auth_token_source(cfg)
    if source == "login":
        try:
            from src.common.token_provider import get_token, clear_token_cache

            user, pwd = at_env.admin_credentials(cfg)
            if user:
                clear_token_cache(user)
                tok = at_env.normalize_bearer_token_value(get_token(user, pwd, force_refresh=True) or "")
                if tok:
                    return tok
        except Exception as ex:
            print('get_token 异常: %s' % str(ex))

    static_token = at_env.static_access_token(cfg)
    if static_token:
        return static_token
    
    return None


def session_setup(session_id, config):
    """
    会话开始前执行，记录初始模型列表
    :param session_id: 会话 ID
    :param config: 测试配置
    """
    global _initial_llm_ids, _initial_small_model_ids
    
    print(f"[modelmanager] Session setup: {session_id}")
    
    if os.environ.get('AT_CLEAN_UP') != '1':
        print('清理功能未启用，跳过记录初始模型')
        return
    
    base_url = config.get('server', {}).get('base_url', '')
    if not base_url:
        print('未配置base_url')
        return
    
    token = _get_auth_token(config)
    if not token:
        print('未获取到认证 token')
        return
    
    print(f'Base URL: {base_url}')
    print(f'获取到认证 token: {token[:20]}...')
    print('Recording initial models...')
    
    headers = {
        "Content-Type": "application/json", 
        "Accept": "application/json",
        "Authorization": f"Bearer {token}"
    }
    
    # 记录大模型列表
    try:
        llm_list_url = f"{base_url}/api/mf-model-manager/v1/llm/list"
        response = requests.get(llm_list_url, params={"page":1, "size": 1000}, headers=headers, verify=False)
        if response.status_code == 200:
            data = response.json()
            llm_ids = [item['model_id'] for item in data.get('data', [])]
            _initial_llm_ids = llm_ids
            os.environ['AT_MODEL_MANAGER_INITIAL_LLM_IDS'] = json.dumps(llm_ids)
            print(f'Recorded {len(llm_ids)} initial LLM models')
        else:
            print(f'Failed to get LLM list: {response.status_code}, {response.text}')
    except Exception as e:
        print(f'Exception getting LLM list: {e}')
    
    # 记录小模型列表
    try:
        small_model_list_url = f"{base_url}/api/mf-model-manager/v1/small-model/list"
        response = requests.get(small_model_list_url, params={"page":1, "size": 1000}, headers=headers, verify=False)
        if response.status_code == 200:
            data = response.json()
            small_model_ids = [item.get('model_id') or item.get('id') for item in data.get('data', [])]
            _initial_small_model_ids = small_model_ids
            os.environ['AT_MODEL_MANAGER_INITIAL_SMALL_MODEL_IDS'] = json.dumps(small_model_ids)
            print(f'Recorded {len(small_model_ids)} initial small models')
        else:
            print(f'Failed to get small model list: {response.status_code}, {response.text}')
    except Exception as e:
        print(f'Exception getting small model list: {e}')


def session_clean_up(config, allure):
    """
    会话清理函数，在测试会话结束时执行
    :param config: 配置对象
    :param allure: allure对象，用于添加报告信息
    """
    global _initial_llm_ids, _initial_small_model_ids, _session_setup_llm_ids, _session_setup_small_model_ids
    
    print(f"Session clean up called, AT_CLEAN_UP={os.environ.get('AT_CLEAN_UP')}")
    
    if os.environ.get('AT_CLEAN_UP') != '1':
        print('清理功能未启用')
        allure.attach('清理功能未启用', '请设置环境变量 AT_CLEAN_UP=1 来启用清理功能', allure.attachment_type.TEXT)
        return
    
    base_url = config.get('server', {}).get('base_url', '')
    if not base_url:
        print('未配置base_url')
        allure.attach('清理失败', '未配置base_url', allure.attachment_type.TEXT)
        return
    
    print(f"Base URL: {base_url}")
    
    token = _get_auth_token(config)
    if token:
        print(f'获取到认证 token: {token[:20]}...')
    else:
        print('未获取到认证 token')
        return
    
    # 从环境变量读取初始模型ID
    if _initial_llm_ids is None:
        env_llm_ids = os.environ.get('AT_MODEL_MANAGER_INITIAL_LLM_IDS')
        if env_llm_ids:
            _initial_llm_ids = json.loads(env_llm_ids)
            print(f'从环境变量读取初始大模型ID: {_initial_llm_ids}')
        else:
            print('环境变量中无初始大模型ID')
            _initial_llm_ids = None
    
    if _initial_small_model_ids is None:
        env_small_ids = os.environ.get('AT_MODEL_MANAGER_INITIAL_SMALL_MODEL_IDS')
        if env_small_ids:
            _initial_small_model_ids = json.loads(env_small_ids)
            print(f'从环境变量读取初始小模型ID: {_initial_small_model_ids}')
        else:
            print('环境变量中无初始小模型ID')
            _initial_small_model_ids = None
    
    # 从环境变量读取session_setup创建的模型ID
    env_session_setup_llm_ids = os.environ.get('AT_MODEL_MANAGER_SESSION_SETUP_LLM_IDS')
    if env_session_setup_llm_ids:
        _session_setup_llm_ids = json.loads(env_session_setup_llm_ids)
        print(f'从环境变量读取session_setup创建的大模型ID: {_session_setup_llm_ids}')
    else:
        print('环境变量中无session_setup创建的大模型ID')
    
    env_session_setup_small_model_ids = os.environ.get('AT_MODEL_MANAGER_SESSION_SETUP_SMALL_MODEL_IDS')
    if env_session_setup_small_model_ids:
        _session_setup_small_model_ids = json.loads(env_session_setup_small_model_ids)
        print(f'从环境变量读取session_setup创建的小模型ID: {_session_setup_small_model_ids}')
    else:
        print('环境变量中无session_setup创建的小模型ID')
    
    # 删除新增的模型
    print('Deleting new models')
    _delete_new_models(base_url, token, allure)
    
    # 清理环境变量
    os.environ.pop('AT_MODEL_MANAGER_INITIAL_LLM_IDS', None)
    os.environ.pop('AT_MODEL_MANAGER_INITIAL_SMALL_MODEL_IDS', None)
    os.environ.pop('AT_MODEL_MANAGER_SESSION_SETUP_LLM_IDS', None)
    os.environ.pop('AT_MODEL_MANAGER_SESSION_SETUP_SMALL_MODEL_IDS', None)
    print('已清理相关环境变量')


def _get_current_model_list(base_url, token, url_path):
    """获取当前模型列表的通用方法"""
    try:
        url = f"{base_url}{url_path}"
        headers = {
            "Content-Type": "application/json", 
            "Accept": "application/json",
            "Authorization": f"Bearer {token}"
        }
        response = requests.get(url, params={"page":1, "size": 1000}, headers=headers, verify=False)
        if response.status_code == 200:
            data = response.json()
            return data.get('data', [])
        else:
            print(f'获取模型列表失败: {response.status_code}')
            return []
    except Exception as e:
        print(f'获取模型列表异常: {e}')
        return []


def _delete_models(base_url, token, allure, model_ids, delete_url, model_type):
    """删除模型的通用方法"""
    if not model_ids:
        print(f'No {model_type} models to delete')
        return
    
    print(f'All {model_type} models to delete: {model_ids}')
    
    try:
        headers = {
            "Content-Type": "application/json", 
            "Accept": "application/json",
            "Authorization": f"Bearer {token}"
        }
        data = {"model_ids": model_ids}
        response = requests.post(delete_url, headers=headers, data=json.dumps(data), verify=False)
        if response.status_code == 200:
            print(f'Successfully deleted {model_type} models: {model_ids}')
            allure.attach(f'删除新增{model_type}模型成功', f'模型ID: {model_ids}', allure.attachment_type.TEXT)
        else:
            print(f'Failed to delete {model_type} models: {response.status_code}, {response.text}')
            allure.attach(f'删除新增{model_type}模型失败', f'模型ID: {model_ids}, 状态码: {response.status_code}, 响应: {response.text}', allure.attachment_type.TEXT)
    except Exception as e:
        print(f'Exception deleting {model_type} models: {e}')
        allure.attach(f'删除新增{model_type}模型异常', f'模型ID: {model_ids}, 异常: {str(e)}', allure.attachment_type.TEXT)


def _delete_new_models(base_url, token, allure):
    """
    删除新增的模型
    :param base_url: 基础URL
    :param token: 认证token
    :param allure: allure对象
    """
    global _initial_llm_ids, _initial_small_model_ids, _session_setup_llm_ids, _session_setup_small_model_ids
    
    print('Deleting new models...')
    
    # 获取当前模型列表
    current_llm_data = _get_current_model_list(base_url, token, "/api/mf-model-manager/v1/llm/list")
    current_llm_ids = [item['model_id'] for item in current_llm_data]
    print(f'Current LLM models: {current_llm_ids}')
    
    current_small_data = _get_current_model_list(base_url, token, "/api/mf-model-manager/v1/small-model/list")
    current_small_model_ids = [item.get('model_id') or item.get('id') for item in current_small_data]
    print(f'Current small models: {current_small_model_ids}')
    
    # 计算需要删除的大模型
    if _initial_llm_ids is not None:
        new_llm_ids = [id for id in current_llm_ids if id not in _initial_llm_ids]
        all_llm_ids_to_delete = list(set(str(id) for id in (new_llm_ids + _session_setup_llm_ids)))
    else:
        print('没有初始模型列表，只删除session_setup创建的模型')
        all_llm_ids_to_delete = list(set(str(id) for id in _session_setup_llm_ids))
    
    # 计算需要删除的小模型
    if _initial_small_model_ids is not None:
        new_small_model_ids = [id for id in current_small_model_ids if id not in _initial_small_model_ids]
        all_small_model_ids_to_delete = list(set(str(id) for id in (new_small_model_ids + _session_setup_small_model_ids)))
    else:
        print('没有初始小模型列表，只删除session_setup创建的模型')
        all_small_model_ids_to_delete = list(set(str(id) for id in _session_setup_small_model_ids))
    
    # 执行删除
    _delete_models(base_url, token, allure, all_llm_ids_to_delete, 
                   f"{base_url}/api/mf-model-manager/v1/llm/delete", "大")
    _delete_models(base_url, token, allure, all_small_model_ids_to_delete, 
                   f"{base_url}/api/mf-model-manager/v1/small-model/delete", "小")
