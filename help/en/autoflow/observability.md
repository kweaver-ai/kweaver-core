# Overview of Dataflow Observability

## Functional Value

Through the **Dataflow Overview** feature, you can achieve the following core capabilities:

- Monitor the overall health of Dataflow operations across the entire platform in real time, including but not limited to run quality and success/failure rate trends.
- Trace the complete execution process of a single Dataflow accurately and inspect detailed node-level data such as data volume, processing duration, and exception logs.
- Quickly identify abnormal stages in a Dataflow, reduce troubleshooting time, and support efficient and stable business operations.

## Role and Permission Description

Different roles in the DIP system have different visibility and operation permissions for data pipelines, as shown below:

| Role Type | Visibility Scope | Operation Permissions | Special Capabilities (planned) |
| --- | --- | --- | --- |
| Data Administrator | All data pipelines in the system | Full permissions for data pipelines, including create, edit, delete, and monitor | Can assign pipeline permissions to business users |
| AI Administrator | Only pipelines created by the user | Full permissions for the pipelines created by the user | Can assign pipeline permissions to business users |
| Business User (planned) | Assigned pipelines only | View and monitoring permissions only | No assignment permissions |

> **Note**: The data shown on the **Dataflow Overview** page is always limited by the viewer's own visibility scope. For example, a data administrator and an AI administrator will see completely different overview data.

## Dataflow Overview Methods

To support monitoring under different scenarios, the system provides two filtering methods to help you quickly locate target Dataflow runs.

**Entry**: Log in to DIP and go to **Autoflow > Dataflow > Overview**.

### Filter by Trigger Type

1. Find the **Trigger Type** dropdown in the filter bar at the top of the page.
2. Click the dropdown and choose one of the following trigger types:
   1. Manual trigger: Dataflow tasks started manually by users
   2. Scheduled trigger: Dataflow tasks executed automatically at preset times
   3. Event trigger: Dataflow tasks triggered by specific business events such as new data creation or file upload
3. After selecting a trigger type, the page refreshes in real time and displays all Dataflow runs under that trigger type.

### Filter by Custom Time Range

1. Find the **Time Range** selector in the filter bar at the top of the page.
2. Click the selector and choose one of the following methods:
   1. Quick ranges: preset time ranges such as the last 1 hour, last 24 hours, last 7 days, today, or yesterday
   2. Custom range: click **Custom** and manually set the start and end time in the time picker, with minute-level precision
3. After the selection is complete, click **Confirm**. The page refreshes and displays the Dataflow runs within the selected time range.

> **Tip**: You can combine both filters, such as `Scheduled Trigger + Last 24 Hours`, to narrow the data range further and improve monitoring efficiency.
