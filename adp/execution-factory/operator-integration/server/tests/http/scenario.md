# 基于场景测试流程（以外部接口为例）

## 注册算子后，更新算子状态，发布算子

### 注册算子
**请求参数**
```bash
curl -i -k -XPOST "https://10.4.175.99/api/agent-operator-integration/v1/operator/register" \
-H "Content-Type: application/json" \
-H "Authorization: Bearer ory_at_qZBAnm-wJBt0j2ntqRuQZrYU4j_5k5FnQ1XjXSVRXrQ.La48kfvZHZPTS_C13OOSxj7uJr4QVZaaUWUH4FUC5BQ" \
--data-binary @- <<EOF
{
  "data": $(cat /json/full_text_subdoc.json | jq -Rs .),
  "operator_metadata_type": "openapi"
}
EOF
```

**结果**
> 第一次注册算子时，会返回一个版本号，后续更新算子时，会返回一个新版本号
```json
HTTP/1.1 200 OK
Content-Type: application/json
Date: Fri, 18 Apr 2025 05:37:57 GMT
Content-Length: 137

[{"status":"success","operator_id":"11176d4f-bd5c-471d-9e80-93c5830b78f8","version":"71a889c5-b425-4d9e-93b6-d6b3230eb14b","error":null}]
```

### 查询发布算子信息
```bash
curl -i -k -XGET http://127.0.0.1:9000/api/agent-operator-integration/v1/operator/info/b2d8baf0-e31f-4cac-851d-30ad8c2e4722?version=e6ed8888-91fd-4b89-be6a-6dd46527d7f0 \
-H "Content-Type: application/json" \
-H "Authorization: Bearer ory_at_AuOrM3_F5TM7_XwdeojktaJyNDGmb0YJrdLsJioHMgg.-6w7cWnTApmetv2-dYSPKKKDLr2LPlQtJsWdbpfyxmw"
```

### 编辑未发布算子信息
**请求参数**
```bash
curl -i -k -XPOST "http://127.0.0.1:9000/api/agent-operator-integration/internal-v1/operator/info/update" \
-H "Content-Type: application/json" \
--data-binary @- <<EOF
{
  "data": $(cat /root/go/src/github.com/kweaver-ai/operator-hub/operator-integration/server/tests/file/json/file_decrypt.json | jq -Rs .),
  "operator_metadata_type": "openapi",
  "user_token": "ory_at_lFRsxTWm_AfurRpxjs5ECQg_g0IxQKyUEeBgK2Jl_DA.kOVvmulauj2h4GnF7q5sN9fkq85T__D8QjAlSQlUZdQ",
  "operator_id": "b2d8baf0-e31f-4cac-851d-30ad8c2e4722",
  "version": "e6ed8888-91fd-4b89-be6a-6dd46527d7f0"
}
EOF
```
**响应结果**
```json
HTTP/1.1 200 OK
Content-Type: application/json
Date: Fri, 18 Apr 2025 09:29:33 GMT
Content-Length: 200

[{"status":"success","operator_id":"b2d8baf0-e31f-4cac-851d-30ad8c2e4722","version":"e6ed8888-91fd-4b89-be6a-6dd46527d7f0","error":{"code":"","description":"","solution":"","link":"","details":null}}]
```

### 更改算子状态，发布算子
**请求参数**
```bash
curl -i -k -XPOST "http://127.0.0.1:9000/api/agent-operator-integration/v1/operator/status" \
-H "Content-Type: application/json" \
-H "Authorization: Bearer ory_at_lFRsxTWm_AfurRpxjs5ECQg_g0IxQKyUEeBgK2Jl_DA.kOVvmulauj2h4GnF7q5sN9fkq85T__D8QjAlSQlUZdQ" \
-d '[{"status": "published","operator_id": "b2d8baf0-e31f-4cac-851d-30ad8c2e4722","version": "e6ed8888-91fd-4b89-be6a-6dd46527d7f0"}]'
```



### 编辑已发布算子信息，产生新版本
```bash
curl -i -k -XPOST "http://127.0.0.1:9000/api/agent-operator-integration/internal-v1/operator/info/update" \
-H "Content-Type: application/json" \
--data-binary @- <<EOF
{
  "data": $(cat /root/go/src/github.com/kweaver-ai/operator-hub/operator-integration/server/tests/file/json/file_decrypt.json | jq -Rs .),
  "operator_metadata_type": "openapi",
  "user_token": "ory_at_R7KCiZTDj0rNQaD_D_lVCmtXN54uKF-YggP-qbsmP-I.C3LJUmrQBP0PXmMEdGxQHFgnhepbglfPIOfbnxDc_aE",
  "operator_id": "b2d8baf0-e31f-4cac-851d-30ad8c2e4722",
  "version": "e6ed8888-91fd-4b89-be6a-6dd46527d7f0",
  "extend_info": {"key":"test2"}
}
EOF
```
**响应结果**
```json
HTTP/1.1 200 OK
Content-Type: application/json
Date: Fri, 18 Apr 2025 10:43:19 GMT
Content-Length: 124

[{"status":"success","operator_id":"b2d8baf0-e31f-4cac-851d-30ad8c2e4722","version":"8cb1a81f-2ecd-4a3f-8db1-6c004db3bde5"}]
```

### 将编辑后的状态变更为已发布

**请求参数**
```bash
curl -i -k -XPOST "http://127.0.0.1:9000/api/agent-operator-integration/v1/operator/status" \
-H "Content-Type: application/json" \
-H "Authorization: Bearer ory_at_R7KCiZTDj0rNQaD_D_lVCmtXN54uKF-YggP-qbsmP-I.C3LJUmrQBP0PXmMEdGxQHFgnhepbglfPIOfbnxDc_aE" \
-d '[{"status": "published","operator_id": "b2d8baf0-e31f-4cac-851d-30ad8c2e4722","version": "8cb1a81f-2ecd-4a3f-8db1-6c004db3bde5"}]'
```

### 将最新草稿发布为正式版本，已发布的算子is_latest为false
>(数据库中已存在的算子f_is_latest=0)

**请求参数**
```bash
curl -i -k -XPOST "http://127.0.0.1:9000/api/agent-operator-integration/v1/operator/status" \
-H "Content-Type: application/json" \
-H "Authorization: Bearer ory_at_R7KCiZTDj0rNQaD_D_lVCmtXN54uKF-YggP-qbsmP-I.C3LJUmrQBP0PXmMEdGxQHFgnhepbglfPIOfbnxDc_aE" \
-d '[{"status": "published","operator_id": "b2d8baf0-e31f-4cac-851d-30ad8c2e4722","version": "416278e0-2816-4537-a974-fbe46a3a7720"}]'
```
**提示命名冲突**
```json
HTTP/1.1 409 Conflict
Content-Type: application/json
Date: Fri, 18 Apr 2025 10:46:45 GMT
Content-Length: 287

{"code":"Public.Conflict","description":"算子已存在","solution":"无","link":"无","details":{"operator_id":"b2d8baf0-e31f-4cac-851d-30ad8c2e4722","version":"416278e0-2816-4537-a974-fbe46a3a7720","name":"文件解密aaaa","error":"operator already exists，please change the name"}}
```

**发布成功**

```json
HTTP/1.1 200 OK
Content-Type: application/json
Date: Fri, 18 Apr 2025 10:49:38 GMT
Content-Length: 0
```
### 已发布算子下架
**请求参数**

```bash
curl -i -k -XPOST "http://127.0.0.1:9000/api/agent-operator-integration/v1/operator/status" \
-H "Content-Type: application/json" \
-H "Authorization: Bearer ory_at_97y8dO-8uUIg_ofpfBgT5go6l7WmIbgGHnoJtvRP-NA.Wz-4oRRTwLIE5cjTXtZtPcFlaC9rNYr8QTQR8MZ4Ed4" \
-d '[{"status": "offline","operator_id": "b2d8baf0-e31f-4cac-851d-30ad8c2e4722","version": "416278e0-2816-4537-a974-fbe46a3a7720"}]'
```

