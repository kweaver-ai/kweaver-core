#!/usr/bin/python3
# -*- coding:utf-8 -*-

import json
from jsonschema import ValidationError
from tornado.web import RequestHandler

from errors.errors import BAD_REQUEST_ERROR, INTERNAL_ERROR, BadParameterException, RequestProcessingException
from logics.ocr import OCRService
from common.response import ErrorHandler
from utils.utils import *

class BuildInOCRHandler(RequestHandler):

    async def post(self):
        error_handler = ErrorHandler(self)
        try:
            headers = self.request.headers
            host, access_token = str(headers.get('x-as-address')), str(headers.get('x-authorization'))
            if access_token.startswith("Bearer "):
                access_token = access_token[len("Bearer "):]
            ocr_opt = OCRService(host, access_token)
            requestBody = json.loads(self.request.body)
            validate_params(requestBody, "build_in_ocr.json")
            res = await ocr_opt.buildin_ocr(requestBody)
            self.set_status(200)
            self.set_header("Content-Type", "application/json")
            self.write(res)
        except ValidationError as e:
            error_handler.BAD_REQUEST(e)
        except BadParameterException as e:
            error_handler.BAD_REQUEST(e)
        except Exception as e:
            error_handler.INTERNAL_ERROR(e)

class ExternalOCRHandler(RequestHandler):

    async def post(self, task_id = None):
        error_handler = ErrorHandler(self)
        try:
            if task_id != None:
                error_handler.PAGE_NOT_FOUND()
                return
            headers = self.request.headers
            host, access_token = str(headers.get('x-as-address')), str(headers.get('x-authorization'))
            if access_token.startswith("Bearer "):
                access_token = access_token[len("Bearer "):]
            ocr_opt = OCRService(host, access_token)
            requestBody = json.loads(self.request.body)
            validate_params(requestBody, "external_ocr.json")
            res = await ocr_opt.external_ocr(requestBody)
            self.set_status(200)
            self.set_header("Content-Type", "application/json")
            self.write(res)
        except ValidationError as e:
            error_handler.BAD_REQUEST(e)
        except BadParameterException as e:
            error_handler.BAD_REQUEST(e)
        except Exception as e:
            error_handler.INTERNAL_ERROR(e)

    async def delete(self, task_id = None):
        error_handler = ErrorHandler(self)
        try:
            if task_id == None:
                error_handler.PAGE_NOT_FOUND()
                return
            task_ids = task_id.strip(",").split(',')
            ocr_opt = OCRService()
            await ocr_opt.delete_ocr_task(task_ids)
        except Exception as e:
            error_handler.INTERNAL_ERROR(e)

class OCRResultHandler(RequestHandler):

    async def get(self):
        error_handler = ErrorHandler(self)
        try:
            task_id = self.get_argument("task_id", default='')
            rec_type = self.get_argument("rec_type", default='')
            ocr_opt = OCRService()
            res = await ocr_opt.get_ocr_result(task_id=task_id, rec_type=rec_type)
            self.set_status(200)
            self.set_header("Content-Type", "application/json")
            self.write(res)
        except RequestProcessingException as re:
            self.set_status(202)
        except Exception as e:
            error_handler.INTERNAL_ERROR(e)
