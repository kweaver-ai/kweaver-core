import json
import time
from tornado.web import RequestHandler
from common.response import ErrorHandler
from logics.analysis import AnalysisService
from jsonschema import ValidationError
from utils.utils import *
from utils.request import *
from errors.errors import *

PARSE_DOC_PARAM_SCHEMA = {"rule":"analysis/parse_doc.json"}

class ParseDoc(RequestHandler):
    async def get(self):
        error_handler = ErrorHandler(self)
        try:
            headers = self.request.headers
            _, access_token = getHeadersInfo(headers)
            docid = self.get_query_argument('docid')
            validate_params({'docid': docid}, PARSE_DOC_PARAM_SCHEMA['rule'])

            analysis_services = AnalysisService()
            res = await analysis_services.parse_doc(docid, access_token)
            self.set_status(200)
            # response_content = {"tags": tags}
            self.set_header("Content-Type", "application/json")
            self.write(res)
        except ValidationError as e:
            error_handler.BAD_REQUEST(e)
        except Exception as e:
            error_handler.INTERNAL_ERROR(e)

class SlotInfo(RequestHandler):
    async def get(self):
        error_handler = ErrorHandler(self)
        try:
            headers = self.request.headers
            host, access_token = str(headers.get('x-as-address')), str(headers.get('x-authorization'))
            if access_token == '':
                access_token = str(headers.get('Authorization'))
            if access_token.startswith("Bearer "):
                access_token = access_token[len("Bearer "):]
            requestBody = json.loads(self.request.body)
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
        return self.write("health")
