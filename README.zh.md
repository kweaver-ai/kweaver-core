<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="./assets/logo/dark.png" />
    <source media="(prefers-color-scheme: light)" srcset="./assets/logo/light.png" />
    <img alt="KWeaver" src="./assets/logo/light.png" width="320" />
  </picture>
</p>

中文 | [English](README.md)

[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE.txt) [![skills.sh kweaver-core](https://img.shields.io/badge/skills.sh-kweaver--core-blue)](https://skills.sh/kweaver-ai/kweaver-sdk/kweaver-core) [![skills.sh create-bkn](https://img.shields.io/badge/skills.sh-create--bkn-green)](https://skills.sh/kweaver-ai/kweaver-sdk/create-bkn)

KWeaver Core 是面向企业决策智能体的治理优先（harness-first）基础平台。它将分散的数据、知识、工具和策略转化为受治理的上下文、安全的执行和可验证的反馈闭环。通过语义建模、实时访问、运行时管控和 TraceAI，帮助 AI 系统在复杂企业环境中可靠地推理、适应和行动。

> **注意：** KWeaver Core 是**纯后台框架**，不提供 Web 界面。所有交互通过 CLI、SDK 或 API 完成。如需界面访问，请安装 [**KWeaver DIP**](https://github.com/kweaver-ai/kweaver)。

## 📚 快速链接

- 🌐 [KWeaver DIP](https://dip-poc.aishu.cn/studio/agent/development/my-agent-list) - KWeaver Web 界面（用户名：`kweaver`，密码：`111111`）
- 🤝 [贡献指南](rules/CONTRIBUTING.zh.md) - 项目贡献指南
- 🚢 [部署指南](deploy/README.zh.md) - 一键部署到 Kubernetes
- 📘 [产品文档](help/) - 产品文档与使用指南
- 📝 [博客](https://kweaver-ai.github.io/kweaver-core/) - KWeaver 技术文章与更新
- 🚀 [发布规范](rules/RELEASE.zh.md) - 版本管理与发布流程
- 🏗️ [架构规范](rules/ARCHITECTURE.zh.md) - 架构设计规范
- 🧾 [版本发布](release-notes/) - 重要变更记录
- 📄 [许可证](LICENSE.txt) - Apache License 2.0
- 🐛 [报告 Bug](https://github.com/kweaver-ai/kweaver-core/issues) - 报告问题或 Bug
- 💡 [功能建议](https://github.com/kweaver-ai/kweaver-core/issues) - 提出新功能建议

## 🎬 演示视频

<div align="center">
<a href="https://www.bilibili.com/video/BV1nGXVBTEmo/?vd_source=4cdad687b2ac18a0b25e434f1fafe2f7" target="_blank">
<img src="./help/demo-cover.png" alt="KWeaver 演示视频" width="75%"/>
</a>

点击图片在 Bilibili 观看 KWeaver 演示视频
</div>

## 🚀 快速开始

1. 源码部署：参考 [部署文档](deploy/README.zh.md)。
2. 前置要求：参考 `deploy/README.zh.md` 中前置条件。
3. 执行安装部署脚本：

```bash
git clone https://github.com/kweaver-ai/kweaver-core.git
cd kweaver-core/deploy
chmod +x deploy.sh

# 完整一键部署（推荐）
./deploy.sh kweaver-core install     # 基础设施 + KWeaver 应用服务

# 查看帮助
./deploy.sh --help
```

4. 验证部署：

```bash
# 检查集群状态
kubectl get nodes
kubectl get pods -A

# 检查服务状态
./deploy.sh kweaver status
```

5. 验证 API 访问：

```bash
kweaver auth login https://<节点IP> -k
kweaver bkn list
```

> **尚未部署？** 可访问 [KWeaver DIP](https://dip-poc.aishu.cn/studio/agent/development/my-agent-list) Web 界面在线体验（用户名：`kweaver`，密码：`111111`），或将 CLI/SDK 直接连接到 Demo 环境（见下方说明）。

### 核心子系统

| 子项目 | 描述 | 仓库地址 |
| --- | --- | --- |
| **KWeaver SDK** | 面向 AI Agent 和开发者的 CLI 及 SDK（TypeScript/Python），用于以编程方式访问 KWeaver 知识网络与 Decision Agent | [kweaver-sdk](https://github.com/kweaver-ai/kweaver-sdk) |
| **KWeaver Core** | AI 原生平台底座 — Decision Agent、AI Data Platform（BKN Engine、VEGA Engine、Context Loader、Execution Factory）、Info Security Fabric、Trace AI | [kweaver-core](https://github.com/kweaver-ai/kweaver-core)（本仓库） |

## KWeaver SDK

[**kweaver-sdk**](https://github.com/kweaver-ai/kweaver-sdk) 通过 `kweaver` CLI 为 AI 智能体（Claude Code、GPT、自定义 Agent 等）提供对 KWeaver 知识网络与 Decision Agent 的访问能力，同时提供 Python 与 TypeScript SDK 用于编程集成。

### AI Agent Skills

从 [**kweaver-sdk**](https://github.com/kweaver-ai/kweaver-sdk) 使用 [`npx skills`](https://www.npmjs.com/package/skills) 安装技能。

**一次安装两个 skill**（推荐）：

```bash
npx skills add https://github.com/kweaver-ai/kweaver-sdk \
  --skill kweaver-core --skill create-bkn
```

- **`kweaver-core`** — 让 AI 编程助手掌握 KWeaver 的 API 与 CLI，可代替用户操作 KWeaver 平台。详见 [skills/kweaver-core/SKILL.md](https://github.com/kweaver-ai/kweaver-sdk/blob/main/skills/kweaver-core/SKILL.md)。
- **`create-bkn`** — 在 AI 编程助手中创建与管理**业务知识网络（BKN）**的流程与工具。详见 [skills/create-bkn/SKILL.md](https://github.com/kweaver-ai/kweaver-sdk/blob/main/skills/create-bkn/SKILL.md)。

**仅安装其中一个**（可选）：

```bash
npx skills add https://github.com/kweaver-ai/kweaver-sdk --skill kweaver-core
# 或
npx skills add https://github.com/kweaver-ai/kweaver-sdk --skill create-bkn
```

**使用任意 skill 前**，需先完成 KWeaver 实例认证：

```bash
npm install -g @kweaver-ai/kweaver-sdk
kweaver auth login https://your-kweaver-instance.com
```

> **自签名证书？** 如果你的实例使用了自签名或不受信任的 TLS 证书（新部署且未配置 CA 签发证书时很常见），添加 `-k` 参数跳过证书验证：
>
> ```bash
> kweaver auth login https://your-kweaver-instance.com -k
> ```

### 使用 Demo 环境快速体验

无需部署 — 将 AI Agent 直接连接到 Demo 环境即可立即上手（如需 Web 界面，请访问 [KWeaver DIP](https://dip-poc.aishu.cn/studio/agent/development/my-agent-list)）：

```bash
npx skills add https://github.com/kweaver-ai/kweaver-sdk \
  --skill kweaver-core --skill create-bkn

npm install -g @kweaver-ai/kweaver-sdk
kweaver auth login https://dip-poc.aishu.cn
```

然后在 AI 编程助手（Cursor、Claude Code 等）中通过自然语言提问：

```
列出所有知识网络
这个知识网络里面有什么
搜索业务知识网络中关于"供应链风险"的内容
查看两个客户的样例
有哪些 Agent
跟 Agent xxx 对话，问他"当前库存情况"
```

或使用 `/kweaver-core` 斜杠命令（skill 会自动接管）：

```
/kweaver-core 列出所有知识网络
/kweaver-core 这个知识网络里面有什么
/kweaver-core 搜索业务知识网络中关于"供应链风险"的内容
/kweaver-core 查看两个客户的样例
/kweaver-core 有哪些 Agent
/kweaver-core 跟 Agent xxx 对话，问他"当前库存情况"
```

> **Demo 账号**：用户名 `kweaver`，密码 `111111`

### 无浏览器环境登录（SSH、CI、容器等）

**npm 版** `kweaver` CLI 可在没有本地图形浏览器的环境完成 OAuth 登录：

1. 在**有浏览器**的机器上执行 `kweaver auth login https://你的实例`；登录成功后，从回调页复制一行命令，或执行 `kweaver auth export` / `kweaver auth export --json`。
2. 在**无头（headless）**机器上执行该命令，通过 `--client-id`、`--client-secret`、`--refresh-token` 换取令牌并写入 `~/.kweaver/`。

若已持有上述参数，也可在无头机器上直接执行：`kweaver auth login <url> --client-id … --client-secret … --refresh-token …`。

完整说明见 [kweaver-sdk — Headless / Server Authentication](https://github.com/kweaver-ai/kweaver-sdk/blob/main/packages/typescript/README.md#headless--server-authentication)（TypeScript 包 README）。Python 版 `kweaver` CLI 仍为交互式浏览器登录；可将 Node CLI 已完成登录的机器上的 `~/.kweaver/` 目录拷贝过来使用，或配置 `KWEAVER_BASE_URL` / `KWEAVER_TOKEN` 等环境变量（见 [kweaver-sdk 认证说明](https://github.com/kweaver-ai/kweaver-sdk#authentication)）。

### CLI

```bash
kweaver auth login https://your-kweaver.com     # 登录认证（自签名证书加 -k）
kweaver bkn list                                 # 浏览知识网络
kweaver bkn search <kn-id> "关键词"              # 知识网络语义搜索
kweaver bkn build <kn-id> --wait                # 重建索引并等待完成
kweaver bkn object-type list <kn-id>            # 查看对象类型
kweaver bkn action-type execute <kn-id> <at-id> # 执行 Action
kweaver agent list                               # 列出 Decision Agent
kweaver agent chat <agent-id> -m "你好"          # 与 Agent 对话
kweaver ds import-csv <ds-id> --files "*.csv"   # 将 CSV 文件导入数据源
kweaver context-loader kn-search "关键词"        # 通过 Context Loader 语义搜索
kweaver call /api/...                            # 原始 API 调用
```

### TypeScript & Python SDK

**简洁 API（推荐）：**

```typescript
import kweaver from "@kweaver-ai/kweaver-sdk/kweaver";
kweaver.configure({ config: true, bknId: "your-bkn-id", agentId: "your-agent-id" });

const results = await kweaver.search("供应链有哪些风险？");
const reply   = await kweaver.chat("总结前三大风险");
await kweaver.weaver({ wait: true });   // 重建 BKN 索引
```

```python
import kweaver
kweaver.configure(config=True, bkn_id="your-bkn-id", agent_id="your-agent-id")

results = kweaver.search("供应链有哪些风险？")
reply   = kweaver.chat("总结前三大风险")
```

**底层客户端（高级用法）：**

```typescript
import { KWeaverClient } from "@kweaver-ai/kweaver-sdk";
const client = new KWeaverClient();   // 读取 ~/.kweaver/ 凭据

const kns   = await client.knowledgeNetworks.list();
const reply = await client.agents.chat("agent-id", "你好");
await client.agents.stream("agent-id", "你好", {
  onTextDelta: (chunk) => process.stdout.write(chunk),
});
```

```python
from kweaver import KWeaverClient, ConfigAuth
client = KWeaverClient(auth=ConfigAuth())
kns  = client.knowledge_networks.list()
msg  = client.conversations.send_message("", "你好", agent_id="agent-id")
```

---

## KWeaver Core

**KWeaver Core** 是自主决策型 AI 原生平台底座。它位于 AI Agent（上层）与 AI/数据基础设施（下层）之间，以**业务知识网络（BKN）**为核心，为 Agent 提供统一的数据访问、执行与安全治理能力。

```text
            ┌─────────────────────────────────┐
            │     AI Agents（Decision Agent、   │
            │     Data Agent、HiAgent 等）      │
            └───────────────┬─────────────────┘
                            │
            ┌───────────────▼─────────────────┐
            │         业务知识网络              │
            │         KWeaver Core             │
            └───────────────┬─────────────────┘
                            │
            ┌───────────────▼─────────────────┐
            │   AI 基础设施 & 数据基础设施       │
            └─────────────────────────────────┘
```

KWeaver Core 解决专有数据与自主智能体结合时的两大核心技术痛点：

### 上下文工程 — 为 Agent 提供高质量上下文

在持续运行的 Agent 场景下，上下文不可避免地面临爆炸、腐烂、污染和高 Token 消耗问题。KWeaver Core 通过业务知识网络解决这些挑战：

- **上下文爆炸可收敛** — 多源候选先经 BKN 语义网络组织与聚合，再由 Context Loader 统一精排（召回→粗排→精排）仅保留关键证据与约束，避免海量片段直塞提示词导致决策失焦，整体准确率达到 **93%+**。
- **长上下文腐烂可缓解** — 以"实时事实 + 证据引用"替代长文堆叠，让推理围绕稳定对象展开，降低超长输入下的遗忘与幻觉风险，各类型场景下准确率相比同类型平台提升 **15%+**。
- **上下文污染可隔离** — 通过 BKN 网络准确构建企业数字孪生，把低可信内容、矛盾信息与潜在注入风险挡在知识与执行边界之外，保证推理链路干净可控。
- **Token 成本可压缩** — 将多源材料转为结构化对象信息按需获取而非全文拼接，在相同预算下提升信息密度，实现准确率提升的同时 Token 平均下降 **30%+**。

### 约束工程 — 安全可控的执行能力

Agent 不仅要"看得更全"，更要"做得更稳"。KWeaver Core 提供约束工程能力，实现企业级安全执行：

- **可解释决策** — 以"对象→动作→规则→约束"的知识结构表达业务意图，把工具调用与参数选择落到明确语义边界与规则依据上，解释清楚"为什么这么做"。
- **可追溯证据链** — 从行动意图→知识节点→数据来源→映射/算子→最终调用全链路留痕，支持按实体/关系回溯到源数据与生效规则，做到可审计、可复盘。
- **可管控执行闭环** — 统一身份与访问控制绑定到知识网络的对象/动作权限，执行前置校验、执行中策略拦截、执行后审计回写，实现"可授权、可批准、可收敛"的安全闭环。
- **可风险预防机制** — 将风险建模为"风险类"并与行动类关联，执行前做风险评估与仿真，命中阈值自动降级/阻断/二次确认，把高风险动作挡在执行之前。

### 核心架构

```text
┌──────────────────────── KWeaver Core ────────────────────────┐
│         │                                          │         │
│         │  Decision Agent                          │         │
│  信息    │  (Dolphin Runtime / Agent Executor)       │  Trace │
│  安全    │──────────────────────────────────────────│         │
│  编织    │                                          │   AI    │
│         │  AI Data Platform                        │         │
│ (ISF)   │  ┌────────────────────────────────────┐  │   可    │
│         │  │ Context Loader                     │  │   观    │
│  访问    │  │  ┌───────────┐   ┌─────────────┐  │  │   测    │
│  控制    │  │  │ Retrieval │ → │   Ranker    │  │  │    /   │
│   /     │  │  └───────────┘   └─────────────┘  │  │  证据   │
│  安全    │  ├────────────────────────────────────┤  │   链    │
│  管控    │  │ 业务知识网络                         │  │   追    │
│         │  │  ┌────────────────────────────┐    │  │   溯    │
│         │  │  │ BKN Engine                 │    │  │         │
│         │  │  │（数据 / 逻辑 / 风险 / 行动）  │    │  │         │
│         │  │  └──────┬─────────────┬───────┘    │  │         │
│         │  │         ↓ 映射         ↓ 映射       │  │         │
│         │  │  ┌────────┐ ┌───────────┐ ┌──────┐ │  │         │
│         │  │  │  VEGA  │ │ Execution │ │ Data │ │  │         │
│         │  │  │ Engine │ │  Factory  │ │ flow │ │  │         │
│         │  │  └────────┘ └───────────┘ └──────┘ │  │         │
│         │  └────────────────────────────────────┘  │         │
│         │                                          │         │
└─────────┴──────────────────────────────────────────┴─────────┘
               ↕                ↕                ↕
       多源 & 多模态数据（内置 30+ 开箱即用数据源）
```

| 组件 | 说明 |
| --- | --- |
| **AI Data Platform** | 采用非侵入式访问架构，通过业务知识网络，实现统一数据访问、统一执行、统一安全管控 |
| **Decision Agent** | 面向业务目标自主进行任务规划，基于 AI Data Platform 获取高质量上下文并进行有效的运行时管理，抑制幻觉和上下文腐烂等问题，在权限管控下调用工具与技能，形成安全的"推理→风险评估→执行→反馈"的自主业务闭环 |
| **Info Security Fabric** | 以身份、权限与策略为统一入口，对数据访问、模型输出与工具调用实施端到端控制与审计，降低越权、泄露与提示注入带来的安全风险，保障系统可管控 |
| **Trace AI** | 基于全链路可观测与证据链穿透，支持问题回放定位与自动优化建议，赋能 AI 应用实现可解释和可追溯 |

### BKN Lang

BKN Lang 是基于 Markdown 扩展语法的业务知识建模语言，人机双向友好：

- **易开发** — 基于通用 Markdown 语法，彻底消除代码壁垒。业务专家可通过所见即所得编辑器直接编写、阅读和修改业务逻辑定义，支持快速版本比对与协作审计，像修改文档一样修改系统规则。
- **易理解** — "对象类-关系类-风险类-行动类"四位一体模型完美映射企业业务模型。人类读懂业务含义，智能体实时"阅读"并解析出精准的上下文约束。逻辑显性化，拒绝性黑盒，从根本上降低大模型的推理幻觉与逻辑偏差。
- **易集成** — 定义仅作为全量文本存储于数据库特定字段，无复杂底层表结构强耦合。通过 Context Loader 按需动态加载，摒弃静态硬编码。跨系统、跨智能体高度兼容，作为轻量级资产在 AI Data Platform 中流畅流转。

### 基准测试与实验

查看更多：[KWeaver Blog](https://kweaver-ai.github.io/kweaver-core/)

#### 非结构化数据问答 — 跨平台横向对比

基于 145 个 HR 场景样本（简历库含 118 份多格式 PDF），覆盖简单信息查询、跨段落经验分析、多跳综合推理三类场景。各平台统一使用 DeepSeek V3.2 + BGE M3-Embedding，相同数据源，均采用 Agentic 模式测试。

| 指标 | KWeaver Core (v0.3.0) | BiSheng | Dify (v0.15.3) | RAGFlow (v0.17.0) |
| --- | --- | --- | --- | --- |
| **通过率** | **99.31%** (144/145) | 86.90% (126/145) | 96.55% (140/145) | 86.90% (126/145) |
| **平均响应时间** | 43.69s | **19.52s** | 63.82s | 71.56s |
| **P90 响应时间** | 56.92s | 32.53s | 79.15s | 95.37s |
| **平均 Token 消耗** | 21.36K | **4.98K** | 36.25K | 16.28K |

唯有 KWeaver Core 突破了传统 RAG 架构的"性能不可能三角"——在满足企业级严苛可靠性（>99% 准确率）的同时，将推理成本和时间延迟控制在高度生产可用的水平。Dify 属于"高消耗换取高通过率（力大砖飞）"；BiSheng 则是"牺牲准确推理换取表面上的快速低耗"；RAGFlow 在准确率和延迟上均落后。

#### 消融实验 — 关键技术杠杆

以下消融实验定位了 KWeaver Core 各组件的贡献度：

**检索深度** — 将检索 limit 从 10 提升至 20，通过率从 96.67% 提升至 100%，响应时间仅增加 +6.48s。得益于 Context Loader 的语义重排与上下文压缩能力，KWeaver Core 有效消化了扩大的上下文窗口，突破了传统 RAG 的"中间迷失"瓶颈。

**Schema 预加载** — 通过 Context Loader 提前预加载业务本体 Schema：

| 配置 | 通过率 | 平均推理步数 | 平均 Token 消耗 |
| --- | --- | --- | --- |
| 预加载 Schema | **100.0%** | **3.2 步** | **20.07K** |
| 不预加载 Schema | 93.33% | 5.8 步 | 38.54K |

Schema 为智能体提供了清晰的实体关系"地图"，推理步数平均减少 44.8%，Token 消耗降低 47.9%。

**工具治理** — 精准裁剪工具集优于提供全部可用工具：

| 配置 | 通过率 | 平均推理步数 | 平均 Token 消耗 |
| --- | --- | --- | --- |
| 完整工具集（6 个） | 75.0% | 7.1 步 | 42.3K |
| 精简工具集（3 个） | **100.0%** | **2.4 步** | **12.8K** |

提供全库搜索工具时，智能体倾向于选择"看起来功能更强大"的工具，但其返回结果噪声远高于限定范围的查询工具，触发多轮反思循环导致路径发散。精简后智能体被"约束"在正确路径上，实现一次通过。

**路径指引** — 将领域专家经验编码为可执行的推理路径模板：

| 配置 | 路径指引 | 工具数 | 通过率 | 平均响应时间 | 平均 Token |
| --- | --- | --- | --- | --- | --- |
| Explore-kn_search | 有 | 3 | **100.0%** | **37.82s** | **15,420** |
| 无路径，2 工具 | 无 | 2 | 100.0% | 53.06s | 23,287 |
| 无路径，3 工具 | 无 | 3 | 80.0% | 53.28s | 19,870 |

路径指引通过"告诉智能体怎么走"来提升效率，工具精简通过"减少可选的岔路"来保证稳定性。在生产环境中，两者的组合可以实现最优的性能表现。

#### 异构数据推理 — F1 Bench

F1 Bench 由 BIRD 测试集中 Formula-1 数据库混合 30 篇非结构化文档设计，测试智能体结合结构化+非结构化异构数据推理能力。

| 指标 | KWeaver Core | Dify 召回基线 |
| --- | --- | --- |
| **整体准确率** | **92.96%** | 78.87% |
| **SQL 浪费率** | **8.2%** | 24.5% |
| **SQL 命中效率** | **0.226** | 0.137 |
| **总 SQL 调用次数** | **292** | 408 |

### 核心价值总结

| 指标 | 数据 |
| --- | --- |
| **场景覆盖** | 问答、流程执行、情报分析、决策判断、探索类场景 |
| **TCO 降低** | 一体化平台降低建设成本 70% |
| **BKN 构建效率** | 业务知识网络构建效率提升 300% |
| **Token 成本** | 上下文优化与压缩，消耗降低 50% |

## 💬 交流社区

<div align="center">
<img src="./help/qrcode.png" alt="KWeaver 交流群二维码" width="30%"/>

扫码加入 KWeaver 交流群
</div>

## 支持与联系

- **贡献指南**: [贡献指南](rules/CONTRIBUTING.zh.md)
- **问题反馈**: [GitHub Issues](https://github.com/kweaver-ai/kweaver-core/issues)
- **许可证**: [Apache License 2.0](LICENSE.txt)

---

后续更多组件开源，敬请期待！
