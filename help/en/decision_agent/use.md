# How to Create and Use Decision Agent

The template feature allows users to publish an existing Decision Agent as a reusable template so that other users can quickly create similar agents.

### Create a Template

#### Steps

1. **Open the template creation entry**
   - Log in to ADP and go to **Development > Decision Agent**.
   - Find the Decision Agent card you want to publish as a template and hover over it.
   - Click the `···` menu in the upper-right corner of the card and select **Publish as Template**.

2. **Configure the template category**
   - After clicking **Publish as Template**, the system opens the **Publish Template** dialog.
   - Select the category for the template.
   - Click **Confirm** to complete template publishing.

3. **View published templates**
   - After a successful publish, the template is available under **Square > Templates**.
   - The template is displayed in the template marketplace according to the selected category and can be used by all users.

### Use a Template

#### Steps

1. **Open the template library and choose a template**
   - Log in to ADP and go to **Square > Templates**.
   - Browse or search for available templates.
   - Hover over the target template card, click the `···` menu in the upper-right corner, and select **Create from Template**.

2. **Enter the configuration page and customize**
   - After clicking **Create from Template**, the system jumps to the Decision Agent creation page and automatically loads all configuration information from the template.
   - On this page, you can review the preset configuration values and modify any configuration item as needed.

3. **Complete the configuration and publish**
   - Modify or keep the template settings based on your needs.
   - After editing, click **Publish**.

### Edit a Template

#### Steps

1. **Open the template editing entry**
   - Log in to ADP and go to **Development > My Templates**.
   - Find the template card you want to edit and click it to enter the editing page.

2. **Modify the template configuration**
   - On the template editing page, you can modify all template settings.

3. **Publish the updated template**
   - After editing, you can publish in either of the following ways:
   - **Option 1**: Click **Publish** directly on the editing page to republish the template to the template marketplace.
   - **Option 2**: Click **Save** to keep the changes, then return to **My Templates**, hover over the template card, open the `···` menu, and click **Publish Template**.

### Delete a Template

#### Steps

1. **Open the template deletion entry**
   - Log in to ADP and go to **Development > My Templates**.
   - Find the template card you want to delete and hover over it.

2. **Delete the template**
   - Click the `···` menu in the upper-right corner of the card and select **Delete**.
   - The system asks for confirmation. Click **Confirm** to complete the deletion.

### FAQ

1. **Q1: Which users can publish templates?**  
   **A1:** Users with the **template publishing permission** can publish the Decision Agent instances they created as templates.

2. **Q2: Will the original Decision Agent be affected after publishing it as a template?**  
   **A2:** No. Publishing as a template only creates a template copy. The original Decision Agent remains independent and can still be used and edited normally.

3. **Q3: After a template is published, can other users see my template configuration details?**  
   **A3:** Yes. Once the template is published to the template marketplace, all users can view its detailed configuration information, including role instructions, knowledge sources, and skill settings. The core value of a template is the reusability of its configuration.

4. **Q4: Can template access permissions be configured?**  
   **A4:** No. The current template feature does not support access control. All templates published to the template marketplace are visible to all users. If you need to restrict access, do not publish the agent as a template.

5. **Q5: If a template is deleted, will the Decision Agent instances created from it be affected?**  
   **A5:** No. Deleting a template does not affect any Decision Agent that was already created from it. Those agents continue to exist independently.
