# Agent Factory API 文档

这个目录只保留对外可查看的最终 OpenAPI 文档产物。

## 目录里的文件

- `agent-factory.json`
  - OpenAPI 3 JSON 版本，适合工具接入
- `agent-factory.yaml`
  - OpenAPI 3 YAML 版本，适合人工阅读和分发
- `agent-factory.html`
  - Scalar 风格的静态文档页面，适合 Try it Out
- `agent-factory-redoc.html`
  - Redoc 风格的静态文档页面，适合只读展示
- `favicon.png`
  - 静态文档页面图标
- `ui/`
  - 本地化后的 JS 资源目录，供静态 HTML 离线加载

## 如何生成

在 `agent-backend/agent-factory` 目录下执行：

```bash
make gen-api-docs
```

## 如何校验

```bash
make validate-api-docs
```

这个命令会同时检查：

- OpenAPI 结构是否合法
- 路径数、接口数是否符合预期
- 这里的公共文档是否与运行时副本一致

## v3 Agent Config 模式说明

`/api/agent-factory/v3/agent` 的 `config` 现在以 `mode` 作为主模式字段：

- `default`
  - 使用 `system_prompt`、`plan_mode`
- `dolphin`
  - 使用 `dolphin`
  - `is_dolphin_mode` 仍保留，用于兼容旧请求
  - 新增 `is_use_tool_id_in_dolphin`
- `react`
  - 使用 `system_prompt`、`react_config`、`plan_mode`

对外文档与详情响应统一使用 `react_config`。

## 如何查看

### 直接查看静态页面

直接打开当前目录下的任一页面：

- `agent-factory.html`
- `agent-factory-redoc.html`

这两个页面都会加载当前目录下的 `ui/` 本地资源，不依赖外部 CDN。

### 启动服务后查看

启动 Agent Factory 后访问：

- `http://127.0.0.1:30777/scalar/index.html`
- `http://127.0.0.1:30777/redoc/index.html`
- `http://127.0.0.1:30777/scalar/doc.json`
- `http://127.0.0.1:30777/scalar/doc.yaml`

推荐方式：

- 需要发起请求、Try it Out：使用 Scalar 页面
- 需要更好的只读文档展示：使用 Redoc 页面

## 深入维护说明

如果你需要了解生成链路、overlay、baseline 或 Swagger 中间产物，请查看：

- `../../cmd/openapi-docs/README.simple.md`
- `../../cmd/openapi-docs/docs/OPENAPI_AUTOMATION_GUIDE.md`
