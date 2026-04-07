## 1.0 背景分析 (Background Analysis)
### 1.1 风险与机遇：Data Agent 时代的执行环境
在人工智能（AI） Agent 和用户自定义代码执行的新范式下，构建一个安全、隔离的执行环境已不再是可选项，而是至关重要的战略基础设施。直接在生产环境中运行由大语言模型（LLM）动态生成或用户提交的非受信代码，会带来一系列固有风险。这些风险包括但不限于：利用系统漏洞进行恶意攻击、因无限循环或内存泄漏导致的资源滥用、以及因环境不一致引发的不可预测行为。一个健壮的沙箱环境是释放 Data Agent 全部潜力的前提，它通过创建一道坚固的屏障，确保创新探索不会以牺牲系统安全与稳定性为代价。

### 1.2  算子平台对沙箱执行环境的要求
算子平台需要提供“函数”的核心能力，让开发者能够以最小成本编写并运行自定义 Python 函数，以满足多样化的数据处理、算子扩展和业务逻辑需求。为保证整体系统的安全性、可控性与可运维性，这些函数必须在沙箱环境中执行，避免对宿主系统造成影响。

沙箱系统需支持对** Python 代码的隔离执行**，包括**运行时资源限制**、**权限限制、网络访问隔离**等。平台还需具备对 **Python 依赖包的离线管理能力**，可在无外网/受限网络的环境下为用户函数自动装载依赖，确保执行环境一致性与可复现性。同时，沙箱需具备较高的执行效率和并发能力，支持函数的快速启动、稳定运行和可预测的资源开销，以满足算子平台对延迟、吞吐、成本和可靠性的要求。

最终，沙箱系统将作为算子平台的基础执行层，使用户能够在受控环境中灵活、安全地扩展数据处理能力。

### 1.3 范式跃迁：从“提示工程”到“环境工程”
行业正经历一场深刻的范式变迁，即从传统的“提示工程”（Prompt Engineering）演进为更为系统和可控的“环境工程”（Environment Engineering）。正如 Anthropic 公司从程序化工具调用（PTC）等技术演进所示，其核心思想是将模型从一个单纯的对话者，重塑为一个拥有标准化 I/O 和逻辑执行能力的智能内核。这一演进直接解决了多步工具调用产生的高延迟，以及在 JSON 中表达循环、条件判断等复杂控制流时的逻辑脆弱性，这些是先前模型固有的问题。

在这种新范式下，Data Agent 不再仅仅是生成文本或简单的函数调用，而是能够编写和执行包含复杂逻辑的完整代码脚本。静态代码分析在面对 AI 生成代码的多变性和不确定性时已显乏力，这进一步凸显了动态执行环境的必要性。因此，一个高性能、低延迟且逻辑确定性的沙箱执行环境，是实现高级 Agent 场景的关键基础设施，而非简单的功能附加。

![](https://cdn.nlark.com/yuque/0/2025/png/1904465/1764745304495-21e10320-31a5-447b-b471-0d030d752213.png)



### 1.4 面向 PTC 数据分析 + 上下文问答的 统一沙箱临时区
在智能体执行过程中，系统需要同时满足两类文件处理需求：

+ 其一，面向数据分析、计算类任务的 **PTC 场景**，模型需通过程序化代码执行方式对用户上传的文件进行过滤、聚合、统计等操作，因此需要一个可安全运行 Python、可安装依赖、支持文件持久化的 **高隔离沙箱环境**，保证计算的确定性与资源可控性。
+ 其二，面向普通对话问答的 **非 PTC 场景**，用户上传文件仅作为知识来源，但文件格式可能复杂（如 PDF、图片、CSV）且存在潜在风险，因此仍需通过沙箱对文件进行安全解析、内容抽取与结构化处理，以便可靠地生成可注入模型上下文的精简语料。

基于这两类场景的共同需求，系统需要构建一个统一的 **临时区沙箱**：既可作为智能体的计算执行引擎，又可作为文件解析与安全处理的中间层，从而在不同任务模式下提供一致、隔离、可控的文件处理能力。

```plain
                   /-------------------\
                   | 用户上传文件 File  |
                   \--------+----------/
                            |
                     判断使用场景？
                +-----------+-------------+
                |                         |
        (A) PTC 数据分析           (B) 非 PTC 问答
                |                         |
                |                         |
   /-------------------------\   /-----------------------------\
   | 沙箱执行 Python 代码    |   | 是否需要解析？(PDF/OCR/表格) |
   | (pandas, numpy...)     |   \-----------+-----------------/
   \-----------+-------------/               |
               |                             |
    结构化结果返回给 LLM              是需要 → 进入沙箱
               |                             |
         LLM 生成结论                     解析文件（OCR/PDF/CSV）
                                            |
                                 输出干净文本/结构化数据
                                            |
                             将结果注入 LLM 上下文或 RAG
                                            |
                                      LLM 回答用户

```

### 1.5 核心需求
当前市场对沙箱方案的需求日益明确和严苛。根据行业分析，一个理想的沙箱方案必须在以下几个维度表现出色：

+ **隔离性与安全性 (Isolation & Security):** 提供强有力的隔离机制，从业界领先方案使用的 Firecracker microVMs（如 E2B、Fly.io）到 V8 isolates（如 Cloudflare Workers），其共同点是严格限制对文件系统、网络和系统调用的访问，防止代码“逃逸”。
+ **性能与速度 (Performance & Speed):** 必须将性能开销降至最低，实现快速的环境启动和代码执行，以满足 Data Agent 对实时响应的要求。
+ **语言与运行时支持 (Language & Runtime Support):** 需支持主流编程语言，特别是 Python，并提供灵活的依赖管理能力。
+ **可扩展性与资源管理 (Scalability & Resource Management):** 能够高效地并发运行大量沙箱实例，并提供精细化的资源（CPU、内存、执行时间）控制能力。
+ **开发者体验与集成 (Developer Experience & Integration):** 提供清晰的 API、SDK 和文档，简化与现有系统的集成，并支持模板化配置。
+ **可观测性与调试 (Observability & Debugging):** 提供充足的日志、指标和调试工具，以便于问题排查和性能优化。

### 1.6 方案愿景
上述分析揭示了在 Data Agent 时代，传统执行模式已无法满足安全、性能和功能上的新要求。本技术方案旨在直面这些挑战，通过设计一个模块化、多层次隔离、高性能的下一代沙箱平台，为 Data Agent 和自定义函数提供一个安全、可靠且高效的执行基石。

## 2.0 关键目标 (Key Objectives)
### 2.1 目标导向设计
明确定义项目的关键目标是成功构建复杂系统的基石。这不仅为技术选型和架构设计提供了清晰的指引，也为后续衡量方案的成功与否设立了量化的标准。本方案的核心目标围绕技术、功能、性能和体验四个维度展开。

### 2.2 核心目标详解
以下是我们为下一代沙箱技术方案设定的核心目标：

1. **技术目标 (Technical Objectives)**
    - **环境隔离与安全 (Isolation & Security):**
        * **强隔离:** 必须实现进程、文件系统、网络和系统调用层面的强隔离，有效防止任何形式的代码“逃逸”攻击。
        * **访问控制:** 提供精细化的权限控制，严格限制对文件系统、网络资源（如特定域名）和系统调用的访问。
    - **多环境支持 (Multi-Environment Support):**
        * **技术栈灵活性:** 系统需原生支持多种沙箱隔离技术，包括基于 `bubblewrap` 的轻量级命名空间隔离和基于容器（如 Docker）的强环境隔离，以灵活适应不同场景对安全与性能的平衡需求。
    - **语言与运行时 (Language & Runtime):**
        * **核心语言支持:** 首要目标是为 Python 代码执行提供一流的支持，满足 PTC & Function 的核心需求。
        * **架构可扩展性:** 系统架构必须具备前瞻性，能够平滑扩展至支持 JavaScript 等其他主流语言。
2. **功能目标 (Functional Objectives)**
    - **会话生命周期管理 (Session Lifecycle):**
        * **全生命周期:** 实现对执行会话（Session）从创建（Creation）、激活（Active）、空闲（Idle）到最终销毁（Destruction）的完整生命周期管理。
        * **状态管理:** 包含自动化的空闲超时和资源回收机制，优化资源利用率。
    - **动态调度与负载均衡 (Dynamic Scheduling & Load Balancing):**
        * **智能调度:** 管理中心需具备智能调度能力，能根据请求的场景和安全级别，从资源池中选择最合适的沙箱类型。
        * **负载均衡:** 动态监控所有沙箱实例的负载，并对会话执行请求进行有效的负载均衡，确保系统整体的稳定性和响应能力。
    - **流式与调试支持 (Streaming & Debugging):**
        * **实时交互:** 必须支持通过 WebSockets、SSE 等技术进行实时流式输出，为交互式应用（如 AI Chat）提供最佳用户体验。
        * **可观测性:** 提供全面的调试日志、性能指标和事件追踪功能，以支持开发者进行问题排查和系统监控。
3. **性能目标 (Performance Objectives)**
    - **执行效率 (Execution Speed):**
        * **执行延迟:** 对于简单函数的执行，端到端延迟需控制在 **100 毫秒** 以内。
        * **关键指标:** 明确“命令执行速度”是比“沙箱启动时间”更为关键的性能指标，直接影响用户和 Agent 的交互体验。
    - **资源占用 (Resource Consumption):**
        * **默认限制:** 为每个沙箱会话设定默认的资源限制，例如内存占用不超过 **100MB**，CPU 使用率不超过 **100%**（相当于在 100ms 周期内分配 100ms 的 CPU 时间，即单核完全占用）。
    - **并发能力 (Concurrency):**
        * **高并发处理:** 系统需设计为能够支持至少 **100 个并发函数执行**，以满足规模化应用的需求。
4. **体验目标 (Developer Experience Objectives)**
    - **易于集成 (Ease of Integration):**
        * **标准化接口:** 提供定义清晰、文档完备的 API 和 SDK，最大程度地简化与上游系统（如 Data Agent、算子平台）的集成工作。
    - **模板化配置 (Templated Configuration):**
        * **配置即代码:** 支持会话模板（Template）的配置，允许开发者预定义执行环境、依赖库、资源限制和网络策略，实现环境的快速复用和标准化管理。

### 2.3 目标驱动架构
这些明确的目标共同勾画出一个安全、高效、易用且可扩展的沙箱执行平台蓝图。为了实现这些目标，我们首先需要通过领域驱动设计（DDD）来定义系统的核心概念与对象。

## 3.0 核心概念和对象 (Core Concepts and Objects - DDD)
### 3.1 领域驱动设计思想
为了构建一个清晰、健壮且易于演进的系统，我们采用领域驱动设计（DDD）的思想来定义核心概念和对象。通过 DDD，我们能够建立统一的语言（Ubiquitous Language），明确限界上下文（Bounded Context），并构建出一个反映业务需求的、高内聚的领域模型。这有助于降低系统复杂度，提升代码的可维护性和团队沟通效率。

### 3.2 核心领域对象


![](https://cdn.nlark.com/yuque/__mermaid_v3/3a11b73fd2be5439a7c5d17632fb189a.svg)

#### 3.2.1 `执行会话` (Execution Session) - 聚合根 (Aggregate Root)
+ **定义:**`执行会话`（Session）是本领域的核心聚合根，代表了一次完整且独立的函数执行生命周期单元。它封装了所有与单次执行相关的数据和行为，确保了业务规则的一致性。一个 `Session` 实例包含待执行的代码（Code）、输入参数（Input）、所引用的`沙箱模板`（Template ID）、自身的执行状态（Execution Status），以及最终的`执行结果`（Result）。
+ **核心职责:**
    - **状态一致性维护:** 管理自身的执行状态机，确保状态（如 `PENDING`, `RUNNING`, `COMPLETED`, `FAILED`）的转换符合预定义的业务规则。
    - **数据封装:** 封装待执行的代码、输入参数和环境变量，确保执行上下文的完整性。
    - **结果记录:** 作为执行产物的唯一所有者，负责记录和存储执行完成后生成的`执行结果`，包括标准输出、标准错误和函数返回值。
+ **状态澄清:**`Session` 的执行状态（Execution Status）与底层`沙箱 Runtime` 资源的生命周期状态（Resource Lifecycle）是两个独立的概念。`Session` 关注的是代码执行的进程，而资源生命周期关注的是承载执行的沙箱实例本身的状态（例如 `CREATING`, `WARM`, `ACTIVE`, `IDLE`, `DESTROYED`）。这种分离使得领域模型更加清晰。

#### 3.2.2 `沙箱模板` (Sandbox Template) - 实体 (Entity)
+ **定义:**`沙箱模板`（Template）是一个可配置的实体，它定义了沙箱执行环境的规格和约束。每个`执行会话`都必须关联一个模板，该模板决定了会话将在何种环境下运行。模板是可复用的，允许开发者标准化不同应用场景的环境配置。
+ **关键属性:**
    - **沙箱类型 (Sandbox Type):** 指定底层的隔离技术，例如 `bubblewrap`（轻量级隔离）或 `microVMs`（强隔离）。
    - **资源限制 (Resource Limits):** 定义执行环境的资源配额，包括 CPU 限制、内存上限（如 100MB）和最大执行时间（如 3分钟）。
    - **预装依赖 (Dependencies):** 一个 Python 包列表，定义了在代码执行前需要在沙箱环境中预先安装的第三方库。
    - **网络策略 (Network Policy):** 定义网络访问规则，如一个明确的允许访问域名列表（allowlist）或禁止访问的域名列表（denylist）。

#### 3.2.3 `执行结果` (Execution Result) - 值对象 (Value Object)
+ **定义:**`执行结果`（Result）是一个不可变的（Immutable）值对象，用于封装一次`执行会话`完成后的所有产物。其不可变性确保了执行历史的准确性和不可篡改性。
+ **标准结构:** 为了便于上游系统（尤其是 Data Agent）进行确定性的解析，`执行结果`采用标准化的数据结构，必须包含以下字段：
    - **执行状态码 (Exit Code):** 整数，表示执行是否成功（通常 0 为成功）。
    - **标准输出 (stdout):** 字符串，捕获脚本执行期间的所有标准输出。
    - **标准错误 (stderr):** 字符串，捕获脚本执行期间的所有标准错误信息。
    - **函数返回值 (Return Value):** 任意类型，捕获脚本主函数的返回值。
    - **性能指标 (Metrics):** 一个包含关键性能数据的对象，如`执行耗时`（毫秒）和`内存峰值`（MB）。

### 3.3 构建领域模型
这三个核心对象——`执行会话`、`沙箱模板`和`执行结果`——共同构成了我们沙箱系统的核心领域模型。它们清晰地划分了职责，为接下来的架构设计奠定了坚实的基础。接下来，我们将探讨这些对象如何在两大核心能力中心（沙箱 Runtime 和沙箱管理中心）中发挥作用。

## 4.0 沙箱核心能力拆解 (Sandbox Core Capabilities)
### 4.1 关注点分离架构
为构建一个模块化、可扩展且易于维护的系统，我们将整个沙箱平台在逻辑上拆分为两大核心组件：**沙箱 Runtime (Sandbox Runtime)** 和 **沙箱管理中心 (Sandbox Management Center)**。`沙箱 Runtime` 专注于“执行”，提供一个纯粹、安全的执行内核；而`沙箱管理中心`则专注于“管理”，扮演着系统的大脑和指挥中心。这种关注点分离的设计模式，极大地提升了系统的整体灵活性和可维护性。

### 4.2 沙箱 Runtime (Sandbox Runtime)
#### 4.2.1 职责定义
`沙箱 Runtime` 的核心职责是提供一个安全、隔离的 **“执行内核”**。它的任务是接收来自管理中心的指令，实例化一个符合`沙箱模板`规范的隔离环境，并在其中原子性地执行一个`执行会话`（Execution Session），最终将完整的`执行结果`返回。

#### 4.2.2 关键能力分析
+ **会话执行引擎 (Session Execution Engine):** 作为 Runtime 的核心，该引擎负责接收一个 `Execution Session` 对象，解析其中的代码和参数，在隔离环境中启动相应的解释器（如 Python），执行代码，并实时捕获其标准输出（stdout）、标准错误（stderr）以及最终的函数返回值。
+ **多环境隔离技术实现 (Multi-Environment Isolation):** 为了满足不同场景下的安全与性能需求，Runtime 必须支持可插拔的隔离技术。调度器将默认选择 `bubblewrap` 以优化性能和成本，仅当沙箱模板明确要求或安全策略强制时，才调度到资源开销更高的容器化 Runtime。
    - **轻量级隔离 (**`**bubblewrap**`**):** 参考 `anthropic-experimental/sandbox-runtime` 的实现，利用 Linux 的 `bubblewrap` 或 macOS 的 `sandbox-exec` 工具创建独立的网络和文件系统命名空间。这种方式启动快、开销小，适用于信任度较高或对性能极度敏感的场景。
    - **强隔离 (容器化):** 利用 Docker 或其他兼容 OCI 的容器技术，为每次执行提供一个完全隔离的容器环境。这种方式提供了包括依赖库、操作系统环境在内的更强隔离保证，适用于执行完全不受信代码的场景。
+ **资源限制强制执行 (Resource Limit Enforcement):** Runtime 必须严格执行`沙箱模板`中定义的资源限制。我们将利用操作系统底层的 `**cgroups**` 功能，对每个沙箱实例的 CPU 使用率、内存消耗和 I/O 带宽进行精确控制和强制约束，防止单个恶意或有缺陷的会话耗尽整个节点的资源（即“吵闹的邻居”问题）。
+ **依赖管理 (Dependency Management):** 在无网络或严格限制网络的沙箱环境中管理 Python 依赖是一个关键挑战。本方案将以**“两阶段执行模型”**为核心策略，因为它在灵活性和安全性之间取得了最佳平衡。对于高度固定的依赖环境，我们亦支持**“预置镜像模型”**作为一种性能优化手段。
    1. **核心策略：两阶段执行模型:** 受 `pydantic/mcp-run-python` 启发，执行过程分为两步。第一阶段，在一个有网络访问权限的预备环境中，根据`沙箱模板`中的依赖列表下载并安装依赖到一个共享的缓存卷。第二阶段，在启动完全无网络的执行沙箱时，将此缓存卷以只读方式挂载进去，供 Python 代码加载。
    2. **优化策略：预置镜像模型:** 对于常用的依赖组合，可以预先构建包含这些依赖的不可变容器镜像或文件系统根（rootfs）。当请求匹配时，直接启动对应的预置环境。这种方式启动速度最快，环境最确定，但灵活性稍差。

### 4.3 沙箱管理中心 (Sandbox Management Center)
#### 4.3.1 职责定义
`沙箱管理中心`是整个沙箱系统的 **“指挥大脑”**。它对外提供统一的 API，对内负责所有`沙箱 Runtime`实例的调度、会话生命周期的管理、系统资源的池化以及整体的可观测性。

#### 4.3.2 关键能力分析
+ **会话生命周期管理 (Session Lifecycle Management):** 借鉴 Cloudflare Sandbox 的设计理念，管理中心负责维护每个`执行会话`的完整状态机，包括从创建、排队、运行、完成到销毁的各个阶段。它还需实现空闲会话的超时回收机制，自动释放长时间未使用的沙箱资源。
+ **智能调度与负载均衡 (Intelligent Scheduling & Load Balancing):** 调度器是管理中心的核心。当接收到新的执行请求时，它会解析请求中指定的`沙箱模板`类型（如 `bubblewrap` 或 `container`），然后从相应的资源池中选择一个负载最低的`沙箱 Runtime`实例来执行任务。如果池中没有可用实例，调度器将负责按需创建新的实例。
+ **缓存与队列管理 (Caching & Queue Management):**
    - **温沙箱池 (Warm Sandbox Pools):** 为了大幅降低会话的冷启动延迟，管理中心会维护一个“温沙箱池”。池中预先启动了一定数量的、处于待命状态的沙箱实例（容器或 `bubblewrap` 环境）。当新请求到达时，可以直接从池中获取一个“温”实例，从而实现毫秒级的会话启动。
    - **请求队列:** 为应对高并发的执行请求，管理中心使用消息队列（如 Kafka 或 Redis Streams）作为缓冲。API 入口将请求快速推入队列后即可返回，由后端的调度器异步消费队列中的任务，实现了系统削峰填谷和异步处理。
+ **流式与非流式通信 (Streaming & Non-Streaming Communication):** 为满足不同应用场景的需求，管理中心必须提供两种通信模式的 API 端点。参考 Google Cloud Run 的设计，它应同时支持：
    - **HTTP 同步调用:** 用于简单的、请求-响应式的函数执行。
    - **WebSocket 长连接:** 用于需要实时流式返回输出的交互式场景。这将消除 HTTP 轮询的需要，通过避免重复握手来降低延迟，并允许服务器向客户端“推送”实时更新，这对于观察 Data Agent 的思考过程或执行日志至关重要。
+ **模板与配置管理 (Template & Configuration Management):** 管理中心提供一套 CRUD API，用于对`沙箱模板`进行全生命周期的管理，包括创建、更新、读取和版本控制，为开发者提供灵活的环境定义能力。
+ **可观测性与调试 (Observability & Debugging):** 管理中心是系统可观测性的中枢。它负责聚合来自所有`沙箱 Runtime`实例的日志、性能指标（如执行时间、内存使用）和事件，并通过统一的查询 API 对外提供服务，为开发者和运维人员提供强大的调试和监控能力。

### 4.4 协同工作模式
`沙箱管理中心`与`沙箱 Runtime`通过清晰的职责划分协同工作：管理中心负责“决策与调度”，Runtime 负责“执行与隔离”。这种架构不仅实现了高内聚低耦合，还为系统未来的水平扩展和功能演进奠定了坚实的基础。接下来，我们将使用 C4 模型来更直观地展示这一系统架构。

## 5.0 系统架构设计 (System Architecture using C4 Model)
### 5.1 C4 模型简介
为了清晰、多层次地呈现本沙箱技术方案的架构，我们选择采用 C4 模型。C4 模型通过上下文（Context）、容器（Container）、组件（Component）和代码（Code）四个层次，由高到低地逐步展开系统设计。这种方法使得不同角色的干系人（如业务分析师、开发者、运维工程师）都能在各自关心的抽象层次上理解系统架构，极大地促进了沟通和协作。

### 5.2 C1: 系统上下文图 (System Context Diagram)
![](https://cdn.nlark.com/yuque/__puml/9314c88f3b068699dab7c3d00b2afcc0.svg)



在最高层次的系统上下文中，我们将“沙箱执行平台”视为一个黑盒，重点描述其与外部世界的交互关系。

+ **核心系统:****沙箱执行平台 (Sandbox Execution Platform)**。
+ ** 用户与外部接入层  :**
    - **Data Agent:** 作为核心使用者，通过 API 提交代码（例如，由 PTC 生成的脚本）以在隔离环境中执行，并接收执行结果。
    - **开发者 (Developer):** 通过管理 API 或 SDK 配置沙箱模板、查询执行日志，并将沙箱平台集成到其应用程序中。
+ ** 基础设施与依赖层  :**
    - **代码仓库 (Code Repository, e.g., Git):** 沙箱在执行过程中可能需要克隆代码库。
    - **软件包索引 (Package Index, e.g., PyPI):** 沙箱环境在构建阶段需要从此类仓库拉取指定的依赖包。
    - **可观测系统（Observability System）**: 可观测数据上报系统 
    - **元数据存储 (Metadata Store, e.g., Mariadb):** 负责持久化存储沙箱模板、会话信息以及执行结果等关键数据。
    - **消息队列 (Message Queue, e.g., Redis/Kafka):** 用于缓冲高并发的执行请求，实现 API 服务与后端调度器的解耦，并处理执行结果的异步返回。
    - 

### 5.3 C2: 容器图 (Container Diagram)
![](https://cdn.nlark.com/yuque/__mermaid_v3/449c1747103372fc164a6e02d496ec6c.svg)

在容器层，我们将“沙箱执行平台”的内部结构进行分解，展示构成系统的、可独立部署的关键服务单元（“容器”）。

+ **核心容器:**
    1. **沙箱管理中心 (Sandbox Management Center):** 一个对外的 API 服务，作为系统的总入口和控制大脑。
    2. **消息队列 (Message Queue, e.g., Redis/Kafka):** 用于缓冲高并发的执行请求，实现 API 服务与后端调度器的解耦，并处理执行结果的异步返回。
    3. **元数据存储 (Metadata Store, e.g., Mariadb):** 负责持久化存储沙箱模板、会话信息以及执行结果等关键数据。
    4. **沙箱 Runtime 池 (Sandbox Runtime Pool):** 一组可动态伸缩的、真正执行代码的隔离环境。每个 Runtime 实例都是一个独立的进程或容器，负责执行单个会话。
+ **交互关系:**
    - **Data Agent** 或 **开发者** 通过 HTTPS 调用 **沙箱管理中心** 的 API。
    - **沙箱管理中心** 接收到执行请求后，将其封装成任务消息推送到 **消息队列** 的请求主题。
    - **沙箱管理中心** 的调度器从 **消息队列** 中获取任务，并根据模板从 **沙箱 Runtime 池** 中选择或创建一个合适的 Runtime 实例来执行。
    - **沙箱 Runtime** 在执行完毕后，将完整的`执行结果`发布到 **消息队列** 的一个专用响应主题。
    - **沙箱管理中心** 订阅该响应主题，获取结果并将其持久化到 **元数据存储**。

### 5.4 C3: 组件图 (Component Diagram)
在组件层，我们进一步深入 **“沙箱管理中心”和 “沙箱 Runtime”** 内部，展示其内部的逻辑组件及其职责划分。

#### 5.4.1 沙箱管理中心
![](https://cdn.nlark.com/yuque/__mermaid_v3/d0d665f9507f0fa6cb6f8789e8e121f9.svg)



+ **核心组件:**
    1. **API Layer :** 负责处理所有外部的 HTTP 和 WebSocket 请求，执行认证、鉴权和请求路由。
    2. **会话管理器 (Session Manager):** 核心业务逻辑组件，负责`执行会话`的创建、状态转换和生命周期管理。
    3. **调度器 (Scheduler):**  从 MQ/队列消费执行任务；根据模板、当前资源利用率与策略选择或创建合适的 Runtime 实例；负责 Warm Pool 的获取/释放/扩缩容决策。  
    4. **模板服务 (Template Service):** 提供对`沙箱模板`的 CRUD（创建、读取、更新、删除）管理功能。
    5. ** Result Handler : **订阅 MQ 的响应主题，接收 Runtime 发布的 Execution Result，进行格式校验、结果富化（添加审计/追踪信息）、并持久化到元数据存储；触发回调/通知给 API 层。**  **
    6. ** Message Queue Client（mq_client）: **MQ 的读写适配器：发布任务到 Request Topic，从 Response Topic 消费结果；负责消息确认、重试、死信队列（DLQ）。  **  **
    7. **监控与日志聚合器 (Monitoring & Log Aggregator):** 收集并聚合所有 Runtime 实例的日志和性能指标，提供统一的查询接口。

#### 5.4.2 沙箱 Runtime 层
![](https://cdn.nlark.com/yuque/__mermaid_v3/3401ee57b3f97413bf8aaff99c6d1ffe.svg)

+ **核心组件:**
    1. **Environment Initializer（init）： **基于 Template 构建运行时环境：选择 base rootfs 或预置镜像、挂载共享依赖缓存卷、准备工作目录、写入环境变量、创建虚拟环境（如 venv）或解包依赖层。  
    2.  **Policy Engine（policy_engine）**  : 将模板/系统的安全策略下发到隔离层：应用网络策略（allowlist/denylist）、系统调用过滤（seccomp profile）、文件系统权限、用户映射（uid/gid namespace）。  
    3. **Code Executor（executor）**: 在隔离环境中启动目标解释器/运行时（如 python3），执行用户代码；负责 stdin/stdout/stderr 的非阻塞处理，监控子进程状态（exit code、signals）。  
    4.  **Metrics Collector（metrics_collector）**  : 采集并上报每次执行的关键指标（duration_ms、cpu_time、memory_peak_mb、io_bytes、syscalls count）；产生用于调度/报警/计费的数据。  
    5.  **Result Assembler（result_assembler）** : 所有产物（exit_code, stdout, stderr, return_value, metrics）封装为 Execution Result 值对象，做 schema 校验与大小限制（超长日志的存储策略：将大型 stdout 切分或存对象存储并在 result 中引用）。  
    6.  **MQ Publisher（mq_publisher）**  :  把 ExecutionResult 发布到响应主题（Response Topic），保证消息至少一次或精确一次交付（由 MQ 类型决定），并处理重试与 DLQ。  



### 5.5 动态架构：执行流程时序描述
![](https://cdn.nlark.com/yuque/__mermaid_v3/b4274d5ed28be156023344f4031302c3.svg)

为了展示系统的动态行为，我们以“Data Agent 通过 PTC (程序化工具调用) 请求执行一个包含多步骤的 Python 脚本”为例，描述一次典型的交互流程：

1. **请求发起:** Data Agent 生成一个包含多步操作的 Python 脚本，并通过 SDK 调用**沙箱管理中心**的 `/execute` API 端点。
2. **会话创建:****API 网关** 验证请求后，将其转发给**会话管理器**。**会话管理器**根据请求内容（代码、模板 ID）创建一个新的`执行会话`对象，状态为 `PENDING`，并将其信息存入**元数据存储**。
3. **任务入队:****会话管理器**将该会话的 ID 封装成一个执行任务，推送到**消息队列**中。
4. **任务调度:****调度器**组件从**消息队列**中消费该任务。它首先从**模板服务**获取对应的`沙箱模板`配置，然后从**沙箱 Runtime 池**中请求一个匹配该模板的“温”实例。
5. **代码执行:****调度器**将完整的`执行会话`信息发送给选定的 **沙箱 Runtime**。Runtime 开始执行代码。
6. **流式输出 (可选):** 如果是 WebSocket 连接，**沙箱 Runtime** 会将执行过程中的 `stdout` 和 `stderr` 实时流式传输回**管理中心**，再由**API 网关**转发给 Data Agent。
7. **结果返回:** 代码执行完毕后，**沙箱 Runtime** 将包含 `exit_code`、`stdout`、`stderr` 和返回值的完整`执行结果`发布到**消息队列**的响应主题。
8. **状态更新与持久化:****管理中心**的对应组件消费结果消息，通知**会话管理器**更新会话状态为 `COMPLETED` 或 `FAILED`，并由会话管理器将最终的`执行结果`持久化到**元数据存储**。
9. **最终响应:** Data Agent 接收到包含完整执行结果的最终响应，并基于此结果进行下一步的推理和决策。



## 6.0 关键技术考量 (Special Considerations)
### 6.1 引言
在完成宏观架构设计后，本章节将深入探讨方案实现中几个至关重要的技术决策点。这些细节直接关系到系统的安全性、健壮性、性能以及对上层 AI 应用的友好度，是确保方案最终成功的关键。

### 6.2 执行结果的规范化返回
为了确保上游调用者（特别是逻辑严谨的 Data Agent）能够确定性地解析和处理执行结果，我们必须定义一个标准的、结构化的 JSON 返回格式。任何一次执行，无论成功与否，都应严格遵循此规范。

**标准返回格式定义:**

| 字段 (Field) | 类型 (Type) | 描述 (Description) | 示例 (Example) |
| --- | --- | --- | --- |
| `exit_code` | `integer` | 进程的退出状态码。`0` 通常表示成功，非零值表示失败。 | `0` |
| `stdout` | `string` | 捕获的标准输出流内容。 | `"Processing complete.\n"` |
| `stderr` | `string` | 捕获的标准错误流内容。 | `""` |
| `result` | `any` | Python 脚本的最终返回值，类型取决于脚本逻辑。 | `{"status": "ok", "data": [1, 2, 3]}` |
| `metrics` | `object` | 包含关键性能指标的对象，如 `duration_ms` 和 `memory_peak_mb`。 | `{"duration_ms": 75, "memory_peak_mb": 42.5}` |


这种规范化的返回格式，使得 Data Agent 可以通过解析 `exit_code` 和 `stderr` 来判断执行是否成功，并通过解析 `result` 来获取业务数据，极大地降低了 Agent 的“认知负載”，提升了系统的可靠性。

### 6.3 无网络场景下的 Python 依赖管理
![](https://cdn.nlark.com/yuque/__mermaid_v3/9c74da86774ba9036a1f76ba102ce3f9.svg)

在高安全要求的执行环境中，沙箱通常被严格禁止访问外部网络，这使得 Python 依赖管理成为必须解决的关键问题。为满足智能体任务中对不同依赖的使用需求，我们重点分析了两类可行策略，并形成明确的技术架构选型。

第一类是 **环境预置策略**：通过提前构建包含常用依赖的“黄金镜像”或预制 rootfs，在执行请求到来时直接启动对应沙箱实例。该方案具备启动速度快、环境确定性强等优势，但灵活性有限，一旦依赖发生变更就需重新构建镜像，同时镜像可能因包含大量未使用库而变得臃肿。

第二类是 **两阶段执行策略**（参考 pydantic/mcp-run-python 模型）：系统在具备受限网络访问权限的“准备环境”中，根据任务声明的依赖列表动态拉取和安装依赖，并将其写入独立缓存卷；随后在完全无网络的“执行沙箱”中，以只读方式挂载该依赖卷，保障核心代码执行阶段的绝对安全隔离。该方案灵活性极高，能适配多变的依赖需求，且安全性优于传统方案。

综合评估后，我们选择 **两阶段执行模型** 作为主方案，以满足 Data Agent 场景下需求多样、依赖复杂的特点。同时，对于部分稳定、性能敏感的模板场景，我们也将采用 **环境预置策略** 作为加速优化手段，在可接受一定镜像臃肿与灵活性下降的前提下实现最小启动延迟。

### 6.4 资源占用与性能的综合平衡
在多租户的沙箱环境中，实现资源与性能的平衡至关重要，既要防止“吵闹的邻居”问题，也要保证用户感知的低延迟。

+ 我们必须采用操作系统级别的机制，如 `**cgroups**`，对每个沙箱实例进行严格的资源限制。无论是 `bubblewrap` 还是容器化方案，都必须强制执行`沙箱模板`中定义的 CPU 和内存配额。这确保了单个任务的资源超用不会影响到同一物理节点上的其他任务，保证了系统的整体稳定性。
+ 对于一个可能在单个会话中执行数百个命令的 Data Agent 而言，一次性的启动成本与每个命令的累积延迟相比是微不足道的。每个命令 500 毫秒的延迟会使 Agent 无法使用，而 2 秒的冷启动时间则完全可以接受。因此，**“命令执行延迟”远比“冷启动时间”更为关键**。为了在不牺牲强隔离性的前提下优化性能，本方案的核心策略是维护一个 **“温沙箱池”（Warm Sandbox Pools）**。管理中心会预先启动并初始化一批沙箱实例。当新的执行请求到来时，可直接从池中获取一个“温”实例，从而将实际的会话启动延迟降至毫秒级，极大地提升了用户和 Agent 感知的响应速度。

### 6.5 基于临时区考虑沙箱技术选择
<font style="color:rgb(48, 48, 48);">“统一沙箱临时区”的核心诉求是：</font>**<font style="color:rgb(48, 48, 48);">文件持久化</font>**<font style="color:rgb(48, 48, 48);">（支持多次对话访问）和</font>**<font style="color:rgb(48, 48, 48);">复杂依赖兼容性</font>**<font style="color:rgb(48, 48, 48);">（支持 </font>`<font style="color:rgb(40, 42, 44);">pandas</font>`<font style="color:rgb(48, 48, 48);"> 等库用于 PTC 数据分析）。</font>

<font style="color:rgb(48, 48, 48);"></font>

| <font style="color:rgb(48, 48, 48);">特性</font> | <font style="color:rgb(48, 48, 48);">重量级隔离 (MicroVMs/容器)</font> | `<font style="color:rgb(40, 42, 44);">bubblewrap</font>`<font style="color:rgb(48, 48, 48);"> + </font>`<font style="color:rgb(40, 42, 44);">tmpfs</font>`<font style="color:rgb(48, 48, 48);"> 组合</font> | <font style="color:rgb(48, 48, 48);">结论：适用于临时区？</font> |
| --- | --- | --- | --- |
| **<font style="color:rgb(48, 48, 48);">隔离技术</font>** | **<font style="color:rgb(48, 48, 48);">内核/硬件级隔离</font>**<font style="color:rgb(48, 48, 48);"> (例如：Firecracker MicroVMs</font><font style="color:rgb(48, 48, 48);">)。</font> | **<font style="color:rgb(48, 48, 48);">OS 级别原生隔离</font>**<font style="color:rgb(48, 48, 48);">。</font>`<font style="color:rgb(40, 42, 44);">bubblewrap</font>`<font style="color:rgb(48, 48, 48);"> 在 Linux 上使用容器化和网络命名空间隔离</font><font style="color:rgb(48, 48, 48);">。</font> | **<font style="color:rgb(48, 48, 48);">两者都提供强大的隔离。</font>** |
| **<font style="color:rgb(48, 48, 48);">文件系统持久性</font>** | **<font style="color:rgb(48, 48, 48);">支持</font>**<font style="color:rgb(48, 48, 48);">。提供文件系统持久化和文件上传功能</font><font style="color:rgb(48, 48, 48);">。例如，</font>**<font style="color:rgb(48, 48, 48);">E2B</font>**<font style="color:rgb(48, 48, 48);"> 支持跨停止/启动周期的文件系统持久化</font><font style="color:rgb(48, 48, 48);">。</font>**<font style="color:rgb(48, 48, 48);">Cloudflare Sandbox SDK</font>**<font style="color:rgb(48, 48, 48);"> 支持挂载 S3 兼容的对象存储（如 R2）以实现数据持久性</font><font style="color:rgb(48, 48, 48);">。</font> | **<font style="color:rgb(48, 48, 48);">不支持</font>**<font style="color:rgb(48, 48, 48);">。</font>`<font style="color:rgb(40, 42, 44);">tmpfs</font>`<font style="color:rgb(48, 48, 48);"> 是基于 RAM 的临时文件系统。当沙箱容器因不活动而停止时，</font>**<font style="color:rgb(48, 48, 48);">所有文件和状态都会丢失</font>**<font style="color:rgb(48, 48, 48);">。</font> | **<font style="color:rgb(48, 48, 48);">重量级隔离胜出。</font>** |
| **<font style="color:rgb(48, 48, 48);">库和依赖兼容性</font>** | **<font style="color:rgb(48, 48, 48);">完全兼容</font>**<font style="color:rgb(48, 48, 48);">。可运行几乎任何 Python 库，包括带有 C 扩展的重型库（如 PyTorch）</font><font style="color:rgb(48, 48, 48);">。支持 </font>`<font style="color:rgb(40, 42, 44);">pip</font>`<font style="color:rgb(48, 48, 48);">、</font>`<font style="color:rgb(40, 42, 44);">npm</font>`<font style="color:rgb(48, 48, 48);">、</font>`<font style="color:rgb(40, 42, 44);">apt</font>`<font style="color:rgb(48, 48, 48);"> 等包安装</font><font style="color:rgb(48, 48, 48);">。</font> | **<font style="color:rgb(48, 48, 48);">受限。</font>**`<font style="color:rgb(40, 42, 44);">bubblewrap</font>`<font style="color:rgb(48, 48, 48);"> 是底层工具，不提供完整的包管理，难以支持复杂依赖的安装和运行。</font> | **<font style="color:rgb(48, 48, 48);">重量级隔离胜出。</font>** |
| **<font style="color:rgb(48, 48, 48);">性能/启动速度</font>** | <font style="color:rgb(48, 48, 48);">启动时间通常以秒计</font><font style="color:rgb(48, 48, 48);">。但 </font>**<font style="color:rgb(48, 48, 48);">MicroVMs</font>**<font style="color:rgb(48, 48, 48);"> 已优化至亚秒级启动（例如 </font>**<font style="color:rgb(48, 48, 48);">E2B</font>**<font style="color:rgb(48, 48, 48);"> 的 Firecracker 可达 </font>**<font style="color:rgb(48, 48, 48);">200 毫秒以内</font>**<font style="color:rgb(48, 48, 48);">）</font><font style="color:rgb(48, 48, 48);">。</font> | **<font style="color:rgb(48, 48, 48);">极快。</font>**`<font style="color:rgb(40, 42, 44);">bubblewrap</font>`<font style="color:rgb(48, 48, 48);"> 本身开销小，</font>`<font style="color:rgb(40, 42, 44);">tmpfs</font>`<font style="color:rgb(48, 48, 48);"> 读写速度快，适合轻量级任务</font><font style="color:rgb(48, 48, 48);">。</font> | **<font style="color:rgb(48, 48, 48);">各有优势</font>**<font style="color:rgb(48, 48, 48);">：轻量级更快，但功能受限。</font> |


<font style="color:rgb(48, 48, 48);">结论是：</font>**<font style="color:rgb(48, 48, 48);">重量级隔离</font>**<font style="color:rgb(48, 48, 48);">（MicroVMs/容器）</font>**<font style="color:rgb(48, 48, 48);">远比</font>**`<font style="color:rgb(40, 42, 44);">bubblewrap</font>`<font style="color:rgb(48, 48, 48);"> + </font>`<font style="color:rgb(40, 42, 44);">tmpfs</font>`**<font style="color:rgb(48, 48, 48);">更适合</font>**<font style="color:rgb(48, 48, 48);">作为统一沙箱临时区的技术基础，因为后者的组合</font>**<font style="color:rgb(48, 48, 48);">无法提供必需的文件持久化和完整库兼容性</font>**<font style="color:rgb(48, 48, 48);">。  
</font>

### 6.5 结论
本技术方案通过采用模块化的“管理中心 + Runtime”架构、支持多层次隔离技术、以及对规范化返回、依赖管理和性能优化等关键技术点的深思熟虑，旨在构建一个既安全可靠又高性能的下一代沙箱执行平台。它不仅解决了当前运行不受信代码所面临的直接挑战，更为重要的是，它将为 Data Agent 的进一步发展和落地提供坚实的基础设施，真正赋能能够编写并执行复杂逻辑的下一代人工智能应用。





## <font style="color:rgb(0, 0, 0);">7.0 Data Agent 沙箱执行平台 技术实施 Roadmap</font>
| **<font style="color:rgb(0, 0, 0) !important;">维度</font>** | **<font style="color:rgb(0, 0, 0) !important;">阶段一：基础能力构建（API/SDK + Bubblewrap Runtime）</font>** | **<font style="color:rgb(0, 0, 0) !important;">阶段二：平台化能力（沙箱管理中心 + 会话管理 + 调度器 + 沙箱模板）</font>** | **<font style="color:rgb(0, 0, 0) !important;">阶段三：沙箱能力升级（可观测  + 资源池化 + 分布式调度）</font>** |
| :--- | :--- | :--- | :--- |
| **<font style="color:rgb(0, 0, 0) !important;">核心目标</font>** | <font style="color:rgba(0, 0, 0, 0.85) !important;">实现 Python 代码的</font>**<font style="color:rgb(0, 0, 0) !important;">轻量级隔离执行</font>**<font style="color:rgba(0, 0, 0, 0.85) !important;">，提供开发者友好的 API/SDK，验证 bubblewrap 隔离能力与依赖管理可行性</font> | <font style="color:rgba(0, 0, 0, 0.85) !important;">构建完整的沙箱管控平台，实现会话生命周期管理、调度，支撑完整的沙箱平台化能力</font> | <font style="color:rgba(0, 0, 0, 0.85) !important;">沙箱能力升级，提升可观测性，支持Runtime池管理，满足高并发场景</font> |
| **<font style="color:rgb(0, 0, 0) !important;">交付版本</font>** | <font style="color:rgba(0, 0, 0, 0.85) !important;">ADP 5.0.0.4</font> | <font style="color:rgba(0, 0, 0, 0.85) !important;">ADP 5.0.0.5</font> | <font style="color:rgba(0, 0, 0, 0.85) !important;">ADP 5.0.0.6</font> |
| **<font style="color:rgb(0, 0, 0) !important;">核心任务拆解</font>** | <font style="color:rgba(0, 0, 0, 0.85) !important;">1. </font>**<font style="color:rgb(0, 0, 0) !important;">SandBox Runtime </font>**<font style="color:rgba(0, 0, 0, 0.85) !important;"> ：</font><br/><font style="color:rgba(0, 0, 0, 0.85) !important;">① 基于已有版本优化改造，解决沙箱性能问题，提供独立的沙箱部署包</font><br/><font style="color:rgba(0, 0, 0, 0.85) !important;">② 实现</font><font style="color:rgb(0, 0, 0) !important;">预装依赖管理，满足函数使用常见依赖库支持</font><br/><font style="color:rgba(0, 0, 0, 0.85) !important;">③ 执行结果的规范化返回，封装 Execution Result 标准化输出（exit_code/stdout/stderr/result/metrics）</font><br/><font style="color:rgba(0, 0, 0, 0.85) !important;">2. </font>**<font style="color:rgb(0, 0, 0) !important;">API/SDK 开发</font>**<font style="color:rgba(0, 0, 0, 0.85) !important;">  </font><br/><font style="color:rgba(0, 0, 0, 0.85) !important;">① 提供 Python SDK 核心 API：</font>`~~<font style="color:rgb(0, 0, 0);">create_session</font>~~<font style="color:rgb(0, 0, 0);">()</font>`<br/><font style="color:rgba(0, 0, 0, 0.85) !important;">/</font>`<font style="color:rgb(0, 0, 0);">execute_code()</font>`<br/><font style="color:rgba(0, 0, 0, 0.85) !important;">/</font>`~~<font style="color:rgb(0, 0, 0);">get_result</font>~~<font style="color:rgb(0, 0, 0);">()</font>`<br/>~~<font style="color:rgb(0, 0, 0);">/ kill_session()</font>~~<br/><font style="color:rgb(0, 0, 0);"></font><br/><font style="color:rgb(0, 0, 0);"></font><br/><font style="color:rgb(0, 0, 0);"></font><br/><font style="color:rgb(0, 0, 0);"> </font><br/><font style="color:rgb(0, 0, 0);"> </font> | <font style="color:rgba(0, 0, 0, 0.85) !important;">1.</font><font style="color:rgba(0, 0, 0, 0.85) !important;"> </font>**<font style="color:rgb(0, 0, 0) !important;">沙箱管理中心开发</font>**<font style="color:rgba(0, 0, 0, 0.85) !important;"> </font><br/><font style="color:rgba(0, 0, 0, 0.85) !important;"> ① </font>**<font style="color:rgb(0, 0, 0) !important;">API Layer</font>**<font style="color:rgba(0, 0, 0, 0.85) !important;">：提供 HTTP/WebSocket 接口，支持同步调用 + 流式输出，实现认证鉴权  </font><br/><font style="color:rgba(0, 0, 0, 0.85) !important;">② </font>**<font style="color:rgb(0, 0, 0) !important;">核心组件开发</font>**<font style="color:rgba(0, 0, 0, 0.85) !important;">：   </font><br/><font style="color:rgba(0, 0, 0, 0.85) !important;">- 会话管理器：实现 Session 状态机（PENDING/RUNNING/COMPLETED/FAILED）、超时回收   </font><br/><font style="color:rgba(0, 0, 0, 0.85) !important;">- 调度器：支持模板匹配、调度沙箱 Runtime实例   </font><br/><font style="color:rgba(0, 0, 0, 0.85) !important;">- 模板服务：提供沙箱模板 CRUD 与版本管理  </font> ，支持沙箱模板本地配置（资源限制、依赖列表、网络策略<br/><font style="color:rgba(0, 0, 0, 0.85) !important;">- Result Handler：消费 MQ 结果、持久化到元数据存储、触发回调  </font><br/><br/><font style="color:rgba(0, 0, 0, 0.85) !important;">2. </font>**<font style="color:rgb(0, 0, 0) !important;">中间件集成</font>**<font style="color:rgba(0, 0, 0, 0.85) !important;"> </font><br/><font style="color:rgba(0, 0, 0, 0.85) !important;"> ① 消息队列集成（Kafka/Redis）：实现请求异步入队、结果异步返回 </font><br/><font style="color:rgba(0, 0, 0, 0.85) !important;"> ② 元数据存储（MariaDB）：存储模板配置、会话信息、执行结果  </font> | <font style="color:rgba(0, 0, 0, 0.85) !important;">1. </font>**<font style="color:rgb(0, 0, 0) !important;">沙箱管理中心开发</font>**<font style="color:rgba(0, 0, 0, 0.85) !important;"> </font><br/><font style="color:rgba(0, 0, 0, 0.85) !important;">① </font>**可观测性与调试 (Observability & Debugging):**<font style="color:rgba(0, 0, 0, 0.85) !important;"> 聚合来自所有</font>`<font style="color:rgba(0, 0, 0, 0.85) !important;">沙箱 Runtime</font>`<font style="color:rgba(0, 0, 0, 0.85) !important;">实例的日志、性能指标（如执行时间、内存使用）和事件，并通过统一的查询 API 对外提供服务，为开发者和运维人员提供强大的调试和监控能力。</font><br/><font style="color:rgba(0, 0, 0, 0.85) !important;">②</font>**缓存与队列管理 (Caching & Queue Management):**<br/>+ **温沙箱池 (Warm Sandbox Pools):**<font style="color:rgba(0, 0, 0, 0.85) !important;"> 构建</font><font style="color:rgb(0, 0, 0) !important;">温沙箱池</font><font style="color:rgba(0, 0, 0, 0.85) !important;">，支持 Runtime 实例注册到管理中心，上报负载状态  ：实现 Runtime 与管理中心的心跳机制、任务分发与结果上报</font><br/>+ **智能调度与负载均衡:** 当接收到新的执行请求时，它会解析请求中指定的`沙箱模板`类型（如 `bubblewrap` 或 `E2B`），然后从相应的资源池中选择一个负载最低的`沙箱 Runtime`实例来执行任务。如果池中没有可用实例，调度器将负责按需创建新的实例。 |
| **<font style="color:rgb(0, 0, 0) !important;">  </font>** | <font style="color:rgba(0, 0, 0, 0.85) !important;"></font> | <font style="color:rgba(0, 0, 0, 0.85) !important;"></font> | <font style="color:rgba(0, 0, 0, 0.85) !important;"></font> |


  


