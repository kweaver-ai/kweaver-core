# KWeaver Core 0.3.1 Release Notes

This release focuses on quality optimization and user experience improvements, building on the core capabilities of 0.3.0. It addresses issues found through user feedback and production environments, with emphasis on stability fixes in the BKN engine, VEGA data virtualization, and Execution Factory modules, along with Decision Agent frontend experience improvements.

---

## Highlights

1. **Enhanced Data Management Safety**: The BKN engine introduces an object class deletion protection mechanism, using a `force_delete` parameter to prevent accidental deletion that could damage the business knowledge network structure.
2. **Comprehensive Stability Improvements**: Overhauled exception handling in VEGA data virtualization, Dataflow, and Execution Factory core modules, significantly improving stability and fault tolerance in complex business scenarios.
3. **Decision Agent Interaction Improvements**: Optimized frontend experience with collapsible temporary area, improved API debugging, and unified localization throughout the UI.

---

## 1. BKN Engine

**1. Safe Object Class Deletion Control**

Added a safety control mechanism for object class deletion to prevent accidental damage to the business knowledge network structure:

- **Reference check**: The system automatically checks an object class's references, including relationship class and action class references. By default, only unreferenced object classes can be deleted, preventing accidental operations.
- **force_delete parameter**: Supports a `force_delete` query parameter (default: `false`). Setting `force_delete=true` via the API forces deletion of referenced object classes, for special data cleanup scenarios.
- **Data integrity guarantee**: The deletion control mechanism protects the structural integrity of the business knowledge network, reducing the risk of system failures caused by accidental deletions.

---

## 2. VEGA Data Virtualization

**1. Enhanced Anomalous Data Handling**

Optimized exception handling in VEGA data views for anomalous data scenarios in production:

- **Improved exception catching**: Fixed an issue where certain anomalous data was not caught in data views in the production environment, preventing service interruptions.
- **Improved fault tolerance**: Enhanced data view handling of edge cases and anomalous inputs, ensuring stable system operation in complex data scenarios.

---

## 3. Execution Factory

**1. Composite Operator Version Management Fix**

Fixed composite operator version management issues to improve execution reliability:

- **Version reference consistency**: Fixed an issue where importing a composite operator regenerated its version, causing other operators referencing that version to fail execution.
- **Dependency management**: Optimized dependency handling for composite operators to prevent execution failures caused by version changes.

**2. Toolbox Access Error Handling**

Improved error handling for toolbox access:

- **HTTP status code normalization**: Fixed an issue where accessing a non-existent toolbox returned a 404 status code; replaced with a more appropriate error response to improve API call standards.
- **Error message improvements**: Enhanced error messages to help developers quickly identify root causes.

---

## 4. Dataflow

**1. Operator Loop Execution Stability Fix**

Fixed errors occurring when operators contain loops, improving data processing pipeline reliability:

- **Loop execution support**: Fixed errors that occurred when operators contained loop logic, ensuring data processing workflows with loops run correctly.
- **Pipeline stability**: Enhanced Dataflow's handling of complex control flows, supporting more flexible data processing logic.
- **Execution reliability**: Optimized the operator execution engine for stable operation in complex scenarios including loops.

---

## 5. Decision Agent

**1. Conversation Interface Temporary Area Optimization**

Optimized the temporary area in the Agent conversation interface:

- **Show/hide control**: Added support for toggling the temporary area's visibility, allowing users to flexibly manage screen layout.
- **Display bug fix**: Fixed a display anomaly bug in the Agent conversation interface temporary area.

**2. Localization Improvements**

- **Unified terminology**: All instances of "Decision Agent" in the UI are now consistently displayed as "决策智能体" (Decision Agent in Chinese), improving terminology consistency and reducing user confusion.

**3. API Debugging Enhancements**

Improved the Agent API debugging functionality:

- **Business domain request header**: Fixed a missing `x-business-domain` request header when debugging Agent API endpoints, ensuring debug requests match actual business requests.
- **Improved debug experience**: Enhanced debugger configuration for more complete debugging capabilities.

**4. Code Quality Improvements**

- **Component cleanup**: Removed unused `KNSpaceTree`, `DocTree`, and `ContentDataTree` components, reducing code redundancy and maintenance overhead.

---

## More Resources

**1. GitHub Open Source**

- AI Data Platform: https://github.com/kweaver-ai/adp/tree/release/0.3.1
- Decision Agent: https://github.com/kweaver-ai/decision-agent/tree/release/0.3.1

**2. Technical Documentation**

- AI Data Platform: https://github.com/kweaver-ai/adp/blob/main/README.md
- Decision Agent: https://github.com/kweaver-ai/decision-agent/blob/main/README.md
