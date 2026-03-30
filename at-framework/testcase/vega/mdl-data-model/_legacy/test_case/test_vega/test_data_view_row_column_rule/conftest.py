import pytest
import json
from src.api.vega.data_views_api import DataViewsApi

@pytest.fixture(scope="module")
def rule_id():
    """创建一个测试权限规则并返回其ID"""
    # 创建一个测试权限规则
    rule_data = [{
        "name": "测试权限规则",
        "view_id": "test_view_id",
        "tags": ["test"],
        "comment": "这是一个测试权限规则",
        "fields": ["field1", "field2"]
    }]
    
    resp = DataViewsApi().create_row_column_rule(rule_data)
    assert resp.status_code == 201
    
    # 提取规则ID
    rule_id = json.loads(resp.text)[0].get("id")
    assert rule_id is not None
    
    yield rule_id
    
    # 清理：删除创建的规则
    try:
        DataViewsApi().delete_row_column_rule(rule_id)
    except Exception as e:
        print(f"清理规则时出错: {e}")