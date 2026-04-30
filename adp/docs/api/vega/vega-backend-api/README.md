# Vega Backend API 文档

> Vega Backend HTTP API 的 OpenAPI 3.1.1 定义。每个文件聚焦一个资源概念；跨资源的"便利端点"按"按谁过滤就归谁"的原则归属到对应资源文件，避免 schema 重复定义和文档撕裂。

## 文件索引

| 文件 | 资源 | 包含的端点 |
|---|---|---|
| [catalog.yaml](catalog.yaml) | Catalog | `GET/POST /catalogs`、`GET/PUT/DELETE /catalogs/{id(s)}`、`GET /catalogs/{ids}/health-status`、`POST /catalogs/{id}/test-connection` |
| [connector-type.yaml](connector-type.yaml) | ConnectorType | `GET/POST /connector-types`、`GET/PUT/DELETE /connector-types/{type}`、`POST .../enable`、`POST .../disable` |
| [discover-task.yaml](discover-task.yaml) | DiscoverTask | `POST /catalogs/{cid}/discover`（手动触发）、`GET /discover-tasks`、`GET /discover-tasks/{id}` |
| [discover-schedule.yaml](discover-schedule.yaml) | DiscoverSchedule | `GET/POST /discover-schedules`、`GET/PUT/DELETE /discover-schedules/{sid}`、`POST .../enable`、`POST .../disable`、`GET /catalogs/{cid}/discover-schedules`（便利视图） |
| [resource.yaml](resource.yaml) | Resource | `GET/POST /resources`、`GET/PUT/DELETE /resources/{id(s)}`、`GET /catalogs/{ids}/resources`（便利视图）、其它 dataset / buildtask 子端点 |
| [query.yaml](query.yaml) | Query | `POST /query/execute`（结构化查询） |
| [query-endpoints-comparison.md](query-endpoints-comparison.md) | — | 设计参考：查询相关端点的对比与归属说明（非 OpenAPI 文档） |

## 约定

- **OpenAPI 版本**：3.1.1。
- **错误响应**：所有非 2xx 响应统一返回 `Error` schema（对应 `kweaver-go-lib/rest.BaseError`），各文件自含一份定义。
- **内部接口**：`/api/vega-backend/in/v1/...` 与 `/api/vega-backend/v1/...` 一一对应，请求 / 响应结构完全一致；区别仅在鉴权方式（外部 OAuth Token，内部 Header `X-Account-ID` / `X-Account-Type`）。各文件仅描述外部接口，内部接口在 `info.description` 中统一说明。
- **跨资源便利端点**：归属到"被过滤的目标资源"所在的文件。例如 `GET /catalogs/{cid}/discover-schedules` 是按 catalog 过滤 schedule，归 `discover-schedule.yaml`；`POST /catalogs/{cid}/discover` 创建 DiscoverTask 实例，归 `discover-task.yaml`。

## 待生成

以下文件尚未落地，待对应资源 / 设计稿评审通过后补全：

- `discover-task.yaml`
- `discover-schedule.yaml`（依赖 [discover_schedule_redesign.md](../../../design/vega/features/vega-backend/discover_schedule_redesign.md) 的重构落地）
- `resource.yaml`
