import allure
import pytest
import time
from lib.dataflow_como_operator import AutomationClient

@allure.epic("流程全景可观测")
class TestFullView:
    client = AutomationClient()

    @allure.title("流程全景可观测 - 基本查询")
    def test_full_view_basic(self, Headers):
        end_time = int(time.time())
        start_time = end_time - 24 * 3600
        params = {"start_time": start_time, "end_time": end_time}
        result = self.client.GetFullView(params, Headers)
        assert result[0] == 200, f"应返回200，实际为: {result[0]}"
        for key in ["dag_cnt", "success", "failed", "total"]:
            assert key in result[1], f"响应体缺少字段: {key}"
        assert isinstance(result[1]["dag_cnt"], int)
        assert isinstance(result[1]["success"], int)
        assert isinstance(result[1]["failed"], int)
        assert isinstance(result[1]["total"], int)


    @allure.title("流程全景可观测 - 按类型查询（combo-operator）")
    def test_full_view_combo_operator(self, Headers):
        end_time = int(time.time())
        start_time = end_time - 24 * 3600
        params = {"start_time": start_time, "end_time": end_time, "type": "combo-operator"}
        result = self.client.GetFullView(params, Headers)
        assert result[0] == 200, f"应返回200，实际为: {result[0]}"
        for key in ["dag_cnt", "success", "failed", "total"]:
            assert key in result[1], f"响应体缺少字段: {key}"

    @allure.title("流程全景可观测 - 按类型查询（data-flow）")
    def test_full_view_data_flow(self, Headers):
        end_time = int(time.time())
        start_time = end_time - 24 * 3600
        params = {"start_time": start_time, "end_time": end_time, "type": "data-flow"}
        result = self.client.GetFullView(params, Headers)
        assert result[0] == 200, f"应返回200，实际为: {result[0]}"
        for key in ["dag_cnt", "success", "failed", "total"]:
            assert key in result[1], f"响应体缺少字段: {key}"

    @allure.title("流程全景可观测 - 不同时间范围查询")
    @pytest.mark.parametrize("hours", [1, 24, 168])
    def test_full_view_time_range(self, Headers, hours):
        end_time = int(time.time())
        start_time = end_time - hours * 3600
        params = {"start_time": start_time, "end_time": end_time}
        result = self.client.GetFullView(params, Headers)
        assert result[0] == 200, f"应返回200，实际为: {result[0]}"
        for key in ["dag_cnt", "success", "failed", "total"]:
            assert key in result[1], f"响应体缺少字段: {key}"

    @allure.title("流程全景可观测 - 结束时间早于开始时间")
    def test_full_view_invalid_time(self, Headers):
        params = {"start_time": 1747564800, "end_time": 1747391537}
        result = self.client.GetFullView(params, Headers)
        assert result[0] == 400, f"应返回400，实际为: {result[0]}"

    @allure.title("流程全景可观测 - 非法类型参数")
    def test_full_view_invalid_type(self, Headers):
        end_time = int(time.time())
        start_time = end_time - 24 * 3600
        params = {"start_time": start_time, "end_time": end_time, "type": "invalid-type"}
        result = self.client.GetFullView(params, Headers)
        assert result[0] == 400, f"应返回400，实际为: {result[0]}"

    @allure.title("流程全景可观测 - 缺少start_time参数")
    def test_full_view_missing_start_time(self, Headers):
        end_time = int(time.time())
        params = {"end_time": end_time}
        result = self.client.GetFullView(params, Headers)
        assert result[0] == 200, f"应返回200，实际为: {result[0]}"

    @allure.title("流程全景可观测 - 缺少end_time参数")
    def test_full_view_missing_end_time(self, Headers):
        start_time = int(time.time()) - 24 * 3600
        params = {"start_time": start_time}
        result = self.client.GetFullView(params, Headers)
        assert result[0] == 200, f"应返回200，实际为: {result[0]}"

    @allure.title("流程全景可观测 - start_time等于end_time")
    def test_full_view_start_time_equal_end_time(self, Headers):
        now = int(time.time())
        params = {"start_time": now, "end_time": now}
        result = self.client.GetFullView(params, Headers)
        assert result[0] == 400, f"应返回400，实际为: {result[0]}"


    @allure.title("流程全景可观测 - 超大时间跨度")
    def test_full_view_large_time_range(self, Headers):
        end_time = int(time.time())
        start_time = end_time - 10 * 365 * 24 * 3600  # 10年
        params = {"start_time": start_time, "end_time": end_time}
        result = self.client.GetFullView(params, Headers)
        assert result[0] in [200, 400], f"应返回200或400，实际为: {result[0]}"
