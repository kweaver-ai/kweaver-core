import sys
from pathlib import Path

from fastapi import FastAPI

# 获取项目根目录（src 的父目录）
if hasattr(sys, "_MEIPASS"):
    ROOT_DIR = Path(sys._MEIPASS) / "src"
else:
    ROOT_DIR = Path(__file__).parent.parent

sys.path.append(str(ROOT_DIR))
print(sys.path)
print(ROOT_DIR)

from src.config import config
from src.interfaces.api.middleware import error_handler_middleware
from src.interfaces.api.routes import external_router, internal_router

app = FastAPI(
    title=config.get("app.name"),
    version=config.get("app.version"),
    debug=config.get("app.debug", False),
)

# 注册中间件
app.middleware("http")(error_handler_middleware)

# 注册路由
app.include_router(internal_router)
app.include_router(external_router)


@app.get("/")
async def root():
    return {"message": "Agent Memory Service"}


@app.get("/health")
async def health_check():
    return {"status": "ok"}


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(
        "src.main:app",
        host=config.get("server.host", "0.0.0.0"),
        port=config.get("server.port", 8000),
        workers=config.get("server.workers", 1),
        reload=config.get("app.debug", False),
    )
