<p align="center">
  <img alt="KWeaver" src="./assets/logo/banner.png" width="800" />
</p>

中文 | [English](README.md)

[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE.txt) [![skills.sh kweaver-core](https://img.shields.io/badge/skills.sh-kweaver--core-blue)](https://skills.sh/kweaver-ai/kweaver-sdk/kweaver-core) [![skills.sh create-bkn](https://img.shields.io/badge/skills.sh-create--bkn-green)](https://skills.sh/kweaver-ai/kweaver-sdk/create-bkn)

KWeaver Core 是面向企业决策智能体的治理优先（harness-first）基础平台。它将分散的数据、知识、工具和策略转化为受治理的上下文、安全的执行和可验证的反馈闭环。通过语义建模、实时访问、运行时管控和 TraceAI，帮助 AI 系统在复杂企业环境中可靠地推理、适应和行动。

**本文目录：** [📚 快速链接](#toc-quick-links) · [🎬 演示视频](#toc-demo) · [🚀 快速开始](#toc-quick-start) · [🛠️ KWeaver SDK](#toc-kweaver-sdk) · [🛡️ KWeaver Admin CLI](#toc-kweaver-admin) · [🏗️ KWeaver Core](#toc-kweaver-core) · [📐 BKN Lang](#toc-bkn-lang) · [📊 基准测试](#toc-benchmarks) · [💬 社区](#toc-community)

> **注意：** KWeaver Core 是**纯后台框架**，不提供 Web 界面。所有交互通过 CLI、SDK 或 API 完成。如需界面访问，请安装 [**KWeaver DIP**](https://github.com/kweaver-ai/kweaver)。

想直观感受 KWeaver DIP 的核心功能？点击下方链接报名，立即开启试用体验，快速解锁产品价值！👉 申请试用：https://kweaver-ai.feishu.cn/share/base/form/shrcni732cNDY4x3A5SYTncrguf

<a id="toc-quick-links"></a>

## 📚 快速链接

- 🌐 [KWeaver DIP](https://dip-poc.aishu.cn/studio/agent/development/my-agent-list) - KWeaver DIP Demo 环境，[申请试用](https://kweaver-ai.feishu.cn/share/base/form/shrcni732cNDY4x3A5SYTncrguf)
- 🛠️ [KWeaver SDK](https://github.com/kweaver-ai/kweaver-sdk) - 终端用户 / Agent 使用的 `kweaver` CLI、TypeScript / Python SDK 与 Agent Skill
- 🛡️ [kweaver-admin](https://github.com/kweaver-ai/kweaver-admin) - 平台管理员 CLI（用户/组织/角色/模型/审计），完整安装后使用
- 🤝 [贡献指南](rules/CONTRIBUTING.zh.md) - 项目贡献指南
- 🚢 [部署指南](deploy/README.zh.md) - 一键部署到 Kubernetes
- 📘 [产品文档](help/README.md) - 产品文档与使用指南（[中文](help/zh/README.md) / [EN](help/en/README.md)）
- 📦 [示例](examples/README.zh.md) - 端到端 CLI 示例（数据库 / CSV / Action）
- 📝 [博客](https://kweaver-ai.github.io/kweaver-core/) - KWeaver 技术文章与更新
- 🧾 [版本发布](release-notes/) - 重要变更记录

<a id="toc-demo"></a>

## 🎬 演示视频

[在 Bilibili 观看 KWeaver 演示](https://www.bilibili.com/video/BV1nGXVBTEmo/?vd_source=4cdad687b2ac18a0b25e434f1fafe2f7)。

<a id="toc-quick-start"></a>

## 🚀 快速开始

1. **前置与规划** — 阅读 [部署文档](deploy/README.zh.md) 并满足其中前置条件。**正式安装以 Linux 为主**；**macOS** 可选本机 kind 验证见 [Mac 安装（开发向）](deploy/dev/README.md)。
2. **装机前自检 / 修复：`preflight.sh`**（推荐）

   在**安装目标主机**上，安装前先做一次系统体检：内核 / sysctl / containerd / `kubectl` / `helm` / Node / `kweaver` CLI 等，缺什么可按需修（每项默认 y/N 询问，`-y` 全自动）：

```bash
git clone https://github.com/kweaver-ai/kweaver-core.git
cd kweaver-core/deploy
chmod +x preflight.sh deploy.sh onboard.sh

sudo bash ./preflight.sh                # 仅检查（默认）
sudo bash ./preflight.sh --fix          # 检查 + 交互修复
sudo bash ./preflight.sh --fix -y       # 全部自动确认修复
sudo bash ./preflight.sh --list-fixes   # 预览将会执行哪些修复，不改任何东西
sudo bash ./preflight.sh --help         # 全部参数（--role、--skip、--report、--output=json 等）
```

   退出码：**0** 全 OK；**1** 有 FAIL；**2** 仅有 WARN。`--report=/tmp/preflight.txt` 可保存完整日志。

3. **执行安装部署脚本**：

```bash
# （仍在第 2 步的 deploy/ 目录下）

# 最小化安装 — 首次体验推荐
./deploy.sh kweaver-core install --minimum

# 完整安装（包含 auth 和 business-domain 模块）
./deploy.sh kweaver-core install

# 或显式指定地址（跳过交互提示）：
#   --access_address       客户端访问 KWeaver 服务的地址（可以是 IP 或域名）
#   --api_server_address   K8s API Server 绑定的本机网卡 IP（必须是真实的网卡地址）
./deploy.sh kweaver-core install \
  --access_address=<你的IP> \
  --api_server_address=<你的IP>

# 查看帮助
./deploy.sh --help
```

4. **验证部署**：

```bash
# 检查集群状态
kubectl get nodes
kubectl get pods -A

# 检查服务状态
./deploy.sh kweaver-core status
```

5. **安装后引导：`onboard.sh`**（推荐）

   在**同一台机器**（kubectl 能访问集群）上执行安装后引导脚本：注册一个 LLM + 一个 embedding；只有当**默认 embedding 实际变更**时才会 patch BKN ConfigMap 并滚动重启 `bkn-backend` / `ontology-query`；在**完整安装**下还会创建业务用户 **`test`**、把 `kweaver-admin role list` 中的所有角色都挂上、切换 `kweaver` 到该用户身份，并导入 Context Loader 工具集：

```bash
cd deploy
bash ./onboard.sh        # 交互模式；或：bash ./onboard.sh -y
bash ./onboard.sh --help # 所有参数（--config=models.yaml、--enable-bkn-search、--skip-context-loader 等）
```

   可重复运行：脚本会先探测平台已有的模型与 ConfigMap 状态，已注册 / 已配置的会自动跳过。完整流程图、ISF 全量下 `kweaver` 与 `kweaver-admin` 双 CLI 的鉴权说明见 [help/zh/install.md — Post-install：`onboard.sh`](help/zh/install.md#post-installonboardsh安装后引导)。

6. **验证 API 访问**

   KWeaver Core 为纯后台，无 Web 控制台。在**访问端**（本机、跳板机等）通过 [**kweaver-sdk**](https://github.com/kweaver-ai/kweaver-sdk) 使用 `kweaver` CLI：可全局安装 `npm install -g @kweaver-ai/kweaver-sdk`，或直接用 `npx kweaver`（无需全局安装；详见下文 [KWeaver SDK](#toc-kweaver-sdk)）。再执行：

```bash
# 最小化安装（未启用鉴权）：
kweaver auth login https://<节点IP> -k
# 完整安装：以 onboard.sh 创建的业务用户登录（未自定义时默认密码 111111）：
kweaver auth login https://<节点IP> -u test -p '<密码>' -k

kweaver bkn list
# 或使用 npx 免全局安装：
# npx kweaver auth login https://<节点IP> -k
# npx kweaver bkn list
```

7. **查看帮助**：

```bash
kweaver --help                   # 列出所有命令
kweaver <command> --help         # 查看某命令的帮助，例如 kweaver bkn --help
```

完整产品文档参见[文档中心](help/README.md)（[中文](help/zh/README.md) / [EN](help/en/README.md)）。

> **完整安装（未加 `--minimum`）？** 建议同时安装 [`kweaver-admin`](#toc-kweaver-admin) 进行用户、组织、角色、模型与审计管理 — 详见 [help/zh/install.md — 完整安装后的管理员工具](help/zh/install.md#-完整安装后的管理员工具kweaver-admin)。

> **尚未部署？** KWeaver DIP / 公开 Demo 可能需先申请试用 — 请填写 [申请试用](https://kweaver-ai.feishu.cn/share/base/form/shrcni732cNDY4x3A5SYTncrguf)。获得权限后，可打开 [KWeaver DIP](https://dip-poc.aishu.cn/studio/agent/development/my-agent-list) Web 界面，或将 CLI/SDK 连接到 Demo 环境（见下方说明）。

<a id="toc-kweaver-core"></a>

## 🏗️ KWeaver Core

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

<a id="toc-bkn-lang"></a>

### 📐 BKN Lang

BKN Lang 是基于 Markdown 扩展语法的业务知识建模语言，人机双向友好：

- **易开发** — 基于通用 Markdown 语法，彻底消除代码壁垒。业务专家可通过所见即所得编辑器直接编写、阅读和修改业务逻辑定义，支持快速版本比对与协作审计，像修改文档一样修改系统规则。
- **易理解** — "对象类-关系类-风险类-行动类"四位一体模型完美映射企业业务模型。人类读懂业务含义，智能体实时"阅读"并解析出精准的上下文约束。逻辑显性化，拒绝性黑盒，从根本上降低大模型的推理幻觉与逻辑偏差。
- **易集成** — 定义仅作为全量文本存储于数据库特定字段，无复杂底层表结构强耦合。通过 Context Loader 按需动态加载，摒弃静态硬编码。跨系统、跨智能体高度兼容，作为轻量级资产在 AI Data Platform 中流畅流转。

### 核心价值总结

| 指标 | 数据 |
| --- | --- |
| **场景覆盖** | 问答、流程执行、情报分析、决策判断、探索类场景 |
| **TCO 降低** | 一体化平台降低建设成本 70% |
| **BKN 构建效率** | 业务知识网络构建效率提升 300% |
| **Token 成本** | 上下文优化与压缩，消耗降低 50% |

<a id="toc-kweaver-sdk"></a>

## 🛠️ KWeaver SDK

<a id="toc-kweaver-core-and-sdk"></a>

### 在客户端上安装 SDK

安装完 KWeaver Core 后，建议首先安装 [kweaver-sdk](https://github.com/kweaver-ai/kweaver-sdk)。SDK 提供 `kweaver` CLI 和 AI Agent Skills，是与平台交互的主要方式。

[**kweaver-sdk**](https://github.com/kweaver-ai/kweaver-sdk) 通过 `kweaver` CLI 为 AI 智能体（Claude Code、GPT、自定义 Agent 等）提供对 KWeaver 知识网络与 Decision Agent 的访问能力，同时提供 Python 与 TypeScript SDK 用于编程集成。

使用以下命令安装 CLI：

```bash
npm install -g @kweaver-ai/kweaver-sdk
```

或者无需全局安装，直接运行：

```bash
npx kweaver --help
```

### AI Agent Skills

从 [**kweaver-sdk**](https://github.com/kweaver-ai/kweaver-sdk) 使用 [`npx skills`](https://www.npmjs.com/package/skills) 安装技能。

**一次安装两个 skill**（推荐）：

```bash
npx skills add https://github.com/kweaver-ai/kweaver-sdk \
  --skill kweaver-core --skill create-bkn
```

- **`kweaver-core`** — 让 AI 编程助手掌握 KWeaver 的 API 与 CLI，可代替用户操作 KWeaver 平台。详见 [skills/kweaver-core/SKILL.md](https://github.com/kweaver-ai/kweaver-sdk/blob/main/skills/kweaver-core/SKILL.md)。
- **`create-bkn`** — 在 AI 编程助手中创建与管理**业务知识网络（BKN）**的流程与工具。详见 [skills/create-bkn/SKILL.md](https://github.com/kweaver-ai/kweaver-sdk/blob/main/skills/create-bkn/SKILL.md)。

**安装一个 skill**（可选）：

```bash
npx skills add https://github.com/kweaver-ai/kweaver-sdk --skill kweaver-core
# 或
npx skills add https://github.com/kweaver-ai/kweaver-sdk --skill create-bkn
```

**使用任意 skill 前**，需先完成 KWeaver 实例认证：

```bash
kweaver auth login https://your-kweaver-instance.com
```

> **注意** 如果你的实例使用了自签名或不受信任的 TLS 证书（新部署且未配置 CA 签发证书时很常见），添加 `-k` 参数跳过证书验证：
>
> ```bash
> kweaver auth login https://your-kweaver-instance.com -k
> ```

### 使用 Demo 环境快速体验

无需部署 — 请先完成 [申请试用](https://kweaver-ai.feishu.cn/share/base/form/shrcni732cNDY4x3A5SYTncrguf)，再将 AI Agent 连接到 Demo 环境（Web 界面见 [KWeaver DIP](https://dip-poc.aishu.cn/studio/agent/development/my-agent-list)）。

```bash
npx skills add https://github.com/kweaver-ai/kweaver-sdk \
  --skill kweaver-core --skill create-bkn

npm install -g @kweaver-ai/kweaver-sdk
kweaver auth login https://dip-poc.aishu.cn -k
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

> **Demo 访问：** 如需请先 [申请试用](https://kweaver-ai.feishu.cn/share/base/form/shrcni732cNDY4x3A5SYTncrguf)。

### 无浏览器环境认证（SSH、CI、容器等）

`kweaver` CLI 支持在没有本地图形浏览器、或不便粘贴回调 URL 的场景下完成认证。

**如何选用：**

| 你的情况 | 请用 | 说明 |
| --- | --- | --- |
| **方便使用用户名与密码** | **方式 1**（HTTP，`-u` / `-p`） | 本机一条命令完成，无需在终端外复制 OAuth 回调。 |
| **已安装** [kweaver-sdk](https://github.com/kweaver-ai/kweaver-sdk)（本机可执行 `kweaver`） | **方式 2**（`auth export` / 重放） | 在能打开浏览器的机器上完成登录后执行 `kweaver auth export`，将导出的一行命令在目标环境**重放**。 |
| **尚未安装** kweaver-sdk、通常用 `npx kweaver` 等运行 | **方式 3**（`--no-browser`） | 在其它设备上打开 OAuth 链接登录后**多一步**：从地址栏**复制完整回调 URL 或只复制 authorization code**（界面常提示 *copy code*），粘贴到无头端 `Paste URL or code` 处。 |

**方式 1 — 用户名密码 HTTP 登录（本机全自动化，无需浏览器）**

无需 Node/Chromium，CLI 通过 HTTPS 直接调用平台的 `/oauth2/signin` 接口并保存返回的令牌，适用于 CI Runner、最小化 Linux 容器以及任何无浏览器主机：

```bash
kweaver auth login https://你的实例 -u <用户名> -p <密码> -k
```

同时提供 `-u` 与 `-p` 时会自动走该路径（也可显式追加 `--http-signin`）。如果不带 `-u` / `-p`，CLI 会从 stdin 交互获取（TTY 下密码不回显）。令牌写入 `~/.kweaver/`；若 IdP 返回 `refresh_token`，后续换发 access token 的行为与普通浏览器登录一致。

**方式 2 — 导出凭据后重放（已安装 kweaver-sdk 时；export / 重放）**

在**已安装** `kweaver` CLI 的机器上，用浏览器完成一次正常登录后，用 `export` 得到可粘贴到无头环境的一行命令，无需在终端里手抄 OAuth 回调 URL 或 code。

1. 在**有浏览器**的机器上执行 `kweaver auth login https://你的实例`，登录成功后导出凭据：

```bash
kweaver auth export              # 输出一行命令，可直接在无头机器上粘贴执行
```

2. 在**无头机器**上粘贴导出的命令，它通过 `--client-id`、`--client-secret`、`--refresh-token` 换取令牌并写入 `~/.kweaver/`：

```bash
kweaver auth login https://你的实例 \
  --client-id <ID> --client-secret <SECRET> --refresh-token <TOKEN>
```

**方式 3 — `--no-browser`（未安装 kweaver-sdk 时；多一步复制 URL 或 code）**

适用于尚未安装 CLI、常配合 `npx kweaver ...` 使用的场景。与**方式 2**相比，在另一设备上登录后需**多一步**手动从浏览器**复制** URL 或 code。

```bash
kweaver auth login https://你的实例 --no-browser
# 或: npx kweaver auth login https://你的实例 --no-browser
```

CLI 不会打开本机图形浏览器，而是打印一个 OAuth URL。将该 URL 复制到**任意有浏览器的设备**（手机、笔记本等）打开并登录。登录后浏览器会跳转到 `localhost` 回调地址 — 页面可能显示错误（这是正常的）。从浏览器地址栏复制**完整 URL**，**或仅复制 authorization code**，粘贴回 CLI 提示符（即下方 *Paste URL or code* 多出的这一步）：

```
Open this URL on any device (use a private/incognito window if you need the full sign-in form):

  https://你的实例/oauth2/auth?redirect_uri=...&client_id=...

After login, the browser may show an error page (this is expected if nothing listens on localhost).
Copy the FULL URL from the address bar and paste it here, or paste only the authorization code.

Paste URL or code>
```

> 已有 `~/.kweaver/` 保存的会话时，CLI 会在 access token 过期后自动使用 `refresh_token` 换取新令牌，无需额外操作。也可通过环境变量（`KWEAVER_BASE_URL`、`KWEAVER_TOKEN`）传入凭据，无需持久化到磁盘。

完整说明见 [kweaver-sdk — 认证说明](https://github.com/kweaver-ai/kweaver-sdk#authentication) 与 [Headless / Server Authentication](https://github.com/kweaver-ai/kweaver-sdk/blob/main/packages/typescript/README.md#headless--server-authentication)。Python 版 `kweaver` CLI 仍为交互式浏览器登录；可将 Node CLI 已完成登录的机器上的 `~/.kweaver/` 目录拷贝过来使用，或配置上述环境变量。

### CLI

```bash
kweaver auth login https://your-kweaver.com -k    # 登录（自签名证书加 -k）
kweaver bkn list                                 # 列出知识网络
kweaver bkn search <kn-id> "关键词"              # 语义搜索
kweaver agent chat <agent-id> -m "你好"         # 与 Decision Agent 对话
kweaver --help                                   # 查看全部子命令
```

### TypeScript & Python SDK

精简示例（需先完成 CLI 登录或等价凭据配置）：

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

流式对话、`KWeaverClient` 与完整 API 等更多用法，请参阅 [kweaver-sdk](https://github.com/kweaver-ai/kweaver-sdk) 仓库文档与示例。

<a id="toc-kweaver-admin"></a>

## 🛡️ KWeaver Admin CLI

[`kweaver-admin`](https://github.com/kweaver-ai/kweaver-admin) 是面向**平台管理员**的独立 npm CLI，与面向终端用户/Agent 的 `kweaver`（kweaver-sdk）互补：

| CLI | 受众 | 覆盖范围 |
| --- | --- | --- |
| `kweaver`（`@kweaver-ai/kweaver-sdk`） | 业务用户 / Agent | BKN、Decision Agent、Action、Skill、查询 |
| `kweaver-admin`（`@kweaver-ai/kweaver-admin`） | 平台管理员 | 用户、组织、角色、模型、审计、原始 HTTP |

> `kweaver-admin` 主要操作的服务来自**完整安装**（`auth.enabled=true`、`businessDomain.enabled=true`）：`user-management`、`deploy-manager`、`deploy-auth`、`eacp`、`mf-model-manager`、OAuth2(Hydra)。`--minimum` 安装下大多数命令会返回 401 / 404，属正常裁剪。

### 安装

要求 Node.js 22+（与 [npm 上 `@kweaver-ai/kweaver-sdk`](https://www.npmjs.com/package/@kweaver-ai/kweaver-sdk) 的 `engines` 一致）。凭据保存在 `~/.kweaver-admin/platforms/`，与 `~/.kweaver/` 隔离。

```bash
npm install -g @kweaver-ai/kweaver-admin
kweaver-admin --version
```

### 登录

```bash
# 浏览器 OAuth2（自签名证书加 -k）
kweaver-admin auth login https://<访问地址> -k

# 用户名/密码（CI 或无浏览器场景）
kweaver-admin auth login https://<访问地址> -u <用户名> -p <密码> -k

# 或通过环境变量
export KWEAVER_BASE_URL=https://<访问地址>
export KWEAVER_ADMIN_TOKEN=<bearer-token>   # 也可回退到 KWEAVER_TOKEN
```

### 常用管理命令

```bash
kweaver-admin org tree                           # 列出部门
kweaver-admin user create --login alice          # 默认密码 123456，首次登录强制改密
kweaver-admin user reset-password -u alice       # 管理员重置密码
kweaver-admin role list
kweaver-admin role add-member <roleId> -u alice
kweaver-admin llm add                            # 注册 LLM
kweaver-admin small-model add                    # 注册 Embedding 模型
kweaver-admin audit list --user alice --start 2026-04-01 --end 2026-04-30
kweaver-admin call /api/user-management/v1/management/users -X GET   # 自动带认证头的原始 HTTP
```

> `user create` 新建账号的初始密码固定为 **`123456`**，首次登录强制改密（这是 ISF 用户存储的上游既定行为，并非 CLI 选择）。忘/失密请用 `kweaver-admin user reset-password`。

> 内置「三权分立」账号 `system / admin / security / audit` 不可随意删改；请使用各自的个人账号操作，避免共用 `admin`。

完整命令树、安全说明与 `auth change-password`（EACP `modifypassword`，与 `kweaver` CLI 一致的 `401001017` 首登流程）：见 [kweaver-admin README](https://github.com/kweaver-ai/kweaver-admin) 与 [help/zh/install.md — 完整安装后的管理员工具](help/zh/install.md#-完整安装后的管理员工具kweaver-admin)。

### AI Agent Skill —— `kweaver-admin`

`kweaver-admin` 仓库自带渐进式（progressive disclosure）Skill，可让 AI 编程助手代你执行管理员操作（auth、org、user、role、model、audit、原始 HTTP）。通过 `npx skills` 安装：

```bash
npx skills add https://github.com/kweaver-ai/kweaver-admin --skill kweaver-admin
```

安装后先用 `kweaver-admin auth login https://<访问地址> -k` 登录一次，然后用自然语言（也支持 `/kweaver-admin` 斜杠命令）指挥助手：

```text
列出所有角色
新建用户 alice，并把 role list 中的所有角色都赋给她
重置 alice 的密码
查看 alice 最近 7 天的登录审计
注册一个名为 bge-m3 的 Embedding 模型，base 是 https://api.siliconflow.cn
```

Skill 源文件：[skills/kweaver-admin/SKILL.md](https://github.com/kweaver-ai/kweaver-admin/blob/main/skills/kweaver-admin/SKILL.md)。它与上文面向 `kweaver` CLI 的 `kweaver-core` / `create-bkn` Skill 互补。

<a id="toc-benchmarks"></a>

## 📊 基准测试与实验

查看更多：[KWeaver Blog](https://kweaver-ai.github.io/kweaver-core/)

### 非结构化数据问答 — 跨平台横向对比

基于 145 个 HR 场景样本（简历库含 118 份多格式 PDF），覆盖简单信息查询、跨段落经验分析、多跳综合推理三类场景。各平台统一使用 DeepSeek V3.2 + BGE M3-Embedding，相同数据源，均采用 Agentic 模式测试。

| 指标 | KWeaver Core (v0.3.0) | BiSheng | Dify (v0.15.3) | RAGFlow (v0.17.0) |
| --- | --- | --- | --- | --- |
| **通过率** | **99.31%** (144/145) | 86.90% (126/145) | 96.55% (140/145) | 86.90% (126/145) |
| **平均响应时间** | 43.69s | **19.52s** | 63.82s | 71.56s |
| **P90 响应时间** | 56.92s | 32.53s | 79.15s | 95.37s |
| **平均 Token 消耗** | 21.36K | **4.98K** | 36.25K | 16.28K |

唯有 KWeaver Core 突破了传统 RAG 架构的"性能不可能三角"——在满足企业级严苛可靠性（>99% 准确率）的同时，将推理成本和时间延迟控制在高度生产可用的水平。Dify 属于"高消耗换取高通过率（力大砖飞）"；BiSheng 则是"牺牲准确推理换取表面上的快速低耗"；RAGFlow 在准确率和延迟上均落后。

### 消融实验 — 关键技术杠杆

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

### 异构数据推理 — F1 Bench

F1 Bench 由 BIRD 测试集中 Formula-1 数据库混合 30 篇非结构化文档设计，测试智能体结合结构化+非结构化异构数据推理能力。

| 指标 | KWeaver Core | Dify 召回基线 |
| --- | --- | --- |
| **整体准确率** | **92.96%** | 78.87% |
| **SQL 浪费率** | **8.2%** | 24.5% |
| **SQL 命中效率** | **0.226** | 0.137 |
| **总 SQL 调用次数** | **292** | 408 |

<a id="toc-community"></a>

## 🤝 社区

**KWeaver 微信交流群二维码**

<img src="./help/qrcode.png" width="200" alt="KWeaver 微信交流群二维码" />

扫码加入 KWeaver 交流群。
