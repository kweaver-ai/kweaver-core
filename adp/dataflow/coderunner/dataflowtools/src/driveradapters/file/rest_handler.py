import json
from tornado.web import RequestHandler
from common.response import ErrorHandler
from logics.file_operator import FileOperationService
from jsonschema import ValidationError
from utils.utils import *
from errors.errors import *

CREATE_FILE_PARAM_SCHEMA = {
    "xlsx": "create_file.json",
    "docx": "create_file.json",
    "pdf": "create_file.json",
    "md": "create_file.json",
}
UPDATE_FILE_PARAM_SCHEMA = {
    "xlsx": "update_excel_file.json",
    "docx": "update_docx_file.json",
    "pdf": "update_docx_file.json",
    "md": "update_docx_file.json",
}

class CreateFileHandler(RequestHandler):

    async def post(self):
        error_handler = ErrorHandler(self)
        try:
            headers = self.request.headers
            host, access_token = str(headers.get('x-as-address')), str(headers.get('x-authorization'))
            if access_token.startswith("Bearer "):
                access_token = access_token[len("Bearer "):]
            requestBody = json.loads(self.request.body)
            file_type = requestBody["type"]
            if file_type in CREATE_FILE_PARAM_SCHEMA:
                validate_params(requestBody, CREATE_FILE_PARAM_SCHEMA[file_type])
            else:
                raise ValidationError(f"'{file_type}' is not one of ['xlsx', 'docx', 'pdf', 'md']")
            file_opt = FileOperationService(host, access_token)
            doc_id = await file_opt.async_create_file(requestBody)
            self.set_status(200)
            response_content = {"docid": doc_id}
            self.set_header("Content-Type", "application/json")
            self.write(response_content)
        except ValidationError as e:
            error_handler.BAD_REQUEST(e)
        except BadParameterException as e:
            error_handler.BAD_REQUEST(e)
        except Exception as e:
            error_handler.INTERNAL_ERROR(e)

class UpdateFileHandler(RequestHandler):

    async def put(self):
        error_handler = ErrorHandler(self)
        try:
            headers = self.request.headers
            host, access_token = str(headers.get('x-as-address')), str(headers.get('x-authorization'))
            if access_token.startswith("Bearer "):
                access_token = access_token[len("Bearer "):]
            requestBody = json.loads(self.request.body)
            file_type = requestBody["type"]
            if file_type in UPDATE_FILE_PARAM_SCHEMA:
                validate_params(requestBody, UPDATE_FILE_PARAM_SCHEMA[file_type])
            else:
                raise ValidationError(f"'{file_type}' is not one of ['xlsx', 'docx', 'pdf', 'md']")
            file_opt = FileOperationService(host, access_token)
            doc_id = await file_opt.async_update_file(requestBody)
            self.set_status(200)
            response_content = {"docid": doc_id}
            self.set_header("Content-Type", "application/json")
            self.write(response_content)
        except ValidationError as e:
            error_handler.BAD_REQUEST(e)
        except BadParameterException as e:
            error_handler.BAD_REQUEST(e)
        except Exception as e:
            error_handler.INTERNAL_ERROR(e)
