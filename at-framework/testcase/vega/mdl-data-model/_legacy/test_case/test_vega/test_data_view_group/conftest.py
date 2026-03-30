import pytest
import json
from src.api.vega.data_views_api import DataViewsApi

@pytest.fixture(scope="module")
def group_id():
    """创建一个测试分组并返回其ID"""
    # 创建一个测试分组，使用时间戳确保唯一性
    import time
    timestamp = int(time.time())
    resp = DataViewsApi().create_group(f"测试分组_{timestamp}")
    assert resp.status_code == 201
    
    # 提取分组ID
    group_id = json.loads(resp.text).get("id")
    assert group_id is not None
    
    yield group_id
    
    # 清理：删除创建的分组
    try:
        DataViewsApi().delete_group(group_id)
    except Exception as e:
        print(f"清理分组时出错: {e}")