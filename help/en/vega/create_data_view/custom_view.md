# Custom View

## Overview

A custom view is generated from one or more existing **atomic views**. It supports **filtering** and **deduplication** on the data from those atomic views to generate a new view that satisfies specific business requirements. It also provides **import**, **export**, and **group management** capabilities for efficient organization and maintenance of view resources.

## Create a Custom View

### Prerequisites

Before creating a custom view, make sure the following steps are complete:

1. A valid data connection has been created under **VEGA Virtualization > Data Connections**
2. At least one atomic view has already been scanned and synchronized under **VEGA Virtualization > Data Views > Atomic Views**

### Steps

#### Step 1: Open the Creation Page

1. On the **Custom View** page, click **New** at the top to open the **New Custom View** dialog.
2. Fill in the basic information:
   - **View Name**: must be unique. It is recommended to include the business scenario, such as `2025Q3 Active User View`
   - **Group**: choose an existing group from the dropdown, or select `Ungrouped`
   - **Description**: optional explanation of the purpose of the view
3. Click **Next** to enter the visual canvas page.

#### Step 2: Add Atomic Views

1. On the left toolbar of the canvas, click the **View Reference** icon to open the **Select Atomic View** window.
2. Select the atomic views to reference. Multiple selection is supported.
3. The selected atomic views appear on the canvas as **Referenced View** nodes.

#### Step 3: Configure Filtering and Deduplication

1. Click a **Referenced View** node. The **Data Configuration** panel appears at the bottom of the canvas.
2. Configure **Data Filtering** if needed:
   - Click **Add Filter Condition**, choose a field, an operator, and a value
   - Multiple conditions can be added, combined with `AND` or `OR`
3. Configure **Data Deduplication** if needed:
   - Enable **Deduplication**
   - Select the deduplication field, such as `User ID`
   - The system removes duplicate data according to the selected field and keeps the first matched record
4. Click **Confirm** to save the configuration of the referenced view node.

#### Step 4: Connect Nodes and Preview

1. On the canvas, drag from the output port of the **Referenced View** node to the input port of the **Output View** node to create the connection.
2. Click the **Output View** node. The **Field Preview** panel appears at the bottom of the canvas and shows the final field list.
3. If needed, drag field names in the **Field Preview** panel to adjust their order.

#### Step 5: Save the View

1. After confirming the configuration, click **Save** in the upper-right corner of the canvas. The system shows **Saved successfully**.
2. The custom view appears in the **Custom View** list.

### Alternative: Import a Custom View

If a custom view configuration file in JSON format has been exported from another environment, it can be imported directly:

1. On the **Custom View** page, click **Import** at the top.
2. Select the local JSON configuration file and click **Confirm**.
3. The system validates the file, including field mapping and atomic view relationships. If validation passes, the view is added to the selected group.

### View Details

1. In the **Custom View** list, click the name of the target view to enter the detail page.
2. On the detail page, you can inspect:
   - Basic information such as view name, group, creation time, and update time
   - Configuration information such as referenced atomic views, filter rules, and deduplication fields
   - A preview of the first 100 rows, with support for field-based sorting

## FAQ

Q1: **What should I do if the system says `Atomic view does not exist` when creating a custom view?**  
A1: Check the following:

- Open the **Atomic View** page and confirm whether the target atomic view exists
- If it was deleted, recreate the atomic view before configuring the custom view again

Q2: **What should I do if importing an exported custom view into another environment fails?**  
A2: Check the following:

- In the target environment, make sure the **Atomic View** list already contains atomic views with the same structure as the source environment
- If the fields are inconsistent, adjust the field configuration of the custom view in the source environment, export it again, and then reimport it
