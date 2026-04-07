# What Is Decision Agent

Decision Agent is built on BKN and integrates Dataflow processing capabilities. Its core responsibility is to create, configure, and manage agents throughout their full lifecycle. By combining multiple knowledge sources, extensible skill modules, intelligent models, and the Dolphin-Lang programming language, it can efficiently handle complex decision-making tasks and provide intelligent support for business scenarios.

## Core Architecture

The functionality of Decision Agent depends on four core modules. Each module has a clear role and works together to form the complete capability system, as shown below:

| Core Module | Positioning | Details |
| --- | --- | --- |
| Knowledge Sources | The "blood" of the agent | Provides the foundational information required for agent execution, including but not limited to:<br>- Document libraries (structured / unstructured documents)<br>- BKN (domain knowledge relationships)<br>- Indicator systems (core business metrics)<br>- Knowledge entries (fragmented and precise knowledge units) |
| Skills | The "capability extender" of the agent | Extends the agent beyond its basic functions to support complex task processing, including:<br>- Skill Agents (sub-agents for specialized tasks)<br>- MCP (Model Context Protocol for consistent model interaction)<br>- Tools (external functional components such as data query tools)<br>- Operators (basic computing / processing units for data operations) |
| Models | The "brain and thinking center" of the agent | Responsible for intelligent decision logic, with the following characteristics:<br>- Multi-model compatibility: supports both large models (such as general conversation models) and smaller specialized reasoning models<br>- Standardized interfaces: provides a unified invocation method to reduce integration and switching costs |
| Dolphin-Lang | The "nervous system" of the agent | A dedicated flow configuration language used to:<br>- Orchestrate Dataflow (define data paths and processing order)<br>- Schedule task handling (assign tasks to the right modules and monitor execution)<br>- Implement decision logic (turn business rules into executable flows) |

## Application Scenarios

With flexible configuration and strong decision support capabilities, Decision Agent has been widely applied in multiple domains. Typical scenarios include:

1. **Decision support systems**: Provide data-driven recommendations for business leaders, such as market trend analysis and risk assessment, reducing subjectivity in decisions.
2. **Business process automation**: Replace manual work in repetitive and rule-based decision steps, such as order review and compliance checks, to improve operational efficiency.
3. **Intelligent customer service**: Combine business knowledge with conversation models to answer user questions accurately and anticipate needs, improving the customer service experience.
