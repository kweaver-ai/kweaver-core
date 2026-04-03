from datetime import datetime
from unittest import TestCase, mock

import requests
from llmadapter.llms.llm_factory import llm_factory
from sse_starlette import EventSourceResponse

from app.dao.llm_model_dao import llm_model_dao
from app.logs.stand_log import StandLogger
from app.utils import llm_utils
from app.utils.llm_utils import model_config
from app.controller import llm_controller
import asyncio
import json


class ResponseTest:
    status_code = 200
    def json(self):
        return {
            "choices": []
        }


class TestUsedModelOpenai(TestCase):
    def setUp(self) -> None:
        self.get_data_from_model_list_by_name = llm_model_dao.get_data_from_model_list_by_name
        self.OtherClient = llm_utils.OtherClient
        self.get_context_size = llm_utils.get_context_size
        self.openai_series_stream = llm_utils.openai_series_stream
        self.get_model_config = model_config.get_model_config
        self.encode = llm_utils.encode
        model_config.get_model_config = mock.Mock(return_value={"context_size": 32 * 1024})
        pass

    def tearDown(self) -> None:
        llm_model_dao.get_data_from_model_list_by_name = self.get_data_from_model_list_by_name
        llm_utils.OtherClient = self.OtherClient
        llm_utils.get_context_size = self.get_context_size
        llm_utils.openai_series_stream = self.openai_series_stream
        model_config.get_model_config = self.get_model_config
        llm_utils.encode = self.encode
        StandLogger.stand_log_shutdown()

    def test_used_model_openai_success1(self):
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        request = {
            "messages": [
                {
                    "content": "You are a helpful assistant",
                    "role": "system"
                },
                {
                    "content": "Hi",
                    "role": "user"
                }
            ],
            "model": "test_model",
            "frequency_penalty": 0,
            "max_tokens": 2048,
            "presence_penalty": 0,
            "stream": False,
            "temperature": 1,
            "top_p": 1,
            "top_k": 1,
            "response_format": {},
            "cache": False,
            "stop": None
        }
        llm_model_dao.get_data_from_model_list_by_name = mock.Mock(return_value=[{
            "f_model_id": "1234567890987654321",
            "f_model_name": "test_model",
            "f_model_series": "others",
            "f_model_type": "chat",
            "f_model_config": '{"api_key": "xxx", "api_model": "qianxun-l-128k", "api_type": "openai", "api_url": "https://qianxun.rcrai.com/open/qianxun/v1/chat/completions"}',
            "f_model": "qianxun-l-128k",
            "f_max_model_len": 32,
            "f_model_parameters": 72
        }])
        model_config.get_model_config = mock.Mock(return_value={"context_size": 4096})
        m1 = mock.MagicMock()
        m1.chat_completion = mock.AsyncMock(return_value={})
        llm_utils.OtherClient = mock.Mock(return_value=m1)
        res = loop.run_until_complete(
            llm_controller.used_model_openai(request, "111"))
        self.assertEqual({}, json.loads(res.body))

    def test_used_model_openai_success2(self):
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        request = {
            "messages": [
                {
                    "content": "You are a helpful assistant",
                    "role": "system"
                },
                {
                    "content": "Hi",
                    "role": "user"
                }
            ],
            "model": "test_model",
            "frequency_penalty": 0,
            "max_tokens": 2048,
            "presence_penalty": 0,
            "stream": True,
            "temperature": 1,
            "top_p": 1,
            "top_k": 1,
            "cache": False,
            "stop": None,
            "response_format": None
        }
        llm_model_dao.get_data_from_model_list_by_name = mock.Mock(return_value=[{
            "f_model_id": "1234567890987654321",
            "f_model_name": "test_model",
            "f_model_series": "openai",
            "f_model_type": "chat",
            "f_model_config": '{"api_key": "xxx", "api_model": "qianxun-l-128k", "api_type": "openai", "api_url": "https://qianxun.rcrai.com/open/qianxun/v1/chat/completions"}',
            "f_model": "gpt-35-turbo-16k",
            "f_max_model_len": 32,
            "f_model_parameters": 72
        }])
        m1 = mock.MagicMock()
        m1.chat_completion.return_value = {}
        llm_utils.openai_series_stream = mock.Mock(return_value={})
        res = loop.run_until_complete(
            llm_controller.used_model_openai(request, "111"))
        self.assertEqual(isinstance(res, EventSourceResponse), True)


class TestAddModel(TestCase):
    def setUp(self) -> None:
        self.add_data_into_model_list = llm_model_dao.add_data_into_model_list
        self.get_model_by_name = llm_model_dao.get_model_by_name
        self.check_model_unique = llm_model_dao.check_model_unique

    def tearDown(self) -> None:
        llm_model_dao.add_data_into_model_list = self.add_data_into_model_list
        llm_model_dao.get_model_by_name = self.get_model_by_name
        llm_model_dao.check_model_unique = self.check_model_unique
        StandLogger.stand_log_shutdown()

    def test_add_model_success(self):
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        request = {
            "model_config": {

                "api_model": "qianxun-l-128k",
                "api_url": "https://qianxun.rcrai.com/open/qianxun/v1",
                "api_key": "ckm-652a0795c43b8abca48ce7627d65e910",
                "api_type": "openai"
            },
            "model_series": "openai",
            "model_name": "qianxun-l-128k",
            "max_model_len": 128
        }
        llm_model_dao.add_data_into_model_list = mock.Mock(return_value=None)
        llm_model_dao.get_model_by_name = mock.Mock(return_value=())
        llm_model_dao.check_model_unique = mock.Mock(return_value=False)
        res = loop.run_until_complete(
            llm_controller.add_model(request, "111"))
        self.assertEqual(json.loads(res.body)["res"], True)




# test_model
class TestTestModel(TestCase):
    def setUp(self) -> None:
        self.get_data_from_model_list_by_id = llm_model_dao.get_data_from_model_list_by_id
        self.get_all_model_list = llm_model_dao.get_all_model_list
        self.create_llm = llm_factory.create_llm
        self.add_model_op_log = model_used_audit_controller.add_model_op_log
        self.encode = llm_utils.encode
        self.post = requests.post

    def tearDown(self) -> None:
        llm_model_dao.get_data_from_model_list_by_id = self.get_data_from_model_list_by_id
        llm_model_dao.get_all_model_list = self.get_all_model_list
        requests.post = self.post
        llm_factory.create_llm = self.create_llm
        model_used_audit_controller.add_model_op_log = self.add_model_op_log
        llm_utils.encode = self.encode
        StandLogger.stand_log_shutdown()

    def test_test_model_success1(self):
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        request = {
            "model_config": {
                "model_name": "qianxun-l-128k-1",
                "api_model": "qianxun-l-128k",
                "api_url": "https://qianxun.rcrai.com/open/qianxun/v1/chat/completions",
                "api_key": "ckm-652a0795c43b8abca48ce7627d65e910",
                "api_type": "openai"
            },
            "model_series": "others"
        }
        response = ResponseTest()
        requests.post = mock.Mock(return_value=response)
        res = loop.run_until_complete(
            llm_controller.test_model(request, "111"))
        self.assertEqual(json.loads(res.body)["res"]["status"], True)

    def test_test_model_success2(self):
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        request = {"model_id": "111"}
        llm_model_dao.get_all_model_list = mock.Mock(return_value=[{"f_model_id": "111"}])
        llm_model_dao.get_data_from_model_list_by_id = mock.Mock(return_value=[{"f_model_id": "111", "f_create_by": "111",
                                                                            "f_is_delete": 0, "f_model_config": '{"api_key": "111", "api_model": "gpt-4-32k", "api_url": "https://artificial-intelligence-01.openai.azure.com/"}',
                                                                            "f_model_series": "openai"}])
        m1 = mock.MagicMock()
        m1.predict.return_value = "你好"
        llm_factory.create_llm = mock.Mock(return_value=m1)
        model_used_audit_controller.add_model_op_log = mock.Mock(return_value=None)
        res = loop.run_until_complete(
            llm_controller.test_model(request, "111"))
        self.assertEqual(json.loads(res.body)["res"]["status"], True)

    def test_test_model_success3(self):
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        request = {"model_id": "111"}
        llm_model_dao.get_all_model_list = mock.Mock(return_value=[{"f_model_id": "111"}])
        llm_model_dao.get_data_from_model_list_by_id = mock.Mock(return_value=[{"f_model_id": "111", "f_create_by": "111",
                                                                            "f_is_delete": 0, "f_model_config": '{"api_key": "123", "api_model": "AIshuReader", "api_url": "http://192.168.102.250:8303/v1", "api_type": "openai"}',
                                                                            "f_model_series": "openai"}])
        m1 = mock.MagicMock()
        m1.predict.return_value = "你好"
        m1.get_num_tokens.return_value = 2
        llm_factory.create_llm = mock.Mock(return_value=m1)
        model_used_audit_controller.add_model_op_log = mock.Mock(return_value=None)
        llm_utils.encode = mock.Mock(return_value={"count": 5})
        res = loop.run_until_complete(
            llm_controller.test_model(request, "111"))
        self.assertEqual(json.loads(res.body)["res"]["status"], True)

    def test_test_model_fail1(self):
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        request = {"model_id": "111"}
        llm_model_dao.get_all_model_list = mock.Mock(return_value=[{"f_model_id": "111"}])
        llm_model_dao.get_data_from_model_list_by_id = mock.Mock(return_value=[{"f_model_id": "111", "f_create_by": "111",
                                                                            "f_is_delete": 0, "f_model_config": '{"api_key": "111", "api_model": "gpt-4-32k", "api_base": "https://artificial-intelligence-01.openai.azure.com/"}',
                                                                            "f_model_series": "aishu"}])
        m1 = mock.MagicMock()
        m1.predict.return_value = "你好"
        m1.get_num_tokens.return_value = 2
        llm_factory.create_llm = mock.Mock(return_value=m1)
        model_used_audit_controller.add_model_op_log = mock.Mock(return_value=None)
        res = loop.run_until_complete(
            llm_controller.test_model(request, "111"))
        self.assertEqual(json.loads(res.body)["code"], "ModelFactory.ModelController.TestModel.Error")


class TestEditModel(TestCase):
    def setUp(self) -> None:
        self.get_all_model_list = llm_model_dao.get_all_model_list
        self.get_data_from_model_list_by_id = llm_model_dao.get_data_from_model_list_by_id
        self.rename_model = llm_model_dao.rename_model

    def tearDown(self) -> None:
        llm_model_dao.get_all_model_list = self.get_all_model_list
        llm_model_dao.get_data_from_model_list_by_id = self.get_data_from_model_list_by_id
        llm_model_dao.rename_model = self.rename_model
        StandLogger.stand_log_shutdown()

    def test_edit_model_success(self):
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        request = {
            "model_name": "gpt-35s",
            "model_config": {
                "api_model": "gpt-4-32k",
                "api_url": "https://artificial-intelligence-01.openai.azure.com/",
                "api_type": "openai",
                "api_key": "111"
            },
            "model_series": "aishu",
            "icon": "azure",
            "model_id": "111",
            "max_model_len": 32
        }
        llm_model_dao.get_all_model_list = mock.Mock(return_value=[{"f_model_id": "111", "f_model_name": "111"}])
        llm_model_dao.get_data_from_model_list_by_id = mock.Mock(return_value=[{"f_model_id": "111", "f_create_by": "111",
                                                                            "f_is_delete": 0,
                                                                            "f_model_config": '{"api_key": "111", "api_model": "gpt-4-32k", "api_base": "https://artificial-intelligence-01.openai.azure.com/"}',
                                                                            "f_model_series": "aishu",
                                                                            "f_model_name": "111"}])
        llm_model_dao.rename_model = mock.Mock(return_value=None)
        res = loop.run_until_complete(
            llm_controller.edit_model(request, "111"))
        self.assertEqual(json.loads(res.body)["res"], True)

    def test_edit_model_fail1(self):
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        request = {
            "model_name": "gpt-35s",
            "model_config": {
                "api_model": "gpt-4-32k",
                "api_url": "https://artificial-intelligence-01.openai.azure.com/",
                "api_type": "aishu",
                "api_key": "111"
            },
            "model_series": "openai",
            "icon": "azure",
            "model_id": "111",
            "max_model_len": 32
        }
        llm_model_dao.get_all_model_list = mock.Mock(return_value=[{"f_model_id": "111", "f_model_name": "111"}])
        llm_model_dao.get_data_from_model_list_by_id = mock.Mock(return_value=[{"f_model_id": "111", "f_create_by": "111",
                                                                            "f_is_delete": 0,
                                                                            "f_model_config": '{"api_key": "111", "api_model": "gpt-4-32k", "api_base": "https://artificial-intelligence-01.openai.azure.com/"}',
                                                                            "f_model_series": "aishu",
                                                                            "f_model_name": "111"}])
        llm_model_dao.rename_model = mock.Mock(return_value=None)
        res = loop.run_until_complete(
            llm_controller.edit_model(request, "111"))
        self.assertEqual(json.loads(res.body)["code"], "ModelFactory.ConnectController.LLMEdit.ParameterError")

    def test_edit_model_fail2(self):
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        request = {
            "model_name": "gpt-35s",
            "model_config": {
                "api_model": "gpt-4-32k1",
                "api_url": "https://artificial-intelligence-01.openai.azure.com/",
                "api_key": "111"
            },
            "model_series": "aishu-m",
            "icon": "azure",
            "model_id": "111",
            "max_model_len": 32
        }
        llm_model_dao.get_all_model_list = mock.Mock(return_value=[{"f_model_id": "111", "f_model_name": "111"}])
        llm_model_dao.get_data_from_model_list_by_id = mock.Mock(return_value=[{"f_model_id": "111", "f_create_by": "111",
                                                                            "f_is_delete": 0,
                                                                            "f_model_config": '{"api_key": "111", "api_model": "gpt-4-32k", "api_base": "https://artificial-intelligence-01.openai.azure.com/"}',
                                                                            "f_model_series": "aishu",
                                                                            "f_model_name": "111"}])
        llm_model_dao.rename_model = mock.Mock(return_value=None)
        res = loop.run_until_complete(
            llm_controller.edit_model(request, "111"))
        self.assertEqual(json.loads(res.body)["code"], "ModelFactory.ConnectController.LLMEdit.ParameterError")


class TestSourceModel(TestCase):
    def setUp(self) -> None:
        self.get_data_from_model_list_by_name_fuzzy = llm_model_dao.get_data_from_model_list_by_name_fuzzy
        self.get_data_from_model_list_by_name_fuzzy_and_series = llm_model_dao.get_data_from_model_list_by_name_fuzzy_and_series

    def tearDown(self) -> None:
        llm_model_dao.get_data_from_model_list_by_name_fuzzy = self.get_data_from_model_list_by_name_fuzzy
        llm_model_dao.get_data_from_model_list_by_name_fuzzy_and_series = self.get_data_from_model_list_by_name_fuzzy_and_series
        StandLogger.stand_log_shutdown()

    def test_source_model_success1(self):
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        llm_model_dao.get_data_from_model_list_by_name_fuzzy = mock.Mock(return_value=[
            {
                "f_model_id": "111",
                "f_model_name": "111",
                "f_model_series": "aishu",
                "f_model": "AIshuReader",
                "f_model_api": "111",
                "f_create_by": "111",
                "f_update_by": "111",
                "f_create_time": datetime.today(),
                "f_update_time": datetime.today(),
                "f_icon": "aishu",
                "f_quota": 0,
                "f_max_model_len": 32,
                "f_model_parameters": 72
            }
        ])
        res = loop.run_until_complete(
            llm_controller.source_model("111", "1", "10", "", "desc", "all", "update_time", "AIshuReader", None, 0))
        self.assertEqual(json.loads(res.body)["res"]["data"][0]["model_id"], "111")

    def test_source_model_success2(self):
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        llm_model_dao.get_data_from_model_list_by_name_fuzzy_and_series = mock.Mock(return_value=[
            {
                "f_model_id": "111",
                "f_model_name": "111",
                "f_model_series": "aishu",
                "f_model": "AIshuReader",
                "f_model_api": "111",
                "f_create_by": "111",
                "f_update_by": "111",
                "f_create_time": datetime.today(),
                "f_update_time": datetime.today(),
                "f_icon": "aishu",
                "f_quota": 0,
                "f_max_model_len": 32,
                "f_model_parameters": 72
            }
        ])
        res = loop.run_until_complete(
            llm_controller.source_model("111", "1", "10", "", "desc", "aishu", "update_time", "AIshuReader", None, 0))
        self.assertEqual(json.loads(res.body)["res"]["data"][0]["model_id"], "111")


class TestCheckModel(TestCase):
    def setUp(self) -> None:
        self.get_all_model_list = llm_model_dao.get_all_model_list
        self.get_data_from_model_list_by_id = llm_model_dao.get_data_from_model_list_by_id

    def tearDown(self) -> None:
        llm_model_dao.get_all_model_list = self.get_all_model_list
        llm_model_dao.get_data_from_model_list_by_id = self.get_data_from_model_list_by_id
        StandLogger.stand_log_shutdown()

    def test_check_model_success(self):
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        llm_model_dao.get_data_from_model_list_by_id = mock.Mock(return_value=[{"f_model_id": "111", "f_create_by": "111",
                                                                            "f_is_delete": 0,
                                                                            "f_model_config": '{"api_key": "111", "api_model": "gpt-4-32k", "api_base": "https://artificial-intelligence-01.openai.azure.com/"}',
                                                                            "f_model_series": "aishu",
                                                                            "f_model_name": "111", "f_model_url": "1",
                                                                            "f_icon": "aishu",
                                                                            "f_quota": 0,
                "f_max_model_len": 32,
                "f_model_parameters": 72}])
        llm_model_dao.get_all_model_list = mock.Mock(return_value=[{"f_model_id": "111", "f_model_name": "111",
                                                                "f_model": "AIshuReader", "f_create_by": "111",
                "f_max_model_len": 32,
                "f_model_parameters": 72}])
        res = loop.run_until_complete(
            llm_controller.check_model("111", "111"))
        self.assertEqual(json.loads(res.body)["res"]["model_id"], "111")





if __name__ == '__main__':
    import unittest

    unittest.main()
