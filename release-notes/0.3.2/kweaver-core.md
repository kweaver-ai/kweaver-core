# KWeaver Core 0.3.2 Release Notes

KWeaver DIP 0.3.2 focuses on enhancing data analysis capabilities and improving product stability, with improvements in three key areas: **intelligent data discovery enhancements, smart data querying experience optimization, and bug fixes**.

---

## 1. Intelligent Data Discovery

**1. Automatic Task Planning for Data Discovery**
Supports automatically decomposing complex data discovery requests into executable steps, forming a complete task planning workflow. A to-do management mechanism tracks task execution status throughout the process, marking completed/pending steps to ensure no critical step is missed.

**2. Knowledge Enhancement for Improved Discovery Accuracy**
Supports knowledge augmentation based on business data. During data discovery, the system combines metadata information, sample data, and data quality probe results from data resources to validate and filter recall results, ultimately outputting a list of data resources that meet business requirements — effectively improving accuracy and reliability.

**3. Department Responsibility-Driven Data Discovery**
Supports data discovery driven by department responsibility definitions. By matching department information and information systems in data tables, the search upgrades from simple "keyword matching" to "department responsibility-driven matching." Results include responsible units, unit responsibilities, and information systems, clearly identifying data ownership and accountability.

**4. Optimized Intent Understanding**
Supports precise identification of the core intent behind user queries, automatically distinguishing between data discovery, follow-up, counter-question, and ambiguous intent scenarios, then applying the corresponding processing logic for efficient and accurate responses.

**5. Custom Search Ranking Strategies**
Supports defining custom search ranking strategies via a rule library. When a user's question involves keywords defined in the rule library, the preset strategy is automatically triggered to prioritize or top-rank specified data tables in results.

**6. Follow-Up Queries Based on Historical Discovery Results**
Supports follow-up queries in continuous conversation. When a user builds on previous discovery results — by adding query conditions or narrowing the scope — the system continues reasoning from existing analysis results without restarting the task, allowing query logic to converge progressively while avoiding redundant computation and improving response efficiency.

**7. Enhanced Data Discovery Under Specified Constraints**
Supports precise identification of constraint conditions in user queries (e.g., "excluding"), strictly enforcing those constraints during data discovery to filter out resources related to the restricted conditions and ensure results match user requirements.

**8. Templated Output for Discovery Results**
Supports defining output templates for data discovery results via prompts. Users can flexibly adjust template content to suit their business needs, making results more standardized and consistent with business usage habits.

**9. Related Question Recommendations**
While answering a user's current question, the system automatically generates related follow-up questions, helping users discover potential data query directions, reducing input cost, and enabling users to comprehensively mine data value.

**10. Support for Data View (Table/View) Resources**
Data discovery now supports data views. Users can locate the view resources needed for their business scenarios, quickly preview relevant data through view details, improving data discovery efficiency and reducing comprehension costs.

**11. Unified Entry Point for Data Discovery and Querying**
Provides a unified intelligent data query entry point. Users can initiate data queries and analysis tasks directly via an on-page dialog, with data discovery enabled by default. Users can also switch between different skills in the search box to access more data analysis capabilities.

---

## 2. Smart Data Querying

**1. Enhanced Data Query Toolbox Experience**
Provides a data query toolbox covering the complete analysis pipeline — from problem analysis and LLM-generated analysis code to sandbox execution and result generation. Key improvements include fixes for sandbox creation failures, knowledge network call exceptions, and Text2SQL-to-sandbox connectivity failures, ensuring stable execution of analysis tasks.

**2. Related Question Recommendations**
Supports related question recommendations alongside answers, guiding users toward deeper analysis and helping them continuously uncover business value hidden in data.

---

## 3. Bug Fixes and Product Experience Improvements

**1. Decision Agent Experience Improvements**
- Supports quick startup of the Decision Agent frontend environment via Docker Compose.
- Completed Decision Agent frontend refactoring; local frontend-backend debugging no longer requires the proton main application.
- Fixed incorrect status filter values on the "My Templates" page in the Decision Agent development UI.
- Fixed an issue where the interruption panel reappeared after refreshing the browser following a mid-execution Agent dialog interrupt.
- Fixed a page crash caused by value rendering errors during Agent conversations.
- Changed the Agent temporary area ID from a dynamic ID to a fixed value `sess-agent-default`, reducing pod resource overuse caused by excessive temporary sandbox instances.

**2. Business Knowledge Network Fix**
- Fixed an issue where Logical Properties could not be added during object class creation, restoring normal logical property configuration in the object class creation workflow.

**3. Data Model - Metric Model Fixes**
- Fixed an issue where previously created groups in the metric model could not be deleted, restoring normal group management functionality.
- Fixed an issue where composite metric creation fields were not inputtable, ensuring composite metrics can be created and used normally.

**4. Data View Fix**
- Fixed a style stacking display issue in the custom view data management page caused by incorrect layer configuration of UI elements, restoring normal page layout.

**5. Execution Factory Fixes**
- Fixed an issue where input parameter display was incorrect on the MCP detail page, ensuring proper display of field names, types, and descriptions.
- Fixed an issue where incorrect parameter passing in MCP debug interface calls led to erroneous debug results, ensuring tools can be debugged and run correctly.

---

## More Resources

**1. GitHub Open Source**

- Intelligent Data Discovery & Querying: https://github.com/kweaver-ai/chat-data/tree/v0.3.2
- Decision Agent: https://github.com/kweaver-ai/decision-agent/tree/release/0.3.5
- AI Data Platform: https://github.com/kweaver-ai/adp/tree/release/0.3.2

**2. Installation Packages and Technical Documentation**

https://anyshare.aishu.cn/link/AR4787DE0ADD3043629C7CB4AC8E24BCFB

Folder name: KWeaver DIP 0.3.2版本
Folder path: AnyShare://产品安装包和补丁库/KWeaver DIP/KWeaver DIP 0.3.2版本

**3. Product Release Materials**

https://anyshare.aishu.cn/link/ARFC4B6FEA00444B12B97EEF421804162A

Folder name: KWeaver DIP 0.3.2
Folder path: AnyShare://产品资料(Product Document)/09.爱数决策智能平台DIP/01-产品发布/KWeaver DIP 0.3.2
