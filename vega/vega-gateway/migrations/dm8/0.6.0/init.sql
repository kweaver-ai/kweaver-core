-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

SET SCHEMA kweaver;


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


CREATE TABLE IF NOT EXISTS "client_id" (
  "id" int NOT NULL COMMENT '主键id',
  "client_name" varchar(128 char) DEFAULT NULL COMMENT '客户端名称',
  "client_id" varchar(64 char) DEFAULT NULL COMMENT '客户端id',
  "client_secret" varchar(64 char) DEFAULT NULL COMMENT '客户端密码',
  "create_time" datetime DEFAULT NULL COMMENT '创建时间',
  "update_time" datetime DEFAULT NULL COMMENT '更新时间',
  CLUSTER PRIMARY KEY ("id")
);


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
