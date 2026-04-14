# Info Security Fabric (ISF)

## Overview

The **Info Security Fabric** is the **cross-cutting security layer**: unified **identity**, **permissions**, **policies**, and **audit** across data access, model output, and tool invocation. In full installs it may integrate with OAuth2/OIDC stacks (e.g. Hydra) and business-domain services.

With **`--minimum` install**, many auth components are disabled for a simpler lab setup — APIs may not require tokens. For production, enable the full auth profile per [deploy/README.md](../../deploy/README.md).

**Related modules:** All subsystems that accept `Authorization` headers; [Decision Agent](decision-agent.md) and [VEGA Engine](vega.md) are primary consumers.

## Usage

```bash
export KWEAVER_BASE="https://<access-address>"
export TOKEN="<bearer-token>"   # often unused when auth.enabled=false
```

### CLI

```bash
kweaver auth login https://<access-address> -k
kweaver auth --help
# Token is stored for subsequent CLI calls
```

### Python SDK

```python
from kweaver_sdk import KWeaverClient

client = KWeaverClient(base_url="https://<access-address>")
# client.auth.login(username="...", password="...")
# Subsequent requests attach the session token automatically
```

### TypeScript SDK

```typescript
import { KWeaverClient } from '@kweaver-ai/kweaver-sdk';

const client = new KWeaverClient({ baseUrl: 'https://<access-address>' });
// await client.auth.login({ username: '...', password: '...' });
```

### curl

```bash
# When OAuth2 is enabled, token endpoint and client credentials depend on your IdP config.
# Illustrative resource call with a bearer token:
curl -sk "$KWEAVER_BASE/api/agent-factory/v1/agents" \
  -H "Authorization: Bearer $TOKEN"

# Discover OpenID configuration (path varies by deployment)
curl -sk "$KWEAVER_BASE/.well-known/openid-configuration" || true
```
