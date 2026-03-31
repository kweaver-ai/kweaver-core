# Effect Evaluation Task

## Overview

### Feature Introduction

In Benchmark, a task is used to combine **evaluation rules** with **evaluation target types** and run the evaluation in practice.

- Evaluation rules: define the datasets and test indicators used in evaluation.
- Evaluation targets: act as the objects being tested and participate in evaluation under the same rule.

### Application Scenarios

- Compare the response accuracy and generation speed of different large models under the same prompt.
- Verify the task completion rate and stability of an agent in a specific business scenario.
- Evaluate API call success rates and data processing efficiency of custom applications or externally integrated targets.

## Prerequisites

Before creating an effect evaluation task, make sure the following preparation work is complete:

1. You are logged in to ADP and have access to the **Intelligent Evaluation** module. If not, contact the platform administrator.
2. The **evaluation targets** have been configured and passed validation. Evaluation targets may be large models, agents, or external algorithms integrated by API.
3. The **evaluation rules** have already been configured. These rules include available evaluation datasets and test indicators and can be referenced by effect evaluation tasks.

## Steps

### Create a New Evaluation Task

1. Log in to ADP and choose **Effect Evaluation** under **Intelligent Evaluation** from the left navigation bar.
2. In the right panel, switch to the **Effect Evaluation Task** tab.
3. Click **New** to open the task creation page.
4. Fill in the following information. Fields marked with `*` are required:
   - **Task Name***: the name used to identify the current effect evaluation task
   - **Evaluation Rule***: choose an available evaluation rule that defines the dataset, metric dimensions, and scoring criteria for this evaluation
   - **Evaluation Target Type***: select the target type participating in this evaluation. Available options include **Agent**, **Large Model**, and **External Algorithm via API**
   - **Select Evaluation Target**: specify the concrete targets participating in this evaluation. Multiple targets can be selected for comparison
   - **Upload Adapter File**: adapts the input and output formats among the evaluation dataset, evaluation target, and indicators, so data can be transmitted and parsed correctly
   - **Description**: add notes about the purpose and scenario of the task
   - **Color**: choose a display color for the task for easier visual distinction
5. After configuration, choose one of the following actions:
   - **Save**: save the task configuration without running it immediately
   - **Run**: save the configuration and submit the task to run immediately or wait in the queue

## Task Management Operations

### Edit a Task

- Purpose: modify the configuration of an existing task
- Operation: choose **Edit** in the actions column of the target task, then adjust the settings and save
- Note: if the task is already `Running`, stop it before editing

### Copy a Task

- Purpose: quickly create a new task with the same configuration as an existing one
- Operation: choose **Copy** in the actions column. The system creates a new task from the original configuration, and you can modify it before saving

### View Run Details

- Purpose: inspect the execution results and logs of a task
- Operation: choose **View Run Details** in the actions column to see the run status, evaluation results, and detailed logs

### Delete a Task

- Purpose: permanently remove a task that is no longer needed
- Operation: choose **Delete** in the actions column and confirm as prompted
- Note: only tasks in the `Completed` or `Stopped` states can be deleted. A `Running` task must be stopped first
