"""测试 response 模块"""
import pytest
from fastapi import status
from fastapi.responses import JSONResponse
from app.commons.response import error_response, correct_response


class TestErrorResponse:
    """测试error_response函数"""

    def test_error_response_basic(self):
        """测试基本错误响应"""
        response = error_response(
            status_code=400,
            code="TEST.Error",
            detail="测试错误详情",
            language="zh"
        )
        assert isinstance(response, JSONResponse)
        assert response.status_code == 400

    def test_error_response_with_solution(self):
        """测试带解决方案的错误响应"""
        response = error_response(
            status_code=500,
            code="TEST.ServerError",
            detail="服务器错误",
            solution="请联系管理员",
            language="zh"
        )
        assert isinstance(response, JSONResponse)
        assert response.status_code == 500

    def test_error_response_with_link(self):
        """测试带链接的错误响应"""
        response = error_response(
            status_code=404,
            code="TEST.NotFound",
            detail="资源未找到",
            link="http://help.example.com",
            language="zh"
        )
        assert isinstance(response, JSONResponse)
        assert response.status_code == 404

    def test_error_response_all_params(self):
        """测试所有参数的错误响应"""
        response = error_response(
            status_code=403,
            code="TEST.Forbidden",
            detail="权限不足",
            solution="请申请权限",
            link="http://help.example.com",
            language="zh"
        )
        assert isinstance(response, JSONResponse)
        assert response.status_code == 403

    def test_error_response_english(self):
        """测试英文错误响应"""
        response = error_response(
            status_code=400,
            code="TEST.Error",
            detail="Test error detail",
            language="en"
        )
        assert isinstance(response, JSONResponse)
        assert response.status_code == 400


class TestCorrectResponse:
    """测试correct_response函数"""

    def test_correct_response_default(self):
        """测试默认正确响应"""
        data = {"message": "success"}
        response = correct_response(data=data)
        assert isinstance(response, JSONResponse)
        assert response.status_code == 200

    def test_correct_response_custom_status(self):
        """测试自定义状态码的正确响应"""
        data = {"id": "123", "status": "created"}
        response = correct_response(http_code=201, data=data)
        assert isinstance(response, JSONResponse)
        assert response.status_code == 201

    def test_correct_response_none_data(self):
        """测试空数据的正确响应"""
        response = correct_response(data=None)
        assert isinstance(response, JSONResponse)
        assert response.status_code == 200

    def test_correct_response_list_data(self):
        """测试列表数据的正确响应"""
        data = [{"id": 1}, {"id": 2}]
        response = correct_response(data=data)
        assert isinstance(response, JSONResponse)
        assert response.status_code == 200

    def test_correct_response_dict_data(self):
        """测试字典数据的正确响应"""
        data = {"count": 10, "items": []}
        response = correct_response(data=data)
        assert isinstance(response, JSONResponse)
        assert response.status_code == 200

    def test_correct_response_202_accepted(self):
        """测试202 Accepted响应"""
        data = {"task_id": "abc123"}
        response = correct_response(http_code=status.HTTP_202_ACCEPTED, data=data)
        assert isinstance(response, JSONResponse)
        assert response.status_code == 202

    def test_correct_response_204_no_content(self):
        """测试204 No Content响应"""
        response = correct_response(http_code=status.HTTP_204_NO_CONTENT, data=None)
        assert isinstance(response, JSONResponse)
        assert response.status_code == 204

