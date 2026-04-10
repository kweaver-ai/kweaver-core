-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

-- ==========================================
-- 迁移脚本：在 kweaver schema 下创建 vega-gateway 相关表，并从 adp schema 复制数据
-- ==========================================

SET SCHEMA kweaver;

-- ==========================================
-- 1. cache_table
-- ==========================================
CREATE TABLE IF NOT EXISTS "cache_table" (
  "id" varchar(36 char) NOT NULL COMMENT '主键',
  "catalog_name" varchar(36 char) NOT NULL COMMENT '对应的逻辑视图的catalog名称',
  "schema_name" varchar(36 char) NOT NULL COMMENT '对应的逻辑视图的schema名称',
  "table_name" varchar(36 char) NOT NULL COMMENT '对应的逻辑视图的table名称',
  "cts_sql" text DEFAULT NULL COMMENT '表的建表sql',
  "source_create_sql" text DEFAULT NULL COMMENT '样例数据查询sql',
  "current_view_original_text" text DEFAULT NULL COMMENT '最近一次的原始加密sql',
  "status" varchar(36 char) NOT NULL COMMENT '可用；异常；正在初始化',
  "mid_status" varchar(36 char) DEFAULT NULL COMMENT '在FSM任务的时候的中间状态',
  "deps" varchar(255 char) DEFAULT '' COMMENT '生成的结果缓存表的id用,分隔',
  "create_time" datetime(3) NOT NULL DEFAULT current_timestamp(3) COMMENT '更新时间',
  "update_time" datetime(3) NOT NULL DEFAULT current_timestamp(3) COMMENT '更新时间',
  CLUSTER PRIMARY KEY ("id")
);

INSERT INTO kweaver."cache_table" (
    "id", "catalog_name", "schema_name", "table_name",
    "cts_sql", "source_create_sql", "current_view_original_text",
    "status", "mid_status", "deps",
    "create_time", "update_time"
)
SELECT
    s."id", s."catalog_name", s."schema_name", s."table_name",
    s."cts_sql", s."source_create_sql", s."current_view_original_text",
    s."status", s."mid_status", s."deps",
    s."create_time", s."update_time"
FROM adp."cache_table" s;

-- ==========================================
-- 2. client_id
-- ==========================================
CREATE TABLE IF NOT EXISTS "client_id" (
  "id" int NOT NULL COMMENT '主键id',
  "client_name" varchar(128 char) DEFAULT NULL COMMENT '客户端名称',
  "client_id" varchar(64 char) DEFAULT NULL COMMENT '客户端id',
  "client_secret" varchar(64 char) DEFAULT NULL COMMENT '客户端密码',
  "create_time" datetime DEFAULT NULL COMMENT '创建时间',
  "update_time" datetime DEFAULT NULL COMMENT '更新时间',
  CLUSTER PRIMARY KEY ("id")
);

INSERT INTO kweaver."client_id" (
    "id", "client_name", "client_id", "client_secret",
    "create_time", "update_time"
)
SELECT
    s."id", s."client_name", s."client_id", s."client_secret",
    s."create_time", s."update_time"
FROM adp."client_id" s;

-- ==========================================
-- 3. excel_column_type（IDENTITY 列，需显式插入）
-- ==========================================
CREATE TABLE IF NOT EXISTS "excel_column_type" (
  "id" bigint NOT NULL IDENTITY(1,1) COMMENT '主键id',
  "catalog" varchar(256 char) NOT NULL COMMENT '数据源',
  "vdm_catalog" varchar(256 char) DEFAULT NULL COMMENT 'vdm数据源',
  "schema_name" varchar(256 char) NOT NULL COMMENT '库名',
  "table_name" varchar(512 char) NOT NULL COMMENT '表名',
  "column_name" varchar(128 char) NOT NULL COMMENT '列名',
  "column_comment" varchar(512 char) DEFAULT NULL COMMENT '列注释',
  "type" varchar(128 char) NOT NULL COMMENT '字段类型',
  "order_no" int NOT NULL COMMENT '列序号',
  "create_time" timestamp NOT NULL DEFAULT current_timestamp() COMMENT '创建时间',
  "update_time" timestamp NOT NULL DEFAULT current_timestamp() COMMENT '更新时间',
  CLUSTER PRIMARY KEY ("id")
);

SET IDENTITY_INSERT kweaver.excel_column_type ON;

INSERT INTO kweaver."excel_column_type" (
    "id", "catalog", "vdm_catalog", "schema_name", "table_name",
    "column_name", "column_comment", "type", "order_no",
    "create_time", "update_time"
)
SELECT
    s."id", s."catalog", s."vdm_catalog", s."schema_name", s."table_name",
    s."column_name", s."column_comment", s."type", s."order_no",
    s."create_time", s."update_time"
FROM adp."excel_column_type" s;

SET IDENTITY_INSERT kweaver.excel_column_type OFF;

-- ==========================================
-- 4. excel_table_config（IDENTITY 列，需显式插入）
-- ==========================================
CREATE TABLE IF NOT EXISTS "excel_table_config" (
  "id" bigint NOT NULL IDENTITY(1,1) COMMENT '主键id',
  "catalog" varchar(256 char) NOT NULL COMMENT '数据源',
  "vdm_catalog" varchar(256 char) DEFAULT NULL COMMENT 'vdm数据源',
  "schema_name" varchar(256 char) NOT NULL COMMENT '库名',
  "file_name" varchar(512 char) NOT NULL COMMENT 'excel文件名',
  "table_name" varchar(512 char) NOT NULL COMMENT '表名',
  "table_comment" varchar(512 char) DEFAULT NULL COMMENT '表注释',
  "sheet" varchar(128 char) DEFAULT NULL COMMENT 'sheet名称',
  "all_sheet" tinyint NOT NULL DEFAULT 0 COMMENT '是否加载所有sheet',
  "sheet_as_new_column" tinyint NOT NULL DEFAULT 0 COMMENT 'sheet是否作为列 1:是 0:否',
  "start_cell" varchar(32 char) DEFAULT NULL COMMENT '起始单元格',
  "end_cell" varchar(32 char) DEFAULT NULL COMMENT '结束单元格',
  "has_headers" tinyint NOT NULL DEFAULT 1 COMMENT '是否有表头  1：有； 0：没有',
  "create_time" timestamp NULL DEFAULT current_timestamp() COMMENT '创建时间',
  "update_time" timestamp NOT NULL DEFAULT current_timestamp() COMMENT '更新时间',
  CLUSTER PRIMARY KEY ("id")
);

CREATE UNIQUE INDEX IF NOT EXISTS "excel_table_config_vdm_table_uindex" on "excel_table_config" ("catalog","table_name");

SET IDENTITY_INSERT kweaver.excel_table_config ON;

INSERT INTO kweaver."excel_table_config" (
    "id", "catalog", "vdm_catalog", "schema_name",
    "file_name", "table_name", "table_comment", "sheet",
    "all_sheet", "sheet_as_new_column", "start_cell", "end_cell",
    "has_headers", "create_time", "update_time"
)
SELECT
    s."id", s."catalog", s."vdm_catalog", s."schema_name",
    s."file_name", s."table_name", s."table_comment", s."sheet",
    s."all_sheet", s."sheet_as_new_column", s."start_cell", s."end_cell",
    s."has_headers", s."create_time", s."update_time"
FROM adp."excel_table_config" s;

SET IDENTITY_INSERT kweaver.excel_table_config OFF;

-- ==========================================
-- 5. query_info
-- ==========================================
CREATE TABLE IF NOT EXISTS "query_info" (
  "query_id" varchar(30 char) NOT NULL COMMENT 'query id',
  "result" text DEFAULT NULL COMMENT '查询结果集',
  "msg" varchar(500 char) DEFAULT NULL COMMENT '错误详情',
  "task_id" varchar(200 char) NOT NULL COMMENT '任务Id',
  "state" varchar(30 char) NOT NULL COMMENT '状态',
  "create_time" varchar(30 char) NOT NULL COMMENT '创建时间',
  "update_time" varchar(30 char) NOT NULL COMMENT '更新时间',
  CLUSTER PRIMARY KEY ("query_id")
);

INSERT INTO kweaver."query_info" (
    "query_id", "result", "msg", "task_id",
    "state", "create_time", "update_time"
)
SELECT
    s."query_id", s."result", s."msg", s."task_id",
    s."state", s."create_time", s."update_time"
FROM adp."query_info" s;

-- ==========================================
-- 6. task_info
-- ==========================================
CREATE TABLE IF NOT EXISTS "task_info" (
  "task_id" varchar(200 char) NOT NULL COMMENT '主键taskid',
  "state" varchar(30 char) DEFAULT NULL COMMENT 'task状态',
  "query" text DEFAULT NULL,
  "create_time" varchar(30 char) DEFAULT NULL COMMENT '创建时间',
  "update_time" varchar(30 char) DEFAULT NULL COMMENT '修改时间',
  "topic" varchar(100 char) DEFAULT NULL COMMENT 'topic名称',
  "sub_task_id" varchar(200 char) NOT NULL COMMENT '子任务Id',
  "type" int NOT NULL DEFAULT 1 COMMENT '类型,0:异步查询,1:字段探查',
  "elapsed_time" varchar(30 char) NOT NULL COMMENT '总耗时',
  "update_count" text NOT NULL COMMENT '结果集大小,只针对insert into或create table as记录大小',
  "schedule_time" varchar(30 char) NOT NULL COMMENT '调度耗时',
  "queued_time" varchar(30 char) NOT NULL COMMENT '队列耗时',
  "cpu_time" varchar(30 char) NOT NULL COMMENT 'cpu耗时',
  CLUSTER PRIMARY KEY ("task_id","sub_task_id")
);

INSERT INTO kweaver."task_info" (
    "task_id", "state", "query",
    "create_time", "update_time", "topic", "sub_task_id",
    "type", "elapsed_time", "update_count",
    "schedule_time", "queued_time", "cpu_time"
)
SELECT
    s."task_id", s."state", s."query",
    s."create_time", s."update_time", s."topic", s."sub_task_id",
    s."type", s."elapsed_time", s."update_count",
    s."schedule_time", s."queued_time", s."cpu_time"
FROM adp."task_info" s;
