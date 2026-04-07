# 在 Autoflow 中使用 Python 节点

## 节点作用
在 「Autoflow」>「数据流」中，Python 节点用于实现**自定义逻辑操作**，补充当前数据流内置节点未覆盖的功能场景，例如：
- 调用第三方 API（如 OCR 接口）
- 执行复杂数据处理（如正则提取、格式转换）
- 集成 AnyShare OpenAPI 实现文档操作

## 背景说明
DIP 平台已提供内置 Python 节点，若内置节点无法满足业务需求（如需特殊依赖包、自定义代码逻辑），用户可通过以下方式扩展：
- 导入自定义 Python 依赖包
- 编写自定义 Python 代码并集成至数据流


# 安装 Python 依赖包
在数据流中运行 Python 节点前，需先安装所需 Python 依赖包。以下为完整安装流程，包含接口调用、安装包制作步骤。

## 前提条件
- 拥有**流程管理员权限**（需使用 admin token 调用接口）
- 已明确当前 coderunner 服务版本对应的 Python 版本：

| coderunner 服务版本 | 兼容 Python 版本 |
| --- | --- |
| release/6.3 及以前 | 3.8.16 |
| release/6.4 及以后 | 3.9.13 |

## 配置流程管理员
通过接口配置管理员权限，步骤如下：
1. 调用接口：`/api/automation/v1/admin`
2. 认证方式：使用 admin token（需在请求头中携带）
3. 接口作用：获取流程管理员操作权限，为后续导入依赖包做准备

## 上传依赖安装包
通过接口上传已制作好的依赖包，步骤如下：
1. 调用接口：`/api/coderunner/v1/py-package`
2. 请求方法：PUT
3. 认证方式：Bearer token（需管理员身份）
4. 请求体格式（form/data）：

| 参数名 | 类型 | 说明 | 前提条件 |
| --- | --- | --- | --- |
| file | object | 待上传的依赖安装包 | 需为 tar 格式 |

## 制作依赖安装包
考虑到用户环境可能存在**无网络场景**，需手动制作离线依赖包，步骤如下：

### 步骤 1：下载依赖包（需有网络环境）
1. 登录有网络的服务器，打开终端
2. 执行命令：`pip3 download [包名] -d [目标目录]`
   - 示例：下载 requests 包至 `/home/python-packages` 目录，命令为：`pip3 download requests -d /home/python-packages`
   - 说明：命令会自动下载指定包及所有依赖包至目标目录

### 步骤 2：打包为 tar 格式
1. 切换至依赖包所在目录的**上层目录**（示例：若依赖包在 `/home/python-packages`，则切换至 `/home`）
2. 执行打包命令：`tar -cvf [包名].tar [依赖包目录]`
   - 示例：打包 requests 依赖包，命令为：`tar -cvf requests.tar python-packages`
3. 注意事项：
   - 打包文件名仅包含名称，不携带版本（如 `requests.tar`，而非 `requests-2.31.0.tar`）
   - 确保 tar 包内直接包含依赖包文件（而非多层目录嵌套）


# 在数据流中使用 Python 节点
以下为 Python 节点的配置、代码编写及集成步骤，包含示例代码与原理说明。

## 前提条件
- 拥有**基础 Python 编码能力**
- 了解**AnyShare OpenAPI**的调用规则（若需集成文档操作）
- 已完成所需依赖包的安装（参考前述章节）

## 内置库说明
Python 节点默认包含以下内置库，无需额外安装：
- aishu_anyshare_api（AnyShare OpenAPI 封装 SDK）
- requests（HTTP 请求库）
- json（JSON 格式处理）
- re（正则表达式）
- base64（Base64 编码解码）
- datetime（时间处理）

## 示例：调用 OCR 接口的 Python 代码
以下代码实现 “获取文档内容→调用 OCR 接口→提取 OCR 结果” 的完整逻辑，可根据业务需求调整。

```Plain
from aishu_anyshare_api.api_client import ApiClient
import aishu_anyshare_api_efast as efast
from datetime import datetime
import requests
import json
import re
import base64

def main(doc_id):
    """
    主函数：实现OCR接口调用与结果提取
    参数：
        doc_id: str - 触发流程的文件ID（来自数据流上下文）
    返回：
        texts: list - OCR识别后的文本列表（若失败则返回错误信息）
    """
    client = ApiClient(verify_ssl=False)  # 初始化AnyShare API客户端（禁用SSL验证）
    try:
        # 步骤1：下载文件（调用AnyShare OpenAPI）
        json_file = file_download(client, doc_id)
        text = json_file.content  # 获取文件二进制内容
        
        # 步骤2：调用OCR接口
        ocr_url = 'http://10.4.132.197:8507/lab/ocr/predict/general'  # OCR接口地址
        b64_text = base64.b64encode(text).decode()  # 对文件内容进行Base64编码
        request_data = {'scene': 'chinese_print', 'image': b64_text}  # 构造OCR请求参数
        ocr_response = requests.post(ocr_url, json=request_data).json()  # 发送POST请求
        
        # 步骤3：提取OCR结果（根据OCR模型返回格式调整）
        texts = ocr_response['data']['json']['general_ocr_res']['texts']
        return texts  # 返回结果至数据流上下文
    
    except Exception as e:
        return str(e)  # 捕获异常并返回错误信息

def file_download(client: ApiClient, doc_id: str) -> requests.Response:
    """
    辅助函数：通过AnyShare OpenAPI下载文件
    参数：
        client: ApiClient - AnyShare API客户端实例
        doc_id: str - 文件ID
    返回：
        requests.Response - 文件下载响应对象
    """
    # 调用AnyShare文件下载接口
    download_resp = client.efast.efast_v1_file_osdownload_post(
        efast.FileOsdownloadReq.from_dict({\"docid\": doc_id})
    )
    # 解析接口返回的请求信息（方法、URL、请求头）
    method, url, headers = download_resp.authrequest[0], download_resp.authrequest[1], download_resp.authrequest[2:]
    # 构造请求头字典
    header_dict = {}
    for header in headers:
        key, value = header.split(\": \")
        header_dict[key] = value.strip()
    # 发送文件下载请求
    file_resp = requests.request(method, url, headers=header_dict, verify=False)
    return file_resp
```

### 代码说明与原理
#### 代码功能说明
1. **文件下载**：通过 `file_download` 函数调用 AnyShare OpenAPI，根据 `doc_id` 获取文件内容
2. **OCR 调用**：将文件内容 Base64 编码后，发送至 OCR 接口获取识别结果
3. **结果提取**：从 OCR 接口返回的 JSON 数据中，提取 `texts` 字段（文本识别结果）
4. **异常处理**：捕获代码执行过程中的错误（如接口调用失败、格式解析错误），并返回错误信息

#### 数据流集成原理
Python 节点与 Autoflow 数据流的集成依赖**上下文（context）** 机制，核心规则如下：
1. **触发条件**：数据流在 “文件上传” 事件触发时，自动获取文件 `doc_id` 并传入 Python 节点
2. **输入变量**：Python 节点的输入变量（如示例中的 `doc_id`）需与数据流上下文变量绑定，值由数据流自动传入
3. **输出变量**：Python 节点的输出变量与 `main` 函数的返回值一一对应，返回结果会存入数据流上下文，供后续节点使用（如编目、标签设置）

#### 配置注意事项
1. **输入变量配置**：需在 Autoflow 数据流中为 Python 节点设置输入变量 `doc_id`，并绑定 “文件上传” 事件的 `docid` 字段
2. **输出变量配置**：根据 `main` 函数的返回值类型（如示例中的 `list`），在数据流中配置对应类型的输出变量
3. **SSL 验证**：示例中 `verify_ssl=False` 用于禁用 SSL 证书验证（适用于测试环境），生产环境需启用 SSL 验证（删除该参数或设为 `True`）
