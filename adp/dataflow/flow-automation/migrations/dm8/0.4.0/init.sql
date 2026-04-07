SET SCHEMA adp;


CREATE TABLE IF NOT EXISTS "t_model" (
  "f_id" BIGINT  NOT NULL,
  "f_name" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  "f_description" VARCHAR(300 CHAR) NOT NULL DEFAULT '',
  "f_train_status" VARCHAR(16 CHAR) NOT NULL DEFAULT '',
  "f_status" TINYINT NOT NULL,
  "f_rule" TEXT DEFAULT NULL,
  "f_userid" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  "f_type" TINYINT NOT NULL DEFAULT -1,
  "f_created_at" BIGINT DEFAULT NULL,
  "f_updated_at" BIGINT DEFAULT NULL,
  "f_scope" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS t_model_idx_t_model_f_name ON t_model(f_name);

CREATE INDEX IF NOT EXISTS t_model_idx_t_model_f_userid_status ON t_model(f_userid, f_status);

CREATE INDEX IF NOT EXISTS t_model_idx_t_model_f_status_type ON t_model(f_status, f_type);

CREATE TABLE IF NOT EXISTS "t_train_file" (
  "f_id" BIGINT  NOT NULL,
  "f_train_id" BIGINT  NOT NULL,
  "f_oss_id" VARCHAR(36 CHAR) DEFAULT '',
  "f_key" VARCHAR(36 CHAR) DEFAULT '',
  "f_created_at" BIGINT DEFAULT NULL,
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS t_train_file_idx_t_train_file_f_train_id ON t_train_file(f_train_id);

CREATE TABLE IF NOT EXISTS "t_automation_executor" (
  "f_id" BIGINT  NOT NULL,
  "f_name" VARCHAR(256 CHAR) NOT NULL DEFAULT '',
  "f_description" VARCHAR(1024 CHAR) NOT NULL DEFAULT '',
  "f_creator_id" VARCHAR(40 CHAR) NOT NULL,
  "f_status" TINYINT NOT NULL,
  "f_created_at" BIGINT DEFAULT NULL,
  "f_updated_at" BIGINT DEFAULT NULL,
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS t_automation_executor_idx_t_automation_executor_name ON t_automation_executor("f_name");

CREATE INDEX IF NOT EXISTS t_automation_executor_idx_t_automation_executor_creator_id ON t_automation_executor("f_creator_id");

CREATE INDEX IF NOT EXISTS t_automation_executor_idx_t_automation_executor_status ON t_automation_executor("f_status");

CREATE TABLE IF NOT EXISTS "t_automation_executor_accessor" (
  "f_id" BIGINT  NOT NULL,
  "f_executor_id" BIGINT  NOT NULL,
  "f_accessor_id" VARCHAR(40 CHAR) NOT NULL,
  "f_accessor_type" VARCHAR(20 CHAR) NOT NULL,
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS t_automation_executor_accessor_idx_t_automation_executor_accessor ON t_automation_executor_accessor("f_executor_id", "f_accessor_id", "f_accessor_type");

CREATE UNIQUE INDEX IF NOT EXISTS t_automation_executor_accessor_uk_executor_accessor ON t_automation_executor_accessor("f_executor_id", "f_accessor_id", "f_accessor_type");

CREATE TABLE IF NOT EXISTS "t_automation_executor_action" (
  "f_id" BIGINT  NOT NULL,
  "f_executor_id" BIGINT  NOT NULL,
  "f_operator" VARCHAR(64 CHAR) NOT NULL,
  "f_name" VARCHAR(256 CHAR) NOT NULL DEFAULT '',
  "f_description" VARCHAR(1024 CHAR) NOT NULL DEFAULT '',
  "f_group" VARCHAR(64 CHAR) NOT NULL DEFAULT '',
  "f_type" VARCHAR(16 CHAR) NOT NULL DEFAULT 'python',
  "f_inputs" text,
  "f_outputs" text,
  "f_config" text,
  "f_created_at" BIGINT DEFAULT NULL,
  "f_updated_at" BIGINT DEFAULT NULL,
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS t_automation_executor_action_idx_t_automation_executor_action_executor_id ON t_automation_executor_action("f_executor_id");

CREATE INDEX IF NOT EXISTS t_automation_executor_action_idx_t_automation_executor_action_operator ON t_automation_executor_action("f_operator");

CREATE INDEX IF NOT EXISTS t_automation_executor_action_idx_t_automation_executor_action_name ON t_automation_executor_action("f_name");

CREATE TABLE IF NOT EXISTS "t_content_admin" (
  "f_id" BIGINT  NOT NULL,
  "f_user_id" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  "f_user_name" VARCHAR(128 CHAR) NOT NULL DEFAULT '',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_content_admin_uk_f_user_id ON t_content_admin("f_user_id");

CREATE TABLE IF NOT EXISTS "t_audio_segments" (
  "f_id" BIGINT  NOT NULL,
  "f_task_id" VARCHAR(32 CHAR) NOT NULL,
  "f_object" VARCHAR(1024 CHAR) NOT NULL,
  "f_summary_type" VARCHAR(12 CHAR) NOT NULL,
  "f_max_segments" TINYINT NOT NULL,
  "f_max_segments_type" VARCHAR(12 CHAR) NOT NULL,
  "f_need_abstract" TINYINT NOT NULL,
  "f_abstract_type" VARCHAR(12 CHAR) NOT NULL,
  "f_callback" VARCHAR(1024 CHAR) NOT NULL,
  "f_created_at" BIGINT DEFAULT NULL,
  "f_updated_at" BIGINT DEFAULT NULL,
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE TABLE IF NOT EXISTS "t_automation_conf" (
  "f_key" VARCHAR(32 CHAR) NOT NULL,
  "f_value" VARCHAR(255 CHAR) NOT NULL,
  CLUSTER PRIMARY KEY ("f_key")
);

CREATE TABLE IF NOT EXISTS "t_automation_agent" (
  "f_id" BIGINT  NOT NULL,
  "f_name" VARCHAR(128 CHAR) NOT NULL DEFAULT '',
  "f_agent_id" VARCHAR(64 CHAR) NOT NULL DEFAULT '',
  "f_version" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS t_automation_agent_idx_t_automation_agent_agent_id ON t_automation_agent("f_agent_id");

CREATE UNIQUE INDEX IF NOT EXISTS t_automation_agent_uk_t_automation_agent_name ON t_automation_agent("f_name");

CREATE TABLE IF NOT EXISTS "t_alarm_rule" (
  "f_id" BIGINT  NOT NULL,
  "f_rule_id" BIGINT  NOT NULL,
  "f_dag_id" BIGINT  NOT NULL,
  "f_frequency" SMALLINT  NOT NULL,
  "f_threshold" INT  NOT NULL,
  "f_created_at" BIGINT DEFAULT NULL,
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS t_alarm_rule_idx_t_alarm_rule_rule_id ON t_alarm_rule("f_rule_id");

CREATE TABLE IF NOT EXISTS "t_alarm_user" (
  "f_id" BIGINT  NOT NULL,
  "f_rule_id" BIGINT  NOT NULL,
  "f_user_id" VARCHAR(36 CHAR) NOT NULL,
  "f_user_name" VARCHAR(128 CHAR) NOT NULL,
  "f_user_type" VARCHAR(10 CHAR) NOT NULL,
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS t_alarm_user_idx_t_alarm_user_rule_id ON t_alarm_user("f_rule_id");

CREATE TABLE IF NOT EXISTS "t_automation_dag_instance_ext_data" (
    "f_id" VARCHAR(64 CHAR) NOT NULL,
    "f_created_at" BIGINT DEFAULT NULL,
    "f_updated_at" BIGINT DEFAULT NULL,
    "f_dag_id" VARCHAR(64 CHAR),
    "f_dag_ins_id" VARCHAR(64 CHAR),
    "f_field" VARCHAR(64 CHAR) NOT NULL DEFAULT '',
    "f_oss_id" VARCHAR(64 CHAR) NOT NULL DEFAULT '',
    "f_oss_key" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
    "f_size" BIGINT  DEFAULT NULL,
    "f_removed" TINYINT NOT NULL DEFAULT 1,
    CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS t_automation_dag_instance_ext_data_idx_t_automation_dag_instance_ext_data_dag_ins_id ON t_automation_dag_instance_ext_data("f_dag_ins_id");

CREATE TABLE IF NOT EXISTS "t_task_cache_0" (
  "f_id" BIGINT  NOT NULL,
  "f_hash" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  "f_type" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
  "f_status" TINYINT NOT NULL DEFAULT '0',
  "f_oss_id" VARCHAR(36 CHAR) NOT NULL DEFAULT '',
  "f_oss_key" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  "f_ext" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  "f_size" BIGINT NOT NULL DEFAULT '0',
  "f_err_msg" TEXT NULL DEFAULT NULL,
  "f_create_time" BIGINT NOT NULL DEFAULT '0',
  "f_modify_time" BIGINT NOT NULL DEFAULT '0',
  "f_expire_time" BIGINT NOT NULL DEFAULT '0',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_task_cache_0_uk_hash ON t_task_cache_0("f_hash");

CREATE INDEX IF NOT EXISTS t_task_cache_0_idx_expire_time ON t_task_cache_0("f_expire_time");

CREATE TABLE IF NOT EXISTS "t_task_cache_1" (
  "f_id" BIGINT  NOT NULL,
  "f_hash" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  "f_type" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
  "f_status" TINYINT NOT NULL DEFAULT '0',
  "f_oss_id" VARCHAR(36 CHAR) NOT NULL DEFAULT '',
  "f_oss_key" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  "f_ext" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  "f_size" BIGINT NOT NULL DEFAULT '0',
  "f_err_msg" TEXT NULL DEFAULT NULL,
  "f_create_time" BIGINT NOT NULL DEFAULT '0',
  "f_modify_time" BIGINT NOT NULL DEFAULT '0',
  "f_expire_time" BIGINT NOT NULL DEFAULT '0',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_task_cache_1_uk_hash ON t_task_cache_1("f_hash");

CREATE INDEX IF NOT EXISTS t_task_cache_1_idx_expire_time ON t_task_cache_1("f_expire_time");

CREATE TABLE IF NOT EXISTS "t_task_cache_2" (
  "f_id" BIGINT  NOT NULL,
  "f_hash" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  "f_type" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
  "f_status" TINYINT NOT NULL DEFAULT '0',
  "f_oss_id" VARCHAR(36 CHAR) NOT NULL DEFAULT '',
  "f_oss_key" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  "f_ext" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  "f_size" BIGINT NOT NULL DEFAULT '0',
  "f_err_msg" TEXT NULL DEFAULT NULL,
  "f_create_time" BIGINT NOT NULL DEFAULT '0',
  "f_modify_time" BIGINT NOT NULL DEFAULT '0',
  "f_expire_time" BIGINT NOT NULL DEFAULT '0',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_task_cache_2_uk_hash ON t_task_cache_2("f_hash");

CREATE INDEX IF NOT EXISTS t_task_cache_2_idx_expire_time ON t_task_cache_2("f_expire_time");

CREATE TABLE IF NOT EXISTS "t_task_cache_3" (
  "f_id" BIGINT  NOT NULL,
  "f_hash" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  "f_type" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
  "f_status" TINYINT NOT NULL DEFAULT '0',
  "f_oss_id" VARCHAR(36 CHAR) NOT NULL DEFAULT '',
  "f_oss_key" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  "f_ext" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  "f_size" BIGINT NOT NULL DEFAULT '0',
  "f_err_msg" TEXT NULL DEFAULT NULL,
  "f_create_time" BIGINT NOT NULL DEFAULT '0',
  "f_modify_time" BIGINT NOT NULL DEFAULT '0',
  "f_expire_time" BIGINT NOT NULL DEFAULT '0',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_task_cache_3_uk_hash ON t_task_cache_3("f_hash");

CREATE INDEX IF NOT EXISTS t_task_cache_3_idx_expire_time ON t_task_cache_3("f_expire_time");

CREATE TABLE IF NOT EXISTS "t_task_cache_4" (
  "f_id" BIGINT  NOT NULL,
  "f_hash" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  "f_type" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
  "f_status" TINYINT NOT NULL DEFAULT '0',
  "f_oss_id" VARCHAR(36 CHAR) NOT NULL DEFAULT '',
  "f_oss_key" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  "f_ext" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  "f_size" BIGINT NOT NULL DEFAULT '0',
  "f_err_msg" TEXT NULL DEFAULT NULL,
  "f_create_time" BIGINT NOT NULL DEFAULT '0',
  "f_modify_time" BIGINT NOT NULL DEFAULT '0',
  "f_expire_time" BIGINT NOT NULL DEFAULT '0',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_task_cache_4_uk_hash ON t_task_cache_4("f_hash");

CREATE INDEX IF NOT EXISTS t_task_cache_4_idx_expire_time ON t_task_cache_4("f_expire_time");

CREATE TABLE IF NOT EXISTS "t_task_cache_5" (
  "f_id" BIGINT  NOT NULL,
  "f_hash" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  "f_type" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
  "f_status" TINYINT NOT NULL DEFAULT '0',
  "f_oss_id" VARCHAR(36 CHAR) NOT NULL DEFAULT '',
  "f_oss_key" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  "f_ext" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  "f_size" BIGINT NOT NULL DEFAULT '0',
  "f_err_msg" TEXT NULL DEFAULT NULL,
  "f_create_time" BIGINT NOT NULL DEFAULT '0',
  "f_modify_time" BIGINT NOT NULL DEFAULT '0',
  "f_expire_time" BIGINT NOT NULL DEFAULT '0',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_task_cache_5_uk_hash ON t_task_cache_5("f_hash");

CREATE INDEX IF NOT EXISTS t_task_cache_5_idx_expire_time ON t_task_cache_5("f_expire_time");

CREATE TABLE IF NOT EXISTS "t_task_cache_6" (
  "f_id" BIGINT  NOT NULL,
  "f_hash" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  "f_type" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
  "f_status" TINYINT NOT NULL DEFAULT '0',
  "f_oss_id" VARCHAR(36 CHAR) NOT NULL DEFAULT '',
  "f_oss_key" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  "f_ext" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  "f_size" BIGINT NOT NULL DEFAULT '0',
  "f_err_msg" TEXT NULL DEFAULT NULL,
  "f_create_time" BIGINT NOT NULL DEFAULT '0',
  "f_modify_time" BIGINT NOT NULL DEFAULT '0',
  "f_expire_time" BIGINT NOT NULL DEFAULT '0',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_task_cache_6_uk_hash ON t_task_cache_6("f_hash");

CREATE INDEX IF NOT EXISTS t_task_cache_6_idx_expire_time ON t_task_cache_6("f_expire_time");

CREATE TABLE IF NOT EXISTS "t_task_cache_7" (
  "f_id" BIGINT  NOT NULL,
  "f_hash" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  "f_type" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
  "f_status" TINYINT NOT NULL DEFAULT '0',
  "f_oss_id" VARCHAR(36 CHAR) NOT NULL DEFAULT '',
  "f_oss_key" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  "f_ext" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  "f_size" BIGINT NOT NULL DEFAULT '0',
  "f_err_msg" TEXT NULL DEFAULT NULL,
  "f_create_time" BIGINT NOT NULL DEFAULT '0',
  "f_modify_time" BIGINT NOT NULL DEFAULT '0',
  "f_expire_time" BIGINT NOT NULL DEFAULT '0',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_task_cache_7_uk_hash ON t_task_cache_7("f_hash");

CREATE INDEX IF NOT EXISTS t_task_cache_7_idx_expire_time ON t_task_cache_7("f_expire_time");

CREATE TABLE IF NOT EXISTS "t_task_cache_8" (
  "f_id" BIGINT  NOT NULL,
  "f_hash" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  "f_type" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
  "f_status" TINYINT NOT NULL DEFAULT '0',
  "f_oss_id" VARCHAR(36 CHAR) NOT NULL DEFAULT '',
  "f_oss_key" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  "f_ext" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  "f_size" BIGINT NOT NULL DEFAULT '0',
  "f_err_msg" TEXT NULL DEFAULT NULL,
  "f_create_time" BIGINT NOT NULL DEFAULT '0',
  "f_modify_time" BIGINT NOT NULL DEFAULT '0',
  "f_expire_time" BIGINT NOT NULL DEFAULT '0',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_task_cache_8_uk_hash ON t_task_cache_8("f_hash");

CREATE INDEX IF NOT EXISTS t_task_cache_8_idx_expire_time ON t_task_cache_8("f_expire_time");

CREATE TABLE IF NOT EXISTS "t_task_cache_9" (
  "f_id" BIGINT  NOT NULL,
  "f_hash" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  "f_type" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
  "f_status" TINYINT NOT NULL DEFAULT '0',
  "f_oss_id" VARCHAR(36 CHAR) NOT NULL DEFAULT '',
  "f_oss_key" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  "f_ext" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  "f_size" BIGINT NOT NULL DEFAULT '0',
  "f_err_msg" TEXT NULL DEFAULT NULL,
  "f_create_time" BIGINT NOT NULL DEFAULT '0',
  "f_modify_time" BIGINT NOT NULL DEFAULT '0',
  "f_expire_time" BIGINT NOT NULL DEFAULT '0',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_task_cache_9_uk_hash ON t_task_cache_9("f_hash");

CREATE INDEX IF NOT EXISTS t_task_cache_9_idx_expire_time ON t_task_cache_9("f_expire_time");

CREATE TABLE IF NOT EXISTS "t_task_cache_a" (
  "f_id" BIGINT  NOT NULL,
  "f_hash" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  "f_type" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
  "f_status" TINYINT NOT NULL DEFAULT '0',
  "f_oss_id" VARCHAR(36 CHAR) NOT NULL DEFAULT '',
  "f_oss_key" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  "f_ext" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  "f_size" BIGINT NOT NULL DEFAULT '0',
  "f_err_msg" TEXT NULL DEFAULT NULL,
  "f_create_time" BIGINT NOT NULL DEFAULT '0',
  "f_modify_time" BIGINT NOT NULL DEFAULT '0',
  "f_expire_time" BIGINT NOT NULL DEFAULT '0',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_task_cache_a_uk_hash ON t_task_cache_a("f_hash");

CREATE INDEX IF NOT EXISTS t_task_cache_a_idx_expire_time ON t_task_cache_a("f_expire_time");

CREATE TABLE IF NOT EXISTS "t_task_cache_b" (
  "f_id" BIGINT  NOT NULL,
  "f_hash" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  "f_type" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
  "f_status" TINYINT NOT NULL DEFAULT '0',
  "f_oss_id" VARCHAR(36 CHAR) NOT NULL DEFAULT '',
  "f_oss_key" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  "f_ext" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  "f_size" BIGINT NOT NULL DEFAULT '0',
  "f_err_msg" TEXT NULL DEFAULT NULL,
  "f_create_time" BIGINT NOT NULL DEFAULT '0',
  "f_modify_time" BIGINT NOT NULL DEFAULT '0',
  "f_expire_time" BIGINT NOT NULL DEFAULT '0',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_task_cache_b_uk_hash ON t_task_cache_b("f_hash");

CREATE INDEX IF NOT EXISTS t_task_cache_b_idx_expire_time ON t_task_cache_b("f_expire_time");

CREATE TABLE IF NOT EXISTS "t_task_cache_c" (
  "f_id" BIGINT  NOT NULL,
  "f_hash" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  "f_type" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
  "f_status" TINYINT NOT NULL DEFAULT '0',
  "f_oss_id" VARCHAR(36 CHAR) NOT NULL DEFAULT '',
  "f_oss_key" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  "f_ext" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  "f_size" BIGINT NOT NULL DEFAULT '0',
  "f_err_msg" TEXT NULL DEFAULT NULL,
  "f_create_time" BIGINT NOT NULL DEFAULT '0',
  "f_modify_time" BIGINT NOT NULL DEFAULT '0',
  "f_expire_time" BIGINT NOT NULL DEFAULT '0',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_task_cache_c_uk_hash ON t_task_cache_c("f_hash");

CREATE INDEX IF NOT EXISTS t_task_cache_c_idx_expire_time ON t_task_cache_c("f_expire_time");

CREATE TABLE IF NOT EXISTS "t_task_cache_d" (
  "f_id" BIGINT  NOT NULL,
  "f_hash" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  "f_type" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
  "f_status" TINYINT NOT NULL DEFAULT '0',
  "f_oss_id" VARCHAR(36 CHAR) NOT NULL DEFAULT '',
  "f_oss_key" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  "f_ext" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  "f_size" BIGINT NOT NULL DEFAULT '0',
  "f_err_msg" TEXT NULL DEFAULT NULL,
  "f_create_time" BIGINT NOT NULL DEFAULT '0',
  "f_modify_time" BIGINT NOT NULL DEFAULT '0',
  "f_expire_time" BIGINT NOT NULL DEFAULT '0',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_task_cache_d_uk_hash ON t_task_cache_d("f_hash");

CREATE INDEX IF NOT EXISTS t_task_cache_d_idx_expire_time ON t_task_cache_d("f_expire_time");

CREATE TABLE IF NOT EXISTS "t_task_cache_e" (
  "f_id" BIGINT  NOT NULL,
  "f_hash" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  "f_type" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
  "f_status" TINYINT NOT NULL DEFAULT '0',
  "f_oss_id" VARCHAR(36 CHAR) NOT NULL DEFAULT '',
  "f_oss_key" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  "f_ext" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  "f_size" BIGINT NOT NULL DEFAULT '0',
  "f_err_msg" TEXT NULL DEFAULT NULL,
  "f_create_time" BIGINT NOT NULL DEFAULT '0',
  "f_modify_time" BIGINT NOT NULL DEFAULT '0',
  "f_expire_time" BIGINT NOT NULL DEFAULT '0',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_task_cache_e_uk_hash ON t_task_cache_e("f_hash");

CREATE INDEX IF NOT EXISTS t_task_cache_e_idx_expire_time ON t_task_cache_e("f_expire_time");

CREATE TABLE IF NOT EXISTS "t_task_cache_f" (
  "f_id" BIGINT  NOT NULL,
  "f_hash" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  "f_type" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
  "f_status" TINYINT NOT NULL DEFAULT '0',
  "f_oss_id" VARCHAR(36 CHAR) NOT NULL DEFAULT '',
  "f_oss_key" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  "f_ext" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
  "f_size" BIGINT NOT NULL DEFAULT '0',
  "f_err_msg" TEXT NULL DEFAULT NULL,
  "f_create_time" BIGINT NOT NULL DEFAULT '0',
  "f_modify_time" BIGINT NOT NULL DEFAULT '0',
  "f_expire_time" BIGINT NOT NULL DEFAULT '0',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_task_cache_f_uk_hash ON t_task_cache_f("f_hash");

CREATE INDEX IF NOT EXISTS t_task_cache_f_idx_expire_time ON t_task_cache_f("f_expire_time");

CREATE TABLE IF NOT EXISTS "t_dag_instance_event" (
  "f_id" BIGINT  NOT NULL,
  "f_type" TINYINT NOT NULL DEFAULT '0',
  "f_instance_id" VARCHAR(64 CHAR) NOT NULL DEFAULT '',
  "f_operator" VARCHAR(128 CHAR) NOT NULL DEFAULT '',
  "f_task_id" VARCHAR(64 CHAR) NOT NULL DEFAULT '',
  "f_status" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
  "f_name" VARCHAR(128 CHAR) NOT NULL DEFAULT '',
  "f_data" TEXT NOT NULL,
  "f_size" BIGINT NOT NULL DEFAULT '0',
  "f_inline" TINYINT NOT NULL DEFAULT '0',
  "f_visibility" TINYINT NOT NULL DEFAULT '0',
  "f_timestamp" BIGINT NOT NULL DEFAULT '0',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS t_dag_instance_event_idx_instance_id ON t_dag_instance_event("f_instance_id", "f_id");

CREATE INDEX IF NOT EXISTS t_dag_instance_event_idx_instance_type_vis ON t_dag_instance_event("f_instance_id", "f_type", "f_visibility", "f_id");

CREATE INDEX IF NOT EXISTS t_dag_instance_event_idx_instance_name_type ON t_dag_instance_event("f_instance_id", "f_name", "f_type", "f_id");

INSERT INTO "t_automation_conf" (f_key, f_value) SELECT 'process_template', 1 FROM DUAL WHERE NOT EXISTS(SELECT "f_key", "f_value" FROM "t_automation_conf" WHERE "f_key"='process_template');

INSERT INTO "t_automation_conf" (f_key, f_value) SELECT 'ai_capabilities', 1 FROM DUAL WHERE NOT EXISTS(SELECT "f_key", "f_value" FROM "t_automation_conf" WHERE "f_key"='ai_capabilities');

CREATE TABLE IF NOT EXISTS "t_cron_job"
(
    "f_key_id" BIGINT NOT NULL IDENTITY(1, 1),
    "f_job_id" VARCHAR(36 CHAR) NOT NULL,
    "f_job_name" VARCHAR(64 CHAR) NOT NULL,
    "f_job_cron_time" VARCHAR(32 CHAR) NOT NULL,
    "f_job_type" TINYINT NOT NULL,
    "f_job_context" VARCHAR(10240 CHAR),
    "f_tenant_id" VARCHAR(36 CHAR),
    "f_enabled" TINYINT NOT NULL DEFAULT 1,
    "f_remarks" VARCHAR(256 CHAR),
    "f_create_time" BIGINT NOT NULL,
    "f_update_time" BIGINT NOT NULL,
    CLUSTER PRIMARY KEY ("f_key_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_cron_job_index_job_id ON t_cron_job("f_job_id");
CREATE UNIQUE INDEX IF NOT EXISTS t_cron_job_index_job_name ON t_cron_job("f_job_name", "f_tenant_id");
CREATE INDEX IF NOT EXISTS t_cron_job_index_tenant_id ON t_cron_job("f_tenant_id");
CREATE INDEX IF NOT EXISTS t_cron_job_index_time ON t_cron_job("f_create_time", "f_update_time");



CREATE TABLE IF NOT EXISTS "t_cron_job_status"
(
    "f_key_id" BIGINT NOT NULL IDENTITY(1, 1),
    "f_execute_id" VARCHAR(36 CHAR) NOT NULL,
    "f_job_id" VARCHAR(36 CHAR) NOT NULL,
    "f_job_type" TINYINT NOT NULL,
    "f_job_name" VARCHAR(64 CHAR) NOT NULL,
    "f_job_status" TINYINT NOT NULL,
    "f_begin_time" BIGINT,
    "f_end_time" BIGINT,
    "f_executor" VARCHAR(1024 CHAR),
    "f_execute_times" INT,
    "f_ext_info" VARCHAR(1024 CHAR),
    CLUSTER PRIMARY KEY ("f_key_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_cron_job_status_index_execute_id ON t_cron_job_status("f_execute_id");
CREATE INDEX IF NOT EXISTS t_cron_job_status_index_job_id ON t_cron_job_status("f_job_id");
CREATE INDEX IF NOT EXISTS t_cron_job_status_index_job_status ON t_cron_job_status("f_job_status");
CREATE INDEX IF NOT EXISTS t_cron_job_status_index_time ON t_cron_job_status("f_begin_time","f_end_time");

CREATE TABLE IF NOT EXISTS "t_flow_dag" (
 "f_id" BIGINT NOT NULL,
 "f_created_at" BIGINT NOT NULL DEFAULT 0,
 "f_updated_at" BIGINT NOT NULL DEFAULT 0,
 "f_user_id" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
 "f_name" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
 "f_desc" VARCHAR(310 CHAR) NOT NULL DEFAULT '',
 "f_trigger" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
 "f_cron" VARCHAR(64 CHAR) NOT NULL DEFAULT '',
 "f_vars" TEXT DEFAULT NULL,
 "f_status" VARCHAR(16 CHAR) NOT NULL DEFAULT '',
 "f_tasks" TEXT DEFAULT NULL,
 "f_steps" TEXT DEFAULT NULL,
 "f_description" VARCHAR(310 CHAR) NOT NULL DEFAULT '',
 "f_shortcuts" TEXT DEFAULT NULL,
 "f_accessors" TEXT DEFAULT NULL,
 "f_type" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
 "f_policy_type" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
 "f_appinfo" TEXT DEFAULT NULL,
 "f_priority" VARCHAR(16 CHAR) NOT NULL DEFAULT '',
 "f_removed" TINYINT NOT NULL DEFAULT 0,
 "f_emails" TEXT DEFAULT NULL,
 "f_template" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
 "f_published" TINYINT NOT NULL DEFAULT 0,
 "f_trigger_config" TEXT DEFAULT NULL,
 "f_sub_ids" TEXT DEFAULT NULL,
 "f_exec_mode" VARCHAR(8 CHAR) NOT NULL DEFAULT '',
 "f_category" VARCHAR(64 CHAR) NOT NULL DEFAULT '',
 "f_outputs" TEXT DEFAULT NULL,
 "f_instructions" TEXT DEFAULT NULL,
 "f_operator_id" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
 "f_inc_values" VARCHAR(4096 CHAR) DEFAULT NULL,
 "f_version" VARCHAR(64 CHAR) DEFAULT NULL,
 "f_version_id" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
 "f_modify_by" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
 "f_is_debug" TINYINT NOT NULL DEFAULT 0,
 "f_debug_id" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
 "f_biz_domain_id" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS "idx_dag_user_id" ON "t_flow_dag" ("f_user_id");
CREATE INDEX IF NOT EXISTS "idx_dag_type" ON "t_flow_dag" ("f_type");
CREATE INDEX IF NOT EXISTS "idx_dag_trigger" ON "t_flow_dag" ("f_trigger");
CREATE INDEX IF NOT EXISTS "idx_dag_name" ON "t_flow_dag" ("f_name");
CREATE INDEX IF NOT EXISTS "idx_dag_biz_domain" ON "t_flow_dag" ("f_biz_domain_id");

CREATE TABLE IF NOT EXISTS "t_flow_dag_var" (
 "f_id" BIGINT NOT NULL,
 "f_dag_id" BIGINT NOT NULL DEFAULT 0,
 "f_var_name" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
 "f_default_value" TEXT DEFAULT NULL,
 "f_var_type" VARCHAR(16 CHAR) NOT NULL DEFAULT '',
 "f_description" TEXT DEFAULT NULL,
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS "idx_dag_vars_dag_id" ON "t_flow_dag_var" ("f_dag_id");

CREATE TABLE IF NOT EXISTS "t_flow_dag_instance_keyword" (
 "f_id" BIGINT NOT NULL,
 "f_dag_ins_id" BIGINT NOT NULL DEFAULT 0,
 "f_keyword" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS "idx_dag_ins_kw" ON "t_flow_dag_instance_keyword" ("f_dag_ins_id", "f_keyword");

CREATE TABLE IF NOT EXISTS "t_flow_dag_step" (
 "f_id" BIGINT NOT NULL DEFAULT 0,
 "f_dag_id" BIGINT NOT NULL DEFAULT 0,
 "f_operator" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
 "f_source_id" TEXT NOT NULL,
 "f_has_datasource" TINYINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS "idx_dag_step_op" ON "t_flow_dag_step" ("f_operator");
CREATE INDEX IF NOT EXISTS "idx_dag_step_op_dag" ON "t_flow_dag_step" ("f_dag_id", "f_operator");
CREATE INDEX IF NOT EXISTS "idx_dag_step_has_ds_dag" ON "t_flow_dag_step" ("f_dag_id", "f_has_datasource");

CREATE TABLE IF NOT EXISTS "t_flow_dag_accessor" (
 "f_id" BIGINT NOT NULL,
 "f_dag_id" BIGINT NOT NULL DEFAULT 0,
 "f_accessor_id" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS "idx_dag_accessor_id_dag" ON "t_flow_dag_accessor" ("f_accessor_id", "f_dag_id");

CREATE TABLE IF NOT EXISTS "t_flow_dag_version" (
 "f_id" BIGINT NOT NULL,
 "f_created_at" BIGINT NOT NULL DEFAULT 0,
 "f_updated_at" BIGINT NOT NULL DEFAULT 0,
 "f_dag_id" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
 "f_user_id" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
 "f_version" VARCHAR(64 CHAR) NOT NULL DEFAULT '',
 "f_version_id" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
 "f_change_log" VARCHAR(512 CHAR) DEFAULT NULL,
 "f_config" TEXT DEFAULT NULL,
 "f_sort_time" BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS "idx_dag_versions_dag_version" ON "t_flow_dag_version" ("f_version_id", "f_dag_id");
CREATE INDEX IF NOT EXISTS "idx_dag_versions_dag_sort" ON "t_flow_dag_version" ("f_dag_id", "f_sort_time");

CREATE TABLE IF NOT EXISTS "t_flow_dag_instance" (
 "f_id" BIGINT NOT NULL,
 "f_created_at" BIGINT NOT NULL DEFAULT 0,
 "f_updated_at" BIGINT NOT NULL DEFAULT 0,
 "f_dag_id" BIGINT NOT NULL DEFAULT 0,
 "f_trigger" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
 "f_worker" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
 "f_source" TEXT DEFAULT NULL,
 "f_vars" TEXT DEFAULT NULL,
 "f_keywords" TEXT DEFAULT NULL,
 "f_event_persistence" TINYINT NOT NULL DEFAULT 0,
 "f_event_oss_path" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
 "f_share_data" TEXT DEFAULT NULL,
 "f_share_data_ext" TEXT DEFAULT NULL,
 "f_status" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
 "f_reason" TEXT DEFAULT NULL,
 "f_cmd" TEXT DEFAULT NULL,
 "f_has_cmd" TINYINT NOT NULL DEFAULT 0,
 "f_batch_run_id" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
 "f_user_id" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
 "f_ended_at" BIGINT NOT NULL DEFAULT 0,
 "f_dag_type" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
 "f_policy_type" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
 "f_appinfo" TEXT DEFAULT NULL,
 "f_priority" VARCHAR(16 CHAR) NOT NULL DEFAULT '',
 "f_mode" TINYINT NOT NULL DEFAULT 0,
 "f_dump" TEXT DEFAULT NULL,
 "f_dump_ext" TEXT DEFAULT NULL,
 "f_success_callback" VARCHAR(1024 CHAR) DEFAULT NULL,
 "f_error_callback" VARCHAR(1024 CHAR) DEFAULT NULL,
 "f_call_chain" TEXT DEFAULT NULL,
 "f_resume_data" TEXT DEFAULT NULL,
 "f_resume_status" VARCHAR(64 CHAR) NOT NULL DEFAULT '',
 "f_version" VARCHAR(64 CHAR) NOT NULL DEFAULT '',
 "f_version_id" VARCHAR(20 CHAR) NOT NULL DEFAULT '',
 "f_biz_domain_id" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS "idx_dag_ins_dag_status" ON "t_flow_dag_instance" ("f_dag_id", "f_status");
CREATE INDEX IF NOT EXISTS "idx_dag_ins_status_upd" ON "t_flow_dag_instance" ("f_status", "f_updated_at");
CREATE INDEX IF NOT EXISTS "idx_dag_ins_status_user_pri" ON "t_flow_dag_instance" ("f_status", "f_user_id", "f_priority");
CREATE INDEX IF NOT EXISTS "idx_dag_ins_user_id" ON "t_flow_dag_instance" ("f_user_id");
CREATE INDEX IF NOT EXISTS "idx_dag_ins_batch_run" ON "t_flow_dag_instance" ("f_batch_run_id");
CREATE INDEX IF NOT EXISTS "idx_dag_ins_worker" ON "t_flow_dag_instance" ("f_worker");

CREATE TABLE IF NOT EXISTS "t_flow_inbox" (
 "f_id" BIGINT NOT NULL,
 "f_created_at" BIGINT NOT NULL DEFAULT 0,
 "f_updated_at" BIGINT NOT NULL DEFAULT 0,
 "f_msg" TEXT DEFAULT NULL,
 "f_topic" VARCHAR(128 CHAR) NOT NULL DEFAULT '',
 "f_docid" VARCHAR(512 CHAR) NOT NULL DEFAULT '',
 "f_dag" TEXT DEFAULT NULL,
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS "idx_inbox_docid" ON "t_flow_inbox" ("f_docid");
CREATE INDEX IF NOT EXISTS "idx_inbox_topic_created" ON "t_flow_inbox" ("f_topic", "f_created_at");

CREATE TABLE IF NOT EXISTS "t_flow_outbox" (
 "f_id" BIGINT NOT NULL,
 "f_created_at" BIGINT NOT NULL DEFAULT 0,
 "f_updated_at" BIGINT NOT NULL DEFAULT 0,
 "f_msg" TEXT DEFAULT NULL,
 "f_topic" VARCHAR(128 CHAR) NOT NULL DEFAULT '',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS "idx_outbox_created" ON "t_flow_outbox" ("f_created_at");

CREATE TABLE IF NOT EXISTS "t_flow_task_instance" (
 "f_id" BIGINT NOT NULL,
 "f_created_at" BIGINT NOT NULL DEFAULT 0,
 "f_updated_at" BIGINT NOT NULL DEFAULT 0,
 "f_expired_at" BIGINT NOT NULL DEFAULT 0,
 "f_task_id" VARCHAR(64 CHAR) NOT NULL DEFAULT '',
 "f_dag_ins_id" BIGINT NOT NULL DEFAULT 0,
 "f_name" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
 "f_depend_on" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
 "f_action_name" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
 "f_timeout_secs" BIGINT NOT NULL DEFAULT 0,
 "f_params" TEXT DEFAULT NULL,
 "f_traces" TEXT DEFAULT NULL,
 "f_status" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
 "f_reason" TEXT DEFAULT NULL,
 "f_pre_checks" TEXT DEFAULT NULL,
 "f_results" TEXT DEFAULT NULL,
 "f_steps" TEXT DEFAULT NULL,
 "f_last_modified_at" BIGINT NOT NULL DEFAULT 0,
 "f_rendered_params" TEXT DEFAULT NULL,
 "f_hash" VARCHAR(64 CHAR) NOT NULL DEFAULT '',
 "f_settings" TEXT DEFAULT NULL,
 "f_metadata" TEXT DEFAULT NULL,
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS "idx_task_ins_dag_ins_id" ON "t_flow_task_instance" ("f_dag_ins_id");
CREATE INDEX IF NOT EXISTS "idx_task_ins_hash" ON "t_flow_task_instance" ("f_hash");
CREATE INDEX IF NOT EXISTS "idx_task_ins_action" ON "t_flow_task_instance" ("f_action_name");
CREATE INDEX IF NOT EXISTS "idx_task_ins_status_expire" ON "t_flow_task_instance" ("f_status", "f_expired_at");
CREATE INDEX IF NOT EXISTS "idx_task_ins_status_upd_id" ON "t_flow_task_instance" ("f_status", "f_updated_at", "f_id");

CREATE TABLE IF NOT EXISTS "t_flow_token" (
 "f_id" BIGINT NOT NULL,
 "f_created_at" BIGINT NOT NULL DEFAULT 0,
 "f_updated_at" BIGINT NOT NULL DEFAULT 0,
 "f_user_id" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
 "f_user_name" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
 "f_refresh_token" TEXT DEFAULT NULL,
 "f_token" TEXT DEFAULT NULL,
 "f_expires_in" INT NOT NULL DEFAULT 0,
 "f_login_ip" VARCHAR(64 CHAR) NOT NULL DEFAULT '',
 "f_is_app" TINYINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS "idx_token_user_id" ON "t_flow_token" ("f_user_id");

CREATE TABLE IF NOT EXISTS "t_flow_client" (
 "f_id" BIGINT NOT NULL,
 "f_created_at" BIGINT NOT NULL DEFAULT 0,
 "f_updated_at" BIGINT NOT NULL DEFAULT 0,
 "f_client_name" VARCHAR(64 CHAR) NOT NULL DEFAULT '',
 "f_client_id" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
 "f_client_secret" VARCHAR(16 CHAR) NOT NULL DEFAULT '',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS "idx_client_name" ON "t_flow_client" ("f_client_name");

CREATE TABLE IF NOT EXISTS "t_flow_switch" (
 "f_id" BIGINT NOT NULL,
 "f_created_at" BIGINT NOT NULL DEFAULT 0,
 "f_updated_at" BIGINT NOT NULL DEFAULT 0,
 "f_name" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
 "f_status" TINYINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS "idx_switch_name" ON "t_flow_switch" ("f_name");

CREATE TABLE IF NOT EXISTS "t_flow_log" (
 "f_id" BIGINT NOT NULL,
 "f_created_at" BIGINT NOT NULL DEFAULT 0,
 "f_updated_at" BIGINT NOT NULL DEFAULT 0,
 "f_ossid" VARCHAR(64 CHAR) NOT NULL DEFAULT '',
 "f_key" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
 "f_filename" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  CLUSTER PRIMARY KEY ("f_id")
);
