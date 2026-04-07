# Create an Operator

## Overview

An operator is an independent functional unit in a Dataflow that transforms, calculates, or processes data. By combining and connecting these operators, you can build a complete, automated processing pipeline from raw data to final results.

## Create an Operator

You can create custom operators in two ways: **Function Compute** (writing custom code) and **Operator Orchestration** (visually orchestrating existing operators). Created operators can then be referenced by the Dataflow module.

### Key Terms

| Term | Meaning |
| --- | --- |
| Function Compute | Implement custom logic by writing Python code, with support for code editing and debugging |
| Operator Orchestration | Orchestrate existing operators through a drag-and-drop visual interface to build composite business logic |
| Debugging | Run the operator with test data to verify that the code or logic works correctly |
| Timeout | A configurable time limit that prevents a single task from taking too long and blocking the system |

## Prerequisites

1. You are logged in to the ADP system and have permission to create items in **Operator Management**.
2. If you choose **Function Compute**, you should have basic Python coding skills.

## Scenario 1: Create a Function Compute Operator

### Steps

1. Open the operator list page: go to **Execution Factory > Operator Management > Operator**.
2. Start creation: click **Create Operator** in the upper-right corner and choose **Function Compute** in the dialog.
3. Write the code logic:
   - Open the **Code Editor** tab and write the `handler` function in the editor by following the example format. The function must include an `event` parameter and return the processed result.
   - Add your logic based on the page comments. For example, use `event.get` to read input parameters and define the return structure.
4. Configure metadata:
   - Switch to the **Metadata** tab. Under **Basic Settings**, fill in the operator name (required), operator description (required), and timeout (default `3000ms`, editable).
   - In the **Input Parameters** and **Output Parameters** sections, click **Add Parameter** and fill in the parameter name (required), parameter description (required), type (selected from the dropdown list), and required status (required / optional).
5. Debug and verify:
   - Enter test data in JSON format in the **Debugging** area on the right.
   - Click **Run** and check the result in the **Output** area to verify that the logic works correctly.
6. Save the operator: after confirming that the configuration is correct, click **Save** in the upper-right corner of the page.

### Notes

- After you choose the operator creation method (**Function Compute** or **Operator Orchestration**), it cannot be changed. To use a different method, delete the operator and create a new one.
- The input and output parameters in metadata must strictly match the code logic. Otherwise, parameter mismatches may occur when the agent invokes the operator.
- Debugging data must be valid JSON, and the timeout should be set appropriately to avoid execution interruption caused by a timeout that is too short.

## Scenario 2: Create an Operator Orchestration Operator

1. Open the creation page: go to **Execution Factory > Operator Management > Operator**.
2. Start creation: click **Create Operator** in the upper-right corner and choose **Operator Orchestration** in the dialog.
3. Configure the workflow nodes:
   - Set the **Start Operator** and configure the trigger parameters.
   - Choose the **Execution Actions**, select the required actions, and configure the related parameters.
   - Set the **End Operator** and define the output format.
4. Verify the run: click **Run** to test the composite operator. Use **Details** to check the run status. If the run fails, review the error logs to locate the issue.
5. Use it in calls: when creating a Dataflow, use this operator as an execution node.

## FAQ

### Q1: What should I do if I click **Run** while debugging a Function Compute operator and no output is returned?

A1: Check the following:

1. Make sure the code correctly implements the return logic in the `handler` function and includes a `return` statement.
2. Verify that the test data is valid JSON. You can use an online JSON validator if needed.
3. Confirm that the timeout is not too short. Increase it appropriately and try again.

### Q2: After saving an operator, can I change its creation method (**Function Compute** / **Operator Orchestration**)?

A2: No. The creation method is fixed when the operator is created. After the operator is saved, it cannot be changed. To use a different method, delete the operator and create it again.
