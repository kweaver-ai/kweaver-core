# 相关需求设计
- feature-799460 对逻辑属性进行召回
- feature-799472 支持行动召回
- feature-803607 MCP 类型行动召回支持
- feature-806350 将kn_rerank和kn_search从data-retrieval迁移至agent-retrieval
- [feature-806350](feature-806350/) KnSearch、KnowledgeRerank 模块（逻辑设计、实现设计、OpenAPI3、接口测试用例）
- [feature-mcp-server](feature-mcp-server/) Agent Retrieval 作为 Streamable HTTP MCP Server 暴露 kn_search 工具（kn_id 通过 Header X-Kn-ID 配置、用户配置与使用示例）
