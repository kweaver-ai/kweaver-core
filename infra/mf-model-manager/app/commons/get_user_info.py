import json

import aiohttp
from app.core.config import base_config
from app.commons.errors import UserManagementError
from app.core.config import base_config


async def get_username_by_ids(user_ids):
    if base_config.DEBUG:
        return {}
    if not user_ids:
        return {}
    user_management_url = f"http://{base_config.USERMANAGEMENTPRIVATEHOST}:{base_config.USERMANAGEMENTPRIVATEPORT}/api/user-management/v1/batch-get-user-info"
    user_management_app_url = f"http://{base_config.USERMANAGEMENTPRIVATEHOST}:{base_config.USERMANAGEMENTPRIVATEPORT}/api/user-management/v2/names"
    payload = {
        "user_ids": user_ids,
        "method": "GET",
        "fields": [
            "name"
        ]
    }
    
    # 存储最终的用户信息
    final_user_infos = {}
    
    for i in range(2):
        async with aiohttp.ClientSession() as session:
            async with session.post(user_management_url, json=payload) as response:
                if response.status != 200:
                    if response.status == 404:
                        res = await response.text()
                        result = json.loads(res)
                        invalid_ids = result.get("detail", {}).get("ids", [])
                        effective_ids = [user_id for user_id in user_ids if user_id not in invalid_ids]
                        payload["user_ids"] = effective_ids
                        
                        # 如果有无效ID，调用应用名称接口
                        if invalid_ids:
                            app_payload = {
                                "method": "GET",
                                "app_ids": invalid_ids
                            }
                            
                            # 调用应用名称接口
                            async with session.post(user_management_app_url, json=app_payload) as app_response:
                                if app_response.status == 200:
                                    app_res = await app_response.text()
                                    app_result = json.loads(app_res)
                                    # 获取应用名称信息
                                    app_names = app_result.get("app_names", [])
                                    for app_info in app_names:
                                        final_user_infos[app_info['id']] = app_info['name']
                                elif app_response.status == 400:
                                    app_res = await app_response.text()
                                    app_result = json.loads(app_res)
                                    # 处理400错误，排除无效的应用ID
                                    invalid_app_ids = app_result.get("detail", {}).get("ids", [])
                                    valid_app_ids = [app_id for app_id in invalid_ids if app_id not in invalid_app_ids]
                                    
                                    if valid_app_ids:
                                        # 重新调用应用名称接口，只包含有效的应用ID
                                        app_payload["app_ids"] = valid_app_ids
                                        async with session.post(user_management_app_url, json=app_payload) as retry_response:
                                            if retry_response.status == 200:
                                                retry_res = await retry_response.text()
                                                retry_result = json.loads(retry_res)
                                                app_names = retry_result.get("app_names", [])
                                                for app_info in app_names:
                                                    final_user_infos[app_info['id']] = app_info['name']
                                else:
                                    # 其他错误状态，抛出异常
                                    raise Exception("user-management app service error,please check")
                        
                        continue
                    raise Exception("user-management service error,please check")
                else:
                    res = await response.text()
                    result = json.loads(res)
                    user_infos = {info['id']: info['name'] for info in result}
                    # 合并用户信息和应用信息
                    final_user_infos.update(user_infos)
                    return final_user_infos


async def get_userid_by_search(result):
    user_ids = []
    if base_config.DEBUG:
        return user_ids
    for line in result:
        create_by = line["f_create_by"]
        update_by = line["f_update_by"]
        user_ids.append(create_by) if create_by else user_ids
        user_ids.append(update_by) if update_by else user_ids
    user_ids = list(set(user_ids))
    return user_ids
