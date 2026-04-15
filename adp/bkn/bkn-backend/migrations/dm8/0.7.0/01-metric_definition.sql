-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

-- BKN 原生指标定义表（与 mariadb/0.7.0 对齐）

SET SCHEMA kweaver;

CREATE TABLE IF NOT EXISTS t_metric_definition (
  f_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_kn_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_branch VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_name VARCHAR(128 CHAR) NOT NULL DEFAULT '',
  f_comment VARCHAR(1000 CHAR) NOT NULL DEFAULT '',
  f_unit_type VARCHAR(64 CHAR) NOT NULL DEFAULT '',
  f_unit VARCHAR(64 CHAR) NOT NULL DEFAULT '',
  f_metric_type VARCHAR(32 CHAR) NOT NULL DEFAULT 'atomic',
  f_scope_type VARCHAR(32 CHAR) NOT NULL DEFAULT 'object_type',
  f_scope_ref VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_time_dimension TEXT DEFAULT NULL,
  f_calculation_formula TEXT NOT NULL,
  f_analysis_dimensions TEXT DEFAULT NULL,
  f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator_type VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_updater VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_updater_type VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  f_update_time BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_kn_id, f_branch, f_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_metric_name ON t_metric_definition(f_kn_id, f_branch, f_name);