from unittest import TestCase, mock

from app.dao import user_info
from app.dao.llm_model_dao import llm_model_dao
from app.logs.stand_log import StandLogger
from app.utils import llm_utils
from app.mydb.pymysql_pool import PymysqlPool


class TestGetModelNameById(TestCase):
    def setUp(self) -> None:
        self.mysqlPool = PymysqlPool

    def tearDown(self) -> None:
        PymysqlPool = self.mysqlPool
        StandLogger.stand_log_shutdown()

    def test_get_model_name_by_id_success(self):
        m1 = mock.MagicMock()
        m2 = mock.MagicMock()
        m1.connection.return_value = m2
        m3 = mock.MagicMock()
        m3.fetchall.return_value = [{"f_model_name": "name"}]
        m2.cursor.return_value = m3
        # 为PymysqlPool.get_pool函数创建mock对象
        PymysqlPool.get_pool = mock.Mock(return_value=m1)
        res = llm_model_dao.get_model_name_by_id("222")
        self.assertEqual("name", res)


# get_model_id_by_name函数的测试类
class TestGetModelIdByName(TestCase):
    def setUp(self) -> None:
        self.mysqlPool = PymysqlPool

    def tearDown(self) -> None:
        PymysqlPool = self.mysqlPool
        StandLogger.stand_log_shutdown()

    def test_get_model_id_by_name_success(self):
        m1 = mock.MagicMock()
        m2 = mock.MagicMock()
        m1.connection.return_value = m2
        m3 = mock.MagicMock()
        m3.fetchall.return_value = [{"f_model_id": "222"}]
        m2.cursor.return_value = m3
        # 为PymysqlPool.get_pool函数创建mock对象
        PymysqlPool.get_pool = mock.Mock(return_value=m1)
        res = llm_model_dao.get_model_id_by_name("name")
        self.assertEqual([{"f_model_id": "222"}], res)


# get_model_series_by_id函数的测试类
class TestGetModelSeriesByID(TestCase):
    def setUp(self) -> None:
        self.mysqlPool = PymysqlPool

    def tearDown(self) -> None:
        PymysqlPool = self.mysqlPool
        StandLogger.stand_log_shutdown()

    def test_get_model_series_by_id_success(self):
        m1 = mock.MagicMock()
        m2 = mock.MagicMock()
        m1.connection.return_value = m2
        m3 = mock.MagicMock()
        m3.fetchall.return_value = [{"f_model_series": "series"}]
        m2.cursor.return_value = m3
        # 为PymysqlPool.get_pool函数创建mock对象
        PymysqlPool.get_pool = mock.Mock(return_value=m1)
        res = llm_model_dao.get_model_series_by_id("222")
        self.assertEqual("series", res)


# get_model_model_from_model_list_by_id函数的测试类
class TestGetModelModelFromModelListByID(TestCase):
    def setUp(self) -> None:
        self.mysqlPool = PymysqlPool

    def tearDown(self) -> None:
        PymysqlPool = self.mysqlPool
        StandLogger.stand_log_shutdown()

    def test_get_model_model_from_model_list_by_id_success(self):
        m1 = mock.MagicMock()
        m2 = mock.MagicMock()
        m1.connection.return_value = m2
        m3 = mock.MagicMock()
        m3.fetchall.return_value = [{"f_model": "222"}]
        m2.cursor.return_value = m3
        # 为PymysqlPool.get_pool函数创建mock对象
        PymysqlPool.get_pool = mock.Mock(return_value=m1)
        res = llm_model_dao.get_model_model_from_model_list_by_id("222")
        self.assertEqual("222", res)


# get_all_model_list函数的测试类
class TestGetAllModelList(TestCase):
    def setUp(self) -> None:
        self.mysqlPool = PymysqlPool

    def tearDown(self) -> None:
        PymysqlPool = self.mysqlPool
        StandLogger.stand_log_shutdown()

    def test_get_all_model_list_success(self):
        m1 = mock.MagicMock()
        m2 = mock.MagicMock()
        m1.connection.return_value = m2
        m3 = mock.MagicMock()
        m3.fetchall.return_value = [{"f_model_id": "222"}]
        m2.cursor.return_value = m3
        # 为PymysqlPool.get_pool函数创建mock对象
        PymysqlPool.get_pool = mock.Mock(return_value=m1)
        res = llm_model_dao.get_all_model_list()
        self.assertEqual([{"f_model_id": "222"}], res)


# get_data_from_model_list_by_id函数的测试类
class TestGetDataFromModelListByID(TestCase):
    def setUp(self) -> None:
        self.mysqlPool = PymysqlPool

    def tearDown(self) -> None:
        PymysqlPool = self.mysqlPool
        StandLogger.stand_log_shutdown()

    def test_get_data_from_model_list_by_id_success(self):
        m1 = mock.MagicMock()
        m2 = mock.MagicMock()
        m1.connection.return_value = m2
        m3 = mock.MagicMock()
        m3.fetchall.return_value = [{"f_model_id": "222"}]
        m2.cursor.return_value = m3
        # 为PymysqlPool.get_pool函数创建mock对象
        PymysqlPool.get_pool = mock.Mock(return_value=m1)
        res = llm_model_dao.get_data_from_model_list_by_id("222")
        self.assertEqual([{"f_model_id": "222"}], res)


# get_data_from_model_list_by_name函数的测试类
class TestGetDataFromModelListByName(TestCase):
    def setUp(self) -> None:
        self.mysqlPool = PymysqlPool

    def tearDown(self) -> None:
        PymysqlPool = self.mysqlPool
        StandLogger.stand_log_shutdown()

    def test_get_data_from_model_list_by_name_success(self):
        m1 = mock.MagicMock()
        m2 = mock.MagicMock()
        m1.connection.return_value = m2
        m3 = mock.MagicMock()
        m3.fetchall.return_value = [{"f_model_id": "222"}]
        m2.cursor.return_value = m3
        # 为PymysqlPool.get_pool函数创建mock对象
        PymysqlPool.get_pool = mock.Mock(return_value=m1)
        res = llm_model_dao.get_data_from_model_list_by_name("222")
        self.assertEqual([{"f_model_id": "222"}], res)


# add_data_into_model_list函数的测试类
class TestAddDataIntoModelList(TestCase):
    def setUp(self) -> None:
        self.mysqlPool = PymysqlPool

    def tearDown(self) -> None:
        PymysqlPool = self.mysqlPool
        StandLogger.stand_log_shutdown()

    def test_add_data_into_model_list_success(self):
        m1 = mock.MagicMock()
        m2 = mock.MagicMock()
        m1.connection.return_value = m2
        m3 = mock.MagicMock()
        m3.fetchall.return_value = [{"f_model_id": "222"}]
        m2.cursor.return_value = m3
        # 为PymysqlPool.get_pool函数创建mock对象
        PymysqlPool.get_pool = mock.Mock(return_value=m1)
        res = llm_model_dao.add_data_into_model_list("222", "222", "222", "222", "222", "222", "222", "222", "222", "222",
                                                 0, 32, 72)
        self.assertEqual(None, res)


# rename_model函数的测试类
class TestRenameModel(TestCase):
    def setUp(self) -> None:
        self.mysqlPool = PymysqlPool
        self.model_config_rename = llm_utils.model_config.model_config_rename

    def tearDown(self) -> None:
        PymysqlPool = self.mysqlPool
        self.model_config_rename = llm_utils.model_config.model_config_rename
        StandLogger.stand_log_shutdown()

    def test_rename_model_success(self):
        m1 = mock.MagicMock()
        m2 =mock.MagicMock()
        m1.connection.return_value = m2
        m3 = mock.MagicMock()
        m3.fetchall.return_value = [{"f_model_id": "222"}]
        m2.cursor.return_value = m3
        # 为PymysqlPool.get_pool函数创建mock对象
        PymysqlPool.get_pool = mock.Mock(return_value=m1)
        llm_utils.model_config.model_config_rename = mock.Mock(return_value=None)
        res = llm_model_dao.rename_model("222", "222", "222", 1, 32, 72)
        self.assertEqual(None, res)


# get_data_from_model_list_by_name_fuzzy函数的测试类
class TestGetDataFromModelListByNameFuzzy(TestCase):
    def setUp(self) -> None:
        self.mysqlPool = PymysqlPool
        self.get_admin_user_id = user_info.get_admin_user_id

    def tearDown(self) -> None:
        PymysqlPool = self.mysqlPool
        user_info.get_admin_user_id = self.get_admin_user_id
        StandLogger.stand_log_shutdown()

    def test_get_data_from_model_list_by_name_fuzzy_success(self):
        m1 = mock.MagicMock()
        m2 = mock.MagicMock()
        m1.connection.return_value = m2
        m3 = mock.MagicMock()
        m3.fetchall.return_value = [{"f_model_id": "222"}]
        m2.cursor.return_value = m3
        # 为PymysqlPool.get_pool函数创建mock对象
        PymysqlPool.get_pool = mock.Mock(return_value=m1)
        user_info.get_admin_user_id = mock.Mock(return_value="12321")
        res = llm_model_dao.get_data_from_model_list_by_name_fuzzy("222", 1, 1, "222", "222", "222", "222", 0, None)
        self.assertEqual([{"f_model_id": "222"}], res)


# get_data_from_model_list_by_name_fuzzy_and_series函数的测试类
class TestGetDataFromModelListByNameFuzzyAndSeries(TestCase):
    def setUp(self) -> None:
        self.mysqlPool = PymysqlPool
        self.get_admin_user_id = user_info.get_admin_user_id

    def tearDown(self) -> None:
        PymysqlPool = self.mysqlPool
        user_info.get_admin_user_id = self.get_admin_user_id
        StandLogger.stand_log_shutdown()

    def test_get_data_from_model_list_by_name_fuzzy_and_series_success(self):
        m1 = mock.MagicMock()
        m2 = mock.MagicMock()
        m1.connection.return_value = m2
        m3 = mock.MagicMock()
        m3.fetchall.return_value = [{"f_model_id": "222"}]
        m2.cursor.return_value = m3
        # 为PymysqlPool.get_pool函数创建mock对象
        PymysqlPool.get_pool = mock.Mock(return_value=m1)
        user_info.get_admin_user_id = mock.Mock(return_value="12321")
        res = llm_model_dao.get_data_from_model_list_by_name_fuzzy_and_series("222", "222", 1, 1, "222", "222", "222", "222", 1, 0)
        self.assertEqual([{"f_model_id": "222"}], res)


# get_all_data_from_model_series函数的测试类
class TestGetAllDataFromModelSeries(TestCase):
    def setUp(self) -> None:
        self.mysqlPool = PymysqlPool

    def tearDown(self) -> None:
        PymysqlPool = self.mysqlPool
        StandLogger.stand_log_shutdown()



# get_all_data_from_model_param函数的测试类
class TestGetAllDataFromModelParam(TestCase):
    def setUp(self) -> None:
        self.mysqlPool = PymysqlPool

    def tearDown(self) -> None:
        PymysqlPool = self.mysqlPool
        StandLogger.stand_log_shutdown()

    def test_get_all_data_from_model_param_success(self):
        m1 = mock.MagicMock()
        m2 = mock.MagicMock()
        m1.connection.return_value = m2
        m3 = mock.MagicMock()
        m3.fetchall.return_value = [{"f_model_id": "222"}]
        m2.cursor.return_value = m3
        # 为PymysqlPool.get_pool函数创建mock对象
        PymysqlPool.get_pool = mock.Mock(return_value=m1)
        res = llm_model_dao.get_all_data_from_model_param()
        self.assertEqual([{"f_model_id": "222"}], res)


# get_data_from_model_param_by_param_id函数的测试类
class TestGetDataFromModelParamByParamId(TestCase):
    def setUp(self) -> None:
        self.mysqlPool = PymysqlPool

    def tearDown(self) -> None:
        PymysqlPool = self.mysqlPool
        StandLogger.stand_log_shutdown()

    def test_get_data_from_model_param_by_param_id_success(self):
        m1 = mock.MagicMock()
        m2 = mock.MagicMock()
        m1.connection.return_value = m2
        m3 = mock.MagicMock()
        m3.fetchall.return_value = [{"f_model_id": "222"}]
        m2.cursor.return_value = m3
        # 为PymysqlPool.get_pool函数创建mock对象
        PymysqlPool.get_pool = mock.Mock(return_value=m1)
        res = llm_model_dao.get_data_from_model_param_by_param_id("222")
        self.assertEqual([{"f_model_id": "222"}], res)




# get_model_by_name函数的测试类
class TestGetModelByName(TestCase):
    def setUp(self) -> None:
        self.mysqlPool = PymysqlPool

    def tearDown(self) -> None:
        PymysqlPool = self.mysqlPool
        StandLogger.stand_log_shutdown()

    def test_get_model_by_name_success(self):
        m1 = mock.MagicMock()
        m2 = mock.MagicMock()
        m1.connection.return_value = m2
        m3 = mock.MagicMock()
        m3.fetchall.return_value = [{"f_model_id": "222"}]
        m2.cursor.return_value = m3
        # 为PymysqlPool.get_pool函数创建mock对象
        PymysqlPool.get_pool = mock.Mock(return_value=m1)
        res = llm_model_dao.get_model_by_name("222")
        self.assertEqual([{"f_model_id": "222"}], res)


# check_model_is_exist函数的测试类
class TestCheckModelIsExist(TestCase):
    def setUp(self) -> None:
        self.mysqlPool = PymysqlPool

    def tearDown(self) -> None:
        PymysqlPool = self.mysqlPool
        StandLogger.stand_log_shutdown()

    def test_check_model_is_exist_success(self):
        m1 = mock.MagicMock()
        m2 = mock.MagicMock()
        m1.connection.return_value = m2
        m3 = mock.MagicMock()
        m3.fetchall.return_value = [{"COUNT(f_model_id)": 0}]
        m2.cursor.return_value = m3
        # 为PymysqlPool.get_pool函数创建mock对象
        PymysqlPool.get_pool = mock.Mock(return_value=m1)
        res = llm_model_dao.check_model_is_exist("222")
        self.assertEqual(False, res)


# check_model_unique函数的测试类
class TestCheckModelUnique(TestCase):
    def setUp(self) -> None:
        self.mysqlPool = PymysqlPool

    def tearDown(self) -> None:
        PymysqlPool = self.mysqlPool
        StandLogger.stand_log_shutdown()

    def test_check_model_unique_success(self):
        m1 = mock.MagicMock()
        m2 = mock.MagicMock()
        m1.connection.return_value = m2
        m3 = mock.MagicMock()
        m3.fetchall.return_value = [{"account_id": "111", "f_model_config": '{"api_base": "111"}'}]
        m2.cursor.return_value = m3
        # 为PymysqlPool.get_pool函数创建mock对象
        PymysqlPool.get_pool = mock.Mock(return_value=m1)
        res = llm_model_dao.check_model_unique("222", "222", "222", None)
        self.assertEqual(False, res)


# get_model_default_paras函数的测试类
class TestGetModelDefaultParas(TestCase):
    def setUp(self) -> None:
        self.mysqlPool = PymysqlPool

    def tearDown(self) -> None:
        PymysqlPool = self.mysqlPool
        StandLogger.stand_log_shutdown()

    def test_get_model_default_paras_success(self):
        m1 = mock.MagicMock()
        m2 = mock.MagicMock()
        m1.connection.return_value = m2
        m3 = mock.MagicMock()
        m3.fetchall.return_value = [{"f_model_id": "222", "f_model_name": "222", "f_model_series": "222", "f_model": "222"}]
        m2.cursor.return_value = m3
        # 为PymysqlPool.get_pool函数创建mock对象
        PymysqlPool.get_pool = mock.Mock(return_value=m1)
        res = llm_model_dao.get_model_default_paras()
        self.assertEqual({'222': {'model': '222', 'model_name': '222', 'model_series': '222'}}, res)


class TestGetApiModelByModelType(TestCase):
    def setUp(self) -> None:
        self.mysqlPool = PymysqlPool
        self.get_admin_user_id = user_info.get_admin_user_id

    def tearDown(self) -> None:
        PymysqlPool = self.mysqlPool
        user_info.get_admin_user_id = self.get_admin_user_id
        StandLogger.stand_log_shutdown()

    def test_get_api_model_by_model_type_success1(self):
        m1 = mock.MagicMock()
        m2 = mock.MagicMock()
        m1.connection.return_value = m2
        m3 = mock.MagicMock()
        m3.fetchall.return_value = [
            {"f_model_id": "222", "f_model_name": "222", "f_model_series": "222", "f_model": "222"}]
        m2.cursor.return_value = m3
        # 为PymysqlPool.get_pool函数创建mock对象
        PymysqlPool.get_pool = mock.Mock(return_value=m1)
        user_info.get_admin_user_id = mock.Mock(return_value="12321")
        self.assertEqual([{"f_model_id": "222", "f_model_name": "222", "f_model_series": "222", "f_model": "222"}], res)

    def test_get_api_model_by_model_type_success2(self):
        m1 = mock.MagicMock()
        m2 = mock.MagicMock()
        m1.connection.return_value = m2
        m3 = mock.MagicMock()
        m3.fetchall.return_value = [
            {"f_model_id": "222", "f_model_name": "222", "f_model_series": "222", "f_model": "222"}]
        m2.cursor.return_value = m3
        # 为PymysqlPool.get_pool函数创建mock对象
        PymysqlPool.get_pool = mock.Mock(return_value=m1)
        user_info.get_admin_user_id = mock.Mock(return_value="12321")
        self.assertEqual([{"f_model_id": "222", "f_model_name": "222", "f_model_series": "222", "f_model": "222"}], res)

    def test_get_api_model_by_model_type_success3(self):
        m1 = mock.MagicMock()
        m2 = mock.MagicMock()
        m1.connection.return_value = m2
        m3 = mock.MagicMock()
        m3.fetchall.return_value = [
            {"f_model_id": "222", "f_model_name": "222", "f_model_series": "222", "f_model": "222"}]
        m2.cursor.return_value = m3
        # 为PymysqlPool.get_pool函数创建mock对象
        PymysqlPool.get_pool = mock.Mock(return_value=m1)
        user_info.get_admin_user_id = mock.Mock(return_value="12321")
        self.assertEqual([{"f_model_id": "222", "f_model_name": "222", "f_model_series": "222", "f_model": "222"}], res)


if __name__ == '__main__':
    import unittest

    unittest.main()
