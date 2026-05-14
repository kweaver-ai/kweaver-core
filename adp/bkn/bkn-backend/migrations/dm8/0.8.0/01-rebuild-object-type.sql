-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

-- ==========================================
-- 0.7.0 → 0.8.0 升级脚本 (DM8) - t_object_type 重建
-- 目的：t_object_type 新增 f_condition TEXT 字段用于存储过滤条件。
-- DM8 不允许对带 CLUSTER PRIMARY KEY 的表 ALTER 新增 LOB 字段，
-- 因此通过「建新表 → 迁数据 → 老表重命名保底 → 新表改名 → 删保底」重建。
-- ==========================================
SET SCHEMA kweaver;


DROP TABLE IF EXISTS t_object_type_new;

DROP TABLE IF EXISTS t_object_type_bak;

CREATE TABLE t_object_type_new (
  f_id               VARCHAR(40 CHAR)   NOT NULL DEFAULT '',
  f_name             VARCHAR(40 CHAR)   NOT NULL DEFAULT '',
  f_tags             VARCHAR(255 CHAR)  DEFAULT NULL,
  f_comment          TEXT               NOT NULL,
  f_icon             VARCHAR(255 CHAR)  NOT NULL DEFAULT '',
  f_color            VARCHAR(40 CHAR)   NOT NULL DEFAULT '',
  f_bkn_raw_content  TEXT               NOT NULL,
  f_kn_id            VARCHAR(40 CHAR)   NOT NULL DEFAULT '',
  f_branch           VARCHAR(40 CHAR)   NOT NULL DEFAULT '',
  f_data_source      VARCHAR(255 CHAR)  NOT NULL,
  f_data_properties  TEXT               DEFAULT NULL,
  f_logic_properties TEXT               DEFAULT NULL,
  f_primary_keys     VARCHAR(8192 CHAR) DEFAULT NULL,
  f_display_key      VARCHAR(40 CHAR)   NOT NULL DEFAULT '',
  f_incremental_key  VARCHAR(40 CHAR)   NOT NULL DEFAULT '',
  f_condition        TEXT               DEFAULT NULL,
  f_creator          VARCHAR(40 CHAR)   NOT NULL DEFAULT '',
  f_creator_type     VARCHAR(40 CHAR)   NOT NULL DEFAULT '',
  f_create_time      BIGINT             NOT NULL DEFAULT 0,
  f_updater          VARCHAR(40 CHAR)   NOT NULL DEFAULT '',
  f_updater_type     VARCHAR(40 CHAR)   NOT NULL DEFAULT '',
  f_update_time      BIGINT             NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_kn_id, f_branch, f_id)
);

INSERT INTO t_object_type_new (
  f_id, f_name, f_tags, f_comment, f_icon, f_color, f_bkn_raw_content,
  f_kn_id, f_branch, f_data_source, f_data_properties, f_logic_properties,
  f_primary_keys, f_display_key, f_incremental_key, f_condition,
  f_creator, f_creator_type, f_create_time,
  f_updater, f_updater_type, f_update_time
)
SELECT
  f_id, f_name, f_tags, f_comment, f_icon, f_color, f_bkn_raw_content,
  f_kn_id, f_branch, f_data_source, f_data_properties, f_logic_properties,
  f_primary_keys, f_display_key, f_incremental_key, NULL,
  f_creator, f_creator_type, f_create_time,
  f_updater, f_updater_type, f_update_time
FROM t_object_type;

ALTER TABLE t_object_type RENAME TO t_object_type_bak;

ALTER TABLE t_object_type_new RENAME TO t_object_type;

DROP TABLE t_object_type_bak;

CREATE UNIQUE INDEX uk_t_object_type_ot_name ON t_object_type(f_kn_id, f_branch, f_name);
