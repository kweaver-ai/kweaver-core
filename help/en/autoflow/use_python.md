# Use Python Nodes in Autoflow

## Purpose of the Node

In **Autoflow > Dataflow**, the Python node is used to implement **custom logic operations** and cover scenarios that are not supported by built-in nodes, for example:

- Calling third-party APIs such as OCR services
- Executing complex data processing such as regular-expression extraction or format conversion
- Integrating AnyShare OpenAPI to operate on documents

## Background

The DIP platform already provides a built-in Python node. If built-in nodes cannot meet business requirements, such as when you need special dependency packages or custom logic, you can extend capabilities in the following ways:

- Import custom Python dependency packages
- Write custom Python code and integrate it into a Dataflow

# Install Python Dependency Packages

Before running a Python node in a Dataflow, the required Python dependencies must be installed. The full procedure below includes API invocation and package preparation.

## Prerequisites

- You have **workflow administrator permissions** and can call the API with an admin token
- You know the Python version corresponding to the current coderunner service version

| coderunner Service Version | Compatible Python Version |
| --- | --- |
| release/6.3 and earlier | 3.8.16 |
| release/6.4 and later | 3.9.13 |

## Configure Workflow Administrator Permissions

Use the API to grant administrator permissions:

1. Call the endpoint `/api/automation/v1/admin`
2. Authenticate with an admin token carried in the request header
3. The purpose of this endpoint is to obtain workflow administrator permissions so that dependency packages can be imported later

## Upload the Dependency Package

Upload the prepared dependency package through the API:

1. Call the endpoint `/api/coderunner/v1/py-package`
2. Use the `PUT` method
3. Authenticate with a Bearer token and an administrator identity
4. Use `form/data` as the request body format

| Parameter | Type | Description | Prerequisite |
| --- | --- | --- | --- |
| file | object | Dependency package to upload | Must be in `tar` format |

## Create a Dependency Package

Because some user environments may not have external network access, offline dependency packages must be prepared manually.

### Step 1: Download the Dependency Package in a Networked Environment

1. Log in to a server with network access and open a terminal
2. Run `pip3 download [package_name] -d [target_directory]`
   - Example: `pip3 download requests -d /home/python-packages`
   - The command downloads the specified package together with all dependencies to the target directory

### Step 2: Package It as a Tar File

1. Switch to the **parent directory** of the dependency folder. For example, if the dependency folder is `/home/python-packages`, switch to `/home`
2. Run `tar -cvf [package_name].tar [dependency_directory]`
   - Example: `tar -cvf requests.tar python-packages`
3. Notes:
   - The package file name should contain only the package name, not the version number
   - Ensure the tar package contains the dependency files directly, not unnecessary multi-level directory nesting

# Use Python Nodes in a Dataflow

The following section explains configuration, code writing, and integration steps for a Python node, including example code and its operating principle.

## Prerequisites

- Basic Python coding ability
- Familiarity with the invocation rules of **AnyShare OpenAPI** if document operations are needed
- Completion of required dependency installation

## Built-in Libraries

The Python node includes the following libraries by default and does not require extra installation:

- `aishu_anyshare_api`
- `requests`
- `json`
- `re`
- `base64`
- `datetime`

## Example: Python Code for Calling an OCR API

The following code implements a complete process of **getting document content -> calling an OCR API -> extracting the OCR result**. It can be adjusted according to business requirements.

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
    Main function: call the OCR API and extract the result
    Parameters:
        doc_id: str - file ID from the Dataflow context
    Returns:
        texts: list - list of OCR-recognized text, or an error message if execution fails
    """
    client = ApiClient(verify_ssl=False)
    try:
        # Step 1: download the file through AnyShare OpenAPI
        json_file = file_download(client, doc_id)
        text = json_file.content
        
        # Step 2: call the OCR API
        ocr_url = 'http://10.4.132.197:8507/lab/ocr/predict/general'
        b64_text = base64.b64encode(text).decode()
        request_data = {'scene': 'chinese_print', 'image': b64_text}
        ocr_response = requests.post(ocr_url, json=request_data).json()
        
        # Step 3: extract the OCR result
        texts = ocr_response['data']['json']['general_ocr_res']['texts']
        return texts
    
    except Exception as e:
        return str(e)

def file_download(client: ApiClient, doc_id: str) -> requests.Response:
    """
    Helper function: download a file through AnyShare OpenAPI
    """
    download_resp = client.efast.efast_v1_file_osdownload_post(
        efast.FileOsdownloadReq.from_dict({\"docid\": doc_id})
    )
    method, url, headers = download_resp.authrequest[0], download_resp.authrequest[1], download_resp.authrequest[2:]
    header_dict = {}
    for header in headers:
        key, value = header.split(\": \")
        header_dict[key] = value.strip()
    file_resp = requests.request(method, url, headers=header_dict, verify=False)
    return file_resp
```

### Code Explanation and Principles

#### Functional Explanation

1. **File download**: use `file_download` to call AnyShare OpenAPI and get the file content by `doc_id`
2. **OCR invocation**: Base64-encode the file content and send it to the OCR API
3. **Result extraction**: extract the `texts` field from the OCR API response JSON
4. **Exception handling**: catch errors such as API call failures or format parsing issues and return an error message

#### Dataflow Integration Principle

The integration between the Python node and Autoflow relies on the **context** mechanism:

1. **Trigger condition**: when the Dataflow is triggered by a file upload event, the system automatically gets the file `doc_id` and passes it into the Python node
2. **Input variables**: the input variables of the Python node, such as `doc_id`, must be bound to context variables in the Dataflow
3. **Output variables**: the output variables correspond one-to-one with the return value of the `main` function. The returned value is stored in the Dataflow context and can be used by downstream nodes

#### Configuration Notes

1. **Input variable configuration**: configure an input variable named `doc_id` for the Python node and bind it to the `docid` field of the file upload event
2. **Output variable configuration**: configure an output variable with a type that matches the return type of `main`, such as `list`
3. **SSL verification**: the example sets `verify_ssl=False` for testing environments. In production, SSL verification should be enabled
