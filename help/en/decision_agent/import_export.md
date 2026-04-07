# How to Import and Export Decision Agent?

## Import Decision Agent

### Steps

1. **Open the import entry**
   - Log in to ADP and click **Development > Decision Agent > Import**.

2. **Select the import mode**
   - After clicking **Import**, the system opens an **Import Mode** dialog.
   - Choose one of the following modes:
   - **Update mode**: If a Decision Agent with the same identifier already exists, the system updates that agent; otherwise, it creates a new one.
   - **Create mode**: The system only creates a new Decision Agent. If the identifier already exists, the import fails.
   - Click **Confirm** to enter the file selection screen.

3. **Select the import file**
   - In the file selection dialog, choose the local `Decision Agent.json` file.
   - Click **Open** to start the import.

4. **Complete the import**
   - The system processes the import file based on the selected mode.
   - After a successful import, you can view the imported Decision Agent under **Development > Decision Agent**.

## Export Decision Agent

### Steps

1. **Open the export entry**
   - Log in to ADP and click **Development > Decision Agent > Export**.

2. **Run the export**
   - After clicking **Export**, the page enters batch selection mode.  
     Note: the original source text says "click Import," which appears to be a wording mistake. Based on the context, it should mean "click Export."
   - Select the Decision Agent instances you want to export. Multiple selection is supported.
   - After making your selection, click **Export Selected**.

3. **Complete the export**
   - The system automatically generates a `.json` export file containing the selected agents and downloads it locally.

## FAQ

Q1: **What should I do if the system says "Invalid configuration" during import?**  
A1: Check the following:

- Confirm that the imported `.json` file was not manually modified in a way that breaks the required configuration format.
- It is recommended to import the original file exported by the system to avoid errors caused by format or field changes.

Q2: **What should I do if the system says "Insufficient permissions" during import?**  
A2: Check the following:

- Verify whether the imported `.json` file contains a **system Decision Agent** configuration.
- If it does and your current account lacks permission, contact the administrator to obtain the necessary access before importing again.

Q3: **What should I do if there is a "Decision Agent identifier conflict" during import?**  
A3: Handle it based on the selected import mode:

- Update mode: If the target system already has the same identifier and it does not belong to the current user, the import fails. You need to change the target or obtain permission.
- Create mode: If the target system already has the same identifier, the system automatically generates a new identifier to avoid the conflict.

## Notes

- Imported Decision Agent instances do not include data source information. You need to reconfigure the data source connection after the import.
- Built-in Decision Agent instances in the system cannot be exported, because those built-in agents already exist in other systems and cannot be modified.
- The exported file is in `.json` format and contains all configuration information for the Decision Agent, but does not include runtime data.
