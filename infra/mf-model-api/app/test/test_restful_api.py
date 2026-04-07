"""测试 restful_api 模块"""
import pytest
from app.commons.restful_api import (
    get_model_restful_api_document,
    get_prompt_restful_api_document,
    get_embedding_restful_api_document,
    get_reranker_restful_api_document,
    get_spr_restful_api_document,
    get_info_extract_restful_api_document,
    get_audio_restful_api_document
)


class TestGetModelRestfulApiDocument:
    """测试get_model_restful_api_document函数"""

    def test_basic_structure(self):
        """测试基本结构"""
        result = get_model_restful_api_document("test_llm_id")
        assert "openapi" in result
        assert "info" in result
        assert "paths" in result
        assert "components" in result

    def test_openapi_version(self):
        """测试OpenAPI版本"""
        result = get_model_restful_api_document("test_llm_id")
        assert result["openapi"] == "3.0.2"

    def test_info_section(self):
        """测试info部分"""
        result = get_model_restful_api_document("test_llm_id")
        assert result["info"]["title"] == "大语言模型服务"
        assert result["info"]["version"] == "1.0.0"

    def test_path_with_llm_id(self):
        """测试路径包含llm_id"""
        llm_id = "123456"
        result = get_model_restful_api_document(llm_id)
        expected_path = f"/api/model-factory/v1/llm-used/{llm_id}"
        assert expected_path in result["paths"]

    def test_components_schemas(self):
        """测试components schemas"""
        result = get_model_restful_api_document("test_id")
        assert "serviceReq" in result["components"]["schemas"]
        assert "Error" in result["components"]["schemas"]
        assert "LLMResp" in result["components"]["schemas"]


class TestGetPromptRestfulApiDocument:
    """测试get_prompt_restful_api_document函数"""

    def test_basic_structure(self):
        """测试基本结构"""
        result = get_prompt_restful_api_document("prompt_id", {"var1": "value1"}, "test prompt")
        assert "openapi" in result
        assert "info" in result
        assert "paths" in result

    def test_info_section(self):
        """测试info部分"""
        result = get_prompt_restful_api_document("prompt_id", {}, "")
        assert result["info"]["title"] == "提示词服务"
        assert result["info"]["version"] == "1.0.0"

    def test_path_with_prompt_id(self):
        """测试路径包含prompt_id"""
        prompt_id = "prompt_123"
        result = get_prompt_restful_api_document(prompt_id, {}, "")
        expected_path = f"/api/model-factory/v1/prompt/{prompt_id}/used"
        assert expected_path in result["paths"]

    def test_variable_dict_in_example(self):
        """测试变量字典在示例中"""
        var_dict = {"name": "test", "age": 25}
        result = get_prompt_restful_api_document("prompt_id", var_dict, "")
        # 检查示例中包含变量
        path_key = list(result["paths"].keys())[0]
        example = result["paths"][path_key]["post"]["requestBody"]["content"]["application/json"]["examples"]["request"]["value"]
        assert "inputs" in example

    def test_prompt_in_description(self):
        """测试提示词在描述中"""
        prompt = "你是一个AI助手"
        result = get_prompt_restful_api_document("prompt_id", {}, prompt)
        path_key = list(result["paths"].keys())[0]
        description = result["paths"][path_key]["post"]["description"]
        assert prompt in description


class TestGetEmbeddingRestfulApiDocument:
    """测试get_embedding_restful_api_document函数"""

    def test_basic_structure(self):
        """测试基本结构"""
        result = get_embedding_restful_api_document("embedding_model")
        assert "openapi" in result
        assert "info" in result
        assert "paths" in result

    def test_info_section(self):
        """测试info部分"""
        result = get_embedding_restful_api_document("model")
        assert result["info"]["title"] == "embedding小模型"

    def test_model_name_in_example(self):
        """测试示例中的模型名称"""
        model_name = "test_embedding_model"
        result = get_embedding_restful_api_document(model_name)
        path = "/api/model-factory/v1/small_model_run"
        example = result["paths"][path]["post"]["requestBody"]["content"]["application/json"]["examples"]["request"]["value"]
        assert example["model_name"] == model_name

    def test_response_schema(self):
        """测试响应schema"""
        result = get_embedding_restful_api_document("model")
        assert "EmbeddingResp" in result["components"]["schemas"]


class TestGetRerankerRestfulApiDocument:
    """测试get_reranker_restful_api_document函数"""

    def test_basic_structure(self):
        """测试基本结构"""
        result = get_reranker_restful_api_document("reranker_model")
        assert "openapi" in result
        assert "info" in result

    def test_info_section(self):
        """测试info部分"""
        result = get_reranker_restful_api_document("model")
        assert result["info"]["title"] == "reranker小模型"

    def test_model_name_in_example(self):
        """测试示例中的模型名称"""
        model_name = "test_reranker"
        result = get_reranker_restful_api_document(model_name)
        path = "/api/model-factory/v1/small_model_run"
        example = result["paths"][path]["post"]["requestBody"]["content"]["application/json"]["examples"]["request"]["value"]
        assert example["model_name"] == model_name

    def test_reranker_parameters(self):
        """测试reranker参数"""
        result = get_reranker_restful_api_document("model")
        schema = result["components"]["schemas"]["serviceReq"]
        assert "slices" in schema["properties"]["param_data"]["properties"]
        assert "query" in schema["properties"]["param_data"]["properties"]


class TestGetSprRestfulApiDocument:
    """测试get_spr_restful_api_document函数"""

    def test_basic_structure(self):
        """测试基本结构"""
        result = get_spr_restful_api_document("spr_model")
        assert "openapi" in result
        assert "info" in result

    def test_info_section(self):
        """测试info部分"""
        result = get_spr_restful_api_document("model")
        assert result["info"]["title"] == "spr小模型"

    def test_model_name_in_example(self):
        """测试示例中的模型名称"""
        model_name = "test_spr"
        result = get_spr_restful_api_document(model_name)
        path = "/api/model-factory/v1/small_model_run"
        example = result["paths"][path]["post"]["requestBody"]["content"]["application/json"]["examples"]["request"]["value"]
        assert example["model_name"] == model_name

    def test_spr_param_data_structure(self):
        """测试SPR参数数据结构"""
        result = get_spr_restful_api_document("model")
        schema = result["components"]["schemas"]["serviceReq"]
        assert schema["properties"]["param_data"]["type"] == "array"


class TestGetInfoExtractRestfulApiDocument:
    """测试get_info_extract_restful_api_document函数"""

    def test_basic_structure(self):
        """测试基本结构"""
        result = get_info_extract_restful_api_document("info_extract_model")
        assert "openapi" in result
        assert "info" in result

    def test_info_section(self):
        """测试info部分"""
        result = get_info_extract_restful_api_document("model")
        assert result["info"]["title"] == "info_extract小模型"

    def test_model_name_in_example(self):
        """测试示例中的模型名称"""
        model_name = "test_info_extract"
        result = get_info_extract_restful_api_document(model_name)
        path = "/api/model-factory/v1/small_model_run"
        example = result["paths"][path]["post"]["requestBody"]["content"]["application/json"]["examples"]["request"]["value"]
        assert example["model_name"] == model_name

    def test_info_extract_parameters(self):
        """测试信息抽取参数"""
        result = get_info_extract_restful_api_document("model")
        schema = result["components"]["schemas"]["serviceReq"]
        param_data = schema["properties"]["param_data"]["properties"]
        assert "schema" in param_data
        assert "texts" in param_data


class TestGetAudioRestfulApiDocument:
    """测试get_audio_restful_api_document函数"""

    def test_basic_structure(self):
        """测试基本结构"""
        result = get_audio_restful_api_document("audio_model")
        assert "openapi" in result
        assert "info" in result

    def test_info_section(self):
        """测试info部分"""
        result = get_audio_restful_api_document("model")
        assert result["info"]["title"] == "audio小模型"

    def test_model_name_in_example(self):
        """测试示例中的模型名称"""
        model_name = "test_audio"
        result = get_audio_restful_api_document(model_name)
        path = "/api/model-factory/v1/small_model_run"
        example = result["paths"][path]["post"]["requestBody"]["content"]["application/json"]["examples"]["request"]["value"]
        assert example["model_name"] == model_name

    def test_audio_parameters(self):
        """测试音频参数"""
        result = get_audio_restful_api_document("model")
        schema = result["components"]["schemas"]["serviceReq"]
        param_data = schema["properties"]["param_data"]["properties"]
        assert "file_url" in param_data
        assert "file_size" in param_data
        assert "file_name" in param_data

    def test_audio_response(self):
        """测试音频响应"""
        result = get_audio_restful_api_document("model")
        assert "AudioResp" in result["components"]["schemas"]
        audio_resp = result["components"]["schemas"]["AudioResp"]
        assert "id" in audio_resp["properties"]

