import os
import urllib3
import tornado
import tornado.ioloop

from tornado.options import options
from tornado.httpserver import HTTPServer
from tornado.web import Application
from tornado.routing import (Rule, PathMatches)

from driveradapters.health.rest_handler import (HealthHandler, AliveHandler)
from driveradapters.ocr.rest_handler import *
from driveradapters.file.rest_handler import (CreateFileHandler, UpdateFileHandler)
from driveradapters.tag.rest_handler import (TagExtractionHandler, TagExtractionV2Handler)
from driveradapters.data_analysis.rest_handler import (SlotInfo, ParseDoc)
from driveradapters.py_pkg.rest_handler import PyPkgHandler
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

def make_private_app():
    url = "/api/coderunner/v1"
    url_v2 = "/api/coderunner/v2"
    return Application(
        [
            Rule(PathMatches(rf"{url}/health/ready/?$"), HealthHandler),
            Rule(PathMatches(rf"{url}/health/alive/?$"), AliveHandler),
            Rule(PathMatches(rf"{url}/documents/file?$"), CreateFileHandler),
            Rule(PathMatches(rf"{url}/documents/content?$"), UpdateFileHandler),
            Rule(PathMatches(rf"{url}/built-in/ocr/task?$"), BuildInOCRHandler),
            Rule(PathMatches(rf"{url}/external/ocr/task(?:/([\w,]+))?$"), ExternalOCRHandler),
            Rule(PathMatches(rf"{url}/external/ocr/result?$"), OCRResultHandler),
            Rule(PathMatches(rf"{url}/tag/extraction?$"), TagExtractionHandler),
            Rule(PathMatches(rf"{url_v2}/tag/extraction?$"), TagExtractionV2Handler),
            Rule(PathMatches(rf"{url}/data-analisis/parse-doc?$"), ParseDoc),
            Rule(PathMatches(rf"{url}/py-packages/([\w,]+)?$"), PyPkgHandler, dict(need_auth=False)),
        ]
    )

def make_public_app():
    url = "/api/coderunner/v1"
    return Application(
        [
            Rule(PathMatches(rf"{url}/health/ready/?$"), HealthHandler),
            Rule(PathMatches(rf"{url}/health/alive/?$"), AliveHandler),
            Rule(PathMatches(rf"{url}/py-package?$"), PyPkgHandler, dict(need_automation_admin=True)),
            Rule(PathMatches(rf"{url}/py-packages"), PyPkgHandler, dict(need_automation_admin=True)),
            Rule(PathMatches(rf"{url}/py-packages/([\w,]+)?$"), PyPkgHandler, dict(need_automation_admin=True))
        ]
    )

def start_server():
    private_app = make_private_app()
    public_app = make_public_app()

    private_http_server = HTTPServer(private_app)
    private_port = os.getenv("API_SERVER_PORT", "8086")
    private_http_server.listen(int(private_port))

    public_http_server = HTTPServer(public_app, max_body_size=1<<31)
    public_port = os.getenv("API_SERVER_PUBLIC_PORT", "8087")
    public_http_server.listen(int(public_port))

    tornado.ioloop.IOLoop.current().start()