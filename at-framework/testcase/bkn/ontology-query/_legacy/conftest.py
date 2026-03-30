"""
pytest配置文件，提供通用的API客户端工厂和模块级前置条件
"""

import pytest
import json
from src.config.setting import config
from src.common.global_var import global_vars
from src.common.utils import filter_tables_by_name
from src.api.vega.data_views_api import DataViewsApi
from src.api.agent.agetn_operator import AgentOperatorApi
from src.api.ontology.kn_network_api import KnowledgeNetworkApi
from src.api.ontology.kn_group_api import ConceptGroupApi
from src.common.logger import logger
from src.common.data_gen import random_str


@pytest.fixture(scope="session")
def update_kn_id():
    """
    对象类型测试模块的自动前置条件
    """
    from src.config.setting import config

    old_kn_id = config.get("view_data", "kn_id")
    prefix = "auto_test_"
    suffix_str = old_kn_id[len(prefix) :]
    suffix_num = int(suffix_str) + 1
    new_suffix = str(suffix_num).zfill(len(suffix_str))
    new_kn_id = prefix + new_suffix
    config.set("view_data", "kn_id", new_kn_id)
    config.save()
    kn_api = KnowledgeNetworkApi()
    kn_api.create_kn(new_kn_id, new_kn_id)
    logger.info(f"测试模块初始化完成，使用KN: {new_kn_id}")
    yield new_kn_id


@pytest.fixture(scope="class")
def test_kn_id_use():
    """
    对象类型测试模块的自动前置条件
    自动使用ensure_test_kn fixture确保KN存在
    """
    kn_id = f"test_kn_{random_str()}"
    kn_api = KnowledgeNetworkApi()
    kn_api.create_kn(kn_id, kn_id)
    logger.info(f"测试模块初始化完成，使用KN: {kn_id}")
    yield kn_id
    logger.info(f"测试模块完成，删除KN: {kn_id}")
    kn_api.delete_kn(kn_id)


@pytest.fixture(scope="session")
def get_views_id():
    """检查及获取数据视图ID"""
    view_api = DataViewsApi()
    resp = view_api.list_views(data_source_id=config.get("view_data", "data_source_id"))
    views = json.loads(resp.text)["entries"]
    objects_dict = [
        "organisation",
        "person",
        "post",
        "comment",
        "place",
        "tag",
        "forum",
        "tagclass",
    ]
    relations_dict = [
        "comment_hascreator_person",
        "comment_hastag_tag",
        "comment_islocatedin_place",
        "comment_replyof_comment",
        "comment_replyof_post",
        "forum_containerof_post",
        "forum_hasmember_person",
        "forum_hasmoderator_person",
        "forum_hastag_tag",
        "organisation_islocatedin_place",
        "person_hasinterest_tag",
        "person_islocatedin_place",
        "person_knows_person",
        "person_likes_comment",
        "person_likes_post",
        "person_speaks_language",
        "person_studyat_organisation",
        "person_workat_organisation",
        "place_ispartof_place",
        "post_hascreator_person",
        "post_hastag_tag",
        "post_islocatedin_place",
        "tag_hastype_tagclass",
        "tagclass_issubclassof_tagclass",
    ]
    objects_list = filter_tables_by_name(views, objects_dict)
    relations_list = filter_tables_by_name(views, relations_dict)
    global_vars.set_var("objects_list", objects_list)
    global_vars.set_var("relations_list", relations_list)
    logger.info(f"ldbc视图初始化完成")
    return objects_list, relations_list


@pytest.fixture(scope="session")
def get_tool_id_parameters():
    """
    行动类类型测试模块的前置条件
    """
    agent_api = AgentOperatorApi()

    boxs_resp = agent_api.list_tool_boxs_market()
    box = json.loads(boxs_resp.text)["data"]
    box_id = next(
        (item["box_id"] for item in box if item["box_name"] == "联网搜索添加引用工具"),
        None,
    )

    tool_list_resp = agent_api.list_tool(box_id)
    tool_list = json.loads(tool_list_resp.text)["tools"]
    tool_id = next(
        (
            item["tool_id"]
            for item in tool_list
            if item["name"] == "online_search_cite_tool"
        ),
        None,
    )

    parameters = [
        {"name": "api_key", "value_from": "input", "type": "string", "source": "Body"},
        {
            "name": "model_name",
            "value_from": "input",
            "type": "string",
            "source": "Body",
        },
        {"name": "query", "value_from": "input", "type": "string", "source": "Body"},
        {
            "name": "search_tool",
            "value_from": "input",
            "type": "string",
            "source": "Body",
        },
        {"name": "stream", "value_from": "input", "type": "boolean", "source": "Body"},
        {"name": "user_id", "value_from": "input", "type": "string", "source": "Body"},
    ]
    global_vars.set_var("box_id", box_id)
    global_vars.set_var("tool_id", tool_id)
    global_vars.set_var("parameters", parameters)


@pytest.fixture(scope="session")
def create_test_concept_group_add_objects():
    api = ConceptGroupApi()
    api.create_concept_groups(
        kn_id=config.get("view_data", "kn_id"),
        groups_id="test_add",
        groups_name="测试添加对象类",
    )
    yield
    api.delete_concept_groups(
        kn_id=config.get("view_data", "kn_id"),
        groups_id="test_add",
    )


@pytest.fixture(scope="session")
def create_test_concept_group_remove_objects():
    api = ConceptGroupApi()

    api.create_concept_groups(
        kn_id=config.get("view_data", "kn_id"),
        groups_id="test_remove",
        groups_name="测试删除对象类",
    )
    api.concept_groups_add_object(
        kn_id=config.get("view_data", "kn_id"),
        groups_id="test_remove",
        objects=["person", "comment", "post", "tag", "forum", "tagclass"],
    )
    yield
    api.delete_concept_groups(
        kn_id=config.get("view_data", "kn_id"),
        groups_id="test_remove",
    )
