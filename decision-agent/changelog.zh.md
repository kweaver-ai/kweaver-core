# 版本 changelog 说明

## 0.7.1

### 功能与改进

- 新增带校验的 Agent `Config` 值对象，完善 Dolphin 模式配置校验
  - 为 Agent 配置和 Dolphin 模式补充 `Config` 结构与测试覆盖
  - 同步 Agent Mode 文档和仓库 Agent 指引中的配置结构说明

### Bug 修复

- 修复内置工具名称以符合 DeepSeek tool naming 规范
  - issue: https://github.com/kweaver-ai/kweaver-core/issues/390
  - 将 `获取agent详情` 重命名为 `get_agent_detail`
  - 移除已废弃的 `Agent可观测数据查询API` 工具（对应接口已下线）
- 在工具注册时增加名称合法性校验，提前发现不合规名称
  - 校验工具名称仅包含 `a-z`、`A-Z`、`0-9`、`_`、`-`，且长度不超过 64 字符
  - 对不合规的工具名称记录警告日志，避免在 DeepSeek 模型下运行失败
- 升级 `kweaver-dolphin` 依赖到 v0.7.6，修复 DeepSeek v4 兼容性问题
  - 包含 v0.7.5 过渡升级和 v0.7.6 兼容性修复

### 重构与清理

- 优化内置工具箱描述文案，提升可读性
  - 改进搜索工具、记忆管理、Agent配置、联网搜索、沙箱代码执行等工具的描述
- 重构 Skill 内置工具注入机制和服务边界
  - 将 Skill contract tools 移入 agent core logic 层，通用工具装配与平台内置工具注入解耦
  - 将 Skill API 能力切换为 Agent Operator integration service，修复错误的下游服务依赖
  - 新增 Skill HTTP 端点、请求/响应模型和 OpenAPI 工具定义
- 重组用户手册 examples，并补充共享状态管理
  - 按能力目录拆分 API、CLI、TypeScript SDK 示例和 Makefile target
  - 新增共享环境与状态处理，便于示例流程复用

### 文档

- 新增完整的 Decision Agent 用户手册体系
  - 补充 API、CLI、概念、TypeScript SDK 指南及聚合版文档
  - 新增可运行的 API、CLI、SDK 示例，并更新 help 文档和 OpenAPI 产物
  - 新增 Agent 临时区（Sandbox Workspace）技术说明
- 新增 Decision Agent 集成 cookbook 场景文档
  - 覆盖合同摘要、Sub-Agent 合同审查、人工干预/终止等 API、CLI、SDK 和场景指南

## 0.7.0

### 功能与改进

- 在 agent 中新增 React Agent 模式支持
  - issue: https://github.com/kweaver-ai/kweaver-core/issues/171
  - docs 目录: decision-agent/agent-backend/agent-factory/docs/feature/03-agent-mode
- 新增 LLM 消息日志功能，用于调试和监控
  - 引入 `LLMMessageLoggingConfig` 配置类
  - 支持通过 agent-executor.yaml 配置消息日志
  - 更新 agent_core_v2.py 处理 LLM 消息日志标志和参数
  - 【注意】：这个dolphin sdk还没有适配，需要等dolphin sdk适配后才能使用
- 新增 Agent 模式枚举并重构配置结构，提升可扩展性
- 更新 kweaver-dolphin 依赖到 v0.6.0

### Bug 修复

- 修复复制 Agent 模板时未清除已发布字段的问题

### 重构与清理

- 简化可观测性系统并移除冗余组件
  - 使用新的结构化模型重构可观测性配置
  - 移除已废弃的可观测性处理器，启用 OpenTelemetry 追踪
  - 整合各服务的遥测配置
  - 清理环境变量和 configmap 模板
- 移除 benchmark 相关代码并优化 API 文档
- 移除已废弃的服务配置和 disable_biz_domain_init 配置

### 文档

- 增强 API 文档，在 Swagger 基础上新增 Redoc 支持
- 更新 Swagger 路由和文档结构
- 更新 API Chat 页面文档信息并新增 API Chat API 集成指南
  - issue: https://github.com/kweaver-ai/kweaver-core/issues/173
  - doc: decision-agent/agent-backend/agent-factory/docs/api-chat-integration.zh.md
- 更新 API 文档和版本信息

## 0.6.0

### 功能与改进

- 为 `agent-factory` 新增 Skill 类型支持
  - 添加带校验的 `SkillSkill` 值对象
  - 在 agent 创建/详情/更新处理器及运行服务中支持 Skill 配置
- 为 `agent-executor` 新增 TraceAI Evidence 请求头支持
  - 在 `FeaturesConfig` 中引入 `enable_traceai_evidence` 功能开关
  - 在 API 工具代理请求中注入 `X-TraceAi-Enable-Evidence` 请求头
- 添加 v0.6.0 数据库迁移：Skill 相关表及 `agent-memory` 历史表（dm8 和 mariadb）

### 重构与清理

- 重构 `agent-factory` 发布请求校验：使用构造函数语义，对请求字段进行校验和清洗，校验失败时返回 400 并携带明确错误信息，而非 500
- 重组数据库迁移文件结构，移除冗余的 `pre` 目录层级
- 从部署中移除未使用的 `ENABLE_EVIDENCE_EXTRACTION` 环境变量和字典文件；更新 `agent-executor` 依赖并优化 Dockerfile

## 0.5.3

### 重构与清理

- 简化 `agent-factory` 的 API Chat 路由：改从请求体获取 `agent_key`，路由更新为 `/api/agent-factory/v1/api/chat/completion`，同步更新 Swagger 文档及 `agent-web` API 文档组件
- 为 `agent-factory` 新增基于 handler 代码注释的 OpenAPI 3 文档自动生成能力，提供 Scalar 文档服务（`/scalar`）、运行时动态 server URL 回填及嵌入式静态产物（JSON、YAML、HTML、favicon）

### 测试

- 为 API Chat 处理器和 API 文档恢复对话流程补充单元测试

## 0.5.2

### Bug 修复

- 修复 agent-executor dolphin 路径中的日志方法调用
  - 将 dolphin trace listener 创建失败时的 `warning` 改为 `warn`
  - 将 run-agent options 回退到自动生成会话 ID 时的 `warning` 改为 `warn`

## 0.5.1

### Bug 修复

- 修复开启长期记忆后可能报错的问题
  - 将召回记忆改为以 system message 的形式追加到 `_history`，避免错误写入历史变量导致执行异常
  - 在注入提示词前先序列化记忆召回结果，确保后续提示拼装过程稳定可用

### 可观测性

- 改进 `agent-executor` 的 OpenTelemetry 故障诊断日志
  - 将 trace wrapper 中直接输出到标准输出的调试信息统一改为结构化日志
  - 在 trace provider 初始化失败时补充完整 traceback，便于排查问题

## 0.5.0

### 功能与改进

- 为 `agent-factory` 和 `agent-executor` 升级 OpenTelemetry 可观测能力
  - 在 `agent-factory` 中初始化新版 OTel 链路，并开放服务名、版本、环境、OTLP 端点、Trace 采样率和日志级别等配置
  - 统一 Agent ID、用户 ID、会话 ID 和操作名等 GenAI Trace 属性，便于请求链路与下游 Span 关联排查
  - 将 `agent-executor` 的 Trace 配置统一到 OTLP 端点和基于环境变量的采样设置
- 为 `agent-executor` 的 API 工具新增 evidence 提取能力
  - 支持从 API 工具返回结果中的 `nodes` 字段提取 `_evidence` 结构
  - 新增 `ENABLE_EVIDENCE_EXTRACTION` 开关，用于按环境控制该行为
- 增强部署与运行控制能力
  - 为 Sandbox Platform 开关和业务域开关补充 Helm 渲染与配置支持
  - 当 `disable_pms_check` 开启时，内部 agent-app 权限中间件支持跳过用户 ID 强校验

### 文档

- 补充 `agent-factory` 与 `agent-executor` 生成版 API 文档
- 为新版 OTel 配置补充 Helm 组件元数据说明

### 前端 (agent-web)

- Bug 修复：支持对话中断场景下 list 类型的中断参数。

## 0.4.4

### 前端 (agent-web)

- Bug 修复：修复 Agent 配置时无法添加知识条目的问题。
- Bug 修复：修复 Agent 配置时无法添加指标的问题。

## 0.4.3

### 功能与改进

- 新增业务域禁用功能 (DisableBizDomain)
  - 在 SwitchFields 中新增 DisableBizDomain 配置选项
  - 实现 IsBizDomainDisabled() 辅助方法
  - 更新代理配置服务以支持业务域禁用
  - 修改个人空间和已发布服务以适应业务域禁用
  - 更新 OAuth 中间件以处理禁用的业务域
  - 添加配置示例和文档更新
  - 部署 Helm chart 配置更新

### 测试

- 为业务域禁用功能添加全面的单元测试覆盖
  - 添加 disable_biz_domain_test.go 测试文件
  - 更新相关服务的测试用例
  - 增强中间件测试覆盖

### 修复问题
- Bug 修复：工具初始化阶段遇到不可用工具时，降级为跳过问题工具，而不是直接导致整个 `dolphin_run` 请求失败
  - 将 `get_tool_info` 调整为只记录工具可用性错误日志并返回空结果，不再直接抛出异常
  - 从 `skills.tools` 中移除不可用工具，确保其余可用工具仍能继续加载和执行
  - 为沙盒 execute-sync OpenAPI 的 `session_id` 路径参数补充默认值 `sess-agent-default`
  - 补充单元测试，覆盖问题工具过滤和工具信息失败降级路径

- Bug 修复：修复 agent-executor 进程被杀死等其他异常时，对话状态未更新为 failed 的问题
  - 修改 chat_process.go 文件，在 errChan 关闭或收到 EOF 错误时设置 messageChanClosed = true
  - 确保 agent-executor 进程被杀死或其他异常时，对话状态能够正确更新为 failed
  - 避免 UI 上显示状态为 running，且重新打开历史对话时报错 "conversation_id not found"

- Bug 修复：修复用户中止对话后，assistant 消息内容为空的问题
  - 修改 HandleStopChan 函数，当用户点击中止按钮时，将 session 中已输出的临时消息内容更新到数据库
  - 确保中止对话的内容能够在对话列表中正常显示

## 0.4.2

### 前端 (agent-web)

- Bug 修复：Agent 配置界面移除对旧业务知识网络判断逻辑。
- Bug 修复：Agent 配置界面移除对文档的判断逻辑。

## 0.4.1

### 功能与改进

- 在 `agent-factory` 中新增对话历史配置功能
  - 支持通过配置接口设置历史对话保留策略
  - 实现按对话条数控制历史记录（默认策略）
  - 预留按时间窗口和 token 占用量的配置选项
  - 支持无历史对话的配置模式
  - 新增 `conversationHistoryConfig` 配置结构和相关枚举类型
  - 添加历史对话数量限制常量（范围 1-1000）

### Bug 修复

- 移除重复的参数校验：调用方已对 `historyLimit` 参数进行校验，后续函数不再重复校验
- 修正 count limit 配置范围说明：更新为 1-1000

### 测试

- 为对话历史配置验证逻辑添加单元测试
- 为历史对话获取边界场景添加测试用例

### 文档

- 新增历史对话策略配置设计文档

## 0.4.0

### 功能与改进

- 增强 `agent-executor` 日志目录创建的权限处理
  - 实现目录权限处理，使用 0o755 模式
  - 添加 LOG_DIR 写权限检查和适当的错误处理
  - 为新权限检查逻辑添加单元测试
- 修复 `agent-executor` 中缺少的 agent router v1 注册
  - 导入并注册 agent_router_v1，与现有的 v2 路由并存
  - 恢复 v1 端点的完整 API 路由功能

## 0.3.6

### 功能与改进

- 在 `agent-factory` 中新增按 Agent ID 或 Agent Key 查询智能体的能力，并同步更新相关对话 / API 文档流程
- 在 `agent-executor` 中提升 API 工具超时缓冲时间，增强长耗时请求稳定性

### Bug 修复

- 修复基于 Agent Key 的权限校验问题：校验前先解析出真实 Agent ID
- 更新沙盒会话 ID 为代码执行要求的默认格式 `sess-agent-default`

### 重构与清理

- 更新 square 服务接口及相关服务逻辑，统一复用新的 ID / Key 查询路径
- 优化 `api_doc.go` 中的步骤标注，提升代码可读性

### 测试

- 为 API 工具超时处理和按 ID / Key 查询智能体补充单元测试
- 更新 `agent-factory` 相关服务测试，覆盖新的查询逻辑和权限校验流程

### 其他

- 项目版本升级到 0.3.6


## 0.3.5

### 功能与改进

- 为 decision-agent 添加 CI/CD 工作流和代码质量改进
- 添加 pre-commit 钩子并优化代码格式化
- 增强 docker-compose 部署，支持可配置端口和自动化
- 为领域服务和基础设施添加全面的单元测试
- 为 agent-factory 添加全面的单元测试
- 为 agent-executor 添加配置文件并更新 gitignore

### Bug 修复

- 修复 EnsureSandboxSession 在重建会话前先删除失败状态的会话
- 修复恢复中断对话时未清除 Ext 中的 InterruptInfo 的问题
- 修复 agent-web 独立运行时路由匹配错误
- 修复 agent-web 独立运行时静态资源加载错误

### 重构与清理

- 修复临时区会话 ID，检查对话中的 name 属性值，移除旧文件上传组件 (agent-web)
- 移除未使用的导入并优化代码结构 (agent-executor)
- 移除已废弃的空间相关模块
- 重构单元测试以避免环境变量竞争条件
- 重命名 helm 模板并移除未使用的 configmap
- 将 agent-factory HTTP 访问层替换为内部服务调用
- 改进环境变量加载并添加 air 热重载支持
- 移除未使用的 DelInternationalPath 函数
- 重构 agent-web Dockerfile 以支持 Docker Compose

### 测试

- 为 DTO 和基础设施层添加全面的单元测试
- 为基础设施和处理器层添加全面的单元测试
- 为领域服务和值对象添加全面的单元测试
- 为 agentrunsvc 和相关服务添加单元测试
- 为 agentconfigsvc、agentrunsvc 和 releasesvc 添加单元测试
- 为智能体导入/导出功能添加全面的测试用例
- 为 tool_v2 API 工具包模块添加单元测试
- 为 agent factory 配置添加全面的测试覆盖
- 为依赖模块和启动模块添加测试
- 为异常处理器和路由中间件模块添加测试
- 扩展 tool_requests 和 json 模块的测试

### 文档

- 添加代码质量指南和 lint 说明
- 添加 Claude Code 指导文档 (CLAUDE.md)

### 其他

- 为 agent-executor 添加 pre-commit 配置并更新开发工作流
- 添加覆盖率配置以排除测试和构建文件
- 更新生产部署配置文件
- 更新 PyInstaller spec 以包含 setuptools 依赖

## 0.3.4

### Bug 修复

- 修复智能体开启记忆功能后对话报错的问题
- 添加 dbutilsx 依赖并重构记忆配置解析逻辑
- 更新 numpy 最低版本到 1.23.5 以提高兼容性
- 修复记忆召回重排序时 relevance_score 为 None 导致的比较错误

## 0.3.3

### 前端 (agent-web)

- 修复我的Agent模板页面筛选状态异常的bug


## 0.3.2

### Bug 修复

- 修复已知问题

## 0.3.1

### 前端 (agent-web)

- Agent 对话界面临时区支持隐藏
- 修复 Agent 对话界面临时区显示异常的 bug
- 页面所有 Decision Agent 中文显示为"决策智能体"
- 修复 Agent API 页面调试接口时缺少 x-business-domain 请求头
- 移除未使用的 KNSpaceTree、DocTree、ContentDataTree 组件

## 0.3.0

### 功能与改进

- 新增沙盒平台集成，支持代码执行和文件管理
- 新增 PyCodeGenerate Agent 及沙盒执行工具
- 为 agent-factory 添加 Swagger 文档支持
- 调试接口添加 SelectedFiles 字段
- Helm Chart 添加开发控制开关和新配置选项
- 更新 kweaver-dolphin 依赖到 v0.2.4

### Bug 修复

- 修复用户跳过中断工具步骤时执行状态显示错误
- 修复中断时 JSON 值类型渲染错误
- 修复调试模式文件显示问题
- 修复调试模式文件参数传递方式

### 重构与清理

- 移除 EcoIndex 和 DataHubCentral 配置
- 移除废弃的 doc_qa 和 graph_qa 工具
- 从 agent-executor 移除 pandas 依赖
- 移除废弃的服务类和提示工具
- 移除未使用的数据访问对象和旧 dataset 表
- 简化配置并移除未使用的停用词文件配置
- 隐藏 Agent 轨迹分析模块
- 前端移除批量检查索引状态接口调用

### 测试

- 为 agent-memory 模块添加全面的单元测试
- 添加配置加载器和工具函数的单元测试

### 前端 (agent-web)

- 支持中断工具执行的跳过状态(skipped)显示
- Agent 对话适配沙盒文件上传
- 从 Agent 配置界面移除文件相关配置
- 从调试模式移除文件删除和预览功能

## 0.2.3

### Bug 修复

- 修复 Agent 角色指令模式下 Agent 运行白屏的问题

## 0.2.2

### Bug 修复

- 修复前端 Agent 中断参数传递问题
- 修复对话界面白屏问题
- 修复输入配置类型下拉选择框失效问题
- 修复模版创建 Agent 跳转 404 错误
- 修复 agent-memory 权限错误并提升可观测性

### 功能与改进

- 新增工具中断恢复支持，通过统一 Run API 实现
- 将 agent-executor 中的 TelemetrySDK 设为可选依赖
- 优化消息扩展结构并添加状态处理
- 简化中断处理和类型转换
- 优化聊天恢复，使用统一 DTO 类型和中断恢复机制

### 前端 (agent-web)

- 支持脱离微前端独立运行
- 优化中断流式聊天接口，仅传递用户更改的参数
- 移除冗余的 changelog 文件

## 0.2.1

### Bug 修复

- 解决 agent-web 安装阻塞问题
- 修复 Agent 检索功能 (#37, #38)

### 基础设施

- 将 Helm Chart 从 agent-factory 重命名为 agent-backend
- 删除 tests/tools 中的编译产物以减少仓库大小

### 文档

- 更新近期变更的 changelog

## 0.2.0

### 架构与部署

- 统一多服务 Docker 架构,使用 supervisor 进程管理
- 修复 agent-factory Helm Chart 配置问题
- 添加缺失的服务配置 (agent_executor, efast, docset, ecoconfig, uniquery)
- 修复 volumeMounts 使用 subPath 精确挂载
- 更新 securityContext runAsUser/runAsGroup 为 1001
- 支持 GOPROXY 优化 Docker 构建
- 启用 mq-sdk 和 telemetrysdk-python 依赖

### Agent 中断与恢复

- 新增 Agent 中断和恢复功能
- 自定义 ToolInterruptException 处理工具中断
- 修复中断会话的进度处理
- 前端适配中断操作

### Agent Executor

- 将 agent-executor 模块迁移到 agent-backend 目录
- 添加 PascalCase 函数名的向后兼容别名
- 修复 memory handler 参数处理
- 重构工具中断处理和 DTO 命名

### Agent Factory

- 新增 agent-factory-v2 完整实现,采用 DDD 架构
- 重构 httpserver 模块,支持旧路径配置
- 添加流式响应日志和改进请求日志
- 启用 keep_legacy_app_path 配置

### 前端 (agent-web)

- Agent 流式接口支持 agent_run_id 参数
- 工具配置支持确认提示
- 修复添加技能时 MCP 树节点无法展开的 bug
- 修复部署文件中的 YAML 语法错误
- 菜单注册更新

### 代码质量与重构

- 移除 agent-go-common-pkg 外部依赖
- 迁移 DolphinLanguageSDK 导入到新的 dolphin 包结构
- 移除废弃的函数错误类
- 简化 Dockerfile 统一复制命令
- 添加 opencode 工作流用于自动化代码审查
- 删除 tests/tools/fetch-log/build 中的编译产物以减少仓库大小
- 更新 .gitignore 排除构建产物和日志文件

### 数据检索 (Data Retrieval)

- 添加 Jupyter Gateway runner 用于代码执行
- 添加代码运行器工具 (exec_runner, ipython_runner)
- 增强 DIP 服务集成
- 添加 MCP 测试工具和示例
- 添加 text-to-DIP 指标工具和提示
