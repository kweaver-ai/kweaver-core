import requests
import argparse
import warnings
warnings.filterwarnings('ignore')

def call_api_method(host, token, dag_id):
    """
    Calls an API method with the specified host and authorization token.

    :param host: The host URL of the API.
    :param token: Authorization token for the API.
    :return: Response from the API.
    """
    headers = {'Authorization': f'Bearer {token}'}
    url = f'https://{host}/api/automation/v1/dag/{dag_id}'  # 修改为实际的API端点
    response = requests.delete(url, headers=headers, verify=False)
    response.raise_for_status()
    return response.json()  # 假设API返回JSON格式数据

def main():
    parser = argparse.ArgumentParser(description='API Call Script')
    parser.add_argument('--host', type=str, help='Host of the API', required=True)
    parser.add_argument('--token', type=str, help='Authorization token', required=True)
    parser.add_argument('--dags', nargs='+', help='List of dag IDs.', required=True)

    args = parser.parse_args()
    for dag in args.dags:
        try:
            call_api_method(args.host, args.token, dag)
            print("delete dag success, id is {}".format(dag))
        except Exception as e:
            print("delete dag failed, id is {}, error is {}".format(dag, e))
            continue

if __name__ == "__main__":
    main()