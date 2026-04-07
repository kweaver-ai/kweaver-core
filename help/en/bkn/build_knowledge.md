# Build BKN

## Overview

### Positioning of BKN

BKN takes ontology as its core methodology and adopts a three-in-one ontology model of **data - logic - action**. By extracting object classes, relationship classes, and action classes and establishing links among them, it ultimately forms BKN that can support intelligent decision-making and automated operations.

### Layered Architecture of BKN

BKN evolves vertically from **general -> industry -> customer-defined**, with each layer serving a different role:

| Layer | Core Purpose | Example |
| --- | --- | --- |
| General BKN | Accumulates cross-industry concepts, relationships, and attributes, providing a standardized knowledge foundation | Concepts: `Person`, `Customer`; Relationships: `Belongs To`, `Associated With` |
| Industry BKN | Inherits the general layer and adds industry-specific concepts and rules to reduce vertical-domain modeling costs | Finance: `Risk Control Indicator`; Retail: `Supply Chain Node` |
| Customer-defined BKN | Allows enterprises to customize business models on demand for accurate scenario mapping | Business processes for a specific product, customer segmentation strategies |

## Core Concepts

BKN is built on **ontology models + resources**. The core elements are defined below:

| Element | Subtype | Definition | Example |
| --- | --- | --- | --- |
| Ontology Model | Concept Group | A logical grouping of concepts. Concepts can belong to multiple groups. | The `User Management` group contains the `Customer` object class and the `Associated Order` relationship class |
|  | Concept - Object Class | Defines the essential properties of a business entity, including data properties and logical properties | - |
|  | Concept - Relationship Class | Describes structural relationships between entities without involving business actions | - |
|  | Concept - Action Class | Defines a concrete business operation together with its conditional constraints | - |
|  | Mapping - Association Rules | Binds concepts to data, logic, and action resources | - |
| Resources | Data Resources | Data views from VEGA virtualization | Customer consumption record views, device status views |
|  | Logic Resources | Resources from VEGA data models and Execution Factory operators and functions | Customer activity scoring model, risk scoring operator |
|  | Action Resources | Tools and MCP interfaces from Execution Factory | SMS push tool, device scheduling MCP interface |

## Build Process

### Prerequisites

1. Complete business scenario analysis and identify the core entities and relationships that need to be modeled.
2. Create the required data views in VEGA virtualization for binding data resources.

### Create a BKN

1. Log in to the system and click **BKN > BKN**.
2. Click **+ New** and configure the following in the dialog:
   1. **ID**: Custom identifier. If left blank, the system generates one automatically.
   2. **Name**: Should reflect the business scenario, such as `Retail Customer Management Knowledge Network`.
   3. **Tags**: Used to classify and quickly search BKN.
   4. **Icon + Color**: Visual identifiers used to distinguish different BKN instances in the system.
   5. **Description**: A short explanation of the network purpose. Optional.
3. Click **Save** to complete creation.

### Define Concepts: Object Class / Relationship Class / Action Class

#### Create an Object Class

1. Open **BKN > Object Class**.
2. Click **+ New** and configure the object class as follows:
   - **Step 1: Data View**
     - If DIP already has a data view containing object-class data, select **Import Attributes from Data View**. The system maps data-view columns to attributes automatically, and unwanted attributes can be discarded later.
     - If there is no existing data source, choose **Create Attributes Manually**.
   - **Step 2: Basic Information**
     - **Name**: Display name of the object class
     - **ID**: Unique identifier of the object class
     - **Icon / Color**: Default icon and theme color for easier visual distinction
     - **Concept Group**: Assign the object class to a specific group for organization and filtering
     - **Description**: Briefly explain the purpose or meaning of the object class
   - **Step 3: Attribute Definition**
     - Every object class must contain at least one attribute used as the primary key to uniquely identify instances. Additional attributes can be added as needed.
     - Key parameters:
       - Primary key: choose one or more attributes, such as `Customer ID`, as the unique identifier
       - Title: specify one attribute as the default display name
       - Incremental: only integer (`integer`, `unsigned integer`) and time types (`datetime`, `timestamp`) are supported
   - **Step 4: Attribute Mapping**
     - Data properties: map basic attributes to the data view by dragging fields from the logic view on the left to logic properties on the right.
     - Logic properties: map to data models, operators, or functions. Bind the resource ID and configure the mapping between object-class data properties and logic-resource fields through **Logic Property > Target Property > Configure**.
3. Click **Save and Exit** to finish creating the object class.

#### Create a Relationship Class

1. Open **BKN > Relationship Class**.
2. Click **+ New** and configure the following:
   - **Name**: Should describe the association, such as `Customer - Associated With - Order`
   - **Start Object Class / End Object Class**: Choose the two object classes to be linked
   - **Association Method**:
     - Direct association: directly connect an attribute in one object class to an attribute in another object class, such as `Customer ID -> Order Table.Customer ID`
     - Indirect association: connect through a data view. This is used when direct association is not possible, such as `Device -> Device Status View -> Fault Record`
3. Click **Save and Exit** to finish creating the relationship class.

#### Create an Action Class

1. Open **BKN > Action Class**.
2. Click **+ New** and configure the following:
   - **Name**: Should represent the business operation, such as `Customer Churn Alert`
   - **Conditional Association**: Set the trigger rule for the action, such as `customer purchase count in the last 30 days < 2`
   - **Associated Tool**: Select the target tool from Execution Factory, such as an SMS push tool
3. Click **Save and Exit** to finish creating the action class.

### Configure Indexes

1. Open the detail page of an object class and click **Actions > Index Settings**.
2. Select the target property and click **Configure**. Complete the configuration in the side panel.
3. Click **Confirm** to finish the index configuration.

#### Index Configuration Rules

- Only fields of type **string**, **text**, and **vector** support index configuration. All other types are disabled by default.
- **string / text** fields support the following index capabilities:
  - Keyword index: used for exact matching and requires field length configuration
  - Full-text index: used for text search and requires a tokenizer selection (`standard`, `ik_max_word`, or `english`)
  - Vector index: after a small model is selected, the system automatically shows vector dimension, batch size, and max token count
- **vector** fields support **vector indexes only**.

### Import a BKN

1. On the **BKN** list page, click **Import**.
2. Upload a historical configuration file in JSON format.
3. The system automatically parses and imports concepts and mapping relationships without requiring you to remap views.
4. **Note**: After import, create index build tasks as needed to regenerate data indexes. Persisted data is not retained.

### Export a BKN

1. On the **BKN** list page, select the target BKN and click **More > Export**.
2. Choose the export format (`JSON`) and click **Confirm** to download.

## Appendix

### Attribute Type Reference

| Attribute Type | Description | Typical Usage | Example |
| --- | --- | --- | --- |
| boolean | Boolean value (`true` / `false`) | State flags, switch settings | `is_active: true` |
| short | 16-bit signed integer | Small numeric ranges, status codes | `age: 25` |
| integer | 32-bit signed integer | IDs, counters | `user_id: 12345` |
| long | 64-bit signed integer | Large numeric IDs, timestamps | `timestamp: 1640995200000` |
| float | Single-precision floating-point | Numeric values without high precision requirements | `price: 99.99` |
| double | Double-precision floating-point | Precise calculations, geographic coordinates | `latitude: 39.9042` |
| decimal | High-precision decimal | Financial amounts, interest rates | `amount: 12345.6789` |
| varchar | Variable-length string | Names, emails, and general text | `name: "Zhang San"` |
| keyword | Non-tokenized string for exact matching | Category tags, status codes | `status: "completed"` |
| text | Tokenized text for full-text search | Product descriptions, articles | `description: "This is a product description"` |
| date | Date without time | Birthdays, order dates | `birthday: "1990-05-15"` |
| datetime | Date and time | Creation time, update time | `created_at: "2024-01-20 14:30:25"` |
| timestamp | Timestamp in milliseconds or seconds | System logs, event times | `login_timestamp: 1642671025000` |
| vector | Vector data for similarity search | AI embeddings, recommendation systems | `embedding: [0.1, 0.2, 0.3]` |
| metric | Metric data for monitoring | System monitoring, business indicators | `cpu_usage: 75.5` |
| operator | Operator or function reference | Logical computation, business rules | `calculate_score` |

### Tokenizer Selection Guide

| Tokenizer | Suitable Language / Scenario | Characteristics |
| --- | --- | --- |
| Standard tokenizer | Mixed-language text | Splits by Unicode rules, automatically lowercases, and is broadly applicable |
| IK max-word tokenizer | Chinese text | Maximizes semantic completeness of words and reduces ambiguity |
| English tokenizer | Pure English text | Supports stemming, such as `running -> run`, to improve recall |
