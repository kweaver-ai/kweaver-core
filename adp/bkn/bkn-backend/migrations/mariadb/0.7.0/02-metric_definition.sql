-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

-- BKN 原生指标定义表（DESIGN §3.2 / IMPLEMENTATION_PLAN Task 2）

USE kweaver;

CREATE TABLE IF NOT EXISTS t_metric_definition (
  f_id VARCHAR(40) NOT NULL DEFAULT '' COMMENT '指标ID',
  f_kn_id VARCHAR(40) NOT NULL DEFAULT '' COMMENT '业务知识网络ID',
  f_branch VARCHAR(40) NOT NULL DEFAULT '' COMMENT '分支',
  f_name VARCHAR(128) NOT NULL DEFAULT '' COMMENT '指标技术名',
  f_comment VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '描述',
  f_unit_type VARCHAR(64) NOT NULL DEFAULT '' COMMENT '单位类型',
  f_unit VARCHAR(64) NOT NULL DEFAULT '' COMMENT '单位',
  f_metric_type VARCHAR(32) NOT NULL DEFAULT 'atomic' COMMENT '指标类型 atomic|derived|composite',
  f_scope_type VARCHAR(32) NOT NULL DEFAULT 'object_type' COMMENT '统计主体类型',
  f_scope_ref VARCHAR(40) NOT NULL DEFAULT '' COMMENT '对象类或子图ID',
  f_time_dimension LONGTEXT DEFAULT NULL COMMENT '时间维度 JSON',
  f_calculation_formula LONGTEXT NOT NULL COMMENT '计算公式 JSON',
  f_analysis_dimensions LONGTEXT DEFAULT NULL COMMENT '分析维度 JSON',
  f_creator VARCHAR(40) NOT NULL DEFAULT '' COMMENT '创建者id',
  f_creator_type VARCHAR(20) NOT NULL DEFAULT '' COMMENT '创建者类型',
  f_create_time BIGINT(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
  f_updater VARCHAR(40) NOT NULL DEFAULT '' COMMENT '更新者id',
  f_updater_type VARCHAR(20) NOT NULL DEFAULT '' COMMENT '更新者类型',
  f_update_time BIGINT(20) NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (f_kn_id, f_branch, f_id),
  UNIQUE KEY uk_metric_name (f_kn_id, f_branch, f_name)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = 'BKN 指标定义';
