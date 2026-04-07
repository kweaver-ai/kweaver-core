# -*- coding:UTF-8 -*-

from common.get_content import GetContent
from common.request import Request

class AutomationClient():
    def __init__(self):
        file = GetContent("./config/env.ini")
        self.config = file.config()
        self.base_url = self.config["requests"]["protocol"] + "://" + self.config["server"]["host"] + ":" + self.config["server"]["port"] + "/api/automation/v1"

    '''创建组合算子'''
    def CreateCombinationOperator(self, data, headers):
        url = self.base_url + "/operators"
        return Request.post(self, url, data, headers)

    '''获取组合算子列表'''
    def GetOperatorsList(self, params, headers):
        url = self.base_url + "/operators"
        return Request.query(self, url, params, headers)

    '''获取组合算子详情'''
    def GetOperatorDetail(self, operator_id, headers):
        url = self.base_url + "/operators/" + operator_id   
        return Request.get(self, url, headers)
    

    '''更新组合算子'''
    def UpdateOperator(self, operator_id, data, headers):
        url = self.base_url + "/operators/" + operator_id
        result = Request.put(self, url, data, headers=headers)
        print(result)
        # 处理204 No Content响应
        if result[0] == 204:
            return [204, {}]  # 返回空对象而不是空字符串
        return result


    '''删除组合算子'''
    def DeleteOperator(self, operator_id, headers):
        url = self.base_url + "/operators/" + operator_id
        return Request.delete(self, url, {}, headers)
    

    

    '''运行组合算子'''
    def RuneOperator(self, dag_id, data , headers):
        url = self.base_url+"/operators/" + dag_id+"/executions" 
        return Request.post(self, url, data, headers)


    '''获取运行记录（v2）'''
    def GetDagResultsV2(self, dag_id, params, headers):
        base_url_v2 = self.base_url.replace('/v1', '/v2')
        url = f"{base_url_v2}/dag/{dag_id}/results"
        return Request.query(self, url, params, headers)

    '''获取执行日志（v2）'''
    def GetDagResultLogV2(self, dag_id, result_id, params, headers):
        base_url_v2 = self.base_url.replace('/v1', '/v2')
        url = f"{base_url_v2}/dag/{dag_id}/result/{result_id}"
        return Request.query(self, url, params, headers)
    
    
    '''流程全景可观测接口'''
    def GetFullView(self, params, headers):
        """
        获取流程全景统计数据
        Args:
            params (dict): 查询参数，包含 start_time, end_time, 可选 type
            headers (dict): 请求头
        Returns:
            (status_code, resp_json)
        """
        url = self.base_url + "/observability/full-view"
        return Request.query(self, url, params, headers)

    '''流程运行可观测接口'''
    def GetRunView(self, params, headers):
        """
        获取流程运行统计数据
        Args:
            params (dict): 查询参数，包含 start_time, end_time, 可选 type
            headers (dict): 请求头
        Returns:
            (status_code, resp_json)
        """
        url = self.base_url + "/observability/runtime-view"
        return Request.query(self, url, params, headers)

    '''创建工作流'''
    def CreateDag(self, data, headers):
        url = self.base_url + "/dag"
        return Request.post(self, url, data, headers)

    '''运行工作流实例'''
    def RunInstance(self, dag_id, data, headers):
        url = self.base_url + "/run-instance/" + dag_id
        return Request.post(self, url, data, headers)

    '''获取工作流详情'''
    def GetDagDetail(self, dag_id, headers):
        url = self.base_url + "/dag/" + dag_id
        return Request.get(self, url, headers)

    '''获取执行日志详情（v1）'''
    def GetDagResultDetail(self, dag_id, result_id, headers):
        url = self.base_url + "/dag/" + dag_id + "/result/" + result_id
        return Request.get(self, url, headers)

    '''获取工作流列表（v2）'''
    def ListDagsV2(self, params, headers):
        base_url_v2 = self.base_url.replace('/v1', '/v2')
        url = base_url_v2 + "/dags"
        return Request.query(self, url, params, headers)

    '''获取我的工作流列表（v1）'''
    def ListMyDags(self, params, headers):
        url = self.base_url + "/dags"
        return Request.query(self, url, params, headers)

    '''获取共享给我的工作流列表'''
    def ListSharedDags(self, params, headers):
        url = self.base_url + "/shared-dags"
        return Request.query(self, url, params, headers)

    '''删除数据流'''
    def DeleteDataFlow(self, flow_id, headers):
        url = self.base_url + "/data-flow/flow/" + flow_id
        return Request.delete(self, url, {}, headers)

    '''运行工作流实例（表单触发）'''
    def RunInstanceForm(self, dag_id, data, headers):
        url = self.base_url + "/run-instance-form/" + dag_id
        return Request.post(self, url, data, headers)