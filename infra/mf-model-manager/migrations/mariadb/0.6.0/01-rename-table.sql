-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

-- ==========================================
-- 迁移脚本：将 mf-model-manager 相关表从 adp 库迁移至 kweaver 库
-- ==========================================
USE kweaver;

RENAME TABLE adp.t_llm_model TO kweaver.t_llm_model;
RENAME TABLE adp.t_small_model TO kweaver.t_small_model;
RENAME TABLE adp.t_prompt_item_list TO kweaver.t_prompt_item_list;
RENAME TABLE adp.t_prompt_list TO kweaver.t_prompt_list;
RENAME TABLE adp.t_prompt_template_list TO kweaver.t_prompt_template_list;
RENAME TABLE adp.t_model_monitor TO kweaver.t_model_monitor;
RENAME TABLE adp.t_model_quota_config TO kweaver.t_model_quota_config;
RENAME TABLE adp.t_user_quota_config TO kweaver.t_user_quota_config;
RENAME TABLE adp.t_model_op_detail TO kweaver.t_model_op_detail;
