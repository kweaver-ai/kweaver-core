commmon目录放一些提炼的公告函数，下面是一个示例：

没有优化前的代码：
```python
    # lib目录下某个py文件中的部分函数
    @allure.step('撤销定密审核')
    def DeletePendingDetail(self, ip, token, apply_id):
        requrl = 'https://%s/api/document/v1/security_classification_approval/pending_detail/%s' % (ip, apply_id)
        dict1 = {}
        dict1['Authorization'] = "Bearer %s" % token
        r = requests.request('delete', requrl, verify=False, headers=dict1)
        allure.attach(r.url, "url", allure.attachment_type.TEXT)
        allure.attach(str(r.status_code), "status_code", allure.attachment_type.TEXT)
        allure.attach(r.content, "content", allure.attachment_type.TEXT)
        if r.status_code == 204:
            return r.status_code
        else:
            return r.status_code, json.loads(r.content)

    @allure.step('设置文件密级枚举')
    def SetConsoleFileClassifications(self, ip, tokenid, body):
        requrl = 'https://%s/api/document/v1/console/file-classifications' % (ip)
        dict1 = {}
        dict1['Authorization'] = "Bearer %s" % tokenid
        body = body
        r = requests.request('PUT', requrl, verify=False, json=body, headers=dict1)
        allure.attach(r.url, "url", allure.attachment_type.TEXT)
        allure.attach(str(r.status_code), "status_code", allure.attachment_type.TEXT)
        allure.attach(r.content, "content", allure.attachment_type.TEXT)
        if r.status_code == 204:
            return r.status_code
        else:
            return r.status_code, json.loads(r.content)

    @allure.step('设置系统密级')
    def SetConsoleSystemClassifications(self, ip, tokenid, body):
        requrl = 'https://%s/api/document/v1/console/system-classification' % (ip)
        dict1 = {}
        dict1['Authorization'] = "Bearer %s" % tokenid
        body = body
        r = requests.request('PUT', requrl, verify=False, json=body, headers=dict1)
        allure.attach(r.url, "url", allure.attachment_type.TEXT)
        allure.attach(str(r.status_code), "status_code", allure.attachment_type.TEXT)
        allure.attach(r.content, "content", allure.attachment_type.TEXT)
        if r.status_code == 204:
            return r.status_code
        else:
            return r.status_code, json.loads(r.content)
```
优化后的代码：

```python
# common目录下的某个py文件，提炼了一些公共的方法
@allure.step('API请求')
def _make_api_request(self, method, ip, token, endpoint, body=None, path_param=None):
    """基础API请求方法
    
    Args:
        method (str): 请求方法 (GET, POST, PUT, DELETE等)
        ip (str): 服务器IP
        token (str): 认证token
        endpoint (str): API路径
        body (dict, optional): 请求体
        path_param (str, optional): 路径参数
    
    Returns:
        int or tuple: 状态码(204)或状态码与响应内容的元组
    """
    # 构建URL
    if path_param:
        requrl = f'https://{ip}/api/document/v1/{endpoint}/{path_param}'
    else:
        requrl = f'https://{ip}/api/document/v1/{endpoint}'
    
    # 请求头
    headers = {'Authorization': f"Bearer {token}"}
    
    # 发送请求
    r = requests.request(method, requrl, verify=False, json=body, headers=headers)
    
    # Allure报告附件
    allure.attach(r.url, "url", allure.attachment_type.TEXT)
    allure.attach(str(r.status_code), "status_code", allure.attachment_type.TEXT)
    allure.attach(r.content, "content", allure.attachment_type.TEXT)
    
    # 返回响应
    if r.status_code == 204:
        return r.status_code
    else:
        return r.status_code, json.loads(r.content)

# 下面是如何使用，还是放在原lib目录下的某个py文件中
@allure.step('撤销定密审核')
def DeletePendingDetail(self, ip, token, apply_id):
    return self._make_api_request(
        method='delete',
        ip=ip,
        token=token,
        endpoint='security_classification_approval/pending_detail',
        path_param=apply_id
    )


@allure.step('设置文件密级枚举')
def SetConsoleFileClassifications(self, ip, tokenid, body):
    return self._make_api_request(
        method='PUT',
        ip=ip,
        token=tokenid,
        endpoint='console/file-classifications',
        body=body
    )


@allure.step('设置系统密级')
def SetConsoleSystemClassifications(self, ip, tokenid, body):
    return self._make_api_request(
        method='PUT',
        ip=ip,
        token=tokenid,
        endpoint='console/system-classification',
        body=body
    )
```