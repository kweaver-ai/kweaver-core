-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

SET SCHEMA adp;

-- 移除关系类名称唯一性约束，同一 BKN 内允许同名关系类存在
DROP INDEX IF EXISTS adp.uk_t_relation_type_rt_name;

-- Risk Type
CREATE TABLE IF NOT EXISTS t_risk_type (
  f_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_name VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_tags VARCHAR(255 CHAR) DEFAULT NULL,
  f_icon VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_color VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_comment VARCHAR(1000 CHAR) NOT NULL DEFAULT '',
  f_kn_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_branch VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator_type VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_updater VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_updater_type VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  f_update_time BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_kn_id, f_branch, f_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_risk_type_name ON t_risk_type(f_kn_id, f_branch, f_name);
