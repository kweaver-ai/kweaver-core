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
                if first_step['dataSource']['operator'] == '@anyshare-data/dept-tree':
                    
                    dag['steps'][1] = {
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
                        
                    response = call_api_method(host, args.token, dag, id)
                    print("update dag success, id is {}, title is {}".format(id, dag['title']))

            if first_step['operator'] == '@anyshare-trigger/move-dept':                    
                dag['steps'][1] = {
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
                
                response = call_api_method(host, args.token, dag, id)
                print("update dag success, id is {}, title is {}".format(id, dag['title']))

            if first_step['operator'] == '@anyshare-trigger/create-dept':                    
                dag['steps'][1] = {
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
                    
                response = call_api_method(host, args.token, dag, id)
                print("update dag success, id is {}, title is {}".format(id, dag['title']))

        except Exception as e:
            print("update dag failed, id is {}, title is {}, error is {}".format(id, dag['title'], e))
            continue

if __name__ == "__main__":
    main()