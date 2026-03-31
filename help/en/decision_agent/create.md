# Create Decision Agent

This document describes the complete process for creating, configuring, and publishing a Decision Agent in the DIP system, including prerequisites, key steps, notes, and frequently asked questions. It is intended for users of the development platform.

## Prerequisites

- The user must be logged in to the system and have entered the development platform.
- The user must configure the required product type for the Decision Agent, such as ChatBI, DIP, or AnyShare.
- The user must select and configure an appropriate data source, such as BKN or a document library.

## Key Terms

No specific key terms are defined in the original source document. This section is reserved for future additions.

## Create Decision Agent

### Steps

1. Log in to ADP and click **Development > Decision Agent > New**.

### Decision Agent Configuration Page

After clicking **New**, the system enters the Decision Agent creation and configuration page. This page contains several configuration modules that should be completed one by one.

#### Basic Information

- **Name**: Enter the name of the Decision Agent. The maximum length is 50 characters.
- **Summary**: Briefly describe the purpose and function of this agent, or generate it with AI assistance. The maximum length is 200 characters.
- **Avatar**: Upload or select an identifying avatar for the agent, or generate one with AI assistance.
- **Product**: Select the product associated with this agent, such as `DIP`.

#### Role Instructions

Two configuration modes are supported:

- **Natural language mode**: Write role instructions in prompt format, or generate them with AI assistance.
- **Dolphin (expert mode)**: Write Dolphin code to implement more complex tasks and decision flows.

Role instructions should include:

- **Role definition**: Clearly define the identity and responsibility of the agent.
- **Target tasks**: Describe the main tasks the agent is expected to complete.
- **Execution constraints**: Define permission boundaries, parameters, and behavioral limits.

#### Knowledge Sources

Configure the knowledge sources used by the agent. Supported options include:

- **BKN**: Connect to the enterprise's internal BKN.
- **Knowledge entries**: Add specific knowledge entries to enrich the agent's knowledge base.

#### Skill Configuration

Select one or more skill modules from the skill library, such as:

- Data analysis
- Contract matching
- Retrieval augmentation
- Task planning

#### Decision Models

Choose the AI model used when the agent performs tasks, such as:

- Large language models for general conversation and generation
- Reasoning models for structured logic and decision tasks

#### Advanced Configuration (Optional)

- **Long-term memory**: Configure whether to retain conversation history.
- **Default opening**: Set the greeting or initial prompt shown when the agent starts. AI-assisted generation is supported.
- **Related questions**: Configure a list of related questions the agent can answer.
- **Task planning**: Set the strategy for task decomposition and planning.
- **Preset questions**: Configure a list of common questions so users can choose quickly. AI-assisted generation is supported.

#### Save and Publish

After completing the configuration above:

- Click **Save** to store the agent as a draft.
- Click **Publish** to deploy the agent to the platform for user access.

##### Notes

1. Long-term memory: when this feature is enabled, the agent can record and reuse historical user interactions to improve decision accuracy, such as personalized recommendations and context continuity.
2. Configuration validation: before publishing, it is recommended to review the input settings, knowledge sources, and skill modules one by one to avoid runtime issues caused by missing or incorrect configuration.

## Notes and Boundaries

### Permission Management

- Only users with application administrator permission or designated development-role permission can create, edit, publish, or delete a Decision Agent.
- Regular users can only use published agents, while also having edit permission for Decision Agent instances they created themselves.
- Users can only modify unpublished versions they created. Published agents must be updated through version management.

### Publishing Limits

- Published Decision Agent instances can still be edited and republished.
- After republishing, the new version directly overwrites the old version, and the online environment is updated immediately.
- All historical versions are recorded on the **Configuration Information** page of the Decision Agent and can be reviewed at any time.
- If rollback is required, you can select a historical version on the **Configuration Information** page and publish it again. That version then becomes the latest online version.

### Data Privacy

- When enabling the **long-term memory** module, follow enterprise data privacy policies and relevant regulations, such as GDPR and personal information protection laws.
- Audit user data stored in long-term memory regularly to prevent leakage or non-compliant retention of sensitive information.
- Storage and processing of user data should follow enterprise security standards, and sensitive data should be desensitized.

### Performance and Stability

- A single agent can only process a limited number of tasks concurrently. Exceeding the limit may lead to response delays.
- The processing duration of complex tasks depends on factors such as model performance and knowledge base size.

## FAQ

Q1: **What should I do if I cannot find the right skill when creating a Decision Agent?**  
A1: Check the following:

- Confirm that the required skill has already been created in the **Skill Module Library**. If not, create it first on the **Skill Management** page.
- Check whether the skill has been enabled. If it is disabled, turn on its availability on the **Skill Management** page.
- Verify that the skill is compatible with the current Decision Agent product type, and confirm that the skill's applicable product scope includes the target type.

Q2: **How do I modify a published Decision Agent?**  
A2: Follow these steps:

- Go to the **Decision Agent** list page and find the target agent.
- Open the agent for editing.
- After updating the configuration, click **Publish**. The system updates the agent to the latest version automatically.
