# Evaluation Dataset

## Overview

In evaluation scenarios, an evaluation dataset is a carefully designed standardized test dataset used for systematic effectiveness evaluation. It usually contains input samples and expected outputs. Input samples are used as the input data for the evaluation target, while expected outputs provide the benchmark for evaluation. Before running evaluation experiments, you must prepare the dataset first. This document explains how to create and manage evaluation datasets in Intelligent Evaluation.

**Target users**: ADP administrators and data evaluation operators

## Introduction to Evaluation Datasets

An evaluation dataset is a set of data used to evaluate an evaluation target. It usually contains input data and expected outputs, helping developers verify the effectiveness of the target being evaluated.

An evaluation dataset usually contains the following columns:

- Input data: standardized test samples provided to the evaluation target and used to evaluate how an AI agent performs in different scenarios
- Expected output: ideal results used as the evaluation benchmark so evaluators or evaluation programs can judge the output

## Prerequisites

You are logged in to ADP and have access to the **Intelligent Evaluation** module. If not, contact the platform administrator.

## Steps

### Create a New Evaluation Dataset

1. Log in to ADP and choose **Evaluation Dataset** under **Intelligent Evaluation** from the left navigation bar.
2. Click **New** in the upper-right corner of the page.
3. In the dialog, fill in the required information. Fields marked with `*` are required:
   - **Evaluation Dataset Name***: enter a clear and recognizable name, such as `2025Q4 Customer Behavior Evaluation Dataset`. Maximum 50 characters
   - **CSV parsing rules***: configure the parsing method of the file and specify the field delimiter
   - **Description**: briefly describe the dataset purpose and included data types. Optional. Maximum 200 characters
   - **Color***: configure the default color for visual distinction in the UI
4. After configuration, click **Create** to enter the dataset detail page.

#### Evaluation Dataset Version Management

Evaluation datasets support **version management**, helping data-driven teams iterate datasets continuously during different evaluation stages and improve evaluation quality and result reliability.

- After an evaluation dataset is created for the first time, the system automatically generates the first version
- The system automatically creates the initial **version number** and **version description**
- You can click **Edit** on the right side of the page to modify the version name and description
- Additional versions can be created later to support continued iteration and evaluation

#### File and Version Management

This area is used to manage test files in the evaluation dataset and the versions they belong to.

- Open the target evaluation dataset, switch to the **Files and Versions** tab, and click **Upload** on the right side of the page to upload a test file or folder. The uploaded content is automatically included in the current version
- For uploaded files, download and delete operations are supported. Some file types, such as CSV and JSON, also support online preview
- Adding or removing files changes the composition of the current version, and these changes should be managed consistently at the version level

#### Evaluation Dataset Settings

This section is used to maintain the basic information of the evaluation dataset.

1. Open the target evaluation dataset and switch to the **Settings** tab.
2. On the settings page, you can maintain the dataset's **basic information**, including:
   - Modify the **name** of the evaluation dataset so that it more accurately reflects the business scenario or evaluation purpose
   - Edit the **description** to add information about the dataset content, source, or usage instructions
3. If the evaluation dataset is no longer needed, you can initiate a **delete** operation on the settings page. The system asks for confirmation before deletion. After deletion, the dataset, related versions, and files are permanently removed and **cannot be recovered**.
