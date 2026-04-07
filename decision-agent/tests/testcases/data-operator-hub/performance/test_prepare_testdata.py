# -*- coding:UTF-8 -*-

from datetime import datetime
import random
import string

from common.get_content import GetContent
from common.get_token import GetToken
from lib.operator import Operator
from lib.tool_box import ToolBox
from lib.mcp import MCP

op_client = Operator()
tb_client = ToolBox()
mcp_client = MCP()

def prepare_testdata():
    """准备测试数据：注册并发布算子/工具/mcp"""
    filepath = "./resource/openapi/compliant/test0.json"
    api_data = GetContent(filepath).jsonfile()
    # 定义所有可能的category
    categories = [
        "other_category",
        "data_process",
        "data_transform",
        "data_store",
        "data_analysis",
        "data_query",
        "data_extract",
        "data_split",
        "model_train"
    ]
    for i in range(1,101):
        configfile = "./config/env.ini"
        file = GetContent(configfile)
        config = file.config()
        host = config["server"]["host"]
        token = GetToken(host=host).get_token(host, str(i), "111111")
        headers = {
            "Authorization": f"Bearer {token[1]}"
        }
        # 注册1000个算子（每次10个，共100次）
        for i in range(100):
            # 修改每个路径下的summary字段避免重名
            for path in api_data["paths"]:
                for method in api_data["paths"][path]:
                    if "summary" in api_data["paths"][path][method]:
                        now = datetime.now()
                        timestamp_seconds = now.timestamp()
                        timestamp_millis = str(int(timestamp_seconds * 1000))
                        name = ''.join(random.choice(string.ascii_letters+string.digits) for i in range(8))
                        api_data["paths"][path][method]["summary"] = f"Test_operator_{name}_{timestamp_millis}"
            
            # 设置category（每次注册10个算子，确保均匀分布）
            current_category = categories[i % len(categories)]          
            
            data = {
                "data": str(api_data),
                "operator_metadata_type": "openapi",
                "operator_info": {
                    "category": current_category
                }
            }
            result = op_client.RegisterOperator(data, headers)
            assert result[0] == 200
            operators = result[1]            
            # 处理每个算子
            for operator in operators:
                if operator["status"] == "success":       
                    update_data = [{
                        "operator_id": operator["operator_id"],
                        "version": operator["version"],
                        "status": "published"
                    }]
                    re = op_client.UpdateOperatorStatus(update_data, headers)
                    assert re[0] == 200 
        # 创建1000个工具箱并发布
        for i in range(1000):
            now = datetime.now()
            timestamp_seconds = now.timestamp()
            timestamp_millis = str(int(timestamp_seconds * 1000))
            name = "toolbox_" + ''.join(random.choice(string.ascii_letters+string.digits) for i in range(8)) + timestamp_millis
            data = {
                "box_name": name,
                "data": api_data,
                "metadata_type": "openapi"
            }
            result = tb_client.CreateToolbox(data, headers)
            assert result[0] == 200
            box_id = result[1]["box_id"]
            # 发布工具箱
            update_data = {
                "status": "published"
            }
            re = tb_client.UpdateToolboxStatus(box_id, update_data, headers)
            assert re[0] == 200
        # 注册1000个mcp并发布
        for i in range(1000):
            now = datetime.now()
            timestamp_seconds = now.timestamp()
            timestamp_millis = str(int(timestamp_seconds * 1000))
            name = "mcp_" + ''.join(random.choice(string.ascii_letters+string.digits) for i in range(8)) + timestamp_millis
            data = {
                "name": name,
                "description": "test mcp server",
                "mode": "sse",
                "url": "https://mcp.map.baidu.com/sse?ak=bW9A9vyhGcYmdKRvWJCkySpekiBUTeUL",
                "source": "custom",
                "category": "data_analysis"
            }
            result = mcp_client.RegisterMCP(data, headers)
            assert result[0] == 200
            mcp_id = result[1]["mcp_id"]
            update_data = {
                "status": "published"
            }
            result = mcp_client.MCPReleaseAction(mcp_id, update_data, headers)
            assert result[0] == 200

if __name__ == '__main__':
    prepare_testdata()