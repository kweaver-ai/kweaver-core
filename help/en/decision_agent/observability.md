# Decision Agent Observability (Trace Analysis)

## Feature Overview

- **Definition**: the agent observability module, also called **Trace Analysis**, is an execution monitoring and log analysis platform for application administrators. It is used to track the execution status, performance, and problem details of published agents.
- **Purpose**: through multi-level drill-down analysis, administrators can fully understand agent runtime conditions, identify anomalies, optimize performance, and collect data for later version iterations.
- **Target users**: application administrators
- **Location**: part of the agent governance and operations layer of the DIP platform, providing runtime insight upward and connecting to the execution engine and log system downward

## Typical Scenario

### Scenario: An Application Administrator Reviews the Runtime Behavior of Published Agents and Locates Execution Exceptions

- **Trigger**: business users report abnormal handling results from the agent, or the administrator performs routine health inspections
- **Cause**: the agent may experience performance degradation, invocation failures, or logic errors during execution
- **Goal**: use the observability module to drill down from overview to details, quickly locate problematic sessions, analyze the execution chain, identify the root cause, and obtain optimization suggestions

## Key Terms

| Term | Meaning |
| --- | --- |
| Decision Agent | An intelligent agent component with decision-making capability. It can execute specific business tasks based on predefined rules or algorithms and is the monitored object in observability |
| Runtime Metrics | Quantitative data used to measure the operating state of Decision Agent, such as CPU usage, memory usage, task success rate, and response latency |
| Runtime Logs | Textual data that records key operations, events, and error information during Decision Agent execution, including timestamps, operations, and status results |
| Exception Events | Events that deviate from normal operation, such as task failures, resource usage above threshold, or API timeouts |
| Monitoring Dashboard | The core page of the Decision Agent observability feature. It shows runtime metrics and exception events through charts and lists |

## Functional Structure and Operation Flow

### 1. Open the Observability Page

- **Entry**: Log in to ADP and open **Development > Decision Agent > New button / click card > Trace Analysis tab**
- **Page load behavior**: after loading, the page shows aggregated runtime data and session lists for all published Decision Agent instances over the last week by default

### 2. Browse Homepage Monitoring Data

The homepage gives you a global view of all agent runtime states.

#### 2.1 View the Performance Overview

- **Displayed data**: the top of the page shows core performance overview data for all agents within the selected time range
- **Metrics**:
  - Total requests
  - Total sessions
  - Average session rounds
  - Task success rate
  - Average execution duration
  - Average first-token response time
  - Tool call success rate

#### 2.2 Conversation Log List

- **Displayed data**: the list shows all session records generated within the selected time range, where each record represents a complete interaction between the user and the agent
- **Quick overview**: session ID, start time, end time, total duration, and status
- **Core action**: click **View Details** on any session to enter the detail page for deep analysis

#### 2.3 Time Filter

- **Description**: use the time filter at the top of the page to adjust the time range for analysis
- **Usage**: select a preset range such as `Today`, `Last 7 Days`, or `Last 30 Days`, or choose a custom date range. All charts and lists refresh in real time

#### 2.4 Get Optimization Suggestions with AI Analysis

- **Description**: this feature uses AI to analyze recent runtime data and provide potential optimization directions and concrete suggestions
- **Workflow**:
  1. The page initially shows `No optimization suggestions yet` and an **AI Analysis** button
  2. After you click the button, the system analyzes recent performance data and error logs
  3. When analysis is complete, the panel displays structured suggestions, such as prompt optimization, skill configuration adjustments, or resource allocation improvements

### 3. Analyze the Details of a Single Session

When you need root-cause analysis for a particular problematic session, drill down into the session detail page.

#### 3.1 How to Enter

- On the homepage in the **Conversation Log List**, click **View Details** next to the target session

#### 3.2 View Session Details

- **Description**: this page provides the full execution trace and detailed data of a selected session for troubleshooting and performance analysis
- **Core information**:
  1. **Session overview**: session ID, time range, total duration, and overall status
  2. **Execution step pipeline**: all execution steps in time order, such as intent recognition, tool invocation, and reply generation, together with step status and duration
  3. **Intelligent analysis panel**: an analysis view for the current session and its execution steps

### 4. View Execution Step Details

This page provides atomic-level transparency and is the key to locating performance bottlenecks or errors.

#### 4.1 How to Enter

- On the **Single Session Detail** page, click **View Details** on any item in the **Execution Step Pipeline**

#### 4.2 View Step Details

- **Description**: this page shows the detailed runtime information of a specific step, including precise performance data and complete input and output information
- **Core information**:
  1. **Basic step information**: step name, execution status, exact start and end timestamps, and token usage
  2. **Execution timeline**: visualize the time distribution of internal sub-stages, such as network requests and model inference
  3. **Raw input and output**: full raw input data and the output result of the step
  4. **Error details**: if the step fails, the page provides error code, error description, and a link to related logs

## Data Filtering and Drill-Down Logic

### 1. Filter Data by Time

All monitoring data is counted and displayed based on the selected time range. On the homepage, you can conveniently switch among `Today`, `Last 7 Days`, `Last 30 Days`, or any custom range. After the time range changes, all charts and session lists update in real time.

### 2. Drill Down Step by Step to Locate Problems

The platform provides a two-level drill-down path to help you move from abnormal symptoms to root cause quickly.

#### Level 1: From the Global List to a Problem Session

- When you notice abnormal global metrics, such as a drop in success rate, use the **Conversation Log List** to filter sessions by time or status and find potentially problematic records
- Click **View Details** for the target session to inspect its full execution process

#### Level 2: From the Session to a Problem Step

- In the session detail page, inspect the status and duration of each execution step
- If a step fails or takes abnormally long, click **View Details** for that step to inspect the most detailed logs, performance breakdown, and input/output data

## FAQ

### Q1: Why does every chart show `No Data` after I enter the observability page?

- A1: Check the following:
  1. Confirm that the target Decision Agent has been published and is running normally
  2. Adjust the time range. If there were no calls in the default range, choose a longer range such as `Last 7 Days`
  3. Confirm that the account has permission to view runtime metrics

### Q2: Why does `Task Success Rate` show `0%` even though I know some tasks succeeded?

- A2: Check the following:
  1. Confirm that the selected time range covers the time when those successful tasks ran
  2. Search for tasks from that time range in the **Conversation Log List** and confirm the filter conditions

### Q3: Can I export monitoring data for offline analysis?

- A3: Yes:
  1. Click **Export** above the **Conversation Log List** on the homepage
  2. The system exports session records as a CSV file based on the current filters

### Q4: Can regular users view and handle exception events?

- A4: The rules are:
  1. Viewing: all users can enter the exception event list and inspect handled events through **View Record**
  2. Handling: only application administrators or users with designated operations permissions can accept, edit, and resolve exception events

### Q5: Does `Average Task Response Latency` include failed tasks?

- A5: No:
  1. This metric only counts the average response latency of successful tasks
  2. The duration of failed tasks can be viewed in the corresponding task's step detail page together with the error information
