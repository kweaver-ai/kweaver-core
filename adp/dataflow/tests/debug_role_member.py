#!/usr/bin/env python3
# -*- coding:UTF-8 -*-
"""
单独调试 RoleMember fixture 的脚本
用于测试将 A0 用户设置为 AI 管理员的功能
"""

import sys
import os
import json

# 添加项目根目录到路径
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from common.get_content import GetContent
from common.create_user import CreateUser
from common.get_token import GetToken
from lib.permission import Perm

def debug_role_member():
    """调试 RoleMember 功能"""
    print("=" * 80)
    print("开始调试 RoleMember")
    print("=" * 80)
    
    # 1. 读取配置
    configfile = "./config/env.ini"
    file = GetContent(configfile)
    config = file.config()
    host = config["server"]["host"]
    print(f"\n1. 配置信息:")
    print(f"   Host: {host}")
    
    # 2. 创建用户 A0（模拟 APrepare）
    print(f"\n2. 创建用户 A0（模拟 APrepare）:")
    try:
        client = CreateUser(host=host)
        orgId = client.CreateOrganization("AISHU")
        print(f"   组织 ID: {orgId}")
        
        depId = client.AddDepartment(orgId, "测试部")
        print(f"   部门 ID: {depId}")
        
        depIds = [depId]
        userId = client.AddUser("A0", depIds, orgId)
        print(f"   用户 ID (A0): {userId}")
        
        if hasattr(userId, '__class__') and hasattr(userId, '__name__'):
            print(f"   ❌ 用户创建失败: {userId}")
            return False
    except Exception as e:
        print(f"   ❌ 创建用户时发生异常: {str(e)}")
        import traceback
        traceback.print_exc()
        return False
    
    # 3. 获取管理员 token
    print(f"\n3. 获取管理员 token:")
    try:
        token = GetToken(host=host).get_token(host, "admin", "eisoo.com123")
        admin_token = token[1]
        admin_user_id = token[0]
        print(f"   管理员用户 ID: {admin_user_id}")
        print(f"   管理员 Token: {admin_token[:50]}...")
    except Exception as e:
        print(f"   ❌ 获取管理员 token 失败: {str(e)}")
        import traceback
        traceback.print_exc()
        return False
    
    # 4. 准备权限客户端
    print(f"\n4. 准备权限客户端:")
    try:
        perm_client = Perm()
        print(f"   Base URL: {perm_client.base_url}")
    except Exception as e:
        print(f"   ❌ 创建权限客户端失败: {str(e)}")
        import traceback
        traceback.print_exc()
        return False
    
    # 5. 设置角色成员
    print(f"\n5. 设置角色成员:")
    roleid = "3fb94948-5169-11f0-b662-3a7bdba2913f"
    print(f"   角色 ID: {roleid}")
    print(f"   用户 ID: {userId}")
    
    headers = {
        "Authorization": f"Bearer {admin_token}"
    }
    
    data = {
        "method": "POST",
        "members": [
            {
                "id": userId,
                "type": "user"
            }
        ]
    }
    
    print(f"\n   请求 Headers:")
    print(f"   {json.dumps(headers, indent=2, ensure_ascii=False)}")
    print(f"\n   请求 Body:")
    print(f"   {json.dumps(data, indent=2, ensure_ascii=False)}")
    
    try:
        result = perm_client.ManageMember(roleid, data, headers)
        print(f"\n   响应状态码: {result[0]}")
        print(f"   响应内容:")
        if isinstance(result[1], dict):
            print(f"   {json.dumps(result[1], indent=2, ensure_ascii=False)}")
        else:
            print(f"   {result[1]}")
        
        if result[0] == 204:
            print(f"\n   ✅ 成功！用户 A0 已被设置为 AI 管理员")
            return True
        else:
            print(f"\n   ❌ 失败！状态码: {result[0]}")
            if isinstance(result[1], dict):
                error_code = result[1].get("code", "Unknown")
                error_desc = result[1].get("description", "Unknown")
                print(f"   错误代码: {error_code}")
                print(f"   错误描述: {error_desc}")
            return False
    except Exception as e:
        print(f"\n   ❌ 调用 ManageMember 时发生异常: {str(e)}")
        import traceback
        traceback.print_exc()
        return False

def cleanup_user():
    """清理测试用户"""
    print(f"\n" + "=" * 80)
    print("清理测试用户")
    print("=" * 80)
    
    try:
        from common.delete_user import DeleteUser
        configfile = "./config/env.ini"
        file = GetContent(configfile)
        config = file.config()
        host = config["server"]["host"]
        
        token = GetToken(host=host).get_token(host, "admin", "eisoo.com123")
        admin_token = token[1]
        
        client = DeleteUser(host=host)
        # 注意：这里需要知道用户 ID，实际清理时需要从之前的步骤获取
        print("   提示：需要手动清理用户和组织")
    except Exception as e:
        print(f"   清理时发生异常: {str(e)}")

if __name__ == "__main__":
    print("\n" + "=" * 80)
    print("RoleMember 调试脚本")
    print("=" * 80)
    
    success = debug_role_member()
    
    print("\n" + "=" * 80)
    if success:
        print("✅ 调试完成：RoleMember 设置成功")
    else:
        print("❌ 调试完成：RoleMember 设置失败")
        print("\n请检查以下内容：")
        print("1. 角色 ID 是否正确: 3fb94948-5169-11f0-b662-3a7bdba2913f")
        print("2. 用户 A0 是否已成功创建")
        print("3. 管理员账号和密码是否正确")
        print("4. 权限服务是否正常运行")
        print("5. API 路径是否正确: /api/authorization/v1/role-members/{roleid}")
    print("=" * 80)
    
    # 询问是否清理
    # cleanup_user()
