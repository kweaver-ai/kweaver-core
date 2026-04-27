-- Copyright The kweaver.ai Authors.
--
-- Licensed under the Apache License, Version 2.0.
-- See the LICENSE file in the project root for details.

SET SCHEMA adp;

CREATE TABLE IF NOT EXISTS t_knowledge_network (
  f_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_name VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_tags VARCHAR(255 CHAR) DEFAULT NULL,
  f_comment VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_icon VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_color VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_detail TEXT DEFAULT NULL,
  f_branch VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_business_domain VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator_type VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_updater VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_updater_type VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_update_time BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_id,f_branch)
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_t_knowledge_network_kn_name ON t_knowledge_network(f_name,f_branch);


CREATE TABLE IF NOT EXISTS t_object_type (
  f_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_name VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_tags VARCHAR(255 CHAR) DEFAULT NULL,
  f_comment VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_icon VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_color VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_detail TEXT DEFAULT NULL,
  f_kn_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_branch VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_data_source VARCHAR(255 CHAR) NOT NULL,
  f_data_properties TEXT DEFAULT NULL,
  f_logic_properties TEXT DEFAULT NULL,
  f_primary_keys VARCHAR(8192 CHAR) DEFAULT NULL,
  f_display_key VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_incremental_key VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator_type VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_updater VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_updater_type VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_update_time BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_kn_id,f_branch,f_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_t_object_type_ot_name ON t_object_type(f_kn_id,f_branch,f_name);


-- 对象类状态
CREATE TABLE IF NOT EXISTS t_object_type_status (
  f_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_kn_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_branch VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_incremental_key VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_incremental_value VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_index VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_index_available BIT NOT NULL DEFAULT 0,
  f_doc_count BIGINT NOT NULL DEFAULT 0,
  f_storage_size BIGINT NOT NULL DEFAULT 0,
  f_update_time BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_kn_id,f_branch,f_id)
);


CREATE TABLE IF NOT EXISTS t_relation_type (
  f_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_name VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_tags VARCHAR(255 CHAR) DEFAULT NULL,
  f_comment VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_icon VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_color VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_detail TEXT DEFAULT NULL,
  f_kn_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_branch VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_source_object_type_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_target_object_type_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_type VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_mapping_rules text DEFAULT NULL,
  f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator_type VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_updater VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_updater_type VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_update_time BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_kn_id,f_branch,f_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_t_relation_type_rt_name ON t_relation_type(f_kn_id,f_branch,f_name);


CREATE TABLE IF NOT EXISTS t_action_type (
  f_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_name VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_tags VARCHAR(255 CHAR) DEFAULT NULL,
  f_comment VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_icon VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_color VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_detail TEXT DEFAULT NULL,
  f_kn_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_branch VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_action_type VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_object_type_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_condition TEXT DEFAULT NULL,
  f_affect TEXT DEFAULT NULL,
  f_action_source VARCHAR(255 CHAR) NOT NULL,
  f_parameters TEXT DEFAULT NULL,
  f_schedule VARCHAR(255 CHAR) DEFAULT NULL,
  f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator_type VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_updater VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_updater_type VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_update_time BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_kn_id,f_branch,f_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_t_action_type_at_name ON t_action_type(f_kn_id,f_branch,f_name);


CREATE TABLE IF NOT EXISTS t_kn_job (
  f_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_name VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_kn_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_branch VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_job_type VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_job_concept_config TEXT DEFAULT NULL,
  f_state VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_state_detail TEXT DEFAULT NULL,
  f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator_type VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_finish_time BIGINT NOT NULL DEFAULT 0,
  f_time_cost BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_id)
);


CREATE TABLE IF NOT EXISTS t_kn_task (
  f_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_name VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_job_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_concept_type VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_concept_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_index VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_doc_count BIGINT NOT NULL DEFAULT 0,
  f_state VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_state_detail TEXT DEFAULT NULL,
  f_start_time BIGINT NOT NULL DEFAULT 0,
  f_finish_time BIGINT NOT NULL DEFAULT 0,
  f_time_cost BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_id)
);


CREATE TABLE IF NOT EXISTS t_concept_group (
  f_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_name VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_tags VARCHAR(255 CHAR) DEFAULT NULL,
  f_comment VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_icon VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  f_color VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_detail TEXT DEFAULT NULL,
  f_kn_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_branch VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_creator_type VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  f_updater VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_updater_type VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_update_time BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_kn_id,f_branch,f_id)
) ;

CREATE UNIQUE INDEX IF NOT EXISTS uk_concept_group_name ON t_concept_group(f_kn_id,f_branch,f_name);


CREATE TABLE IF NOT EXISTS t_concept_group_relation (
  f_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_kn_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_branch VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_group_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_concept_type VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_concept_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  f_create_time BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY (f_id)
) ;

CREATE UNIQUE INDEX IF NOT EXISTS uk_concept_group_relation ON t_concept_group_relation(f_kn_id,f_branch,f_group_id,f_concept_type,f_concept_id);
