from unittest import TestCase, mock
from app.dao.small_model_dao import external_small_model_dao
from app.interfaces.dbaccess import AddExternalSmallModelInfo
from app.mydb.pymysql_pool import PymysqlPool
from app.utils.stand_log import StandLogger


class TestAddModelInfo(TestCase):
    def setUp(self) -> None:
        self.mysqlPool = PymysqlPool

    def tearDown(self) -> None:
        PymysqlPool = self.mysqlPool
        StandLogger.stand_log_shutdown()

    def test_add_model_info_success(self):
        m1 = mock.MagicMock()
        m2 = mock.MagicMock()
        m1.connection.return_value = m2
        m3 = mock.MagicMock()
        m3.execute.return_value = True
        m3.fetchall.return_value = "test"
        m2.cursor.return_value = m3
        m3.lastrowid = 1
        PymysqlPool.get_pool = mock.Mock(return_value=m1)
        config_info = AddExternalSmallModelInfo(
            config_id="1",
            model_name="1",
            model_series="1",
            model_type="1",
            model_config={}
        )
        res = external_small_model_dao.add_model_info(config_info)
        self.assertEqual(res, None)

class TestEditModelInfo(TestCase):
    def setUp(self) -> None:
        self.mysqlPool = PymysqlPool

    def tearDown(self) -> None:
        PymysqlPool = self.mysqlPool
        StandLogger.stand_log_shutdown()

    def test_edit_model_info_success(self):
        m1 = mock.MagicMock()
        m2 = mock.MagicMock()
        m1.connection.return_value = m2
        m3 = mock.MagicMock()
        m3.execute.return_value = True
        m3.fetchall.return_value = "test"
        m2.cursor.return_value = m3
        m3.lastrowid = 1
        PymysqlPool.get_pool = mock.Mock(return_value=m1)
        config_info = AddExternalSmallModelInfo(
            config_id="1",
            model_name="1",
            model_series="1",
            model_type="1",
            model_config={}
        )
        res = external_small_model_dao.edit_model_info(config_info)
        self.assertEqual(res, None)


class TestGetModelInfoById(TestCase):
    def setUp(self) -> None:
        self.mysqlPool = PymysqlPool

    def tearDown(self) -> None:
        PymysqlPool = self.mysqlPool
        StandLogger.stand_log_shutdown()

    def test_get_model_info_by_id_success(self):
        m1 = mock.MagicMock()
        m2 = mock.MagicMock()
        m1.connection.return_value = m2
        m3 = mock.MagicMock()
        m3.execute.return_value = True
        m3.fetchall.return_value = "test"
        m2.cursor.return_value = m3
        m3.lastrowid = 1
        PymysqlPool.get_pool = mock.Mock(return_value=m1)
        res = external_small_model_dao.get_model_info_by_id("!")
        self.assertEqual(res, "test")


class TestGetModelInfoByName(TestCase):
    def setUp(self) -> None:
        self.mysqlPool = PymysqlPool

    def tearDown(self) -> None:
        PymysqlPool = self.mysqlPool
        StandLogger.stand_log_shutdown()

    def test_get_model_info_by_name_success(self):
        m1 = mock.MagicMock()
        m2 = mock.MagicMock()
        m1.connection.return_value = m2
        m3 = mock.MagicMock()
        m3.execute.return_value = True
        m3.fetchall.return_value = "test"
        m2.cursor.return_value = m3
        m3.lastrowid = 1
        PymysqlPool.get_pool = mock.Mock(return_value=m1)
        res = external_small_model_dao.get_model_info_by_name("!")
        self.assertEqual(res, "test")


class TestGetModelInfoList(TestCase):
    def setUp(self) -> None:
        self.mysqlPool = PymysqlPool

    def tearDown(self) -> None:
        PymysqlPool = self.mysqlPool
        StandLogger.stand_log_shutdown()

    def test_get_model_info_list_success(self):
        m1 = mock.MagicMock()
        m2 = mock.MagicMock()
        m1.connection.return_value = m2
        m3 = mock.MagicMock()
        m3.execute.return_value = True
        m3.fetchall.return_value = "test"
        m2.cursor.return_value = m3
        m3.lastrowid = 1
        PymysqlPool.get_pool = mock.Mock(return_value=m1)
        res = external_small_model_dao.get_model_info_list(1, 10, "asc", "update_time", "1", "baidu", "embedding")
        self.assertEqual(res, "test")


class TestDeleteModelInfoById(TestCase):
    def setUp(self) -> None:
        self.mysqlPool = PymysqlPool

    def tearDown(self) -> None:
        PymysqlPool = self.mysqlPool
        StandLogger.stand_log_shutdown()

    def test_delete_model_info_by_id_success(self):
        m1 = mock.MagicMock()
        m2 = mock.MagicMock()
        m1.connection.return_value = m2
        m3 = mock.MagicMock()
        m3.execute.return_value = True
        m3.fetchall.return_value = "test"
        m2.cursor.return_value = m3
        m3.lastrowid = 1
        PymysqlPool.get_pool = mock.Mock(return_value=m1)
        res = external_small_model_dao.delete_model_info_by_id("1")
        self.assertEqual(res, None)


class TestNameCheck(TestCase):
    def setUp(self) -> None:
        self.mysqlPool = PymysqlPool

    def tearDown(self) -> None:
        PymysqlPool = self.mysqlPool
        StandLogger.stand_log_shutdown()

    def test_name_check_success(self):
        m1 = mock.MagicMock()
        m2 = mock.MagicMock()
        m1.connection.return_value = m2
        m3 = mock.MagicMock()
        m3.execute.return_value = True
        m3.fetchall.return_value = "test"
        m2.cursor.return_value = m3
        m3.lastrowid = 1
        PymysqlPool.get_pool = mock.Mock(return_value=m1)
        res = external_small_model_dao.name_check("1")
        self.assertEqual(res, "test")


if __name__ == '__main__':
    import unittest

    unittest.main()
