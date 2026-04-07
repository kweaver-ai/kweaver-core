# How to Create a Data Connection?

This document explains how to create a data connection in the VEGA virtualization module of the DIP platform. It covers connection types, the general creation process, typical configuration scenarios, and follow-up scan management operations, helping you connect and manage multi-source data in a unified way.

> **Prerequisites**
1. You already have access to the DIP platform and the required permissions for the VEGA virtualization module, such as data connection management and scan management.
2. The target data source, such as a MySQL database or AnyShare storage, has been deployed and is reachable, and you already have the connection information such as IP address, port, username, and password.

## Introduction to Data Connections

The VEGA virtualization module supports **unified data connection management**. It can connect to structured data, unstructured data, and other types of data sources, providing the foundation for integrating multi-source data. Through the data processing engine layer, data can be extracted from different systems and intelligently combined to ensure complete and accurate decision support.

### Supported Data Connection Types

| Category | Specific Types | Default Connection Method |
| --- | --- | --- |
| Structured Data | MySQL, MariaDB, Oracle, PostgreSQL, SQLServer, Apache Doris, Hologres, openGauss, Dameng, GaussDB, MongoDB, Apache Hive, ClickHouse, TDH Inceptor, MaxCompute | JDBC, with Thrift optional for Hive |
| Unstructured Data | Excel, AnyShare 7.0 | HTTPS, with Excel depending on the storage medium |
| Other | TingYun, OpenSearch | Platform-specific adapter protocol |

## General Process for Creating a Data Connection

No matter which data connection type you use, the connection can be created through the following two steps. Additional type-specific settings are described later in the typical scenarios section.

### Open the Data Connection Management Page

1. Log in to the DIP platform and open **VEGA Virtualization > Data Connections > Connection Management**.
2. Click **New** to enter the data connection creation page.

### Select the Connection Type and Configure Connection Information

1. In the **Select Connection Type** step, choose the required data connection type from the category list on the left, such as **Structured Data > MySQL**, and click **Next**.
2. In the **Detailed Configuration** step, fill in the connection information according to the prompts on the page. Different connection types have slightly different fields. Typical items include:
   - Basic information: connection name and description
   - Connection information: IP address or domain, port, database name or instance name, account, and password
   - Advanced configuration, planned: timeout, character encoding, SSL configuration, and similar options
3. Click **Test Connection** to verify whether the connection is available:
   1. If the system reports **Connection failed**, check the following:
   - Whether the IP address and port are correct and the network is reachable
   - Whether the account and password are correct and have the required permissions
   - Whether the data source is running normally
   2. If the test succeeds, click **Confirm** to finish creating the data connection.

> **Notes**
On the **Connection Management** page, you can also perform the following operations on existing connections:

- View: open the detailed configuration and related information
- Edit: modify the configuration and retest the connection
- Delete: delete the connection only if it is not referenced by scan tasks or data views
- Search: filter connections by connection name or connection type

## Typical Data Connection Scenarios

This section provides extra configuration guidance and examples for common data connection types. For the general procedure, refer to the previous section.

### Structured Data - Create a MySQL Data Source

Example configuration for a `Service Management System Database`:

| Configuration Item | Example Value | Description |
| --- | --- | --- |
| Connection Name | MySQL - Service Management System | Custom and easy to identify |
| Data Source Type | Structured Data > MySQL | Fixed selection |
| Host Address | 192.168.1.100 | MySQL server IP or domain |
| Port | 3306 | Default MySQL port unless changed |
| Database Name | service_db | The target MySQL database |
| Account | db_user | MySQL account with access to the target database |
| Password | ******** | Password of the selected account |
| Connection Method | JDBC | Default, no change needed |

### Unstructured Data - Create an Excel Data Connection

Excel data connections support two storage media options: **AnyShare** for external storage and **Document Library** for internal DIP storage.

#### Storage Medium: AnyShare

| Configuration Item | Example Value | Description |
| --- | --- | --- |
| Connection Name | Excel - Aishu Technical Materials - AnyShare | Custom |
| Data Source Type | Unstructured Data > Excel | Fixed selection |
| Storage Medium | AnyShare | Fixed selection |
| Connection Address | anyshare.aishu.cn | IP or domain of the AnyShare service |
| Port | 443 | Default HTTPS port of AnyShare |
| User ID | app_tech | AnyShare application account created in the AnyShare admin console |
| Password | ******** | Password of the application account |
| Storage Path | Technical Documents / Excel Materials | Only department libraries or custom libraries are supported. No AnyShare prefix is needed. A single `.xlsx` file or a directory can be specified |

#### Storage Medium: Document Library

| Configuration Item | Example Value | Description |
| --- | --- | --- |
| Connection Name | Excel - Technical Materials - Document Library | Custom |
| Data Connection Type | Unstructured Data > Excel | Fixed selection |
| Storage Medium | Document Library | Fixed selection |
| Storage Path | Personal Library / Technical Materials | Only personal document libraries are supported. No AnyShare prefix is needed. A single `.xlsx` file or a directory can be specified |

#### Notes

1. Access permissions: you must use the user corresponding to the document library. Otherwise, the system reports that the file is not accessible.
2. AnyShare version: only AnyShare 7.0 and earlier versions are supported. Proton 3.0 and later versions changed the way application accounts are created and authorized, so department and custom libraries are not currently supported.
3. File type: only `.xlsx` files are supported. `.xls` and other formats are not supported.

## View Data Connection Details

After creating a data connection, use the following steps to inspect its details for verification or troubleshooting:

1. Open **VEGA Virtualization > Data Connections > Connection Management**.
2. In the connection list, find the target connection and click **View** in the actions column.
3. On the details page, you can inspect:
   - **Basic information**: name, type, creator, creation time, update time
   - **Table properties** for structured data: table name, table type, description
   - **Field properties** for structured data: field name, data type, whether it is a primary key, field description
   - **Connection configuration**: full connection details with masked passwords

## Scan Management

After creating a data connection, you need to use **Scan Management** to import source tables or files into the platform so that they can later be used to build data views and data models. Two scan methods are supported: create from the entire data connection, and create from selected tables in the data connection.

### Create a Scan Task from the Entire Data Connection

This method is suitable when you need to import all tables from a data source.

1. Open **VEGA Virtualization > Data Connections > Scan Management**.
2. Click **+ New Scan Task** in the upper-right corner and choose **Create from Entire Data Connection**.
3. In the dialog, select the data source to scan from the dropdown list, then click **Start Scan**.
4. During scanning, the page shows the real-time progress. **Refreshing or closing the page stops the scan**.
5. After the scan stops or finishes, the page shows the number of successfully scanned tables.
6. Return to **Scan Management** to view all scan tasks, including status, scanned count, creator, and creation time.

### Create a Scan Task from Tables in a Data Connection

This method is supported only for OpenSearch and is suitable when only part of the source tables need to be imported.

1. Open **VEGA Virtualization > Scan Management**.
2. Click **+ New Scan Task** and choose **Create from Tables in a Data Connection**.
3. In the dialog, choose the target OpenSearch connection, select one or more tables from the available table list, and click **Start Scan**.
4. Progress display and result prompts are the same as in the previous method.

## FAQ

Q1: **The system reports `Connection timeout` when testing a data connection. How should I handle it?**  
A1: Check the following:

- Verify that the data source server is running normally
- Check network reachability between the DIP platform server and the data source server
- Confirm that the target port is not blocked by a firewall

Q2: **If a scan task is interrupted, do the already scanned data connections need to be rescanned?**  
A2: No.

- Data that has already been scanned successfully remains in the **Scan Management** list
- If you need to continue scanning unfinished parts, create a new scan task for the same connection. The platform skips completed content automatically and scans only the unfinished part
