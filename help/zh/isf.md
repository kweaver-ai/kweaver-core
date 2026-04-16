# Info Security Fabric（ISF）

## 概述

**Info Security Fabric** 是**横切的安全层**：在数据访问、模型输出与工具调用上提供统一的**身份**、**权限**、**策略**与**审计**。完整安装可能对接 OAuth2/OIDC（如 Hydra）与业务域服务。

使用 **`--minimum` 安装**时，多数认证组件关闭，便于实验环境快速上手，部分 API 可能无需 Token。生产环境请按 [deploy/README.zh.md](../../deploy/README.zh.md) 启用完整认证配置。

**相关模块：** 所有接受 `Authorization` 的子系统；主要消费者包括 [Decision Agent](decision-agent.md)、[VEGA 引擎](vega.md)。

### CLI

#### 登录

```bash
# 标准登录（跳过 TLS 证书验证）
kweaver auth login https://kweaver.example.com -k

# 登录并设置别名，方便多环境切换
kweaver auth login https://kweaver.example.com --alias prod -k

# 使用用户名密码直接登录（非交互式）
kweaver auth login https://kweaver.example.com \
  -u admin -p 'MySecurePassword' -k

# 最小化安装时跳过认证
kweaver auth login https://localhost:30000 --no-auth -k

# 使用 Playwright 浏览器完成 OAuth 登录流程
kweaver auth login https://kweaver.example.com --playwright -k

# 自定义本地回调端口与重定向 URI
kweaver auth login https://kweaver.example.com \
  --port 8765 \
  --redirect-uri http://localhost:8765/callback \
  -k
```

#### 会话管理

```bash
# 列出所有已保存的登录会话
kweaver auth list

# 切换当前使用的会话（按别名）
kweaver auth use prod

# 列出当前会话下的用户列表
kweaver auth users

# 切换到不同用户
kweaver auth switch --user analyst@example.com

# 查看当前登录身份
kweaver auth whoami

# 查看当前会话的详细状态（Token 有效期、刷新状态等）
kweaver auth status
```

**`auth whoami` 与 no-auth**：`whoami` 需 OAuth 登录写入的 `id_token`。若会话为 **`auth login … --no-auth`** 或平台关闭鉴权，CLI 为 **no-auth**，`whoami` 会报错提示无 `id_token`，属正常；请用 `auth status` 确认模式，勿与登录失败混淆。

```bash
# 导出当前会话的 Token（用于脚本或 CI/CD）
kweaver auth export
```

在已登录会话下，REST 调用可直接使用 **`kweaver token`**（与 `kweaver auth export` 均可得到 Bearer 串；示例优先用前者）：

```bash
curl -sk "https://<访问地址>/api/agent-factory/v1/agents" \
  -H "Authorization: Bearer $(kweaver token)"
```

#### 登出与删除

```bash
# 登出当前会话（Token 失效，本地凭据保留）
kweaver auth logout

# 删除已保存的会话（同时清理本地凭据）
kweaver auth delete prod
```

#### 多账户工作流

```bash
# 1. 登录生产环境
kweaver auth login https://prod.kweaver.example.com --alias prod -k

# 2. 登录开发环境
kweaver auth login https://dev.kweaver.example.com --alias dev -k

# 3. 查看所有会话
kweaver auth list

# 4. 切换到生产环境
kweaver auth use prod

# 5. 确认身份
kweaver auth whoami

# 6. 在生产环境操作
kweaver agent list --limit 5

# 7. 切换到开发环境
kweaver auth use dev

# 8. 在开发环境操作
kweaver agent list --limit 5
```

#### 配置与业务域

```bash
# 显示当前完整配置
kweaver config show

# 列出所有已配置的业务域
kweaver config list-bd

# 设置当前业务域
kweaver config set-bd bd_sales
```

**`config list-bd` / `config set-bd` 与最小化安装**：**`--minimum` / 最小化安装** **不包含**这两条子命令依赖的**业务域管理服务**（未随最小化部署），`list-bd` 常 **404** 等，属部署裁剪，不是 CLI 故障。平台仍有默认业务域，请用 `config show` 查看。**完整安装**下再用 `list-bd` / `set-bd` 枚举或切换域；若仍失败，再查网关或相关服务。

**业务域优先级说明**：当设置了业务域后，所有 API 调用会在请求头中携带 `X-Business-Domain` 字段。平台根据此字段进行数据隔离与权限控制。优先级为：命令行 `--bd` 参数 > `kweaver config set-bd` 配置 > 默认业务域。

```bash
# 命令级覆盖业务域
kweaver agent list --bd bd_finance

# 查看当前生效的业务域配置
kweaver config show | grep business_domain
```

#### 端到端流程

```bash
# 1. 首次登录
kweaver auth login https://kweaver.example.com --alias prod -k -u admin -p secret

# 2. 确认身份
kweaver auth whoami

# 3. 设置业务域
kweaver config set-bd bd_sales

# 4. 开始使用平台功能
kweaver bkn list --limit 5
kweaver agent list --limit 5

# 5. 会话结束后登出
kweaver auth logout
```

---

### Python SDK

```python
from kweaver_sdk import KWeaverClient

client = KWeaverClient(base_url="https://<访问地址>")

client.auth.login(username="admin", password="MySecurePassword")

user = client.auth.whoami()
print(f"用户: {user['username']}")
print(f"角色: {user['roles']}")
print(f"业务域: {user.get('business_domain', '默认')}")

status = client.auth.status()
print(f"Token 有效: {status['token_valid']}")
print(f"过期时间: {status['expires_at']}")
print(f"刷新 Token: {'可用' if status['refresh_available'] else '不可用'}")

token = client.auth.export_token()
print(f"Bearer Token: {token[:20]}...")

agents = client.agent.list(limit=5)
for agt in agents["data"]:
    print(agt["id"], agt["name"])

client.config.set_business_domain("bd_sales")

agents_sales = client.agent.list(limit=5)
for agt in agents_sales["data"]:
    print(agt["id"], agt["name"])

client.auth.logout()

client_noauth = KWeaverClient(
    base_url="https://localhost:30000",
    skip_auth=True,
    verify_ssl=False
)
networks = client_noauth.bkn.list_networks()
print(f"知识网络数: {len(networks['data'])}")
```

---

### TypeScript SDK

```typescript
import { KWeaverClient } from '@kweaver-ai/kweaver-sdk';

const client = new KWeaverClient({ baseUrl: 'https://<访问地址>' });

await client.auth.login({ username: 'admin', password: 'MySecurePassword' });

const user = await client.auth.whoami();
console.log('用户:', user.username);
console.log('角色:', user.roles);
console.log('业务域:', user.businessDomain ?? '默认');

const status = await client.auth.status();
console.log('Token 有效:', status.tokenValid);
console.log('过期时间:', status.expiresAt);

const token = await client.auth.exportToken();
console.log('Bearer Token:', token.slice(0, 20) + '...');

const agents = await client.agent.list({ limit: 5 });
agents.data.forEach((agt) => console.log(agt.id, agt.name));

client.config.setBusinessDomain('bd_sales');

const agentsSales = await client.agent.list({ limit: 5 });
agentsSales.data.forEach((agt) => console.log(agt.id, agt.name));

await client.auth.logout();

const clientNoAuth = new KWeaverClient({
  baseUrl: 'https://localhost:30000',
  skipAuth: true,
  verifySsl: false,
});
const networks = await clientNoAuth.bkn.listNetworks();
console.log('知识网络数:', networks.data.length);
```

---

### curl

```bash
# OAuth2 Token 获取（密码模式，适用于启用完整认证的环境）
curl -sk -X POST "https://<访问地址>/oauth2/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=password&username=admin&password=MySecurePassword&client_id=kweaver-cli&scope=openid"

# 使用 Token 访问受保护资源
curl -sk "https://<访问地址>/api/agent-factory/v1/agents" \
  -H "Authorization: Bearer $(kweaver token)"

# 查看当前用户信息
curl -sk "https://<访问地址>/api/isf/v1/userinfo" \
  -H "Authorization: Bearer $(kweaver token)"

# 发现 OpenID 配置
curl -sk "https://<访问地址>/.well-known/openid-configuration"

# Token 内省（检查 Token 有效性）
curl -sk -X POST "https://<访问地址>/oauth2/introspect" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "token=<access-token>&client_id=kweaver-cli"

# 刷新 Token
curl -sk -X POST "https://<访问地址>/oauth2/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=refresh_token&refresh_token=<refresh-token>&client_id=kweaver-cli"

# 最小化安装 — 无需 Token 直接访问
curl -sk "https://localhost:30000/api/agent-factory/v1/agents"
```
