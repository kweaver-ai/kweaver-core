-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

-- ==========================================
-- 迁移脚本：将 bkn 相关表从 adp 库迁移至 kweaver 库
-- ==========================================
USE kweaver;

RENAME TABLE adp.t_knowledge_network TO kweaver.t_knowledge_network;
RENAME TABLE adp.t_object_type TO kweaver.t_object_type;
RENAME TABLE adp.t_object_type_status TO kweaver.t_object_type_status;
RENAME TABLE adp.t_relation_type TO kweaver.t_relation_type;
RENAME TABLE adp.t_action_type TO kweaver.t_action_type;
RENAME TABLE adp.t_kn_job TO kweaver.t_kn_job;
RENAME TABLE adp.t_kn_task TO kweaver.t_kn_task;
RENAME TABLE adp.t_concept_group TO kweaver.t_concept_group;
RENAME TABLE adp.t_concept_group_relation TO kweaver.t_concept_group_relation;
RENAME TABLE adp.t_action_schedule TO kweaver.t_action_schedule;
RENAME TABLE adp.t_risk_type TO kweaver.t_risk_type;
