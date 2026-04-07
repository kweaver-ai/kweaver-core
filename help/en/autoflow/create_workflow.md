# How to Create a Workflow

## What Is Workflow Automation?

Workflow automation can automatically complete repetitive and mechanical tasks and processes, reduce manual intervention, improve content circulation and processing efficiency, help teams save time, and reduce operational errors.

## Common Workflow Use Cases

In traditional business scenarios, manually handling data synchronization, document copying, and document moving is error-prone and inefficient. Workflow automation can automatically execute tasks such as:

- Triggering an approval flow automatically after a file is uploaded to a specific folder
- Adding tags automatically when a new file is created
- Saving attachments automatically to a specified path
- Other custom automation scenarios, with support for more than 30 action nodes and more than 10 workflow templates, without requiring coding skills

## Core Workflow Concepts

The execution logic of a workflow can be summarized as: **when a defined trigger action occurs, the system checks the execution condition and then automatically runs the execution action**.

| Concept | Definition | Common Types / Examples |
| --- | --- | --- |
| Trigger Action | The start condition of the workflow, determining whether the flow begins | Event trigger: folder upload / delete / modify operations; Manual trigger: manually start a flow; Form trigger: start after a form is submitted; Scheduled trigger: start at a fixed interval |
| Execution Condition | The branching rule in the flow. Subsequent actions run only when the condition is met | Example: if the review result is approved, publish the file; if not approved, delete the file |
| Execution Action | The concrete operation automatically executed after the condition is met | Example: send a notification, synchronize data, modify file properties, archive files |

# How Workflows Run

You only need two steps to start workflow automation: **set the trigger** and **set the execution actions**.

## Step 1: Set the Trigger

- **Purpose**: the starting point of the workflow. A workflow supports only one trigger, and it cannot be deleted.
- **Example**: in the scenario `automatically add a tag after a file is uploaded`, the trigger is `file upload`.
- **Supported types**: event trigger, manual trigger, form trigger, and scheduled trigger.

## Step 2: Set the Execution Actions

- **Purpose**: the concrete actions that run after the trigger fires. One or more actions can be configured.
- **Example**: in the flow `notify related staff after a file is updated`, `send notification to designated users` is the execution action.

# Create and Manage Workflows

## Common Workflow Scenarios

The following scenarios can be used directly as configuration references or adjusted according to actual needs.

### Scenario 1: Knowledge Base File Publishing Approval

- **Requirement**: before a file is published to the knowledge base, it must be reviewed and approved by a manager
- **Core configuration**:
  - Trigger: `file uploaded to the pending publishing folder in the knowledge base`
  - Execution condition: `review result = approved`
  - Execution action: `publish the file to the public directory of the knowledge base`

### Scenario 2: Automatic Archiving of Expired E-commerce Images

- **Requirement**: seasonal sample images in an e-commerce image library should be taken offline and archived automatically
- **Core configuration**:
  - Trigger: `file creation time exceeds a defined period, such as 90 days`
  - Execution action: `move the file to the archive folder and mark it as archived`

## Create a Workflow

1. Log in to DIP, go to **Autoflow > Workflow > My Flows**, and click **New**.
2. Choose a trigger type, such as `Event Trigger - File Upload`, and configure the trigger condition, such as the target folder path.
3. Optionally add execution conditions, such as an approval result check.
4. Configure execution actions, such as sending notifications or moving files.
5. Enter a workflow name, such as `Knowledge Base Approval Flow`, then click **Save and Enable** to finish.

## Manage Existing Workflows

On the **Autoflow > Workflow** page, you can perform the following operations on created workflows:

- **Edit**: modify the trigger, execution conditions, or execution actions
- **Run**: execute a manually triggered workflow
- **Enable / Disable**: temporarily pause or resume workflow execution
- **Delete**: remove a workflow that is no longer used. Make sure data has been backed up first
- **View Logs**: inspect workflow run records, including trigger time, run duration, execution result, and error information

# Workflow Node Details

## Core Node Type: Review Node

A review node is used in business scenarios such as document collection, circulation, and sharing to ensure content accuracy and data security.

Using a `document circulation` scenario as an example, the configuration steps are:

1. Start document circulation and complete the basic configuration such as name, target location, and scope.
2. Find the **Review** module and click **Settings** to open the review flow configuration page.
3. Choose how to create the review flow:
   - **Option 1: Use an existing review template**  
     Click **Select from Existing Templates**, choose a suitable template, and click **Use This Flow Template**
   - **Option 2: Create a new review flow**  
     Click **+** to add a review stage, then configure:
     - Review stage name, such as `Department Manager Review` or `Director Review`
     - Whether countersigning is allowed
     - Reviewer selection method
     - Review mode, such as parallel review, joint review, or sequential review

### Reviewer Selection Methods

| Selection Method | Suitable Scenario | Description |
| --- | --- | --- |
| Fixed user review | All requests must be handled by fixed reviewers | Select a specific user directly as the reviewer |
| Rule-based review | Reviewers should be matched automatically by department or organization structure | Choose the same department, choose a reviewer from the organization structure, or configure one-level or multi-level superior review |

## FAQ

Q1: **What if no reviewer is found at the configured matching level?**  
A1: Configure the review node to automatically reject or automatically approve so that the process is not interrupted.

Q2: **What if the reviewer and the initiator are the same person?**  
A2: Configure the review node to automatically reject or automatically approve so that the flow is not blocked by self-review.

### Review Modes

| Review Mode | Rule | Suitable Scenario |
| --- | --- | --- |
| Parallel review | When multiple reviewers are involved, approval from any one reviewer is enough | Scenarios that require fast approval and do not need all reviewers to confirm |
| Joint review | All reviewers must approve. A rejection from any reviewer terminates the process | Scenarios that require confirmation by multiple reviewers, such as contract review |
| Sequential review | Review runs level by level, and a rejection at any level terminates the process | Scenarios that require hierarchical approval, such as data access requests |

### CC for Review Results

When configuring a review node, you can add **CC for Review Results** so that non-reviewers, such as department members or related stakeholders, also receive the result.

Recipients can view the result in **Autoflow > Workflow > Review To-do > My To-do**.

## Other Nodes

For the configuration of AI capability nodes, text processing, JSON processing, Python code execution, content processing, variables, and other nodes, refer to the document **How to Create a Dataflow?**
