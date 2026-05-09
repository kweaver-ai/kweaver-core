-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

-- ==========================================
-- 0.6.0 → 0.7.0 升级脚本 (DM8) - t_risk_type 重建
-- DM8 不允许对带 CLUSTER PRIMARY KEY 的表 ALTER 新增 LOB 字段，
-- 因此通过「建新表 → 迁数据 → 老表重命名保底 → 新表改名 → 删保底」重建。
-- ==========================================
SET SCHEMA kweaver;


DROP TABLE IF EXISTS t_risk_type_new;

DROP TABLE IF EXISTS t_risk_type_bak;

CREATE TABLE t_risk_type_new (
  f_id              VARCHAR(40 CHAR)  NOT NULL DEFAULT '',
  f_name            VARCHAR(40 CHAR)  NOT NULL DEFAULT '',
  f_comment         TEXT              NOT NULL,
  f_tags            VARCHAR(255 CHAR) DEFAULT NULL,
  f_icon            VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_color           VARCHAR(40 CHAR)  NOT NULL DEFAULT '',
  f_kn_id           VARCHAR(40 CHAR)  NOT NULL DEFAULT '',
  f_branch          VARCHAR(40 CHAR)  NOT NULL DEFAULT '',
  f_creator         VARCHAR(40 CHAR)  NOT NULL DEFAULT '',
  f_creator_type    VARCHAR(20 CHAR)  NOT NULL DEFAULT '',
  f_create_time     BIGINT            NOT NULL DEFAULT 0,
  f_updater         VARCHAR(40 CHAR)  NOT NULL DEFAULT '',
  f_updater_type    VARCHAR(20 CHAR)  NOT NULL DEFAULT '',
  f_update_time     BIGINT            NOT NULL DEFAULT 0,
  f_bkn_raw_content TEXT              NOT NULL,
  CLUSTER PRIMARY KEY (f_kn_id, f_branch, f_id)
);

INSERT INTO t_risk_type_new (
  f_id, f_name, f_comment, f_tags, f_icon, f_color,
  f_kn_id, f_branch, f_creator, f_creator_type, f_create_time,
  f_updater, f_updater_type, f_update_time, f_bkn_raw_content
)
SELECT
  f_id, f_name, f_comment, f_tags, f_icon, f_color,
  f_kn_id, f_branch, f_creator, f_creator_type, f_create_time,
  f_updater, f_updater_type, f_update_time, ''
FROM t_risk_type;

ALTER TABLE t_risk_type RENAME TO t_risk_type_bak;

ALTER TABLE t_risk_type_new RENAME TO t_risk_type;

DROP TABLE t_risk_type_bak;

CREATE UNIQUE INDEX uk_risk_type_name ON t_risk_type(f_kn_id, f_branch, f_name);

