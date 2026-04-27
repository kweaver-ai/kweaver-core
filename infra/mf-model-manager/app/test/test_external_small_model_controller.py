from datetime import datetime
from unittest import TestCase, mock

from app.dao.small_model_dao import small_model_dao
from app.interfaces import logics
from app.logs.stand_log import StandLogger
import asyncio
import json


class TestAddModel(TestCase):
    def setUp(self) -> None:
        self.name_check = small_model_dao.name_check
        self.add_model_info = small_model_dao.add_model_info

    def tearDown(self) -> None:
        small_model_dao.name_check = self.name_check
        small_model_dao.add_model_info = self.add_model_info
        StandLogger.stand_log_shutdown()

    def test_add_model_success(self):
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        small_model_dao.name_check = mock.Mock(return_value=[])
        small_model_dao.add_model_info = mock.Mock(return_value=None)
        para = logics.AddExternalSmallModel(
            model_name="1",
            model_series="baidu",
            model_type="embedding",
            model_config={}
        )
        res = loop.run_until_complete(
            external_small_model_controller.add_model(para, "1"))
        self.assertEqual(isinstance(json.loads(res.body)["id"], str), True)


class TestEditModel(TestCase):
    def setUp(self) -> None:
        self.name_check = small_model_dao.name_check
        self.edit_model_info = small_model_dao.edit_model_info

    def tearDown(self) -> None:
        small_model_dao.name_check = self.name_check
        small_model_dao.edit_model_info = self.edit_model_info
        StandLogger.stand_log_shutdown()

    def test_edit_model_success(self):
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        small_model_dao.name_check = mock.Mock(return_value=[])
        small_model_dao.edit_model_info = mock.Mock(return_value=None)
        para = logics.EditExternalSmallModel(
            config_id="1234567890987654321",
            model_name="1",
            model_series="baidu",
            model_type="embedding",
            model_config={}
        )
        res = loop.run_until_complete(
            external_small_model_controller.edit_model(para, "1"))
        self.assertEqual(isinstance(json.loads(res.body)["id"], str), True)


class TestGetInfoList(TestCase):
    def setUp(self) -> None:
        self.get_model_info_list = small_model_dao.get_model_info_list

    def tearDown(self) -> None:
        small_model_dao.get_model_info_list = self.get_model_info_list
        StandLogger.stand_log_shutdown()

    def test_get_info_list_success(self):
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        small_model_dao.get_model_info_list = mock.Mock(return_value=[])
        res = loop.run_until_complete(
            external_small_model_controller.get_info_list("asc", "update_time", 1, 10, "1", "baidu", "embedding", "1"))
        self.assertEqual(isinstance(json.loads(res.body)["res"], list), True)


class TestGetInfo(TestCase):
    def setUp(self) -> None:
        self.get_model_info_by_id = small_model_dao.get_model_info_by_id

    def tearDown(self) -> None:
        small_model_dao.get_model_info_by_id = self.get_model_info_by_id
        StandLogger.stand_log_shutdown()

    def test_get_info_success(self):
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        small_model_dao.get_model_info_by_id = mock.Mock(return_value=[
            {
                "f_model_id": "1",
                "f_model_name": "1",
                "f_series": "baidu",
                "f_model_type": "embedding",
                "f_model_config": "{}",
                "f_create_time": datetime.today(),
                "f_update_time": datetime.today()
            }
        ])
        res = loop.run_until_complete(
            external_small_model_controller.get_info("asc", "update_time"))
        self.assertEqual(isinstance(json.loads(res.body)["res"], dict), True)


class TestDeleteModel(TestCase):
    def setUp(self) -> None:
        self.delete_model_info_by_id = small_model_dao.delete_model_info_by_id

    def tearDown(self) -> None:
        small_model_dao.delete_model_info_by_id = self.delete_model_info_by_id
        StandLogger.stand_log_shutdown()

    def test_delete_model_success(self):
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        small_model_dao.delete_model_info_by_id = mock.Mock(return_value=[])
        res = loop.run_until_complete(
            external_small_model_controller.delete_model("asc", "update_time"))
        self.assertEqual(isinstance(json.loads(res.body)["res"], str), True)


if __name__ == '__main__':
    import unittest

    unittest.main()
