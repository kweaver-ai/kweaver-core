-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

-- ==========================================
-- 0.7.0 → 0.8.0 升级脚本
-- 1) t_metric_definition：补充 0.7.0 增量建表脚本遗漏的字段
-- 2) t_object_type：新增 f_condition 字段用于存储过滤条件
-- 使用 IF NOT EXISTS：兼容已通过 init.sql 建表的环境（已有这些列）。
-- ==========================================
USE kweaver;

ALTER TABLE t_metric_definition
  ADD COLUMN IF NOT EXISTS f_tags VARCHAR(255) DEFAULT NULL COMMENT '标签' AFTER f_comment;

ALTER TABLE t_metric_definition
  ADD COLUMN IF NOT EXISTS f_icon VARCHAR(255) NOT NULL DEFAULT '' COMMENT '图标' AFTER f_tags;

ALTER TABLE t_metric_definition
  ADD COLUMN IF NOT EXISTS f_color VARCHAR(40) NOT NULL DEFAULT '' COMMENT '颜色' AFTER f_icon;

ALTER TABLE t_metric_definition
  ADD COLUMN IF NOT EXISTS f_bkn_raw_content MEDIUMTEXT NULL COMMENT 'BKNRawContent' AFTER f_color;

UPDATE t_metric_definition SET f_bkn_raw_content = '' WHERE f_bkn_raw_content IS NULL;

ALTER TABLE t_metric_definition
  MODIFY COLUMN f_bkn_raw_content MEDIUMTEXT NOT NULL COMMENT 'BKNRawContent' AFTER f_color;

-- ==========================================
-- t_action_type：行动意图、影响契约（与代码 DTO 对齐）
-- 使用 IF NOT EXISTS：兼容已通过 init.sql 建表的环境（已有这些列）。
-- ==========================================
ALTER TABLE t_action_type
  ADD COLUMN IF NOT EXISTS f_action_intent VARCHAR(40) NOT NULL DEFAULT '' COMMENT '行动意图' AFTER f_action_type;

ALTER TABLE t_action_type
  ADD COLUMN IF NOT EXISTS f_impact_contracts TEXT DEFAULT NULL COMMENT '行动影响契约' AFTER f_action_intent;

-- t_object_type 新增 f_condition 字段
ALTER TABLE t_object_type
  ADD COLUMN IF NOT EXISTS f_condition TEXT DEFAULT NULL COMMENT '对象类过滤条件' AFTER f_incremental_key;
