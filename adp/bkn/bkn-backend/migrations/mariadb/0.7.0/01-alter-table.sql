-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

-- ==========================================
-- 0.6.0 → 0.7.0 升级脚本
-- 1) 各概念表 f_comment: VARCHAR(1000) NOT NULL DEFAULT '' → TEXT NOT NULL
-- 2) 各概念表 f_bkn_raw_content: MEDIUMTEXT DEFAULT NULL → MEDIUMTEXT NOT NULL
-- 3) t_knowledge_network 新增 f_skill_content MEDIUMTEXT NOT NULL
-- 4) t_risk_type 新增 f_bkn_raw_content MEDIUMTEXT NOT NULL
-- 说明: 存量 NULL 统一回填为 '' 后再收紧为 NOT NULL
-- ==========================================
USE kweaver;


-- ------------------------------------------
-- t_knowledge_network
-- ------------------------------------------
ALTER TABLE t_knowledge_network
  MODIFY COLUMN f_comment TEXT NOT NULL COMMENT '备注';

UPDATE t_knowledge_network SET f_bkn_raw_content = '' WHERE f_bkn_raw_content IS NULL;

ALTER TABLE t_knowledge_network
  MODIFY COLUMN f_bkn_raw_content MEDIUMTEXT NOT NULL COMMENT 'BKNRawContent';

ALTER TABLE t_knowledge_network
  ADD COLUMN f_skill_content MEDIUMTEXT NULL COMMENT 'SkillContent' AFTER f_bkn_raw_content;

UPDATE t_knowledge_network SET f_skill_content = '' WHERE f_skill_content IS NULL;

ALTER TABLE t_knowledge_network
  MODIFY COLUMN f_skill_content MEDIUMTEXT NOT NULL COMMENT 'SkillContent';


-- ------------------------------------------
-- t_object_type
-- ------------------------------------------
ALTER TABLE t_object_type
  MODIFY COLUMN f_comment TEXT NOT NULL COMMENT '备注';

UPDATE t_object_type SET f_bkn_raw_content = '' WHERE f_bkn_raw_content IS NULL;

ALTER TABLE t_object_type
  MODIFY COLUMN f_bkn_raw_content MEDIUMTEXT NOT NULL COMMENT 'BKNRawContent';


-- ------------------------------------------
-- t_relation_type
-- ------------------------------------------
ALTER TABLE t_relation_type
  MODIFY COLUMN f_comment TEXT NOT NULL COMMENT '备注';

UPDATE t_relation_type SET f_bkn_raw_content = '' WHERE f_bkn_raw_content IS NULL;

ALTER TABLE t_relation_type
  MODIFY COLUMN f_bkn_raw_content MEDIUMTEXT NOT NULL COMMENT 'BKNRawContent';


-- ------------------------------------------
-- t_action_type
-- ------------------------------------------
ALTER TABLE t_action_type
  MODIFY COLUMN f_comment TEXT NOT NULL COMMENT '备注';

UPDATE t_action_type SET f_bkn_raw_content = '' WHERE f_bkn_raw_content IS NULL;

ALTER TABLE t_action_type
  MODIFY COLUMN f_bkn_raw_content MEDIUMTEXT NOT NULL COMMENT 'BKNRawContent';


-- ------------------------------------------
-- t_concept_group
-- ------------------------------------------
ALTER TABLE t_concept_group
  MODIFY COLUMN f_comment TEXT NOT NULL COMMENT '备注';

UPDATE t_concept_group SET f_bkn_raw_content = '' WHERE f_bkn_raw_content IS NULL;

ALTER TABLE t_concept_group
  MODIFY COLUMN f_bkn_raw_content MEDIUMTEXT NOT NULL COMMENT 'BKNRawContent';


-- ------------------------------------------
-- t_risk_type
-- ------------------------------------------
ALTER TABLE t_risk_type
  MODIFY COLUMN f_comment TEXT NOT NULL COMMENT '描述';

ALTER TABLE t_risk_type
  ADD COLUMN f_bkn_raw_content MEDIUMTEXT NULL COMMENT 'BKNRawContent' AFTER f_color;

UPDATE t_risk_type SET f_bkn_raw_content = '' WHERE f_bkn_raw_content IS NULL;

ALTER TABLE t_risk_type
  MODIFY COLUMN f_bkn_raw_content MEDIUMTEXT NOT NULL COMMENT 'BKNRawContent';
