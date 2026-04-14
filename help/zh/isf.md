# Info Security Fabric（ISF）

## 概述

**Info Security Fabric** 是**横切的安全层**：在数据访问、模型输出与工具调用上提供统一的**身份**、**权限**、**策略**与**审计**。完整安装可能对接 OAuth2/OIDC（如 Hydra）与业务域服务。

使用 **`--minimum` 安装**时，多数认证组件关闭，便于实验环境快速上手，部分 API 可能无需 Token。生产环境请按 [deploy/README.zh.md](../../deploy/README.zh.md) 启用完整认证配置。

**相关模块：** 所有接受 `Authorization` 的子系统；主要消费者包括 [Decision Agent](decision-agent.md)、[VEGA 引擎](vega.md)。

## 使用方式

```bash
export KWEAVER_BASE="https://<访问地址>"
export TOKEN="<bearer-token>"   # auth.enabled=false 时常可省略
```

### CLI

```bash
kweaver auth login https://<访问地址> -k
kweaver auth --help
# Token 会保存在本地供后续 CLI 调用使用
```

### Python SDK

```python
from kweaver_sdk import KWeaverClient

client = KWeaverClient(base_url="https://<访问地址>")
# client.auth.login(username="...", password="...")
# 后续请求自动附带会话令牌
```

### TypeScript SDK

```typescript
import { KWeaverClient } from '@kweaver-ai/kweaver-sdk';

const client = new KWeaverClient({ baseUrl: 'https://<访问地址>' });
// await client.auth.login({ username: '...', password: '...' });
```

### curl

```bash
# OAuth2 启用时，令牌端点与客户端凭据取决于 IdP 配置。
# 示例：使用 Bearer 访问受保护资源
curl -sk "$KWEAVER_BASE/api/agent-factory/v1/agents" \
  -H "Authorization: Bearer $TOKEN"

# 发现 OpenID 配置（路径因部署而异）
curl -sk "$KWEAVER_BASE/.well-known/openid-configuration" || true
```
