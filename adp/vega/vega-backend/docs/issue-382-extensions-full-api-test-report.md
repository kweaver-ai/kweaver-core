# Issue #382 extensions（方案 B）全量 API 手工/自动联调报告

- 生成时间（UTC）: 2026-05-12T02:05:06.671211+00:00
- 服务根地址: `http://127.0.0.1:51203`
- 内网 API 前缀: `http://127.0.0.1:51203/api/vega-backend/in/v1`
- 鉴权: `AUTH_ENABLED=false`，请求头 `x-account-id` / `x-account-type`（内部接口）
- 报告文件: `D:\kweaver-core\adp\vega\vega-backend\docs\issue-382-extensions-full-api-test-report.md`

- 已 `go build` 并启动二进制（pid=24420），日志: `D:\kweaver-core\adp\vega\vega-backend\docs\.issue382-server.log`

## 用例执行明细


### C01 POST 创建 logical catalog（含 extensions）

#### C01 POST 创建 logical catalog（含 extensions）

- **HTTP** `POST http://127.0.0.1:51203/api/vega-backend/in/v1/catalogs`
- **状态码**: `201`

**请求体 JSON**

```json
{
  "name": "catalog-a-ext382-20260512020506",
  "tags": [
    "ext-test"
  ],
  "description": "Issue382 logical catalog A",
  "extensions": {
    "env": "prod",
    "tier": "gold"
  }
}
```

**响应体（原文）**

```
{"id":"d818lrsck8k5up394f4g"}
```


### C02 POST 创建 logical catalog B（不同 extension 便于筛选）

#### C02 POST 创建 logical catalog B（不同 extension 便于筛选）

- **HTTP** `POST http://127.0.0.1:51203/api/vega-backend/in/v1/catalogs`
- **状态码**: `201`

**请求体 JSON**

```json
{
  "name": "catalog-b-ext382-20260512020506",
  "tags": [
    "ext-test"
  ],
  "description": "Issue382 logical catalog B",
  "extensions": {
    "env": "staging",
    "tier": "silver"
  }
}
```

**响应体（原文）**

```
{"id":"d818lrsck8k5up394f50"}
```


### C03 POST 负例：保留前缀 vega_ 的 key

#### C03 POST 负例：保留前缀 vega_ 的 key

- **HTTP** `POST http://127.0.0.1:51203/api/vega-backend/in/v1/catalogs`
- **状态码**: `400`

**请求体 JSON**

```json
{
  "name": "catalog-c-ext382-20260512020506",
  "tags": [
    "bad"
  ],
  "description": "should fail",
  "extensions": {
    "vega_reserved": "x"
  }
}
```

**响应体（原文）**

```
{"error_code":"VegaBackend.Extensions.ReservedKey","description":"extensions 的 key 使用了保留前缀","solution":"请勿使用 vega_ 前缀的 key","error_link":"暂无","error_details":"extensions key 不能使用保留前缀 vega_"}
```


### C10 GET 列表（默认，不应依赖 include_extensions）

#### C10 GET 列表（默认，不应依赖 include_extensions）

- **HTTP** `GET http://127.0.0.1:51203/api/vega-backend/in/v1/catalogs?offset=0&limit=50&sort=update_time&direction=desc`
- **状态码**: `200`

**响应体（原文）**

```
{"entries":[{"id":"d818ir4ck8k1s15mc5bg","name":"catalog-a-ext382-20260512015851","tags":["ext-test"],"description":"Issue382 logical catalog A","type":"logical","enabled":false,"connector_type":"","connector_config":null,"metadata":null,"health_check_enabled":true,"health_check_status":"healthy","last_check_time":1778551148645,"health_check_result":"","creator":{"id":"e2e-ext382-20260512015851","type":"user","name":"e2e-ext382-20260512015851"},"create_time":1778551148645,"updater":{"id":"e2e-ext382-20260512015851","type":"user","name":"e2e-ext382-20260512015851"},"update_time":1778551148645,"operations":["view_detail","create","modify","delete","authorize","task_manage"]},{"id":"d818ir4ck8k1s15mc5c0","name":"catalog-b-ext382-20260512015851","tags":["ext-test"],"description":"Issue382 logical catalog B","type":"logical","enabled":false,"connector_type":"","connector_config":null,"metadata":null,"health_check_enabled":true,"health_check_status":"healthy","last_check_time":1778551148696,"health_check_result":"","creator":{"id":"e2e-ext382-20260512015851","type":"user","name":"e2e-ext382-20260512015851"},"create_time":1778551148696,"updater":{"id":"e2e-ext382-20260512015851","type":"user","name":"e2e-ext382-20260512015851"},"update_time":1778551148696,"operations":["view_detail","create","modify","delete","authorize","task_manage"]},{"id":"d818jscck8kark4dcqm0","name":"catalog-a-ext382-20260512020115","tags":["ext-test"],"description":"Issue382 logical catalog A","type":"logical","enabled":false,"connector_type":"","connector_config":null,"metadata":null,"health_check_enabled":true,"health_check_status":"healthy","last_check_time":1778551281583,"health_check_result":"","creator":{"id":"e2e-ext382-20260512020115","type":"user","name":"e2e-ext382-20260512020115"},"create_time":1778551281583,"updater":{"id":"e2e-ext382-20260512020115","type":"user","name":"e2e-ext382-20260512020115"},"update_time":1778551281583,"operations":["view_detail","create","modify","delete","authorize","task_manage"]},{"id":"d818jscck8kark4dcqmg","name":"catalog-b-ext382-20260512020115","tags":["ext-test"],"description":"Issue382 logical catalog B","type":"logical","enabled":false,"connector_type":"","connector_config":null,"metadata":null,"health_check_enabled":true,"health_check_status":"healthy","last_check_time":1778551281686,"health_check_result":"","creator":{"id":"e2e-ext382-20260512020115","type":"user","name":"e2e-ext382-20260512020115"},"create_time":1778551281686,"updater":{"id":"e2e-ext382-20260512020115","type":"user","name":"e2e-ext382-20260512020115"},"update_time":1778551281686,"operations":["view_detail","create","modify","delete","authorize","task_manage"]},{"id":"d818lrsck8k5up394f4g","name":"catalog-a-ext382-20260512020506","tags":["ext-test"],"description":"Issue382 logical catalog A","type":"logical","enabled":false,"connector_type":"","connector_config":null,"metadata":null,"health_check_enabled":true,"health_check_status":"healthy","last_check_time":1778551535150,"health_check_result":"","creator":{"id":"e2e-ext382-20260512020506","type":"user","name":"e2e-ext382-20260512020506"},"create_time":1778551535150,"updater":{"id":"e2e-ext382-20260512020506","type":"user","name":"e2e-ext382-20260512020506"},"update_time":1778551535150,"operations":["view_detail","create","modify","delete","authorize","task_manage"]},{"id":"d818lrsck8k5up394f50","name":"catalog-b-ext382-20260512020506","tags":["ext-test"],"description":"Issue382 logical catalog B","type":"logical","enabled":false,"connector_type":"","connector_config":null,"metadata":null,"health_check_enabled":true,"health_check_status":"healthy","last_check_time":1778551535266,"health_check_result":"","creator":{"id":"e2e-ext382-20260512020506","type":"user","name":"e2e-ext382-20260512020506"},"create_time":1778551535266,"updater":{"id":"e2e-ext382-20260512020506","type":"user","name":"e2e-ext382-20260512020506"},"update_time":1778551535266,"operations":["view_detail","create","modify","delete","authorize","task_manage"]}],"total_count":6}
```


### C11 GET 列表 include_extensions=true

#### C11 GET 列表 include_extensions=true

- **HTTP** `GET http://127.0.0.1:51203/api/vega-backend/in/v1/catalogs?offset=0&limit=50&include_extensions=true`
- **状态码**: `200`

**响应体（原文）**

```
{"entries":[{"id":"d818ir4ck8k1s15mc5bg","name":"catalog-a-ext382-20260512015851","tags":["ext-test"],"description":"Issue382 logical catalog A","type":"logical","enabled":false,"connector_type":"","connector_config":null,"metadata":null,"extensions":{"env":"prod","tier":"gold"},"health_check_enabled":true,"health_check_status":"healthy","last_check_time":1778551148645,"health_check_result":"","creator":{"id":"e2e-ext382-20260512015851","type":"user","name":"e2e-ext382-20260512015851"},"create_time":1778551148645,"updater":{"id":"e2e-ext382-20260512015851","type":"user","name":"e2e-ext382-20260512015851"},"update_time":1778551148645,"operations":["view_detail","create","modify","delete","authorize","task_manage"]},{"id":"d818ir4ck8k1s15mc5c0","name":"catalog-b-ext382-20260512015851","tags":["ext-test"],"description":"Issue382 logical catalog B","type":"logical","enabled":false,"connector_type":"","connector_config":null,"metadata":null,"extensions":{"env":"staging","tier":"silver"},"health_check_enabled":true,"health_check_status":"healthy","last_check_time":1778551148696,"health_check_result":"","creator":{"id":"e2e-ext382-20260512015851","type":"user","name":"e2e-ext382-20260512015851"},"create_time":1778551148696,"updater":{"id":"e2e-ext382-20260512015851","type":"user","name":"e2e-ext382-20260512015851"},"update_time":1778551148696,"operations":["view_detail","create","modify","delete","authorize","task_manage"]},{"id":"d818jscck8kark4dcqm0","name":"catalog-a-ext382-20260512020115","tags":["ext-test"],"description":"Issue382 logical catalog A","type":"logical","enabled":false,"connector_type":"","connector_config":null,"metadata":null,"extensions":{"tier":"gold","env":"prod"},"health_check_enabled":true,"health_check_status":"healthy","last_check_time":1778551281583,"health_check_result":"","creator":{"id":"e2e-ext382-20260512020115","type":"user","name":"e2e-ext382-20260512020115"},"create_time":1778551281583,"updater":{"id":"e2e-ext382-20260512020115","type":"user","name":"e2e-ext382-20260512020115"},"update_time":1778551281583,"operations":["view_detail","create","modify","delete","authorize","task_manage"]},{"id":"d818jscck8kark4dcqmg","name":"catalog-b-ext382-20260512020115","tags":["ext-test"],"description":"Issue382 logical catalog B","type":"logical","enabled":false,"connector_type":"","connector_config":null,"metadata":null,"extensions":{"env":"staging","tier":"silver"},"health_check_enabled":true,"health_check_status":"healthy","last_check_time":1778551281686,"health_check_result":"","creator":{"id":"e2e-ext382-20260512020115","type":"user","name":"e2e-ext382-20260512020115"},"create_time":1778551281686,"updater":{"id":"e2e-ext382-20260512020115","type":"user","name":"e2e-ext382-20260512020115"},"update_time":1778551281686,"operations":["view_detail","create","modify","delete","authorize","task_manage"]},{"id":"d818lrsck8k5up394f4g","name":"catalog-a-ext382-20260512020506","tags":["ext-test"],"description":"Issue382 logical catalog A","type":"logical","enabled":false,"connector_type":"","connector_config":null,"metadata":null,"extensions":{"env":"prod","tier":"gold"},"health_check_enabled":true,"health_check_status":"healthy","last_check_time":1778551535150,"health_check_result":"","creator":{"id":"e2e-ext382-20260512020506","type":"user","name":"e2e-ext382-20260512020506"},"create_time":1778551535150,"updater":{"id":"e2e-ext382-20260512020506","type":"user","name":"e2e-ext382-20260512020506"},"update_time":1778551535150,"operations":["view_detail","create","modify","delete","authorize","task_manage"]},{"id":"d818lrsck8k5up394f50","name":"catalog-b-ext382-20260512020506","tags":["ext-test"],"description":"Issue382 logical catalog B","type":"logical","enabled":false,"connector_type":"","connector_config":null,"metadata":null,"extensions":{"env":"staging","tier":"silver"},"health_check_enabled":true,"health_check_status":"healthy","last_check_time":1778551535266,"health_check_result":"","creator":{"id":"e2e-ext382-20260512020506","type":"user","name":"e2e-ext382-20260512020506"},"create_time":1778551535266,"updater":{"id":"e2e-ext382-20260512020506","type":"user","name":"e2e-ext382-20260512020506"},"update_time":1778551535266,"operations":["view_detail","create","modify","delete","authorize","task_manage"]}],"total_count":6}
```


### C12 GET 列表 include_extensions=true & include_extension_keys 投影

#### C12 GET 列表 include_extensions=true & include_extension_keys 投影

- **HTTP** `GET http://127.0.0.1:51203/api/vega-backend/in/v1/catalogs?offset=0&limit=50&include_extensions=true&include_extension_keys=env`
- **状态码**: `200`

**响应体（原文）**

```
{"entries":[{"id":"d818ir4ck8k1s15mc5bg","name":"catalog-a-ext382-20260512015851","tags":["ext-test"],"description":"Issue382 logical catalog A","type":"logical","enabled":false,"connector_type":"","connector_config":null,"metadata":null,"extensions":{"env":"prod"},"health_check_enabled":true,"health_check_status":"healthy","last_check_time":1778551148645,"health_check_result":"","creator":{"id":"e2e-ext382-20260512015851","type":"user","name":"e2e-ext382-20260512015851"},"create_time":1778551148645,"updater":{"id":"e2e-ext382-20260512015851","type":"user","name":"e2e-ext382-20260512015851"},"update_time":1778551148645,"operations":["view_detail","create","modify","delete","authorize","task_manage"]},{"id":"d818ir4ck8k1s15mc5c0","name":"catalog-b-ext382-20260512015851","tags":["ext-test"],"description":"Issue382 logical catalog B","type":"logical","enabled":false,"connector_type":"","connector_config":null,"metadata":null,"extensions":{"env":"staging"},"health_check_enabled":true,"health_check_status":"healthy","last_check_time":1778551148696,"health_check_result":"","creator":{"id":"e2e-ext382-20260512015851","type":"user","name":"e2e-ext382-20260512015851"},"create_time":1778551148696,"updater":{"id":"e2e-ext382-20260512015851","type":"user","name":"e2e-ext382-20260512015851"},"update_time":1778551148696,"operations":["view_detail","create","modify","delete","authorize","task_manage"]},{"id":"d818jscck8kark4dcqm0","name":"catalog-a-ext382-20260512020115","tags":["ext-test"],"description":"Issue382 logical catalog A","type":"logical","enabled":false,"connector_type":"","connector_config":null,"metadata":null,"extensions":{"env":"prod"},"health_check_enabled":true,"health_check_status":"healthy","last_check_time":1778551281583,"health_check_result":"","creator":{"id":"e2e-ext382-20260512020115","type":"user","name":"e2e-ext382-20260512020115"},"create_time":1778551281583,"updater":{"id":"e2e-ext382-20260512020115","type":"user","name":"e2e-ext382-20260512020115"},"update_time":1778551281583,"operations":["view_detail","create","modify","delete","authorize","task_manage"]},{"id":"d818jscck8kark4dcqmg","name":"catalog-b-ext382-20260512020115","tags":["ext-test"],"description":"Issue382 logical catalog B","type":"logical","enabled":false,"connector_type":"","connector_config":null,"metadata":null,"extensions":{"env":"staging"},"health_check_enabled":true,"health_check_status":"healthy","last_check_time":1778551281686,"health_check_result":"","creator":{"id":"e2e-ext382-20260512020115","type":"user","name":"e2e-ext382-20260512020115"},"create_time":1778551281686,"updater":{"id":"e2e-ext382-20260512020115","type":"user","name":"e2e-ext382-20260512020115"},"update_time":1778551281686,"operations":["view_detail","create","modify","delete","authorize","task_manage"]},{"id":"d818lrsck8k5up394f4g","name":"catalog-a-ext382-20260512020506","tags":["ext-test"],"description":"Issue382 logical catalog A","type":"logical","enabled":false,"connector_type":"","connector_config":null,"metadata":null,"extensions":{"env":"prod"},"health_check_enabled":true,"health_check_status":"healthy","last_check_time":1778551535150,"health_check_result":"","creator":{"id":"e2e-ext382-20260512020506","type":"user","name":"e2e-ext382-20260512020506"},"create_time":1778551535150,"updater":{"id":"e2e-ext382-20260512020506","type":"user","name":"e2e-ext382-20260512020506"},"update_time":1778551535150,"operations":["view_detail","create","modify","delete","authorize","task_manage"]},{"id":"d818lrsck8k5up394f50","name":"catalog-b-ext382-20260512020506","tags":["ext-test"],"description":"Issue382 logical catalog B","type":"logical","enabled":false,"connector_type":"","connector_config":null,"metadata":null,"extensions":{"env":"staging"},"health_check_enabled":true,"health_check_status":"healthy","last_check_time":1778551535266,"health_check_result":"","creator":{"id":"e2e-ext382-20260512020506","type":"user","name":"e2e-ext382-20260512020506"},"create_time":1778551535266,"updater":{"id":"e2e-ext382-20260512020506","type":"user","name":"e2e-ext382-20260512020506"},"update_time":1778551535266,"operations":["view_detail","create","modify","delete","authorize","task_manage"]}],"total_count":6}
```


### C13 GET 列表 extension 筛选 AND 单组（应命中 catalog A）

#### C13 GET 列表 extension 筛选 AND 单组（应命中 catalog A）

- **HTTP** `GET http://127.0.0.1:51203/api/vega-backend/in/v1/catalogs?offset=0&limit=50&extension_key=env&extension_value=prod&include_extensions=true`
- **状态码**: `200`

**响应体（原文）**

```
{"entries":[{"id":"d818ir4ck8k1s15mc5bg","name":"catalog-a-ext382-20260512015851","tags":["ext-test"],"description":"Issue382 logical catalog A","type":"logical","enabled":false,"connector_type":"","connector_config":null,"metadata":null,"extensions":{"env":"prod","tier":"gold"},"health_check_enabled":true,"health_check_status":"healthy","last_check_time":1778551148645,"health_check_result":"","creator":{"id":"e2e-ext382-20260512015851","type":"user","name":"e2e-ext382-20260512015851"},"create_time":1778551148645,"updater":{"id":"e2e-ext382-20260512015851","type":"user","name":"e2e-ext382-20260512015851"},"update_time":1778551148645,"operations":["view_detail","create","modify","delete","authorize","task_manage"]},{"id":"d818jscck8kark4dcqm0","name":"catalog-a-ext382-20260512020115","tags":["ext-test"],"description":"Issue382 logical catalog A","type":"logical","enabled":false,"connector_type":"","connector_config":null,"metadata":null,"extensions":{"env":"prod","tier":"gold"},"health_check_enabled":true,"health_check_status":"healthy","last_check_time":1778551281583,"health_check_result":"","creator":{"id":"e2e-ext382-20260512020115","type":"user","name":"e2e-ext382-20260512020115"},"create_time":1778551281583,"updater":{"id":"e2e-ext382-20260512020115","type":"user","name":"e2e-ext382-20260512020115"},"update_time":1778551281583,"operations":["view_detail","create","modify","delete","authorize","task_manage"]},{"id":"d818lrsck8k5up394f4g","name":"catalog-a-ext382-20260512020506","tags":["ext-test"],"description":"Issue382 logical catalog A","type":"logical","enabled":false,"connector_type":"","connector_config":null,"metadata":null,"extensions":{"env":"prod","tier":"gold"},"health_check_enabled":true,"health_check_status":"healthy","last_check_time":1778551535150,"health_check_result":"","creator":{"id":"e2e-ext382-20260512020506","type":"user","name":"e2e-ext382-20260512020506"},"create_time":1778551535150,"updater":{"id":"e2e-ext382-20260512020506","type":"user","name":"e2e-ext382-20260512020506"},"update_time":1778551535150,"operations":["view_detail","create","modify","delete","authorize","task_manage"]}],"total_count":3}
```


### C14 GET 列表 extension 筛选 AND 两组（应命中 catalog A）

#### C14 GET 列表 extension 筛选 AND 两组（应命中 catalog A）

- **HTTP** `GET http://127.0.0.1:51203/api/vega-backend/in/v1/catalogs?offset=0&limit=50&extension_key=env&extension_value=prod&extension_key=tier&extension_value=gold&include_extensions=true`
- **状态码**: `200`

**响应体（原文）**

```
{"entries":[{"id":"d818ir4ck8k1s15mc5bg","name":"catalog-a-ext382-20260512015851","tags":["ext-test"],"description":"Issue382 logical catalog A","type":"logical","enabled":false,"connector_type":"","connector_config":null,"metadata":null,"extensions":{"env":"prod","tier":"gold"},"health_check_enabled":true,"health_check_status":"healthy","last_check_time":1778551148645,"health_check_result":"","creator":{"id":"e2e-ext382-20260512015851","type":"user","name":"e2e-ext382-20260512015851"},"create_time":1778551148645,"updater":{"id":"e2e-ext382-20260512015851","type":"user","name":"e2e-ext382-20260512015851"},"update_time":1778551148645,"operations":["view_detail","create","modify","delete","authorize","task_manage"]},{"id":"d818jscck8kark4dcqm0","name":"catalog-a-ext382-20260512020115","tags":["ext-test"],"description":"Issue382 logical catalog A","type":"logical","enabled":false,"connector_type":"","connector_config":null,"metadata":null,"extensions":{"env":"prod","tier":"gold"},"health_check_enabled":true,"health_check_status":"healthy","last_check_time":1778551281583,"health_check_result":"","creator":{"id":"e2e-ext382-20260512020115","type":"user","name":"e2e-ext382-20260512020115"},"create_time":1778551281583,"updater":{"id":"e2e-ext382-20260512020115","type":"user","name":"e2e-ext382-20260512020115"},"update_time":1778551281583,"operations":["view_detail","create","modify","delete","authorize","task_manage"]},{"id":"d818lrsck8k5up394f4g","name":"catalog-a-ext382-20260512020506","tags":["ext-test"],"description":"Issue382 logical catalog A","type":"logical","enabled":false,"connector_type":"","connector_config":null,"metadata":null,"extensions":{"env":"prod","tier":"gold"},"health_check_enabled":true,"health_check_status":"healthy","last_check_time":1778551535150,"health_check_result":"","creator":{"id":"e2e-ext382-20260512020506","type":"user","name":"e2e-ext382-20260512020506"},"create_time":1778551535150,"updater":{"id":"e2e-ext382-20260512020506","type":"user","name":"e2e-ext382-20260512020506"},"update_time":1778551535150,"operations":["view_detail","create","modify","delete","authorize","task_manage"]}],"total_count":3}
```


### C15 GET 列表负例：extension_key/value 数量不成对

#### C15 GET 列表负例：extension_key/value 数量不成对

- **HTTP** `GET http://127.0.0.1:51203/api/vega-backend/in/v1/catalogs?offset=0&limit=50&extension_key=a&extension_key=b&extension_value=1`
- **状态码**: `400`

**响应体（原文）**

```
{"error_code":"VegaBackend.Extensions.MismatchedQueryPairs","description":"extension_key 与 extension_value 查询参数不合法","solution":"请成对传入且数量一致、非空","error_link":"暂无","error_details":"extension_key 与 extension_value 必须成对且数量一致"}
```


### C16 GET 列表负例：筛选组数 > 5

#### C16 GET 列表负例：筛选组数 > 5

- **HTTP** `GET http://127.0.0.1:51203/api/vega-backend/in/v1/catalogs?offset=0&limit=50&extension_key=k0&extension_value=v0&extension_key=k1&extension_value=v1&extension_key=k2&extension_value=v2&extension_key=k3&extension_value=v3&extension_key=k4&extension_value=v4&extension_key=k5&extension_value=v5`
- **状态码**: `400`

**响应体（原文）**

```
{"error_code":"VegaBackend.Extensions.TooManyFilterPairs","description":"扩展筛选条件组数超过上限","solution":"请减少筛选组数","error_link":"暂无","error_details":"扩展筛选条件最多 5 组"}
```


### C20 GET 单条 catalog（应含 extensions）

#### C20 GET 单条 catalog（应含 extensions）

- **HTTP** `GET http://127.0.0.1:51203/api/vega-backend/in/v1/catalogs/d818lrsck8k5up394f4g`
- **状态码**: `200`

**响应体（原文）**

```
{"entries":[{"id":"d818lrsck8k5up394f4g","name":"catalog-a-ext382-20260512020506","tags":["ext-test"],"description":"Issue382 logical catalog A","type":"logical","enabled":false,"connector_type":"","connector_config":null,"metadata":null,"extensions":{"env":"prod","tier":"gold"},"health_check_enabled":true,"health_check_status":"healthy","last_check_time":1778551535150,"health_check_result":"","creator":{"id":"e2e-ext382-20260512020506","type":"user","name":"e2e-ext382-20260512020506"},"create_time":1778551535150,"updater":{"id":"e2e-ext382-20260512020506","type":"user","name":"e2e-ext382-20260512020506"},"update_time":1778551535150,"operations":["view_detail","create","modify","delete","authorize","task_manage"]}]}
```


### C21 GET 批量 catalog（逗号分隔 ids）

#### C21 GET 批量 catalog（逗号分隔 ids）

- **HTTP** `GET http://127.0.0.1:51203/api/vega-backend/in/v1/catalogs/d818lrsck8k5up394f4g,d818lrsck8k5up394f50`
- **状态码**: `200`

**响应体（原文）**

```
{"entries":[{"id":"d818lrsck8k5up394f4g","name":"catalog-a-ext382-20260512020506","tags":["ext-test"],"description":"Issue382 logical catalog A","type":"logical","enabled":false,"connector_type":"","connector_config":null,"metadata":null,"extensions":{"env":"prod","tier":"gold"},"health_check_enabled":true,"health_check_status":"healthy","last_check_time":1778551535150,"health_check_result":"","creator":{"id":"e2e-ext382-20260512020506","type":"user","name":"e2e-ext382-20260512020506"},"create_time":1778551535150,"updater":{"id":"e2e-ext382-20260512020506","type":"user","name":"e2e-ext382-20260512020506"},"update_time":1778551535150,"operations":["view_detail","create","modify","delete","authorize","task_manage"]},{"id":"d818lrsck8k5up394f50","name":"catalog-b-ext382-20260512020506","tags":["ext-test"],"description":"Issue382 logical catalog B","type":"logical","enabled":false,"connector_type":"","connector_config":null,"metadata":null,"extensions":{"env":"staging","tier":"silver"},"health_check_enabled":true,"health_check_status":"healthy","last_check_time":1778551535266,"health_check_result":"","creator":{"id":"e2e-ext382-20260512020506","type":"user","name":"e2e-ext382-20260512020506"},"create_time":1778551535266,"updater":{"id":"e2e-ext382-20260512020506","type":"user","name":"e2e-ext382-20260512020506"},"update_time":1778551535266,"operations":["view_detail","create","modify","delete","authorize","task_manage"]}]}
```


### C22 PUT 更新 catalog：extensions 整包替换为单键

#### C22 PUT 更新 catalog：extensions 整包替换为单键

- **HTTP** `PUT http://127.0.0.1:51203/api/vega-backend/in/v1/catalogs/d818lrsck8k5up394f4g`
- **状态码**: `204`

**请求体 JSON**

```json
{
  "id": "d818lrsck8k5up394f4g",
  "name": "catalog-a-ext382-20260512020506",
  "tags": [
    "ext-test"
  ],
  "description": "Issue382 logical catalog A (updated)",
  "connector_type": "",
  "connector_config": {},
  "extensions": {
    "env": "prod-only"
  }
}
```

**响应体（原文）**

```

```


### C23 GET 单条 catalog 验证替换后 extensions

#### C23 GET 单条 catalog 验证替换后 extensions

- **HTTP** `GET http://127.0.0.1:51203/api/vega-backend/in/v1/catalogs/d818lrsck8k5up394f4g`
- **状态码**: `200`

**响应体（原文）**

```
{"entries":[{"id":"d818lrsck8k5up394f4g","name":"catalog-a-ext382-20260512020506","tags":["ext-test"],"description":"Issue382 logical catalog A (updated)","type":"logical","enabled":false,"connector_type":"","connector_config":null,"metadata":null,"extensions":{"env":"prod-only"},"health_check_enabled":true,"health_check_status":"healthy","last_check_time":1778551536473,"health_check_result":"","creator":{"id":"e2e-ext382-20260512020506","type":"user","name":"e2e-ext382-20260512020506"},"create_time":1778551535150,"updater":{"id":"e2e-ext382-20260512020506","type":"user","name":"e2e-ext382-20260512020506"},"update_time":1778551536473,"operations":["view_detail","create","modify","delete","authorize","task_manage"]}]}
```


### R01 POST 创建 table resource（根 extensions + Property.extensions）

#### R01 POST 创建 table resource（根 extensions + Property.extensions）

- **HTTP** `POST http://127.0.0.1:51203/api/vega-backend/in/v1/resources`
- **状态码**: `201`

**请求体 JSON**

```json
{
  "catalog_id": "d818lrsck8k5up394f4g",
  "name": "res-table-ext382-20260512020506",
  "tags": [
    "r-ext"
  ],
  "description": "resource ext test",
  "category": "table",
  "status": "active",
  "source_identifier": "src_ext382-20260512020506",
  "extensions": {
    "owner": "team-a",
    "cost_center": "cc99"
  },
  "schema_definition": [
    {
      "name": "id",
      "type": "string",
      "orig_type": "varchar",
      "display_name": "id",
      "original_name": "id",
      "description": "",
      "features": [],
      "attributes": {},
      "extensions": {
        "col_meta": "v1"
      }
    }
  ]
}
```

**响应体（原文）**

```
{"id":"d818ls4ck8k5up394f5g"}
```


### R02 POST 负例：根 extensions 保留前缀

#### R02 POST 负例：根 extensions 保留前缀

- **HTTP** `POST http://127.0.0.1:51203/api/vega-backend/in/v1/resources`
- **状态码**: `400`

**请求体 JSON**

```json
{
  "catalog_id": "d818lrsck8k5up394f4g",
  "name": "res-table-ext382-20260512020506-bad-root",
  "tags": [],
  "description": "",
  "category": "table",
  "status": "active",
  "source_identifier": "x",
  "extensions": {
    "vega_bad": "1"
  }
}
```

**响应体（原文）**

```
{"error_code":"VegaBackend.Extensions.ReservedKey","description":"extensions 的 key 使用了保留前缀","solution":"请勿使用 vega_ 前缀的 key","error_link":"暂无","error_details":"extensions key 不能使用保留前缀 vega_"}
```


### R03 POST 负例：Property.extensions 保留前缀

#### R03 POST 负例：Property.extensions 保留前缀

- **HTTP** `POST http://127.0.0.1:51203/api/vega-backend/in/v1/resources`
- **状态码**: `400`

**请求体 JSON**

```json
{
  "catalog_id": "d818lrsck8k5up394f4g",
  "name": "res-table-ext382-20260512020506-bad-prop",
  "tags": [],
  "description": "",
  "category": "table",
  "status": "active",
  "source_identifier": "y",
  "schema_definition": [
    {
      "name": "c1",
      "type": "string",
      "orig_type": "varchar",
      "display_name": "c1",
      "original_name": "c1",
      "description": "",
      "features": [],
      "attributes": {},
      "extensions": {
        "vega_x": "1"
      }
    }
  ]
}
```

**响应体（原文）**

```
{"error_code":"VegaBackend.Extensions.ReservedKey","description":"extensions 的 key 使用了保留前缀","solution":"请勿使用 vega_ 前缀的 key","error_link":"暂无","error_details":"extensions key 不能使用保留前缀 vega_"}
```


### R10 GET resources 列表（默认）

#### R10 GET resources 列表（默认）

- **HTTP** `GET http://127.0.0.1:51203/api/vega-backend/in/v1/resources?catalog_id=d818lrsck8k5up394f4g&offset=0&limit=50`
- **状态码**: `200`

**响应体（原文）**

```
{"total_count":1,"entries":[{"id":"d818ls4ck8k5up394f5g","catalog_id":"d818lrsck8k5up394f4g","name":"res-table-ext382-20260512020506","tags":["r-ext"],"description":"resource ext test","category":"table","status":"active","status_message":"","source_identifier":"src_ext382-20260512020506","creator":{"id":"e2e-ext382-20260512020506","type":"user","name":"e2e-ext382-20260512020506"},"create_time":1778551536690,"updater":{"id":"e2e-ext382-20260512020506","type":"user","name":"e2e-ext382-20260512020506"},"update_time":1778551536690,"operations":["view_detail","create","modify","delete","authorize","task_manage"]}]}
```


### R11 GET resources 列表 include_extensions=true

#### R11 GET resources 列表 include_extensions=true

- **HTTP** `GET http://127.0.0.1:51203/api/vega-backend/in/v1/resources?catalog_id=d818lrsck8k5up394f4g&offset=0&limit=50&include_extensions=true`
- **状态码**: `200`

**响应体（原文）**

```
{"entries":[{"id":"d818ls4ck8k5up394f5g","catalog_id":"d818lrsck8k5up394f4g","name":"res-table-ext382-20260512020506","tags":["r-ext"],"description":"resource ext test","category":"table","status":"active","status_message":"","source_identifier":"src_ext382-20260512020506","extensions":{"cost_center":"cc99","owner":"team-a"},"creator":{"id":"e2e-ext382-20260512020506","type":"user","name":"e2e-ext382-20260512020506"},"create_time":1778551536690,"updater":{"id":"e2e-ext382-20260512020506","type":"user","name":"e2e-ext382-20260512020506"},"update_time":1778551536690,"operations":["view_detail","create","modify","delete","authorize","task_manage"]}],"total_count":1}
```


### R12 GET resources 列表 extension 筛选（owner=team-a）

#### R12 GET resources 列表 extension 筛选（owner=team-a）

- **HTTP** `GET http://127.0.0.1:51203/api/vega-backend/in/v1/resources?catalog_id=d818lrsck8k5up394f4g&offset=0&limit=50&extension_key=owner&extension_value=team-a&include_extensions=true`
- **状态码**: `200`

**响应体（原文）**

```
{"entries":[{"id":"d818ls4ck8k5up394f5g","catalog_id":"d818lrsck8k5up394f4g","name":"res-table-ext382-20260512020506","tags":["r-ext"],"description":"resource ext test","category":"table","status":"active","status_message":"","source_identifier":"src_ext382-20260512020506","extensions":{"cost_center":"cc99","owner":"team-a"},"creator":{"id":"e2e-ext382-20260512020506","type":"user","name":"e2e-ext382-20260512020506"},"create_time":1778551536690,"updater":{"id":"e2e-ext382-20260512020506","type":"user","name":"e2e-ext382-20260512020506"},"update_time":1778551536690,"operations":["view_detail","create","modify","delete","authorize","task_manage"]}],"total_count":1}
```


### R13 GET resource 单条

#### R13 GET resource 单条

- **HTTP** `GET http://127.0.0.1:51203/api/vega-backend/in/v1/resources/d818ls4ck8k5up394f5g`
- **状态码**: `200`

**响应体（原文）**

```
{"entries":[{"id":"d818ls4ck8k5up394f5g","catalog_id":"d818lrsck8k5up394f4g","name":"res-table-ext382-20260512020506","tags":["r-ext"],"description":"resource ext test","category":"table","status":"active","status_message":"","source_identifier":"src_ext382-20260512020506","schema_definition":[{"name":"id","type":"string","orig_type":"varchar","display_name":"id","original_name":"id","description":"","features":[],"attributes":{},"extensions":{"col_meta":"v1"}}],"extensions":{"cost_center":"cc99","owner":"team-a"},"creator":{"id":"e2e-ext382-20260512020506","type":"user","name":"e2e-ext382-20260512020506"},"create_time":1778551536690,"updater":{"id":"e2e-ext382-20260512020506","type":"user","name":"e2e-ext382-20260512020506"},"update_time":1778551536690,"operations":["view_detail","create","modify","delete","authorize","task_manage"]}]}
```


### R14 PUT resource extensions 整包替换

#### R14 PUT resource extensions 整包替换

- **HTTP** `PUT http://127.0.0.1:51203/api/vega-backend/in/v1/resources/d818ls4ck8k5up394f5g`
- **状态码**: `204`

**请求体 JSON**

```json
{
  "catalog_id": "d818lrsck8k5up394f4g",
  "name": "res-table-ext382-20260512020506",
  "tags": [
    "r-ext"
  ],
  "description": "resource ext test (put)",
  "extensions": {
    "owner": "team-b"
  }
}
```

**响应体（原文）**

```

```


### R15 GET resource 验证 PUT 后 extensions

#### R15 GET resource 验证 PUT 后 extensions

- **HTTP** `GET http://127.0.0.1:51203/api/vega-backend/in/v1/resources/d818ls4ck8k5up394f5g`
- **状态码**: `200`

**响应体（原文）**

```
{"entries":[{"id":"d818ls4ck8k5up394f5g","catalog_id":"d818lrsck8k5up394f4g","name":"res-table-ext382-20260512020506","tags":["r-ext"],"description":"resource ext test (put)","category":"table","status":"active","status_message":"","source_identifier":"src_ext382-20260512020506","schema_definition":[{"name":"id","type":"string","orig_type":"varchar","display_name":"id","original_name":"id","description":"","features":[],"attributes":{},"extensions":{"col_meta":"v1"}}],"extensions":{"owner":"team-b"},"creator":{"id":"e2e-ext382-20260512020506","type":"user","name":"e2e-ext382-20260512020506"},"create_time":1778551536690,"updater":{"id":"e2e-ext382-20260512020506","type":"user","name":"e2e-ext382-20260512020506"},"update_time":1778551537471,"operations":["view_detail","create","modify","delete","authorize","task_manage"]}]}
```


### R90 DELETE resource

#### R90 DELETE resource

- **HTTP** `DELETE http://127.0.0.1:51203/api/vega-backend/in/v1/resources/d818ls4ck8k5up394f5g`
- **状态码**: `500`

**响应体（原文）**

```
{"error_code":"VegaBackend.Resource.InternalError.GetFailed","description":"获取数据资源失败","solution":"请联系管理员","error_link":"暂无","error_details":"Error 1054 (42S22): Unknown column 'f_creator' in 'SELECT'"}
```


### C90 DELETE catalogs（A,B）

#### C90 DELETE catalogs（A,B）

- **HTTP** `DELETE http://127.0.0.1:51203/api/vega-backend/in/v1/catalogs/d818lrsck8k5up394f4g,d818lrsck8k5up394f50`
- **状态码**: `204`

**响应体（原文）**

```

```


## 说明（未覆盖的 OpenAPI 路径）


- `GET /catalogs/{ids}/resources` 在 OpenAPI 中作为便利视图描述；当前 `router.go` **未注册**该嵌套路由，等价能力为 `GET /resources?catalog_id=...`（本报告 R10–R12 已覆盖 extensions 相关 query）。
- `POST /catalogs/{id}/discover`、`test-connection`、`health-status` 与 extensions 无直接耦合，未纳入本次「extensions 功能」必测范围。
