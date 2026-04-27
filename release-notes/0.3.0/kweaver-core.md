# KWeaver Core 0.3.0 Release Notes

This release focuses on deep integration of business and technology alongside a comprehensive overhaul of the agent architecture. The new sandbox architecture achieves separation of reasoning (LLM) and logic (Code), and combined with action-driven visualization and human intervention mechanisms, delivers a full-stack solution for building safe, controllable, and reliable enterprise-grade agent applications.

---

## Highlights

1. **Redesigned Sandbox Architecture**: Achieves separation of reasoning (LLM) and logic (Code) responsibilities, significantly reducing token consumption. Integrated temporary area management provides session-level data isolation and intermediate state persistence, creating a safer, more stable, and more efficient execution environment for agents.
2. **Agent Human Intervention**: Leveraging the semantic modeling capabilities of the business knowledge network, agents can precisely recall Actions based on task context. A human intervention mechanism is introduced for sensitive operations (e.g., approval, deletion), pausing execution to request human confirmation and display decision rationale — keeping business processes "intelligent yet controlled."
3. **Action-Driven Frontend**: Added visual configuration and task management interfaces for intuitive management of actions and task lists, bridging the gap between backend logic definition and frontend visual management.
4. **MCP Standardized Integration**: Built a Context Loader MCP Server providing a unified service interface based on the MCP standard. Integration requires only a URL, Token, and business knowledge network ID, with support for Claude Desktop, Cursor, Codex, and other mainstream development tools.
5. **Enhanced Data Semantics**: VEGA data virtualization introduces feature fields as semantic capability enhancements. Supports configuring multiple feature types per view field (Keyword exact match, Full-Text full-text search, Vector semantic retrieval), with a default feature mechanism to simplify queries.

---

## 1. BKN Engine

**1. Action-Driven Frontend Configuration**

Supports frontend configuration of action-driven tasks with visual interfaces for defining trigger rules and execution logic:

- **Event-driven triggers**: Supports trigger condition configuration based on object class data attributes. When specific business rules are met, predefined actions are triggered automatically.
- **Visual configuration**: Wizard-style guidance and form-based configuration allow business users to define action-driven workflows without writing code.

**2. Task Management and Execution Status**

Provides task management capabilities for monitoring action execution:

- **Execution monitoring**: View action execution status in the task management interface, including total count, success count, and failure count.
- **Execution records**: View the execution record for each action instance, including execution time, parameters, and results.
- **Async execution**: Action-driven tasks use an asynchronous execution model, returning immediately after submission to support efficient batch action processing.

---

## 2. VEGA Data Virtualization

**1. Feature Fields for Views**

Enhanced semantic expression capabilities for views, providing richer data support for the business knowledge network:

- **Multi-type feature support**: Supports configuring multiple feature types for view fields:
  - Keyword: For exact match queries such as IDs and codes.
  - Full-Text: For full-text search scenarios with fuzzy matching and tokenized queries.
  - Vector: For vectorized retrieval, enabling semantic search and similarity queries.
- **Default feature mechanism**: Each feature type can have one default, automatically used when no feature is specified in a query.
- **Flexible feature configuration**: Supports dynamically adding or modifying feature definitions after view creation.
- **Semantic retrieval**: Feature fields can be indexed and retrieved by the business knowledge network, enabling agents to perform precise data queries and reasoning based on field semantics.

---

## 3. Decision Agent

**1. Deep Integration with Context Loader**

Significantly improved action parameter recognition accuracy through deep integration of Decision Agent and Context Loader:

- **Context enrichment**: Context Loader provides Decision Agent with rich business context including concept definitions, instance data, and logical properties, supplying sufficient semantic information for parameter recognition.
- **Semantic reasoning**: Decision Agent performs intelligent reasoning based on the business knowledge network's semantic information, accurately identifying required action parameters and reducing execution failures from missing or incorrect parameters.

**2. Human Intervention Mechanism**

A human intervention mechanism for sensitive or high-risk actions:

- **Intervention toggle**: Supports configuring a "enable human intervention" option when defining action classes. Sensitive operations (e.g., contract approval, data deletion) pause before execution to await human confirmation.
- **Confirmation UI**: When human intervention is triggered, the system displays action details and parameter information; the operator can choose to "Execute" or "Skip."

**3. Temporary Area Redesign**

Redesigned the temporary area to improve efficiency and stability of intermediate state management:

- **Data isolation**: Each agent session gets an independent temporary data storage area, achieving session-level isolation to prevent data contamination.
- **State persistence**: Intermediate states are persisted to ensure long-running actions are unaffected by system restarts or network disruptions.
- **Performance optimization**: Optimized temporary area read/write performance to reduce the impact of state management on action execution efficiency.

**4. Enhanced File Analysis: Separation of Reasoning and Logic**

The new sandbox architecture brings revolutionary improvements to agent file analysis:

- **Direct file upload**: Files can be uploaded directly to the sandbox, eliminating repeated downloads each conversation turn and significantly improving execution efficiency.
- **Massive token savings**: Instead of loading entire document content into the context, the LLM generates Python code executed in the sandbox — saving over 90% of token consumption for large documents.
- **Intelligent code generation**: The LLM automatically generates data analysis code based on user questions, executes it safely in the sandbox, and returns results — supporting complex data processing and statistical analysis.
- **Minimal configuration**: Only a single "execute code in sandbox" tool needs to be configured for the system to handle complex file analysis tasks.

---

## 4. Context Loader

**1. MCP Standardized Integration**

Built a Context Loader MCP Server implementing seamless integration with mainstream development tools based on the MCP (Model Context Protocol) standard:

- **Minimal configuration**: Integration requires only a URL, account, user token, and business knowledge network ID — perfectly compatible with Claude Desktop, Cursor, Codex, and other clients.
- **Automatic tool discovery**: The system automatically loads all tool descriptions defined in the business knowledge network; agents can dynamically discover and invoke available tools without additional development.

**2. Context Loader and Sandbox Integration**

Deep integration of Context Loader with the sandbox environment:

- **Sandbox function bridging**: Custom functions are executed through the sandbox environment, achieving code isolation that prevents malicious code from damaging the system and ensures production safety.

---

## 5. Dataflow

**1. Python Runtime Sandboxing**

- **File upload to sandbox**: Supports uploading files (e.g., Excel, CSV) directly in node configuration; files are automatically synced to the sandbox for Python code to read and process.
- **Online code editing**: Built-in Python editor for writing data processing logic directly, executed safely in the sandbox for flexible data cleaning and analysis.

---

## More Resources

**1. GitHub Open Source**

- AI Data Platform: https://github.com/kweaver-ai/adp/releases/tag/v0.3.0
- Decision Agent: https://github.com/kweaver-ai/decision-agent/releases/tag/v0.3.0
- AI Store: https://github.com/kweaver-ai/ai-store/releases/tag/v.0.3.0

**2. Technical Documentation**

- AI Data Platform: https://github.com/kweaver-ai/adp/blob/main/README.md
- Decision Agent: https://github.com/kweaver-ai/decision-agent/blob/main/README.md
- AI Store: https://github.com/kweaver-ai/ai-store/blob/main/README.md
