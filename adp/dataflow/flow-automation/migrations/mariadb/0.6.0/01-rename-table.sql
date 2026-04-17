-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

-- ==========================================
-- 迁移脚本：将 flow-automation 相关表从 adp 库迁移至 kweaver 库
-- ==========================================
USE kweaver;

RENAME TABLE adp.t_model TO kweaver.t_model;
RENAME TABLE adp.t_train_file TO kweaver.t_train_file;
RENAME TABLE adp.t_automation_executor TO kweaver.t_automation_executor;
RENAME TABLE adp.t_automation_executor_accessor TO kweaver.t_automation_executor_accessor;
RENAME TABLE adp.t_automation_executor_action TO kweaver.t_automation_executor_action;
RENAME TABLE adp.t_content_admin TO kweaver.t_content_admin;
RENAME TABLE adp.t_audio_segments TO kweaver.t_audio_segments;
RENAME TABLE adp.t_automation_agent TO kweaver.t_automation_agent;
RENAME TABLE adp.t_alarm_rule TO kweaver.t_alarm_rule;
RENAME TABLE adp.t_alarm_user TO kweaver.t_alarm_user;
RENAME TABLE adp.t_automation_dag_instance_ext_data TO kweaver.t_automation_dag_instance_ext_data;
RENAME TABLE adp.t_task_cache_0 TO kweaver.t_task_cache_0;
RENAME TABLE adp.t_task_cache_1 TO kweaver.t_task_cache_1;
RENAME TABLE adp.t_task_cache_2 TO kweaver.t_task_cache_2;
RENAME TABLE adp.t_task_cache_3 TO kweaver.t_task_cache_3;
RENAME TABLE adp.t_task_cache_4 TO kweaver.t_task_cache_4;
RENAME TABLE adp.t_task_cache_5 TO kweaver.t_task_cache_5;
RENAME TABLE adp.t_task_cache_6 TO kweaver.t_task_cache_6;
RENAME TABLE adp.t_task_cache_7 TO kweaver.t_task_cache_7;
RENAME TABLE adp.t_task_cache_8 TO kweaver.t_task_cache_8;
RENAME TABLE adp.t_task_cache_9 TO kweaver.t_task_cache_9;
RENAME TABLE adp.t_task_cache_a TO kweaver.t_task_cache_a;
RENAME TABLE adp.t_task_cache_b TO kweaver.t_task_cache_b;
RENAME TABLE adp.t_task_cache_c TO kweaver.t_task_cache_c;
RENAME TABLE adp.t_task_cache_d TO kweaver.t_task_cache_d;
RENAME TABLE adp.t_task_cache_e TO kweaver.t_task_cache_e;
RENAME TABLE adp.t_task_cache_f TO kweaver.t_task_cache_f;
RENAME TABLE adp.t_dag_instance_event TO kweaver.t_dag_instance_event;
RENAME TABLE adp.t_cron_job TO kweaver.t_cron_job;
RENAME TABLE adp.t_cron_job_status TO kweaver.t_cron_job_status;
RENAME TABLE adp.t_flow_dag TO kweaver.t_flow_dag;
RENAME TABLE adp.t_flow_dag_var TO kweaver.t_flow_dag_var;
RENAME TABLE adp.t_flow_dag_instance_keyword TO kweaver.t_flow_dag_instance_keyword;
RENAME TABLE adp.t_flow_dag_step TO kweaver.t_flow_dag_step;
RENAME TABLE adp.t_flow_dag_accessor TO kweaver.t_flow_dag_accessor;
RENAME TABLE adp.t_flow_dag_version TO kweaver.t_flow_dag_version;
RENAME TABLE adp.t_flow_dag_instance TO kweaver.t_flow_dag_instance;
RENAME TABLE adp.t_flow_inbox TO kweaver.t_flow_inbox;
RENAME TABLE adp.t_flow_outbox TO kweaver.t_flow_outbox;
RENAME TABLE adp.t_flow_task_instance TO kweaver.t_flow_task_instance;
RENAME TABLE adp.t_flow_token TO kweaver.t_flow_token;
RENAME TABLE adp.t_flow_client TO kweaver.t_flow_client;
RENAME TABLE adp.t_flow_switch TO kweaver.t_flow_switch;
RENAME TABLE adp.t_flow_log TO kweaver.t_flow_log;
RENAME TABLE adp.t_flow_storage TO kweaver.t_flow_storage;
RENAME TABLE adp.t_flow_file TO kweaver.t_flow_file;
RENAME TABLE adp.t_flow_file_download_job TO kweaver.t_flow_file_download_job;
RENAME TABLE adp.t_flow_task_resume TO kweaver.t_flow_task_resume;
RENAME TABLE adp.t_automation_conf TO kweaver.t_automation_conf;
