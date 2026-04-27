SET SCHEMA adp;

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