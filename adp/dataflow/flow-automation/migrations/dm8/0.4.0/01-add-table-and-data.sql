SET SCHEMA adp;


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
