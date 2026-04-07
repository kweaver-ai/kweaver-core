# coding=utf-8

import requests
import argparse
import json
from datetime import datetime
import warnings
warnings.filterwarnings('ignore')

def call_api_method(host, token, data, id):
    """
    Calls an API method with the specified host and authorization token.

    :param host: The host URL of the API.
    :param token: Authorization token for the API.
    :return: Response from the API.
    """
    # headers = {'Authorization': f'Bearer {token}'}
    headers = {'Authorization': 'Bearer {}'.format(token)}
    # url = f'{host}/api/automation/v1/dag'  # 修改为实际的API端点
    url = '{}/api/automation/v1/dag/{}'.format(host, id)  # 修改为实际的API端点
    response = requests.put(url, headers=headers, json=data, verify=False)
    if response.status_code > 300:
        print(response.json())
    response.raise_for_status()

data_string = datetime.now().strftime("%Y%m%d%H%M")
code = "import re\nimport requests\nfrom aishu_anyshare_api.api_client import ApiClient\n\ndef main(name, id):\n    pattern = re.compile(r'([^\\（]+)\\（([^）]+)\\）')\n    match = pattern.search(name)\n    chinese_name = name\n    english_name = ''\n    dep = ''\n    if match:\n        chinese_name = match.group(1)\n        english_name = match.group(2)\n\n    try:\n        url = \"{}/api/kc-mc/v2/user-card-info?user_id={}\".format(ApiClient.get_global_host(), id)\n        headers = {'Content-Type': 'application/json','Authorization': 'Bearer {}'.format(ApiClient.get_global_access_token())}\n        resp = requests.request(method=\"GET\", url= url, headers=headers, verify=False)\n        _json = resp.json()\n        _dep = _json[\"data\"]['out_dep']\n        dep = ';'.join(_dep)\n    except:\n        return chinese_name, english_name, dep\n        \n    return chinese_name, english_name, dep"
def main():
    parser = argparse.ArgumentParser(description='API Call Script')
    parser.add_argument('--host', type=str, help='Host of the API', required=True)
    parser.add_argument('--token', type=str, help='Authorization token', required=True)
    # parser.add_argument('--docids', nargs='+', help='List of document IDs.', required=True)
    # parser.add_argument('--cron', type=str, help='Enable create cron task.', required=False)
    # parser.add_argument('--emails', nargs='+', help='Enable email notify.', required=False)
    args = parser.parse_args()
    # headers = {'Authorization': f'Bearer {token}'}
    headers = {'Authorization': 'Bearer {}'.format(args.token)}
    host = args.host
    if not host.startswith('http'):
        host = "https://"+host
    # url = f'{host}/api/automation/v1/dag'  # 修改为实际的API端点
    url = '{}/api/automation/v1/dags?limit=100'.format(host)  # 修改为实际的API端点
    response = requests.get(url, headers=headers, verify=False)
    if response.status_code != 200:
        print("get dag list failed, err: {}".format(response.status_code))
        return
    dagLists = response.json()
    for dag in dagLists['dags']:
        headers = {'Authorization': 'Bearer {}'.format(args.token)}
        # url = f'{host}/api/automation/v1/dag'  # 修改为实际的API端点
        url = '{}/api/automation/v1/dag/{}'.format(host, dag['id'])  # 修改为实际的API端点
        res = requests.get(url, headers=headers, verify=False)
        dag = res.json()
        try:
            id = dag['id']
            first_step = dag['steps'][0]
            if first_step['operator'] =='@trigger/security-policy':
                continue
            del dag['id']
            del dag['created_at']
            del dag['updated_at']
            del dag['cron']
            del dag['shortcuts']
            del dag['accessors']
            
            if first_step['operator'] == '@trigger/cron':
                if first_step['dataSource']['operator'] == '@anyshare-data/user':
                    
                    dag['steps'][1] = {
                    "id": "1",
                    "operator": "@internal/tool/py3",
                    "parameters": {
                        "input_params": [
                            {
                                "id": "bgknp",
                                "key": "name",
                                "type": "string",
                                "value": "{{__0.name}}"
                            },
                            {
                                "id": "ttt",
                                "key": "id",
                                "type": "string",
                                "value": "{{__2.id}}"
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
                            },
                            {
                                "id": "rrttt",
                                "key": "parent_depth",
                                "type": "string"
                            }
                        ],
                        "code": code
                    },
                }
                    dag['steps'][2] = {
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
                        
                    response = call_api_method(host, args.token, dag, id)
                    print("update dag success, id is {}, title is {}".format(id, dag['title']))

            if first_step['operator'] == '@anyshare-trigger/create-user':                    
                dag['steps'][1] = {
                    "id": "1",
                    "operator": "@internal/tool/py3",
                    "parameters": {
                        "input_params": [
                            {
                                "id": "bgknp",
                                "key": "name",
                                "type": "string",
                                "value": "{{__0.name}}"
                            },
                            {
                                "id": "ttt",
                                "key": "id",
                                "type": "string",
                                "value": "{{__0.id}}"
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
                            },
                            {
                                "id": "rrttt",
                                "key": "parent_depth",
                                "type": "string"
                            }
                        ],
                        "code": code
                    },
                }
                dag['steps'][2] = {
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
                
                response = call_api_method(host, args.token, dag, id)
                print("update dag success, id is {}, title is {}".format(id, dag['title']))

            if first_step['operator'] == '@anyshare-trigger/change-user':                    
                dag['steps'][1] = {
                    "id": "1",
                    "operator": "@internal/tool/py3",
                    "parameters": {
                        "input_params": [
                            {
                                "id": "bgknp",
                                "key": "name",
                                "type": "string",
                                "value": "{{__0.name}}"
                            },
                            {
                                "id": "ttt",
                                "key": "id",
                                "type": "string",
                                "value": "{{__0.id}}"
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
                            },
                            {
                                "id": "rrttt",
                                "key": "parent_depth",
                                "type": "string"
                            }
                        ],
                        "code": code
                    },
                }
                dag['steps'][2] = {
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
                    
                response = call_api_method(host, args.token, dag, id)
                print("update dag success, id is {}, title is {}".format(id, dag['title']))

        except Exception as e:
            print("update dag failed, id is {}, title is {}, error is {}".format(id, dag['title'], e))
            continue

if __name__ == "__main__":
    main()