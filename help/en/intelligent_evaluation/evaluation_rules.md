# Evaluation Rules

## Overview

Effect evaluation rules are a set of unified configurations used to standardize the evaluation process. They define the datasets, indicators, and calculation methods used in evaluation, and provide a standardized basis for later task execution and result comparison.

The core capability of effect evaluation rules is the centralized management and storage of **evaluation rule configurations**. The system supports creating, viewing, editing, copying, and deleting rules. Once configured, evaluation rules can be referenced directly by effect evaluation tasks to assess and compare the performance of algorithms or models.

## Prerequisites

1. You are logged in to ADP and have access to the **Intelligent Evaluation** module. If not, contact the platform administrator.
2. You have prepared a standardized sample dataset for effect evaluation, which defines both the algorithm input and the expected output as benchmark truth.
3. Test indicators have already been defined and configured. These indicators are used to quantify algorithm performance on evaluation datasets and serve as the core basis for measuring evaluation effectiveness.

## Steps

### Create a New Evaluation Rule

1. Log in to ADP and choose **Effect Evaluation** under **Intelligent Evaluation** from the left navigation bar.
2. In the **Evaluation Rules** list on the right, click **New** to open the creation dialog. Fill in or choose the following information:
   - **Name**: a concise and clear name that reflects the evaluation goal or scenario
   - **Creation Type**: choose how to create the rule:
     - **Blank Configuration**: configure the entire rule from scratch. The steps below use this method as the example
     - **Template Configuration**: start from an existing rule template and modify it for a similar or repeated scenario
   - **Description**: explain the purpose and scope of the rule. It is recommended to include the target type, business scenario, or version information
   - **Color**: choose an identifying color for easier distinction in lists or on the canvas
3. After filling in the information above, click **Create** to enter the detailed configuration page of the evaluation rule.

#### Create a Task

Tasks are the basic units inside an evaluation rule and are used to separate different evaluation targets. They are usually named by capability type for easier management and reuse.

1. On the canvas, click the **+** button on the right side of the **Evaluation Rule Configuration Information** card to create one or more tasks.
2. In the right-side configuration panel, set the following:
   - **Task Name**: used to identify the task. It is recommended to reflect the evaluated capability directly, such as language understanding, knowledge reasoning, or agent interaction
   - **Description**: explains the task objective, applicable scenario, or any additional notes

> **Note**: One evaluation rule can contain multiple tasks. Different tasks can use different datasets, indicators, and evaluation targets.

#### Configure Evaluation Datasets and Indicators

1. On the canvas, click the **+** button to the right of the task card and choose the **Evaluation Dataset and Indicators** card.
2. Click **Add Evaluation Dataset File**.
3. Complete dataset configuration in the right side panel:
   - Select the evaluation dataset together with its version and file. At least one file must be configured, and multiple files are supported
   - For each file, at least one **Input** and one **Output** must be selected:
     - **Input**: sample data provided to the algorithm for prediction or processing
     - **Output**: the target value the algorithm is expected to produce
4. After dataset configuration is complete, click **Add Indicator** on the canvas:
   - Select one or more indicators from the list. While selecting, you can review each indicator's definition and calculation description

#### Adapter Configuration

Adapter is used to adapt the input and output formats of evaluation datasets and indicators so that data can flow correctly.

- **Scenarios that require configuration**:
  - A single file contains multiple outputs
  - Multiple files are configured, regardless of the number of outputs
- For details, see the **Adapter Usage Guide** document.

#### Ranking Configuration

After evaluation is complete, the ranking page is used to visualize the results. Its core configuration items are **Average Score** and **Per-Task Indicators**.

1. On the canvas, click the **Ranking** card.
2. In the right side panel, configure the following:
   - **Average Score**: automatically calculates the average value of the core indicators across all tasks
   - **Per-Task Indicators**: displays the specific values of selected indicators under each task, such as `Language Understanding Task - Accuracy 92%`

> Note: The ranking updates in real time when datasets or indicators change. No separate save action is required.

#### Save and Publish

- After configuration is complete, click **Save** to save the current rule as **Unpublished**. Unpublished effect evaluation tasks can still be used normally
- After a rule is published, it cannot be edited directly. If you need to modify it, choose **Copy** in the actions column and edit the copy instead

### Edit an Evaluation Rule

To modify an unpublished evaluation rule, choose **Edit** in the actions column of the target rule, enter the configuration page, adjust the relevant settings, and save.

> **Note**: If the rule is already `Published`, you must copy it first before editing.

### Copy an Evaluation Rule

To quickly create a new rule with the same configuration as an existing one, choose **Copy** in the actions column. The system creates a new evaluation rule from the selected configuration, and you can modify and save it afterward.

### Delete an Evaluation Rule

To permanently remove an evaluation rule that is no longer needed, choose **Delete** in the actions column and confirm the operation. The rule is then deleted and cannot be recovered.

> **Note**: If the rule has already been published, you must copy it before editing. Deletion has no state restriction, but deletion is irreversible.
