import os
import urllib3
import tornado
import tornado.ioloop

from tornado.options import options
from tornado.httpserver import HTTPServer
from tornado.web import Application
from tornado.routing import (Rule, PathMatches)

from driveradapters.health.rest_handler import (HealthHandler, AliveHandler)
from driveradapters.pypkg.rest_handler import PyPkgHandler
from driveradapters.runner.rest_handler import (RunnerHandler, AsyncRunnerHandler)
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

def make_app():
    url = "/api/coderunner/v1"
    return Application(
        [
            Rule(PathMatches(rf"{url}/health/ready/?$"), HealthHandler),
            Rule(PathMatches(rf"{url}/health/alive/?$"), AliveHandler),
            Rule(PathMatches(rf"{url}/pycode/run-by-params?$"), RunnerHandler),
            Rule(PathMatches(rf"{url}/pycode/async-run-by-params?$"), AsyncRunnerHandler),
            Rule(PathMatches(rf"{url}/py-packages/([\w,]+)?$"), PyPkgHandler),
        ]
    )


def start_server():
    application = make_app()
    http_server = HTTPServer(application)
    port = os.getenv("API_SERVER_PORT", "8085")
    http_server.listen(int(port))
    tornado.ioloop.IOLoop.current().start()