-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

USE adp;

-- 移除关系类名称唯一性约束，同一 BKN 内允许同名关系类存在
DROP INDEX IF EXISTS uk_relation_type_name ON t_relation_type;


CREATE TABLE IF NOT EXISTS t_risk_type (
  f_id VARCHAR(40) NOT NULL DEFAULT '' COMMENT '风险类ID',
  f_name VARCHAR(40) NOT NULL DEFAULT '' COMMENT '风险类名称',
  f_comment VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '描述',
  f_tags VARCHAR(255) DEFAULT NULL COMMENT '标签',
  f_icon VARCHAR(255) NOT NULL DEFAULT '' COMMENT '图标',
  f_color VARCHAR(40) NOT NULL DEFAULT '' COMMENT '颜色',
  f_kn_id VARCHAR(40) NOT NULL DEFAULT '' COMMENT '业务知识网络ID',
  f_branch VARCHAR(40) NOT NULL DEFAULT '' COMMENT '分支',
  f_creator VARCHAR(40) NOT NULL DEFAULT '' COMMENT '创建者id',
  f_creator_type VARCHAR(20) NOT NULL DEFAULT '' COMMENT '创建者类型',
  f_create_time BIGINT(20) NOT NULL DEFAULT 0 COMMENT '创建时间',
  f_updater VARCHAR(40) NOT NULL DEFAULT '' COMMENT '更新者id',
  f_updater_type VARCHAR(20) NOT NULL DEFAULT '' COMMENT '更新者类型',
  f_update_time BIGINT(20) NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (f_kn_id, f_branch, f_id),
  UNIQUE KEY uk_risk_type_name (f_kn_id, f_branch, f_name)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '风险类';
