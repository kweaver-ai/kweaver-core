import json
import time
from tornado.web import RequestHandler
from common.response import ErrorHandler
from logics.tag import TagExtractionService, TagExtractionV2Service
from jsonschema import ValidationError
from utils.utils import *
from errors.errors import *

TAG_EXTRACTION_PARAM_SCHEMA = {"rule":"tag_extraction.json"}

class TagExtractionHandler(RequestHandler):
    async def post(self):
        error_handler = ErrorHandler(self)
        try:
            headers = self.request.headers
            host, access_token = str(headers.get('x-as-address')), str(headers.get('x-authorization'))
            if access_token.startswith("Bearer "):
                access_token = access_token[len("Bearer "):]
            requestBody = json.loads(self.request.body)
            validate_params(requestBody, TAG_EXTRACTION_PARAM_SCHEMA["rule"])
            tag_services = TagExtractionService(host, access_token)
            tags = await tag_services.extraction_by_rule(requestBody)
            self.set_status(200)
            response_content = {"tags": tags}
            self.set_header("Content-Type", "application/json")
            self.write(response_content)
        except ValidationError as e:
            error_handler.BAD_REQUEST(e)
        except Exception as e:
            error_handler.INTERNAL_ERROR(e)

class TagExtractionV2Handler(RequestHandler):
    async def post(self):
        error_handler = ErrorHandler(self)
        try:
            headers = self.request.headers
            requestBody = json.loads(self.request.body)
            validate_params(requestBody, TAG_EXTRACTION_PARAM_SCHEMA["rule"])
            tag_services = TagExtractionV2Service()
            tags = await tag_services.extraction_by_rule(requestBody)
            self.set_status(200)
            response_content = {"tags": tags}
            self.set_header("Content-Type", "application/json")
            self.write(response_content)
        except ValidationError as e:
            error_handler.BAD_REQUEST(e)
        except Exception as e:
            error_handler.INTERNAL_ERROR(e)
