# Concept Groups

**Scope**: ADP BKN module. Intended for business analysts, data engineers, AI engineers, and other users who need to manage ontology concepts in BKN.  
**Related module**: BKN > BKN

## Overview

### Feature Definition

Concept Group is a core feature in the ADP BKN used to logically organize and manage **object classes, relationship classes, and action classes** within a BKN. With this feature, users can arrange scattered ontology concepts into different groups according to actual business scenarios. A single concept can belong to multiple groups. Its key value is:

- **Improve ontology management efficiency**: business users can quickly locate related ontology concepts from a business perspective such as `Recruitment` or `Customer Service` without searching the full ontology set.
- **Accelerate ontology rollout**: implementers can quickly configure and verify business logic based on groups, reducing repetitive work.
- **Improve retrieval precision**: groups narrow the query scope and help users find target concepts more efficiently, especially in complex BKN scenarios.

### Application Scenarios

| User Role | Scenario Description | Functional Value |
| --- | --- | --- |
| Business Analyst | Needs to view all object classes and relationship classes under a business perspective such as `Recruitment` during requirement analysis | Quickly understand the business structure without searching the full ontology set |
| Data Engineer | Needs to process similar object classes, relationship classes, and action classes in batches when configuring data source mappings for agents | Supports batch export, import, and modification of ontology configurations, reducing repetitive work and configuration errors |
| AI Engineer | Needs to retrieve knowledge under a specific scenario, such as `Customer Service`, when building a RAG system | Restricts search to a concept group and provides more accurate knowledge data, improving the accuracy and relevance of agent responses |

## Prerequisites

Before using the concept group feature, make sure the following conditions are met:

1. You have logged in to ADP and entered the **BKN** module.
2. You have access to the target **BKN**, granted by the system administrator.
3. To create, modify, or delete groups, you must have the corresponding functional permissions. See the ADP permission management guide for details.

## Operation Guide

### Create a Concept Group

1. In ADP, open **BKN**, then click **BKN**. In the list, select the target BKN where the group should be created.
2. After entering the target BKN, click **Concept Group** in the left navigation bar to open the list page.
3. Click **New** in the upper-left corner to open the **Create Concept Group** dialog.
4. Fill in the required information. Fields marked with `*` are mandatory:
   - **Name***: Enter the group name. Up to 40 characters. It cannot be empty and must be unique under the same branch of the same BKN.
   - **ID**: Optional custom value. Only lowercase letters, digits, underscores, and hyphens are allowed. It cannot start with an underscore or hyphen. If not specified, the system generates it automatically. It cannot be changed after creation.
   - **Tags**: Up to 5 tags can be added. Each tag can be up to 40 characters and cannot contain `# \ / : * ? " < > |`.
   - **Remarks**: Optional. Up to 255 characters.
5. Click **Confirm** to create the concept group.

### View the Concept Group List and Details

#### View Group Details

1. Go to the **Concept Group** list page and locate the target group.
2. Click anywhere on the row or click **View** in the actions column to open the detail side panel.
3. In the side panel, you can see:
   - Basic information: group name, ID, tags, remarks, creation time, and update time
   - Related resources: counts and detailed lists of object classes, relationship classes, and action classes

#### Notes

- By default, all concept groups in the current BKN are sorted by **update time** in descending order.
- The **Related Resources** column displays the counts of object classes, relationship classes, and action classes separately.
- By default, only the first two tags are shown. If there are more than two, `+N` is displayed. Hover over `+N` to see all tags.

### Edit a Concept Group

1. Open the **Concept Group** list page and locate the group to edit.
2. Click **Edit** in the actions column to open the **Edit Concept Group** dialog.
3. Modify the information. The rules are the same as during creation, and the **group ID cannot be changed**:
   - Name: must remain unique within the same branch of the same BKN and can be up to 40 characters
   - Tags: new tags can be added or existing ones removed, up to 5 in total
   - Remarks: existing content can be updated or removed, up to 255 characters
4. Click **Confirm**. The system shows **Edit successful** and finishes the update.

### Delete a Concept Group

> ⚠️ Warning: After a group is deleted, all relationships between that group and ontology concepts are automatically removed. Group information and association records cannot be restored. The ontology concepts themselves are not deleted.

1. Open the **Concept Group** list page and locate the group to delete.
2. Click **Delete** in the actions column. A confirmation dialog appears with a warning that deleting the group may affect features using it.
3. Click **Confirm** to delete the group, or **Cancel** to abort.

### Add Object Classes to a Concept Group

> Note: After an object class is added, its related relationship classes and action classes are automatically associated with the group as well.

1. Open the **Concept Group** list page and click the target group to enter the editing page.
2. Under **Group Details** on the left, click the **Object Class** tab, then click **Add** to open the **Select Object Class** dialog.
3. Select the object classes to add:
   - The left side shows the full object class list. You can filter by name or by a single tag.
   - Select the object classes to add. Those already in the current group are marked as selected and cannot be selected again.
   - Selected object classes appear in the **Selected Object Classes** list on the right.
   - To remove a selected class, click the `×` button next to it in the right-side list.
4. After verifying the selection, click **Confirm**. The system shows **Added successfully**.

### Remove Object Classes from a Group

> ⚠️ Warning: After an object class is removed, its related action classes and all dangling relationship classes are also permanently removed from the group and cannot be restored. The object class itself is not deleted.

1. Open the **Concept Group** list page and click the target group to enter the editing page.
2. Under **Group Details** on the left, click the **Object Class** tab to view the current object class list.
3. Select one or more object classes to remove and click **Remove** above the list. A confirmation dialog appears.
4. Click **Confirm** to continue. The system shows **Removed successfully**.

### Import and Export Concept Groups

#### Export a Concept Group

1. Open the **Concept Group** list page and locate the group to export.
2. Click **Export** in the actions column. The system automatically generates and downloads a JSON export file.
3. The export includes:
   - Basic group information such as name, ID, tags, and remarks
   - Related ontology concepts such as object classes, relationship classes, and action classes

#### Import a Concept Group

> Note: Import supports only a single concept group at a time, using one JSON file. Batch import is not supported. Only the relationships between object classes and existing groups are retained. If a referenced group does not exist, the relationship is not retained.

1. Open the **Concept Group** list page in the target BKN and click **Import** in the upper-right corner.
2. Click **Choose File** and select a local JSON file. The system automatically filters out non-JSON files.
3. After upload, the system checks for duplicate IDs or names:
   - If there is no conflict, the import proceeds directly and the system shows **Import successful**.
   - If a conflict exists, the **Conflict Handling** dialog appears with these options:
   - **Overwrite**: replace the existing duplicate group with the information from the import file
   - **Create New**: create a brand-new group and write the imported content into it
   - **Ignore**: keep the existing group and skip the duplicate entry in the import file
   - You can also select **Apply this action to all subsequent identical conflicts**.
4. After choosing the conflict handling method, click **Confirm**. The system starts the import and shows **Import successful** when it finishes. If some content is not imported, the reason is displayed.

## Notes

- Concept groups **do not support nesting**. One group cannot contain another.
- Relationship classes and action classes **cannot be directly added, edited, or assigned to groups**. They can only be associated indirectly through related object classes.
- Group names must be **unique within the same BKN**. Custom group IDs must follow the allowed character rules and cannot be changed after creation.
- A maximum of 5 tags can be added. Each tag must be 40 characters or fewer and cannot contain `# \ / : * ? " < > |`.
- When an object class is deleted, the system automatically removes its association with **all concept groups**. These association records cannot be recovered.
- Import and export support only **single concept group** operations. Export files are in JSON format and include group information and related concepts.

## FAQ

Q1: **Why can't relationship classes or action classes be added to a concept group directly?**  
A1: Because:

- Relationship classes and action classes depend on object classes. For example, in the relationship `Employee - Belongs To - Department`, `Employee` and `Department` are object classes.
- Associating groups through object classes avoids invalid group relationships for relationship classes or action classes and preserves the logical consistency of BKN.
- If you need to add relationship classes or action classes, simply add the related object classes. The system synchronizes the related associations automatically.

Q2: **A conflict prompt appears when importing a concept group. How should I choose a handling method?**  
A2: Use the following guidance:

- **Overwrite**: use this when the imported file is the latest version and should replace the old group.
- **Create New**: use this when you want to keep the imported content without affecting the existing group.
- **Ignore**: use this when you want to keep the original group in the system and only import non-conflicting content.
- If there are multiple duplicate conflicts, you can select **Apply this action to all subsequent identical conflicts**.

Q3: **If a concept group is deleted, will previously associated ontology object classes also be deleted?**  
A3: No.

- Deleting a group only removes the associations between that group and all ontology concepts.
- The group information itself, including name, ID, tags, and remarks, is deleted together with the association records.
- Ontology object classes, relationship classes, and action classes remain in BKN and can still be queried, used, or associated with other groups later.
