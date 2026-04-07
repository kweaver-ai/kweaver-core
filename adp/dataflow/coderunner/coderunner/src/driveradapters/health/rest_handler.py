from tornado.web import RequestHandler

__all__ = [
    'HealthHandler',
    'AliveHandler'
]


class HealthHandler(RequestHandler):

    async def get(self):
        return self.write("health")


class AliveHandler(RequestHandler):

    async def get(self):
        return self.write("alive")