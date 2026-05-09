SET SCHEMA kweaver;


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


-- t_model
INSERT INTO kweaver."t_model" ("f_id", "f_name", "f_description", "f_train_status", "f_status", "f_rule", "f_userid", "f_type", "f_created_at", "f_updated_at", "f_scope")
SELECT s."f_id", s."f_name", s."f_description", s."f_train_status", s."f_status", s."f_rule", s."f_userid", s."f_type", s."f_created_at", s."f_updated_at", s."f_scope"
FROM adp."t_model" s;


CREATE TABLE IF NOT EXISTS "t_train_file" (
  "f_id" BIGINT  NOT NULL,
  "f_train_id" BIGINT  NOT NULL,
  "f_oss_id" VARCHAR(36 CHAR) DEFAULT '',
  "f_key" VARCHAR(36 CHAR) DEFAULT '',
  "f_created_at" BIGINT DEFAULT NULL,
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS t_train_file_idx_t_train_file_f_train_id ON t_train_file(f_train_id);


-- t_train_file
INSERT INTO kweaver."t_train_file" ("f_id", "f_train_id", "f_oss_id", "f_key", "f_created_at")
SELECT s."f_id", s."f_train_id", s."f_oss_id", s."f_key", s."f_created_at"
FROM adp."t_train_file" s;


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


-- t_automation_executor
INSERT INTO kweaver."t_automation_executor" ("f_id", "f_name", "f_description", "f_creator_id", "f_status", "f_created_at", "f_updated_at")
SELECT s."f_id", s."f_name", s."f_description", s."f_creator_id", s."f_status", s."f_created_at", s."f_updated_at"
FROM adp."t_automation_executor" s;


CREATE TABLE IF NOT EXISTS "t_automation_executor_accessor" (
  "f_id" BIGINT  NOT NULL,
  "f_executor_id" BIGINT  NOT NULL,
  "f_accessor_id" VARCHAR(40 CHAR) NOT NULL,
  "f_accessor_type" VARCHAR(20 CHAR) NOT NULL,
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS t_automation_executor_accessor_idx_t_automation_executor_accessor ON t_automation_executor_accessor("f_executor_id", "f_accessor_id", "f_accessor_type");

CREATE UNIQUE INDEX IF NOT EXISTS t_automation_executor_accessor_uk_executor_accessor ON t_automation_executor_accessor("f_executor_id", "f_accessor_id", "f_accessor_type");


-- t_automation_executor_accessor
INSERT INTO kweaver."t_automation_executor_accessor" ("f_id", "f_executor_id", "f_accessor_id", "f_accessor_type")
SELECT s."f_id", s."f_executor_id", s."f_accessor_id", s."f_accessor_type"
FROM adp."t_automation_executor_accessor" s;


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


-- t_automation_executor_action
INSERT INTO kweaver."t_automation_executor_action" ("f_id", "f_executor_id", "f_operator", "f_name", "f_description", "f_group", "f_type", "f_inputs", "f_outputs", "f_config", "f_created_at", "f_updated_at")
SELECT s."f_id", s."f_executor_id", s."f_operator", s."f_name", s."f_description", s."f_group", s."f_type", s."f_inputs", s."f_outputs", s."f_config", s."f_created_at", s."f_updated_at"
FROM adp."t_automation_executor_action" s;


CREATE TABLE IF NOT EXISTS "t_content_admin" (
  "f_id" BIGINT  NOT NULL,
  "f_user_id" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  "f_user_name" VARCHAR(128 CHAR) NOT NULL DEFAULT '',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS t_content_admin_uk_f_user_id ON t_content_admin("f_user_id");


-- t_content_admin
INSERT INTO kweaver."t_content_admin" ("f_id", "f_user_id", "f_user_name")
SELECT s."f_id", s."f_user_id", s."f_user_name"
FROM adp."t_content_admin" s;


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


-- t_audio_segments
INSERT INTO kweaver."t_audio_segments" ("f_id", "f_task_id", "f_object", "f_summary_type", "f_max_segments", "f_max_segments_type", "f_need_abstract", "f_abstract_type", "f_callback", "f_created_at", "f_updated_at")
SELECT s."f_id", s."f_task_id", s."f_object", s."f_summary_type", s."f_max_segments", s."f_max_segments_type", s."f_need_abstract", s."f_abstract_type", s."f_callback", s."f_created_at", s."f_updated_at"
FROM adp."t_audio_segments" s;


CREATE TABLE IF NOT EXISTS "t_automation_conf" (
  "f_key" VARCHAR(32 CHAR) NOT NULL,
  "f_value" VARCHAR(255 CHAR) NOT NULL,
  CLUSTER PRIMARY KEY ("f_key")
);


-- t_automation_conf
INSERT INTO kweaver."t_automation_conf" ("f_key", "f_value")
SELECT s."f_key", s."f_value"
FROM adp."t_automation_conf" s;


CREATE TABLE IF NOT EXISTS "t_automation_agent" (
  "f_id" BIGINT  NOT NULL,
  "f_name" VARCHAR(128 CHAR) NOT NULL DEFAULT '',
  "f_agent_id" VARCHAR(64 CHAR) NOT NULL DEFAULT '',
  "f_version" VARCHAR(32 CHAR) NOT NULL DEFAULT '',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS t_automation_agent_idx_t_automation_agent_agent_id ON t_automation_agent("f_agent_id");

CREATE UNIQUE INDEX IF NOT EXISTS t_automation_agent_uk_t_automation_agent_name ON t_automation_agent("f_name");


-- t_automation_agent
INSERT INTO kweaver."t_automation_agent" ("f_id", "f_name", "f_agent_id", "f_version")
SELECT s."f_id", s."f_name", s."f_agent_id", s."f_version"
FROM adp."t_automation_agent" s;


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


-- t_alarm_rule
INSERT INTO kweaver."t_alarm_rule" ("f_id", "f_rule_id", "f_dag_id", "f_frequency", "f_threshold", "f_created_at")
SELECT s."f_id", s."f_rule_id", s."f_dag_id", s."f_frequency", s."f_threshold", s."f_created_at"
FROM adp."t_alarm_rule" s;


CREATE TABLE IF NOT EXISTS "t_alarm_user" (
  "f_id" BIGINT  NOT NULL,
  "f_rule_id" BIGINT  NOT NULL,
  "f_user_id" VARCHAR(36 CHAR) NOT NULL,
  "f_user_name" VARCHAR(128 CHAR) NOT NULL,
  "f_user_type" VARCHAR(10 CHAR) NOT NULL,
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS t_alarm_user_idx_t_alarm_user_rule_id ON t_alarm_user("f_rule_id");


-- t_alarm_user
INSERT INTO kweaver."t_alarm_user" ("f_id", "f_rule_id", "f_user_id", "f_user_name", "f_user_type")
SELECT s."f_id", s."f_rule_id", s."f_user_id", s."f_user_name", s."f_user_type"
FROM adp."t_alarm_user" s;


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


-- t_automation_dag_instance_ext_data
INSERT INTO kweaver."t_automation_dag_instance_ext_data" ("f_id", "f_created_at", "f_updated_at", "f_dag_id", "f_dag_ins_id", "f_field", "f_oss_id", "f_oss_key", "f_size", "f_removed")
SELECT s."f_id", s."f_created_at", s."f_updated_at", s."f_dag_id", s."f_dag_ins_id", s."f_field", s."f_oss_id", s."f_oss_key", s."f_size", s."f_removed"
FROM adp."t_automation_dag_instance_ext_data" s;


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


-- t_task_cache_0
INSERT INTO kweaver."t_task_cache_0" ("f_id", "f_hash", "f_type", "f_status", "f_oss_id", "f_oss_key", "f_ext", "f_size", "f_err_msg", "f_create_time", "f_modify_time", "f_expire_time")
SELECT s."f_id", s."f_hash", s."f_type", s."f_status", s."f_oss_id", s."f_oss_key", s."f_ext", s."f_size", s."f_err_msg", s."f_create_time", s."f_modify_time", s."f_expire_time"
FROM adp."t_task_cache_0" s;


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


-- t_task_cache_1
INSERT INTO kweaver."t_task_cache_1" ("f_id", "f_hash", "f_type", "f_status", "f_oss_id", "f_oss_key", "f_ext", "f_size", "f_err_msg", "f_create_time", "f_modify_time", "f_expire_time")
SELECT s."f_id", s."f_hash", s."f_type", s."f_status", s."f_oss_id", s."f_oss_key", s."f_ext", s."f_size", s."f_err_msg", s."f_create_time", s."f_modify_time", s."f_expire_time"
FROM adp."t_task_cache_1" s;


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


-- t_task_cache_2
INSERT INTO kweaver."t_task_cache_2" ("f_id", "f_hash", "f_type", "f_status", "f_oss_id", "f_oss_key", "f_ext", "f_size", "f_err_msg", "f_create_time", "f_modify_time", "f_expire_time")
SELECT s."f_id", s."f_hash", s."f_type", s."f_status", s."f_oss_id", s."f_oss_key", s."f_ext", s."f_size", s."f_err_msg", s."f_create_time", s."f_modify_time", s."f_expire_time"
FROM adp."t_task_cache_2" s;


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


-- t_task_cache_3
INSERT INTO kweaver."t_task_cache_3" ("f_id", "f_hash", "f_type", "f_status", "f_oss_id", "f_oss_key", "f_ext", "f_size", "f_err_msg", "f_create_time", "f_modify_time", "f_expire_time")
SELECT s."f_id", s."f_hash", s."f_type", s."f_status", s."f_oss_id", s."f_oss_key", s."f_ext", s."f_size", s."f_err_msg", s."f_create_time", s."f_modify_time", s."f_expire_time"
FROM adp."t_task_cache_3" s;


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


-- t_task_cache_4
INSERT INTO kweaver."t_task_cache_4" ("f_id", "f_hash", "f_type", "f_status", "f_oss_id", "f_oss_key", "f_ext", "f_size", "f_err_msg", "f_create_time", "f_modify_time", "f_expire_time")
SELECT s."f_id", s."f_hash", s."f_type", s."f_status", s."f_oss_id", s."f_oss_key", s."f_ext", s."f_size", s."f_err_msg", s."f_create_time", s."f_modify_time", s."f_expire_time"
FROM adp."t_task_cache_4" s;


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


-- t_task_cache_5
INSERT INTO kweaver."t_task_cache_5" ("f_id", "f_hash", "f_type", "f_status", "f_oss_id", "f_oss_key", "f_ext", "f_size", "f_err_msg", "f_create_time", "f_modify_time", "f_expire_time")
SELECT s."f_id", s."f_hash", s."f_type", s."f_status", s."f_oss_id", s."f_oss_key", s."f_ext", s."f_size", s."f_err_msg", s."f_create_time", s."f_modify_time", s."f_expire_time"
FROM adp."t_task_cache_5" s;


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


-- t_task_cache_6
INSERT INTO kweaver."t_task_cache_6" ("f_id", "f_hash", "f_type", "f_status", "f_oss_id", "f_oss_key", "f_ext", "f_size", "f_err_msg", "f_create_time", "f_modify_time", "f_expire_time")
SELECT s."f_id", s."f_hash", s."f_type", s."f_status", s."f_oss_id", s."f_oss_key", s."f_ext", s."f_size", s."f_err_msg", s."f_create_time", s."f_modify_time", s."f_expire_time"
FROM adp."t_task_cache_6" s;


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


-- t_task_cache_7
INSERT INTO kweaver."t_task_cache_7" ("f_id", "f_hash", "f_type", "f_status", "f_oss_id", "f_oss_key", "f_ext", "f_size", "f_err_msg", "f_create_time", "f_modify_time", "f_expire_time")
SELECT s."f_id", s."f_hash", s."f_type", s."f_status", s."f_oss_id", s."f_oss_key", s."f_ext", s."f_size", s."f_err_msg", s."f_create_time", s."f_modify_time", s."f_expire_time"
FROM adp."t_task_cache_7" s;


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


-- t_task_cache_8
INSERT INTO kweaver."t_task_cache_8" ("f_id", "f_hash", "f_type", "f_status", "f_oss_id", "f_oss_key", "f_ext", "f_size", "f_err_msg", "f_create_time", "f_modify_time", "f_expire_time")
SELECT s."f_id", s."f_hash", s."f_type", s."f_status", s."f_oss_id", s."f_oss_key", s."f_ext", s."f_size", s."f_err_msg", s."f_create_time", s."f_modify_time", s."f_expire_time"
FROM adp."t_task_cache_8" s;


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


-- t_task_cache_9
INSERT INTO kweaver."t_task_cache_9" ("f_id", "f_hash", "f_type", "f_status", "f_oss_id", "f_oss_key", "f_ext", "f_size", "f_err_msg", "f_create_time", "f_modify_time", "f_expire_time")
SELECT s."f_id", s."f_hash", s."f_type", s."f_status", s."f_oss_id", s."f_oss_key", s."f_ext", s."f_size", s."f_err_msg", s."f_create_time", s."f_modify_time", s."f_expire_time"
FROM adp."t_task_cache_9" s;


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


-- t_task_cache_a
INSERT INTO kweaver."t_task_cache_a" ("f_id", "f_hash", "f_type", "f_status", "f_oss_id", "f_oss_key", "f_ext", "f_size", "f_err_msg", "f_create_time", "f_modify_time", "f_expire_time")
SELECT s."f_id", s."f_hash", s."f_type", s."f_status", s."f_oss_id", s."f_oss_key", s."f_ext", s."f_size", s."f_err_msg", s."f_create_time", s."f_modify_time", s."f_expire_time"
FROM adp."t_task_cache_a" s;


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


-- t_task_cache_b
INSERT INTO kweaver."t_task_cache_b" ("f_id", "f_hash", "f_type", "f_status", "f_oss_id", "f_oss_key", "f_ext", "f_size", "f_err_msg", "f_create_time", "f_modify_time", "f_expire_time")
SELECT s."f_id", s."f_hash", s."f_type", s."f_status", s."f_oss_id", s."f_oss_key", s."f_ext", s."f_size", s."f_err_msg", s."f_create_time", s."f_modify_time", s."f_expire_time"
FROM adp."t_task_cache_b" s;


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


-- t_task_cache_c
INSERT INTO kweaver."t_task_cache_c" ("f_id", "f_hash", "f_type", "f_status", "f_oss_id", "f_oss_key", "f_ext", "f_size", "f_err_msg", "f_create_time", "f_modify_time", "f_expire_time")
SELECT s."f_id", s."f_hash", s."f_type", s."f_status", s."f_oss_id", s."f_oss_key", s."f_ext", s."f_size", s."f_err_msg", s."f_create_time", s."f_modify_time", s."f_expire_time"
FROM adp."t_task_cache_c" s;


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


-- t_task_cache_d
INSERT INTO kweaver."t_task_cache_d" ("f_id", "f_hash", "f_type", "f_status", "f_oss_id", "f_oss_key", "f_ext", "f_size", "f_err_msg", "f_create_time", "f_modify_time", "f_expire_time")
SELECT s."f_id", s."f_hash", s."f_type", s."f_status", s."f_oss_id", s."f_oss_key", s."f_ext", s."f_size", s."f_err_msg", s."f_create_time", s."f_modify_time", s."f_expire_time"
FROM adp."t_task_cache_d" s;


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


-- t_task_cache_e
INSERT INTO kweaver."t_task_cache_e" ("f_id", "f_hash", "f_type", "f_status", "f_oss_id", "f_oss_key", "f_ext", "f_size", "f_err_msg", "f_create_time", "f_modify_time", "f_expire_time")
SELECT s."f_id", s."f_hash", s."f_type", s."f_status", s."f_oss_id", s."f_oss_key", s."f_ext", s."f_size", s."f_err_msg", s."f_create_time", s."f_modify_time", s."f_expire_time"
FROM adp."t_task_cache_e" s;


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


-- t_task_cache_f
INSERT INTO kweaver."t_task_cache_f" ("f_id", "f_hash", "f_type", "f_status", "f_oss_id", "f_oss_key", "f_ext", "f_size", "f_err_msg", "f_create_time", "f_modify_time", "f_expire_time")
SELECT s."f_id", s."f_hash", s."f_type", s."f_status", s."f_oss_id", s."f_oss_key", s."f_ext", s."f_size", s."f_err_msg", s."f_create_time", s."f_modify_time", s."f_expire_time"
FROM adp."t_task_cache_f" s;


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


-- t_dag_instance_event
INSERT INTO kweaver."t_dag_instance_event" ("f_id", "f_type", "f_instance_id", "f_operator", "f_task_id", "f_status", "f_name", "f_data", "f_size", "f_inline", "f_visibility", "f_timestamp")
SELECT s."f_id", s."f_type", s."f_instance_id", s."f_operator", s."f_task_id", s."f_status", s."f_name", s."f_data", s."f_size", s."f_inline", s."f_visibility", s."f_timestamp"
FROM adp."t_dag_instance_event" s;


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


-- t_cron_job
INSERT INTO kweaver."t_cron_job" ("f_key_id", "f_job_id", "f_job_name", "f_job_cron_time", "f_job_type", "f_job_context", "f_tenant_id", "f_enabled", "f_remarks", "f_create_time", "f_update_time")
SELECT s."f_key_id", s."f_job_id", s."f_job_name", s."f_job_cron_time", s."f_job_type", s."f_job_context", s."f_tenant_id", s."f_enabled", s."f_remarks", s."f_create_time", s."f_update_time"
FROM adp."t_cron_job" s;


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


-- t_cron_job_status
INSERT INTO kweaver."t_cron_job_status" ("f_key_id", "f_execute_id", "f_job_id", "f_job_type", "f_job_name", "f_job_status", "f_begin_time", "f_end_time", "f_executor", "f_execute_times", "f_ext_info")
SELECT s."f_key_id", s."f_execute_id", s."f_job_id", s."f_job_type", s."f_job_name", s."f_job_status", s."f_begin_time", s."f_end_time", s."f_executor", s."f_execute_times", s."f_ext_info"
FROM adp."t_cron_job_status" s;


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


-- t_flow_dag
INSERT INTO kweaver."t_flow_dag" ("f_id", "f_created_at", "f_updated_at", "f_user_id", "f_name", "f_desc", "f_trigger", "f_cron", "f_vars", "f_status", "f_tasks", "f_steps", "f_description", "f_shortcuts", "f_accessors", "f_type", "f_policy_type", "f_appinfo", "f_priority", "f_removed", "f_emails", "f_template", "f_published", "f_trigger_config", "f_sub_ids", "f_exec_mode", "f_category", "f_outputs", "f_instructions", "f_operator_id", "f_inc_values", "f_version", "f_version_id", "f_modify_by", "f_is_debug", "f_debug_id", "f_biz_domain_id")
SELECT s."f_id", s."f_created_at", s."f_updated_at", s."f_user_id", s."f_name", s."f_desc", s."f_trigger", s."f_cron", s."f_vars", s."f_status", s."f_tasks", s."f_steps", s."f_description", s."f_shortcuts", s."f_accessors", s."f_type", s."f_policy_type", s."f_appinfo", s."f_priority", s."f_removed", s."f_emails", s."f_template", s."f_published", s."f_trigger_config", s."f_sub_ids", s."f_exec_mode", s."f_category", s."f_outputs", s."f_instructions", s."f_operator_id", s."f_inc_values", s."f_version", s."f_version_id", s."f_modify_by", s."f_is_debug", s."f_debug_id", s."f_biz_domain_id"
FROM adp."t_flow_dag" s;


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


-- t_flow_dag_var
INSERT INTO kweaver."t_flow_dag_var" ("f_id", "f_dag_id", "f_var_name", "f_default_value", "f_var_type", "f_description")
SELECT s."f_id", s."f_dag_id", s."f_var_name", s."f_default_value", s."f_var_type", s."f_description"
FROM adp."t_flow_dag_var" s;


CREATE TABLE IF NOT EXISTS "t_flow_dag_instance_keyword" (
 "f_id" BIGINT NOT NULL,
 "f_dag_ins_id" BIGINT NOT NULL DEFAULT 0,
 "f_keyword" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS "idx_dag_ins_kw" ON "t_flow_dag_instance_keyword" ("f_dag_ins_id", "f_keyword");


-- t_flow_dag_instance_keyword
INSERT INTO kweaver."t_flow_dag_instance_keyword" ("f_id", "f_dag_ins_id", "f_keyword")
SELECT s."f_id", s."f_dag_ins_id", s."f_keyword"
FROM adp."t_flow_dag_instance_keyword" s;


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


-- t_flow_dag_step
INSERT INTO kweaver."t_flow_dag_step" ("f_id", "f_dag_id", "f_operator", "f_source_id", "f_has_datasource")
SELECT s."f_id", s."f_dag_id", s."f_operator", s."f_source_id", s."f_has_datasource"
FROM adp."t_flow_dag_step" s;


CREATE TABLE IF NOT EXISTS "t_flow_dag_accessor" (
 "f_id" BIGINT NOT NULL,
 "f_dag_id" BIGINT NOT NULL DEFAULT 0,
 "f_accessor_id" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS "idx_dag_accessor_id_dag" ON "t_flow_dag_accessor" ("f_accessor_id", "f_dag_id");


-- t_flow_dag_accessor
INSERT INTO kweaver."t_flow_dag_accessor" ("f_id", "f_dag_id", "f_accessor_id")
SELECT s."f_id", s."f_dag_id", s."f_accessor_id"
FROM adp."t_flow_dag_accessor" s;


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


-- t_flow_dag_version
INSERT INTO kweaver."t_flow_dag_version" ("f_id", "f_created_at", "f_updated_at", "f_dag_id", "f_user_id", "f_version", "f_version_id", "f_change_log", "f_config", "f_sort_time")
SELECT s."f_id", s."f_created_at", s."f_updated_at", s."f_dag_id", s."f_user_id", s."f_version", s."f_version_id", s."f_change_log", s."f_config", s."f_sort_time"
FROM adp."t_flow_dag_version" s;


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


-- t_flow_dag_instance
INSERT INTO kweaver."t_flow_dag_instance" ("f_id", "f_created_at", "f_updated_at", "f_dag_id", "f_trigger", "f_worker", "f_source", "f_vars", "f_keywords", "f_event_persistence", "f_event_oss_path", "f_share_data", "f_share_data_ext", "f_status", "f_reason", "f_cmd", "f_has_cmd", "f_batch_run_id", "f_user_id", "f_ended_at", "f_dag_type", "f_policy_type", "f_appinfo", "f_priority", "f_mode", "f_dump", "f_dump_ext", "f_success_callback", "f_error_callback", "f_call_chain", "f_resume_data", "f_resume_status", "f_version", "f_version_id", "f_biz_domain_id")
SELECT s."f_id", s."f_created_at", s."f_updated_at", s."f_dag_id", s."f_trigger", s."f_worker", s."f_source", s."f_vars", s."f_keywords", s."f_event_persistence", s."f_event_oss_path", s."f_share_data", s."f_share_data_ext", s."f_status", s."f_reason", s."f_cmd", s."f_has_cmd", s."f_batch_run_id", s."f_user_id", s."f_ended_at", s."f_dag_type", s."f_policy_type", s."f_appinfo", s."f_priority", s."f_mode", s."f_dump", s."f_dump_ext", s."f_success_callback", s."f_error_callback", s."f_call_chain", s."f_resume_data", s."f_resume_status", s."f_version", s."f_version_id", s."f_biz_domain_id"
FROM adp."t_flow_dag_instance" s;


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


-- t_flow_inbox
INSERT INTO kweaver."t_flow_inbox" ("f_id", "f_created_at", "f_updated_at", "f_msg", "f_topic", "f_docid", "f_dag")
SELECT s."f_id", s."f_created_at", s."f_updated_at", s."f_msg", s."f_topic", s."f_docid", s."f_dag"
FROM adp."t_flow_inbox" s;


CREATE TABLE IF NOT EXISTS "t_flow_outbox" (
 "f_id" BIGINT NOT NULL,
 "f_created_at" BIGINT NOT NULL DEFAULT 0,
 "f_updated_at" BIGINT NOT NULL DEFAULT 0,
 "f_msg" TEXT DEFAULT NULL,
 "f_topic" VARCHAR(128 CHAR) NOT NULL DEFAULT '',
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS "idx_outbox_created" ON "t_flow_outbox" ("f_created_at");


-- t_flow_outbox
INSERT INTO kweaver."t_flow_outbox" ("f_id", "f_created_at", "f_updated_at", "f_msg", "f_topic")
SELECT s."f_id", s."f_created_at", s."f_updated_at", s."f_msg", s."f_topic"
FROM adp."t_flow_outbox" s;


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


-- t_flow_task_instance
INSERT INTO kweaver."t_flow_task_instance" ("f_id", "f_created_at", "f_updated_at", "f_expired_at", "f_task_id", "f_dag_ins_id", "f_name", "f_depend_on", "f_action_name", "f_timeout_secs", "f_params", "f_traces", "f_status", "f_reason", "f_pre_checks", "f_results", "f_steps", "f_last_modified_at", "f_rendered_params", "f_hash", "f_settings", "f_metadata")
SELECT s."f_id", s."f_created_at", s."f_updated_at", s."f_expired_at", s."f_task_id", s."f_dag_ins_id", s."f_name", s."f_depend_on", s."f_action_name", s."f_timeout_secs", s."f_params", s."f_traces", s."f_status", s."f_reason", s."f_pre_checks", s."f_results", s."f_steps", s."f_last_modified_at", s."f_rendered_params", s."f_hash", s."f_settings", s."f_metadata"
FROM adp."t_flow_task_instance" s;


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


-- t_flow_token
INSERT INTO kweaver."t_flow_token" ("f_id", "f_created_at", "f_updated_at", "f_user_id", "f_user_name", "f_refresh_token", "f_token", "f_expires_in", "f_login_ip", "f_is_app")
SELECT s."f_id", s."f_created_at", s."f_updated_at", s."f_user_id", s."f_user_name", s."f_refresh_token", s."f_token", s."f_expires_in", s."f_login_ip", s."f_is_app"
FROM adp."t_flow_token" s;


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


-- t_flow_client
INSERT INTO kweaver."t_flow_client" ("f_id", "f_created_at", "f_updated_at", "f_client_name", "f_client_id", "f_client_secret")
SELECT s."f_id", s."f_created_at", s."f_updated_at", s."f_client_name", s."f_client_id", s."f_client_secret"
FROM adp."t_flow_client" s;


CREATE TABLE IF NOT EXISTS "t_flow_switch" (
 "f_id" BIGINT NOT NULL,
 "f_created_at" BIGINT NOT NULL DEFAULT 0,
 "f_updated_at" BIGINT NOT NULL DEFAULT 0,
 "f_name" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
 "f_status" TINYINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS "idx_switch_name" ON "t_flow_switch" ("f_name");


-- t_flow_switch
INSERT INTO kweaver."t_flow_switch" ("f_id", "f_created_at", "f_updated_at", "f_name", "f_status")
SELECT s."f_id", s."f_created_at", s."f_updated_at", s."f_name", s."f_status"
FROM adp."t_flow_switch" s;


CREATE TABLE IF NOT EXISTS "t_flow_log" (
 "f_id" BIGINT NOT NULL,
 "f_created_at" BIGINT NOT NULL DEFAULT 0,
 "f_updated_at" BIGINT NOT NULL DEFAULT 0,
 "f_ossid" VARCHAR(64 CHAR) NOT NULL DEFAULT '',
 "f_key" VARCHAR(40 CHAR) NOT NULL DEFAULT '',
 "f_filename" VARCHAR(255 CHAR) NOT NULL DEFAULT '',
  CLUSTER PRIMARY KEY ("f_id")
);


-- t_flow_log
INSERT INTO kweaver."t_flow_log" ("f_id", "f_created_at", "f_updated_at", "f_ossid", "f_key", "f_filename")
SELECT s."f_id", s."f_created_at", s."f_updated_at", s."f_ossid", s."f_key", s."f_filename"
FROM adp."t_flow_log" s;


CREATE TABLE IF NOT EXISTS "t_flow_storage" (
  "f_id" BIGINT NOT NULL,
  "f_oss_id" VARCHAR(64 CHAR) NOT NULL DEFAULT '',
  "f_object_key" VARCHAR(512 CHAR) NOT NULL DEFAULT '',
  "f_name" VARCHAR(256 CHAR) NOT NULL DEFAULT '',
  "f_content_type" VARCHAR(128 CHAR) NOT NULL DEFAULT '',
  "f_size" BIGINT NOT NULL DEFAULT 0,
  "f_etag" VARCHAR(128 CHAR) NOT NULL DEFAULT '',
  "f_status" TINYINT NOT NULL DEFAULT 1,
  "f_created_at" BIGINT NOT NULL DEFAULT 0,
  "f_updated_at" BIGINT NOT NULL DEFAULT 0,
  "f_deleted_at" BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS "uk_flow_storage_oss_id_object_key" ON "t_flow_storage" ("f_oss_id", "f_object_key");
CREATE INDEX IF NOT EXISTS "idx_flow_storage_status" ON "t_flow_storage" ("f_status");
CREATE INDEX IF NOT EXISTS "idx_flow_storage_created_at" ON "t_flow_storage" ("f_created_at");


-- t_flow_storage
INSERT INTO kweaver."t_flow_storage" ("f_id", "f_oss_id", "f_object_key", "f_name", "f_content_type", "f_size", "f_etag", "f_status", "f_created_at", "f_updated_at", "f_deleted_at")
SELECT s."f_id", s."f_oss_id", s."f_object_key", s."f_name", s."f_content_type", s."f_size", s."f_etag", s."f_status", s."f_created_at", s."f_updated_at", s."f_deleted_at"
FROM adp."t_flow_storage" s;


CREATE TABLE IF NOT EXISTS "t_flow_file" (
  "f_id" BIGINT NOT NULL,
  "f_dag_id" VARCHAR(64 CHAR) NOT NULL DEFAULT '',
  "f_dag_instance_id" VARCHAR(64 CHAR) NOT NULL DEFAULT '',
  "f_storage_id" BIGINT NOT NULL DEFAULT 0,
  "f_status" TINYINT NOT NULL DEFAULT 1,
  "f_name" VARCHAR(256 CHAR) NOT NULL DEFAULT '',
  "f_expires_at" BIGINT NOT NULL DEFAULT 0,
  "f_created_at" BIGINT NOT NULL DEFAULT 0,
  "f_updated_at" BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE INDEX IF NOT EXISTS "idx_flow_file_dag_id" ON "t_flow_file" ("f_dag_id");
CREATE INDEX IF NOT EXISTS "idx_flow_file_dag_instance_id" ON "t_flow_file" ("f_dag_instance_id");
CREATE INDEX IF NOT EXISTS "idx_flow_file_storage_id" ON "t_flow_file" ("f_storage_id");
CREATE INDEX IF NOT EXISTS "idx_flow_file_status" ON "t_flow_file" ("f_status");
CREATE INDEX IF NOT EXISTS "idx_flow_file_expires_at" ON "t_flow_file" ("f_expires_at");


-- t_flow_file
INSERT INTO kweaver."t_flow_file" ("f_id", "f_dag_id", "f_dag_instance_id", "f_storage_id", "f_status", "f_name", "f_expires_at", "f_created_at", "f_updated_at")
SELECT s."f_id", s."f_dag_id", s."f_dag_instance_id", s."f_storage_id", s."f_status", s."f_name", s."f_expires_at", s."f_created_at", s."f_updated_at"
FROM adp."t_flow_file" s;


CREATE TABLE IF NOT EXISTS "t_flow_file_download_job" (
  "f_id" BIGINT NOT NULL,
  "f_file_id" BIGINT NOT NULL,
  "f_status" TINYINT NOT NULL DEFAULT 1,
  "f_retry_count" INT NOT NULL DEFAULT 0,
  "f_max_retry" INT NOT NULL DEFAULT 3,
  "f_next_retry_at" BIGINT NOT NULL DEFAULT 0,
  "f_error_code" VARCHAR(64 CHAR) NOT NULL DEFAULT '',
  "f_error_message" VARCHAR(1024 CHAR) NOT NULL DEFAULT '',
  "f_download_url" VARCHAR(2048 CHAR) NOT NULL DEFAULT '',
  "f_started_at" BIGINT NOT NULL DEFAULT 0,
  "f_finished_at" BIGINT NOT NULL DEFAULT 0,
  "f_created_at" BIGINT NOT NULL DEFAULT 0,
  "f_updated_at" BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS "uk_flow_file_download_job_file_id" ON "t_flow_file_download_job" ("f_file_id");
CREATE INDEX IF NOT EXISTS "idx_flow_file_download_job_status_retry" ON "t_flow_file_download_job" ("f_status", "f_next_retry_at");


-- t_flow_file_download_job
INSERT INTO kweaver."t_flow_file_download_job" ("f_id", "f_file_id", "f_status", "f_retry_count", "f_max_retry", "f_next_retry_at", "f_error_code", "f_error_message", "f_download_url", "f_started_at", "f_finished_at", "f_created_at", "f_updated_at")
SELECT s."f_id", s."f_file_id", s."f_status", s."f_retry_count", s."f_max_retry", s."f_next_retry_at", s."f_error_code", s."f_error_message", s."f_download_url", s."f_started_at", s."f_finished_at", s."f_created_at", s."f_updated_at"
FROM adp."t_flow_file_download_job" s;


CREATE TABLE IF NOT EXISTS "t_flow_task_resume" (
  "f_id" BIGINT NOT NULL,
  "f_task_instance_id" VARCHAR(64 CHAR) NOT NULL DEFAULT '',
  "f_dag_instance_id" VARCHAR(64 CHAR) NOT NULL DEFAULT '',
  "f_resource_type" VARCHAR(32 CHAR) NOT NULL DEFAULT 'file',
  "f_resource_id" BIGINT NOT NULL DEFAULT 0,
  "f_created_at" BIGINT NOT NULL DEFAULT 0,
  "f_updated_at" BIGINT NOT NULL DEFAULT 0,
  CLUSTER PRIMARY KEY ("f_id")
);

CREATE UNIQUE INDEX IF NOT EXISTS "uk_flow_task_resume_task_instance_id" ON "t_flow_task_resume" ("f_task_instance_id");
CREATE INDEX IF NOT EXISTS "idx_flow_task_resume_resource" ON "t_flow_task_resume" ("f_resource_type", "f_resource_id");

-- t_flow_task_resume
INSERT INTO kweaver."t_flow_task_resume" ("f_id", "f_task_instance_id", "f_dag_instance_id", "f_resource_type", "f_resource_id", "f_created_at", "f_updated_at")
SELECT s."f_id", s."f_task_instance_id", s."f_dag_instance_id", s."f_resource_type", s."f_resource_id", s."f_created_at", s."f_updated_at"
FROM adp."t_flow_task_resume" s;