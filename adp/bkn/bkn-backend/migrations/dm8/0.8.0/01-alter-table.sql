-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

-- ==========================================
-- 0.7.0 → 0.8.0 升级脚本 (DM8)
-- t_action_type：新增行动意图、影响契约（与 Mariadb 0.8.0 对齐）
-- ==========================================
SET SCHEMA kweaver;

ALTER TABLE t_action_type ADD f_action_intent VARCHAR(40 CHAR) NOT NULL DEFAULT '';

ALTER TABLE t_action_type ADD f_impact_contracts TEXT DEFAULT NULL;
