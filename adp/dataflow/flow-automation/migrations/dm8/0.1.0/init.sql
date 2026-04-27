SET SCHEMA adp;

CREATE TABLE IF NOT EXISTS "t_model" (
  "f_id" BIGINT  NOT NULL,
  "f_name" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  "f_description" VARCHAR(300 CHAR) NOT NULL DEFAULT '',
  "f_train_status" VARCHAR(16 CHAR) NOT NULL DEFAULT '',
  "f_status" TINYINT NOT NULL,
  "f_rule" text DEFAULT NULL,
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

