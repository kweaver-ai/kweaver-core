-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

-- ==========================================
-- 0.6.0 → 0.7.0 升级脚本 (DM8)
-- 1) 各概念表 f_comment: VARCHAR(1000 CHAR) NOT NULL DEFAULT '' → TEXT NOT NULL
-- 2) 各概念表 f_bkn_raw_content: TEXT DEFAULT NULL → TEXT NOT NULL
-- 3) t_knowledge_network 新增 f_skill_content TEXT NOT NULL
-- 4) t_risk_type 新增 f_bkn_raw_content TEXT NOT NULL
-- 说明: 存量 NULL 统一回填为 '' 后再收紧为 NOT NULL
-- ==========================================
SET SCHEMA kweaver;


-- ------------------------------------------
-- t_knowledge_network
-- ------------------------------------------
ALTER TABLE t_knowledge_network MODIFY f_comment TEXT NOT NULL;

UPDATE t_knowledge_network SET f_bkn_raw_content = '' WHERE f_bkn_raw_content IS NULL;

ALTER TABLE t_knowledge_network MODIFY f_bkn_raw_content TEXT NOT NULL;

ALTER TABLE t_knowledge_network ADD f_skill_content TEXT;

UPDATE t_knowledge_network SET f_skill_content = '' WHERE f_skill_content IS NULL;

ALTER TABLE t_knowledge_network MODIFY f_skill_content TEXT NOT NULL;


-- ------------------------------------------
-- t_object_type
-- ------------------------------------------
ALTER TABLE t_object_type MODIFY f_comment TEXT NOT NULL;

UPDATE t_object_type SET f_bkn_raw_content = '' WHERE f_bkn_raw_content IS NULL;

ALTER TABLE t_object_type MODIFY f_bkn_raw_content TEXT NOT NULL;


-- ------------------------------------------
-- t_relation_type
-- ------------------------------------------
ALTER TABLE t_relation_type MODIFY f_comment TEXT NOT NULL;

UPDATE t_relation_type SET f_bkn_raw_content = '' WHERE f_bkn_raw_content IS NULL;

ALTER TABLE t_relation_type MODIFY f_bkn_raw_content TEXT NOT NULL;


-- ------------------------------------------
-- t_action_type
-- ------------------------------------------
ALTER TABLE t_action_type MODIFY f_comment TEXT NOT NULL;

UPDATE t_action_type SET f_bkn_raw_content = '' WHERE f_bkn_raw_content IS NULL;

ALTER TABLE t_action_type MODIFY f_bkn_raw_content TEXT NOT NULL;


-- ------------------------------------------
-- t_concept_group
-- ------------------------------------------
ALTER TABLE t_concept_group MODIFY f_comment TEXT NOT NULL;

UPDATE t_concept_group SET f_bkn_raw_content = '' WHERE f_bkn_raw_content IS NULL;

ALTER TABLE t_concept_group MODIFY f_bkn_raw_content TEXT NOT NULL;


-- ------------------------------------------
-- t_risk_type
-- ------------------------------------------
ALTER TABLE t_risk_type MODIFY f_comment TEXT NOT NULL;

ALTER TABLE t_risk_type ADD f_bkn_raw_content TEXT;

UPDATE t_risk_type SET f_bkn_raw_content = '' WHERE f_bkn_raw_content IS NULL;

ALTER TABLE t_risk_type MODIFY f_bkn_raw_content TEXT NOT NULL;
