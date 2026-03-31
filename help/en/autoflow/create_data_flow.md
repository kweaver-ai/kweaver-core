# Dataflow Documentation 

## Overview

Dataflow is a data processing process that automatically **extracts**, **transforms**, and finally **loads** data from a **source system** to a **target system** through a series of **transformation nodes**. In traditional business scenarios, manual data querying, analysis, storage, and transformation are time-consuming and error-prone. With Dataflow automation, these processes can be configured to run automatically. Combined with model and agent capabilities, Dataflow can also become more intelligent and greatly improve efficiency.

### Key Mechanisms

- **Data-driven**: the essence of a Dataflow is automation triggered by data movement, not just scheduled execution or manual commands
- **ETL logic**: follows the standard paradigm of Extraction, Transformation, and Loading
- **Write-oriented**: unlike query or display scenarios, the core value of a Dataflow lies in **writing** processed data into a target storage system

### Concepts Related to Dataflow

1. **Data Source**  
   The starting point of the data pipeline. It provides the raw data and determines the type, quality, and structure of the data used for later cleaning, transformation, and analysis.

2. **Execution Condition**  
   Also called a `logical action`. The system runs subsequent actions only when a predefined condition is met. It can be understood as a branch inside the Dataflow.
   - Branch execution order: branches are matched from left to right. Only when the current branch condition is met does the system execute the branch flow. After all actions inside the branch are completed, the actions outside the branch continue.
   - Example: `approval passed -> file upload` can be branch 1, and `approval failed -> delete file` can be branch 2.

3. **Execution Action**  
   A specific operation event automatically executed after the trigger occurs and the execution condition is satisfied.
   - Example: when the `approval passed` condition is met, `upload and publish the document in the knowledge base` is the execution action.

4. **Trigger**  
   A trigger is not a data processing step. It is a monitoring and decision mechanism that continuously listens for one or more predefined conditions or events and automatically starts the downstream flow as soon as the condition is satisfied.

## Create a Dataflow

Data flows can be created in two ways: **from scratch** or **from a template**.

### Create a Dataflow from Scratch

#### Basic Process

Choose data source -> Set execution actions -> Set trigger

#### Detailed Steps

1. **Open the creation page**  
   Go to **Autoflow > Dataflow > Pipeline**, click **New**, choose **Create from Scratch**, and enter the creation page.

2. **Choose the data source**  
   Select the data to be processed. The following source types are supported:

| Data Source Type | Description |
| --- | --- |
| Unstructured Data | Data without a fixed structure, such as documents, images, and audio |
| Structured Data | Data with a fixed format and structure, such as user organization information |
| Data View | A result set generated dynamically by a predefined query over one or more underlying real tables or views |
| Operator Data Source | Third-party data obtained through RESTful APIs. Operators must be configured in advance |

3. **Set execution actions**  
   Execution actions are used to process or pass data. One or more actions can be configured within the same flow.
   - Supported capabilities: adding action nodes, configuring branches, and configuring loops
   - Branches: two branches are added by default. You can configure `AND` and `OR` logic relationships. If no branch condition is set, all branches run.
   - Loops: nodes inside the loop body can be configured, and loop output can be used in later actions.

4. **Detailed settings, using `Execution Action - Large Model` as an example**  
   This section is divided into **Basic Settings** and **Advanced Settings**:
   - **Basic Settings**: configure necessary information such as large model selection and input content
   - **Advanced Settings**: customize node runtime strategy, including timeout, retry count, and retry interval
   - **Typical scenarios**:
     - Large-data enterprise scenarios: when default timeout values cause frequent failures, extend the timeout and set a reasonable retry strategy
     - Small-data enterprise scenarios: shorten the timeout to obtain results faster
   - **Rules**:
     - Value ranges: timeout `1-86400` seconds, retries `1-10`, retry interval `1-60` seconds
     - Existing flows keep their old timeout and retry logic unless they are updated
     - For fixed-interval retries, the timeout should be greater than or equal to `max retries × retry interval`
     - Default node behavior:
       - Review nodes and graph write nodes: frontend does not set a timeout, backend defaults to 30 minutes
       - Asynchronous Python execution nodes: support async and sync, with a default timeout of 24 hours
       - Other synchronous nodes: default timeout is 30 minutes

5. **Set the trigger**  
   The trigger is the start event of the Dataflow. Only one trigger can be configured for a single flow, and it cannot be deleted.
   - Special note: if the data source is a **data view**, only **scheduled trigger** and **manual trigger** are supported.

### Create a Dataflow from a Template

#### Prerequisites

1. The relevant BKN has already been built
2. The required operators have already been built

#### Basic Process

Choose BKN -> Set execution actions -> Set trigger

#### Example Scenario: Automatically Update BKN When a New User Is Added

The goal of this Dataflow is to automatically extract user-related information and update it to BKN whenever a new user is added.

1. **Open the template selection page**  
   Go to **Autoflow > Dataflow > Pipeline**, click **New**, choose **Create from Template**, and enter the creation page.

2. **Configure BKN**  
   Click **Update BKN**, choose the target graph, select the target entities, and configure the document library path used to store files.

3. **Import operators**  
   Choose the configured operators and click **Confirm**.

4. **Complete Dataflow creation**  
   After configuration, the system shows `Dataflow created successfully`.

5. **Verify the running result**  
   - Add a user in **Information Security Management > Users > User Management**
   - Return to the Dataflow page and click **Run Statistics** to verify whether the flow was triggered correctly
   - Check the execution action and verify whether the expected content is output by the operator

6. **Data flow application**  
   After the flow is created, it runs automatically and updates BKN whenever a new user is added. When users ask the super assistant questions, the system can use the latest user information to generate answers.

## Key and Difficult Steps Explained

### Create from Scratch: Choose a Data Source (Data View)

#### Key Configuration Notes

1. **Data retrieval method**

| Retrieval Method | Definition | Typical Scenario |
| --- | --- | --- |
| Full retrieval | Extract all data in the view once within the user-defined time range | Suitable for processing historical data or full datasets in batch |
| Incremental retrieval | Retrieve only newly added or updated data based on an incremental field and filter rules | Suitable for real-time or near-real-time synchronization scenarios |

2. **Incremental field settings**
   - Only numeric or time-based fields can be used as incremental fields
   - The initial value can be defined according to business needs. For example, `0` means counting starts from an initial or zero-value state

3. **Special rule**  
   If the query language of the data view is **DSL**, the retrieval method is fixed to **full retrieval**, and the related UI option is hidden.

4. **Filter rule configuration**  
   Custom SQL `WHERE` clauses can be used to filter data precisely.  
   Example: filter partition data for the previous day where area code is `275`:
   `"partition" = to_char( (CURRENT_DATE - INTERVAL '1 day') ,'yyyymmdd' ) and "region" = '275'`

5. **Batch size setting**  
   - Definition: the maximum number of records processed in a single batch
   - Range: `1000-10000`

6. **Data preview**  
   After the data source is configured, click **Data Preview** to inspect data format, field values, and data volume in real time.

### Create from Scratch: Choose a Data Source (Operator Data Source)

#### Key Configuration Notes

1. **Operator prerequisites**
   - The target operator must already be configured in the system
   - The operator must include a `data source identifier field`
   - Only **synchronous operators** can be used as data sources

2. **Variable usage limits**
   Data source nodes support only **global variables**, such as user tokens or global system parameters. They do not support local variables from other nodes inside the flow.

### Create from Scratch: Choose Execution Actions

#### Write to BKN

- **Description**: write information from built-in entities such as files, users, and organizations, or from custom entities, into a specified BKN
- **Steps**:
  1. Choose the target BKN in the node configuration page
  2. Select the entity class to write to
  3. Configure the entity property values by referencing outputs from previous nodes or by entering static values manually

#### Execute Python Code

- **Description**: implement complex logic such as data cleaning, format transformation, and custom algorithm calculation by writing Python scripts
- **Steps**:
  1. **Input variable configuration**
     - Supported types: `string`, `int`, `array`, `object`
     - Values can come from previous node outputs or be entered manually
  2. **Output variable configuration**
     - The output type must match the return type of the script
  3. **Code writing**
     - Return results with `return`, not `print`
     - Example:
     ```Plain Text
     def process(input_str):
         return input_str.upper()
     result = process({{input_var}})
     return result
     ```
- **Dependency note**: built-in libraries such as `pandas`, `numpy`, and `json` are available. For third-party libraries, refer to **Use Python Nodes in Autoflow**.

#### JSON Processing

Three JSON operation types are supported:

| Operation Type | Purpose | Workflow |
| --- | --- | --- |
| Get value by key | Extract the value of a specified key from a complex JSON object | Provide JSON input, specify the key, and the system outputs the matched value or `null` |
| Set value by key | Modify the value of an existing key or add a new key-value pair | Provide original JSON, the target key, and the new value, and the system outputs a new JSON object |
| Template transformation | Fill JSON data into a predefined text template | Provide JSON input, write a template with `{{Key}}`, and the system renders the final text |

#### Invoke AI Capabilities

The Dataflow can call large models, small models such as Embedding and Rerank, and agents.

1. **Large model**
   - **Attachments** are optional and can improve model understanding:

| Attachment Type | Configuration Rule |
| --- | --- |
| Image | Selected automatically by default. Any number of images can be added |
| Video | Supported only by vision models, and only one video can be selected |
| File | Can be selected directly as a file or file variable, or provided as a URL / URL variable |

   - **Node outputs**:
     - `answer`: plain text result
     - `json`: JSON object automatically parsed from the text result if possible

2. **Small model - Embedding**
   - **Purpose**: convert unstructured information such as text, images, and audio into vectors
   - **Typical uses**: retrieval, clustering, classification, and deduplication
   - **Input / output**:
     - Input: a `text array`
     - Output: `data`, an array of embeddings corresponding to the inputs

3. **Small model - Rerank**
   - **Purpose**: rerank candidate results selected in a previous stage according to relevance
   - **Typical uses**: RAG systems and search engines
   - **Input / output**:
     - Input: `query` and `documents`
     - Output: `results` with detailed relevance scores and `documents` sorted in descending order by score

## Write to an Index Library

- **Prerequisite**: the target index library must already exist in the system
- **Steps**:
  1. Select the target index library from the node configuration page
  2. Configure the content to index either by referencing a JSON variable from a previous node or by entering a JSON string manually
  3. Click **Validate** to confirm that the JSON structure matches the index schema

## SQL Write

- **Steps**:
  1. **Source data settings**
     - Supports `array` for batch data and `json` for a single record
     - Data can be entered manually or referenced from previous node outputs
  2. **Batch write size**: set the number of records inserted in one SQL batch
  3. **Database connection configuration**:
     - Supported database types include `mariadb`, `mysql`, `dameng`, `postgresql`, `sqlserver`, `oracle`, and `kingbase`
     - Choose a configured database connection from the list
  4. **Target table configuration**:
     - Create a new table automatically from source fields
     - Or select an existing table and keep its schema unchanged
  5. **Additional table information**:
     - Table name
     - Table description
     - Field mapping

## Write to a Document Library

Two document operations are supported: **Create File** and **Update File**.

| Operation Type | Supported File Formats | Description |
| --- | --- | --- |
| Create File | `docx`, `xlsx`, `pdf`, `markdown` | Create a new file in the target folder and write content into it |
| Update File | `docx`, `xlsx`, `markdown` | Update the content of an existing file and keep historical versions |

## Content Processing

Four types of content processing are provided. The source document currently retains only one row in this section:

| Operation Type | Description | Configuration and Output |
| --- | --- | --- |
| Extract all text from a file | Extract plain text in batch from a file of a supported format, ignoring formatting information | Configuration: select the target file, optionally choose a file version, and extract text from supported formats such as `.docx`, `.xlsx`, `.xls`, `.odt`, `.wps`, and related variants |
