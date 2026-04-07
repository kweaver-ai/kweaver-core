# coding=utf-8

import requests
import argparse
import json
from datetime import datetime
import warnings
warnings.filterwarnings('ignore')

def call_api_method(host, token, data):
    """
    Calls an API method with the specified host and authorization token.

    :param host: The host URL of the API.
    :param token: Authorization token for the API.
    :return: Response from the API.
    """
    # headers = {'Authorization': f'Bearer {token}'}
    headers = {'Authorization': 'Bearer {}'.format(token)}
    # url = f'https://{host}/api/automation/v1/dag'  # 修改为实际的API端点
    url = '{}/api/automation/v1/dag'.format(host)  # 修改为实际的API端点
    response = requests.post(url, headers=headers, data=data, verify=False)
    print(response)
    response.raise_for_status()
    return response.json()  # 假设API返回JSON格式数据

data_string = datetime.now().strftime("%Y%m%d%H%M")

def getDags(docids, cron_switch = 'false', emails = []):
    dags = [
        json.dumps({
            "emails": emails,
            "title": "新建文件" + data_string,
            "description": "",
            "status": "normal",
            "shortcuts": [],
            "accessors": [],
            "appinfo":{
                "enable": True
            },    
            "steps": [
                {
                    "id": "0",
                    "operator": "@anyshare-trigger/upload-file",
                    "parameters": {
                        "docids": docids,
                        "inherit": True
                    }
                },
                {
                    "id": "1",
                    "operator": "@intelliinfo/transfer",
                    "parameters": {
                        "rule_id": "document_upsert",
                        "data": {
                            "id": "{{__0.item_id}}",
                            "name": "{{__0.name}}",
                            "rev": "{{__0.rev}}",
                            "create_time": "{{__0.create_time}}",
                            "modify_time": "{{__0.modify_time}}",
                            "size": "{{__0.size}}",
                            "path": "{{__0.path}}",
                            "csflevel": "{{__0.csflevel}}",
                            "creator": "{{__0.creator_id}}",
                            "editor": "{{__0.editor_id}}"
                        }
                    }
                }
            ]
        }),
        json.dumps({
            "emails": emails,
            "title": "文件还原版本" + data_string,
            "description": "",
            "status": "normal",
            "shortcuts": [],
            "accessors": [],
            "appinfo":{
                "enable": True
            },    
            "steps": [
                {
                    "id": "0",
                    "operator": "@anyshare-trigger/reversion-file",
                    "parameters": {
                        "docids": docids,
                        "inherit": True
                    }
                },
                {
                    "id": "1",
                    "operator": "@intelliinfo/transfer",
                    "parameters": {
                        "rule_id": "document_upsert",
                        "data": {
                            "id": "{{__0.item_id}}",
                            "name": "{{__0.name}}",
                            "rev": "{{__0.rev}}",
                            "create_time": "{{__0.create_time}}",
                            "modify_time": "{{__0.modify_time}}",
                            "size": "{{__0.size}}",
                            "path": "{{__0.path}}",
                            "csflevel": "{{__0.csflevel}}",
                            "creator": "{{__0.creator_id}}",
                            "editor": "{{__0.editor_id}}"
                        }
                    }
                }
            ]
        }),
        json.dumps({
            "emails": emails,
            "title": "从回收站还原文件" + data_string,
            "description": "",
            "status": "normal",
            "shortcuts": [],
            "accessors": [],
            "appinfo":{
                "enable": True
            },    
            "steps": [
                {
                    "id": "0",
                    "operator": "@anyshare-trigger/restore-file",
                    "parameters": {
                        "docids": docids,
                        "inherit": True
                    }
                },
                {
                    "id": "1",
                    "operator": "@intelliinfo/transfer",
                    "parameters": {
                        "rule_id": "document_upsert",
                        "data": {
                            "id": "{{__0.item_id}}",
                            "name": "{{__0.name}}",
                            "rev": "{{__0.rev}}",
                            "create_time": "{{__0.create_time}}",
                            "modify_time": "{{__0.modify_time}}",
                            "size": "{{__0.size}}",
                            "path": "{{__0.path}}",
                            "csflevel": "{{__0.csflevel}}",
                            "creator": "{{__0.creator_id}}",
                            "editor": "{{__0.editor_id}}"
                        }
                    }
                }
            ]
        }),
        json.dumps({
            "emails": emails,
            "title": "复制文件触发" + data_string,
            "description": "",
            "status": "normal",
            "shortcuts": [],
            "accessors": [],
            "appinfo":{
                "enable": True
            },    
            "steps": [
                {
                    "id": "0",
                    "operator": "@anyshare-trigger/copy-file",
                    "parameters": {
                        "docids": docids,
                        "inherit": True
                    }
                },
                {
                    "id": "1",
                    "operator": "@intelliinfo/transfer",
                    "parameters": {
                        "rule_id": "document_upsert",
                        "data": {
                            "id": "{{__0.new_item_id}}",
                            "name": "{{__0.name}}",
                            "rev": "{{__0.rev}}",
                            "create_time": "{{__0.create_time}}",
                            "modify_time": "{{__0.modify_time}}",
                            "size": "{{__0.size}}",
                            "path": "{{__0.path}}",
                            "csflevel": "{{__0.csflevel}}",
                            "creator": "{{__0.creator_id}}",
                            "editor": "{{__0.editor_id}}"
                        }
                    }
                }
            ]
        }),
        json.dumps({
            "emails": emails,
            "title": "删除文件触发"+data_string,
            "description": "",
            "status": "normal",
            "shortcuts": [],
            "accessors": [],
            "appinfo":{
                "enable": True
            },    
            "steps": [
                {
                    "id": "0",
                    "operator": "@anyshare-trigger/remove-file",
                    "parameters": {
                        "docids": docids,
                        "inherit": True
                    }
                },
                {
                    "id": "1",
                    "operator": "@intelliinfo/transfer",
                    "parameters": {
                        "rule_id": "document_delete",
                        "data": {
                            "id": "{{__0.item_id}}",
                            "path": "{{__0.path}}"
                        }
                    }
                }
            ]
        }),
        json.dumps({
            "emails": emails,
            "title": "彻底删除文件触发" + data_string,
            "description": "",
            "status": "normal",
            "shortcuts": [],
            "accessors": [],
            "appinfo":{
                "enable": True
            },    
            "steps": [
                {
                    "id": "0",
                    "operator": "@anyshare-trigger/delete-file",
                    "parameters": {
                        "docids": docids,
                        "inherit": True
                    }
                },
                {
                    "id": "1",
                    "operator": "@intelliinfo/transfer",
                    "parameters": {
                        "rule_id": "document_delete",
                        "data": {
                            "id": "{{__0.item_id}}",
                            "path": "{{__0.path}}"
                        }
                    }
                }
            ]
        }),
        json.dumps({
            "emails": emails,
            "title": "重命名文件"+data_string,
            "description": "",
            "status": "normal",
            "shortcuts": [],
            "accessors": [],
            "appinfo":{
                "enable": True
            },    
            "steps": [
                {
                    "id": "0",
                    "operator": "@anyshare-trigger/rename-file",
                    "parameters": {
                        "docids": docids,
                        "inherit": True
                    }
                },
                {
                    "id": "1",
                    "operator": "@intelliinfo/transfer",
                    "parameters": {
                        "rule_id": "document_update",
                        "data": {
                            "id": "{{__0.item_id}}",
                            "name": "{{__0.name}}",
                            "rev": "{{__0.rev}}",
                            "create_time": "{{__0.create_time}}",
                            "modify_time": "{{__0.modify_time}}",
                            "size": "{{__0.size}}",
                            "path": "{{__0.path}}",
                            "csflevel": "{{__0.csflevel}}",
                            "creator": "{{__0.creator_id}}",
                            "editor": "{{__0.editor_id}}"
                        }
                    }
                }
            ]
        }),
        json.dumps({
            "emails": emails,
            "title": "移动文件"+data_string,
            "description": "",
            "status": "normal",
            "shortcuts": [],
            "accessors": [],
            "appinfo":{
                "enable": True
            },    
            "steps": [
                {
                    "id": "0",
                    "operator": "@anyshare-trigger/move-file",
                    "parameters": {
                        "docids": docids,
                        "inherit": True
                    }
                },
                {
                    "id": "1",
                    "operator": "@intelliinfo/transfer",
                    "parameters": {
                        "rule_id": "document_update",
                        "data": {
                            "id": "{{__0.item_id}}",
                            "name": "{{__0.name}}",
                            "rev": "{{__0.rev}}",
                            "create_time": "{{__0.create_time}}",
                            "modify_time": "{{__0.modify_time}}",
                            "size": "{{__0.size}}",
                            "path": "{{__0.path}}",
                            "csflevel": "{{__0.csflevel}}",
                            "creator": "{{__0.creator_id}}",
                            "editor": "{{__0.editor_id}}"
                        }
                    }
                }
            ]
        }),
        json.dumps({
            "emails": emails,
            "title": "创建用户"+data_string,
            "description": "",
            "status": "normal",
            "shortcuts": [],
            "accessors": [],
            "appinfo":{
                "enable": True
            },    
            "steps": [
                {
                    "id": "0",
                    "operator": "@anyshare-trigger/create-user",
                    "parameters": {
                    }
                },
                {
                    "id": "1",
                    "operator": "@internal/tool/py3",
                    "parameters": {
                        "input_params": [
                            {
                                "id": "bgknp",
                                "key": "name",
                                "type": "string",
                                "value": "{{__0.name}}"
                            }
                        ],
                        "output_params": [
                            {
                                "id": "w3dty",
                                "key": "chinese_name",
                                "type": "string"
                            },
                            {
                                "id": "rfq4v",
                                "key": "english_name",
                                "type": "string"
                            }
                        ],
                        "code": "import re\n\ndef main(name):\n    pattern = re.compile(r'([^\\（]+)\\（([^）]+)\\）')\n    match = pattern.search(name)\n    chinese_name = name\n    english_name = ''\n    if match:\n        chinese_name = match.group(1)\n        english_name = match.group(2)\n        \n    return chinese_name, english_name"
                    },
                },
                {
                    "id": "2",
                    "operator": "@intelliinfo/transfer",
                    "parameters": {
                        "rule_id": "person_upsert",
                        "data": {
                            "id": "{{__0.id}}",
                            "name": "{{__1.chinese_name}}",
                            "english_name": "{{__1.english_name}}",
                            "contact": "{{__0.contact}}",
                            "email": "{{__0.email}}",
                            "role": "{{__0.role}}",
                            "csflevel": "{{__0.csflevel}}",
                            "status": "{{__0.status}}",
                            "parent_ids": "{{__0.parent_ids}}",
                            "parent": "{{__1.parent_depth}}"
                        }
                    }
                }
            ]
        }),
        json.dumps({
            "emails": emails,
            "title": "删除用户"+data_string,
            "description": "",
            "status": "normal",
            "shortcuts": [],
            "accessors": [],
            "appinfo":{
                "enable": True
            },    
            "steps": [
                {
                    "id": "0",
                    "operator": "@anyshare-trigger/delete-user",
                    "parameters": {
                    }
                },
                {
                    "id": "1",
                    "operator": "@intelliinfo/transfer",
                    "parameters": {
                        "rule_id": "person_delete",
                        "data": {
                            "id": "{{__0.id}}"
                        }
                    }
                }
            ]
        }),
        json.dumps({
            "emails": emails,
            "title": "更新用户信息"+data_string,
            "description": "",
            "status": "normal",
            "shortcuts": [],
            "accessors": [],
            "appinfo":{
                "enable": True
            },    
            "steps": [
                {
                    "id": "0",
                    "operator": "@anyshare-trigger/change-user",
                    "parameters": {
                    }
                },
                {
                    "id": "1",
                    "operator": "@internal/tool/py3",
                    "parameters": {
                        "input_params": [
                            {
                                "id": "bgknp",
                                "key": "name",
                                "type": "string",
                                "value": "{{__0.name}}"
                            }
                        ],
                        "output_params": [
                            {
                                "id": "w3dty",
                                "key": "chinese_name",
                                "type": "string"
                            },
                            {
                                "id": "rfq4v",
                                "key": "english_name",
                                "type": "string"
                            }
                        ],
                        "code": "import re\n\ndef main(name):\n    pattern = re.compile(r'([^\\（]+)\\（([^）]+)\\）')\n    match = pattern.search(name)\n    chinese_name = name\n    english_name = ''\n    if match:\n        chinese_name = match.group(1)\n        english_name = match.group(2)\n        \n    return chinese_name, english_name"
                    },
                },
                {
                    "id": "2",
                    "operator": "@intelliinfo/transfer",
                    "parameters": {
                        "rule_id": "person_update",
                        "data": {
                            "id": "{{__0.id}}",
                            "name": "{{__1.chinese_name}}",
                            "english_name": "{{__1.english_name}}",
                            "email": "{{__0.email}}",
                            "tags": "{{__0.tags}}",
                            "is_expert": "{{__0.is_expert}}",
                            "verification_info": "{{__0.verification_info}}",
                            "university": "{{__0.university}}",
                            "contact": "{{__0.contact}}",
                            "position": "{{__0.position}}",
                            "work_at": "{{__0.work_at}}",
                            "professional": "{{__0.professional}}",
                            "status": "{{__0.status}}",
                            "parent": "{{__1.parent_depth}}"
                        }
                    }
                }
            ]
        }),
        json.dumps({
            "emails": emails,
            "title": "组织名称变更"+data_string,
            "description": "",
            "status": "normal",
            "shortcuts": [],
            "accessors": [],
            "appinfo":{
                "enable": True
            },    
            "steps": [
                {
                    "id": "0",
                    "operator": "@anyshare-trigger/modify-org-name",
                    "parameters": {
                    }
                },
                {
                    "id": "1",
                    "operator": "@intelliinfo/transfer",
                    "parameters": {
                        "rule_id": "orgnization_update",
                        "data": {
                            "id": "{{__0.id}}",
                            "new_name": "{{__0.new_name}}"
                        }
                    }
                }
            ]
        }),
        json.dumps({
            "emails": emails,
            "title": "新增组织及部门触发"+data_string,
            "description": "",
            "status": "normal",
            "shortcuts": [],
            "accessors": [],
            "appinfo":{
                "enable": True
            },    
            "steps": [
                {
                    "id": "0",
                    "operator": "@anyshare-trigger/create-dept",
                    "parameters": {
                    }
                },
                {
                    "id": "1",
                    "operator": "@intelliinfo/transfer",
                    "parameters": {
                        "rule_id": "orgnization_upsert",
                        "data": {
                            "id": "{{__0.id}}",
                            "name": "{{__0.name}}",
                            "parent_id": "{{__0.parent_id}}",
                            "email": "{{__0.email}}",
                            "parent": "{{__0.parent}}",
                        }
                    }
                }
            ]
        }),
        json.dumps({
            "emails": emails,
            "title": "移动组织及部门触发"+data_string,
            "description": "",
            "status": "normal",
            "shortcuts": [],
            "accessors": [],
            "appinfo":{
                "enable": True
            },    
            "steps": [
                {
                    "id": "0",
                    "operator": "@anyshare-trigger/move-dept",
                    "parameters": {
                    }
                },
                {
                    "id": "1",
                    "operator": "@intelliinfo/transfer",
                    "parameters": {
                        "rule_id": "orgnization_upsert",
                        "data": {
                            "id": "{{__0.id}}",
                            "name": "{{__0.name}}",
                            "parent_id": "{{__0.parent_id}}",
                            "parent": "{{__0.parent}}",
                            "email": "{{__0.email}}"
                        }
                    }
                }
            ]
        }),
        json.dumps({
            "emails": emails,
            "title": "删除组织及部门触发"+data_string,
            "description": "",
            "status": "normal",
            "shortcuts": [],
            "accessors": [],
            "appinfo":{
                "enable": True
            },    
            "steps": [
                {
                    "id": "0",
                    "operator": "@anyshare-trigger/delete-dept",
                    "parameters": {
                    }
                },
                {
                    "id": "1",
                    "operator": "@intelliinfo/transfer",
                    "parameters": {
                        "rule_id": "orgnization_delete",
                        "data": {
                            "id": "{{__0.id}}"
                        }
                    }
                }
            ]
        }),
        json.dumps({
            "emails": emails,
            "title": "用户移动触发"+data_string,
            "description": "",
            "status": "normal",
            "shortcuts": [],
            "accessors": [],
            "appinfo":{
                "enable": True
            },    
            "steps": [
                {
                    "id": "0",
                    "operator": "@anyshare-trigger/move-user",
                    "parameters": {}
                },
                {
                    "id": "1",
                    "operator": "@intelliinfo/transfer",
                    "parameters": {
                        "rule_id": "orgnization_relation_update",
                        "data": {
                            "id": "{{__0.id}}",
                            "new_dept_path": "{{__0.new_dept_path}}",
                            "old_dept_path": "{{__0.old_dept_path}}"
                        }
                    }
                }
            ]
        }),
        json.dumps({
            "emails": emails,
            "title": "更新用户组织信息-添加用户到部门"+data_string,
            "description": "",
            "status": "normal",
            "shortcuts": [],
            "accessors": [],
            "appinfo":{
                "enable": True
            },    
            "steps": [
                {
                    "id": "0",
                    "operator": "@anyshare-trigger/add-user-to-dept",
                    "parameters": {}
                },
                {
                    "id": "1",
                    "operator": "@intelliinfo/transfer",
                    "parameters": {
                        "rule_id": "orgnization_relation_update",
                        "data": "{{__0.data}}"
                    }
                }
            ]
        }),
        json.dumps({
            "emails": emails,
            "title": "更新用户组织信息-将用户从部门移除"+data_string,
            "description": "",
            "status": "normal",
            "shortcuts": [],
            "accessors": [],
            "appinfo":{
                "enable": True
            },    
            "steps": [
                {
                    "id": "0",
                    "operator": "@anyshare-trigger/remove-user-from-dept",
                    "parameters": {}
                },
                {
                    "id": "1",
                    "operator": "@intelliinfo/transfer",
                    "parameters": {
                        "rule_id": "orgnization_relation_update",
                        "data": "{{__0.data}}"
                    }
                }
            ]
        }),
        json.dumps({
            "emails": emails,
            "title": "创建标签树触发"+data_string,
            "description": "",
            "status": "normal",
            "shortcuts": [],
            "accessors": [],
            "appinfo":{
                "enable": True
            },    
            "steps": [
                {
                    "id": "0",
                    "operator": "@anyshare-trigger/create-tag-tree",
                    "parameters": {
                    }
                },
                {
                    "id": "1",
                    "operator": "@intelliinfo/transfer",
                    "parameters": {
                        "rule_id": "tag_upsert",
                        "data": {
                            "id": "{{__0.id}}",
                            "path": "{{__0.path}}",
                            "version": "{{__0.version}}",
                            "name": "{{__0.name}}",
                            "parent_id": "{{__0.parent_id}}"
                    }
                    }
                }
            ]
        }),
        json.dumps({
            "emails": emails,
            "title": "添加标签触发"+data_string,
            "description": "",
            "status": "normal",
            "shortcuts": [],
            "accessors": [],
            "appinfo":{
                "enable": True
            },    
            "steps": [
                {
                    "id": "0",
                    "operator": "@anyshare-trigger/add-tag-tree",
                    "parameters": {
                    }
                },
                {
                    "id": "1",
                    "operator": "@intelliinfo/transfer",
                    "parameters": {
                        "rule_id": "tag_upsert",
                        "data": "{{__0.tags}}"
                    }
                }
            ]
        }),
        json.dumps({
            "emails": emails,
            "title": "编辑标签触发"+data_string,
            "description": "",
            "status": "normal",
            "shortcuts": [],
            "accessors": [],
            "appinfo":{
                "enable": True
            },    
            "steps": [
                {
                    "id": "0",
                    "operator": "@anyshare-trigger/edit-tag-tree",
                    "parameters": {
                    }
                },
                {
                    "id": "1",
                    "operator": "@intelliinfo/transfer",
                    "parameters": {
                        "rule_id": "tag_update",
                        "data": "{{__0.tags}}"
                    }
                }
            ]
        }),
        json.dumps({
            "emails": emails,
            "title": "删除标签触发"+data_string,
            "description": "",
            "status": "normal",
            "shortcuts": [],
            "accessors": [],
            "appinfo":{
                "enable": True
            },    
            "steps": [
                {
                    "id": "0",
                    "operator": "@anyshare-trigger/delete-tag-tree",
                    "parameters": {
                    }
                },
                {
                    "id": "1",
                    "operator": "@intelliinfo/transfer",
                    "parameters": {
                        "rule_id": "tag_delete",
                        "data": "{{__0.tags}}"
                    }
                }
            ]
        })
    ]

    if cron_switch != 'true':
        return dags

    cron_dags = [
        json.dumps({
            "emails": emails,
            "title": "定时获取所有的用户"+data_string,
            "description": "",
            "status": "normal",
            "appinfo":{
                "enable": True
            },    
            "steps": [
                {
                    "id": "0",
                    "operator": "@trigger/cron",
                    "dataSource": {
                        "id": "2",
                        "operator": "@anyshare-data/user",
                        "parameters": {
                            "accessorid": "00000000-0000-0000-0000-000000000000"
                        }
                    },
                    "cron": "0 30 0 * * ?"
                },
                {
                    "id": "1",
                    "operator": "@internal/tool/py3",
                    "parameters": {
                        "input_params": [
                            {
                                "id": "bgknp",
                                "key": "name",
                                "type": "string",
                                "value": "{{__0.name}}"
                            }
                        ],
                        "output_params": [
                            {
                                "id": "w3dty",
                                "key": "chinese_name",
                                "type": "string"
                            },
                            {
                                "id": "rfq4v",
                                "key": "english_name",
                                "type": "string"
                            }
                        ],
                        "code": "import re\n\ndef main(name):\n    pattern = re.compile(r'([^\\（]+)\\（([^）]+)\\）')\n    match = pattern.search(name)\n    chinese_name = name\n    english_name = ''\n    if match:\n        chinese_name = match.group(1)\n        english_name = match.group(2)\n        \n    return chinese_name, english_name"
                    },
                },
                {
                    "id": "3",
                    "title": "",
                    "operator": "@intelliinfo/transfer",
                    "parameters" : {
                        "rule_id" : "person_upsert",
                        "data" : {
                            "role" : "{{__2.role}}",
                            "csflevel" : "{{__2.csflevel}}",
                            "id" : "{{__2.id}}",
                            "name" : "{{__1.chinese_name}}",
                            "english_name": "{{__1.english_name}}",
                            "status": "{{__2.status}}",
                            "contact" : "{{__2.contact}}",
                            "email" : "{{__2.email}}",
                            "parent_ids": "{{__2.parent_ids}}",
                            "tags": "{{__2.tags}}",
                            "is_expert": "{{__2.is_expert}}",
                            "verification_info": "{{__2.verification_info}}",
                            "university": "{{__2.university}}",
                            "position": "{{__2.position}}",
                            "work_at": "{{__2.work_at}}",
                            "professional": "{{__2.professional}}",
                            "parent": "{{__1.parent_depth}}"
                        }
                    }
                }
            ]
        }),
        json.dumps({
            "emails": emails,
            "title": "定时获取所有的组织和部门"+data_string,
            "description": "",
            "status": "normal",
            "appinfo":{
                "enable": True
            },    
            "steps": [
                {
                    "id": "0",
                    "operator": "@trigger/cron",
                    "dataSource": {
                        "id": "2",
                        "operator": "@anyshare-data/dept-tree",
                        "parameters": {
                            "accessorid": "00000000-0000-0000-0000-000000000000"
                        }
                    },
                    "cron": "0 0 0 * * ?"
                },
                {
                    "id": "1",
                    "title": "",
                    "operator": "@intelliinfo/transfer",
                    "parameters" : {
                        "rule_id" : "orgnization_upsert",
                        "data" : {
                            "id" : "{{__2.id}}",
                            "name" : "{{__2.name}}",
                            "parent_id" : "{{__2.parent_id}}",
                            "parent" : "{{__2.parent}}",
                            "email" : "{{__2.email}}"
                        }
                    }
                }
            ]
        }),
        json.dumps({
            "emails": emails,
            "title": "定时获取指定目录所有的文档"+data_string,
            "description": "",
            "status": "normal",
            "appinfo":{
                "enable": True
            },
        "steps": [
                {
                    "id": "0",
                    "operator": "@trigger/cron",
                    "dataSource": {
                        "id": "2",
                        "operator": "@anyshare-data/list-files",
                        "parameters": {
                            "docids" : docids,
                            "depth" : -1
                }
                    },
                    "cron": "0 0 1 * * ?"
                },
                {
                    "id": "1",
                    "title": "",
                    "operator": "@intelliinfo/transfer",
                    "parameters" : {
                        "rule_id" : "document_all_upsert",
                        "data" : {
                            "id": "{{__2.item_id}}",
                            "name": "{{__2.name}}",
                            "rev": "{{__2.rev}}",
                            "create_time": "{{__2.create_time}}",
                            "modify_time": "{{__2.modify_time}}",
                            "size": "{{__2.size}}",
                            "path": "{{__2.path}}",
                            "csflevel": "{{__2.csflevel}}",
                            "creator": "{{__2.creator_id}}",
                            "editor": "{{__2.editor_id}}"
                        }
                    }
                }
            ]
        }),
        json.dumps({
            "emails": emails,
            "title": "定时获取所有的标签树"+data_string,
            "description": "",
            "status": "normal",
            "appinfo":{
                "enable": True
            },
            "steps": [
                {
                    "id": "0",
                    "operator": "@trigger/cron",
                    "dataSource": {
                        "id": "2",
                        "operator": "@anyshare-data/tag-tree",
                        "parameters": {
                            
                        }
                    },
                    "cron": "0 0 0 * * ?"
                },
                {
                    "id": "1",
                    "title": "",
                    "operator": "@intelliinfo/transfer",
                    "parameters" : {
                        "rule_id" : "tag_upsert",
                        "data" : {                    
                            "id": "{{__2.id}}",
                            "path": "{{__2.path}}",
                            "version": "{{__2.version}}",
                            "name": "{{__2.name}}",
                            "parent_id": "{{__2.parent_id}}"
                }
                    }
                }
            ]
        })
    ]
    return dags + cron_dags

def main():
    parser = argparse.ArgumentParser(description='API Call Script')
    parser.add_argument('--host', type=str, help='Host of the API', required=True)
    parser.add_argument('--token', type=str, help='Authorization token', required=True)
    parser.add_argument('--docids', nargs='+', help='List of document IDs.', required=True)
    parser.add_argument('--cron', type=str, help='Enable create cron task.', required=False)
    parser.add_argument('--emails', nargs='+', help='Enable email notify.', required=False, default=[])

    args = parser.parse_args()
    host = args.host
    if not host.startswith('http'):
        host = "https://"+host
    dags = getDags(args.docids, args.cron, args.emails)
    for dag in dags:
        try:
            dag_json = json.loads(dag)
            title = dag_json["title"]
            response = call_api_method(host, args.token, dag)
            print("create dag success, id is {}, title is {}".format(response["id"], title))
        except Exception as e:
            print("create dag failed, title is {}, error is {}".format(title, e))
            continue

if __name__ == "__main__":
    main()